package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

//go:generate mockgen -destination=../mocks/mock_image_registry_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ImageRegistryService

type ImageRegistryService interface {
	Get(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	GetForOrg(ctx context.Context, orgID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	ListForOrg(ctx context.Context, orgID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterID(ctx context.Context, orgID, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalCreateSeedRegistry(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError
	InjectClusterResourceService(registryClusterService clusterresource.ClusterImageRegistryService)
	PopulateInClusterRegistryUrlForResource(ctx context.Context, orgID, stackName string, resource *models.StackResource) *errors.ServiceError
	Delete(ctx context.Context, orgID, ID string) *errors.ServiceError
}

type ImageRegistryServiceSpec struct {
	SessionFactory db.SessionFactory
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

func NewClusterImageRegistryService(spec ImageRegistryServiceSpec) ImageRegistryService {
	return &clusterImageRegistryService{
		clusterImageRegistryStore: pgstore.NewClusterImageRegistryStore(pgstore.ClusterImageRegistryStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		clusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger:      spec.Logger,
		permissions: spec.Permissions,
	}
}

type clusterImageRegistryService struct {
	clusterImageRegistryStore stores.ClusterImageRegistryStore
	clusterStore              stores.ClusterStore
	clusterResourceService    clusterresource.ClusterImageRegistryService
	permissions               auth.PermissionService
	logger                    logger.Logger
}

func (s *clusterImageRegistryService) InjectClusterResourceService(registryClusterService clusterresource.ClusterImageRegistryService) {
	s.clusterResourceService = registryClusterService
}

func (s *clusterImageRegistryService) Get(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("failed to fetch identity")
	}
	if permErr := s.permissions.Check(ctx, identity.OrgID, auth.ResourceImageRegistries, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	registry, err := s.clusterImageRegistryStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster image registry: %v", err)
		return nil, err
	}
	return registry, nil
}

func (s *clusterImageRegistryService) InternalGet(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	return s.clusterImageRegistryStore.GetByID(ctx, ID)
}

// validateOwnedCluster requires the target cluster to be owned by the org.
// Seed registries on the shared platform cluster go through
// InternalCreateSeedRegistry instead.
func (s *clusterImageRegistryService) validateOwnedCluster(ctx context.Context, orgID, clusterID string) *errors.ServiceError {
	owned, err := s.clusterStore.GetClusterForOrg(ctx, orgID)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return errors.NotFound("cluster '%s' not found for organisation '%s'", clusterID, orgID)
		}
		return err
	}
	if owned.ID != clusterID {
		return errors.NotFound("cluster '%s' not found for organisation '%s'", clusterID, orgID)
	}
	return nil
}

func (s *clusterImageRegistryService) ListForOrg(ctx context.Context, orgID string) ([]*models.ClusterImageRegistry, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceImageRegistries, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	registries, err := s.clusterImageRegistryStore.ListForOrg(ctx, orgID)
	if err != nil {
		s.logger.Error(ctx, "failed to list cluster image registries for org: %v", err)
		return nil, err
	}
	return registries, nil
}

func (s *clusterImageRegistryService) ListByClusterID(ctx context.Context, orgID, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceImageRegistries, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	registries, err := s.clusterImageRegistryStore.ListByClusterID(ctx, orgID, clusterID)
	if err != nil {
		s.logger.Error(ctx, "failed to list cluster image registries: %v", err)
		return nil, err
	}
	return registries, nil
}

func (s *clusterImageRegistryService) PopulateInClusterRegistryUrlForResource(ctx context.Context, orgID, stackName string, resource *models.StackResource) *errors.ServiceError {
	if resource.BuildConfig == nil || !resource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
		return nil
	}

	clusterRegistry, err := s.GetForOrg(ctx, orgID)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return errors.BadRequest("no cluster registry found for organisation '%s'", orgID)
		}
		return errors.GeneralError("failed to get cluster registry for organisation '%s': %s", orgID, err.Error())
	}

	resource.BuildConfig.BuildImageRepository.ClusterRegistryName = clusterRegistry.Name
	return nil
}

func (s *clusterImageRegistryService) GetForOrg(ctx context.Context, orgID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	registry, err := s.clusterImageRegistryStore.GetForOrg(ctx, orgID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster image registry for org: %v", err)
		return nil, err
	}
	return registry, nil
}

func (s *clusterImageRegistryService) Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, spec.OrganisationID, auth.ResourceImageRegistries, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	if err := s.validateSpec(spec); err != nil {
		return nil, err
	}
	if serr := s.validateOwnedCluster(ctx, spec.OrganisationID, spec.ClusterID); serr != nil {
		return nil, serr
	}
	return s.create(ctx, spec)
}

