package services

import (
	"net"
	"net/url"
	"os"
	"strings"
)

// defaultDomainBase resolves the base domain used to seed a self-hosted
// org's first domain, from APP_BASE_URL in the environment. Hostnames pass
// through; localhost and IPs are wrapped in nip.io so
// `<anything>.<ip>.nip.io` resolves back to the same machine. Empty when
// APP_BASE_URL is unset — no domain is seeded.
func defaultDomainBase() string {
	raw := os.Getenv("APP_BASE_URL")
	if raw == "" {
		return ""
	}
	host := raw
	if u, err := url.Parse(raw); err == nil && u.Host != "" {
		host = u.Host
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.ToLower(host)
	if host == "localhost" {
		return "127.0.0.1.nip.io"
	}
	if ip := net.ParseIP(host); ip != nil {
		return host + ".nip.io"
	}
	return host
}
