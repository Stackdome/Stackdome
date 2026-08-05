package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/Stackdome/stackdome/install"
)

const (
	bootstrapSecretName      = "stackdome-bootstrap-secrets"
	bootstrapSecretNamespace = "stackdome-control-plane"
)

type BootstrapSecrets struct {
	DBPassword             string
	JWTSecret              string
	EncryptionKey          string
	AdminPassword          string
	AdminEmail             string
	GitHubClientID         string
	GitHubClientSecret     string
	GitHubAppID            string
	GitHubAppSlug          string
	GitHubAppPrivateKey    string
	GitHubAppWebhookSecret string
}

// githubFlags are the GitHub credentials install and upgrade both accept.
// Empty values keep whatever the bootstrap secret already stores.
type githubFlags struct {
	clientID         *string
	clientSecret     *string
	appID            *string
	appSlug          *string
	appKeyFile       *string
	appWebhookSecret *string
}

func registerGitHubFlags(fs *flag.FlagSet) *githubFlags {
	return &githubFlags{
		clientID:         fs.String("github-client-id", "", "GitHub client ID for 'Sign in with GitHub'"),
		clientSecret:     fs.String("github-client-secret", "", "GitHub client secret for 'Sign in with GitHub'"),
		appID:            fs.String("github-app-id", "", "Platform GitHub App numeric ID"),
		appSlug:          fs.String("github-app-slug", "", "Platform GitHub App URL slug"),
		appKeyFile:       fs.String("github-app-key-file", "", "Path to the platform GitHub App private key (PEM)"),
		appWebhookSecret: fs.String("github-app-webhook-secret", "", "Platform GitHub App webhook secret"),
	}
}

func (f *githubFlags) applyTo(vals *install.TemplateValues) error {
	// The hub enables the platform app only when all four values are set, so a
	// partial set would silently disable the feature.
	appFlagsGiven := 0
	for _, v := range []string{*f.appID, *f.appSlug, *f.appKeyFile, *f.appWebhookSecret} {
		if v != "" {
			appFlagsGiven++
		}
	}
	if appFlagsGiven != 0 && appFlagsGiven != 4 {
		return fmt.Errorf("--github-app-id, --github-app-slug, --github-app-key-file and --github-app-webhook-secret must be given together")
	}

	vals.GitHubClientID = *f.clientID
	vals.GitHubClientSecret = *f.clientSecret
	vals.GitHubAppID = *f.appID
	vals.GitHubAppSlug = *f.appSlug
	vals.GitHubAppWebhookSecret = *f.appWebhookSecret

	if *f.appKeyFile == "" {
		return nil
	}
	pem, err := os.ReadFile(*f.appKeyFile)
	if err != nil {
		return fmt.Errorf("reading GitHub App private key: %w", err)
	}
	vals.GitHubAppPrivateKey = string(pem)
	return nil
}

func loadOrCreateSecrets(vals *install.TemplateValues) (*BootstrapSecrets, error) {
	existing, err := readExistingSecrets()
	if err == nil && existing != nil {
		stepLog("Found existing bootstrap secrets -- reusing")
		vals.DBPassword = existing.DBPassword
		vals.JWTSecret = existing.JWTSecret
		vals.EncryptionKey = existing.EncryptionKey
		vals.AdminPassword = existing.AdminPassword
		if err := mergeGitHubConfig(vals, existing); err != nil {
			return nil, err
		}
		return existing, nil
	}

	stepLog("Generating new secrets")
	secrets := &BootstrapSecrets{
		DBPassword:    generateAlphanumeric(24),
		JWTSecret:     generateBase64(48),
		EncryptionKey: generateHex(32),
		AdminPassword: generateAlphanumeric(16),
	}

	vals.DBPassword = secrets.DBPassword
	vals.JWTSecret = secrets.JWTSecret
	vals.EncryptionKey = secrets.EncryptionKey
	vals.AdminPassword = secrets.AdminPassword

	manifest, err := install.RenderManifest("bootstrap-secret.yaml", *vals)
	if err != nil {
		return nil, fmt.Errorf("rendering bootstrap secret: %w", err)
	}

	if err := kubectlApply(manifest); err != nil {
		return nil, fmt.Errorf("creating bootstrap secret: %w", err)
	}

	return secrets, nil
}

// mergeGitHubConfig reconciles the --github-* flags already on vals with what
// the bootstrap secret holds: flags win, stored values fill the gaps, and the
// secret is rewritten only when something actually changed. This is what keeps
// GitHub credentials alive across upgrades, which re-render every manifest.
func mergeGitHubConfig(vals *install.TemplateValues, existing *BootstrapSecrets) error {
	fields := []struct{ flag, stored *string }{
		{&vals.GitHubClientID, &existing.GitHubClientID},
		{&vals.GitHubClientSecret, &existing.GitHubClientSecret},
		{&vals.GitHubAppID, &existing.GitHubAppID},
		{&vals.GitHubAppSlug, &existing.GitHubAppSlug},
		{&vals.GitHubAppPrivateKey, &existing.GitHubAppPrivateKey},
		{&vals.GitHubAppWebhookSecret, &existing.GitHubAppWebhookSecret},
	}
	changed := false
	for _, f := range fields {
		if *f.flag == "" {
			*f.flag = *f.stored
		}
		if *f.flag != *f.stored {
			*f.stored = *f.flag
			changed = true
		}
	}
	// The secret is re-rendered whole, so every field it carries must be on
	// vals — upgrade only knows the email from the secret itself.
	if vals.AdminEmail == "" {
		vals.AdminEmail = existing.AdminEmail
	}
	if !changed {
		return nil
	}

	manifest, err := install.RenderManifest("bootstrap-secret.yaml", *vals)
	if err != nil {
		return fmt.Errorf("rendering bootstrap secret: %w", err)
	}
	if err := kubectlApply(manifest); err != nil {
		return fmt.Errorf("updating bootstrap secret: %w", err)
	}
	stepLog("GitHub credentials stored")
	return nil
}

func readExistingSecrets() (*BootstrapSecrets, error) {
	out, err := output("kubectl", "get", "secret", bootstrapSecretName,
		"-n", bootstrapSecretNamespace,
		"-o", "json")
	if err != nil {
		return nil, err
	}

	var secret struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal([]byte(out), &secret); err != nil {
		return nil, fmt.Errorf("parsing secret JSON: %w", err)
	}

	decode := func(key string) string {
		v, ok := secret.Data[key]
		if !ok {
			return ""
		}
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return ""
		}
		return string(b)
	}

	s := &BootstrapSecrets{
		DBPassword:         decode("db-password"),
		JWTSecret:          decode("jwt-secret"),
		EncryptionKey:      decode("encryption-key"),
		AdminPassword:      decode("admin-password"),
		AdminEmail:         decode("admin-email"),
		GitHubClientID:     decode("github-client-id"),
		GitHubClientSecret: decode("github-client-secret"),

		GitHubAppID:            decode("github-app-id"),
		GitHubAppSlug:          decode("github-app-slug"),
		GitHubAppPrivateKey:    decode("github-app-private-key"),
		GitHubAppWebhookSecret: decode("github-app-webhook-secret"),
	}

	if s.DBPassword == "" || s.JWTSecret == "" || s.EncryptionKey == "" || s.AdminPassword == "" {
		return nil, fmt.Errorf("incomplete bootstrap secrets")
	}

	return s, nil
}

func generateBase64(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

func generateHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func generateAlphanumeric(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var sb strings.Builder
	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(charset[idx.Int64()])
	}
	return sb.String()
}
