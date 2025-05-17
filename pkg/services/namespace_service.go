package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type NamespaceService interface {
	Create(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	CreateWithTx(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Update(ctx context.Context, id string, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Namespace, *errors.ServiceError)
	ListByStack(ctx context.Context, stackID string) ([]*models.Namespace, *errors.ServiceError)
}

type namespaceService struct {
	namespacesStore stores.NamespacesStore
	logger          logger.Logger
}

type NamespaceServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewNamespaceService(spec NamespaceServiceSpec) NamespaceService {
	return &namespaceService{
		namespacesStore: pgstore.NewNamespacesStore(pgstore.NamespacesStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *namespaceService) Create(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	if len(ns.Name) == 0 {
		return nil, errors.BadRequest("namespace name is required")
	}
	if len(ns.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}
	return s.namespacesStore.Create(ctx, ns)
}

func (s *namespaceService) CreateWithTx(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	if len(ns.Name) == 0 {
		return nil, errors.BadRequest("namespace name is required")
	}
	if len(ns.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}
	return s.namespacesStore.CreateWithTx(ctx, ns)
}

func (s *namespaceService) Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.Get(ctx, id)
}

func (s *namespaceService) Delete(ctx context.Context, id string) *errors.ServiceError {
	return s.namespacesStore.Delete(ctx, id)
}

func (s *namespaceService) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	return s.namespacesStore.DeleteWithTx(ctx, id)
}

func (s *namespaceService) Update(ctx context.Context, id string, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.Update(ctx, id, ns)
}

func (s *namespaceService) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.ListByOrganisation(ctx, organisationID)
}

func (s *namespaceService) ListByStack(ctx context.Context, stackID string) ([]*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.ListByStack(ctx, stackID)
}
