package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
)

func TestPresentPostgresAddon_includesDeclaredOutputs(t *testing.T) {
	addon := &models.PostgresAddon{
		Name: "pg-main",
	}

	out := presenters.PresentPostgresAddon(addon)

	got := make([]string, 0, len(out.Outputs))
	for _, output := range out.Outputs {
		got = append(got, output.Name)
	}

	want := []string{
		"host",
		"port",
		"database",
		"username",
		"password",
		"sslmode",
		"ca_certificate",
		"url",
	}

	if len(got) != len(want) {
		t.Fatalf("unexpected outputs length: got %d want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected output at %d: got %q want %q", i, got[i], want[i])
		}
	}
}
