package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

// ErrRateLimited indicates the registry responded with a rate-limit signal
// (HTTP 429 / toomanyrequests) rather than a definitive answer. Callers can
// use errors.Is(err, ErrRateLimited) to soft-pass instead of treating it as a
// hard failure.
var ErrRateLimited = errors.New("registry rate limited")

// ErrAuthFailed indicates the registry rejected the request for auth reasons
// (401/403). Callers can use errors.Is to distinguish missing/invalid
// credentials from other failures.
var ErrAuthFailed = errors.New("registry authentication failed")

// ClusterLocalRegistrySuffix marks in-cluster registries; those are not
// reachable from the hub, so existence/push probes are skipped for them.
const ClusterLocalRegistrySuffix = ".svc.cluster.local"

// Docker Hub aliases that all normalize to name.DefaultRegistry.
const (
	dockerHubAliasHost         = "docker.io"
	dockerHubRegistryAliasHost = "registry-1.docker.io"
)

// This is used by a user input as an image reference.
// NormalizeRegistryHost parses an image ref, repository ref, or bare registry
// host and returns the canonical registry host. Docker Hub aliases (docker.io,
// registry-1.docker.io) all normalize to name.DefaultRegistry.
func NormalizeRegistryHost(ref string) (string, error) {
	// A bare registry host has no path component and a dotted hostname
	// (ghcr.io, registry.example.com:5000). Anything else with no path —
	// redis, redis:7, nginx:1.25 — is a Docker Hub image name.
	if !strings.Contains(ref, "/") && looksLikeRegistryHost(ref) {
		reg, err := name.NewRegistry(ref)
		if err != nil {
			return "", fmt.Errorf("invalid registry host '%s': %w", ref, err)
		}
		return normalizeDockerHubHost(reg.RegistryStr()), nil
	}

	parsed, err := name.ParseReference(ref)
	if err != nil {
		return "", fmt.Errorf("invalid image or repository reference '%s': %w", ref, err)
	}
	return normalizeDockerHubHost(parsed.Context().RegistryStr()), nil
}

func looksLikeRegistryHost(s string) bool {
	// A bare registry host is identified by a dot in its hostname
	// (ghcr.io, registry.example.com:5000). Strip any :port before the check
	// so an image:tag whose tag contains a dot (nginx:1.25) is not mistaken
	// for a host. Single-label hosts and localhost are not supported here;
	// they resolve as Docker Hub image names.
	host := s
	if i := strings.IndexByte(s, ':'); i >= 0 {
		host = s[:i]
	}
	return strings.Contains(host, ".")
}

// This used when users want to register a registry credential for a registry host.
// NormalizeRegistryHostInput normalizes a user-supplied registry host — as
// opposed to an image/repository reference — to its canonical form. Unlike
// NormalizeRegistryHost, a bare single-label string like "harbor" is treated
// as a host, never as a Docker Hub repository name.
func NormalizeRegistryHostInput(host string) (string, error) {
	h := strings.TrimSpace(host)
	h = trimSchemePrefix(h)
	h = strings.TrimSuffix(h, "/")
	h = strings.ToLower(h)

	if h == "" || strings.ContainsAny(h, "/@") {
		return "", fmt.Errorf("expected a registry host, got '%s'", host)
	}

	reg, err := name.NewRegistry(h)
	if err != nil {
		return "", fmt.Errorf("expected a registry host, got '%s': %w", host, err)
	}
	return normalizeDockerHubHost(reg.RegistryStr()), nil
}

func trimSchemePrefix(s string) string {
	lower := strings.ToLower(s)
	switch {
	case strings.HasPrefix(lower, "https://"):
		return s[len("https://"):]
	case strings.HasPrefix(lower, "http://"):
		return s[len("http://"):]
	default:
		return s
	}
}

func normalizeDockerHubHost(host string) string {
	switch host {
	case dockerHubAliasHost, dockerHubRegistryAliasHost:
		return name.DefaultRegistry
	}
	return host
}

//go:generate mockgen -destination=../mocks/mock_registry_client.go -package=mocks github.com/Stackdome/stackdome/pkg/clients RegistryClient
type RegistryClient interface {
	CheckImage(ctx context.Context, imageRef string) (bool, error)
	// CheckPushAccess verifies that the client's credentials can authorize a
	// push to repoRef. It requests a push-scoped token and opens (then
	// cancels) a blob upload; it never pushes any content.
	CheckPushAccess(ctx context.Context, repoRef string) error
}

// registryClient provides functionalities to interact with a container registry.
type registryClient struct {
	auth authn.Authenticator
}

func NewRegistryClientWithAuth(username, password string) (RegistryClient, error) {
	var auth authn.Authenticator
	if username != "" && password != "" {
		auth = &authn.Basic{
			Username: username,
			Password: password,
		}
	} else {
		return nil, fmt.Errorf("username and password must be provided for authentication")
	}
	return &registryClient{
		auth: auth,
	}, nil
}

func NewRegistryClientAnonymous() (RegistryClient, error) {
	return &registryClient{
		auth: authn.Anonymous,
	}, nil
}

