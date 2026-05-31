package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/install"
)

const (
	bootstrapSecretName      = "stackdome-bootstrap-secrets"
	bootstrapSecretNamespace = "stackdome-control-plane"
)

type BootstrapSecrets struct {
	DBPassword    string
	JWTSecret     string
	EncryptionKey string
	AdminPassword string
}

func loadOrCreateSecrets(vals *install.TemplateValues) (*BootstrapSecrets, error) {
	existing, err := readExistingSecrets()
	if err == nil && existing != nil {
		stepLog("Found existing bootstrap secrets -- reusing")
		vals.DBPassword = existing.DBPassword
		vals.JWTSecret = existing.JWTSecret
		vals.EncryptionKey = existing.EncryptionKey
		vals.AdminPassword = existing.AdminPassword
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
		DBPassword:    decode("db-password"),
		JWTSecret:     decode("jwt-secret"),
		EncryptionKey: decode("encryption-key"),
		AdminPassword: decode("admin-password"),
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
