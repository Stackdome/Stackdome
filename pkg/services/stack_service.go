package services

import (
	"context"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validation"
	"github.com/google/uuid"
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
	ClusterResourceServiceInjectable
}

type StackServiceSpec struct {
	SessionFactory         db.SessionFactory
	WorkspaceUserService   WorkspaceUserService
	VolumeService          VolumeService
	ClusterService         ClusterService
	OrganisationService    OrganisationService
	StackResourceService   StackResourceService
	ClusterRegistryService ClusterImageRegistryService
	NamespaceService       NamespaceService
	Logger                 logger.Logger
}

type stackService struct {
	stackStore                      stores.StackStore
	logger                          logger.Logger
	sessionFactory                  db.SessionFactory
	workspaceUserService            WorkspaceUserService
	clusterResourceService          clusterresource.ClusterStackService
	volumeService                   VolumeService
	organisationService             OrganisationService
	interpolationValidator          validation.InterpolationValidation
	domainNameService               DomainsService
	stackResourceService            StackResourceService
	namespaceService                NamespaceService
	namespaceClusterResourceService clusterresource.NamespaceClusterResourceService
	volumeClusterResourceService    clusterresource.VolumeClusterResourceService
	clusterRegistryService          ClusterImageRegistryService
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
		stackResourceService:   spec.StackResourceService,
		clusterRegistryService: spec.ClusterRegistryService,
		namespaceService:       spec.NamespaceService,
		domainNameService: NewDomainsService(DomainsServiceSpec{
			SessionFactory: spec.SessionFactory,
			Logger:         spec.Logger,
		}),
	}
}

