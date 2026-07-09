package models

import (
	"testing"
)

func TestReleaseValidationErrorsValueScanRoundTrip(t *testing.T) {
	original := ReleaseValidationErrors{
		{
			ResourceName: "web",
			Field:        "build_config.source_context.git.repo_url",
			Code:         "image_pull_forbidden",
			Message:      "cannot pull image: unauthorized",
		},
		{
			ResourceName: "worker",
			Field:        "image_spec.repository",
			Code:         "image_not_found",
			Message:      "image does not exist",
		},
	}

	value, err := original.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}

	b, ok := value.([]byte)
	if !ok {
		t.Fatalf("expected Value() to return []byte, got %T", value)
	}

	var scanned ReleaseValidationErrors
	if err := scanned.Scan(b); err != nil {
		t.Fatalf("Scan() returned error: %v", err)
	}

	if len(scanned) != len(original) {
		t.Fatalf("expected %d entries, got %d", len(original), len(scanned))
	}
	for i := range original {
		if scanned[i] != original[i] {
			t.Fatalf("entry %d mismatch: expected %+v, got %+v", i, original[i], scanned[i])
		}
	}
}

func TestReleaseValidationErrorsScanNil(t *testing.T) {
	var v ReleaseValidationErrors
	if err := v.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) returned error: %v", err)
	}
	if v != nil {
		t.Fatalf("expected v to remain nil, got %+v", v)
	}
}
