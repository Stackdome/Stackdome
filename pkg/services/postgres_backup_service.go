package services

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type PostgresBackupService interface {
	Create(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresBackup, *errors.ServiceError)
	Update(ctx context.Context, id string, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError)
	ValidateBackupExists(ctx context.Context, backupID string) (bool, *errors.ServiceError)

	// Status update methods for backup lifecycle
	UpdateBackupStatus(ctx context.Context, backupID, phase string, error *string) *errors.ServiceError
	MarkBackupCompleted(ctx context.Context, backupID string, sizeBytes *int32) *errors.ServiceError
	MarkBackupFailed(ctx context.Context, backupID string, errorMessage string) *errors.ServiceError
}

type PostgresBackupServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

type postgresBackupService struct {
	backupStore stores.PostgresBackupStore
	logger      logger.Logger
}

func NewPostgresBackupService(spec PostgresBackupServiceSpec) PostgresBackupService {
	return &postgresBackupService{
		backupStore: pgstore.NewPostgresBackupStore(pgstore.PostgresBackupStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *postgresBackupService) Create(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError) {
	// Set initial backup state
	if backup.Phase == "" {
		backup.Phase = "Pending"
	}
	if backup.Type == "" {
		backup.Type = "Manual"
	}

	// Set start time
	now := time.Now()
	backup.StartedAt = &now

	createdBackup, err := s.backupStore.Create(ctx, backup)
	if err != nil {
		return nil, err
	}

	return createdBackup, nil
}

func (s *postgresBackupService) GetByID(ctx context.Context, ID string) (*models.PostgresBackup, *errors.ServiceError) {
	backup, err := s.backupStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return backup, nil
}

func (s *postgresBackupService) Update(ctx context.Context, id string, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError) {
	existingBackup, err := s.backupStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Preserve immutable fields
	backup.ID = existingBackup.ID
	backup.PostgresAddonID = existingBackup.PostgresAddonID
	backup.StartedAt = existingBackup.StartedAt

	updatedBackup, err := s.backupStore.Update(ctx, backup)
	if err != nil {
		return nil, err
	}

	return updatedBackup, nil
}

func (s *postgresBackupService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := s.backupStore.Delete(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *postgresBackupService) ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError) {
	backups, err := s.backupStore.ListByPostgresAddon(ctx, postgresAddonID)
	if err != nil {
		return nil, err
	}
	return backups, nil
}

func (s *postgresBackupService) ValidateBackupExists(ctx context.Context, backupID string) (bool, *errors.ServiceError) {
	return s.backupStore.ValidateBackupExists(ctx, backupID)
}

func (s *postgresBackupService) UpdateBackupStatus(ctx context.Context, backupID, phase string, errorMessage *string) *errors.ServiceError {
	backup, err := s.backupStore.GetByID(ctx, backupID)
	if err != nil {
		return err
	}

	backup.Phase = phase
	if errorMessage != nil {
		backup.Error = *errorMessage
	}

	_, err = s.backupStore.Update(ctx, backup)
	return err
}

func (s *postgresBackupService) MarkBackupCompleted(ctx context.Context, backupID string, sizeBytes *int32) *errors.ServiceError {
	backup, err := s.backupStore.GetByID(ctx, backupID)
	if err != nil {
		return err
	}

	now := time.Now()
	backup.Phase = "Completed"
	backup.CompletedAt = &now
	backup.SizeBytes = sizeBytes
	backup.Error = "" // Clear any previous errors

	_, err = s.backupStore.Update(ctx, backup)
	return err
}

func (s *postgresBackupService) MarkBackupFailed(ctx context.Context, backupID string, errorMessage string) *errors.ServiceError {
	backup, err := s.backupStore.GetByID(ctx, backupID)
	if err != nil {
		return err
	}

	now := time.Now()
	backup.Phase = "Failed"
	backup.CompletedAt = &now
	backup.Error = errorMessage

	_, err = s.backupStore.Update(ctx, backup)
	return err
}