func (s *stackService) InjectClusterResourceServiceDeps(deps ClusterResourceServiceDeps) {
	s.clusterResourceService = deps.ClusterStackService
	s.namespaceClusterResourceService = deps.NamespaceClusterService
	s.volumeClusterResourceService = deps.VolumeClusterService
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	// Setup namespace
	namespaceForStack := s.prepareNamespaceForStack(spec)
	spec.Namespace = namespaceForStack.Name
	// Set default values and populate fields
	s.setDefaultValues(spec)
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		spec.StackResources[i].Namespace = spec.Namespace
	}

	// Get domain to use for stack
	domainToUse, err := s.domainNameService.DomainToUseForStack(ctx, spec)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return nil, errors.BadRequest("no domain found for organisation '%s'", spec.OrganisationID)
		}
		return nil, errors.GeneralError("failed to get domain for organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	// Validate volume mounts if present
	if spec.HasVolumeMounts() {
		if err := s.validateVolumeMounts(spec.Volumes, spec); err != nil {
			return nil, err
		}
		setVolumeMountType(spec.Volumes, spec)
	}

	// Populate associations and validate
	if err := s.populateAssociations(ctx, spec); err != nil {
		return nil, errors.GeneralError("failed to populate associations for stack '%s': %s", spec.Name, err.Error())
	}

	// Run validations
	if err := s.validateImageToRun(spec); err != nil {
		return nil, err
	}
	if err := s.validateStackEnvVars(spec); err != nil {
		return nil, err
	}
	if err := s.validateStackPorts(spec); err != nil {
		return nil, err
	}

	if spec.UsesInClusterRegistry() {
		clusterRegistry, err := s.clusterRegistryService.GetForOrg(ctx, spec.OrganisationID)
		if err != nil {
			if err.Code == errors.ErrorNotFound {
				return nil, errors.BadRequest("no cluster registry found for organisation '%s'", spec.OrganisationID)
			}
			return nil, errors.GeneralError("failed to get cluster registry for organisation '%s': %s", spec.OrganisationID, err.Error())
		}
		if clusterRegistry.Status.State != models.RegistryStateRunning {
			return nil, errors.BadRequest("cluster registry '%s' is not running", clusterRegistry.Name)
		}
		registryUrl := clusterRegistry.Status.RegistryUrl
		if len(registryUrl) == 0 {
			return nil, errors.BadRequest("cluster registry '%s' has no registry URL", clusterRegistry.Name)
		}

		urlObj, perr := url.Parse(registryUrl)
		if perr != nil {
			return nil, errors.BadRequest("invalid cluster registry URL '%s': %s", registryUrl, perr.Error())
		}
		spec.PopulateInternalImageRegistryUrlsForResources(urlObj.Hostname())
	}

	// Create stack and update domains within a transaction
	var createdStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// step 1: Create namespace in db
		namespace, err := s.namespaceService.CreateWithTx(ctx, namespaceForStack)
		if err != nil {
			return err
		}
		spec.NamespaceID = namespace.ID

		// Step 2: create volumes in db
		createdVolumesMap := make(map[string]*models.Volume)
		for _, volume := range spec.Volumes {
			volume.NamespaceID = namespace.ID
			volume.OrganisationID = spec.OrganisationID
			volume.UserID = spec.UserID
			volume.Namespace = namespace.Name

			createdVolume, err := s.volumeService.CreateInDbWithTx(ctx, volume)
			if err != nil {
				return errors.GeneralError("failed to create volume '%s': %s", volume.Name, err.Error())
			}
			createdVolumesMap[volume.Name] = createdVolume
		}

		// Step 4: Populate volume mounts source IDs in the stack resources.
		for i := range spec.StackResources {
			currentResource := spec.StackResources[i]
			if len(currentResource.VolumeMounts) == 0 {
				continue
			}
			for j := range currentResource.VolumeMounts {
				currentVolumeMount := currentResource.VolumeMounts[j]
				if volume, found := createdVolumesMap[currentVolumeMount.SourceVolumeName]; found {
					currentVolumeMount.SourceVolumeID = volume.ID
					currentVolumeMount.SourceVolumeName = volume.Name
				} else {
					return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
				}
			}
		}

		// Step 5: Create the stack to get real IDs
		var createErr *errors.ServiceError
		createdStack, createErr = s.stackStore.CreateWithTx(ctx, spec)
		if createErr != nil {
			return createErr
		}

		// Step 6: Associate the created stack with the created volumes
		for volumeName, volume := range createdVolumesMap {
			if err := s.volumeService.UpdateVolumeInUseByStackWithTx(ctx, volume.ID, createdStack.ID); err != nil {
				return errors.GeneralError("failed to update volume '%s' with stack ID '%s': %s", volumeName, createdStack.ID, err.Error())
			}
		}

		// Step 7: Now that we have IDs, populate the domain information
		s.populateExposedPortDomainsForStack(ctx, createdStack, domainToUse)

		// Step 8: Validate domain uniqueness with the populated domains
		if err := s.validateExposedPortDomainUniquenessForStackCreate(ctx, createdStack); err != nil {
			return err
		}

		// Step 9: Update resources and create domains for exposed ports
		stackResourcesToUpdate := make([]*models.StackResource, 0)
		for _, stackResource := range createdStack.StackResources {
			if len(stackResource.Ports) == 0 {
				continue
			}

			for _, port := range stackResource.Ports {
				if port.ExposedToPublic {
					domain := &models.Domain{
						Fqdn:      port.ExposedFqdn,
						OwnerID:   stackResource.ID,
						OwnerType: models.OwnerTypeStackResource,
					}
					if _, err := s.domainNameService.InternalCreateWithTx(ctx, domain); err != nil {
						return errors.GeneralError(
							"failed to create domain '%s' for resource '%s' port '%d': %v",
							port.ExposedFqdn,
							stackResource.Name,
							port.Number,
							err,
						)
					}
					// Only update stack resources with exposed ports
					stackResourcesToUpdate = append(stackResourcesToUpdate, stackResource)
				}
			}
		}

		// Step 10: Update stack resources with generated subdomain prefixes
		for _, stackResource := range stackResourcesToUpdate {
			_, err := s.stackResourceService.InternalUpdateWithTx(ctx, stackResource.ID, stackResource)
			if err != nil {
				return errors.GeneralError(
					"failed to update stack resource '%s' with generated subdomain prefix: %s",
					stackResource.Name, err.Error())
			}
		}

		// Step 11: Get updated stack and create in cluster
		createdStack, err = s.GetStack(ctx, createdStack.ID)
		if err != nil {
			return errors.GeneralError("failed to get created stack '%s': %s", createdStack.ID, err.Error())
		}

		// Step 12: Create namespace in cluster
		if err := s.namespaceClusterResourceService.CreateNamespaceInCluster(ctx, namespace); err != nil {
			return errors.GeneralError("failed to create namespace in cluster: %s", err.Error())
		}

		// Step 13: Create volumes in cluster
		for _, volume := range createdStack.Volumes {
			if err := s.volumeService.CreateInCluster(ctx, volume); err != nil {
				return errors.GeneralError("failed to create volume in cluster: %s", err.Error())
			}
		}

		// Step 14: Create stack in cluster
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

func (s *stackService) UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	// Get existing stack
	existingStack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	existingStackVolumeMap := existingStack.VolumesMap()

	// Validate immutable fields
	if spec.Name != existingStack.Name {
		return nil, errors.BadRequest("stack name cannot be updated")
	}
	if spec.UserID != existingStack.UserID {
		return nil, errors.BadRequest("stack user cannot be updated")
	}
	if spec.OrganisationID != existingStack.OrganisationID {
		return nil, errors.BadRequest("stack organisation cannot be updated")
	}

	// Preserve namespace
	spec.Namespace = existingStack.Namespace

	// Set default values and populate fields
	s.setDefaultValues(spec)
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		spec.StackResources[i].Namespace = spec.Namespace
	}

	// Validate volume mounts if present
	if spec.HasVolumeMounts() {
		if err := s.validateVolumeMounts(spec.Volumes, spec); err != nil {
			return nil, err
		}
		setVolumeMountType(spec.Volumes, spec)
	}

	// Get domain to use
	domainToUse, err := s.domainNameService.DomainToUseForStack(ctx, spec)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return nil, errors.BadRequest("no domain found for organisation '%s'", spec.OrganisationID)
		}
		return nil, errors.GeneralError("failed to get domain for organisation '%s': %s", spec.OrganisationID, err.Error())
	}

	// Run validations
	if err := s.validateImageToRun(spec); err != nil {
		return nil, err
	}
	if err := s.validateStackEnvVars(spec); err != nil {
		return nil, err
	}
	if err := s.validateStackPorts(spec); err != nil {
		return nil, err
	}

	// Update stack and domains within transaction
	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var newlyCreatedVolumesInPatch []*models.Volume
		volumesMap := make(map[string]*models.Volume)
		// Create volumes required for this update
		for _, volume := range spec.Volumes {
			volume.NamespaceID = existingStack.NamespaceID
			volume.OrganisationID = spec.OrganisationID
			volume.UserID = spec.UserID
			volume.Namespace = existingStack.Namespace
			if _, found := existingStackVolumeMap[volume.Name]; !found {
				createdVolume, err := s.volumeService.CreateInDbWithTx(ctx, volume)
				if err != nil {
					return errors.GeneralError("failed to create volume '%s': %s", volume.Name, err.Error())
				}
				// Assoicate volume with the stack.
				if err := s.volumeService.UpdateVolumeInUseByStackWithTx(ctx, createdVolume.ID, ID); err != nil {
					return errors.GeneralError("failed to update volume '%s' with stack ID '%s': %s", volume.Name, ID, err.Error())
				}
				// Add to the list of newly created volumes
				newlyCreatedVolumesInPatch = append(newlyCreatedVolumesInPatch, createdVolume)
				// Add to the map of volumes
				volumesMap[volume.Name] = createdVolume
			} else {
				s.logger.Info(ctx, "Volume '%s' already exists, skipping creation", volume.Name)
				volumesMap[volume.Name] = existingStackVolumeMap[volume.Name]
			}
		}

		// Step 0: Populate volume mounts source IDs in the stack resources.
		for i := range spec.StackResources {
			currentResource := spec.StackResources[i]
			if len(currentResource.VolumeMounts) == 0 {
				continue
			}
			for j := range currentResource.VolumeMounts {
				currentVolumeMount := currentResource.VolumeMounts[j]
				if volume, found := volumesMap[currentVolumeMount.SourceVolumeName]; found {
					if volume.ID == "" {
						return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
					}
					currentVolumeMount.SourceVolumeID = volume.ID
					currentVolumeMount.SourceVolumeName = volume.Name
				} else {
					return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
				}
			}
		}

		// Step 1: Update the stack
		var updateErr *errors.ServiceError
		updatedStack, updateErr = s.stackStore.UpdateWithTx(ctx, ID, spec)
		if updateErr != nil {
			return updateErr
		}

		// Step 3. Cleanup domains for existing stack resources
		for _, existingStackResource := range existingStack.StackResources {
			if err := s.domainNameService.DeleteForOwnerWithTx(ctx, existingStackResource.ID, models.OwnerTypeStackResource); err != nil {
				return errors.GeneralError("failed to delete domains for stack resource '%s': %s", existingStackResource.Name, err.Error())
			}
		}

		// Step 4: Populate domains with real IDs
		s.populateExposedPortDomainsForStack(ctx, updatedStack, domainToUse)

		// Step 5: Validate domain uniqueness
		if err := s.validateExposedPortDomainUniquenessForStackUpdate(ctx, existingStack, updatedStack); err != nil {
			return err
		}

		// Step 5: Update domains for each updated stack resource
		for _, stackResource := range updatedStack.StackResources {
			// New stack resource added in the update
			if err := s.createDomainsForStackResource(ctx, stackResource); err != nil {
				return err
			}
			// Update stack resource with new domain information
			_, err = s.stackResourceService.InternalUpdateWithTx(ctx, stackResource.ID, stackResource)
			if err != nil {
				return errors.GeneralError(
					"failed to update stack resource '%s' with domain information: %s",
					stackResource.Name, err.Error())
			}
		}

		// Step 6: Get updated stack and update in cluster
		updatedStack, err = s.GetStack(ctx, updatedStack.ID)
		if err != nil {
			return errors.GeneralError("failed to get updated stack '%s': %s", updatedStack.ID, err.Error())
		}

		// Step 7: Update volumes in cluster
		for _, volume := range newlyCreatedVolumesInPatch {
			if err := s.volumeService.CreateInCluster(ctx, volume); err != nil {
				return errors.GeneralError("failed to create volume in cluster: %s", err.Error())
			}
		}

		updatedStackVolumeMap := updatedStack.VolumesMap()
		for _, existingVolume := range existingStack.Volumes {
			if _, found := updatedStackVolumeMap[existingVolume.Name]; !found {
				if err := s.volumeService.DeleteWithTx(ctx, existingVolume.ID); err != nil {
					return errors.GeneralError("failed to delete volume: %s", err.Error())
				}
			}
		}

		if err := s.clusterResourceService.UpdateStackInCluster(ctx, updatedStack); err != nil {
			return errors.GeneralError("failed to update stack in cluster: %s", err.Error())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedStack, nil
}

