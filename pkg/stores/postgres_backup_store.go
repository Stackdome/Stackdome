package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type PostgresBackupStore interface {
	Create(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresBackup, *errors.ServiceError)
	Update(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError)

	GetByName(ctx context.Context, postgresAddonID string, name string) (*models.PostgresBackup, *errors.ServiceError)

	// Validation
	ValidateBackupExists(ctx context.Context, backupID string) (bool, *errors.ServiceError)
}
