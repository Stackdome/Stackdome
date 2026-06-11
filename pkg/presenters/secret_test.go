package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
)

func TestPresentSecret_includesDeclaredOutputs(t *testing.T) {
	secret := &models.Secret{
		Name: "tls",
		Keys: []string{"JWT_PRIVATE_KEY", "tls.crt"},
	}

	out := presenters.PresentSecret(secret)

	if len(out.Outputs) != 2 {
		t.Fatalf("expected 2 outputs, got %d", len(out.Outputs))
	}
	if out.Outputs[0].Name != "JWT_PRIVATE_KEY" {
		t.Fatalf("unexpected first output: %q", out.Outputs[0].Name)
	}
	if out.Outputs[1].Name != "tls.crt" {
		t.Fatalf("unexpected second output: %q", out.Outputs[1].Name)
	}
	if !out.Outputs[0].Sensitive || !out.Outputs[1].Sensitive {
		t.Fatalf("expected secret outputs to be sensitive")
	}
}
