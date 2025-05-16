package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type DomainsStore interface {
	CreateWithTx(ctx context.Context, domain *models.Domain) (*models.Domain, *errors.ServiceError)
	BulkCreate(ctx context.Context, domains []*models.Domain) ([]*models.Domain, *errors.ServiceError)
	Create(ctx context.Context, domain *models.Domain) (*models.Domain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.Domain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Update(ctx context.Context, id string, domain *models.Domain) (*models.Domain, *errors.ServiceError)
	ListByOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) ([]*models.Domain, *errors.ServiceError)
	GetByFqdn(ctx context.Context, fqdn string) (*models.Domain, *errors.ServiceError)
	ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.Domain, *errors.ServiceError)
	DeleteForOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError
	DeleteForOwnerWithTx(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError
}
