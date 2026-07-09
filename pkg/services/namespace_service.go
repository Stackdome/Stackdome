package services

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/google/uuid"
)

//go:generate mockgen -destination=../mocks/mock_namespace_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services NamespaceService

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
	PrepareNamespaceForAddon(ctx context.Context, addon models.Addon, orgID string) (*models.Namespace, *errors.ServiceError)
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
	// Generate a unique namespace for the stack. The result must be a valid
	// RFC 1123 DNS label (at most models.KubernetesDNSLabelMaxLength
	// characters); the stack validator caps names at
	// models.MaxStackNameLength so the "-<uuid>" suffix
	// (models.NamespaceUUIDSuffixLength) always fits.
	namespace := &models.Namespace{
		Name:           fmt.Sprintf("%s-%s", stack.Name, uuid.New().String()),
		OrganisationID: stack.OrganisationID,
	}
	namespace.AddDefaultLabels()

	return namespace, nil
}

func (s *namespaceService) PrepareNamespaceForAddon(ctx context.Context, addon models.Addon, organisationID string) (*models.Namespace, *errors.ServiceError) {
	// Generate a unique namespace for the addon
	namespace := &models.Namespace{
		Name:           fmt.Sprintf("stackdome-addons-%s-%s", addon.Type(), addon.AddonName()),
		OrganisationID: organisationID,
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
