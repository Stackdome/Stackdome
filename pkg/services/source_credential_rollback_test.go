package services

import (
	"context"
	"fmt"
	"testing"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/credentials"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
)

// gitInlineResource returns a git-build resource carrying inline credentials so
// prepare materializes a managed git secret and GCs the (absent) pull/push
// slots.
func gitInlineResource(name string) *models.StackResource {
	r := gitBuildResource(name)
	r.BuildConfig.SourceContext.Git.InlineCredentials = &models.InlineCredentials{
		Username: "alice",
		Password: "s3cret",
	}
	return r
}

// expectPullPushGC sets up the drop-slot expectations for pull and push through
// the recorder (a GetManagedFor snapshot returning 404, then the delete).
func expectPullPushGC(deps *sourceCredentialMocks, ownerID string) {
	for _, slot := range []models.ManagedSecretSlot{models.ManagedSecretSlotPull, models.ManagedSecretSlotPush} {
		deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, slot).
			Return(nil, errors.NotFound("none"))
		deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, slot).
			Return(nil)
	}
}

func TestPrepareStackSourcesRollbackDeletesCreatedManagedSecret(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	stack.StackResources = []*models.StackResource{gitInlineResource("api")}
	ownerID := models.StackResourceManagedOwnerID(stack.ID, "api")

	// No prior secret at the git slot -> newly created.
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGC(deps, ownerID)

	// Rollback of a newly-created secret deletes it.
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil)

	rollback, err := svc.PrepareStackSources(context.Background(), stack)
	if err != nil {
		t.Fatalf("unexpected prepare error: %v", err)
	}
	rollback.Rollback(context.Background())
}

func TestPrepareStackSourcesRollbackRestoresOverwrittenManagedSecret(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	stack.StackResources = []*models.StackResource{gitInlineResource("api")}
	ownerID := models.StackResourceManagedOwnerID(stack.ID, "api")

	prior := &models.Secret{
		ID:            "managed-git-1",
		ManagedByKind: models.ManagedByKindStackResource,
		ManagedByID:   ownerID,
		ManagedSlot:   models.ManagedSecretSlotGit,
		EncryptedData: "prior-encrypted-blob",
	}

	// A prior secret exists -> the upsert overwrites it, so rollback must
	// restore the captured snapshot.
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(prior, nil)
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGC(deps, ownerID)

	deps.secrets.EXPECT().RestoreManaged(gomock.Any(), prior).Return(nil)

	rollback, err := svc.PrepareStackSources(context.Background(), stack)
	if err != nil {
		t.Fatalf("unexpected prepare error: %v", err)
	}
	rollback.Rollback(context.Background())
}

func TestPrepareStackSourcesAutoRollsBackOnPrepareFailure(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	res := gitInlineResource("api")
	// Force default-branch resolution to run (and fail) after the git secret
	// has already been materialized.
	res.BuildConfig.SourceRevision.Git = &models.GitRevision{}
	stack.StackResources = []*models.StackResource{res}
	ownerID := models.StackResourceManagedOwnerID(stack.ID, "api")

	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGC(deps, ownerID)

	deps.resolver.EXPECT().
		GitCredentials(gomock.Any(), stack.OrganisationID, "https://github.com/acme/api", credentials.GitAuthSelector{
			SecretRef: &models.SecretReference{SecretID: "managed-git-1"},
		}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	deps.gitClients.EXPECT().ClientFor(gomock.Any(), gomock.Any()).Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().GetDefaultBranch(gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("auth: %w", gitclient.ErrAuthFailed))

	// The partial prepare must be undone before returning the error.
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil)

	rollback, err := svc.PrepareStackSources(context.Background(), stack)
	if err == nil {
		t.Fatal("expected prepare to fail on default-branch resolution")
	}
	if rollback != nil {
		t.Fatalf("expected nil rollback handle on prepare failure, got %v", rollback)
	}
}

func TestPrepareStackSourcesRollbackIsIdempotent(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	stack.StackResources = []*models.StackResource{gitInlineResource("api")}
	ownerID := models.StackResourceManagedOwnerID(stack.ID, "api")

	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-git-1"}, nil)
	expectPullPushGC(deps, ownerID)

	// The delete must happen exactly once even though Rollback is called twice.
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil).
		Times(1)

	rollback, err := svc.PrepareStackSources(context.Background(), stack)
	if err != nil {
		t.Fatalf("unexpected prepare error: %v", err)
	}
	rollback.Rollback(context.Background())
	rollback.Rollback(context.Background())
}

func TestPreparePreviewConfigRollbackDeletesCreatedManagedSecret(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	config := &models.StackPreviewConfig{
		ID:             "cfg-1",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "user-1",
		Name:           "previews",
		GitRepository: models.PreviewGitRepository{
			RepoURL:           "https://github.com/acme/api",
			BaseBranch:        "main",
			InlineCredentials: &models.InlineCredentials{Username: "alice", Password: "s3cret"},
		},
	}

	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindPreviewConfig, "cfg-1", models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-preview-1"}, nil)

	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindPreviewConfig, "cfg-1", models.ManagedSecretSlotGit).
		Return(nil)

	rollback, err := svc.PreparePreviewConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected prepare error: %v", err)
	}
	rollback.Rollback(context.Background())
}