// CheckImage checks if the specified image exists and is pullable
func (ic *registryClient) CheckImage(ctx context.Context, imageRef string) (bool, error) {
	// Parse the image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, fmt.Errorf("invalid image reference: %v", err)
	}

	_, err = remote.Get(ref, remote.WithAuth(ic.auth), remote.WithContext(ctx))
	if err != nil {
		if isRateLimitedError(err) {
			return false, fmt.Errorf("%w: %w", ErrRateLimited, err)
		}
		if isAuthError(err) {
			return false, fmt.Errorf("authentication failed for image: %w", ErrAuthFailed)
		} else if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check image: %v", err)
	}

	_, err = remote.Image(ref, remote.WithAuth(ic.auth), remote.WithContext(ctx))
	if err != nil {
		if isRateLimitedError(err) {
			return false, fmt.Errorf("%w: %w", ErrRateLimited, err)
		}
		return false, fmt.Errorf("image exists but not pullable: %v", err)
	}

	return true, nil
}

// CheckPushAccess verifies push access to repoRef without pushing anything.
func (ic *registryClient) CheckPushAccess(ctx context.Context, repoRef string) error {
	repo, err := name.NewRepository(repoRef)
	if err != nil {
		return fmt.Errorf("invalid repository reference: %v", err)
	}

	// CheckPushPermission needs a name.Reference to compute the push scope;
	// the tag itself isn't used for anything beyond that, so DefaultTag is fine.
	ref := repo.Tag(name.DefaultTag)
	// CheckPushPermission doesn't accept a context, so the only way to give
	// callers cancellation/deadline control over the probe is to attach ctx
	// to every request the underlying transport makes.
	rt := ctxRoundTripper{ctx: ctx, base: http.DefaultTransport}
	if err := remote.CheckPushPermission(ref, authKeychain(ic.auth), rt); err != nil {
		if isRateLimitedError(err) {
			return fmt.Errorf("%w: %w", ErrRateLimited, err)
		}
		if isAuthError(err) {
			return fmt.Errorf("push access check failed: %v: %w", err, ErrAuthFailed)
		}
		return fmt.Errorf("push access check failed: %v", err)
	}
	return nil
}

// ctxRoundTripper attaches a fixed context to every request it forwards.
// remote.CheckPushPermission takes an http.RoundTripper but never threads a
// caller context through to the requests it issues, so this is the only seam
// available to make the probe respect cancellation and deadlines.
type ctxRoundTripper struct {
	ctx  context.Context
	base http.RoundTripper
}

func (rt ctxRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.base.RoundTrip(req.WithContext(rt.ctx))
}

// authKeychain adapts a single authn.Authenticator into an authn.Keychain that
// always resolves to it, regardless of the requested resource. go-containerregistry's
// remote.CheckPushPermission takes a Keychain rather than an Authenticator.
func authKeychain(auth authn.Authenticator) authn.Keychain {
	return staticKeychain{auth: auth}
}

type staticKeychain struct {
	auth authn.Authenticator
}

func (k staticKeychain) Resolve(authn.Resource) (authn.Authenticator, error) {
	return k.auth, nil
}

// Helper function to check if error is authentication related. Uses the
// typed transport.Error rather than matching error text, for the same reason
// as isRateLimitedError/isNotFoundError: a token endpoint denial (e.g. GHCR
// rejecting an anonymous token request for a nonexistent repo with 403
// DENIED) must be classified as auth failure rather than a transient error,
// regardless of message wording.
func isAuthError(err error) bool {
	var transportErr *transport.Error
	if !errors.As(err, &transportErr) {
		return false
	}
	if transportErr.StatusCode == http.StatusUnauthorized || transportErr.StatusCode == http.StatusForbidden {
		return true
	}
	for _, diag := range transportErr.Errors {
		if diag.Code == transport.UnauthorizedErrorCode || diag.Code == transport.DeniedErrorCode {
			return true
		}
	}
	return false
}

// Helper function to check if error is rate-limit related. Uses the typed
// transport.Error rather than matching error text, since a 404 whose body
// happens to echo a ref containing "429" would otherwise be misclassified.
func isRateLimitedError(err error) bool {
	var transportErr *transport.Error
	if !errors.As(err, &transportErr) {
		return false
	}
	if transportErr.StatusCode == http.StatusTooManyRequests {
		return true
	}
	for _, diag := range transportErr.Errors {
		if diag.Code == transport.TooManyRequestsErrorCode {
			return true
		}
	}
	return false
}

// Helper function to check if error is not found related. Typed for the same
// reason as isRateLimitedError.
func isNotFoundError(err error) bool {
	var transportErr *transport.Error
	if !errors.As(err, &transportErr) {
		return false
	}
	if transportErr.StatusCode == http.StatusNotFound {
		return true
	}
	for _, diag := range transportErr.Errors {
		if diag.Code == transport.NameUnknownErrorCode || diag.Code == transport.ManifestUnknownErrorCode {
			return true
		}
	}
	return false
}
