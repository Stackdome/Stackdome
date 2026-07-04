package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestPreviewConfigCreateRollsBackManagedSecretsOnValidationFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := NewMockmanagedSecretService(ctrl)
	sourceCreds := NewSourceCredentialService(SourceCredentialServiceSpec{
		SecretService:      secrets,
		CredentialResolver: mocks.NewMockCredentialResolver(ctrl),
		GitClients:         NewMocksourceGitClientProvider(ctrl),
		Logger:             logger.NewLoggerWithPrefix(context.Background(), "test"),
	})
	store := mocks.NewMockStackPreviewConfigStore(ctrl)
	permissions := mocks.NewMockPermissionService(ctrl)

	config := &models.StackPreviewConfig{
		ID:             "cfg-1",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "user-1",
		Name:           "previews",
		GitRepository: models.PreviewGitRepository{
			// RepoURL intentionally empty so validation fails after the managed
			// secret has already been materialized.
			InlineCredentials: &models.InlineCredentials{Username: "alice", Password: "s3cret"},
		},
	}

	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "", auth.ActionCreate).Return(nil)

	secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindPreviewConfig, "cfg-1", models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-preview-1"}, nil)

	// Validation fails -> the orphaned managed secret must be cleaned up.
	secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindPreviewConfig, "cfg-1", models.ManagedSecretSlotGit).
		Return(nil)

	svc := &stackPreviewConfigService{
		store:             store,
		sourceCredentials: sourceCreds,
		permissions:       permissions,
	}

	if _, err := svc.Create(context.Background(), config); err == nil || err.Code != errors.ErrorValidation {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func newPreviewConfigDeleteService(ctrl *gomock.Controller) (*stackPreviewConfigService, *mocks.MockStackPreviewConfigStore, *mocks.MockPreviewStackStore, *mocks.MockSourceCredentialService, *mocks.MockPermissionService) {
	store := mocks.NewMockStackPreviewConfigStore(ctrl)
	previewStore := mocks.NewMockPreviewStackStore(ctrl)
	sourceCreds := mocks.NewMockSourceCredentialService(ctrl)
	permissions := mocks.NewMockPermissionService(ctrl)
	svc := &stackPreviewConfigService{
		store:             store,
		previewStackStore: previewStore,
		sourceCredentials: sourceCreds,
		permissions:       permissions,
	}
	return svc, store, previewStore, sourceCreds, permissions
}

func TestPreviewConfigDeleteCleansUpBeforeDeletingRow(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, sourceCreds, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(0), nil)

	// Cleanup must run before the row is deleted.
	gomock.InOrder(
		sourceCreds.EXPECT().CleanupPreviewConfig(gomock.Any(), "cfg-1").Return(nil),
		store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(nil),
	)

	if err := svc.Delete(context.Background(), "cfg-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPreviewConfigDeleteCleanupFailureKeepsRowAndRetrySucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, sourceCreds, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil).Times(2)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil).Times(2)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(0), nil).Times(2)

	// First cleanup fails -> row is NOT deleted. Retry's cleanup succeeds and
	// the row delete then proceeds.
	gomock.InOrder(
		sourceCreds.EXPECT().CleanupPreviewConfig(gomock.Any(), "cfg-1").Return(errors.GeneralError("cleanup boom")),
		sourceCreds.EXPECT().CleanupPreviewConfig(gomock.Any(), "cfg-1").Return(nil),
	)
	// Delete is only ever reached on the successful retry.
	store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(nil).Times(1)

	if err := svc.Delete(context.Background(), "cfg-1"); err == nil {
		t.Fatal("expected cleanup failure to be returned")
	}
	if err := svc.Delete(context.Background(), "cfg-1"); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
}

func TestPreviewConfigDeleteRowDeleteFailureRetrySucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, sourceCreds, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil).Times(2)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil).Times(2)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(0), nil).Times(2)

	// Cleanup is idempotent: it runs on both the failed attempt and the retry
	// (a no-op the second time).
	sourceCreds.EXPECT().CleanupPreviewConfig(gomock.Any(), "cfg-1").Return(nil).Times(2)
	gomock.InOrder(
		store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(errors.GeneralError("delete boom")),
		store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(nil),
	)

	if err := svc.Delete(context.Background(), "cfg-1"); err == nil {
		t.Fatal("expected row-delete failure to be returned")
	}
	if err := svc.Delete(context.Background(), "cfg-1"); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
}
