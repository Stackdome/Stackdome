package services

import (
	"context"
	"net"
	"strings"
)

// FIXME(hack): first-domain seeding for self-hosted installs infers a base
// domain from the signup request's Host header, carried through context.
// Replace with an explicit "configure your domain" onboarding step (or an
// install-level base-domain setting) and delete this file.

type signupHostKey struct{}

// WithSignupHost stashes the signup request's Host header for the org
// domain seeding hack below.
func WithSignupHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, signupHostKey{}, host)
}

// signupHostBase turns the stashed request host into a wildcard-resolvable
// base domain: real hostnames pass through, localhost/IPs are wrapped in
// nip.io so `<anything>.<ip>.nip.io` resolves back to the same machine.
func signupHostBase(ctx context.Context) string {
	host, _ := ctx.Value(signupHostKey{}).(string)
	if host == "" {
		return ""
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
