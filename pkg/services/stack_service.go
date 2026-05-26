package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	stackvalidator "github.com/ashishmax31/stackdome-api-server/pkg/validator/stack"
	"k8s.io/utils/ptr"
)

type StackService interface {
	CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	ListStackConnections(ctx context.Context, stackID string) (models.StackConnections, *errors.ServiceError)
	CreateStackConnection(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	UpdateStackConnection(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	DeleteStackConnection(ctx context.Context, stackID, connectionID string) *errors.ServiceError
	UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError
	DeleteStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	UpdateStackCrRevision(ctx context.Context, ID string, revision string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
	ClusterResourceServiceInjectable
	BackgroundJobEnqueuerInjectable
	StackQueryService
}

type StackQueryService interface {
	GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	GetStackTopology(ctx context.Context, ID string) (*models.StackTopology, *errors.ServiceError)
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError)
	// GetStacksByUserID(ctx context.Context, teamID, orgID, userID string) ([]*models.Stack, *errors.ServiceError)
	GetStacksByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError)
	GetStacksByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError)
	ListStacksForCurrentUser(ctx context.Context, orgID string) ([]*models.Stack, *errors.ServiceError)
}

type StackServiceSpec struct {
	SessionFactory         db.SessionFactory
	VolumeService          VolumeService
	ClusterService         ClusterService
	OrganisationService    OrganisationService
	StackResourceService   StackResourceService
	ClusterRegistryService ImageRegistryService
	NamespaceService       NamespaceService
	SecretService          SecretService
	PostgresAddonService   PostgresAddonService
	TeamService            TeamService
	Permissions            auth.PermissionService
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
	clusterResourceBuidler builders.ClusterResourceBuilder
	namespaceService       NamespaceService
	clusterService         ClusterService
	secretService          SecretService
	postgresAddonService   PostgresAddonService
	clusterRegistryService ImageRegistryService
	defaultingService      DefaultingService[*models.Stack]
	teamService            TeamService
	permissions            auth.PermissionService
	ClusterResourceServiceDeps
	BackgroundJobEnqueuerDep
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
		clusterService:      spec.ClusterService,
		volumeService:       spec.VolumeService,
		organisationService: spec.OrganisationService,
		logger:              spec.Logger,
		sessionFactory:      spec.SessionFactory,
		stackValidator: stackvalidator.NewStackValidator(stackvalidator.StackValidatorSpec{
			DomainService:        organisationDomainService,
			SecretService:        spec.SecretService,
			PostgresAddonService: spec.PostgresAddonService,
		}),
		stackResourceService:   spec.StackResourceService,
		clusterRegistryService: spec.ClusterRegistryService,
		namespaceService:       spec.NamespaceService,
		domainNameService:      stackDomainNameService,
		secretService:          spec.SecretService,
		postgresAddonService:   spec.PostgresAddonService,
		clusterResourceBuidler: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			SecretService: spec.SecretService,
		}),
		defaultingService: NewStackDefaultingService(),
		teamService:       spec.TeamService,
		permissions:       spec.Permissions,
	}
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, spec.TeamID, auth.ResourceStacks, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	existingStack, _ := s.GetStackByName(ctx, spec.Name, spec.UserID)
	if existingStack != nil {
		return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
	}
	s.logger.Infof("running validation for stack creation: %s", spec.Name)
	if err := s.stackValidator.ValidateForCreate(ctx, spec); err != nil {
		return nil, err
	}
	s.logger.Infof("validation passed for stack creation: %s", spec.Name)
	// Setup namespace
	namespaceForStack, err := s.namespaceService.PrepareNamespaceForStack(ctx, spec)
	if err != nil {
		return nil, errors.GeneralError("failed to prepare namespace for stack '%s': %s", spec.Name, err.Error())
	}
	spec.Namespace = namespaceForStack.Name

	cluster, err := s.clusterService.GetClusterForOrg(ctx, spec.OrganisationID)
	if err != nil {
		return nil, errors.GeneralError("failed to get cluster for organisation '%s': %s", spec.OrganisationID, err.Error())
	}
	spec.ClusterID = cluster.ID

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

	spec.Status = &models.StackStatus{
		State:   models.StackPending,
		Message: "Stack is being created",
	}

	var createdStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Step 1: Create stack and dependencies in DB
		createdStack, err = s.createStackAndDepsInDbWithTx(ctx, spec, namespaceForStack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.BackgroundJobEnqueuer.Enqueue(&models.Stack{
		ID: createdStack.ID,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", spec.Name, err.Error())
	}
	return createdStack, nil
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

	createdStack.Volumes = createdVolumes
	// Step 5: Populate and save the domans for the stack resources with exposed ports.
	if err := s.domainNameService.PopulateAndSaveExposedPortDomainsForStackWithTx(ctx, createdStack); err != nil {
		return nil, err
	}

	// Step 6: Update stack resources with subdomain prefixes from step 5.
	for _, stackResource := range createdStack.StackResources {
		err := s.stackResourceService.InternalUpdateExposedPortDomainsWithTx(ctx, stackResource.ID, stackResource)
		if err != nil {
			return nil, errors.GeneralError(
				"failed to update stack resource '%s' with generated subdomain prefix: %s",
				stackResource.Name, err.Error())
		}
	}
	stack, err := s.GetStack(ctx, createdStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get created stack '%s': %s", createdStack.Name, err.Error())
	}

	// Step 7: Build and update the CR hash for the stack
	crHash, cerr := s.clusterResourceBuidler.GetStackCRHash(stack)
	if cerr != nil {
		return nil, errors.GeneralError("failed to build stack CR hash for stack '%s': %s", stack.Name, cerr.Error())
	}

	stack.CrRevision = crHash
	if err := s.UpdateStackCrRevision(ctx, stack.ID, crHash); err != nil {
		return nil, errors.GeneralError("failed to update stack CR revision for stack '%s': %s", stack.Name, err.Error())
	}
	return stack, nil
}

func (s *stackService) UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	// Get existing stack (includes read permission check)
	existingStack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingStack.TeamID, auth.ResourceStacks, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if err := s.stackValidator.ValidateForUpdate(ctx, existingStack, spec); err != nil {
		return nil, err
	}
	// set namespace
	spec.Namespace = existingStack.Namespace
	spec.ClusterID = existingStack.ClusterID
	spec.OrganisationID = existingStack.OrganisationID
	spec.TeamID = existingStack.TeamID
	spec.UserID = existingStack.UserID

	// Set default values and populate fields
	spec, derr := s.defaultingService.PopulateDefaultValues(spec)
	if derr != nil {
		return nil, errors.GeneralError("failed to populate default values for stack '%s': %s", spec.Name, derr.Error())
	}

	s.populateAssociations(ctx, spec)

	// Update stack and domains within transaction
	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Step 5: Get updated stack and update in cluster
		updatedStack, err = s.updateStackAndDepsInDbWithTx(ctx, spec, existingStack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.BackgroundJobEnqueuer.Enqueue(&models.Stack{
		ID: updatedStack.ID,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", spec.Name, err.Error())
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

	updatedStack.Volumes = volumesForStack

	// Step 4: Populate and save the domains for the stack resources with exposed ports.
	if err := s.domainNameService.PopulateAndSaveExposedPortDomainsForStackWithTx(ctx, updatedStack); err != nil {
		return nil, err
	}

	// Step 5: Update stack resources with subdomain prefixes from step 5.
	for _, stackResource := range updatedStack.StackResources {
		err = s.stackResourceService.InternalUpdateExposedPortDomainsWithTx(ctx, stackResource.ID, stackResource)
		if err != nil {
			return nil, errors.GeneralError(
				"failed to update stack resource '%s' with domain information: %s",
				stackResource.Name, err.Error())
		}
	}

	// Step 5: Get updated stack and update in cluster
	stack, err := s.GetStack(ctx, updatedStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get updated stack '%s': %s", updatedStack.Name, err.Error())
	}

	// Step 6: Build and update the CR hash for the stack
	crHash, cerr := s.clusterResourceBuidler.GetStackCRHash(stack)
	if cerr != nil {
		return nil, errors.GeneralError("failed to build stack CR hash for stack '%s': %s", stack.Name, cerr.Error())
	}

	stack.CrRevision = crHash
	if err := s.UpdateStackCrRevision(ctx, stack.ID, crHash); err != nil {
		return nil, errors.GeneralError("failed to update stack CR revision for stack '%s': %s", stack.Name, err.Error())
	}
	return stack, nil
}

func (s *stackService) GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	stack, err := s.stackStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return stack, nil
}


func (s *stackService) ListStackConnections(ctx context.Context, stackID string) (models.StackConnections, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, stackID)
	if err != nil {
		return nil, err
	}
	return stack.Connections, nil
}

func (s *stackService) CreateStackConnection(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	stack, _, err := s.prepareDesiredStackWithConnectionMutation(ctx, stackID, func(connections models.StackConnections) (models.StackConnections, *models.StackConnection, *errors.ServiceError) {
		newConnection := *connection
		for _, existing := range connections {
			if existing.Kind == newConnection.Kind &&
				existing.From == newConnection.From &&
				existing.To == newConnection.To {
				return nil, nil, errors.Conflict("a '%s' connection from %s to %s already exists",
					newConnection.Kind,
					connectionNodeLabel(newConnection.From),
					connectionNodeLabel(newConnection.To))
			}
		}
		connections = append(connections, newConnection)
		return connections, &newConnection, nil
	})
	if err != nil {
		return nil, err
	}

	createdConnection, err := s.createStackConnection(ctx, stack, connection)
	return createdConnection, err
}

func (s *stackService) UpdateStackConnection(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	stack, _, err := s.prepareDesiredStackWithConnectionMutation(ctx, stackID, func(connections models.StackConnections) (models.StackConnections, *models.StackConnection, *errors.ServiceError) {
		updated := *connection
		updated.Id = connectionID
		found := false
		for i := range connections {
			if connections[i].Id == connectionID {
				connections[i] = updated
				found = true
				break
			}
		}
		if !found {
			return nil, nil, errors.NotFound("stack connection '%s' not found", connectionID)
		}
		return connections, &updated, nil
	})
	if err != nil {
		return nil, err
	}

	return s.updateSingleStackConnection(ctx, stack, connectionID, connection)
}

func (s *stackService) DeleteStackConnection(ctx context.Context, stackID, connectionID string) *errors.ServiceError {
	stack, _, err := s.prepareDesiredStackWithConnectionMutation(ctx, stackID, func(connections models.StackConnections) (models.StackConnections, *models.StackConnection, *errors.ServiceError) {
		result := make(models.StackConnections, 0, len(connections))
		found := false
		for _, existing := range connections {
			if existing.Id == connectionID {
				found = true
				continue
			}
			result = append(result, existing)
		}
		if !found {
			return nil, nil, errors.NotFound("stack connection '%s' not found", connectionID)
		}
		return result, nil, nil
	})
	if err != nil {
		return err
	}

	return s.deleteSingleStackConnection(ctx, stack, connectionID)
}

func (s *stackService) InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	return s.stackStore.GetByID(ctx, ID)
}

