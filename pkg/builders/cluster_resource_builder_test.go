package builders

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestShouldEnableTLS(t *testing.T) {
	tests := []struct {
		name string
		fqdn string
		want bool
	}{
		{"empty string", "", false},
		{"nip.io subdomain", "app.192-168-1-1.nip.io", false},
		{"sslip.io subdomain", "app.10-0-0-1.sslip.io", false},
		{"dot local", "myapp.local", false},
		{"dot localhost", "myapp.localhost", false},
		{"real domain", "app.example.com", true},
		{"subdomain", "api.staging.example.com", true},
		{"bare domain", "example.com", true},
		{"io TLD not matching nip.io", "myapp.io", true},
		{"domain ending in local but not .local suffix", "app.mylocal", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldEnableTLS(tt.fqdn)
			if got != tt.want {
				t.Errorf("shouldEnableTLS(%q) = %v, want %v", tt.fqdn, got, tt.want)
			}
		})
	}
}

func TestHasTLSPorts(t *testing.T) {
	tests := []struct {
		name string
		spec *corev1alpha1.StackResourceSpec
		want bool
	}{
		{
			"no ports",
			&corev1alpha1.StackResourceSpec{},
			false,
		},
		{
			"ports without TLS",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 8080, TLS: false},
					{Number: 3000, TLS: false},
				},
			},
			false,
		},
		{
			"single TLS port",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 443, TLS: true},
				},
			},
			true,
		},
		{
			"mixed ports",
			&corev1alpha1.StackResourceSpec{
				Ports: []corev1alpha1.Port{
					{Number: 8080, TLS: false},
					{Number: 443, TLS: true},
				},
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasTLSPorts(tt.spec)
			if got != tt.want {
				t.Errorf("hasTLSPorts() = %v, want %v", got, tt.want)
			}
		})
	}
}


func TestBuildStackResourceCR_AnnotationMerge(t *testing.T) {
	builder := &clusterResourceBuilder{}

	t.Run("sets annotation when TLS ports exist", func(t *testing.T) {
		sr := &models.StackResource{
			Name:      "web",
			Namespace: "default",
			StackID:   "stack-1",
			Ports: []models.Port{
				{Number: 443, ExposedToPublic: true, ExposedFqdn: "app.example.com", Protocol: "http"},
			},
		}

		cr, err := builder.BuildStackResourceCR(sr, "test-stack")
		if err != nil {
			t.Fatalf("BuildStackResourceCR() error = %v", err)
		}

		val, ok := cr.Annotations[corev1alpha1.ClusterIssuerAnnotation]
		if !ok {
			t.Fatal("expected cluster issuer annotation to be set")
		}
		if val != models.DefaultClusterIssuerName {
			t.Errorf("annotation value = %q, want %q", val, models.DefaultClusterIssuerName)
		}
	})

	t.Run("preserves existing labels when adding annotation", func(t *testing.T) {
		sr := &models.StackResource{
			Name:      "web",
			Namespace: "default",
			StackID:   "stack-1",
			Version:   3,
			Ports: []models.Port{
				{Number: 443, ExposedToPublic: true, ExposedFqdn: "app.example.com", Protocol: "http"},
			},
		}

		cr, err := builder.BuildStackResourceCR(sr, "test-stack")
		if err != nil {
			t.Fatalf("BuildStackResourceCR() error = %v", err)
		}

		if cr.Labels[corev1alpha1.LabelStackID] != "stack-1" {
			t.Errorf("LabelStackID = %q, want %q", cr.Labels[corev1alpha1.LabelStackID], "stack-1")
		}
		if _, ok := cr.Annotations[corev1alpha1.ClusterIssuerAnnotation]; !ok {
			t.Fatal("expected cluster issuer annotation to be set")
		}
	})
}
