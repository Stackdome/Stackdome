package slug

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
)

const MaxSlugLength = 40

var (
	nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)
	fallbackSlug    = "org"
)

func FromOrgName(name string) string {
	s := strings.ToLower(name)
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
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
