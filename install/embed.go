package install

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"
)

//go:embed manifests/*.yaml
var manifestsFS embed.FS

type PlatformConfig struct {
	BaseDomain         string
	TLSEnabled         bool
	CloudflareAPIToken string
	ACMEEnvironment    string
}

type SharedComputeConfig struct {
	ClusterAPIURL string
	ClusterCAData string
	ClusterToken  string
}

func (c PlatformConfig) Enabled() bool {
	return c.BaseDomain != ""
}

func (c SharedComputeConfig) IsSet() bool {
	return c.ClusterAPIURL != "" && c.ClusterCAData != "" && c.ClusterToken != ""
}

// BootstrapConfigRevision changes when a pod restart is required to load
// updated Secret-backed bootstrap configuration. It reveals none of the stored
// credentials.
func BootstrapConfigRevision(platform PlatformConfig, shared SharedComputeConfig) string {
	data := fmt.Sprintf("%t", platform.TLSEnabled) + "\x00" + platform.BaseDomain + "\x00" +
		platform.CloudflareAPIToken + "\x00" + platform.ACMEEnvironment + "\x00" +
		shared.ClusterAPIURL + "\x00" + shared.ClusterCAData + "\x00" + shared.ClusterToken
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))[:16]
}

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
	Platform       PlatformConfig
	SharedCompute  SharedComputeConfig
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
	"bootstrapConfigRevision": BootstrapConfigRevision,
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
