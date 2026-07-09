package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// TestCreateStackConnection_UsesConnectionScopedValidation asserts that
// connection CRUD validates via the narrow ValidateConnections gate rather
// than the full ValidateForUpdate pass, so a pre-existing unrelated stack
// invalidity can't block a connection edit.
func TestCreateStackConnection_UsesConnectionScopedValidation(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	teamID := "team-456"
	stack := &models.Stack{
		ID:     stackID,
		TeamID: teamID,
		StackResources: []*models.StackResource{
			{Name: "web"},
		},
	}
	newConnection := &models.StackConnection{
		ID:   "internal-api",
		Kind: models.ConnectionKindEnv,
		From: models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
		To:   models.TopologyNodeRef{Type: models.TopologyNodeTypeStackResource, Name: "web"},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStackStore := mocks.NewMockStackStore(ctrl)
	mockPermissions := mocks.NewMockPermissionService(ctrl)
	mockValidator := mocks.NewMockStackValidator(ctrl)
	mockReferenceService := mocks.NewMockReferenceService(ctrl)

	svc := &stackService{
		stackStore:       mockStackStore,
		permissions:      mockPermissions,
		stackValidator:   mockValidator,
		referenceService: mockReferenceService,
	}

	mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
	mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
	mockPermissions.EXPECT().Check(ctx, teamID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)

	// ValidateConnections is the narrow gate under test; ValidateForUpdate
	// must never be called for a connection-only mutation.
	mockValidator.EXPECT().
		ValidateConnections(ctx, gomock.Any()).
		DoAndReturn(func(_ context.Context, spec *models.Stack) *errors.ServiceError {
			assert.Len(t, spec.Connections, 1)
			assert.Equal(t, "internal-api", spec.Connections[0].ID)
			return nil
		})

	created := *newConnection
	mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		})
	mockStackStore.EXPECT().CreateConnectionWithTx(ctx, stackID, newConnection).Return(&created, nil)
	mockReferenceService.EXPECT().ReprojectSpec(ctx, stackID).Return(nil)

	got, serr := svc.CreateStackConnection(ctx, stackID, newConnection)
	assert.Nil(t, serr)
	assert.Equal(t, "internal-api", got.ID)
}
