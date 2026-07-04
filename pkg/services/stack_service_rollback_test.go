package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

// realSourceCredentialsForStackTest wires a real sourceCredentialService backed
// by a mock managed-secret service so the stack-service rollback can be observed
// end to end.
func realSourceCredentialsForStackTest(t *testing.T, ctrl *gomock.Controller) (SourceCredentialService, *MockmanagedSecretService) {
	t.Helper()
	secrets := NewMockmanagedSecretService(ctrl)
	svc := NewSourceCredentialService(SourceCredentialServiceSpec{
		SecretService:      secrets,
		CredentialResolver: mocks.NewMockCredentialResolver(ctrl),
		GitClients:         NewMocksourceGitClientProvider(ctrl),
		Logger:             logger.NewLoggerWithPrefix(context.Background(), "test"),
	})
	return svc, secrets
}

func expectPullPushGCStack(secrets *MockmanagedSecretService, ownerID string) {
	for _, slot := range []models.ManagedSecretSlot{models.ManagedSecretSlotPull, models.ManagedSecretSlotPush} {
		secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, slot).
			Return(nil, errors.NotFound("none"))
		secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, slot).
			Return(nil)
	}
}

func TestInternalCreateStackRollsBackManagedSecretsOnValidationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	sourceCreds, secrets := realSourceCredentialsForStackTest(t, ctrl)
	stackStore := mocks.NewMockStackStore(ctrl)
	validator := mocks.NewMockStackValidator(ctrl)

	spec := testStack()
	spec.StackResources = []*models.StackResource{gitInlineResource("api")}
	ownerID := models.StackResourceManagedOwnerID(spec.ID, "api")

	stackStore.EXPECT().GetByName(gomock.Any(), spec.Name, spec.UserID).Return(nil, errors.NotFound("none"))

	// Sources are materialized before validation runs.
	secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGCStack(secrets, ownerID)

	validator.EXPECT().ValidateForCreate(gomock.Any(), spec).Return(errors.Validation("bad stack"))

	// The rejected create must delete the freshly-created managed secret.
	secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil)

	svc := &stackService{
		stackStore:        stackStore,
		stackValidator:    validator,
		sourceCredentials: sourceCreds,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "stack-service-test"),
	}

	_, err := svc.InternalCreateStack(context.Background(), spec)
	if err == nil || err.Code != errors.ErrorValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestInternalUpdateStackRestoresPriorManagedSecretOnValidationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	sourceCreds, secrets := realSourceCredentialsForStackTest(t, ctrl)
	stackStore := mocks.NewMockStackStore(ctrl)
	validator := mocks.NewMockStackValidator(ctrl)

	existing := &models.Stack{
		ID:             "stack-1",
		Name:           "demo",
		Namespace:      "ns-demo",
		ClusterID:      "cluster-1",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "user-1",
	}
	ownerID := models.StackResourceManagedOwnerID(existing.ID, "api")
	prior := &models.Secret{
		ID:            "managed-git-1",
		ManagedByKind: models.ManagedByKindStackResource,
		ManagedByID:   ownerID,
		ManagedSlot:   models.ManagedSecretSlotGit,
		EncryptedData: "prior-encrypted-blob",
	}

	spec := &models.Stack{
		Name:           "demo",
		StackResources: []*models.StackResource{gitInlineResource("api")},
	}

	stackStore.EXPECT().GetByID(gomock.Any(), existing.ID).Return(existing, nil)

	// Prepare overwrites the prior managed git secret in place.
	secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(prior, nil)
	secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGCStack(secrets, ownerID)

	validator.EXPECT().ValidateForUpdate(gomock.Any(), existing, gomock.Any()).Return(errors.Validation("bad update"))

	// The rejected update must restore the prior managed secret data.
	secrets.EXPECT().RestoreManaged(gomock.Any(), prior).Return(nil)

	svc := &stackService{
		stackStore:        stackStore,
		stackValidator:    validator,
		sourceCredentials: sourceCreds,
		logger:            logger.NewLoggerWithPrefix(context.Background(), "stack-service-test"),
	}

	_, err := svc.InternalUpdateStack(context.Background(), existing.ID, spec)
	if err == nil || err.Code != errors.ErrorValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}