func (s *stackService) prepareDesiredStackWithConnectionMutation(
	ctx context.Context,
	stackID string,
	mutate func(connections models.StackConnections) (models.StackConnections, *models.StackConnection, *errors.ServiceError),
) (*models.Stack, *models.Stack, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, stackID)
	if err != nil {
		return nil, nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, nil, permErr
	}

	desired := *stack
	desired.Connections = append(models.StackConnections(nil), stack.Connections...)
	nextConnections, _, serr := mutate(desired.Connections)
	if serr != nil {
		return nil, nil, serr
	}
	desired.Connections = nextConnections
	if err := s.stackValidator.ValidateForUpdate(ctx, stack, &desired); err != nil {
		return nil, nil, err
	}
	return stack, &desired, nil
}

func (s *stackService) updateStackConnections(ctx context.Context, existingStack *models.Stack, desired *models.Stack) (*models.Stack, *errors.ServiceError) {
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		return s.stackStore.UpdateConnectionsWithTx(txCtx, existingStack.ID, desired.Connections)
	}); err != nil {
		return nil, err
	}

	updatedStack, err := s.GetStack(ctx, existingStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get updated stack '%s': %s", existingStack.Name, err.Error())
	}

	crHash, cerr := s.clusterResourceBuidler.GetStackCRHash(updatedStack)
	if cerr != nil {
		return nil, errors.GeneralError("failed to build stack CR hash for stack '%s': %s", updatedStack.Name, cerr.Error())
	}
	updatedStack.CrRevision = crHash
	if err := s.UpdateStackCrRevision(ctx, updatedStack.ID, crHash); err != nil {
		return nil, errors.GeneralError("failed to update stack CR revision for stack '%s': %s", updatedStack.Name, err.Error())
	}
	if err := s.BackgroundJobEnqueuer.Enqueue(&models.Stack{ID: updatedStack.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", updatedStack.Name, err.Error())
	}
	return updatedStack, nil
}

