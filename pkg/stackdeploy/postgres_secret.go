package stackdeploy

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// PostgresCredentialSecretName returns the deterministic K8s Secret name for
// projected postgres credentials in a stack's namespace. The format is
// pgcred-<8-char-hash>-<sanitized-database>.
func PostgresCredentialSecretName(addonID, database string) string {
	h := sha256.Sum256([]byte(addonID + "/" + database))
	hashPrefix := fmt.Sprintf("%x", h[:4]) // 8 hex chars
	return fmt.Sprintf("pgcred-%s-%s", hashPrefix, sanitizeDNS(database))
}

const defaultDNSLabel = "default"

var dnsUnsafe = regexp.MustCompile(`[^a-z0-9-]`)

// sanitizeDNS lowercases the input, replaces underscores with hyphens,
// strips any remaining characters invalid in a DNS label, and trims
// leading/trailing hyphens.
func sanitizeDNS(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	s = dnsUnsafe.ReplaceAllString(s, "")
	s = strings.Trim(s, "-")
	if s == "" {
		s = defaultDNSLabel
	}
	return s
}
