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
	"github.com/ashishmax31/stackdome-api-server/pkg/validation"
)

type StackService interface {
	CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError)
	GetStacksByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError)
	GetStacksByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError)
	UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError
	DeleteStack(ctx context.Context, ID string) *errors.ServiceError
	InjectClusterResourceService(workspaceClusterService clusterresource.ClusterWorkspaceService)
}

type StackServiceSpec struct {
	SessionFactory       db.SessionFactory
	WorkspaceUserService WorkspaceUserService
	VolumeService        VolumeService
	ClusterService       ClusterService
	OrganisationService  OrganisationService
	Logger               logger.Logger
}

type stackService struct {
	stackStore             stores.StackStore
	logger                 logger.Logger
	sessionFactory         db.SessionFactory
	workspaceUserService   WorkspaceUserService
	clusterResourceService clusterresource.ClusterWorkspaceService
	volumeService          VolumeService
	organisationService    OrganisationService
	interpolationValidator validation.InterpolationValidation
}

func NewStackService(spec StackServiceSpec) StackService {
	return &stackService{
		stackStore: pgstore.NewStackStore(&pgstore.StackStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceUserService:   spec.WorkspaceUserService,
		volumeService:          spec.VolumeService,
		organisationService:    spec.OrganisationService,
		logger:                 spec.Logger,
		sessionFactory:         spec.SessionFactory,
		interpolationValidator: validation.NewInterpolationValidation(),
	}
}

func (s *stackService) InjectClusterResourceService(workspaceClusterService clusterresource.ClusterWorkspaceService) {
	s.clusterResourceService = workspaceClusterService
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	workspaceUser, err := s.workspaceUserService.GetWorkspaceUser(ctx, spec.UserID)
	if err != nil {
		return nil, errors.GeneralError("failed to get workspaceuser for user '%s': %s", spec.UserID, err.Error())
	}
	userWorkspaceNamespaceMap := workspaceUser.WorkspaceNamespaceMap()
	if workpaceNamepace, ok := userWorkspaceNamespaceMap[spec.WorkspaceName]; ok {
		spec.Namespace = workpaceNamepace.Namespace
	} else {
		return nil, errors.BadRequest("workspace with name '%s' does not exist for user", spec.Name)
	}

	s.setDefaultValues(spec)

	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		spec.StackResources[i].Namespace = spec.Namespace
	}

	organisation, err := s.organisationService.Get(ctx, spec.OrganisationID)
	if err != nil {
		return nil, errors.GeneralError("failed to get organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	if len(organisation.DomainName) == 0 {
		return nil, errors.BadRequest("domain name is not defined for organisation '%s'", spec.OrganisationID)
	}

	if spec.HasVolumeMounts() {
		volumes, err := s.volumeService.InternalList(ctx, spec.VolumeMountIds())
		if err != nil {
			return nil, errors.GeneralError("failed to get volumes '%s': %s", spec.Name, err.Error())
		}
		if err := s.validateVolumeMounts(volumes, spec); err != nil {
			return nil, err
		}
		setVolumeMountType(volumes, spec)
	}

	if err := s.populateAssociations(ctx, spec); err != nil {
		return nil, errors.GeneralError("failed to populate associations for stack '%s': %s", spec.Name, err.Error())
	}

	if err := s.validateImageToRun(spec); err != nil {
		return nil, err
	}

	if err := s.validateStackEnvVars(spec); err != nil {
		return nil, err
	}
	if err := s.validateStackPorts(spec, organisation); err != nil {
		return nil, err
	}

	var createdStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var creatErr *errors.ServiceError
		createdStack, creatErr = s.stackStore.CreateWithTx(ctx, spec)
		if creatErr != nil {
			return creatErr
		}
		if err := s.clusterResourceService.CreateStackInCluster(ctx, createdStack); err != nil {
			return errors.GeneralError("failed to create stack in cluster: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return createdStack, nil
}

func (s *stackService) setDefaultValues(spec *models.Stack) {
	for i := range spec.StackResources {
		if len(spec.StackResources[i].Ports) > 0 {
			for j := range spec.StackResources[i].Ports {
				// if the port is exposed to public, set the protocol to http
				if spec.StackResources[i].Ports[j].ExposedToPublic {
					spec.StackResources[i].Ports[j].Protocol = "http"
				}
			}
		}
	}
}

func (s *stackService) GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	workspace, err := s.stackStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

func (s *stackService) GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError) {
	stack, err := s.stackStore.GetByName(ctx, name, userID)
	if err != nil {
		return nil, err
	}
	return stack, nil
}

func (s *stackService) GetStacksByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError) {
	stacks, err := s.stackStore.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (s *stackService) GetStacksByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError) {
	stacks, err := s.stackStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (s *stackService) DeleteStack(ctx context.Context, ID string) *errors.ServiceError {
	stack, err := s.GetStack(ctx, ID)
	if err != nil {
		return err
	}
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		if err := s.clusterResourceService.DeleteStackInCluster(ctx, stack); err != nil {
			return errors.GeneralError("failed to delete stack in cluster: %s", err.Error())
		}
		if err := s.stackStore.DeleteWithTx(ctx, ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *stackService) UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if spec.Name != stack.Name {
		return nil, errors.BadRequest("workspace name cannot be updated")
	}
	if spec.UserID != stack.UserID {
		return nil, errors.BadRequest("workspace user cannot be updated")
	}
	if spec.OrganisationID != stack.OrganisationID {
		return nil, errors.BadRequest("workspace organisation cannot be updated")
	}
	spec.Namespace = stack.Namespace

	if spec.HasVolumeMounts() {
		volumes, err := s.volumeService.InternalList(ctx, spec.VolumeMountIds())
		if err != nil {
			return nil, errors.GeneralError("failed to get volumes '%s': %s", spec.Name, err.Error())
		}
		if err := s.validateVolumeMounts(volumes, spec); err != nil {
			return nil, err
		}
		setVolumeMountType(volumes, spec)
	}

	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var updateErr *errors.ServiceError
		updatedStack, updateErr = s.stackStore.UpdateWithTx(ctx, ID, spec)
		if updateErr != nil {
			return updateErr
		}
		if err := s.clusterResourceService.UpdateStackInCluster(ctx, updatedStack); err != nil {
			return errors.GeneralError("failed to update workspace in cluster: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedStack, nil
}

func (s *stackService) UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError {
	err := s.stackStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		return err
	}
	return nil
}

func (s *stackService) validateVolumeMounts(existingVolumes []*models.Volume, spec *models.Stack) *errors.ServiceError {
	existingVolumeMap := make(map[string]*models.Volume)
	for i := range existingVolumes {
		existingVolumeMap[existingVolumes[i].ID] = existingVolumes[i]
	}

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		for j := range spec.StackResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if _, found := existingVolumeMap[currentVolumeMount.SourceVolumeID]; !found {
				return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeID)
			}
		}
	}
	return nil
}

