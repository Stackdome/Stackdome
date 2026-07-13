package services

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStackResourceService_Restart(t *testing.T) {
	t.Run("successful restart", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
		}

		ctx := context.Background()
		stackID := "stack-123"
		resourceName := "web-server"
		projectID := "project-456"

		stack := &models.Stack{
			ID:        stackID,
			ProjectID: projectID,
		}

		resource := &models.StackResource{
			ID:              "resource-789",
			StackID:         stackID,
			Name:            resourceName,
			LifecycleConfig: nil,
		}

		updatedResource := &models.StackResource{
			ID:      resource.ID,
			StackID: stackID,
			Name:    resourceName,
			LifecycleConfig: &models.LifecycleConfig{
				RestartRequestTime: func() *time.Time { t := time.Now().UTC(); return &t }(),
			},
		}

		mockStackStore.EXPECT().
			GetByID(ctx, stackID).
			Return(stack, nil)

		mockPermissions.EXPECT().
			Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).
			Return(nil)

		mockResourceStore.EXPECT().
			GetByStackIDAndResourceName(ctx, stackID, resourceName).
			Return(resource, nil)

		mockResourceStore.EXPECT().
			Update(ctx, resource.ID, gomock.Any(), stack).
			DoAndReturn(func(ctx context.Context, id string, res *models.StackResource, stk *models.Stack) (*models.StackResource, *errors.ServiceError) {
				assert.NotNil(t, res.LifecycleConfig)
				assert.NotNil(t, res.LifecycleConfig.RestartRequestTime)
				return updatedResource, nil
			})

		result, err := svc.Restart(ctx, stackID, resourceName)

		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedResource.ID, result.ID)
		assert.NotNil(t, result.LifecycleConfig)
		assert.NotNil(t, result.LifecycleConfig.RestartRequestTime)
	})

	t.Run("permission denied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
		}

		ctx := context.Background()
		stackID := "stack-123"
		resourceName := "web-server"
		projectID := "project-456"

		stack := &models.Stack{
			ID:        stackID,
			ProjectID: projectID,
		}

		mockStackStore.EXPECT().
			GetByID(ctx, stackID).
			Return(stack, nil)

		mockPermissions.EXPECT().
			Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).
			Return(errors.Forbidden("permission denied"))

		result, err := svc.Restart(ctx, stackID, resourceName)

		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.Equal(t, "error: permission denied", err.Error())
	})

	t.Run("resource not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
		}

		ctx := context.Background()
		stackID := "stack-123"
		resourceName := "nonexistent-resource"
		projectID := "project-456"

		stack := &models.Stack{
			ID:        stackID,
			ProjectID: projectID,
		}

		mockStackStore.EXPECT().
			GetByID(ctx, stackID).
			Return(stack, nil)

		mockPermissions.EXPECT().
			Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).
			Return(nil)

		mockResourceStore.EXPECT().
			GetByStackIDAndResourceName(ctx, stackID, resourceName).
			Return(nil, errors.NotFound("resource not found"))

		result, err := svc.Restart(ctx, stackID, resourceName)

		assert.Nil(t, result)
		assert.NotNil(t, err)
		assert.Equal(t, "error: resource not found", err.Error())
	})
}
