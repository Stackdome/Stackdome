package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestStackResourceService_Create_Validation(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	projectID := "project-456"
	stack := &models.Stack{ID: stackID, ProjectID: projectID, OrganisationID: "org-1"}

	t.Run("field errors from the validator are returned as a 400 ValidationFailed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
			resourceValidator:  mockValidator,
			runtimePolicy:      NewSelfHostedRuntimePolicy(),
		}

		resource := &models.StackResource{StackID: stackID, Name: "web"}
		siblings := []*models.StackResource{{StackID: stackID, Name: "worker"}}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockResourceStore.EXPECT().GetByStackID(ctx, stackID).Return(siblings, nil)
		fieldErrs := []errors.FieldError{
			{Field: "ports[0].expose_to_public", Code: errors.VErrPublicPortNotHTTP},
			{Field: "depends_on[0]", Code: errors.VErrDependencyUnknown},
		}
		mockValidator.EXPECT().Validate(ctx, stack, resource, siblings).Return(fieldErrs, nil)

		got, serr := svc.Create(ctx, resource)

		assert.Nil(t, got)
		if assert.NotNil(t, serr) {
			assert.Equal(t, errors.ErrorValidation, serr.Code)
			details, ok := serr.Details.(errors.ValidationErrorDetails)
			if assert.True(t, ok, "expected Details to be errors.ValidationErrorDetails") {
				codes := []string{details.Errors[0].Code, details.Errors[1].Code}
				assert.Contains(t, codes, errors.VErrPublicPortNotHTTP)
				assert.Contains(t, codes, errors.VErrDependencyUnknown)
			}
		}
	})

	t.Run("valid resource is created", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
		mockReferenceService := mocks.NewMockReferenceService(ctrl)
		mockDomains := mocks.NewMockStackDomainsService(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
			resourceValidator:  mockValidator,
			runtimePolicy:      newCloudRuntimePolicyForTest(),
			referenceService:   mockReferenceService,
			domainNameService:  mockDomains,
		}

		replicas := int32(9)
		resource := &models.StackResource{StackID: stackID, Name: "web", Replicas: &replicas}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockResourceStore.EXPECT().GetByStackID(ctx, stackID).Return(nil, nil)
		mockValidator.EXPECT().Validate(ctx, stack, resource, ([]*models.StackResource)(nil)).DoAndReturn(
			func(_ context.Context, _ *models.Stack, got *models.StackResource, _ []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
				assert.Equal(t, int32(1), *got.Replicas)
				return nil, nil
			})

		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})

		created := &models.StackResource{ID: "resource-1", StackID: stackID, Name: "web"}
		mockResourceStore.EXPECT().CreateWithTx(ctx, resource, stack).Return(created, nil)
		mockDomains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, created).
			Return(nil)
		mockResourceStore.EXPECT().UpdatePortsWithTx(ctx, created.ID, created).Return(nil)
		mockReferenceService.EXPECT().ReprojectSpec(ctx, stackID).Return(nil)
		mockResourceStore.EXPECT().GetByID(ctx, created.ID).Return(created, nil)

		got, serr := svc.Create(ctx, resource)

		assert.Nil(t, serr)
		if assert.NotNil(t, got) {
			assert.Equal(t, created.ID, got.ID)
		}
	})

	t.Run("cloud resource limit rejects before persistence", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
			resourceValidator:  mockValidator,
			runtimePolicy: newCloudRuntimePolicyWithStoreForTest(&fakeStackLimitStore{
				usage: stores.StackUsage{StackResourceCount: 6},
			}),
		}
		resource := &models.StackResource{StackID: stackID, Name: "web"}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockResourceStore.EXPECT().GetByStackID(ctx, stackID).Return(nil, nil)
		mockValidator.EXPECT().Validate(ctx, stack, resource, ([]*models.StackResource)(nil)).Return(nil, nil)
		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})

		got, serr := svc.Create(ctx, resource)

		assert.Nil(t, got)
		if assert.NotNil(t, serr) {
			assert.Contains(t, serr.Reason, "maximum of 6 stack resources")
		}
	})

	t.Run("infra ServiceError from the validator is propagated and the store is never called", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
			resourceValidator:  mockValidator,
			runtimePolicy:      NewSelfHostedRuntimePolicy(),
		}

		resource := &models.StackResource{StackID: stackID, Name: "web"}
		infraErr := errors.GeneralError("db unreachable")

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockResourceStore.EXPECT().GetByStackID(ctx, stackID).Return(nil, nil)
		mockValidator.EXPECT().Validate(ctx, stack, resource, ([]*models.StackResource)(nil)).Return(nil, infraErr)
		// CreateWithTx must never be called.
		mockResourceStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		got, serr := svc.Create(ctx, resource)

		assert.Nil(t, got)
		assert.Same(t, infraErr, serr)
	})
}

