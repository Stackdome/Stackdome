package services

import (
	"context"
	"testing"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/credentials"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
)

func newResolverWithMockedSecrets(t *testing.T) (CredentialResolver, *MockcredentialSecretFetcher) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := NewMockcredentialSecretFetcher(ctrl)
	return NewCredentialResolver(CredentialResolverSpec{SecretService: secrets}), secrets
}

func TestGitCredentialsAnonymousWithoutOverride(t *testing.T) {
	resolver, _ := newResolverWithMockedSecrets(t)

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
	if resolved.SecretID != "" || resolved.DataHash != "" || resolved.Secret != nil {
		t.Fatalf("expected no secret material on anonymous resolution, got %+v", resolved)
	}
}

func TestGitCredentialsOverrideWithTokenSecret(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-1").Return(&models.Secret{
		ID:       "sec-1",
		DataHash: "hash-1",
		Data: map[string]string{
			models.TokenSecretKey: "tok-123",
		},
	}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{SecretRef: &models.SecretReference{SecretID: "sec-1"}})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceSecretRef {
		t.Fatalf("expected source %q, got %q", CredentialSourceSecretRef, resolved.Source)
	}
	if resolved.Credentials.Token != "tok-123" || resolved.Credentials.Username != "" || resolved.Credentials.Password != "" {
		t.Fatalf("expected token credentials, got %+v", resolved.Credentials)
	}
	if resolved.SecretID != "sec-1" || resolved.DataHash != "hash-1" {
		t.Fatalf("expected secret id/hash to be carried, got %+v", resolved)
	}
	if resolved.Secret == nil || resolved.Secret.ID != "sec-1" {
		t.Fatalf("expected backing secret to be carried, got %+v", resolved.Secret)
	}
}

func TestGitCredentialsOverrideWithBasicAuthSecret(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-2").Return(&models.Secret{
		ID:       "sec-2",
		DataHash: "hash-2",
		Data: map[string]string{
			models.UsernameSecretKey: "alice",
			models.PasswordSecretKey: "s3cret",
		},
	}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://git.example.com/acme/api", credentials.GitAuthSelector{SecretRef: &models.SecretReference{SecretID: "sec-2"}})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Credentials.Username != "alice" || resolved.Credentials.Password != "s3cret" || resolved.Credentials.Token != "" {
		t.Fatalf("expected basic-auth credentials, got %+v", resolved.Credentials)
	}
}

func TestGitCredentialsEmptyTokenFallsBackToBasicAuth(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-3").Return(&models.Secret{
		ID: "sec-3",
		Data: map[string]string{
			models.TokenSecretKey:    "",
			models.UsernameSecretKey: "alice",
			models.PasswordSecretKey: "s3cret",
		},
	}, nil)

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://git.example.com/acme/api", credentials.GitAuthSelector{SecretRef: &models.SecretReference{SecretID: "sec-3"}})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Credentials.Username != "alice" || resolved.Credentials.Password != "s3cret" {
		t.Fatalf("expected basic-auth fallback for empty token, got %+v", resolved.Credentials)
	}
}

func TestGitCredentialsMissingSecret(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "missing").Return(nil, errors.NotFound("secret not found"))

	resolved, serr := resolver.GitCredentials(context.Background(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{SecretRef: &models.SecretReference{SecretID: "missing"}})
	if serr == nil {
		t.Fatal("expected missing secret to error")
	}
	if !serr.Is404() {
		t.Fatalf("expected not-found error, got %v", serr)
	}
	if resolved != nil {
		t.Fatalf("expected nil resolution on error, got %+v", resolved)
	}
}

func TestRegistryCredentialsAnonymousWithoutOverride(t *testing.T) {
	resolver, _ := newResolverWithMockedSecrets(t)

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "nginx:latest", RegistryPurposePull, credentials.RegistryAuthSelector{})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceAnonymous {
		t.Fatalf("expected source %q, got %q", CredentialSourceAnonymous, resolved.Source)
	}
	if resolved.Username != "" || resolved.Password != "" || resolved.Secret != nil {
		t.Fatalf("expected no credential material on anonymous resolution, got %+v", resolved)
	}
}

