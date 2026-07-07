// pkg/errors/validation_test.go
package errors

import (
	"net/http"
	"testing"
)

func TestValidationFailedAggregatesFieldErrors(t *testing.T) {
	errs := []FieldError{
		{Field: "ports[0].protocol", Code: VErrPublicPortNotHTTP, Message: "port 'metrics' (tcp) cannot be exposed to public"},
		{Field: "depends_on[1]", Code: VErrDependencyUnknown, Message: "resource 'redis' does not exist in stack"},
	}
	serr := ValidationFailed(errs)

	if serr.HttpCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", serr.HttpCode)
	}
	if serr.Code != ErrorValidation {
		t.Fatalf("expected ErrorValidation code, got %d", serr.Code)
	}
	apiErr := serr.AsOpenapiError()
	det := apiErr.Details["errors"].([]interface{})
	if len(det) != 2 {
		t.Fatalf("expected 2 detail errors, got %d", len(det))
	}
	first := det[0].(map[string]interface{})
	if first["code"] != VErrPublicPortNotHTTP {
		t.Fatalf("unexpected code: %v", first["code"])
	}
	if first["field"] != "ports[0].protocol" {
		t.Fatalf("unexpected field: %v", first["field"])
	}
}