func TestStackResourceService_Update_Validation(t *testing.T) {
	ctx := context.Background()
	stackID := "stack-123"
	projectID := "project-456"
	resourceName := "web"
	stack := &models.Stack{ID: stackID, ProjectID: projectID, OrganisationID: "org-1"}

	t.Run("siblings passed to the validator exclude the resource being updated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStackStore := mocks.NewMockStackStore(ctrl)
		mockResourceStore := mocks.NewMockStackResourceStore(ctrl)
		mockPermissions := mocks.NewMockPermissionService(ctrl)
		mockValidator := mocks.NewMockValidator(ctrl)
		mockReferenceService := mocks.NewMockReferenceService(ctrl)
		mockDomains := mocks.NewMockStackDomainsService(ctrl)

		svc := &stackResourceService{
			stackStore:         mockStackStore,
			stackResourceStore: mockResourceStore,
			permissions:        mockPermissions,
			resourceValidator:  mockValidator,
			runtimePolicy:      newCloudRuntimePolicyForTest(),
			referenceService:   mockReferenceService,
			domainNameService:  mockDomains,
		}

		existing := &models.StackResource{ID: "resource-1", StackID: stackID, Name: resourceName}
		replicas := int32(7)
		update := &models.StackResource{Replicas: &replicas}
		worker := &models.StackResource{ID: "resource-2", StackID: stackID, Name: "worker"}
		all := []*models.StackResource{existing, worker}

		mockStackStore.EXPECT().GetByID(ctx, stackID).Return(stack, nil)
		mockPermissions.EXPECT().Check(ctx, projectID, auth.ResourceStacks, stackID, auth.ActionWrite).Return(nil)
		mockResourceStore.EXPECT().GetByStackIDAndResourceName(ctx, stackID, resourceName).Return(existing, nil)
		mockResourceStore.EXPECT().GetByStackID(ctx, stackID).Return(all, nil)

		mockValidator.EXPECT().Validate(ctx, stack, update, gomock.Any()).DoAndReturn(
			func(_ context.Context, _ *models.Stack, _ *models.StackResource, siblings []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
				assert.Equal(t, int32(1), *update.Replicas)
				assert.Len(t, siblings, 1)
				assert.Equal(t, "worker", siblings[0].Name)
				for _, s := range siblings {
					assert.NotEqual(t, resourceName, s.Name)
				}
				return nil, nil
			})

		mockStackStore.EXPECT().WithTransaction(ctx, gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})

		updated := &models.StackResource{ID: existing.ID, StackID: stackID, Name: resourceName}
		mockResourceStore.EXPECT().UpdateWithTx(ctx, existing.ID, update, stack).Return(updated, nil)
		mockDomains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx, stack, updated).
			Return(nil)
		mockResourceStore.EXPECT().UpdatePortsWithTx(ctx, updated.ID, updated).Return(nil)
		mockReferenceService.EXPECT().ReprojectSpec(ctx, stackID).Return(nil)
		mockResourceStore.EXPECT().GetByID(ctx, updated.ID).Return(updated, nil)

		got, serr := svc.Update(ctx, stackID, resourceName, update)

		assert.Nil(t, serr)
		if assert.NotNil(t, got) {
			assert.Equal(t, updated.ID, got.ID)
		}
	})
}
