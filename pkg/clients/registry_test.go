package clients_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"
)

// repoRefFor turns an httptest server URL into a bare "host:port/repo" reference,
// which name.NewRepository/name.ParseReference expect (no scheme).
func repoRefFor(t *testing.T, srv *httptest.Server, repo string) string {
	t.Helper()
	host := strings.TrimPrefix(srv.URL, "http://")
	return fmt.Sprintf("%s/%s", host, repo)
}

func TestCheckPushAccess_Allowed(t *testing.T) {
	srv := httptest.NewServer(registry.New())
	defer srv.Close()

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	err = client.CheckPushAccess(context.Background(), repoRefFor(t, srv, "some/repo"))
	require.NoError(t, err)
}

func TestCheckPushAccess_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/v2/":
			w.WriteHeader(http.StatusOK)
		case strings.Contains(r.URL.Path, "/blobs/uploads/"):
			http.Error(w, "forbidden", http.StatusForbidden)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
	defer srv.Close()

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	err = client.CheckPushAccess(context.Background(), repoRefFor(t, srv, "some/repo"))
	require.Error(t, err)
}

func TestCheckPushAccess_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, `{"errors":[{"code":"TOOMANYREQUESTS","message":"rate limit exceeded"}]}`)
	}))
	defer srv.Close()

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	err = client.CheckPushAccess(context.Background(), repoRefFor(t, srv, "some/repo"))
	require.Error(t, err)
	require.True(t, errors.Is(err, clients.ErrRateLimited), "expected errors.Is(err, ErrRateLimited); got %v", err)
}

func TestCheckImage_ExistsAfterPush(t *testing.T) {
	srv := httptest.NewServer(registry.New())
	defer srv.Close()

	imageRef := repoRefFor(t, srv, "some/repo") + ":latest"
	ref, err := name.ParseReference(imageRef)
	require.NoError(t, err)
	img, err := random.Image(256, 1)
	require.NoError(t, err)
	require.NoError(t, remote.Write(ref, img))

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	exists, err := client.CheckImage(context.Background(), imageRef)
	require.NoError(t, err)
	require.True(t, exists, "expected pushed image to be reported as pullable")
}

func TestCheckImage_RateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, `{"errors":[{"code":"TOOMANYREQUESTS","message":"rate limit exceeded"}]}`)
	}))
	defer srv.Close()

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	imageRef := repoRefFor(t, srv, "some/repo") + ":latest"
	_, err = client.CheckImage(context.Background(), imageRef)
	require.Error(t, err)
	require.True(t, errors.Is(err, clients.ErrRateLimited), "expected errors.Is(err, ErrRateLimited); got %v", err)
}

func TestNormalizeRegistryHost(t *testing.T) {
	cases := []struct {
		ref  string
		want string
	}{
		{ref: "redis", want: name.DefaultRegistry},
		{ref: "redis:7", want: name.DefaultRegistry},
		{ref: "redis:latest", want: name.DefaultRegistry},
		{ref: "nginx:1.25", want: name.DefaultRegistry},
		{ref: "docker.io/library/redis", want: name.DefaultRegistry},
		{ref: "docker.io", want: name.DefaultRegistry},
		{ref: "registry-1.docker.io/library/redis", want: name.DefaultRegistry},
		{ref: "index.docker.io", want: name.DefaultRegistry},
		{ref: "ghcr.io", want: "ghcr.io"},
		{ref: "ghcr.io/acme/app:v1", want: "ghcr.io"},
		{ref: "ghcr.io/acme/app", want: "ghcr.io"},
		{ref: "registry.example.com:5000", want: "registry.example.com:5000"},
		{ref: "registry.example.com:5000/acme/app:v1.2.3", want: "registry.example.com:5000"},
	}
	for _, tc := range cases {
		t.Run(tc.ref, func(t *testing.T) {
			got, err := clients.NormalizeRegistryHost(tc.ref)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

// TestNormalizeRegistryHostInput covers the host-input variant used for
// user-supplied registry hosts (e.g. registry credential creation), which
// must not treat a bare single-label string like "harbor" as a Docker Hub
// image name the way the ref-parsing NormalizeRegistryHost does.
func TestNormalizeRegistryHostInput(t *testing.T) {
	cases := []struct {
		name string
		host string
		want string
	}{
		{name: "single-label host", host: "harbor", want: "harbor"},
		{name: "single-label host with port", host: "harbor:5000", want: "harbor:5000"},
		{name: "scheme and trailing slash and mixed case", host: "HTTPS://Harbor.Example.com/", want: "harbor.example.com"},
		{name: "http scheme", host: "http://harbor:5000", want: "harbor:5000"},
		{name: "docker.io alias", host: "docker.io", want: name.DefaultRegistry},
		{name: "index.docker.io canonical", host: "index.docker.io", want: name.DefaultRegistry},
		{name: "registry-1.docker.io alias", host: "registry-1.docker.io", want: name.DefaultRegistry},
		{name: "multi-label host", host: "ghcr.io", want: "ghcr.io"},
		{name: "localhost with port", host: "localhost:5000", want: "localhost:5000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := clients.NormalizeRegistryHostInput(tc.host)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeRegistryHostInput_RejectsRepoRefs(t *testing.T) {
	cases := []string{
		"ghcr.io/org/repo",
		"ghcr.io/org/repo:v1",
		"harbor/repo",
		"user@harbor.example.com",
	}
	for _, host := range cases {
		t.Run(host, func(t *testing.T) {
			_, err := clients.NormalizeRegistryHostInput(host)
			require.Error(t, err, "expected an error for input %q", host)
		})
	}
}

// TestCheckImage_NotFoundTagResemblingRateLimit reproduces finding #9: a 404
// whose request URL happens to contain "429" (from the requested tag) must
// not be misclassified as a rate-limit response.
func TestCheckImage_NotFoundTagResemblingRateLimit(t *testing.T) {
	srv := httptest.NewServer(registry.New())
	defer srv.Close()

	imageRef := repoRefFor(t, srv, "some/repo") + ":v1.429"

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	exists, err := client.CheckImage(context.Background(), imageRef)
	require.NoError(t, err, "a not-found tag should soft-fail with exists=false, not an error")
	require.False(t, exists)
	require.False(t, errors.Is(err, clients.ErrRateLimited))
}

// TestCheckPushAccess_ContextCancellation verifies finding #8: a canceled or
// expired context aborts the push-permission probe promptly instead of
// hanging for the OS-level TCP timeout.
func TestCheckPushAccess_ContextCancellation(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block
	}))
	defer func() {
		close(block)
		srv.Close()
	}()

	client, err := clients.NewRegistryClientAnonymous()
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result := make(chan error, 1)
	start := time.Now()
	go func() {
		result <- client.CheckPushAccess(ctx, repoRefFor(t, srv, "some/repo"))
	}()

	select {
	case err := <-result:
		require.Less(t, time.Since(start), 5*time.Second, "probe should return promptly once ctx expires")
		require.Error(t, err)
		require.Contains(t, strings.ToLower(err.Error()), "context deadline exceeded")
	case <-time.After(5 * time.Second):
		t.Fatal("CheckPushAccess did not return after the context deadline expired")
	}
}
