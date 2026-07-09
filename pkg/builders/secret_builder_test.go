package builders

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestBuildDockerConfigJsonSecretFromRawCredentials(t *testing.T) {
	builder := NewSecretBuilder(SecretBuilderSpec{})

	secret, err := builder.BuildDockerConfigJsonSecret(context.Background(), "registry-credential-ghcr.io", "acme-bot", "hunter2", "ghcr.io/acme/app", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if secret.Name != "registry-credential-ghcr.io" {
		t.Fatalf("unexpected secret name %q", secret.Name)
	}
	if secret.Type != corev1.SecretTypeDockerConfigJson {
		t.Fatalf("unexpected secret type %q", secret.Type)
	}
	config := string(secret.Data[corev1.DockerConfigJsonKey])
	if !strings.Contains(config, "https://ghcr.io") || !strings.Contains(config, "acme-bot") {
		t.Fatalf("unexpected docker config: %s", config)
	}
}

func TestBuildDockerConfigJsonSecretRequiresCredentials(t *testing.T) {
	builder := NewSecretBuilder(SecretBuilderSpec{})

	if _, err := builder.BuildDockerConfigJsonSecret(context.Background(), "n", "", "", "ghcr.io/acme/app", false); err == nil {
		t.Fatal("expected missing credentials to error")
	}
}
