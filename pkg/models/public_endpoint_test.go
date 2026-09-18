package models

import "testing"

func TestPublicEndpointTLSHostnames(t *testing.T) {
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
			got := (PublicEndpointConfig{}).UsesTLS(Port{ExposedToPublic: true, ExposedFqdn: tt.fqdn})
			if got != tt.want {
				t.Errorf("UsesTLS(%q) = %v, want %v", tt.fqdn, got, tt.want)
			}
		})
	}
}
