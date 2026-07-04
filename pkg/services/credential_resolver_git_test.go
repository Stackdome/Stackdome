package services

import (
	"context"
	"testing"
	"time"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func newResolverWithGitIntegrations(t *testing.T) (CredentialResolver, *MockcredentialSecretFetcher, *MockgitIntegrationResolverSource) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := NewMockcredentialSecretFetcher(ctrl)
	gitIntegrations := NewMockgitIntegrationResolverSource(ctrl)
	resolver := NewCredentialResolver(CredentialResolverSpec{
		SecretService:         secrets,
		GitIntegrationService: gitIntegrations,
	})
	return resolver, secrets, gitIntegrations
}

func TestGitCredentialsAttachesIntegrationByHost(t *testing.T) {
	cases := []struct {
		name    string
		repoURL string
		host    string
	}{
		{name: "https", repoURL: "https://gitlab.example.com/acme/api.git", host: "gitlab.example.com"},
		{name: "https with port", repoURL: "https://gitlab.example.com:8443/acme/api.git", host: "gitlab.example.com"},
		{name: "scp style", repoURL: "git@gitlab.example.com:acme/api.git", host: "gitlab.example.com"},
		{name: "ssh with port", repoURL: "ssh://git@gitlab.example.com:7999/acme/api.git", host: "gitlab.example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

			gitIntegrations.EXPECT().
				InternalGetForHost(gomock.Any(), "org-1", tc.host).
				Return(&models.GitIntegration{
					ID:             "gi-1",
					OrganisationID: "org-1",
					Type:           models.GitIntegrationTypeGitCredentials,
					Host:           tc.host,
					DataHash:       "hash-gi-1",
					Auth:           &models.GitIntegrationAuth{Token: "tok-123"},
				}, nil)

			resolved, serr := resolver.GitCredentials(context.Background(), "org-1", tc.repoURL, credentials.GitAuthSelector{})
			if serr != nil {
				t.Fatalf("unexpected error: %v", serr)
			}
			if resolved.Source != CredentialSourceIntegration {
				t.Fatalf("expected integration source, got %q", resolved.Source)
			}
			if resolved.Credentials.Token != "tok-123" {
				t.Fatalf("expected integration token, got %+v", resolved.Credentials)
			}
			if resolved.IntegrationID != "gi-1" || resolved.Host != tc.host || resolved.DataHash != "hash-gi-1" {
				t.Fatalf("expected integration identity to be carried, got %+v", resolved)
			}
		})
	}
}

func TestGitCredentialsExplicitIntegrationIDWins(t *testing.T) {
	resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

	// Strict mock: no InternalGetForHost expectation — the explicit ID must win.
	gitIntegrations.EXPECT().
		InternalGetByID(gomock.Any(), "gi-2").
		Return(&models.GitIntegration{
			ID:             "gi-2",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitCredentials,
			Host:           "gitea.example.com",
			Auth: &models.GitIntegrationAuth{
				Basic: &models.GitIntegrationBasicAuth{Username: "alice", Password: "s3cret"},
			},
		}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{IntegrationID: "gi-2"})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceIntegration || resolved.IntegrationID != "gi-2" {
		t.Fatalf("expected explicit integration to win, got %+v", resolved)
	}
	if resolved.Credentials.Username != "alice" || resolved.Credentials.Password != "s3cret" {
		t.Fatalf("expected basic credentials, got %+v", resolved.Credentials)
	}
}

func TestGitCredentialsSecretRefWinsOverIntegration(t *testing.T) {
	resolver, secrets, _ := newResolverWithGitIntegrations(t)

	// Strict mock: no integration lookups — the secret ref tier must win.
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-1").Return(&models.Secret{
		ID:   "sec-1",
		Data: map[string]string{models.TokenSecretKey: "sec-token"},
	}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://gitlab.example.com/acme/api", credentials.GitAuthSelector{
		SecretRef: &models.SecretReference{SecretID: "sec-1"},
	})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceSecretRef {
		t.Fatalf("expected secret ref to win, got %q", resolved.Source)
	}
}

func TestGitCredentialsIntegrationOrgMismatchIsNotFound(t *testing.T) {
	resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

	gitIntegrations.EXPECT().
		InternalGetByID(gomock.Any(), "gi-other").
		Return(&models.GitIntegration{
			ID:             "gi-other",
			OrganisationID: "org-2",
			Type:           models.GitIntegrationTypeGitCredentials,
		}, nil)

	_, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{IntegrationID: "gi-other"})
	if serr == nil || !serr.Is404() {
		t.Fatalf("expected 404 for cross-org integration, got %v", serr)
	}
}

func TestGitCredentialsMintsGitHubAppTokenByOwner(t *testing.T) {
	resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

	mintedAt := time.Now().UTC()
	expiresAt := mintedAt.Add(time.Hour)
	gitIntegrations.EXPECT().
		InternalMintForRepo(gomock.Any(), "org-1", "https://github.com/acme/api").
		Return(&models.GitHubAppMintResult{
			IntegrationID: "gi-app",
			Host:          "github.com",
			Token:         "ghs_minted",
			MintedAt:      mintedAt,
			ExpiresAt:     expiresAt,
			DataHash:      "hash-app",
		}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceIntegration {
		t.Fatalf("expected integration source, got %q", resolved.Source)
	}
	if resolved.Credentials.Username != githubapp.CloneUsername || resolved.Credentials.Password != "ghs_minted" {
		t.Fatalf("expected minted token credentials, got %+v", resolved.Credentials)
	}
	if resolved.TokenExpiresAt == nil || !resolved.TokenExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected token expiry to be carried, got %+v", resolved.TokenExpiresAt)
	}
	if resolved.IntegrationID != "gi-app" || resolved.DataHash != "hash-app" {
		t.Fatalf("expected integration identity, got %+v", resolved)
	}
}

func TestGitCredentialsGitHubAppFallsThroughToCredentials(t *testing.T) {
	resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

	// No installation covers the owner; a git_credentials row for github.com
	// still applies.
	gitIntegrations.EXPECT().
		InternalMintForRepo(gomock.Any(), "org-1", "https://github.com/acme/api").
		Return(nil, errors.NotFound("no installation"))
	gitIntegrations.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "github.com").
		Return(&models.GitIntegration{
			ID:             "gi-creds",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitCredentials,
			Host:           "github.com",
			Auth:           &models.GitIntegrationAuth{Token: "tok-fallback"},
		}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceIntegration || resolved.IntegrationID != "gi-creds" {
		t.Fatalf("expected git_credentials fall-through, got %+v", resolved)
	}
}

func TestGitCredentialsFallsBackToAnonymousWhenNoIntegration(t *testing.T) {
	resolver, _, gitIntegrations := newResolverWithGitIntegrations(t)

	gitIntegrations.EXPECT().
		InternalMintForRepo(gomock.Any(), "org-1", "https://github.com/acme/api").
		Return(nil, errors.NotFound("none"))
	gitIntegrations.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "github.com").
		Return(nil, errors.NotFound("none"))

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceAnonymous || resolved.Credentials != (gitclient.GitCredentials{}) {
		t.Fatalf("expected anonymous fallback, got %+v", resolved)
	}
}
