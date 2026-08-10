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

//go:generate mockgen -destination=stack_service_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services github.com/Stackdome/stackdome/pkg/services StackService
type StackService interface {
	CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	UpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	ApplyStack(ctx context.Context, spec *models.Stack) (*models.Stack, bool, *errors.ServiceError)
	UpdateStackShell(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
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
	InternalListWorkloadAuthorityCandidates(ctx context.Context) ([]*models.Stack, *errors.ServiceError)
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
	GetStacksByProjectID(ctx context.Context, projectID string) ([]*models.Stack, *errors.ServiceError)
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
	ProjectService        ProjectService
	Permissions           auth.PermissionService
	Logger                logger.Logger
	ReferenceService      ReferenceService
	CredentialResolver    CredentialResolver
	GitIntegrationService GitIntegrationService
	PlatformBaseDomain    string
	RuntimePolicy         RuntimePolicy
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
	projectService       ProjectService
	permissions          auth.PermissionService
	releaseService       releaseServiceForStack
	referenceService     ReferenceService
	defaultingService    DefaultingService[*models.Stack]
	ClusterResourceServiceDeps
	BackgroundJobEnqueuerDep
	runtimePolicy RuntimePolicy
}

func NewStackService(spec StackServiceSpec) StackService {
	organisationDomainService := NewOrganisationDomainsService(OrganisationDomainsServiceSpec{
		SessionFactory: spec.SessionFactory,
		Logger:         spec.Logger,
	})
	// The whole-stack path gets its own resourceValidator instance (rather
	// than reusing the one wired into StackResourceService). Volumes is a
	// real DB-backed store here too - stackresource.validateMountedVolumes
	// checks a mount against the request's own stack.Volumes before ever
	// consulting this seam, so volumes bundled in the same (unpersisted)
	// request still resolve correctly; the seam only gets used as a
	// fallback for names/IDs the payload doesn't declare, where a
	// namespace-scoped DB lookup is exactly what we want.
	resourceValidator := stackresourcevalidator.NewValidator(stackresourcevalidator.ValidatorSpec{
		Volumes: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		// Raw store, not the RBAC-enforcing SecretService: env secret_key_ref
		// validation is an org-scoped existence check, not an authorized
		// read. The thin per-resource path (cmd/environment) wires the same
		// raw store, so both paths accept the same payload regardless of
		// whether the caller holds secrets:read.
		Secrets: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		Domains:            organisationDomainService,
		Credentials:        spec.CredentialResolver,
		GitIntegrations:    spec.GitIntegrationService,
		PlatformBaseDomain: spec.PlatformBaseDomain,
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
		projectService:       spec.ProjectService,
		permissions:          spec.Permissions,
		referenceService:     spec.ReferenceService,
		defaultingService:    NewStackDefaultingService(),
		runtimePolicy:        spec.RuntimePolicy,
	}
}

func (s *stackService) SetReleaseService(rs releaseServiceForStack) {
	s.releaseService = rs
}

func (s *stackService) CreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, spec.ProjectID, auth.ResourceStacks, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}
	return s.InternalCreateStack(ctx, spec)
}

// ApplyStack upserts a stack by name within its project scope (stack names are
// unique per project). When a stack with spec.Name exists in spec.ProjectID it is
// fully replaced through the same path as the id-addressed apply; otherwise
// the stack and its children are created atomically after full validation.
// The returned bool is true when a new stack was created.
func (s *stackService) ApplyStack(ctx context.Context, spec *models.Stack) (*models.Stack, bool, *errors.ServiceError) {
	existingStack, lookupErr := s.stackStore.GetByNameAndProjectID(ctx, spec.Name, spec.ProjectID)
	if lookupErr != nil && !lookupErr.Is404() {
		return nil, false, lookupErr
	}
	if existingStack != nil {
		if permErr := s.permissions.Check(ctx, existingStack.ProjectID, auth.ResourceStacks, existingStack.ID, auth.ActionWrite); permErr != nil {
			return nil, false, permErr
		}
		updatedStack, serr := s.InternalUpdateStack(ctx, existingStack.ID, spec)
		if serr != nil {
			return nil, false, serr
		}
		return updatedStack, false, nil
	}
	if permErr := s.permissions.Check(ctx, spec.ProjectID, auth.ResourceStacks, "", auth.ActionCreate); permErr != nil {
		return nil, false, permErr
	}
	createdStack, serr := s.InternalCreateStack(ctx, spec)
	if serr != nil {
		return nil, false, serr
	}
	return createdStack, true, nil
}