func (s *stackService) createStackConnection(ctx context.Context, existingStack *models.Stack, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	var createdConnection *models.StackConnection
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var serr *errors.ServiceError
		createdConnection, serr = s.stackStore.CreateConnectionWithTx(txCtx, existingStack.ID, connection)
		return serr
	}); err != nil {
		return nil, err
	}

	if _, err := s.refreshStackRevisionAndEnqueue(ctx, existingStack); err != nil {
		return nil, err
	}
	return createdConnection, nil
}

func (s *stackService) updateSingleStackConnection(ctx context.Context, existingStack *models.Stack, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	var updatedConnection *models.StackConnection
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var serr *errors.ServiceError
		updatedConnection, serr = s.stackStore.UpdateConnectionWithTx(txCtx, existingStack.ID, connectionID, connection)
		return serr
	}); err != nil {
		return nil, err
	}

	if _, err := s.refreshStackRevisionAndEnqueue(ctx, existingStack); err != nil {
		return nil, err
	}
	return updatedConnection, nil
}

func (s *stackService) deleteSingleStackConnection(ctx context.Context, existingStack *models.Stack, connectionID string) *errors.ServiceError {
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		return s.stackStore.DeleteConnectionWithTx(txCtx, existingStack.ID, connectionID)
	}); err != nil {
		return err
	}

	_, err := s.refreshStackRevisionAndEnqueue(ctx, existingStack)
	return err
}

