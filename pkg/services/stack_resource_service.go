package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/computequota"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	stackresourcevalidator "github.com/Stackdome/stackdome/pkg/validator/stackresource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

//go:generate mockgen -destination=../mocks/mock_stack_resource_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services StackResourceService

type StackResourceService interface {
	InjectClusterManager(clusterManager clustermanager.ClusterManager)
	Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError)
	GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError)
	InternalGetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError)
	Update(ctx context.Context, stackID, resourceName string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	Delete(ctx context.Context, stackID, resourceName string) *errors.ServiceError
	UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError
	InternalCreateWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	InternalUpdateWithTx(ctx context.Context, stack *models.Stack, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)
	InternalDeleteWithTx(ctx context.Context, resourceID string) *errors.ServiceError
	InternalSyncResourcesWithTx(ctx context.Context, stack *models.Stack, existingStack *models.Stack, desired []*models.StackResource) *errors.ServiceError
	Restart(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError)
}

type StackResourceServiceSpec struct {
	SessionFactory         db.SessionFactory
	StorageService         StackStorageService
	Logger                 logger.Logger
	Permissions            auth.PermissionService
	StackStore             stores.StackStore
	StackResourceStore     stores.StackResourceStore
	ClusterRegistryService ImageRegistryService
	StackDomainService     StackDomainsService
	ReferenceService       ReferenceService
	ResourceValidator      stackresourcevalidator.Validator
	ComputePolicy          computequota.Policy
}

type stackResourceService struct {
	stackResourceStore     stores.StackResourceStore
	stackStore             stores.StackStore
	clusterManager         clustermanager.ClusterManager
	logger                 logger.Logger
	sessionFactory         db.SessionFactory
	storageService         StackStorageService
	permissions            auth.PermissionService
	clusterRegistryService ImageRegistryService
	domainNameService      StackDomainsService
	referenceService       ReferenceService
	resourceValidator      stackresourcevalidator.Validator
	computePolicy          computequota.Policy
}

func NewStackResourceService(spec StackResourceServiceSpec) StackResourceService {
	if spec.ResourceValidator == nil {
		panic("services.NewStackResourceService: ResourceValidator is required")
	}
	stackResourceStore := spec.StackResourceStore
	if stackResourceStore == nil {
		stackResourceStore = pgstore.NewStackResourceStore(pgstore.StackResourceStoreSpec{
			SessionFactory: spec.SessionFactory,
		})
	}
	return &stackResourceService{
		stackResourceStore:     stackResourceStore,
		stackStore:             spec.StackStore,
		storageService:         spec.StorageService,
		logger:                 spec.Logger,
		sessionFactory:         spec.SessionFactory,
		permissions:            spec.Permissions,
		clusterRegistryService: spec.ClusterRegistryService,
		domainNameService:      spec.StackDomainService,
		referenceService:       spec.ReferenceService,
		resourceValidator:      spec.ResourceValidator,
		computePolicy:          spec.ComputePolicy,
	}
}

func (s *stackResourceService) InjectClusterManager(clusterManager clustermanager.ClusterManager) {
	s.clusterManager = clusterManager
}

func (s *stackResourceService) Create(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, resource.StackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, resource.StackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	s.computePolicy.ApplyStackResourceDefaults(resource)
	siblings, sErr := s.stackResourceStore.GetByStackID(ctx, resource.StackID)
	if sErr != nil {
		return nil, sErr
	}
	ferrs, vErr := s.resourceValidator.Validate(ctx, stack, resource, siblings)
	if vErr != nil {
		return nil, vErr
	}
	if len(ferrs) > 0 {
		return nil, errors.ValidationFailed(ferrs)
	}

	var created *models.StackResource
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if accessErr := s.computePolicy.EnsureAccess(txCtx, stack.OrganisationID); accessErr != nil {
			return accessErr
		}
		s.computePolicy.ApplyStackResourceDefaults(resource)
		if limitErr := s.computePolicy.ValidateStackLimits(txCtx, computequota.StackLimitChange{
			Operation:      computequota.StackLimitAddResource,
			OrganisationID: stack.OrganisationID,
		}); limitErr != nil {
			return limitErr
		}
		var createErr *errors.ServiceError
		created, createErr = s.InternalCreateWithTx(txCtx, stack, resource)
		if createErr != nil {
			return createErr
		}
		return s.referenceService.ReprojectSpec(txCtx, stack.ID)
	}); err != nil {
		return nil, err
	}

	return s.stackResourceStore.GetByID(ctx, created.ID)
}

