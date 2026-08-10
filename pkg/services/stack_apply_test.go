package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type applyStackTestEnv struct {
	svc               *stackService
	stackStore        *mocks.MockStackStore
	permissions       *mocks.MockPermissionService
	validator         *mocks.MockStackValidator
	namespaceService  *mocks.MockNamespaceService
	clusterService    *mocks.MockClusterService
	volumeService     *mocks.MockVolumeService
	resourceService   *mocks.MockStackResourceService
	referenceService  *mocks.MockReferenceService
	backgroundEnqueue *mocks.MockBackgroundJobEnqueuer
}

func newApplyStackTestEnv(ctrl *gomock.Controller) *applyStackTestEnv {
	env := &applyStackTestEnv{
		stackStore:        mocks.NewMockStackStore(ctrl),
		permissions:       mocks.NewMockPermissionService(ctrl),
		validator:         mocks.NewMockStackValidator(ctrl),
		namespaceService:  mocks.NewMockNamespaceService(ctrl),
		clusterService:    mocks.NewMockClusterService(ctrl),
		volumeService:     mocks.NewMockVolumeService(ctrl),
		resourceService:   mocks.NewMockStackResourceService(ctrl),
		referenceService:  mocks.NewMockReferenceService(ctrl),
		backgroundEnqueue: mocks.NewMockBackgroundJobEnqueuer(ctrl),
	}
	env.svc = &stackService{
		stackStore:           env.stackStore,
		permissions:          env.permissions,
		stackValidator:       env.validator,
		namespaceService:     env.namespaceService,
		clusterService:       env.clusterService,
		volumeService:        env.volumeService,
		stackResourceService: env.resourceService,
		referenceService:     env.referenceService,
		defaultingService:    NewStackDefaultingService(),
		logger:               logger.NewLogger(),
		runtimePolicy:        NewSelfHostedRuntimePolicy(),
		BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
			BackgroundJobEnqueuer: env.backgroundEnqueue,
		},
	}
	return env
}

func TestApplyStack_CloudDraftDoesNotEnqueueCompute(t *testing.T) {
	ctx := context.Background()
	spec := &models.Stack{Name: "demo", ProjectID: "project-1", OrganisationID: "org-1", UserID: "user-1"}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)
	env.svc.runtimePolicy = NewStackdomeCloudRuntimePolicy(&fakeCloudTrialService{})

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", "project-1").Return(nil, errors.NotFound("missing")).Times(2)
	env.permissions.EXPECT().Check(ctx, "project-1", auth.ResourceStacks, "", auth.ActionCreate).Return(nil)
	env.validator.EXPECT().ValidateForCreate(ctx, spec).Return(nil)
	namespace := &models.Namespace{Name: "ns-demo"}
	env.namespaceService.EXPECT().PrepareNamespaceForStack(ctx, spec).Return(namespace, nil)
	env.clusterService.EXPECT().GetClusterForOrg(ctx, "org-1").Return(&models.Cluster{ID: "cluster-1"}, nil)
	env.stackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		})
	env.namespaceService.EXPECT().CreateInDBWithTx(ctx, namespace).Return(&models.Namespace{ID: "ns-1", Name: "ns-demo"}, nil)
	createdShell := &models.Stack{ID: "stack-1", Name: "demo", ProjectID: "project-1"}
	env.stackStore.EXPECT().CreateWithTx(ctx, gomock.Any()).Return(createdShell, nil)
	env.referenceService.EXPECT().ReprojectSpec(ctx, "stack-1").Return(nil)
	created := &models.Stack{ID: "stack-1", Name: "demo", ProjectID: "project-1"}
	env.stackStore.EXPECT().GetByID(ctx, "stack-1").Return(created, nil)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, serr)
	assert.True(t, wasCreated)
	assert.Equal(t, created, got)
}

// TestApplyStack_CreatesWhenMissing: no stack with the spec's name exists in
// the project, so apply creates the stack (and children) atomically and reports
// created=true.
func TestApplyStack_CreatesWhenMissing(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	orgID := "org-1"
	userID := "user-1"
	spec := &models.Stack{Name: "demo", ProjectID: projectID, OrganisationID: orgID, UserID: userID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).
		Return(nil, errors.NotFound("stack with name demo not found"))
	env.permissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, "", auth.ActionCreate).Return(nil)

	// InternalCreateStack path: its own project-scoped duplicate re-check.
	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).
		Return(nil, errors.NotFound("stack with name demo not found"))
	env.validator.EXPECT().ValidateForCreate(ctx, spec).Return(nil)
	namespace := &models.Namespace{Name: "ns-demo"}
	env.namespaceService.EXPECT().PrepareNamespaceForStack(ctx, spec).Return(namespace, nil)
	env.clusterService.EXPECT().GetClusterForOrg(ctx, orgID).Return(&models.Cluster{ID: "cluster-1"}, nil)
	env.stackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		})
	env.namespaceService.EXPECT().CreateInDBWithTx(ctx, namespace).
		Return(&models.Namespace{ID: "ns-1", Name: "ns-demo"}, nil)
	createdShell := &models.Stack{ID: "stack-1", Name: "demo", ProjectID: projectID}
	env.stackStore.EXPECT().CreateWithTx(ctx, gomock.Any()).Return(createdShell, nil)
	env.referenceService.EXPECT().ReprojectSpec(ctx, "stack-1").Return(nil)
	created := &models.Stack{ID: "stack-1", Name: "demo", ProjectID: projectID}
	env.stackStore.EXPECT().GetByID(ctx, "stack-1").Return(created, nil)
	env.backgroundEnqueue.EXPECT().EnqueueAfterCommit(ctx, models.StackOperand{ID: "stack-1"}).Return(nil)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, serr)
	assert.True(t, wasCreated)
	assert.Equal(t, created, got)
}

