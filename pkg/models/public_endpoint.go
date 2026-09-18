package models

import "strings"

// PublicEndpointConfig determines the configured ingress TLS behavior used by
// both cluster resources and public URL outputs, independently of certificate
// readiness or the application's internal protocol and port.
type PublicEndpointConfig struct {
	SharedCompute      bool
	PlatformTLSEnabled bool
}

func (c PublicEndpointConfig) UsesTLS(port Port) bool {
	if !port.ExposedToPublic || port.ExposedFqdn == "" {
		return false
	}
	if c.SharedCompute && !c.PlatformTLSEnabled {
		return false
	}
	for _, suffix := range []string{".nip.io", ".sslip.io", ".local", ".localhost"} {
		if strings.HasSuffix(port.ExposedFqdn, suffix) {
			return false
		}
	}
	return true
}