func (s *stackResourceService) Update(ctx context.Context, stackID, resourceName string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	existing, err := s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
	if err != nil {
		return nil, err
	}

	resource.StackID = stackID
	resource.ID = existing.ID
	resource.Name = resourceName

	s.computePolicy.ApplyStackResourceDefaults(resource)
	all, sErr := s.stackResourceStore.GetByStackID(ctx, stackID)
	if sErr != nil {
		return nil, sErr
	}
	siblings := make([]*models.StackResource, 0, len(all))
	for _, r := range all {
		if r.Name != resourceName {
			siblings = append(siblings, r)
		}
	}
	ferrs, vErr := s.resourceValidator.Validate(ctx, stack, resource, siblings)
	if vErr != nil {
		return nil, vErr
	}
	if len(ferrs) > 0 {
		return nil, errors.ValidationFailed(ferrs)
	}

	var updated *models.StackResource
	if txErr := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if accessErr := s.computePolicy.EnsureAccess(txCtx, stack.OrganisationID); accessErr != nil {
			return accessErr
		}
		s.computePolicy.ApplyStackResourceDefaults(resource)
		if limitErr := s.computePolicy.ValidateStackLimits(txCtx, computequota.StackLimitChange{
			Operation:      computequota.StackLimitUpdateResource,
			OrganisationID: stack.OrganisationID,
		}); limitErr != nil {
			return limitErr
		}
		var updateErr *errors.ServiceError
		updated, updateErr = s.InternalUpdateWithTx(txCtx, stack, existing.ID, resource)
		if updateErr != nil {
			return updateErr
		}
		return s.referenceService.ReprojectSpec(txCtx, stackID)
	}); txErr != nil {
		return nil, txErr
	}

	return s.stackResourceStore.GetByID(ctx, updated.ID)
}

// prepareResource applies defaults and external lookups before persisting a resource.
func (s *stackResourceService) prepareResource(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	resource.StackID = stack.ID
	resource.UserID = stack.UserID
	resource.Namespace = stack.Namespace
	applyStatefulWorkloadDefault(resource)
	applyStackResourcePortDefaults(resource)
	normalizeStackResourceReplicas(resource)
	return s.populateRegistryUrlForResource(ctx, stack, resource)
}

func (s *stackResourceService) populateRegistryUrlForResource(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	if resource.BuildConfig == nil || !resource.BuildConfig.BuildImageRepository.UseInClusterRegistry {
		return nil
	}
	if s.clusterRegistryService == nil {
		return errors.GeneralError("cluster registry service is not configured")
	}
	return s.clusterRegistryService.PopulateInClusterRegistryNameForResource(
		ctx, stack.OrganisationID, stack.ClusterID, stack.Name, resource)
}

func (s *stackResourceService) InternalCreateWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	if err := s.prepareResource(ctx, stack, resource); err != nil {
		return nil, err
	}
	created, err := s.stackResourceStore.CreateWithTx(ctx, resource, stack)
	if err != nil {
		return nil, err
	}
	return s.finalizeExposedPortsForResourceWithTx(ctx, stack, created)
}

func (s *stackResourceService) InternalUpdateWithTx(ctx context.Context, stack *models.Stack, resourceID string, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	resource.ID = resourceID
	if err := s.prepareResource(ctx, stack, resource); err != nil {
		return nil, err
	}
	updated, err := s.stackResourceStore.UpdateWithTx(ctx, resourceID, resource, stack)
	if err != nil {
		return nil, err
	}
	return s.finalizeExposedPortsForResourceWithTx(ctx, stack, updated)
}

func (s *stackResourceService) InternalDeleteWithTx(ctx context.Context, resourceID string) *errors.ServiceError {
	if s.domainNameService != nil {
		if err := s.domainNameService.InternalDeleteDomainsForResourceWithTx(ctx, resourceID); err != nil {
			return err
		}
	}
	return s.stackResourceStore.DeleteWithTx(ctx, resourceID)
}

