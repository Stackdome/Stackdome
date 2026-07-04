package presenters_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

func TestPresentGitIntegration_neverSerializesAuth(t *testing.T) {
	presented := presenters.PresentGitIntegration(&models.GitIntegration{
		ID:             "gi-1",
		OrganisationID: "org-1",
		Type:           models.GitIntegrationTypeGitCredentials,
		Host:           "gitlab.example.com",
		Status:         models.GitIntegrationStatusActive,
		EncryptedAuth:  "encrypted-should-never-leak",
		DataHash:       "hash-should-never-leak",
		Auth: &models.GitIntegrationAuth{
			Token: "token-should-never-leak",
			Basic: &models.GitIntegrationBasicAuth{Username: "user-leak", Password: "pass-leak"},
		},
	})

	raw, err := json.Marshal(presented)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	body := string(raw)
	for _, leaked := range []string{"auth", "token-should-never-leak", "user-leak", "pass-leak", "encrypted-should-never-leak", "hash-should-never-leak"} {
		if strings.Contains(body, leaked) {
			t.Fatalf("presented integration leaks %q: %s", leaked, body)
		}
	}
	if !presented.GetCredentialsConfigured() {
		t.Fatal("expected credentials_configured to be set")
	}
	if !strings.Contains(body, "gitlab.example.com") {
		t.Fatalf("presented integration missing host: %s", body)
	}
}

func TestConvertGitIntegration_mapsAuth(t *testing.T) {
	in := openapi.GitIntegration{}
	in.SetHost("gitlab.example.com")
	auth := openapi.GitIntegrationAuth{}
	auth.SetToken("tok-123")
	in.SetAuth(auth)

	converted := presenters.ConvertGitIntegration(&in)
	if converted.Host != "gitlab.example.com" {
		t.Fatalf("unexpected host %q", converted.Host)
	}
	if converted.Auth == nil || converted.Auth.Token != "tok-123" {
		t.Fatalf("unexpected auth %+v", converted.Auth)
	}
	if converted.Type != models.GitIntegrationTypeGitCredentials {
		t.Fatalf("expected default type, got %q", converted.Type)
	}
}
