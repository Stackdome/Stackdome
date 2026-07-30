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

func FromOrgName(name string) string {
	s := goslug.Make(name)
	s = strings.ReplaceAll(s, "_", "-")
	s = hyphenRuns.ReplaceAllString(s, "-")
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
