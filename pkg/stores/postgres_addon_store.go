package stores

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type PostgresAddonStore interface {
	Create(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	CreateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.PostgresAddon, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.PostgresAddon, *errors.ServiceError)
	Update(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	UpdateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.PostgresAddon, *errors.ServiceError)
	ListByCluster(ctx context.Context, clusterID string) ([]*models.PostgresAddon, *errors.ServiceError)

	// Validation
	ValidateAddonExists(ctx context.Context, addonID string) (bool, *errors.ServiceError)
	ValidateAddonNameUnique(ctx context.Context, organisationID, name, excludeID string) (bool, *errors.ServiceError)

	// Status and lifecycle management
	UpdateStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError
	UpdateBackupRequestedAt(ctx context.Context, id string, timestamp *time.Time) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}
