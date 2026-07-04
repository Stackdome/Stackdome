package services

import (
	"context"
	"fmt"
	"testing"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/credentials"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/mocks"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
)

type sourceCredentialMocks struct {
	secrets             *MockmanagedSecretService
	registryCredentials *MockorgRegistryCredentialCreator
	resolver            *mocks.MockCredentialResolver
	gitClients          *MocksourceGitClientProvider
	gitClient           *mocks.MockGitClient
}

func newSourceCredentialServiceForTest(t *testing.T) (SourceCredentialService, *sourceCredentialMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deps := &sourceCredentialMocks{
		secrets:             NewMockmanagedSecretService(ctrl),
		registryCredentials: NewMockorgRegistryCredentialCreator(ctrl),
		resolver:            mocks.NewMockCredentialResolver(ctrl),
		gitClients:          NewMocksourceGitClientProvider(ctrl),
		gitClient:           mocks.NewMockGitClient(ctrl),
	}
	svc := NewSourceCredentialService(SourceCredentialServiceSpec{
		SecretService:             deps.secrets,
		RegistryCredentialService: deps.registryCredentials,
		CredentialResolver:        deps.resolver,
		GitClients:                deps.gitClients,
	})
	return svc, deps
}

func testStack() *models.Stack {
	return &models.Stack{
		ID:             "stack-1",
		Name:           "demo",
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "user-1",
	}
}

func gitBuildResource(name string) *models.StackResource {
	return &models.StackResource{
		Name: name,
		BuildConfig: &models.BuildConfigSpec{
			SourceContext: models.BuildContextSource{
				Git: &models.GitBuildSource{RepoURL: "https://github.com/acme/api"},
			},
			SourceRevision: models.BuildSourceRevision{
				Git: &models.GitRevision{Branch: "main"},
			},
			BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true, InsecureRegistry: true},
		},
	}
}

func TestPrepareResourceSourceMaterializesInlineGitCredentials(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")
	resource.BuildConfig.SourceContext.Git.InlineCredentials = &models.InlineCredentials{
		Username: "alice",
		Password: "s3cret",
	}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), ManagedSecretSpec{
		OrganisationID: stack.OrganisationID,
		TeamID:         stack.TeamID,
		UserID:         stack.UserID,
		OwnerKind:      models.ManagedByKindStackResource,
		OwnerID:        ownerID,
		Slot:           models.ManagedSecretSlotGit,
		Name:           models.ManagedSecretName(stack.Name, resource.Name, models.ManagedSecretSlotGit),
		Type:           models.SecretTypeGitCredentials,
		Data: map[string]string{
			models.UsernameSecretKey: "alice",
			models.PasswordSecretKey: "s3cret",
		},
	}).Return(&models.Secret{ID: "managed-1"}, nil)
	// pull and push slots are absent -> GC'd
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	git := resource.BuildConfig.SourceContext.Git
	if git.GitSecretRef == nil || git.GitSecretRef.SecretID != "managed-1" {
		t.Fatalf("expected git secret ref to be wired, got %+v", git.GitSecretRef)
	}
	if git.InlineCredentials != nil {
		t.Fatal("expected inline credentials to be cleared after materialization")
	}
}

func TestPrepareResourceSourceRewiresExistingManagedSecret(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(&models.Secret{ID: "managed-1"}, nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref := resource.BuildConfig.SourceContext.Git.GitSecretRef; ref == nil || ref.SecretID != "managed-1" {
		t.Fatalf("expected existing managed secret to be rewired, got %+v", ref)
	}
}

func TestPrepareResourceSourceGCsManagedSecretsWhenSlotDisappears(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := &models.StackResource{
		Name:        "api",
		ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
	}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	// git and push slots gone
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)
	// pull slot with no creds and no existing managed secret
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull).
		Return(nil, errors.NotFound("none"))

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resource.ImageConfig.PullSecretRef != nil {
		t.Fatal("expected no pull secret ref")
	}
}