func (s *stackService) InternalCreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, _ := s.stackStore.GetByNameAndProjectID(ctx, spec.Name, spec.ProjectID)
	if existingStack != nil {
		return nil, errors.Conflict("stack with name '%s' already exists", spec.Name)
	}

	s.applyRuntimeResourceDefaults(spec.StackResources)
	s.logger.Info(ctx, "running validation for stack creation: %s", spec.Name)
	if err := s.stackValidator.ValidateForCreate(ctx, spec); err != nil {
		return nil, err
	}
	s.logger.Info(ctx, "validation passed for stack creation: %s", spec.Name)

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
	if s.runtimePolicy.DraftProvisioningMode() == ProvisioningModeEager {
		if err := s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, models.StackOperand{ID: createdStack.ID}); err != nil {
			return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", spec.Name, err.Error())
		}
		for _, v := range createdStack.Volumes {
			_ = s.BackgroundJobEnqueuer.EnqueueAfterCommit(ctx, models.VolumeOperand{ID: v.ID})
		}
	}
	s.logger.WithFields(map[string]interface{}{
		logger.FieldStackID:   createdStack.ID,
		logger.FieldClusterID: createdStack.ClusterID,
		"name":                createdStack.Name,
	}).Info(ctx, "created stack")
	return createdStack, nil
}

func (s *stackService) InternalCreateWithTx(ctx context.Context, spec *models.Stack, namespaceForStack *models.Namespace) (*models.Stack, *errors.ServiceError) {
	if err := s.runtimePolicy.AdmitStackMutationWithTx(ctx, StackMutation{
		Kind:                 StackMutationCreate,
		OrganisationID:       spec.OrganisationID,
		DesiredResourceCount: int64(len(spec.StackResources)),
	}); err != nil {
		return nil, err
	}
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

	var createdVolumes []*models.Volume
	for _, volume := range desiredVolumes {
		createdVolume, err := s.volumeService.InternalCreateWithTx(ctx, createdStack, volume)
		if err != nil {
			return nil, err
		}
		createdVolumes = append(createdVolumes, createdVolume)
	}
	createdStack.Volumes = createdVolumes

	for _, resource := range desiredResources {
		if _, err := s.stackResourceService.InternalCreateWithTx(ctx, createdStack, resource); err != nil {
			return nil, err
		}
	}

	// This is makes sure that we track all the explicit resources we use in the stack.
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
	if permErr := s.permissions.Check(ctx, existingStack.ProjectID, auth.ResourceStacks, ID, auth.ActionWrite); permErr != nil {
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
	spec.ProjectID = existingStack.ProjectID
	spec.UserID = existingStack.UserID
	s.applyRuntimeResourceDefaults(spec.StackResources)

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
	if s.runtimePolicy.DraftProvisioningMode() == ProvisioningModeEager {
		for _, v := range updatedStack.Volumes {
			if _, existed := existingVolumeIDs[v.ID]; !existed {
				if enqErr := s.BackgroundJobEnqueuer.Enqueue(models.VolumeOperand{ID: v.ID}); enqErr != nil {
					return nil, errors.GeneralError("failed to enqueue volume '%s': %s", v.ID, enqErr.Error())
				}
			}
		}
	}

	s.logger.WithField(logger.FieldStackID, updatedStack.ID).Info(ctx, "updated stack")
	return updatedStack, nil
}

func (s *stackService) UpdateStackShell(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingStack.ProjectID, auth.ResourceStacks, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	return s.InternalUpdateShellStack(ctx, ID, spec)
}

func (s *stackService) InternalUpdateShellStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError) {
	existingStack, err := s.InternalGetStack(ctx, ID)
	if err != nil {
		return nil, err
	}

	// The stack name is immutable: the cluster Stack CR is keyed by it, so a
	// rename would orphan the existing CR at the next release apply.
	if spec.Name != existingStack.Name {
		return nil, errors.BadRequest("stack name cannot be updated")
	}

	// set namespace
	spec.ID = existingStack.ID
	spec.Namespace = existingStack.Namespace
	spec.ClusterID = existingStack.ClusterID
	spec.OrganisationID = existingStack.OrganisationID
	spec.ProjectID = existingStack.ProjectID
	spec.UserID = existingStack.UserID

	// Strip children so only the stack's own columns are updated; connections
	// must NOT be replaced.
	spec.StackResources = nil
	spec.Volumes = nil
	spec.Connections = nil

	if verr := s.stackValidator.ValidateShell(ctx, spec); verr != nil {
		return nil, verr
	}

	var updatedStack *models.Stack
	err = s.stackStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedStack, err = s.InternalUpdateShellWithTx(ctx, spec, existingStack)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return updatedStack, nil
}

func (s *stackService) InternalUpdateShellWithTx(ctx context.Context, spec *models.Stack, existingStack *models.Stack) (*models.Stack, *errors.ServiceError) {
	if _, policyErr := s.runtimePolicy.AdmitComputeMutationWithTx(ctx, existingStack.OrganisationID); policyErr != nil {
		return nil, policyErr
	}
	updatedStack, updateErr := s.stackStore.UpdateShellWithTx(ctx, existingStack.ID, spec)
	if updateErr != nil {
		return nil, updateErr
	}
	return updatedStack, nil
}

