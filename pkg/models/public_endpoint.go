package models

import "strings"

// PublicEndpointConfig is the configured ingress TLS policy shared by the
// cluster-resource builder and output resolution. Both use it so the external
// URL scheme always matches the port's ingress TLS decision, independently of
// certificate readiness and the application's internal protocol or port.
type PublicEndpointConfig struct {
	SharedCompute      bool
	PlatformTLSEnabled bool
}

// UsesTLS reports whether a public port's ingress terminates TLS.
func (c PublicEndpointConfig) UsesTLS(port Port) bool {
	if !port.ExposedToPublic || port.ExposedFqdn == "" {
		return false
	}
	if c.SharedCompute && !c.PlatformTLSEnabled {
		return false
	}
	nonTLSDomains := []string{".nip.io", ".sslip.io", ".local", ".localhost"}
	for _, suffix := range nonTLSDomains {
		if strings.HasSuffix(port.ExposedFqdn, suffix) {
			return false
		}
	}
	return true
}
