package credentials

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

func TestClusterSecretNameForGitHost(t *testing.T) {
	t.Run("short host", func(t *testing.T) {
		got := ClusterSecretNameForGitHost("github.com")
		want := "git-integration-github.com"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
		assertValidClusterSecretName(t, got)
	})

	t.Run("long host is truncated", func(t *testing.T) {
		host := strings.Repeat("a", 100) + ".example.com"
		got := ClusterSecretNameForGitHost(host)
		if len(got) > maxClusterSecretNameLen {
			t.Fatalf("name length %d exceeds max %d: %q", len(got), maxClusterSecretNameLen, got)
		}
		if !strings.HasPrefix(got, gitIntegrationSecretPrefix) {
			t.Fatalf("expected prefix %q, got %q", gitIntegrationSecretPrefix, got)
		}
		assertValidClusterSecretName(t, got)
	})

	t.Run("host with invalid chars is sanitized", func(t *testing.T) {
		got := ClusterSecretNameForGitHost("GitHub.COM:443/path")
		want := "git-integration-github.com-443-path"
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
		assertValidClusterSecretName(t, got)
	})
}

func TestClusterSecretNameForRegistryHost(t *testing.T) {
	host := "reg.example.com"

	for _, tc := range []struct {
		purpose RegistryPurpose
		want    string
	}{
		{RegistryPurposePull, "registry-credential-pull-reg.example.com"},
		{RegistryPurposePush, "registry-credential-push-reg.example.com"},
	} {
		t.Run(string(tc.purpose), func(t *testing.T) {
			got := ClusterSecretNameForRegistryHost(host, tc.purpose)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			assertValidClusterSecretName(t, got)
		})
	}

	t.Run("long host is truncated without colliding prefixes", func(t *testing.T) {
		host := strings.Repeat("b", 100) + ".registry.io"
		pull := ClusterSecretNameForRegistryHost(host, RegistryPurposePull)
		push := ClusterSecretNameForRegistryHost(host, RegistryPurposePush)
		if pull == push {
			t.Fatalf("pull and push names must differ: %q", pull)
		}
		assertValidClusterSecretName(t, pull)
		assertValidClusterSecretName(t, push)
	})

	t.Run("longest purpose leaves room for host suffix", func(t *testing.T) {
		host := strings.Repeat("c", 100)
		got := ClusterSecretNameForRegistryHost(host, RegistryPurposePush)
		if len(got) > maxClusterSecretNameLen {
			t.Fatalf("name length %d exceeds max %d: %q", len(got), maxClusterSecretNameLen, got)
		}
		assertValidClusterSecretName(t, got)
	})
}

func assertValidClusterSecretName(t *testing.T, name string) {
	t.Helper()
	if len(name) == 0 {
		t.Fatal("name must not be empty")
	}
	if len(name) > maxClusterSecretNameLen {
		t.Fatalf("name length %d exceeds max %d: %q", len(name), maxClusterSecretNameLen, name)
	}
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		t.Fatalf("invalid DNS1123 subdomain %q: %v", name, errs)
	}
}