func (s *stackService) InternalUpdateWithTx(ctx context.Context, spec *models.Stack, existingStack *models.Stack) (*models.Stack, *errors.ServiceError) {
	if err := s.runtimePolicy.AdmitStackMutationWithTx(ctx, StackMutation{
		Kind:                 StackMutationUpdate,
		OrganisationID:       existingStack.OrganisationID,
		StackID:              existingStack.ID,
		DesiredResourceCount: int64(len(spec.StackResources)),
	}); err != nil {
		return nil, err
	}
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

func (s *stackService) applyRuntimeResourceDefaults(resources []*models.StackResource) {
	for _, resource := range resources {
		if resource != nil {
			s.runtimePolicy.ApplyStackResourceDefaults(resource)
		}
	}
}

func (s *stackService) GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	stack, err := s.stackStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, ID, auth.ActionRead); permErr != nil {
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
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
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
		if _, policyErr := s.runtimePolicy.AdmitComputeMutationWithTx(ctx, lockedStack.OrganisationID); policyErr != nil {
			return policyErr
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

	if s.runtimePolicy.DraftProvisioningMode() == ProvisioningModeEager {
		if enqErr := s.BackgroundJobEnqueuer.Enqueue(models.VolumeOperand{ID: created.ID}); enqErr != nil {
			return nil, errors.GeneralError("failed to enqueue volume '%s': %s", created.ID, enqErr.Error())
		}
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
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, stackID, auth.ActionWrite); permErr != nil {
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
	if err := s.stackValidator.ValidateConnections(ctx, &desired); err != nil {
		return nil, nil, err
	}
	return stack, &desired, nil
}

func (s *stackService) createStackConnection(ctx context.Context, existingStack *models.Stack, connection *models.StackConnection) (*models.StackConnection, *errors.ServiceError) {
	var createdConnection *models.StackConnection
	if err := s.stackStore.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		if _, policyErr := s.runtimePolicy.AdmitComputeMutationWithTx(txCtx, existingStack.OrganisationID); policyErr != nil {
			return policyErr
		}
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
		if _, policyErr := s.runtimePolicy.AdmitComputeMutationWithTx(txCtx, existingStack.OrganisationID); policyErr != nil {
			return policyErr
		}
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

func (s *stackService) InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError) {
	stacks, err := s.stackStore.InternalList(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}

func (s *stackService) InternalListWorkloadAuthorityCandidates(ctx context.Context) ([]*models.Stack, *errors.ServiceError) {
	return s.stackStore.ListWorkloadAuthorityCandidates(ctx)
}

func (s *stackService) GetStacksByProjectID(ctx context.Context, projectID string) ([]*models.Stack, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, projectID, auth.ResourceStacks, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	stacks, err := s.stackStore.ListByProjectID(ctx, projectID)
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

	memberships, serr := s.projectService.InternalListUserProjects(ctx, identity.UserID, orgID)
	if serr != nil {
		return nil, serr
	}

	var allowedProjectIDs []string
	for _, m := range memberships {
		if permErr := s.permissions.Check(ctx, m.ProjectID, auth.ResourceStacks, "", auth.ActionList); permErr == nil {
			allowedProjectIDs = append(allowedProjectIDs, m.ProjectID)
		}
	}

	return s.stackStore.ListByProjectIDs(ctx, allowedProjectIDs)
}

func (s *stackService) DeleteStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError) {
	// GetStack includes read permission check
	stack, err := s.GetStack(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, stack.ProjectID, auth.ResourceStacks, ID, auth.ActionDelete); permErr != nil {
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
			if _, markErr := s.releaseService.MarkFailed(ctx, active.ID, "stack deleted", nil); markErr != nil {
				s.logger.Error(ctx, "failed to mark release '%s' failed for deleted stack '%s': %s", active.ID, stack.Name, markErr.Error())
			}
		}
	}
	stack.DeletionTimestamp = ptr.To(time.Now().UTC())
	stack.Status.State = models.StackDeleting
	stack.Status.Message = "Stack is being deleted"
	stackMarkedForDelete, err := s.stackStore.UpdateForDelete(ctx, stack.ID, stack)
	if err != nil {
		return nil, errors.GeneralError("failed to update stack '%s' for deletion: %s", stack.Name, err.Error())
	}
	if err := s.BackgroundJobEnqueuer.Enqueue(models.StackOperand{ID: stack.ID}); err != nil {
		return nil, errors.GeneralError("failed to enqueue background job for stack '%s': %s", stack.Name, err.Error())
	}
	s.logger.WithField(logger.FieldStackID, stack.ID).Info(ctx, "marked stack for deletion")
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
