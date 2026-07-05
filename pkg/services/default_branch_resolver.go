package services

import (
	"context"
	stderrors "errors"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -destination=../mocks/mock_default_branch_resolver.go -package=mocks github.com/Stackdome/stackdome/pkg/services DefaultBranchResolver
//go:generate mockgen -source=default_branch_resolver.go -destination=default_branch_resolver_mock_test.go -package=services -exclude_interfaces DefaultBranchResolver

// DefaultBranchResolver fills in a git source's default branch when the source
// specifies neither a branch nor a tag, pinning the result so releases and pins
// stay reproducible.
type DefaultBranchResolver interface {
	ResolveDefaultBranch(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError
}

// sourceGitClientProvider builds git clients so default-branch resolution can
// be faked in tests.
type sourceGitClientProvider interface {
	ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error)
}

type defaultSourceGitClientProvider struct{}

func (defaultSourceGitClientProvider) ClientFor(repoURL string, creds gitclient.GitCredentials) (gitclient.GitClient, error) {
	return gitclient.NewGitClientForRepo(repoURL, creds)
}

type DefaultBranchResolverSpec struct {
	CredentialResolver CredentialResolver
	// GitClients is optional; it defaults to real git clients.
	GitClients sourceGitClientProvider
}

type defaultBranchResolver struct {
	resolver   CredentialResolver
	gitClients sourceGitClientProvider
}

func NewDefaultBranchResolver(spec DefaultBranchResolverSpec) DefaultBranchResolver {
	gitClients := spec.GitClients
	if gitClients == nil {
		gitClients = defaultSourceGitClientProvider{}
	}
	return &defaultBranchResolver{
		resolver:   spec.CredentialResolver,
		gitClients: gitClients,
	}
}

// ResolveDefaultBranch fills in the repository's default branch when a git
// source specifies neither branch nor tag. The result is stored on the resource
// so releases and pins stay reproducible.
func (s *defaultBranchResolver) ResolveDefaultBranch(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	if resource.BuildConfig == nil || resource.BuildConfig.SourceContext.Git == nil {
		return nil
	}
	rev := resource.BuildConfig.SourceRevision.Git
	if rev == nil || rev.Branch != "" || rev.Tag != "" {
		return nil
	}

	git := resource.BuildConfig.SourceContext.Git
	resolved, serr := s.resolver.GitCredentials(ctx, stack.OrganisationID, git.RepoURL, credentials.GitAuthSelector{
		SecretRef:     git.GitSecretRef,
		IntegrationID: git.IntegrationID,
	})
	if serr != nil {
		return serr.WithPrefix("resource '%s': failed to resolve git credentials", resource.Name)
	}

	client, err := s.gitClients.ClientFor(git.RepoURL, resolved.Credentials)
	if err != nil {
		return errors.GeneralError("resource '%s': failed to create git client: %s", resource.Name, err.Error())
	}

	branch, err := client.GetDefaultBranch(ctx, git.RepoURL)
	if err != nil {
		target := errors.CredentialErrorTarget{
			Kind: errors.CredentialTargetKindGitClone,
			Host: gitRepoHost(git.RepoURL),
			Ref:  git.RepoURL,
		}
		if stderrors.Is(err, gitclient.ErrAuthFailed) {
			if resolved.Source == credentials.SourceAnonymous {
				return errors.CredentialsRequired(target, "resource '%s': repository '%s' requires credentials to resolve its default branch", resource.Name, git.RepoURL)
			}
			return errors.CredentialsInvalid(target, "resource '%s': configured git credentials were rejected by '%s'", resource.Name, git.RepoURL)
		}
		return errors.BadRequest("resource '%s': failed to resolve default branch for '%s': %s", resource.Name, git.RepoURL, err.Error())
	}

	rev.Branch = branch
	return nil
}

func gitRepoHost(repoURL string) string {
	return gitclient.RepoHost(repoURL)
}
