package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type PostgresAddonDatabaseStore interface {
	Create(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	CreateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresAddonDatabase, *errors.ServiceError)
	GetByName(ctx context.Context, postgresAddonID, name string) (*models.PostgresAddonDatabase, *errors.ServiceError)
	Update(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresAddonDatabase, *errors.ServiceError)

	// Validation
	ValidateDatabaseExists(ctx context.Context, databaseID string) (bool, *errors.ServiceError)
	ValidateDatabaseNameUnique(ctx context.Context, postgresAddonID, name, excludeID string) (bool, *errors.ServiceError)
}
