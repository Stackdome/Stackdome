package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type OrganisationStore interface {
	GetDefaultOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError)
	Create(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Update(ctx context.Context, id string, org *models.Organisation) (*models.Organisation, *errors.ServiceError)
}
