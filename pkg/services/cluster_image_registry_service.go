package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

type ClusterImageRegistryService interface {
	Get(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterID(ctx context.Context, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError
	InjectClusterResourceService(registryClusterService clusterresource.ClusterImageRegistryService)
	Delete(ctx context.Context, ID string) *errors.ServiceError
}

type ClusterImageRegistryServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewClusterImageRegistryService(spec ClusterImageRegistryServiceSpec) ClusterImageRegistryService {
	return &clusterImageRegistryService{
		clusterImageRegistryStore: pgstore.NewClusterImageRegistryStore(pgstore.ClusterImageRegistryStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type clusterImageRegistryService struct {
	clusterImageRegistryStore stores.ClusterImageRegistryStore
	clusterResourceService    clusterresource.ClusterImageRegistryService
	logger                    logger.Logger
}

func (s *clusterImageRegistryService) InjectClusterResourceService(registryClusterService clusterresource.ClusterImageRegistryService) {
	s.clusterResourceService = registryClusterService
}

func (s *clusterImageRegistryService) Get(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	registry, err := s.clusterImageRegistryStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster image registry: %v", err)
		return nil, err
	}
	return registry, nil
}

func (s *clusterImageRegistryService) ListByClusterID(ctx context.Context, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError) {
	registries, err := s.clusterImageRegistryStore.ListByClusterID(ctx, clusterID)
	if err != nil {
		s.logger.Errorf("failed to list cluster image registries: %v", err)
		return nil, err
	}
	return registries, nil
}

func (s *clusterImageRegistryService) Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	var createdRegistry *models.ClusterImageRegistry
	var err *errors.ServiceError

	// Initialize status if not set
	if spec.Status == nil {
		spec.Status = &models.ClusterImageRegistryStatus{
			State:      models.RegistryStatePending,
			Conditions: []models.Condition{},
		}
	}

	if spec.ClusterID == "" {
		return nil, errors.GeneralError("cluster ID is required")
	}
	if spec.OrganisationID == "" {
		return nil, errors.GeneralError("organisation ID is required")
	}
	if spec.Name == "" {
		return nil, errors.GeneralError("name is required")
	}

	if err := s.validateBackendStorageSize(spec.BackendStorageSize); err != nil {
		return nil, err
	}
	s.setDefaultValues(spec)

	createErr := s.clusterImageRegistryStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Create registry in database
		createdRegistry, err = s.clusterImageRegistryStore.CreateWithTx(ctx, spec)
		if err != nil {
			s.logger.Errorf("failed to create cluster image registry: %v", err)
			return err
		}

		// Create registry in cluster
		cerr := s.clusterResourceService.CreateImageRegistryInCluster(ctx, createdRegistry)
		if cerr != nil {
			s.logger.Errorf("failed to create cluster image registry in cluster: %v", cerr)
			return errors.GeneralError("failed to create cluster image registry in cluster: %s", cerr.Error())
		}
		return nil
	})

	if createErr != nil {
		return nil, createErr
	}
	return createdRegistry, nil
}

func (s *clusterImageRegistryService) validateBackendStorageSize(size string) *errors.ServiceError {
	if size == "" {
		return nil
	}
	if _, err := k8sresource.ParseQuantity(size); err != nil {
		return errors.Validation("backend storage size is not a valid quantity")
	}
	return nil
}

func (s *clusterImageRegistryService) UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError {
	err := s.clusterImageRegistryStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		s.logger.Errorf("failed to update cluster image registry status: %v", err)
		return err
	}
	return nil
}

func (s *clusterImageRegistryService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	registry, err := s.clusterImageRegistryStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get cluster image registry for deletion: %v", err)
		return err
	}

	deleteErr := s.clusterImageRegistryStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Delete from cluster first
		cErr := s.clusterResourceService.DeleteImageRegistryInCluster(ctx, registry)
		if cErr != nil {
			s.logger.Errorf("failed to delete cluster image registry in cluster: %v", cErr)
			return errors.GeneralError("failed to delete cluster image registry in cluster: %s", cErr.Error())
		}

		// Then delete from database
		err = s.clusterImageRegistryStore.DeleteWithTx(ctx, ID)
		if err != nil {
			s.logger.Errorf("failed to delete cluster image registry: %v", err)
			return err
		}
		return nil
	})

	if deleteErr != nil {
		s.logger.Errorf("failed to delete cluster image registry: %v", deleteErr)
		return deleteErr
	}
	return nil
}

func (s *clusterImageRegistryService) setDefaultValues(registry *models.ClusterImageRegistry) {
	if registry.MaxRepositories == 0 {
		registry.MaxRepositories = 10
	}
	if registry.TagsPerRepository == 0 {
		registry.TagsPerRepository = 5
	}
	if !registry.DeleteUntagged {
		registry.DeleteUntagged = true
	}
	if registry.BackendStorageSize == "" {
		registry.BackendStorageSize = "50Gi"
	}
}