func (s *stackService) GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	stack, err := s.stackStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return stack, nil
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
		// Step 1: Delete from cluster
		if err := s.clusterResourceService.DeleteStackInCluster(ctx, stack); err != nil {
			return errors.GeneralError("failed to delete stack in cluster: %s", err.Error())
		}

		// Step 2: Delete domains for each stack resource
		for _, stackResource := range stack.StackResources {
			if err := s.domainNameService.DeleteForOwnerWithTx(ctx, stackResource.ID, models.OwnerTypeStackResource); err != nil {
				return errors.GeneralError("failed to delete domains for stack resource '%s': %s", stackResource.Name, err.Error())
			}
		}

		// Step 4: Delete the stack from database
		if err := s.stackStore.DeleteWithTx(ctx, ID); err != nil {
			return err
		}

		// Step 3: Delete volumes in cluster
		for _, volume := range stack.Volumes {
			if err := s.volumeService.DeleteWithTx(ctx, volume.ID); err != nil {
				return errors.GeneralError("failed to delete volume in cluster: %s", err.Error())
			}
		}
		// Step 5: Delete the namespace from cluster
		if err := s.namespaceClusterResourceService.DeleteNamespaceInCluster(ctx, &models.Namespace{
			ID:             stack.NamespaceID,
			Name:           stack.Namespace,
			OrganisationID: stack.OrganisationID,
		}); err != nil {
			return errors.GeneralError("failed to delete namespace in cluster: %s", err.Error())
		}

		// Step 6: Delete the namespace from database
		if err := s.namespaceService.DeleteWithTx(ctx, stack.NamespaceID); err != nil {
			return errors.GeneralError("failed to delete namespace in database: %s", err.Error())
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *stackService) createDomainsForStackResource(ctx context.Context, stackResource *models.StackResource) *errors.ServiceError {
	for _, port := range stackResource.Ports {
		if port.ExposedToPublic {
			domain := &models.Domain{
				Fqdn:      port.ExposedFqdn,
				OwnerID:   stackResource.ID,
				OwnerType: models.OwnerTypeStackResource,
			}
			if _, err := s.domainNameService.InternalCreateWithTx(ctx, domain); err != nil {
				return errors.GeneralError(
					"failed to create domain '%s' for resource '%s' port '%d': %v",
					port.ExposedFqdn,
					stackResource.Name,
					port.Number,
					err,
				)
			}
		}
	}
	return nil
}