// InternalCreateSeedRegistry creates the org's seed registry on the shared
// platform cluster. Org-provisioning only — the API Create path requires an
// org-owned target cluster.
func (s *clusterImageRegistryService) InternalCreateSeedRegistry(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	if err := s.validateSpec(spec); err != nil {
		return nil, err
	}
	platform, err := s.clusterStore.GetPlatformCluster(ctx)
	if err != nil {
		return nil, err
	}
	if platform.ID != spec.ClusterID {
		return nil, errors.NotFound("cluster '%s' is not the platform cluster", spec.ClusterID)
	}
	return s.create(ctx, spec)
}

func (s *clusterImageRegistryService) validateSpec(spec *models.ClusterImageRegistry) *errors.ServiceError {
	if spec.ClusterID == "" {
		return errors.GeneralError("cluster ID is required")
	}
	if spec.OrganisationID == "" {
		return errors.GeneralError("organisation ID is required")
	}
	if spec.Name == "" {
		return errors.GeneralError("name is required")
	}
	return s.validateBackendStorageSize(spec.BackendStorageSize)
}

func (s *clusterImageRegistryService) create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	var createdRegistry *models.ClusterImageRegistry
	var err *errors.ServiceError

	// Initialize status if not set
	if spec.Status == nil {
		spec.Status = &models.ClusterImageRegistryStatus{
			State:      models.RegistryStatePending,
			Conditions: []models.Condition{},
		}
	}
	s.setDefaultValues(spec)

	createErr := s.clusterImageRegistryStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Create registry in database
		createdRegistry, err = s.clusterImageRegistryStore.CreateWithTx(ctx, spec)
		if err != nil {
			s.logger.Error(ctx, "failed to create cluster image registry: %v", err)
			return err
		}

		// Create registry in cluster
		cerr := s.clusterResourceService.CreateImageRegistryInCluster(ctx, createdRegistry)
		if cerr != nil {
			s.logger.Error(ctx, "failed to create cluster image registry in cluster: %v", cerr)
			return errors.GeneralError("failed to create cluster image registry in cluster: %s", cerr.Error())
		}
		return nil
	})

	if createErr != nil {
		return nil, createErr
	}
	return createdRegistry, nil
}

func (s *clusterImageRegistryService) CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	var createdRegistry *models.ClusterImageRegistry
	var err *errors.ServiceError

	// Initialize status if not set
	if spec.Status == nil {
		spec.Status = &models.ClusterImageRegistryStatus{
			State:      models.RegistryStatePending,
			Conditions: []models.Condition{},
		}
	}

	if err := s.validateSpec(spec); err != nil {
		return nil, err
	}
	s.setDefaultValues(spec)

	// Create registry in database
	createdRegistry, err = s.clusterImageRegistryStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to create cluster image registry: %v", err)
		return nil, err
	}

	// Create registry in cluster
	cerr := s.clusterResourceService.CreateImageRegistryInCluster(ctx, createdRegistry)
	if cerr != nil {
		s.logger.Error(ctx, "failed to create cluster image registry in cluster: %v", cerr)
		return nil, errors.GeneralError("failed to create cluster image registry in cluster: %s", cerr.Error())
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
		s.logger.Error(ctx, "failed to update cluster image registry status: %v", err)
		return err
	}
	return nil
}

func (s *clusterImageRegistryService) Delete(ctx context.Context, orgID, ID string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceImageRegistries, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}

	registry, err := s.clusterImageRegistryStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get cluster image registry for deletion: %v", err)
		return err
	}

	deleteErr := s.clusterImageRegistryStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Delete from cluster first
		cErr := s.clusterResourceService.DeleteImageRegistryInCluster(ctx, registry)
		if cErr != nil {
			s.logger.Error(ctx, "failed to delete cluster image registry in cluster: %v", cErr)
			return errors.GeneralError("failed to delete cluster image registry in cluster: %s", cErr.Error())
		}

		// Then delete from database
		err = s.clusterImageRegistryStore.DeleteWithTx(ctx, ID)
		if err != nil {
			s.logger.Error(ctx, "failed to delete cluster image registry: %v", err)
			return err
		}
		return nil
	})

	if deleteErr != nil {
		s.logger.Error(ctx, "failed to delete cluster image registry: %v", deleteErr)
		return deleteErr
	}
	return nil
}

func (s *clusterImageRegistryService) setDefaultValues(registry *models.ClusterImageRegistry) {
	if registry.BackendStorageSize == "" {
		registry.BackendStorageSize = models.DefaultRegistryStorageSize
	}
}
