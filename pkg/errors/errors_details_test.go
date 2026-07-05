package errors

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCredentialsRequiredSerializesStructuredDetails(t *testing.T) {
	err := CredentialsRequired(CredentialErrorTarget{
		Kind: CredentialTargetKindImagePull,
		Host: "ghcr.io",
		Ref:  "ghcr.io/acme/app:1",
	}, "image requires credentials")

	if err.HttpCode != 400 {
		t.Fatalf("expected 400, got %d", err.HttpCode)
	}

	apiErr := err.AsOpenapiError()
	raw, marshalErr := json.Marshal(apiErr)
	if marshalErr != nil {
		t.Fatalf("marshal failed: %v", marshalErr)
	}
	body := string(raw)
	for _, want := range []string{ErrorCodeCredentialsRequired, CredentialTargetKindImagePull, "ghcr.io/acme/app:1"} {
		if !strings.Contains(body, want) {
			t.Fatalf("expected serialized error to contain %q: %s", want, body)
		}
	}
}

func TestErrorsWithoutDetailsOmitThem(t *testing.T) {
	apiErr := BadRequest("plain failure").AsOpenapiError()
	raw, marshalErr := json.Marshal(apiErr)
	if marshalErr != nil {
		t.Fatalf("marshal failed: %v", marshalErr)
	}
	if strings.Contains(string(raw), "details") {
		t.Fatalf("expected no details key, got %s", raw)
	}
}
