package services

import (
	"context"
	"fmt"
	"testing"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

type defaultBranchResolverMocks struct {
	resolver   *mocks.MockCredentialResolver
	gitClients *MocksourceGitClientProvider
	gitClient  *mocks.MockGitClient
}

func newDefaultBranchResolverForTest(t *testing.T) (DefaultBranchResolver, *defaultBranchResolverMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deps := &defaultBranchResolverMocks{
		resolver:   mocks.NewMockCredentialResolver(ctrl),
		gitClients: NewMocksourceGitClientProvider(ctrl),
		gitClient:  mocks.NewMockGitClient(ctrl),
	}
	svc := NewDefaultBranchResolver(DefaultBranchResolverSpec{
		CredentialResolver: deps.resolver,
		GitClients:         deps.gitClients,
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
				Git: &models.GitRevision{},
			},
			BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true, InsecureRegistry: true},
		},
	}
}

func TestResolveDefaultBranchStoresResolvedBranch(t *testing.T) {
	svc, deps := newDefaultBranchResolverForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")

	deps.resolver.EXPECT().
		GitCredentials(gomock.Any(), stack.OrganisationID, "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	deps.gitClients.EXPECT().ClientFor("https://github.com/acme/api", gitclient.GitCredentials{}).
		Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().GetDefaultBranch(gomock.Any(), "https://github.com/acme/api").Return("trunk", nil)

	if err := svc.ResolveDefaultBranch(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resource.BuildConfig.SourceRevision.Git.Branch; got != "trunk" {
		t.Fatalf("expected resolved default branch to be stored, got %q", got)
	}
}

func TestResolveDefaultBranchSkipsWhenBranchSet(t *testing.T) {
	svc, _ := newDefaultBranchResolverForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")
	resource.BuildConfig.SourceRevision.Git.Branch = "main"

	// No resolver/git-client expectations: a branch is already pinned.
	if err := svc.ResolveDefaultBranch(context.Background(), stack, resource); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resource.BuildConfig.SourceRevision.Git.Branch; got != "main" {
		t.Fatalf("expected branch to be left untouched, got %q", got)
	}
}

func TestResolveDefaultBranchAuthFailureIsStructured(t *testing.T) {
	svc, deps := newDefaultBranchResolverForTest(t)
	stack := testStack()
	resource := gitBuildResource("api")

	deps.resolver.EXPECT().
		GitCredentials(gomock.Any(), stack.OrganisationID, "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)
	deps.gitClients.EXPECT().ClientFor(gomock.Any(), gomock.Any()).Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().GetDefaultBranch(gomock.Any(), gomock.Any()).
		Return("", fmt.Errorf("authentication failed: %w", gitclient.ErrAuthFailed))

	err := svc.ResolveDefaultBranch(context.Background(), stack, resource)
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
