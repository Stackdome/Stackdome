package stackdeploy

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

var validFuncName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// ValidateTemplateKeys checks that all keys in the values map are valid Go
// identifiers and that the template string parses successfully.
func ValidateTemplateKeys(tmplStr string, values map[string]models.OutputValueRef) error {
	funcs := make(template.FuncMap, len(values))
	for k := range values {
		if !validFuncName.MatchString(k) {
			return fmt.Errorf("key %q is not a valid identifier (use letters, digits, and underscores only)", k)
		}
		funcs[k] = func() string { return "" }
	}
	if _, err := template.New("").Funcs(funcs).Parse(tmplStr); err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}
	return nil
}

// ResolveTemplate executes a Go text/template string against the provided data.
// Template keys are registered as zero-arg functions so {{ key }} syntax works
// without a dot prefix. Keys must be valid Go identifiers (letters, digits,
// underscores; cannot start with a digit).
func ResolveTemplate(tmplStr string, data map[string]string) (string, error) {
	funcs := make(template.FuncMap, len(data))
	for k, v := range data {
		if !validFuncName.MatchString(k) {
			return "", fmt.Errorf("template key %q is not a valid identifier (use letters, digits, and underscores only)", k)
		}
		funcs[k] = func() string { return v }
	}
	tmpl, err := template.New("").Option("missingkey=error").Funcs(funcs).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, nil); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}