func (s *stackResourceService) InternalSyncResourcesWithTx(ctx context.Context, stack *models.Stack, existingStack *models.Stack, desired []*models.StackResource) *errors.ServiceError {
	existingMap := existingStack.ResourcesMap()
	desiredMap := make(map[string]*models.StackResource, len(desired))

	for _, resource := range desired {
		desiredMap[resource.Name] = resource
		if existing, ok := existingMap[resource.Name]; ok {
			resource.Name = existing.Name
			if _, err := s.InternalUpdateWithTx(ctx, stack, existing.ID, resource); err != nil {
				return err
			}
		} else {
			if _, err := s.InternalCreateWithTx(ctx, stack, resource); err != nil {
				return err
			}
		}
	}

	for _, existing := range existingStack.StackResources {
		if _, ok := desiredMap[existing.Name]; !ok {
			if err := s.InternalDeleteWithTx(ctx, existing.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *stackResourceService) finalizeExposedPortsForResourceWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) (*models.StackResource, *errors.ServiceError) {
	if err := s.domainNameService.PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, resource); err != nil {
		return nil, err
	}
	if err := s.stackResourceStore.UpdatePortsWithTx(ctx, resource.ID, resource); err != nil {
		return nil, err
	}
	return resource, nil
}

func (s *stackResourceService) Delete(ctx context.Context, stackID, resourceName string) *errors.ServiceError {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	existing, err := s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
	if err != nil {
		return err
	}

	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if err := s.InternalDeleteWithTx(txCtx, existing.ID); err != nil {
			return err
		}
		return s.referenceService.ReprojectSpec(txCtx, stackID)
	}); err != nil {
		return err
	}

	return nil
}

func (s *stackResourceService) GetByStackID(ctx context.Context, stackID string) ([]*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.stackResourceStore.GetByStackID(ctx, stackID)
}

func (s *stackResourceService) GetByID(ctx context.Context, ID string) (*models.StackResource, *errors.ServiceError) {
	resource, err := s.stackResourceStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	stack, stackErr := s.stackStore.GetByID(ctx, resource.StackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, resource.StackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return resource, nil
}

func (s *stackResourceService) GetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
}

func (s *stackResourceService) InternalGetByStackIDAndResourceName(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	return s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
}

func (s *stackResourceService) UpdateStatus(ctx context.Context, resourceID string, status *models.StackResourceStatus) *errors.ServiceError {
	return s.stackResourceStore.UpdateStatus(ctx, resourceID, status)
}

func (s *stackResourceService) Restart(ctx context.Context, stackID, resourceName string) (*models.StackResource, *errors.ServiceError) {
	stack, stackErr := s.stackStore.GetByID(ctx, stackID)
	if stackErr != nil {
		return nil, stackErr
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	resource, err := s.stackResourceStore.GetByStackIDAndResourceName(ctx, stackID, resourceName)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	var updated *models.StackResource
	if txErr := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if accessErr := s.computePolicy.EnsureAccess(txCtx, stack.OrganisationID); accessErr != nil {
			return accessErr
		}
		if resource.LifecycleConfig == nil {
			resource.LifecycleConfig = &models.LifecycleConfig{}
		}
		resource.LifecycleConfig.RestartRequestTime = &now

		var updateErr *errors.ServiceError
		updated, updateErr = s.stackResourceStore.Update(txCtx, resource.ID, resource, stack)
		return updateErr
	}); txErr != nil {
		return nil, txErr
	}

	// Patch the StackResource CR in the cluster directly with the new restart timestamp.
	// Best-effort: if the cluster is unreachable, the DB is already updated and the
	// next release apply will set the correct restart time on the CR.
	if s.clusterManager != nil {
		s.patchRestartRequestInCluster(ctx, stack, resource, now)
	}

	return updated, nil
}

func (s *stackResourceService) patchRestartRequestInCluster(ctx context.Context, stack *models.Stack, resource *models.StackResource, restartTime time.Time) {
	clusterClient, err := s.clusterManager.GetClient(stack.ClusterID)
	if err != nil {
		s.logger.Warn(ctx, "restart: failed to get cluster client for cluster '%s': %v", stack.ClusterID, err)
		return
	}

	existing := &corev1alpha1.StackResource{}
	if err := clusterClient.Get(ctx, client.ObjectKey{
		Name:      resource.Name,
		Namespace: resource.Namespace,
	}, existing); err != nil {
		s.logger.Warn(ctx, "restart: failed to get StackResource CR '%s/%s': %v", resource.Namespace, resource.Name, err)
		return
	}

	existing.Spec.RestartRequest = &metav1.Time{Time: restartTime}
	if err := clusterClient.Update(ctx, existing); err != nil {
		s.logger.Warn(ctx, "restart: failed to update StackResource CR '%s/%s': %v", resource.Namespace, resource.Name, err)
	}
}
