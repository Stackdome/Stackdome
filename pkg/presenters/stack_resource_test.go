package presenters_test

import (
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

func TestPresentStackResource_includesPortName(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Ports: models.Ports{
			{
				Name:            "http",
				Number:          8080,
				Protocol:        "http",
				ExposedToPublic: true,
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if len(out.Ports) != 1 {
		t.Fatalf("expected one port, got %d", len(out.Ports))
	}
	if out.Ports[0].Name != "http" {
		t.Fatalf("expected port name http, got %q", out.Ports[0].Name)
	}
}

func TestPresentStackResource_includesDerivedOutputs(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Ports: models.Ports{
			{
				Name:            "http",
				Number:          8080,
				Protocol:        "http",
				ExposedToPublic: true,
			},
			{
				Name:            "metrics",
				Number:          9090,
				Protocol:        "http",
				ExposedToPublic: false,
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	got := make([]string, 0, len(out.Outputs))
	for _, output := range out.Outputs {
		got = append(got, output.Name)
	}

	want := []string{
		"host",
		"port.http",
		"url.http",
		"public_host.http",
		"public_url.http",
		"port.metrics",
		"url.metrics",
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

func TestPresentStackResource_usesModelOutputsWhenAlreadyPopulated(t *testing.T) {
	resource := &models.StackResource{
		Name: "web",
		Outputs: []models.OutputDescriptor{
			{
				Name:      "custom.output",
				Type:      models.OutputValueTypeString,
				Sensitive: true,
			},
		},
	}

	out := presenters.PresentStackResource(resource)

	if len(out.Outputs) != 1 {
		t.Fatalf("expected one output, got %d", len(out.Outputs))
	}
	if out.Outputs[0].Name != "custom.output" {
		t.Fatalf("expected custom output name, got %q", out.Outputs[0].Name)
	}
	if !out.Outputs[0].Sensitive {
		t.Fatalf("expected custom output to remain sensitive")
	}
}
