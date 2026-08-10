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

func TestStackService_CreateStackVolume(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	projectID := "project-456"
	stack := &models.Stack{ID: stackID, ProjectID: projectID, Volumes: []*models.Volume{{ID: "v-0", Name: "existing-data"}}}
	newVolume := &models.Volume{Name: "web-data"}

	t.Run("creates, associates and enqueues the volume", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockVolumeService := mocks.NewMockVolumeService(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockEnqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)

		svc := &stackService{
			runtimePolicy: NewSelfHostedRuntimePolicy(),
			stackStore:    mockStackStore,
			volumeService: mockVolumeService,
			permissions:   mockPermissions,
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: mockEnqueuer,
			},
		}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockVolumeService.EXPECT().PrepareForCreate(ctx, newVolume).Return(nil)
		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		// Inside the transaction: lock, then re-read for the duplicate check.
		mockStackStore.EXPECT().LockByID(ctx, stackID).Return(nil)
		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		created := &models.Volume{ID: "v-1", Name: "web-data"}
		mockVolumeService.EXPECT().InternalCreateWithTx(ctx, stack, newVolume).Return(created, nil)
		mockEnqueuer.EXPECT().Enqueue(models.VolumeOperand{ID: "v-1"}).Return(nil)

		got, serr := svc.CreateStackVolume(ctx, stackID, newVolume)
		assert.Nil(t, serr)
		assert.Equal(t, "v-1", got.ID)
	})

	t.Run("rejects a duplicate volume name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		svc := &stackService{stackStore: mockStackStore, permissions: mockPermissions, runtimePolicy: NewSelfHostedRuntimePolicy()}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		_, serr := svc.CreateStackVolume(ctx, stackID, &models.Volume{Name: "existing-data"})
		assert.NotNil(t, serr)
		assert.Equal(t, errors.ErrorConflict, serr.Code)
	})

	t.Run("eagerly provisions a cloud volume", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockVolumeService := mocks.NewMockVolumeService(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockEnqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)
		svc := &stackService{
			runtimePolicy: newCloudRuntimePolicyForTest(),
			stackStore:    mockStackStore, volumeService: mockVolumeService, permissions: mockPermissions,
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{BackgroundJobEnqueuer: mockEnqueuer},
		}
		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionRead).Return(nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockVolumeService.EXPECT().PrepareForCreate(ctx, newVolume).Return(nil)
		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		mockStackStore.EXPECT().LockByID(ctx, stackID).Return(nil)
		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		created := &models.Volume{ID: "v-1", Name: "web-data"}
		mockVolumeService.EXPECT().InternalCreateWithTx(ctx, stack, newVolume).Return(created, nil)
		mockEnqueuer.EXPECT().Enqueue(models.VolumeOperand{ID: "v-1"}).Return(nil)

		got, serr := svc.CreateStackVolume(ctx, stackID, newVolume)
		assert.Nil(t, serr)
		assert.Equal(t, created, got)
	})
}
