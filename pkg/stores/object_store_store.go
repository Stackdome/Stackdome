package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type ObjectStoreStore interface {
	Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError)
	Update(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string) ([]*models.ObjectStore, *errors.ServiceError)
	ListByTeamIDs(ctx context.Context, teamIDs []string) ([]*models.ObjectStore, *errors.ServiceError)

	UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError

	// Validation
	ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError)
	ValidateObjectStoreNameUnique(ctx context.Context, organisationID, name, excludeID string) (bool, *errors.ServiceError)

	IsReferencedByAddon(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError)
}