// TestApplyStack_UpdatesWhenExists: a stack with the spec's name exists in the
// project, so apply resolves its ID and delegates to the full-replacement update
// path, reporting created=false.
func TestApplyStack_UpdatesWhenExists(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	stackID := "stack-1"
	existing := &models.Stack{ID: stackID, Name: "demo", ProjectID: projectID, Namespace: "ns-demo", ClusterID: "cluster-1"}
	spec := &models.Stack{Name: "demo", ProjectID: projectID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).Return(existing, nil)
	env.permissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)

	// InternalUpdateStack path, addressed by the resolved existing ID.
	env.stackStore.EXPECT().GetByID(ctx, stackID).Return(existing, nil)
	env.validator.EXPECT().ValidateForUpdate(ctx, existing, spec).
		DoAndReturn(func(_ context.Context, _, got *models.Stack) *errors.ServiceError {
			// The spec inherits identity from the resolved stack.
			assert.Equal(t, stackID, got.ID)
			return nil
		})
	env.stackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		})
	updatedShell := &models.Stack{ID: stackID, Name: "demo", ProjectID: projectID}
	env.stackStore.EXPECT().UpdateWithTx(ctx, stackID, gomock.Any()).Return(updatedShell, nil)
	env.volumeService.EXPECT().InternalSyncVolumesWithTx(ctx, updatedShell, existing, gomock.Any()).Return(nil)
	env.volumeService.EXPECT().ListVolumesUsedByStack(ctx, stackID).Return(nil, nil)
	env.resourceService.EXPECT().InternalSyncResourcesWithTx(ctx, updatedShell, existing, gomock.Any()).Return(nil)
	env.referenceService.EXPECT().ReprojectSpec(ctx, stackID).Return(nil)
	updated := &models.Stack{ID: stackID, Name: "demo", ProjectID: projectID}
	env.stackStore.EXPECT().GetByID(ctx, stackID).Return(updated, nil)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, serr)
	assert.False(t, wasCreated)
	assert.Equal(t, updated, got)
}

// TestApplyStack_ConflictOnDuplicateNameRecheck: the apply-level project-scoped
// lookup misses, but the create path's own sequential duplicate re-check finds
// a stack with the same name in the project — the Conflict propagates untouched.
// (The true concurrent race is covered by the DB unique index on
// stacks(project_id, name); see the StackStore Create/CreateWithTx Conflict
// mapping tests in pkg/stores/pgstore.)
func TestApplyStack_ConflictOnDuplicateNameRecheck(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	userID := "user-1"
	spec := &models.Stack{Name: "demo", ProjectID: projectID, UserID: userID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).
		Return(nil, errors.NotFound("stack with name demo not found"))
	env.permissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, "", auth.ActionCreate).Return(nil)
	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).
		Return(&models.Stack{ID: "stack-raced", Name: "demo", ProjectID: projectID}, nil)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, got)
	assert.False(t, wasCreated)
	assert.NotNil(t, serr)
	assert.True(t, serr.IsConflict())
}

// TestApplyStack_PermissionDeniedOnCreate: missing stack + no stacks:create
// permission short-circuits before any write.
func TestApplyStack_PermissionDeniedOnCreate(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	spec := &models.Stack{Name: "demo", ProjectID: projectID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).
		Return(nil, errors.NotFound("stack with name demo not found"))
	denied := errors.Forbidden("missing stacks:create")
	env.permissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, "", auth.ActionCreate).Return(denied)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, got)
	assert.False(t, wasCreated)
	assert.Equal(t, denied, serr)
}

// TestApplyStack_PermissionDeniedOnUpdate: existing stack + no stacks:write
// permission short-circuits before any write.
func TestApplyStack_PermissionDeniedOnUpdate(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	stackID := "stack-1"
	existing := &models.Stack{ID: stackID, Name: "demo", ProjectID: projectID}
	spec := &models.Stack{Name: "demo", ProjectID: projectID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).Return(existing, nil)
	denied := errors.Forbidden("missing stacks:write")
	env.permissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(denied)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, got)
	assert.False(t, wasCreated)
	assert.Equal(t, denied, serr)
}

// TestApplyStack_LookupErrorPropagates: a non-404 lookup failure aborts the
// apply instead of being misread as "stack missing".
func TestApplyStack_LookupErrorPropagates(t *testing.T) {
	ctx := context.Background()
	projectID := "project-1"
	spec := &models.Stack{Name: "demo", ProjectID: projectID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	env := newApplyStackTestEnv(ctrl)

	lookupErr := errors.GeneralError("db unavailable")
	env.stackStore.EXPECT().GetByNameAndProjectID(ctx, "demo", projectID).Return(nil, lookupErr)

	got, wasCreated, serr := env.svc.ApplyStack(ctx, spec)
	assert.Nil(t, got)
	assert.False(t, wasCreated)
	assert.Equal(t, lookupErr, serr)
}
