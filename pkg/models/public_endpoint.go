package models

import "strings"

// PublicEndpointConfig is the single source of truth for public ingress TLS.
// The cluster resource builder and public URL outputs both consult it so a
// resource's port TLS matches the scheme emitted in its outputs.
type PublicEndpointConfig struct {
	SharedCompute      bool
	PlatformTLSEnabled bool
}

// UsesTLS reports whether a public port terminates TLS at the ingress. It
// ignores the port's internal protocol: an application serving plain HTTP on
// 8080 is still reachable over HTTPS when the public endpoint has TLS.
func (c PublicEndpointConfig) UsesTLS(port Port) bool {
	if !port.ExposedToPublic || port.ExposedFqdn == "" {
		return false
	}
	if c.SharedCompute && !c.PlatformTLSEnabled {
		return false
	}
	fqdn := strings.ToLower(port.ExposedFqdn)
	for _, suffix := range []string{".nip.io", ".sslip.io", ".local", ".localhost"} {
		if strings.HasSuffix(fqdn, suffix) {
			return false
		}
	}
	return true
}
