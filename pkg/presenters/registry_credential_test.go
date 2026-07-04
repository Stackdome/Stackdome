package presenters_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"k8s.io/utils/ptr"
)

func TestPresentRegistryCredential_neverSerializesPassword(t *testing.T) {
	presented := presenters.PresentRegistryCredential(&models.RegistryCredential{
		ID:                "rc-1",
		OrganisationID:    "org-1",
		Host:              "ghcr.io",
		Purpose:           models.RegistryCredentialPurposeBoth,
		Username:          "acme-bot",
		Password:          "plaintext-should-never-leak",
		EncryptedPassword: "encrypted-should-never-leak",
		DataHash:          "hash-should-never-leak",
	})

	raw, err := json.Marshal(presented)
	if err != nil {
		t.Fatalf("failed to marshal presented credential: %v", err)
	}
	body := string(raw)

	for _, leaked := range []string{"password", "plaintext-should-never-leak", "encrypted-should-never-leak", "hash-should-never-leak"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("presented credential leaks %q: %s", leaked, body)
		}
	}
	if !strings.Contains(body, "acme-bot") || !strings.Contains(body, "ghcr.io") {
		t.Fatalf("presented credential missing expected fields: %s", body)
	}
}

func TestConvertRegistryCredential_mapsFields(t *testing.T) {
	in := openapi.RegistryCredential{}
	in.SetHost("ghcr.io")
	in.SetUsername("acme-bot")
	in.SetPassword("hunter2")
	in.SetPurpose(openapi.PULL)

	converted := presenters.ConvertRegistryCredential(&in)
	if converted.Host != "ghcr.io" || converted.Username != "acme-bot" || converted.Password != "hunter2" {
		t.Fatalf("unexpected conversion: %+v", converted)
	}
	if converted.Purpose != models.RegistryCredentialPurposePull {
		t.Fatalf("unexpected purpose: %q", converted.Purpose)
	}
}

func TestConvertRegistryCredentialVerifyPurpose_defaultsToEmpty(t *testing.T) {
	if got := presenters.ConvertRegistryCredentialVerifyPurpose(&openapi.RegistryCredentialVerifyRequest{}); got != "" {
		t.Fatalf("expected empty purpose for omitted field, got %q", got)
	}
	req := openapi.RegistryCredentialVerifyRequest{Purpose: ptr.To(openapi.PUSH)}
	if got := presenters.ConvertRegistryCredentialVerifyPurpose(&req); got != models.RegistryCredentialPurposePush {
		t.Fatalf("expected push purpose, got %q", got)
	}
}
