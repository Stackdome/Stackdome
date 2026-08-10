package stores

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=postgres_addon_store.go -destination=../mocks/mock_postgres_addon_store.go -package=mocks
type PostgresAddonStore interface {
	Create(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	CreateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresAddon, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.PostgresAddon, *errors.ServiceError)
	Update(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	UpdateLifecycleWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.PostgresAddon, *errors.ServiceError)
	ListByProjectID(ctx context.Context, projectID string) ([]*models.PostgresAddon, *errors.ServiceError)
	ListByProjectIDs(ctx context.Context, projectIDs []string) ([]*models.PostgresAddon, *errors.ServiceError)

	// Validation
	ValidateAddonExists(ctx context.Context, addonID string) (bool, *errors.ServiceError)
	ValidateAddonNameUnique(ctx context.Context, organisationID, name, excludeID string) (bool, *errors.ServiceError)

	// Status and lifecycle management
	UpdateStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError
	UpdateBackupRequestedAt(ctx context.Context, id string, timestamp *time.Time) *errors.ServiceError
	UpdateDeletionTimestamp(ctx context.Context, id string, timestamp *time.Time) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}
