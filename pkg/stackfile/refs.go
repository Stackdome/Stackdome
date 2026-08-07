package stackfile

import (
	"regexp"
	"sort"
	"strings"
)

// The env-value reference DSL:
//   - {{ self.port }}            — own output, must be the whole value
//   - {{ redis.host }}           — another resource's output, exact or embedded
//   - redis://{{ redis.host }}   — template; all refs must share one source
//   - addon env uses bare output names: {{ host }}, {{ username }}
var (
	// Matches a full-value ref: the entire string is {{ source.output }}
	exactRefPattern = regexp.MustCompile(`^\{\{\s*([\w-]+(?:\.[\w-]+)+)\s*\}\}$`)
	// Matches embedded refs within a larger string
	embeddedRefPattern = regexp.MustCompile(`\{\{\s*([\w-]+(?:\.[\w-]+)+)\s*\}\}`)
	// addonVarPattern matches {{ varname }} in addon env templates.
	// Addon vars are plain output names (host, port, username) — no source prefix.
	addonVarPattern = regexp.MustCompile(`\{\{\s*([\w-]+)\s*\}\}`)
)

type envRef struct {
	Source   string
	Output   string
	RawMatch string // the exact substring matched, e.g. "{{ redis.host }}"
}

func findRefs(value string) []envRef {
	matches := embeddedRefPattern.FindAllStringSubmatch(value, -1)
	var refs []envRef
	for _, m := range matches {
		parts := strings.SplitN(m[1], ".", 2)
		if len(parts) == 2 {
			refs = append(refs, envRef{Source: parts[0], Output: parts[1], RawMatch: m[0]})
		}
	}
	return refs
}

func isExactRef(value string) bool {
	return exactRefPattern.MatchString(value)
}

func isSelfRef(value string) bool {
	for _, r := range findRefs(value) {
		if r.Source == sourceSelf {
			return true
		}
	}
	return false
}

func extractSelfOutput(value string) string {
	for _, r := range findRefs(value) {
		if r.Source == sourceSelf {
			return r.Output
		}
	}
	return ""
}

func hasResourceRef(value string) bool {
	for _, r := range findRefs(value) {
		if r.Source != sourceSelf {
			return true
		}
	}
	return false
}

// outputToVarName turns an output accessor into a template variable name,
// e.g. "port.http" -> "port_http". Template keys cannot contain dots.
func outputToVarName(output string) string {
	return strings.ReplaceAll(output, ".", "_")
}

type addonRef struct {
	Output   string
	RawMatch string
}

func findAddonRefs(value string) []addonRef {
	matches := addonVarPattern.FindAllStringSubmatch(value, -1)
	var refs []addonRef
	for _, m := range matches {
		refs = append(refs, addonRef{Output: m[1], RawMatch: m[0]})
	}
	return refs
}

// templateVarPattern matches a specific template variable, e.g. {{key}} or {{ key }}.
func templateVarPattern(key string) *regexp.Regexp {
	return regexp.MustCompile(`\{\{\s*` + regexp.QuoteMeta(key) + `\s*\}\}`)
}

// sortedKeys returns the map's keys in stable order. Map iteration is
// randomized in Go; every builder walks maps through this so the generated
// Stack document is byte-identical across runs.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
