package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestInternalDeleteFromDBCleansUpManagedSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	stack := &models.Stack{
		ID:   "stack-1",
		Name: "demo",
		StackResources: []*models.StackResource{
			{Name: "api"},
			{Name: "web"},
		},
	}

	stackStore := mocks.NewMockStackStore(ctrl)
	stackStore.EXPECT().GetByID(gomock.Any(), "stack-1").Return(stack, nil)
	stackStore.EXPECT().Delete(gomock.Any(), "stack-1").Return(nil)

	sourceCredentials := mocks.NewMockSourceCredentialService(ctrl)
	sourceCredentials.EXPECT().CleanupStackSources(gomock.Any(), stack).Return(nil)

	svc := &stackService{
		stackStore:        stackStore,
		sourceCredentials: sourceCredentials,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "stack-service-test"),
	}

	if err := svc.InternalDeleteFromDB(context.Background(), "stack-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCleanupStackSourcesDeletesEveryResourceOwner(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := &models.Stack{
		ID:   "stack-1",
		Name: "demo",
		StackResources: []*models.StackResource{
			{Name: "api"},
			{Name: "web"},
		},
	}

	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource,
		models.StackResourceManagedOwnerID("stack-1", "api")).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource,
		models.StackResourceManagedOwnerID("stack-1", "web")).Return(nil)

	if err := svc.CleanupStackSources(context.Background(), stack); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
