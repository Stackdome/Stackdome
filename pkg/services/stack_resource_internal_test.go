package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestStackResourceService_InternalCreateWithTx(t *testing.T) {
	t.Run("persists resource and finalizes exposed ports", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockDomains := mocks.NewMockStackDomainsService(ctrl)

		svc := &stackResourceService{
			stackResourceStore: mockResourceStore,
			domainNameService:  mockDomains,
		}

		ctx := context.Background()
		stack := &models.Stack{
			ID:             "stack-1",
			UserID:         "user-1",
			Namespace:      "ns-1",
			OrganisationID: "org-1",
		}
		input := &models.StackResource{
			Name: "web",
			Ports: models.Ports{
				{Number: 8080, ExposedToPublic: true},
			},
		}
		created := &models.StackResource{
			ID:      "resource-1",
			StackID: stack.ID,
			Name:    "web",
			Ports:   input.Ports,
		}
		finalized := &models.StackResource{
			ID:      created.ID,
			StackID: stack.ID,
			Name:    "web",
			Ports: models.Ports{
				{
					Number:          8080,
					ExposedToPublic: true,
					Protocol:        "http",
					ExposedFqdn:     "abc.web.example.com",
				},
			},
		}

		mockResourceStore.EXPECT().
			CreateWithTx(ctx, gomock.Any(), stack).
			DoAndReturn(func(_ context.Context, resource *models.StackResource, _ *models.Stack) (*models.StackResource, *errors.ServiceError) {
				assert.Equal(t, "stack-1", resource.StackID)
				assert.Equal(t, "http", resource.Ports[0].Protocol)
				return created, nil
			})

		mockDomains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, created).
			DoAndReturn(func(_ context.Context, _ *models.Stack, resource *models.StackResource) *errors.ServiceError {
				resource.Ports[0].ExposedFqdn = finalized.Ports[0].ExposedFqdn
				resource.Ports[0].Protocol = finalized.Ports[0].Protocol
				return nil
			})

		mockResourceStore.EXPECT().
			UpdatePortsWithTx(ctx, created.ID, gomock.Any()).
			Return(nil)

		result, err := svc.InternalCreateWithTx(ctx, stack, input)

		require.Nil(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "abc.web.example.com", result.Ports[0].ExposedFqdn)
	})

	t.Run("skips port update when resource has no exposed ports", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockDomains := mocks.NewMockStackDomainsService(ctrl)

		svc := &stackResourceService{
			stackResourceStore: mockResourceStore,
			domainNameService:  mockDomains,
		}

		ctx := context.Background()
		stack := &models.Stack{ID: "stack-1", UserID: "user-1", Namespace: "ns-1"}
		input := &models.StackResource{Name: "worker", Ports: models.Ports{{Number: 8080}}}
		created := &models.StackResource{ID: "resource-2", Name: "worker", Ports: input.Ports}

		mockResourceStore.EXPECT().
			CreateWithTx(ctx, gomock.Any(), stack).
			Return(created, nil)

		mockDomains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, created).
			Return(nil)

		result, err := svc.InternalCreateWithTx(ctx, stack, input)

		require.Nil(t, err)
		assert.Equal(t, created.ID, result.ID)
	})

	t.Run("works without domain service configured", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		svc := &stackResourceService{stackResourceStore: mockResourceStore}

		ctx := context.Background()
		stack := &models.Stack{ID: "stack-1", UserID: "user-1", Namespace: "ns-1"}
		input := &models.StackResource{Name: "web"}
		created := &models.StackResource{ID: "resource-3", Name: "web"}

		mockResourceStore.EXPECT().
			CreateWithTx(ctx, gomock.Any(), stack).
			Return(created, nil)

		result, err := svc.InternalCreateWithTx(ctx, stack, input)

		require.Nil(t, err)
		assert.Equal(t, created, result)
	})
}

func TestStackResourceService_InternalSyncResourcesWithTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
	svc := &stackResourceService{stackResourceStore: mockResourceStore}

	ctx := context.Background()
	stack := &models.Stack{ID: "stack-1", UserID: "user-1", Namespace: "ns-1"}
	existingStack := &models.Stack{
		ID: "stack-1",
		StackResources: []*models.StackResource{
			{ID: "res-keep", Name: "api"},
			{ID: "res-drop", Name: "old"},
		},
	}
	desired := []*models.StackResource{
		{Name: "api", Ports: models.Ports{{Number: 80}}},
		{Name: "web"},
	}

	mockResourceStore.EXPECT().
		UpdateWithTx(ctx, "res-keep", gomock.Any(), stack).
		Return(&models.StackResource{ID: "res-keep", Name: "api"}, nil)

	mockResourceStore.EXPECT().
		CreateWithTx(ctx, gomock.Any(), stack).
		Return(&models.StackResource{ID: "res-new", Name: "web"}, nil)

	mockResourceStore.EXPECT().
		DeleteWithTx(ctx, "res-drop").
		Return(nil)

	err := svc.InternalSyncResourcesWithTx(ctx, stack, existingStack, desired)

	assert.Nil(t, err)
}

func TestStackResourceService_InternalDeleteWithTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
	mockDomains := mocks.NewMockStackDomainsService(ctrl)

	svc := &stackResourceService{
		stackResourceStore: mockResourceStore,
		domainNameService:  mockDomains,
	}

	ctx := context.Background()
	resourceID := "res-drop"

	mockDomains.EXPECT().
		InternalDeleteDomainsForResourceWithTx(ctx, resourceID).
		Return(nil)

	mockResourceStore.EXPECT().
		DeleteWithTx(ctx, resourceID).
		Return(nil)

	err := svc.InternalDeleteWithTx(ctx, resourceID)

	assert.Nil(t, err)
}
