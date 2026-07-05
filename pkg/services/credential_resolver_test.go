package services

import (
	"context"
	"testing"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestGitCredentialsAnonymousWithoutOverride(t *testing.T) {
	resolver := NewCredentialResolver(CredentialResolverSpec{})

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceAnonymous {
		t.Fatalf("expected source %q, got %q", CredentialSourceAnonymous, resolved.Source)
	}
	if resolved.Credentials != (gitclient.GitCredentials{}) {
		t.Fatalf("expected zero credentials, got %+v", resolved.Credentials)
	}
	if resolved.IntegrationID != "" || resolved.DataHash != "" {
		t.Fatalf("expected no integration material on anonymous resolution, got %+v", resolved)
	}
}

func TestRegistryCredentialsAnonymousWithoutOverride(t *testing.T) {
	resolver := NewCredentialResolver(CredentialResolverSpec{})

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "nginx:latest", RegistryPurposePull, credentials.RegistryAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceAnonymous {
		t.Fatalf("expected source %q, got %q", CredentialSourceAnonymous, resolved.Source)
	}
	if resolved.Username != "" || resolved.Password != "" || resolved.CredentialID != "" {
		t.Fatalf("expected no credential material on anonymous resolution, got %+v", resolved)
	}
}

func newResolverWithRegistryCredentials(t *testing.T) (CredentialResolver, *MockregistryCredentialResolverSource) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	registryCredentials := NewMockregistryCredentialResolverSource(ctrl)
	resolver := NewCredentialResolver(CredentialResolverSpec{
		RegistryCredentialService: registryCredentials,
	})
	return resolver, registryCredentials
}

func TestRegistryCredentialsAttachesOrgCredentialByHost(t *testing.T) {
	resolver, registryCredentials := newResolverWithRegistryCredentials(t)

	registryCredentials.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "ghcr.io", models.RegistryCredentialPurposePull).
		Return(&models.RegistryCredential{
			ID:       "rc-1",
			Host:     "ghcr.io",
			Purpose:  models.RegistryCredentialPurposeBoth,
			Username: "acme-bot",
			Password: "hunter2",
			DataHash: "hash-rc-1",
		}, nil)

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/app:1", RegistryPurposePull, credentials.RegistryAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceIntegration {
		t.Fatalf("expected source %q, got %q", CredentialSourceIntegration, resolved.Source)
	}
	if resolved.Username != "acme-bot" || resolved.Password != "hunter2" {
		t.Fatalf("expected org credential material, got %+v", resolved)
	}
	if resolved.CredentialID != "rc-1" || resolved.Host != "ghcr.io" || resolved.DataHash != "hash-rc-1" {
		t.Fatalf("expected credential identity to be carried, got %+v", resolved)
	}
}

func TestRegistryCredentialsFallsBackToAnonymousWhenNoOrgCredential(t *testing.T) {
	resolver, registryCredentials := newResolverWithRegistryCredentials(t)

	registryCredentials.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "ghcr.io", models.RegistryCredentialPurposePush).
		Return(nil, errors.NotFound("no credential"))

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/app", RegistryPurposePush, credentials.RegistryAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceAnonymous {
		t.Fatalf("expected anonymous fallback, got %q", resolved.Source)
	}
}

func TestRegistryCredentialsPropagatesLookupErrors(t *testing.T) {
	resolver, registryCredentials := newResolverWithRegistryCredentials(t)

	registryCredentials.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "ghcr.io", models.RegistryCredentialPurposePull).
		Return(nil, errors.GeneralError("db down"))

	_, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/app", RegistryPurposePull, credentials.RegistryAuthSelector{})
	if serr == nil {
		t.Fatal("expected lookup error to propagate")
	}
}
