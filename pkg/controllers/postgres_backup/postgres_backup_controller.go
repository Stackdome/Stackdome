package postgres_backup

import (
	"context"
	"fmt"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/source"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
)

const controllerName = "postgres-backup-controller"

type postgresBackupReconciler struct {
	Client                client.Client
	Log                   logger.Logger
	PostgresBackupService services.PostgresBackupService
}

type PostgresBackupReconcilerSpec struct {
	Log                   logger.Logger
	PostgresBackupService services.PostgresBackupService
}

func NewPostgresBackupReconciler(spec PostgresBackupReconcilerSpec) *postgresBackupReconciler {
	return &postgresBackupReconciler{
		Log:                   spec.Log,
		PostgresBackupService: spec.PostgresBackupService,
	}
}

func (r *postgresBackupReconciler) AddToManager(manager manager.Manager) error {
	r.Client = manager.GetClient()
	ctrl, err := controller.New(controllerName, manager, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	src := source.Kind(
		manager.GetCache(),
		&cnpgv1.Backup{},
		&handler.TypedEnqueueRequestForObject[*cnpgv1.Backup]{},
	)

	return ctrl.Watch(src)
}

func (r *postgresBackupReconciler) Name() string {
	return controllerName
}

func (r *postgresBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("reconciling backup: %s in namespace %s", req.Name, req.Namespace)

	backup := &cnpgv1.Backup{}
	if err := r.Client.Get(ctx, req.NamespacedName, backup); err != nil {
		r.Log.Errorf("failed to get CNPG Backup: %v", err)
		return ctrl.Result{}, nil
	}

	addonID, err := r.resolveAddonID(ctx, backup)
	if err != nil {
		r.Log.Errorf("failed to resolve addon ID for backup %s: %v", backup.Name, err)
		return ctrl.Result{}, nil
	}
	if addonID == "" {
		// Not a stackdome-managed backup
		return ctrl.Result{}, nil
	}

	existing, serr := r.PostgresBackupService.GetByName(ctx, addonID, backup.Name)
	if serr != nil && serr.Code == apperrors.ErrorNotFound {
		return r.createBackupRecord(ctx, addonID, backup)
	}
	if serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to lookup backup: %w", serr)
	}

	return r.updateBackupRecord(ctx, existing, backup)
}

// resolveAddonID finds the PostgresAddon ID for a CNPG Backup by matching
// the backup's cluster reference against PostgresCluster CRs in the same namespace.
func (r *postgresBackupReconciler) resolveAddonID(ctx context.Context, backup *cnpgv1.Backup) (string, error) {
	cnpgClusterName := backup.Spec.Cluster.Name

	pgClusterList := &addonsv1alpha1.PostgresClusterList{}
	if err := r.Client.List(ctx, pgClusterList, client.InNamespace(backup.Namespace)); err != nil {
		return "", fmt.Errorf("failed to list PostgresCluster CRs: %w", err)
	}

	for i := range pgClusterList.Items {
		pgCluster := &pgClusterList.Items[i]
		if pgCluster.CnpgClusterName() == cnpgClusterName {
			return pgCluster.Labels[models.PostgresAddonIDLabel], nil
		}
	}

	return "", nil
}

func (r *postgresBackupReconciler) createBackupRecord(ctx context.Context, addonID string, backup *cnpgv1.Backup) (ctrl.Result, error) {
	record := &models.PostgresBackup{
		PostgresAddonID: addonID,
		Name:            backup.Name,
		Type:            backupType(backup),
		Phase:           string(backup.Status.Phase),
		Error:           backup.Status.Error,
	}

	if backup.Status.StartedAt != nil {
		t := backup.Status.StartedAt.Time
		record.StartedAt = &t
	}
	if backup.Status.StoppedAt != nil {
		t := backup.Status.StoppedAt.Time
		record.CompletedAt = &t
	}

	if _, serr := r.PostgresBackupService.Create(ctx, record); serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create backup record: %w", serr)
	}

	r.Log.Infof("created backup record for %s (addon %s)", backup.Name, addonID)
	return ctrl.Result{}, nil
}

func (r *postgresBackupReconciler) updateBackupRecord(ctx context.Context, existing *models.PostgresBackup, backup *cnpgv1.Backup) (ctrl.Result, error) {
	newPhase := string(backup.Status.Phase)
	if existing.Phase == newPhase && existing.Error == backup.Status.Error {
		return ctrl.Result{}, nil
	}

	existing.Phase = newPhase
	existing.Error = backup.Status.Error

	if backup.Status.StartedAt != nil {
		t := backup.Status.StartedAt.Time
		existing.StartedAt = &t
	}
	if backup.Status.StoppedAt != nil {
		t := backup.Status.StoppedAt.Time
		existing.CompletedAt = &t
	}

	if _, serr := r.PostgresBackupService.Update(ctx, existing.ID, existing); serr != nil {
		return ctrl.Result{}, fmt.Errorf("failed to update backup record: %w", serr)
	}

	r.Log.Infof("updated backup %s phase to %s", backup.Name, newPhase)
	return ctrl.Result{}, nil
}

func backupType(backup *cnpgv1.Backup) models.PostgresBackupType {
	// Scheduled backups are owned by a ScheduledBackup resource
	for _, ref := range backup.OwnerReferences {
		if ref.Kind == "ScheduledBackup" {
			return models.PostgresBackupTypeScheduled
		}
	}
	return models.PostgresBackupTypeManual
}
