// Package signupprotection applies bot verification and bounded throttles to
// public password signup.
package signupprotection

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

const cloudflareConnectingIPHeader = "CF-Connecting-IP"

// ErrClientIPUnavailable means the configured source did not contain one valid address.
var ErrClientIPUnavailable = errors.New("client IP unavailable")

// ClientIPResolver derives a client address from one configured source.
type ClientIPResolver interface {
	Resolve(*http.Request) (netip.Addr, error)
}

type cloudflareClientIPResolver struct{}

// NewCloudflareClientIPResolver trusts exactly one CF-Connecting-IP value.
// The origin must accept traffic exclusively from Cloudflare when this is used.
func NewCloudflareClientIPResolver() ClientIPResolver {
	return cloudflareClientIPResolver{}
}

func (cloudflareClientIPResolver) Resolve(request *http.Request) (netip.Addr, error) {
	values := request.Header.Values(cloudflareConnectingIPHeader)
	if len(values) != 1 || strings.Contains(values[0], ",") {
		return netip.Addr{}, ErrClientIPUnavailable
	}

	return parseClientIP(strings.TrimSpace(values[0]))
}

type directClientIPResolver struct{}

// NewDirectClientIPResolver reads the network peer and ignores proxy headers.
func NewDirectClientIPResolver() ClientIPResolver {
	return directClientIPResolver{}
}

func (directClientIPResolver) Resolve(request *http.Request) (netip.Addr, error) {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return netip.Addr{}, ErrClientIPUnavailable
	}

	return parseClientIP(host)
}

func parseClientIP(value string) (netip.Addr, error) {
	clientIP, err := netip.ParseAddr(value)
	if err != nil || clientIP.Zone() != "" {
		return netip.Addr{}, ErrClientIPUnavailable
	}
	return clientIP.Unmap(), nil
}

var _ ClientIPResolver = cloudflareClientIPResolver{}
var _ ClientIPResolver = directClientIPResolver{}
