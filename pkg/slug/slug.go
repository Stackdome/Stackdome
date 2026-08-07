package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"

	goslug "github.com/gosimple/slug"
)

const MaxSlugLength = 40

var (
	// gosimple/slug allows underscores; DNS-1123 labels do not.
	hyphenRuns   = regexp.MustCompile(`-{2,}`)
	fallbackSlug = "org"
)

// Make normalizes a value into a DNS-safe slug without applying a length cap
// or fallback value.
func Make(value string) string {
	value = goslug.Make(value)
	value = strings.ReplaceAll(value, "_", "-")
	value = hyphenRuns.ReplaceAllString(value, "-")
	return strings.Trim(value, "-")
}

func FromOrgName(name string) string {
	s := Make(name)
	if len(s) > MaxSlugLength {
		s = strings.Trim(s[:MaxSlugLength], "-")
	}
	if s == "" {
		return fallbackSlug
	}
	return s
}

func RandomSuffix() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
