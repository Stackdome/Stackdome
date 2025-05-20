package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type StackDomainsStore interface {
	CreateWithTx(ctx context.Context, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	BulkCreate(ctx context.Context, domains []*models.StackDomain) (models.StackDomainList, *errors.ServiceError)
	Create(ctx context.Context, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.StackDomain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Update(ctx context.Context, id string, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	ListByStackID(ctx context.Context, stackID string) (models.StackDomainList, *errors.ServiceError)
	ListByStackResourceID(ctx context.Context, stackResourceID string) (models.StackDomainList, *errors.ServiceError)
	GetByStackResourceAndPort(ctx context.Context, stackResourceID string, port int) (*models.StackDomain, *errors.ServiceError)
	DeleteForStackResourceAndPort(ctx context.Context, stackResourceID string, port int) *errors.ServiceError
	DeleteForStackResourceAndPortWithTx(ctx context.Context, stackResourceID string, port int) *errors.ServiceError
	GetByFqdn(ctx context.Context, fqdn string) (*models.StackDomain, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) (models.StackDomainList, *errors.ServiceError)
	ListByFqdnPrefix(ctx context.Context, prefix string) (models.StackDomainList, *errors.ServiceError)
}