func (s *stackService) UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError {
	err := s.stackStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		return err
	}
	return nil
}

func (s *stackService) validateVolumeMounts(definedVolumes []*models.Volume, spec *models.Stack) *errors.ServiceError {
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range definedVolumes {
		definedVolumesMap[definedVolumes[i].Name] = definedVolumes[i]
	}

	for i := range spec.StackResources {
		currentResource := spec.StackResources[i]
		for j := range spec.StackResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if _, found := definedVolumesMap[currentVolumeMount.SourceVolumeName]; !found {
				return errors.BadRequest("volume '%s' does not exist", currentVolumeMount.SourceVolumeName)
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

func (s *stackService) validateStackPorts(spec *models.Stack) *errors.ServiceError {
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

func (s *stackService) validateExposedPortDomainUniquenessForStackCreate(
	ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if len(spec.StackResources) == 0 {
		return nil
	}
	for i := range spec.StackResources {
		curr := spec.StackResources[i]
		if curr.Ports == nil {
			continue
		}
		for j := range curr.Ports {
			if curr.Ports[j].ExposedToPublic {
				existingDomain, err := s.domainNameService.GetByFqdn(ctx, curr.Ports[j].ExposedFqdn)
				if err != nil && err.Code != errors.ErrorNotFound {
					return errors.GeneralError("failed to get domain by fqdn '%s': %s", curr.Ports[j].ExposedFqdn, err.Error())
				}
				if existingDomain != nil {
					var usedSubdomainPrefix string
					if curr.Ports[j].SubdomainPrefix == "" {
						usedSubdomainPrefix = curr.Ports[j].GeneratedSubdomainPrefix
					} else {
						usedSubdomainPrefix = curr.Ports[j].SubdomainPrefix
					}
					return errors.Conflict("Domain with fqdn '%s' already exists. Subdomain prefix '%s' for this domain already in use", curr.Ports[j].ExposedFqdn, usedSubdomainPrefix)
				}
			}
		}
	}
	return nil
}

func (s *stackService) validateExposedPortDomainUniquenessForStackUpdate(
	ctx context.Context, existingStack *models.Stack, patch *models.Stack) *errors.ServiceError {
	if len(patch.StackResources) == 0 {
		return nil
	}
	existingResourcePortFqdnMap := existingStack.ExposedPortFqdnMap()
	for i := range patch.StackResources {
		currPatchResource := patch.StackResources[i]
		if currPatchResource.Ports == nil {
			continue
		}
		for j := range currPatchResource.Ports {
			if currPatchResource.Ports[j].ExposedToPublic {
				existingResourcePortFqdnMap, ok := existingResourcePortFqdnMap[currPatchResource.Name]
				if ok {
					_, alreadyExists := existingResourcePortFqdnMap[currPatchResource.Ports[j].Number]
					if alreadyExists {
						// This port is already exposed in the existing stack, so we don't need to check for uniqueness.
						continue
					}
				}
				existingDomain, err := s.domainNameService.GetByFqdn(ctx, currPatchResource.Ports[j].ExposedFqdn)
				if err != nil && err.Code != errors.ErrorNotFound {
					return errors.GeneralError("failed to get domain by fqdn '%s': %s", currPatchResource.Ports[j].ExposedFqdn, err.Error())
				}
				if existingDomain != nil {
					return errors.Conflict("Domain with fqdn '%s' already exists", currPatchResource.Ports[j].ExposedFqdn)
				}
			}
		}
	}
	return nil
}

func (s *stackService) populateAssociations(ctx context.Context, spec *models.Stack) *errors.ServiceError {
	if spec.StackResources == nil {
		return nil
	}
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range spec.Volumes {
		definedVolumesMap[spec.Volumes[i].Name] = spec.Volumes[i]
	}

	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		if spec.StackResources[i].BuildConfig != nil {
			buildConfig := spec.StackResources[i].BuildConfig
			if buildConfig.SourceContext.Volume != nil {
				volume, found := definedVolumesMap[buildConfig.SourceContext.Volume.SourceVolumeName]
				if !found {
					return errors.BadRequest("volume '%s' does not exist", buildConfig.SourceContext.Volume.SourceVolumeName)
				}
				buildConfig.SourceContext.Volume.SourceVolumeName = volume.Name
			}
		}
	}

	return nil
}

func (s *stackService) populateExposedPortDomainsForStack(
	ctx context.Context, spec *models.Stack, domainToUse *models.Domain) {
	for i := range spec.StackResources {
		curr := spec.StackResources[i]
		if len(curr.Ports) == 0 {
			continue
		}
		for j := range curr.Ports {
			if curr.Ports[j].ExposedToPublic {
				// If the exposed port does not have a subdomain prefix, set it to the
				// encoded stack resource ID and port number.
				if curr.Ports[j].SubdomainPrefix == "" {
					curr.Ports[j].GeneratedSubdomainPrefix = encodeStackResourceIDAndPort(curr.ID, curr.Ports[j].Number)
					// Set the exposed port's FQDN using the subdomain prefix and the domain to use.
					curr.Ports[j].ExposedFqdn = fmt.Sprintf(
						"%s.%s", curr.Ports[j].GeneratedSubdomainPrefix, domainToUse.Fqdn)
				} else {
					// Set the exposed port's FQDN using the subdomain prefix and the domain to use.
					curr.Ports[j].ExposedFqdn = fmt.Sprintf(
						"%s.%s", curr.Ports[j].SubdomainPrefix, domainToUse.Fqdn)
				}
			}
		}
	}
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

func (s *stackService) prepareNamespaceForStack(spec *models.Stack) *models.Namespace {
	// Generate a unique namespace for the stack
	namespace := &models.Namespace{
		Name:           fmt.Sprintf("%s-%s", spec.Name, uuid.New().String()),
		OrganisationID: spec.OrganisationID,
	}
	namespace.AddDefaultLabels()

	return namespace
}

func setVolumeMountType(definedVolumes []*models.Volume, spec *models.Stack) {
	definedVolumesMap := make(map[string]*models.Volume)
	for i := range definedVolumes {
		definedVolumesMap[definedVolumes[i].Name] = definedVolumes[i]
	}

	for i := range spec.StackResources {
		for j := range spec.StackResources[i].VolumeMounts {
			if volume, found := definedVolumesMap[spec.StackResources[i].VolumeMounts[j].SourceVolumeName]; found {
				spec.StackResources[i].VolumeMounts[j].SourceVolumeType = volume.VolumeSourceType()
			}
		}
	}
}

func encodeStackResourceIDAndPort(uuid string, port int) string {
	// Combine UUID and port into a single string
	input := uuid + ":" + strconv.Itoa(port)

	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	// Encode the hash using Base32 (URL-safe) and trim padding
	base32Encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash)

	// Truncate the result to 16 characters for a shorter subdomain
	if len(base32Encoded) > 16 {
		base32Encoded = base32Encoded[:16]
	}

	return strings.ToLower(base32Encoded)
}
