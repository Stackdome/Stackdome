package postgres_addon

// Package postgres_addon contains the implementation of the PostgresAddon controller for managing PostgreSQL clusters in Kubernetes.

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/controllers"
	apperrors "github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
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
	r.Log.Infof("reconciling postgres addon: %s in namespace %s", req.Name, req.Namespace)

	clusterInstance := &addonsv1alpha1.PostgresCluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterInstance); err != nil {
		r.Log.Errorf("failed to get postgres cluster from cluster: %v", err)
		return ctrl.Result{}, nil
	}

	postgresAddonID, ok := clusterInstance.Labels[models.PostgresAddonIDLabel]
	if !ok {
		r.Log.Errorf("postgres addon ID not found in PostgresCluster labels")
		return ctrl.Result{}, nil
	}

	dbInstance, serr := r.PostgresAddonService.GetPostgresAddon(ctx, postgresAddonID)
	if serr != nil {
		if serr.Code == apperrors.ErrorNotFound {
			r.Log.Infof("postgres addon %s in namespace %s not found in DB", clusterInstance.Name, clusterInstance.Namespace)
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get postgres addon from db: %v", serr)
	}

	// Update if status hash changed, or if hash is empty (cluster-agent doesn't compute it yet)
	if clusterInstance.Status.StatusHash == "" || clusterInstance.Status.StatusHash != dbInstance.Status.LastObservedStatusHash {
		newStatus := mapToPostgresAddonStatus(clusterInstance.Status)
		serr = r.PostgresAddonService.UpdatePostgresAddonStatus(ctx, dbInstance.ID, newStatus)
		if serr != nil {
			return ctrl.Result{}, fmt.Errorf("failed to update postgres addon status in db: %v", serr)
		}
		r.Log.Infof("updated postgres addon %s status: phase=%s", postgresAddonID, newStatus.State)
	}

	return ctrl.Result{}, nil
}

func mapPhaseToState(phase string) string {
	switch phase {
	case "Cluster in healthy state":
		return "Ready"
	case addonsv1alpha1.PendingPhase, addonsv1alpha1.ErrorPhase,
		addonsv1alpha1.DeletingPhase, addonsv1alpha1.HibernatedPhase,
		addonsv1alpha1.ReadyPhase:
		return phase
	default:
		// Pass through any other CNPG phase strings as-is
		return phase
	}
}

func mapToPostgresAddonStatus(clusterStatus addonsv1alpha1.PostgresClusterStatus) *models.PostgresAddonStatus {
	status := &models.PostgresAddonStatus{
		State:                  mapPhaseToState(string(clusterStatus.Phase)),
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
			// Derive connection info from write service when ClusterConnection is not populated
			status.ConnectionInfo.Host = clusterStatus.Outputs.WriteService
			status.ConnectionInfo.Port = 5432
			status.ConnectionInfo.SSLMode = "verify-full"
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
