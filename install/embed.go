package install

import (
	"bytes"
	"embed"
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
	TLSEnabled     bool
}

func RenderManifest(name string, vals TemplateValues) ([]byte, error) {
	path := "manifests/" + name
	raw, err := manifestsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading embedded manifest %s: %w", path, err)
	}

	tmpl, err := template.New(name).Parse(string(raw))
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
