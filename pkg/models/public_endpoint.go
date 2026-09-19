package models

import "strings"

// nonTLSDomainSuffixes are hosts that never terminate TLS at the platform
// ingress (raw IPs and local development names).
var nonTLSDomainSuffixes = []string{".nip.io", ".sslip.io", ".local", ".localhost"}

// PublicEndpointConfig is the authoritative ingress TLS policy for public
// endpoints. It is shared by the cluster-resource builder (which configures the
// ingress) and the deployment resolver (which derives public URL outputs) so
// both agree on the scheme.
type PublicEndpointConfig struct {
	// SharedCompute marks platform-managed clusters. On shared compute without
	// platform TLS there is no wildcard certificate, so endpoints stay HTTP.
	SharedCompute      bool
	PlatformTLSEnabled bool
}

// UsesTLS reports whether the public endpoint for a port terminates TLS. It
// depends on configured ingress behavior, not certificate readiness, and
// ignores the application's internal protocol and port.
func (c PublicEndpointConfig) UsesTLS(port Port) bool {
	if !port.ExposedToPublic || port.ExposedFqdn == "" {
		return false
	}
	if c.SharedCompute && !c.PlatformTLSEnabled {
		return false
	}
	for _, suffix := range nonTLSDomainSuffixes {
		if strings.HasSuffix(port.ExposedFqdn, suffix) {
			return false
		}
	}
	return true
}
