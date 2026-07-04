// Package credentials defines the credential-resolution seam shared by the
// service layer and its consumers (validators, workers). The canonical
// implementation lives in pkg/services; the types are declared here so
// packages that the services layer itself depends on can consume the
// resolver without an import cycle.
package credentials

import (
	"context"
	"strings"
	"time"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

// Source identifies where a resolved credential came from.
type Source string

const (
	// SourceSecretRef is an explicit per-resource secret reference.
	SourceSecretRef Source = "secret_ref"
	// SourceIntegration is an org-level integration matched by host
	// (registry credentials / git integrations; filled in by later phases).
	SourceIntegration Source = "integration"
	// SourceAnonymous means no credentials apply.
	SourceAnonymous Source = "anonymous"
)

// SourceRollback reverts the managed-secret mutations performed by a reversible
// source-credential prepare. It is declared here (rather than in pkg/services)
// so generated mocks of service interfaces that return it do not create an
// import cycle back into pkg/services. Rollback is best-effort: it logs failures
// with enough identifiers to find any orphan and never returns an error, so
// callers always surface the original error. Rolling back twice is a no-op.
type SourceRollback interface {
	Rollback(ctx context.Context)
}

// RegistryPurpose distinguishes pull from push resolution for registries.
type RegistryPurpose string

const (
	RegistryPurposePull RegistryPurpose = "pull"
	RegistryPurposePush RegistryPurpose = "push"
)

type ResolvedGitCredential struct {
	Source      Source
	Credentials gitclient.GitCredentials // zero value == anonymous
	SecretID    string                   // set when Source == SourceSecretRef
	DataHash    string                   // for SecretDataHashAnnotation / release freezing

	// IntegrationID and Host identify the org-level git integration when
	// Source == SourceIntegration.
	IntegrationID string
	Host          string

	// TokenMintedAt/TokenExpiresAt are set when the credentials are a minted
	// GitHub App installation token; they drive hub-side refresh.
	TokenMintedAt  *time.Time
	TokenExpiresAt *time.Time

	// Secret is the backing secret when Source == SourceSecretRef. The
	// cluster secret builders still consume models.Secret; this goes away
	// once they take raw credentials instead.
	Secret *models.Secret
}

type ResolvedRegistryCredential struct {
	Source   Source
	Username string
	Password string
	SecretID string
	DataHash string

	// CredentialID and Host identify the org-level registry credential when
	// Source == SourceIntegration. Host is the normalized registry host.
	CredentialID string
	Host         string

	// Purpose is the resolution purpose (pull vs push) this credential was
	// resolved for. It disambiguates the deterministic cluster-secret name for
	// org-level credentials: a pull credential and a push credential can share
	// the same host but hold different accounts, so they must not collide.
	Purpose RegistryPurpose

	// Secret is the backing secret when Source == SourceSecretRef; see
	// ResolvedGitCredential.Secret.
	Secret *models.Secret
}

const maxClusterSecretNameLen = 63

const (
	gitIntegrationSecretPrefix     = "git-integration-"
	registryCredentialSecretPrefix = "registry-credential-"
)

// ClusterSecretNameForGitHost returns the deterministic name of the cluster
// secret synthesized for an org-level git integration.
func ClusterSecretNameForGitHost(host string) string {
	return clusterSecretName(gitIntegrationSecretPrefix, sanitizeHostForSecretName(host))
}

// ClusterSecretNameForRegistryHost returns the deterministic name of the
// cluster secret synthesized for an org-level registry credential. The purpose
// is part of the name so a pull credential and a push credential sharing a host
// (potentially different accounts) map to distinct cluster secrets.
func ClusterSecretNameForRegistryHost(host string, purpose RegistryPurpose) string {
	prefix := registryCredentialSecretPrefix + sanitizePurposeForSecretName(purpose) + "-"
	return clusterSecretName(prefix, sanitizeHostForSecretName(host))
}

func clusterSecretName(prefix, hostSuffix string) string {
	maxHostLen := maxClusterSecretNameLen - len(prefix)
	if maxHostLen < 1 {
		return trimClusterSecretName(prefix[:maxClusterSecretNameLen])
	}
	if len(hostSuffix) > maxHostLen {
		hostSuffix = hostSuffix[:maxHostLen]
	}
	return trimClusterSecretName(prefix + hostSuffix)
}

func trimClusterSecretName(name string) string {
	if len(name) > maxClusterSecretNameLen {
		name = name[:maxClusterSecretNameLen]
	}
	return strings.TrimRight(name, "-.")
}

func sanitizePurposeForSecretName(purpose RegistryPurpose) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, string(purpose))
}

func sanitizeHostForSecretName(host string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '-'
		}
	}, host)
}

// GitAuthSelector carries the explicit per-resource overrides for git
// credential resolution. The zero value selects no override.
type GitAuthSelector struct {
	SecretRef *models.SecretReference
	// IntegrationID pins an org-level git integration; resolution lands with
	// git integrations.
	IntegrationID string
}

// RegistryAuthSelector carries the explicit per-resource overrides for
// registry credential resolution. The zero value selects no override.
type RegistryAuthSelector struct {
	SecretRef *models.SecretReference
	// RegistryCredentialID pins an org-level registry credential, overriding
	// host auto-attach.
	RegistryCredentialID string
}

// Resolver resolves credentials for git/registry targets with fixed
// precedence: explicit secret ref > explicit integration/credential ID >
// org-level integration matched by host > anonymous.
//
//go:generate mockgen -destination=../mocks/mock_credential_resolver.go -package=mocks -mock_names Resolver=MockCredentialResolver github.com/ashishmax31/stackdome-api-server/pkg/credentials Resolver
type Resolver interface {
	GitCredentials(ctx context.Context, orgID, repoURL string, selector GitAuthSelector) (*ResolvedGitCredential, *errors.ServiceError)
	RegistryCredentials(ctx context.Context, orgID, ref string, purpose RegistryPurpose, selector RegistryAuthSelector) (*ResolvedRegistryCredential, *errors.ServiceError)
}