func TestRegistryCredentialsOverride(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-4").Return(&models.Secret{
		ID:       "sec-4",
		DataHash: "hash-4",
		Data: map[string]string{
			models.UsernameSecretKey: "bob",
			models.PasswordSecretKey: "hunter2",
		},
	}, nil)

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/api:latest", RegistryPurposePush, credentials.RegistryAuthSelector{SecretRef: &models.SecretReference{SecretID: "sec-4"}})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceSecretRef {
		t.Fatalf("expected source %q, got %q", CredentialSourceSecretRef, resolved.Source)
	}
	if resolved.Username != "bob" || resolved.Password != "hunter2" {
		t.Fatalf("expected secret credentials, got %+v", resolved)
	}
	if resolved.SecretID != "sec-4" || resolved.DataHash != "hash-4" || resolved.Secret == nil {
		t.Fatalf("expected secret id/hash/secret to be carried, got %+v", resolved)
	}
}

func TestRegistryCredentialsMissingSecret(t *testing.T) {
	resolver, secrets := newResolverWithMockedSecrets(t)
	secrets.EXPECT().InternalGetByID(gomock.Any(), "missing").Return(nil, errors.NotFound("secret not found"))

	_, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/api:latest", RegistryPurposePull, credentials.RegistryAuthSelector{SecretRef: &models.SecretReference{SecretID: "missing"}})
	if serr == nil {
		t.Fatal("expected missing secret to error")
	}
	if !serr.Is404() {
		t.Fatalf("expected not-found error, got %v", serr)
	}
}

func newResolverWithRegistryCredentials(t *testing.T) (CredentialResolver, *MockcredentialSecretFetcher, *MockregistryCredentialResolverSource) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	secrets := NewMockcredentialSecretFetcher(ctrl)
	registryCredentials := NewMockregistryCredentialResolverSource(ctrl)
	resolver := NewCredentialResolver(CredentialResolverSpec{
		SecretService:             secrets,
		RegistryCredentialService: registryCredentials,
	})
	return resolver, secrets, registryCredentials
}

func TestRegistryCredentialsAttachesOrgCredentialByHost(t *testing.T) {
	resolver, _, registryCredentials := newResolverWithRegistryCredentials(t)

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
	if resolved.Secret != nil || resolved.SecretID != "" {
		t.Fatalf("integration tier must not carry a backing secret, got %+v", resolved)
	}
}

func TestRegistryCredentialsOverrideWinsOverOrgCredential(t *testing.T) {
	resolver, secrets, _ := newResolverWithRegistryCredentials(t)

	// Strict mock: no InternalGetForHost expectation — the explicit ref must win.
	secrets.EXPECT().InternalGetByID(gomock.Any(), "sec-1").Return(&models.Secret{
		ID:       "sec-1",
		DataHash: "hash-1",
		Data: map[string]string{
			models.UsernameSecretKey: "alice",
			models.PasswordSecretKey: "s3cret",
		},
	}, nil)

	resolved, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/app:1", RegistryPurposePull, credentials.RegistryAuthSelector{SecretRef: &models.SecretReference{SecretID: "sec-1"}})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if resolved.Source != CredentialSourceSecretRef {
		t.Fatalf("expected explicit ref to win, got source %q", resolved.Source)
	}
}

func TestRegistryCredentialsFallsBackToAnonymousWhenNoOrgCredential(t *testing.T) {
	resolver, _, registryCredentials := newResolverWithRegistryCredentials(t)

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
	resolver, _, registryCredentials := newResolverWithRegistryCredentials(t)

	registryCredentials.EXPECT().
		InternalGetForHost(gomock.Any(), "org-1", "ghcr.io", models.RegistryCredentialPurposePull).
		Return(nil, errors.GeneralError("db down"))

	_, serr := resolver.RegistryCredentials(context.Background(), "org-1", "ghcr.io/acme/app", RegistryPurposePull, credentials.RegistryAuthSelector{})
	if serr == nil {
		t.Fatal("expected lookup error to propagate")
	}
}