func TestPrepareResourceSourceSaveToOrgPromotesRegistryCredential(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := &models.StackResource{
		Name: "api",
		ImageConfig: &models.ImageConfigSpec{
			Image: "ghcr.io/acme/app:1",
			InlineCredentials: &models.InlineCredentials{
				Username:  "alice",
				Password:  "s3cret",
				SaveToOrg: true,
			},
		},
	}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.registryCredentials.EXPECT().Create(gomock.Any(), &models.RegistryCredential{
		OrganisationID: stack.OrganisationID,
		Host:           "ghcr.io/acme/app:1",
		Purpose:        models.RegistryCredentialPurposePull,
		Username:       "alice",
		Password:       "s3cret",
	}).Return(&models.RegistryCredential{ID: "rc-1"}, nil)
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).Return(&models.Secret{ID: "managed-1"}, nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPrepareResourceSourceSaveToOrgConflictPropagates(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := &models.StackResource{
		Name: "api",
		ImageConfig: &models.ImageConfigSpec{
			Image: "ghcr.io/acme/app:1",
			InlineCredentials: &models.InlineCredentials{
				Username:  "alice",
				Password:  "s3cret",
				SaveToOrg: true,
			},
		},
	}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.registryCredentials.EXPECT().Create(gomock.Any(), gomock.Any()).
		Return(nil, errors.Conflict("already exists"))
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).Return(nil)

	err := svc.PrepareResourceSource(context.Background(), stack, resource)
	if err == nil || !err.IsConflict() {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
}

func TestPrepareResourceSourceSaveToOrgForGitIsNotSupported(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")
	resource.BuildConfig.SourceContext.Git.InlineCredentials = &models.InlineCredentials{
		Username:  "alice",
		Password:  "s3cret",
		SaveToOrg: true,
	}
	_ = deps

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err == nil {
		t.Fatal("expected save_to_org for git credentials to be rejected")
	}
}

func TestPrepareResourceSourceResolvesDefaultBranch(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")
	resource.BuildConfig.SourceRevision.Git = &models.GitRevision{}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)

	deps.resolver.EXPECT().
		GitCredentials(gomock.Any(), stack.OrganisationID, "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	deps.gitClients.EXPECT().ClientFor("https://github.com/acme/api", gitclient.GitCredentials{}).
		Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().GetDefaultBranch(gomock.Any(), "https://github.com/acme/api").Return("trunk", nil)

	if err := svc.PrepareResourceSource(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resource.BuildConfig.SourceRevision.Git.Branch; got != "trunk" {
		t.Fatalf("expected resolved default branch to be stored, got %q", got)
	}
}

func TestPrepareResourceSourceDefaultBranchAuthFailureIsStructured(t *testing.T) {
	svc, deps := newSourceCredentialServiceForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")
	resource.BuildConfig.SourceRevision.Git = &models.GitRevision{}

	ownerID := models.StackResourceManagedOwnerID(stack.ID, resource.Name)
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPull).Return(nil)
	deps.secrets.EXPECT().DeleteManagedFor(gomock.Any(), models.ManagedByKindStackResource, ownerID, models.ManagedSecretSlotPush).Return(nil)

	deps.resolver.EXPECT().
		GitCredentials(gomock.Any(), stack.OrganisationID, "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	deps.gitClients.EXPECT().ClientFor(gomock.Any(), gomock.Any()).Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().GetDefaultBranch(gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("authentication failed: %w", gitclient.ErrAuthFailed))

	err := svc.PrepareResourceSource(context.Background(), stack, resource)
	if err == nil {
		t.Fatal("expected auth failure to error")
	}
	details, ok := err.Details.(errors.CredentialErrorDetails)
	if !ok {
		t.Fatalf("expected structured credential details, got %#v", err.Details)
	}
	if details.Code != errors.ErrorCodeCredentialsRequired {
		t.Fatalf("expected %s, got %s", errors.ErrorCodeCredentialsRequired, details.Code)
	}
	if details.Target.Kind != errors.CredentialTargetKindGitClone {
		t.Fatalf("expected git_clone target, got %s", details.Target.Kind)
	}
}

func TestPreparePreviewConfigMaterializesInlineCredentials(t *testing.T) {
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

	// The reversible prepare snapshots the prior secret (none) before the upsert.
	deps.secrets.EXPECT().GetManagedFor(gomock.Any(), models.ManagedByKindPreviewConfig, "cfg-1", models.ManagedSecretSlotGit).
		Return(nil, errors.NotFound("none"))
	deps.secrets.EXPECT().CreateManaged(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, spec ManagedSecretSpec) (*models.Secret, *errors.ServiceError) {
			if spec.OwnerKind != models.ManagedByKindPreviewConfig || spec.OwnerID != "cfg-1" {
				t.Fatalf("unexpected owner: %s/%s", spec.OwnerKind, spec.OwnerID)
			}
			return &models.Secret{ID: "managed-preview-1"}, nil
		})

	if _, err := svc.PreparePreviewConfig(context.Background(), config); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.GitRepository.GitSecretID == nil || *config.GitRepository.GitSecretID != "managed-preview-1" {
		t.Fatalf("expected managed secret to be wired, got %v", config.GitRepository.GitSecretID)
	}
	if config.GitRepository.InlineCredentials != nil {
		t.Fatal("expected inline credentials to be cleared")
	}
}