func (s *stackService) refreshStackRevisionAndEnqueue(ctx context.Context, existingStack *models.Stack) (*models.Stack, *errors.ServiceError) {
	updatedStack, err := s.GetStack(ctx, existingStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get updated stack '%s': %s", existingStack.Name, err.Error())
	}
	crHash, cerr := s.clusterResourceBuidler.GetStackCRHash(updatedStack)
	if cerr != nil {
		return nil, errors.GeneralError("failed to build stack CR hash for stack '%s': %s", updatedStack.Name, cerr.Error())
	}
	updatedStack.CrRevision = crHash
	if err := s.UpdateStackCrRevision(ctx, updatedStack.ID, crHash); err != nil {
		return nil, errors.GeneralError("failed to update stack CR revision for stack '%s': %s", updatedStack.Name, err.Error())
	}
	if err := s.BackgroundJobEnqueuer.Enqueue(&models.Stack{ID: updatedStack.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", updatedStack.Name, err.Error())
	}
	return updatedStack, nil
}

func (s *stackService) GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError) {
	stack, err := s.stackStore.GetByName(ctx, name, userID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stack.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return stack, nil
}

// func (s *stackService) GetStacksByUserID(ctx context.Context, teamID, orgID, userID string) ([]*models.Stack, *errors.ServiceError) {
// 	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceStacks, "", auth.ActionList); permErr != nil {
// 		return nil, permErr
// 	}
// 	stacks, err := s.stackStore.ListByUserID(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return stacks, nil
// }

func (s *stackService) InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError) {
	stacks, err := s.stackStore.InternalList(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (s *stackService) GetStacksByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceStacks, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	stacks, err := s.stackStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (s *stackService) GetStacksByOrganisationID(ctx context.Context, organisationID string) ([]*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceStacks, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.stackStore.ListByOrganisationID(ctx, organisationID)
}

func (s *stackService) ListStacksForCurrentUser(ctx context.Context, orgID string) ([]*models.Stack, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}

	if identity.IsOrgAdmin() {
		return s.stackStore.ListByOrganisationID(ctx, orgID)
	}

	memberships, serr := s.teamService.InternalListUserTeams(ctx, identity.UserID, orgID)
	if serr != nil {
		return nil, serr
	}

	var allowedTeamIDs []string
	for _, m := range memberships {
		if permErr := s.permissions.Check(ctx, m.TeamID, auth.ResourceStacks, "", auth.ActionList); permErr == nil {
			allowedTeamIDs = append(allowedTeamIDs, m.TeamID)
		}
	}

	return s.stackStore.ListByTeamIDs(ctx, allowedTeamIDs)
}

func (s *stackService) DeleteStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	// GetStack includes read permission check
	stack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, ID, auth.ActionDelete); permErr != nil {
		return nil, permErr
	}
	if stack.Status.State == models.StackDeleting {
		return stack, nil
	}
	stack.DeletionTimestamp = ptr.To(time.Now().UTC())
	stack.Status.State = models.StackDeleting
	stack.Status.Message = "Stack is being deleted"
	stackMarkedForDelete, err := s.stackStore.UpdateForDelete(ctx, ID, stack)
	if err != nil {
		return nil, errors.GeneralError("failed to update stack '%s' for deletion: %s", stack.Name, err.Error())
	}
	if err := s.BackgroundJobEnqueuer.Enqueue(&models.Stack{
		ID: stack.ID,
	}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", stack.Name, err.Error())
	}
	return stackMarkedForDelete, nil
}

func (s *stackService) InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError {
	err := s.stackStore.Delete(ctx, ID)
	if err != nil {
		if err.Is404() {
			// If the stack is not found, we can return nil as it is already deleted.
			return nil
		}
		return errors.GeneralError("failed to delete stack with ID '%s' from database: %s", ID, err.Error())
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

func (s *stackService) UpdateStackCrRevision(ctx context.Context, ID string, revision string) *errors.ServiceError {
	if err := s.stackStore.UpdateRevision(ctx, ID, revision); err != nil {
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

func connectionNodeLabel(ref models.TopologyNodeRef) string {
	if ref.Name != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Name)
	}
	if ref.Id != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Id)
	}
	return string(ref.Type)
}
