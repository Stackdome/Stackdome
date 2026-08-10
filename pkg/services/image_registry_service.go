package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

//go:generate mockgen -destination=../mocks/mock_image_registry_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ImageRegistryService

type ImageRegistryService interface {
	BackgroundJobEnqueuerInjectable
	Get(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalGetForOrgAndCluster(ctx context.Context, orgID, clusterID string) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalMarkAllDeletingByClusterIDWithTx(ctx context.Context, clusterID string) *errors.ServiceError
	ListForOrg(ctx context.Context, orgID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	ListByClusterID(ctx context.Context, orgID, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError)
	Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	InternalCreatePendingSeedRegistry(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError)
	DeleteWithTx(ctx context.Context, orgID string, registry *models.ClusterImageRegistry) *errors.ServiceError
	UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError
	PopulateInClusterRegistryNameForResource(ctx context.Context, orgID, clusterID, stackName string, resource *models.StackResource) *errors.ServiceError
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
	permissions               auth.PermissionService
	logger                    logger.Logger
	BackgroundJobEnqueuerDep
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

func (s *clusterImageRegistryService) InternalGetForOrgAndCluster(ctx context.Context, orgID, clusterID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	return s.clusterImageRegistryStore.GetForOrgAndCluster(ctx, orgID, clusterID)
}

func (s *clusterImageRegistryService) InternalMarkAllDeletingByClusterIDWithTx(ctx context.Context, clusterID string) *errors.ServiceError {
	return s.clusterImageRegistryStore.MarkAllDeletingByClusterIDWithTx(ctx, clusterID)
}

// validateOwnedCluster requires the target cluster to be owned by the org.
// Seed registries on the shared-compute cluster go through
// InternalCreatePendingSeedRegistry instead.
func (s *clusterImageRegistryService) validateOwnedCluster(ctx context.Context, orgID, clusterID string) *errors.ServiceError {
	owned, err := s.clusterStore.ListBYOCClustersForOrg(ctx, orgID)
	if err != nil {
		return err
	}
	for _, cluster := range owned {
		if cluster.ID == clusterID {
			return nil
		}
	}
	return errors.NotFound("cluster '%s' not found for organisation '%s'", clusterID, orgID)
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

// PopulateInClusterRegistryNameForResource resolves the registry on the
// stack's cluster: builds must push where the workload pulls.
func (s *clusterImageRegistryService) PopulateInClusterRegistryNameForResource(ctx context.Context, orgID, clusterID, stackName string, resource *models.StackResource) *errors.ServiceError {
	if resource.BuildConfig == nil || !resource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
		return nil
	}

	clusterRegistry, err := s.clusterImageRegistryStore.GetForOrgAndCluster(ctx, orgID, clusterID)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return errors.BadRequest("no cluster registry found for organisation '%s' on cluster '%s'", orgID, clusterID)
		}
		return errors.GeneralError("failed to get cluster registry for organisation '%s': %s", orgID, err.Error())
	}

	resource.BuildConfig.BuildImageRepository.ClusterRegistryName = clusterRegistry.Name
	return nil
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

// InternalCreatePendingSeedRegistry records the org's shared-compute registry
// without provisioning it. A release creates the cluster resource on first use.
func (s *clusterImageRegistryService) InternalCreatePendingSeedRegistry(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	if err := s.validateSpec(spec); err != nil {
		return nil, err
	}
	cluster, err := s.clusterStore.Get(ctx, spec.ClusterID)
	if err != nil {
		return nil, err
	}
	if !cluster.SharedCompute {
		return nil, errors.NotFound("cluster '%s' is not a shared-compute cluster", spec.ClusterID)
	}
	if spec.Status == nil {
		spec.Status = &models.ClusterImageRegistryStatus{State: models.RegistryStatePending, Conditions: []models.Condition{}}
	}
	s.setDefaultValues(spec)
	var created *models.ClusterImageRegistry
	createErr := s.clusterImageRegistryStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var createErr *errors.ServiceError
		created, createErr = s.clusterImageRegistryStore.CreateWithTx(txCtx, spec)
		return createErr
	})
	return created, createErr
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

		return s.enqueueAfterCommit(ctx, createdRegistry.ClusterID)
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

	if enqueueErr := s.enqueueAfterCommit(ctx, createdRegistry.ClusterID); enqueueErr != nil {
		return nil, enqueueErr
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
		return s.DeleteWithTx(ctx, orgID, registry)
	})

	if deleteErr != nil {
		s.logger.Error(ctx, "failed to delete cluster image registry: %v", deleteErr)
		return deleteErr
	}
	return nil
}

// DeleteWithTx records durable deletion intent. Kubernetes cleanup and the
// final database delete are performed asynchronously by the registry worker.
func (s *clusterImageRegistryService) DeleteWithTx(ctx context.Context, orgID string, registry *models.ClusterImageRegistry) *errors.ServiceError {
	if registry == nil {
		return errors.GeneralError("image registry is nil")
	}
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceImageRegistries, registry.ID, auth.ActionDelete); permErr != nil {
		return permErr
	}

	if err := s.clusterImageRegistryStore.MarkDeletingWithTx(ctx, registry.ID); err != nil {
		s.logger.Error(ctx, "failed to mark cluster image registry for deletion: %v", err)
		return err
	}
	return s.enqueueAfterCommit(ctx, registry.ClusterID)
}

func (s *clusterImageRegistryService) enqueueAfterCommit(ctx context.Context, clusterID string) *errors.ServiceError {
	if s.BackgroundJobEnqueuer == nil {
		return errors.GeneralError("background job enqueuer is not configured")
	}
	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, models.ClusterImageRegistryOperand{ClusterID: clusterID}); err != nil {
		return errors.GeneralError("failed to enqueue image registry reconciliation: %s", err.Error())
	}
	return nil
}

func (s *clusterImageRegistryService) setDefaultValues(registry *models.ClusterImageRegistry) {
	if registry.BackendStorageSize == "" {
		registry.BackendStorageSize = models.DefaultRegistryStorageSize
	}
}