func (s *stackService) validateStackEnvVars(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.ExecutionConfig == nil || currentResource.ExecutionConfig.Env == nil {
			continue
		}
		currentEnvVars := currentResource.ExecutionConfig.Env

		for _, envVar := range currentEnvVars {
			if len(envVar.Name) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var name", currentResource.Name)
			}
			if len(envVar.Value) == 0 {
				return errors.BadRequest("stack resource '%s' has empty env var value", currentResource.Name)
			}
		}
	}

	if err := s.interpolationValidator.ValidateStackInterpolations(spec); err != nil {
		return errors.BadRequest("stack resource '%s' has invalid interpolation: %s", spec.Name, err.Error())
	}

	return nil
}

func (s *stackService) validateStackPorts(spec *models.Stack, org *models.Organisation) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if len(currentResource.Ports) == 0 {
			continue
		}
		currentPorts := currentResource.Ports

		for _, port := range currentPorts {
			if port.Number <= 0 {
				return errors.BadRequest("stack resource '%s' has invalid port number", currentResource.Name)
			}
		}
	}
	return nil
}

func (s *stackService) validateImageToRun(spec *models.Stack) *errors.ServiceError {
	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		if currentResource.BuildConfig != nil && currentResource.ImageConfig != nil {
			return errors.BadRequest("stack resource '%s' cannot have both build and image config", currentResource.Name)
		}
		if currentResource.ImageConfig == nil && currentResource.BuildConfig == nil {
			return errors.BadRequest("stack resource '%s' must have either build or image config", currentResource.Name)
		}
		if currentResource.ImageConfig != nil {
			if err := currentResource.ImageConfig.Validate(); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid image config: %s", currentResource.Name, err.Error())
			}
		}

		if currentResource.BuildConfig != nil {
			if err := currentResource.BuildConfig.Validate(); err != nil {
				return errors.BadRequest("stack resource '%s' has invalid build config: %s", currentResource.Name, err.Error())
			}
		}
	}
	return nil
}

func (s *stackService) populateAssociations(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if spec.StackResources == nil {
		return nil
	}
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID

		if spec.StackResources[i].BuildConfig != nil {
			buildConfig := spec.StackResources[i].BuildConfig
			if buildConfig.SourceContext.Volume != nil {
				volume, err := s.volumeService.Get(ctx, buildConfig.SourceContext.Volume.SourceVolumeID)
				if err != nil {
					return errors.GeneralError(
						"failed to get volume specified in the build source '%s': %s",
						buildConfig.SourceContext.Volume.SourceVolumeID, err.Error(),
					)
				}
				buildConfig.SourceContext.Volume.SourceVolumeName = volume.Name
			}
		}
	}

	return nil
}

func setVolumeMountType(existingVolumes []*models.Volume, spec *models.Stack) {
	existingVolumeMap := make(map[string]*models.Volume)
	for i := range existingVolumes {
		existingVolumeMap[existingVolumes[i].ID] = existingVolumes[i]
	}

	for i := range spec.StackResources {
		for j := range spec.StackResources[i].VolumeMounts {
			if volume, found := existingVolumeMap[spec.StackResources[i].VolumeMounts[j].SourceVolumeID]; found {
				spec.StackResources[i].VolumeMounts[j].SourceVolumeType = volume.VolumeSourceType()
			}
		}
	}
}
