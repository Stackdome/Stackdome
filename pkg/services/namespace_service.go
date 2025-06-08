package services

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/google/uuid"
)

type NamespaceService interface {
	CreateInDB(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	CreateInCluster(ctx context.Context, ns *models.Namespace) *errors.ServiceError
	CreateInDBWithTx(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError)
	DeleteFromDB(ctx context.Context, id string) *errors.ServiceError
	InternalDeleteFromDB(ctx context.Context, id string) *errors.ServiceError
	DeleteFromDBWithTx(ctx context.Context, id string) *errors.ServiceError
	UpdateInDB(ctx context.Context, id string, ns *models.Namespace) (*models.Namespace, *errors.ServiceError)
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Namespace, *errors.ServiceError)
	ListByStack(ctx context.Context, stackID string) ([]*models.Namespace, *errors.ServiceError)
	PrepareNamespaceForStack(ctx context.Context, stack *models.Stack) (*models.Namespace, *errors.ServiceError)
	ClusterResourceServiceInjectable
}

type namespaceService struct {
	namespacesStore stores.NamespacesStore
	logger          logger.Logger
	ClusterResourceServiceDeps
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

func (s *namespaceService) PrepareNamespaceForStack(ctx context.Context, stack *models.Stack) (*models.Namespace, *errors.ServiceError) {
	// Generate a unique namespace for the stack
	namespace := &models.Namespace{
		Name:           fmt.Sprintf("%s-%s", stack.Name, uuid.New().String()),
		OrganisationID: stack.OrganisationID,
	}
	namespace.AddDefaultLabels()

	return namespace, nil
}

func (s *namespaceService) CreateInCluster(ctx context.Context, ns *models.Namespace) *errors.ServiceError {
	namespace, err := s.namespacesStore.Get(ctx, ns.ID)
	if err != nil {
		return err
	}
	if err := s.ClusterNamespaceService.CreateNamespaceInCluster(ctx, namespace); err != nil {
		return errors.GeneralError("failed to create namespace in cluster: %v", err)
	}
	return nil
}

func (s *namespaceService) CreateInDB(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	if len(ns.Name) == 0 {
		return nil, errors.BadRequest("namespace name is required")
	}
	if len(ns.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}
	return s.namespacesStore.Create(ctx, ns)
}

func (s *namespaceService) CreateInDBWithTx(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
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

func (s *namespaceService) DeleteFromDB(ctx context.Context, id string) *errors.ServiceError {
	return s.namespacesStore.Delete(ctx, id)
}

func (s *namespaceService) InternalDeleteFromDB(ctx context.Context, id string) *errors.ServiceError {
	if err := s.DeleteFromDB(ctx, id); err != nil {
		if err.Is404() {
			return nil
		}
		return err
	}
	return nil
}

func (s *namespaceService) DeleteFromDBWithTx(ctx context.Context, id string) *errors.ServiceError {
	return s.namespacesStore.DeleteWithTx(ctx, id)
}

func (s *namespaceService) UpdateInDB(ctx context.Context, id string, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.Update(ctx, id, ns)
}

func (s *namespaceService) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.ListByOrganisation(ctx, organisationID)
}

func (s *namespaceService) ListByStack(ctx context.Context, stackID string) ([]*models.Namespace, *errors.ServiceError) {
	return s.namespacesStore.ListByStack(ctx, stackID)
}
