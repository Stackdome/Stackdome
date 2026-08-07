package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=organisation_domain.go -destination=../mocks/mock_organisation_domain_store.go -package=mocks

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
