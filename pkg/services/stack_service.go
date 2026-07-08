package services

import (
	"context"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/Stackdome/stackdome/pkg/validator"
	stackvalidator "github.com/Stackdome/stackdome/pkg/validator/stack"
	stackresourcevalidator "github.com/Stackdome/stackdome/pkg/validator/stackresource"
	"k8s.io/utils/ptr"
)

type StackService interface {
	CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	ListStackConnections(ctx context.Context, stackID string) (models.StackConnections, *errors.ServiceError)
	CreateStackConnection(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	CreateStackVolume(ctx context.Context, stackID string, volume *models.Volume) (*models.Volume, *errors.ServiceError)
	UpdateStackConnection(ctx context.Context, stackID, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError)
	DeleteStackConnection(ctx context.Context, stackID, connectionID string) *errors.ServiceError
	UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError
	DeleteStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	InternalCreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	InternalUpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	InternalDeleteStack(ctx context.Context, stack *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStackCrRevision(ctx context.Context, ID string, revision string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
	SetReleaseService(rs releaseServiceForStack)
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

type releaseServiceForStack interface {
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
}

type StackServiceSpec struct {
	SessionFactory        db.SessionFactory
	VolumeService         VolumeService
	ClusterService        ClusterService
	OrganisationService   OrganisationService
	StackResourceService  StackResourceService
	NamespaceService      NamespaceService
	SecretService         SecretService
	PostgresAddonService  PostgresAddonService
	TeamService           TeamService
	Permissions           auth.PermissionService
	Logger                logger.Logger
	ReferenceService      ReferenceService
	CredentialResolver    CredentialResolver
	GitIntegrationService GitIntegrationService
}

type stackService struct {
	stackStore           stores.StackStore
	logger               logger.Logger
	sessionFactory       db.SessionFactory
	volumeService        VolumeService
	organisationService  OrganisationService
	stackValidator       validator.StackValidator
	stackResourceService StackResourceService
	namespaceService     NamespaceService
	clusterService       ClusterService
	secretService        SecretService
	postgresAddonService PostgresAddonService
	teamService          TeamService
	permissions          auth.PermissionService
	releaseService       releaseServiceForStack
	referenceService     ReferenceService
	defaultingService    DefaultingService[*models.Stack]
	ClusterResourceServiceDeps
	BackgroundJobEnqueuerDep
}

func NewStackService(spec StackServiceSpec) StackService {
	organisationDomainService := NewOrganisationDomainsService(OrganisationDomainsServiceSpec{
		SessionFactory: spec.SessionFactory,
		Logger:         spec.Logger,
	})
	// The whole-stack path gets its own resourceValidator instance (rather
	// than reusing the one wired into StackResourceService) with Volumes left
	// unset: on create the stack's namespace doesn't exist yet, and volumes
	// bundled in the same request aren't persisted until after validation
	// succeeds, so a namespace-scoped DB lookup would always report them
	// missing. validateVolumeReferences in the stack validator checks bundled
	// volumes against the request payload instead.
	resourceValidator := stackresourcevalidator.NewValidator(stackresourcevalidator.ValidatorSpec{
		Secrets:         spec.SecretService,
		Domains:         organisationDomainService,
		Credentials:     spec.CredentialResolver,
		GitIntegrations: spec.GitIntegrationService,
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
			SecretService:        spec.SecretService,
			PostgresAddonService: spec.PostgresAddonService,
			ResourceValidator:    resourceValidator,
		}),
		stackResourceService: spec.StackResourceService,
		namespaceService:     spec.NamespaceService,
		secretService:        spec.SecretService,
		postgresAddonService: spec.PostgresAddonService,
		teamService:          spec.TeamService,
		permissions:          spec.Permissions,
		referenceService:     spec.ReferenceService,
		defaultingService:    NewStackDefaultingService(),
	}
}

func (s *stackService) SetReleaseService(rs releaseServiceForStack) {
	s.releaseService = rs
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, spec.TeamID, auth.ResourceStacks, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	return s.InternalCreateStack(ctx, spec)
}

func (s *stackService) InternalCreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, _ := s.stackStore.GetByName(ctx, spec.Name, spec.UserID)
	if existingStack != nil {
		return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
	}

	s.logger.Infof("running validation for stack creation: %s", spec.Name)
	if err := s.stackValidator.ValidateForCreate(ctx, spec); err != nil {
		return nil, err
	}
	s.logger.Infof("validation passed for stack creation: %s", spec.Name)

	spec, _ = s.defaultingService.PopulateDefaultValues(spec)

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

	spec.Status = &models.StackStatus{
		State:   models.StackPending,
		Message: "Stack is being created",
	}

	var createdStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		createdStack, err = s.InternalCreateWithTx(ctx, spec, namespaceForStack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, &models.Stack{ID: createdStack.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", spec.Name, err.Error())
	}
	for _, v := range createdStack.Volumes {
		_ = s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, &models.Volume{ID: v.ID})
	}
	return createdStack, nil
}

func (s *stackService) InternalCreateWithTx(ctx context.Context, spec *models.Stack, namespaceForStack *models.Namespace) (*models.Stack, *errors.ServiceError) {
	namespace, err := s.namespaceService.CreateInDBWithTx(ctx, namespaceForStack)
	if err != nil {
		return nil, err
	}
	spec.NamespaceID = namespace.ID

	desiredVolumes := spec.Volumes
	desiredResources := spec.StackResources
	shellSpec := stackShellFrom(spec)

	createdStack, createErr := s.stackStore.CreateWithTx(ctx, &shellSpec)
	if createErr != nil {
		return nil, createErr
	}

	for _, volume := range desiredVolumes {
		if _, err := s.volumeService.InternalCreateWithTx(ctx, createdStack, volume); err != nil {
			return nil, err
		}
	}

	volumesForStack, err := s.volumeService.ListVolumesUsedByStack(ctx, createdStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to list volumes for stack '%s': %s", createdStack.ID, err.Error())
	}
	createdStack.Volumes = volumesForStack

	for _, resource := range desiredResources {
		if _, err := s.stackResourceService.InternalCreateWithTx(ctx, createdStack, resource); err != nil {
			return nil, err
		}
	}

	if err := s.referenceService.ReprojectSpec(ctx, createdStack.ID); err != nil {
		return nil, err
	}

	stack, err := s.stackStore.GetByID(ctx, createdStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get created stack '%s': %s", createdStack.Name, err.Error())
	}
	return stack, nil
}

func (s *stackService) UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingStack.TeamID, auth.ResourceStacks, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	return s.InternalUpdateStack(ctx, ID, spec)
}

func (s *stackService) InternalUpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := s.InternalGetStack(ctx, ID)
	if err != nil {
		return nil, err
	}

	// set namespace
	spec.ID = existingStack.ID
	spec.Namespace = existingStack.Namespace
	spec.ClusterID = existingStack.ClusterID
	spec.OrganisationID = existingStack.OrganisationID
	spec.TeamID = existingStack.TeamID
	spec.UserID = existingStack.UserID

	if err := s.stackValidator.ValidateForUpdate(ctx, existingStack, spec); err != nil {
		return nil, err
	}

	spec, _ = s.defaultingService.PopulateDefaultValues(spec)

	// Update stack and domains within transaction
	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedStack, err = s.InternalUpdateWithTx(ctx, spec, existingStack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	existingVolumeIDs := make(map[string]struct{}, len(existingStack.Volumes))
	for _, v := range existingStack.Volumes {
		existingVolumeIDs[v.ID] = struct{}{}
	}
	for _, v := range updatedStack.Volumes {
		if _, existed := existingVolumeIDs[v.ID]; !existed {
			if enqErr := s.BackgroundJobEnqueuer.Enqueue(&models.Volume{ID: v.ID}); enqErr != nil {
				return nil, errors.GeneralError("failed to enqueue volume '%s': %s", v.ID, enqErr.Error())
			}
		}
	}

	return updatedStack, nil
}

func (s *stackService) InternalUpdateWithTx(ctx context.Context, spec *models.Stack, existingStack *models.Stack) (*models.Stack, *errors.ServiceError) {
	desiredVolumes := spec.Volumes
	desiredResources := spec.StackResources
	shellSpec := stackShellFrom(spec)

	updatedStack, updateErr := s.stackStore.UpdateWithTx(ctx, existingStack.ID, &shellSpec)
	if updateErr != nil {
		return nil, updateErr
	}

	if err := s.volumeService.InternalSyncVolumesWithTx(ctx, updatedStack, existingStack, desiredVolumes); err != nil {
		return nil, err
	}

	volumesForStack, err := s.volumeService.ListVolumesUsedByStack(ctx, existingStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to list volumes used by stack '%s': %s", existingStack.ID, err.Error())
	}
	updatedStack.Volumes = volumesForStack

	if err := s.stackResourceService.InternalSyncResourcesWithTx(ctx, updatedStack, existingStack, desiredResources); err != nil {
		return nil, err
	}
	if err := s.referenceService.ReprojectSpec(ctx, updatedStack.ID); err != nil {
		return nil, err
	}
	stack, err := s.stackStore.GetByID(ctx, updatedStack.ID)
	if err != nil {
		return nil, errors.GeneralError("failed to get updated stack '%s': %s", updatedStack.Name, err.Error())
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

// CreateStackVolume creates a volume and associates it with the stack in one
// transaction — the thin counterpart of the whole-stack PUT's volume sync.
func (s *stackService) CreateStackVolume(ctx context.Context, stackID string, volume *models.Volume) (*models.Volume, *errors.ServiceError) {
	stack, err := s.GetStack(ctx, stackID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	var created *models.Volume
	txErr := s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		// Lock the stack row so concurrent creates serialize; the duplicate-name
		// check below then observes any volume a competing request committed.
		// There is no DB unique constraint on volume name within a stack (the
		// association lives in a join table), so the lock IS the invariant.
		if lockErr := s.stackStore.LockByID(ctx, stackID); lockErr != nil {
			return lockErr
		}
		lockedStack, serr := s.stackStore.GetByID(ctx, stackID)
		if serr != nil {
			return serr
		}
		for _, existing := range lockedStack.Volumes {
			if existing.Name == volume.Name {
				return errors.Conflict("a volume named '%s' already exists in this stack", volume.Name)
			}
		}
		created, serr = s.volumeService.InternalCreateWithTx(ctx, stack, volume)
		return serr
	})
	if txErr != nil {
		return nil, txErr
	}

	if enqErr := s.BackgroundJobEnqueuer.Enqueue(&models.Volume{ID: created.ID}); enqErr != nil {
		return nil, errors.GeneralError("failed to enqueue volume '%s': %s", created.ID, enqErr.Error())
	}
	return created, nil
}

func (s *stackService) CreateStackConnection(ctx context.Context, stackID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	stack, _, err := s.prepareDesiredStackWithConnectionMutation(ctx, stackID, func(connections models.StackConnections) (models.StackConnections, *models.StackConnection, *errors.ServiceError) {
		newConnection := *connection
		newDiscriminator := newConnection.ComputeDiscriminator()
		for _, existing := range connections {
			if existing.Kind == newConnection.Kind &&
				existing.From == newConnection.From &&
				existing.To == newConnection.To &&
				existing.ComputeDiscriminator() == newDiscriminator {
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
		updated.ID = connectionID
		updatedDiscriminator := updated.ComputeDiscriminator()
		found := false
		for i := range connections {
			if connections[i].ID == connectionID {
				connections[i] = updated
				found = true
				continue
			}
			if connections[i].Kind == updated.Kind &&
				connections[i].From == updated.From &&
				connections[i].To == updated.To &&
				connections[i].ComputeDiscriminator() == updatedDiscriminator {
				return nil, nil, errors.Conflict("a '%s' connection from %s to %s already exists",
					updated.Kind,
					connectionNodeLabel(updated.From),
					connectionNodeLabel(updated.To))
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
			if existing.ID == connectionID {
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
	// Copy references to allow mutation
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

func (s *stackService) createStackConnection(ctx context.Context, existingStack *models.Stack, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	var createdConnection *models.StackConnection
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var serr *errors.ServiceError
		createdConnection, serr = s.stackStore.CreateConnectionWithTx(txCtx, existingStack.ID, connection)
		if serr != nil {
			return serr
		}
		return s.referenceService.ReprojectSpec(txCtx, existingStack.ID)
	}); err != nil {
		return nil, err
	}

	return createdConnection, nil
}

func (s *stackService) updateSingleStackConnection(ctx context.Context, existingStack *models.Stack, connectionID string, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	var updatedConnection *models.StackConnection
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var serr *errors.ServiceError
		updatedConnection, serr = s.stackStore.UpdateConnectionWithTx(txCtx, existingStack.ID, connectionID, connection)
		if serr != nil {
			return serr
		}
		return s.referenceService.ReprojectSpec(txCtx, existingStack.ID)
	}); err != nil {
		return nil, err
	}

	return updatedConnection, nil
}

func (s *stackService) deleteSingleStackConnection(ctx context.Context, existingStack *models.Stack, connectionID string) *errors.ServiceError {
	return s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if err := s.stackStore.DeleteConnectionWithTx(txCtx, existingStack.ID, connectionID); err != nil {
			return err
		}
		return s.referenceService.ReprojectSpec(txCtx, existingStack.ID)
	})
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
	return s.InternalDeleteStack(ctx, stack)
}

func (s *stackService) InternalDeleteStack(ctx context.Context, stack *models.Stack) (*models.Stack, *errors.ServiceError) {
	if stack.Status.State == models.StackDeleting {
		return stack, nil
	}
	if s.releaseService != nil {
		if active, _ := s.releaseService.InternalGetActiveByStackID(ctx, stack.ID); active != nil {
			s.releaseService.MarkFailed(ctx, active.ID, "stack deleted", nil)
		}
	}
	stack.DeletionTimestamp = ptr.To(time.Now().UTC())
	stack.Status.State = models.StackDeleting
	stack.Status.Message = "Stack is being deleted"
	stackMarkedForDelete, err := s.stackStore.UpdateForDelete(ctx, stack.ID, stack)
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

func connectionNodeLabel(ref models.TopologyNodeRef) string {
	if ref.Name != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Name)
	}
	if ref.Id != "" {
		return fmt.Sprintf("%s:%s", ref.Type, ref.Id)
	}
	return string(ref.Type)
}
