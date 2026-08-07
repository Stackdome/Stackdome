package services

import (
	"context"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
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

var _ = ginkgo.Describe("stackResourceService workload type defaulting", func() {
	var (
		ctrl              *gomock.Controller
		mockResourceStore *mocks.MockStackResourceStore
		mockDomains       *mocks.MockStackDomainsService
		svc               *stackResourceService
		ctx               context.Context
		stack             *models.Stack
	)

	datastoreSpec := func() *models.StackResource {
		return &models.StackResource{
			Name:         "db",
			WorkloadType: models.WorkloadTypeService,
			ImageConfig:  &models.ImageConfigSpec{Image: "postgres:16"},
		}
	}

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockResourceStore = mocks.NewMockStackResourceStore(ctrl)
		mockDomains = mocks.NewMockStackDomainsService(ctrl)
		svc = &stackResourceService{
			stackResourceStore: mockResourceStore,
			domainNameService:  mockDomains,
		}
		ctx = context.Background()
		stack = &models.Stack{ID: "stack-1", UserID: "user-1", Namespace: "ns-1"}

		mockDomains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, gomock.Any()).
			Return(nil)
		mockResourceStore.EXPECT().
			UpdatePortsWithTx(ctx, gomock.Any(), gomock.Any()).
			Return(nil)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.It("promotes a datastore image to StatefulService on create", func() {
		var persisted *models.StackResource
		mockResourceStore.EXPECT().
			CreateWithTx(ctx, gomock.Any(), stack).
			DoAndReturn(func(_ context.Context, resource *models.StackResource, _ *models.Stack) (*models.StackResource, *errors.ServiceError) {
				persisted = resource
				return resource, nil
			})

		_, err := svc.InternalCreateWithTx(ctx, stack, datastoreSpec())

		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(persisted.WorkloadType).To(gomega.Equal(models.WorkloadTypeStatefulService))
	})

	ginkgo.It("re-promotes a datastore image to StatefulService on update", func() {
		var persisted *models.StackResource
		mockResourceStore.EXPECT().
			UpdateWithTx(ctx, "resource-1", gomock.Any(), stack).
			DoAndReturn(func(_ context.Context, _ string, resource *models.StackResource, _ *models.Stack) (*models.StackResource, *errors.ServiceError) {
				persisted = resource
				return resource, nil
			})

		_, err := svc.InternalUpdateWithTx(ctx, stack, "resource-1", datastoreSpec())

		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(persisted.WorkloadType).To(gomega.Equal(models.WorkloadTypeStatefulService))
	})

	ginkgo.It("keeps a non-datastore image a Service on update", func() {
		var persisted *models.StackResource
		mockResourceStore.EXPECT().
			UpdateWithTx(ctx, "resource-1", gomock.Any(), stack).
			DoAndReturn(func(_ context.Context, _ string, resource *models.StackResource, _ *models.Stack) (*models.StackResource, *errors.ServiceError) {
				persisted = resource
				return resource, nil
			})

		spec := datastoreSpec()
		spec.Name = "api"
		spec.ImageConfig = &models.ImageConfigSpec{Image: "nginx:1.27"}

		_, err := svc.InternalUpdateWithTx(ctx, stack, "resource-1", spec)

		gomega.Expect(err).To(gomega.BeNil())
		gomega.Expect(persisted.WorkloadType).To(gomega.Equal(models.WorkloadTypeService))
	})
})
