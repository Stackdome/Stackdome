package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=organisation_store.go -destination=../mocks/mock_organisation_store.go -package=mocks
type OrganisationStore interface {
	GetPlatformOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError)
	OrganisationNameExists(ctx context.Context, name string) (bool, *errors.ServiceError)
	Create(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
	// LockByID takes a row-level lock on the organisation. It must be called
	// inside a transaction to serialize topology mutations owned by that org.
	LockByID(ctx context.Context, ID string) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Update(ctx context.Context, id string, org *models.Organisation) (*models.Organisation, *errors.ServiceError)
}
