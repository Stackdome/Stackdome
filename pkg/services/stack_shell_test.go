package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStackShellFrom(t *testing.T) {
	spec := &models.Stack{
		Name: "demo",
		Volumes: []*models.Volume{
			{Name: "data"},
		},
		StackResources: []*models.StackResource{
			{Name: "web"},
		},
	}

	shell := stackShellFrom(spec)

	assert.Equal(t, "demo", shell.Name)
	assert.Nil(t, shell.Volumes)
	assert.Nil(t, shell.StackResources)
}

// TestInternalUpdateShellStack_RejectsInvalidSettings asserts the shell
// update path runs the shell-scoped stack validation: when the validator
// rejects the settings, nothing is written.
func TestInternalUpdateShellStack_RejectsInvalidSettings(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	existing := &models.Stack{ID: stackID, Name: "demo"}
	spec := &models.Stack{
		Name:     "demo",
		Settings: &models.StackSettings{ReleaseRetentionLimit: models.MaxReleaseRetentionLimit + 1},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStackStore := mocks.NewMockStackStore(ctrl)
	mockValidator := mocks.NewMockStackValidator(ctrl)

	svc := &stackService{
		runtimePolicy:  NewSelfHostedRuntimePolicy(),
		stackStore:     mockStackStore,
		stackValidator: mockValidator,
	}

	mockStackStore.EXPECT().GetByID(ctx, stackID).Return(existing, nil)
	verr := errors.ValidationFailed([]errors.FieldError{{
		Field: "spec.settings", Code: errors.VErrStackSettingsInvalid, Message: "release_retention_limit must be at most 50",
	}})
	mockValidator.EXPECT().ValidateShell(ctx, spec).Return(verr)
	// No WithTransaction / UpdateShellWithTx expectations: any write attempt
	// fails via gomock's controller.

	got, serr := svc.InternalUpdateShellStack(ctx, stackID, spec)
	assert.Nil(t, got)
	assert.Equal(t, verr, serr)
}

// TestInternalUpdateShellStack_ValidSettings_Updates asserts the happy path:
// validation passes (on the child-stripped spec) and the shell is written.
func TestInternalUpdateShellStack_ValidSettings_Updates(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	existing := &models.Stack{ID: stackID, Name: "demo", Namespace: "ns-demo"}
	spec := &models.Stack{
		Name:     "demo",
		Settings: &models.StackSettings{ReleaseRetentionLimit: 20, MinSuccessfulReleases: 5},
		StackResources: []*models.StackResource{
			{Name: "web"},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStackStore := mocks.NewMockStackStore(ctrl)
	mockValidator := mocks.NewMockStackValidator(ctrl)

	svc := &stackService{
		runtimePolicy:  NewSelfHostedRuntimePolicy(),
		stackStore:     mockStackStore,
		stackValidator: mockValidator,
	}

	mockStackStore.EXPECT().GetByID(ctx, stackID).Return(existing, nil)
	mockValidator.EXPECT().ValidateShell(ctx, spec).
		DoAndReturn(func(_ context.Context, got *models.Stack) *errors.ServiceError {
			// Children are stripped before validation; the shell path never
			// validates (or writes) them.
			assert.Nil(t, got.StackResources)
			assert.Nil(t, got.Volumes)
			assert.Nil(t, got.Connections)
			return nil
		})
	mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		})
	updated := &models.Stack{ID: stackID, Name: "demo"}
	mockStackStore.EXPECT().UpdateShellWithTx(ctx, stackID, spec).Return(updated, nil)

	got, serr := svc.InternalUpdateShellStack(ctx, stackID, spec)
	assert.Nil(t, serr)
	assert.Equal(t, updated, got)
}

// TestInternalUpdateShellStack_RenameRejected asserts the stack name is
// immutable on the shell path: the cluster Stack CR is keyed by name, so a
// rename would orphan it at the next release apply.
func TestInternalUpdateShellStack_RenameRejected(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	existing := &models.Stack{ID: stackID, Name: "demo", Namespace: "ns-demo"}
	spec := &models.Stack{Name: "renamed"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStackStore := mocks.NewMockStackStore(ctrl)
	mockValidator := mocks.NewMockStackValidator(ctrl)

	svc := &stackService{
		runtimePolicy:  NewSelfHostedRuntimePolicy(),
		stackStore:     mockStackStore,
		stackValidator: mockValidator,
	}

	mockStackStore.EXPECT().GetByID(ctx, stackID).Return(existing, nil)
	// No ValidateShell / WithTransaction expectations: the rename is rejected
	// before validation or any write.

	got, serr := svc.InternalUpdateShellStack(ctx, stackID, spec)
	assert.Nil(t, got)
	assert.NotNil(t, serr)
	assert.Equal(t, "stack name cannot be updated", serr.Reason)
}
