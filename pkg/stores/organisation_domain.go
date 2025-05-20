package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type OrganisationDomainStore interface {
	CreateWithTx(ctx context.Context, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	BulkCreate(ctx context.Context, domains []*models.OrganisationDomain) ([]*models.OrganisationDomain, *errors.ServiceError)
	Create(ctx context.Context, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.OrganisationDomain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Update(ctx context.Context, id string, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}
