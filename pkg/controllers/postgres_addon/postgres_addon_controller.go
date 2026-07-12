package postgres_addon

// Package postgres_addon contains the implementation of the PostgresAddon controller for managing PostgreSQL clusters in Kubernetes.

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/controllers"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
)

const (
	controllerName = "postgres-addon-controller"
)

type postgresAddonReconciler struct {
	Client               client.Client
	Log                  logger.Logger
	PostgresAddonService services.PostgresAddonService
	Env                  string
}

type PostgresAddonReconcilerSpec struct {
	Log                  logger.Logger
	PostgresAddonService services.PostgresAddonService
	Env                  string
}

func NewPostgresAddonReconciler(spec PostgresAddonReconcilerSpec) *postgresAddonReconciler {
	return &postgresAddonReconciler{
		Log:                  spec.Log,
		PostgresAddonService: spec.PostgresAddonService,
		Env:                  spec.Env,
	}
}

func (r *postgresAddonReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	controller, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}
	src := source.Kind(
		manager.GetCache(),
		&addonsv1alpha1.PostgresCluster{},
		&handler.TypedEnqueueRequestForObject[*addonsv1alpha1.PostgresCluster]{},
		controllers.DBObjectIDPresentPredicate[*addonsv1alpha1.PostgresCluster](models.PostgresAddonIDLabel),
	)

	return controller.Watch(src)
}

func (r *postgresAddonReconciler) Name() string {
	return controllerName
}

func (r *postgresAddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info(ctx, "reconciling postgres addon: %s in namespace %s", req.Name, req.Namespace)

	clusterInstance := &addonsv1alpha1.PostgresCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterInstance); err != nil {
		r.Log.Error(ctx, "failed to get postgres cluster from cluster: %v", err)
		return ctrl.Result{}, nil
	}

	postgresAddonID, ok := clusterInstance.Labels[models.PostgresAddonIDLabel]
	if !ok {
		r.Log.Error(ctx, "postgres addon ID not found in PostgresCluster labels")
		return ctrl.Result{}, nil
	}

	dbInstance, serr := r.PostgresAddonService.InternalGetPostgresAddon(ctx, postgresAddonID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			r.Log.Info(ctx, "postgres addon %s in namespace %s not found in DB", clusterInstance.Name, clusterInstance.Namespace)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get postgres addon from db: %w", serr)
	}

	// TODO(cluster-agent): Remove StatusHash=="" fallback once cluster-agent computes
	// StatusHash on PostgresCluster CRs (see docs/plans/cluster-agent-fixes.md #2).
	if clusterInstance.Status.StatusHash == "" || clusterInstance.Status.StatusHash != dbInstance.Status.LastObservedStatusHash {
		newStatus := mapToPostgresAddonStatus(clusterInstance.Status)
		// Check if current status has ReadyOnce condition = true. If yes then that is always added.
		if models.IsConditionTrue(dbInstance.Status.Conditions, string(models.PostgresAddonConditionReadyOnce)) {
			readyOnceCond := models.FindCondition(dbInstance.Status.Conditions, string(models.PostgresAddonConditionReadyOnce))
			newStatus.Conditions = append(newStatus.Conditions, *readyOnceCond)
		}

		// If the new status is Ready and the old status doesn't have ReadyOnce=true, set ReadyOnce=true with current timestamp.
		// This ensures that ReadyOnce condition is set when addon reaches Ready state for the first time.
		if newStatus.State == models.PostgresAddonStateReady && !models.IsConditionTrue(dbInstance.Status.Conditions, string(models.PostgresAddonConditionReadyOnce)) {
			models.SetCondition(&newStatus.Conditions, models.Condition{
				Type:               string(models.PostgresAddonConditionReadyOnce),
				Status:             string(metav1.ConditionTrue),
				LastTransitionTime: time.Now().UTC(),
				Reason:             "PostgresReadyOnce",
				Message:            "Postgres reached Ready state at least once",
			})
		}

		if equality.Semantic.DeepEqual(dbInstance.Status, &newStatus) {
			r.Log.Info(ctx, "postgres addon %s status is up to date, skipping DB update", postgresAddonID)
			return ctrl.Result{}, nil
		}

		serr = r.PostgresAddonService.UpdatePostgresAddonStatus(ctx, dbInstance.ID, newStatus)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update postgres addon status in db: %w", serr)
		}
		r.Log.Info(ctx, "updated postgres addon %s status: phase=%s", postgresAddonID, newStatus.State)
	}

	return ctrl.Result{}, nil
}

// TODO(cluster-agent): Remove this mapping once cluster-agent uses defined phase
// constants instead of passing through raw CNPG strings (see docs/plans/cluster-agent-fixes.md #3).
func mapPhaseToState(phase string) models.PostgresAddonState {
	switch phase {
	case "Cluster in healthy state", addonsv1alpha1.ReadyPhase:
		return models.PostgresAddonStateReady
	case addonsv1alpha1.PendingPhase:
		return models.PostgresAddonStatePending
	case addonsv1alpha1.ErrorPhase:
		return models.PostgresAddonStateError
	case addonsv1alpha1.DeletingPhase:
		return models.PostgresAddonStateDeleting
	case addonsv1alpha1.HibernatedPhase:
		return models.PostgresAddonStateHibernated
	default:
		return models.PostgresAddonState(phase)
	}
}

func mapToPostgresAddonStatus(clusterStatus addonsv1alpha1.PostgresClusterStatus) *models.PostgresAddonStatus {
	status := &models.PostgresAddonStatus{
		State:                  mapPhaseToState(clusterStatus.Phase),
		Conditions:             models.ConvertConditions(clusterStatus.Conditions),
		LastObservedStatusHash: clusterStatus.StatusHash,
	}

	if clusterStatus.Outputs != nil {
		status.ConnectionInfo = &models.PostgresAddonConnectionInfo{
			WriteService: clusterStatus.Outputs.WriteService,
			ReadService:  clusterStatus.Outputs.ReadService,
			ClusterSecrets: &models.PostgresAddonClusterSecrets{
				SuperuserSecret:     clusterStatus.Outputs.SuperUserCredentialSecret,
				UserSecrets:         clusterStatus.Outputs.UserCredentialSecrets,
				CACertificateSecret: clusterStatus.Outputs.ClientCASecret,
			},
		}

		if clusterStatus.Outputs.ClusterConnection != nil {
			status.ConnectionInfo.Host = clusterStatus.Outputs.ClusterConnection.Host
			status.ConnectionInfo.Port = clusterStatus.Outputs.ClusterConnection.Port
			status.ConnectionInfo.SSLMode = clusterStatus.Outputs.ClusterConnection.SSLMode
		} else if clusterStatus.Outputs.WriteService != "" {
			// TODO(cluster-agent): Remove this fallback once cluster-agent populates
			// ClusterConnection in PostgresCluster outputs (see docs/plans/cluster-agent-fixes.md #4).
			status.ConnectionInfo.Host = clusterStatus.Outputs.WriteService
			status.ConnectionInfo.Port = 5432
			status.ConnectionInfo.SSLMode = "require"
		}

		status.Databases = make([]models.PostgresDatabaseInfo, len(clusterStatus.Outputs.Databases))
		for i, db := range clusterStatus.Outputs.Databases {
			status.Databases[i] = models.PostgresDatabaseInfo{
				Name:  db.Name,
				Owner: db.Owner,
			}
		}
	}

	return status
}
