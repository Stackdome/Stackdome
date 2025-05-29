package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	stackvalidator "github.com/ashishmax31/stackdome-api-server/pkg/validator/stack"
)

type StackService interface {
	CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError
	DeleteStack(ctx context.Context, ID string) *errors.ServiceError
	ClusterResourceServiceInjectable
	StackQueryService
}

type StackQueryService interface {
	GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError)
	GetStacksByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError)
	GetStacksByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError)
}

type StackServiceSpec struct {
	SessionFactory         db.SessionFactory
	VolumeService          VolumeService
	ClusterService         ClusterService
	OrganisationService    OrganisationService
	StackResourceService   StackResourceService
	ClusterRegistryService ImageRegistryService
	NamespaceService       NamespaceService
	Logger                 logger.Logger
}

type stackService struct {
	stackStore             stores.StackStore
	logger                 logger.Logger
	sessionFactory         db.SessionFactory
	volumeService          VolumeService
	organisationService    OrganisationService
	stackValidator         validator.StackValidator
	domainNameService      StackDomainsService
	stackResourceService   StackResourceService
	namespaceService       NamespaceService
	clusterRegistryService ImageRegistryService
	defaultingService      DefaultingService[*models.Stack]
	ClusterResourceServiceDeps
}

func NewStackService(spec StackServiceSpec) StackService {
	stackDomainNameService := NewStackDomainsService(StackDomainsServiceSpec{
		SessionFactory: spec.SessionFactory,
		Logger:         spec.Logger,
	})

	organisationDomainService := NewOrganisationDomainsService(OrganisationDomainsServiceSpec{
		SessionFactory: spec.SessionFactory,
		Logger:         spec.Logger,
	})
	return &stackService{
		stackStore: pgstore.NewStackStore(&pgstore.StackStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		volumeService:          spec.VolumeService,
		organisationService:    spec.OrganisationService,
		logger:                 spec.Logger,
		sessionFactory:         spec.SessionFactory,
		stackValidator:         stackvalidator.NewStackValidator(organisationDomainService),
		stackResourceService:   spec.StackResourceService,
		clusterRegistryService: spec.ClusterRegistryService,
		namespaceService:       spec.NamespaceService,
		domainNameService:      stackDomainNameService,
		defaultingService:      NewStackDefaultingService(),
	}
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, _ := s.GetStackByName(ctx, spec.Name, spec.UserID)
	if existingStack != nil {
		return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
	}

	if err := s.stackValidator.ValidateForCreate(ctx, spec); err != nil {
		return nil, err
	}
	// Setup namespace
	namespaceForStack, err := s.namespaceService.PrepareNamespaceForStack(ctx, spec)
	if err != nil {
		return nil, errors.GeneralError("failed to prepare namespace for stack '%s': %s", spec.Name, err.Error())
	}
	spec.Namespace = namespaceForStack.Name

	// Set default values
	spec, dErr := s.defaultingService.PopulateDefaultValues(spec)
	if dErr != nil {
		return nil, errors.GeneralError("failed to populate default values for stack '%s': %s", spec.Name, dErr.Error())
	}

	// Populate associations
	s.populateAssociations(ctx, spec)

	// Set registry urls for stack resources.
	if err := s.clusterRegistryService.PopulateInClusterRegistryUrlsForStack(ctx, spec); err != nil {
		return nil, errors.GeneralError("failed to populate in-cluster registry URLs for stack '%s': %s", spec.Name, err.Error())
	}

	var createdStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Step 1: Create stack and dependencies in DB
		createdStack, err = s.createStackAndDepsInDbWithTx(ctx, spec, namespaceForStack)
		if err != nil {
			return err
		}
		// Step 2: Create namespace in cluster
		if err := s.namespaceService.CreateInCluster(ctx, namespaceForStack); err != nil {
			return errors.GeneralError("failed to create namespace in cluster: %s", err.Error())
		}

		// Step 3: Create volumes in cluster
		for _, volume := range createdStack.Volumes {
			if err := s.volumeService.CreateInCluster(ctx, volume); err != nil {
				return errors.GeneralError("failed to create volume in cluster: %s", err.Error())
			}
		}

		// Step 4: Create stack in cluster
		if err := s.ClusterStackService.CreateStackInCluster(ctx, createdStack); err != nil {
			return errors.GeneralError("failed to create stack in cluster: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetStack(ctx, createdStack.ID)
}

// Creates stack and all its dependencies in DB wrapped in a transaction. (ctx should be a transaction context)
func (s *stackService) createStackAndDepsInDbWithTx(ctx context.Context, spec *models.Stack, namespaceForStack *models.Namespace) (*models.Stack, *errors.ServiceError) {
	// step 1: Create namespace in db
	namespace, err := s.namespaceService.CreateInDBWithTx(ctx, namespaceForStack)
	if err != nil {
		return nil, err
	}
	spec.NamespaceID = namespace.ID

	// Step 2: create volumes in db
	createdVolumes, err := s.volumeService.CreateVolumesInDBForStackWithTx(ctx, spec)
	if err != nil {
		return nil, errors.GeneralError("failed to create volumes for stack '%s': %s", spec.Name, err.Error())
	}
	spec.Volumes = createdVolumes

	// Step 3: Create the stack to get real IDs
	var createErr *errors.ServiceError
	createdStack, createErr := s.stackStore.CreateWithTx(ctx, spec)
	if createErr != nil {
		return nil, createErr
	}

	// Step 4: Associate the created stack with the created volumes
	for _, volume := range createdVolumes {
		if err := s.volumeService.UpdateVolumeInUseByStackWithTx(ctx, volume.ID, createdStack.ID); err != nil {
			return nil, errors.GeneralError("failed to update volume '%s' with stack ID '%s': %s", volume.Name, createdStack.ID, err.Error())
		}
	}

	// Step 5: Populate and save the domans for the stack resources with exposed ports.
	if err := s.domainNameService.PopulateAndSaveExposedPortDomainsForStackWithTx(ctx, createdStack); err != nil {
		return nil, err
	}

	// Step 6: Update stack resources with subdomain prefixes from step 5.
	for _, stackResource := range createdStack.StackResources {
		err := s.stackResourceService.InternalUpdateExposedPortDomainsWithTx(ctx, stackResource.ID, stackResource, createdStack)
		if err != nil {
			return nil, errors.GeneralError(
				"failed to update stack resource '%s' with generated subdomain prefix: %s",
				stackResource.Name, err.Error())
		}
	}
	return s.GetStack(ctx, createdStack.ID)
}

func (s *stackService) UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	// Get existing stack
	existingStack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}

	if err := s.stackValidator.ValidateForUpdate(ctx, existingStack, spec); err != nil {
		return nil, err
	}
	// set namespace
	spec.Namespace = existingStack.Namespace

	// Set default values and populate fields
	spec, derr := s.defaultingService.PopulateDefaultValues(spec)
	if derr != nil {
		return nil, errors.GeneralError("failed to populate default values for stack '%s': %s", spec.Name, derr.Error())
	}

	s.populateAssociations(ctx, spec)

	existingVolumes := existingStack.VolumesMap()
	// Update stack and domains within transaction
	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Step 5: Get updated stack and update in cluster
		updatedStack, err = s.updateStackAndDepsInDbWithTx(ctx, spec, existingStack)
		if err != nil {
			return err
		}
		// Step 6: Update volumes in cluster
		for _, volume := range updatedStack.Volumes {
			// Check if the volume is new or existing
			// If it is new, create it in the cluster
			if _, ok := existingVolumes[volume.Name]; !ok {
				if err := s.volumeService.CreateInCluster(ctx, volume); err != nil {
					return errors.GeneralError("failed to create volume in cluster: %s", err.Error())
				}
			}
		}
		// We dont delete volumes not present in this update request.We let them be.
		// either user will delete them or we will delete them when stack is deleted.
		if err := s.ClusterStackService.UpdateStackInCluster(ctx, updatedStack); err != nil {
			return errors.GeneralError("failed to update stack in cluster: %s", err.Error())
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedStack, nil
}

func (s *stackService) updateStackAndDepsInDbWithTx(ctx context.Context, spec *models.Stack, existingStack *models.Stack) (*models.Stack, *errors.ServiceError) {
	// Step 1: Update volumes in db
	newlyCreatedVolumesInPatch, err := s.volumeService.UpdateVolumesInDBForStackWithTx(ctx, spec, existingStack)
	if err != nil {
		return nil, err
	}

	// Step 2: Associate the newly created volumes with the stack.
	for _, volume := range newlyCreatedVolumesInPatch {
		if err := s.volumeService.UpdateVolumeInUseByStackWithTx(ctx, volume.ID, existingStack.ID); err != nil {
			return nil, errors.GeneralError("failed to update volume '%s' with stack ID '%s': %s", volume.Name, existingStack.ID, err.Error())
		}
	}

	volumesForStack, err := s.volumeService.ListVolumesUsedByStack(ctx, existingStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to list volumes used by stack '%s': %s", existingStack.ID, err.Error())
	}
	spec.Volumes = volumesForStack

	updatedStack, updateErr := s.stackStore.UpdateWithTx(ctx, existingStack.ID, spec)
	if updateErr != nil {
		return nil, updateErr
	}

	// Step 4: Populate and save the domains for the stack resources with exposed ports.
	if err := s.domainNameService.PopulateAndSaveExposedPortDomainsForStackWithTx(ctx, updatedStack); err != nil {
		return nil, err
	}

	// Step 5: Update stack resources with subdomain prefixes from step 5.
	for _, stackResource := range updatedStack.StackResources {
		err = s.stackResourceService.InternalUpdateExposedPortDomainsWithTx(ctx, stackResource.ID, stackResource, updatedStack)
		if err != nil {
			return nil, errors.GeneralError(
				"failed to update stack resource '%s' with domain information: %s",
				stackResource.Name, err.Error())
		}
	}

	// Step 5: Get updated stack and update in cluster
	return s.GetStack(ctx, updatedStack.ID)
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
		if err := s.ClusterStackService.DeleteStackInCluster(ctx, stack); err != nil {
			return errors.GeneralError("failed to delete stack in cluster: %s", err.Error())
		}

		// Step 2: Delete the stack from database
		if err := s.stackStore.DeleteWithTx(ctx, ID); err != nil {
			return err
		}

		// Step 3: Delete volumes in cluster in DB.
		for _, volume := range stack.Volumes {
			if err := s.volumeService.DeleteWithTx(ctx, volume.ID); err != nil {
				return errors.GeneralError("failed to delete volume in cluster: %s", err.Error())
			}
		}
		// Step 4: Delete the namespace from cluster
		if err := s.ClusterNamespaceService.DeleteNamespaceInCluster(ctx, &models.Namespace{
			ID:             stack.NamespaceID,
			Name:           stack.Namespace,
			OrganisationID: stack.OrganisationID,
		}); err != nil {
			return errors.GeneralError("failed to delete namespace in cluster: %s", err.Error())
		}

		// Step 5: Delete the namespace from database
		if err := s.namespaceService.DeleteFromDBWithTx(ctx, stack.NamespaceID); err != nil {
			return errors.GeneralError("failed to delete namespace in database: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return err
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

func (s *stackService) populateAssociations(ctx context.Context, spec *models.Stack) {
	// Populate the stack resources with the user ID and namespace
	for i := range spec.StackResources {
		spec.StackResources[i].UserID = spec.UserID
		spec.StackResources[i].Namespace = spec.Namespace
	}
}
