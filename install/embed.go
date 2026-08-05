package install

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"
)

//go:embed manifests/*.yaml
var manifestsFS embed.FS

type TemplateValues struct {
	DBPassword     string
	JWTSecret      string
	EncryptionKey  string
	AdminEmail     string
	AdminPassword  string
	Domain         string
	APIServerImage string
	DBWorkloadType string
	StackUID       string
	TLSEnabled     bool
	// GitHub OAuth login credentials; empty leaves login disabled.
	GitHubClientID     string
	GitHubClientSecret string
	// Platform-wide GitHub App; empty leaves each org creating its own.
	GitHubAppID            string
	GitHubAppSlug          string
	GitHubAppPrivateKey    string
	GitHubAppWebhookSecret string
}

// manifestFuncs renders a Go string as a quoted YAML scalar — JSON encoding is
// valid YAML — so multi-line values such as a PEM survive templating.
var manifestFuncs = template.FuncMap{
	"yamlStr": func(s string) (string, error) {
		b, err := json.Marshal(s)
		return string(b), err
	},
}

func RenderManifest(name string, vals TemplateValues) ([]byte, error) {
	path := "manifests/" + name
	raw, err := manifestsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded manifest %s: %w", path, err)
	}

	tmpl, err := template.New(name).Funcs(manifestFuncs).Parse(string(raw))
	if err != nil {
		return nil, fmt.Errorf("parsing template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vals); err != nil {
		return nil, fmt.Errorf("rendering template %s: %w", name, err)
	}

	return buf.Bytes(), nil
}

func ReadManifest(name string) ([]byte, error) {
	path := "manifests/" + name
	raw, err := manifestsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded manifest %s: %w", path, err)
	}
	return raw, nil
}
