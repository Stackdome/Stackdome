package builders

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/go-containerregistry/pkg/name"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// tokenAuthPlaceholderUsername is used when syncing token-based git
// credentials, which don't require a real username.
const tokenAuthPlaceholderUsername = "username"

type SecretBuilder interface {
	// BuildDockerConfigJsonSecret synthesizes a dockerconfigjson secret from
	// raw credentials — used for org-level registry credentials. ref may be an
	// image or repository reference; insecure derives an http:// auth URL for
	// in-cluster/plain-HTTP registries.
	BuildDockerConfigJsonSecret(ctx context.Context, secretName, username, password, ref string, insecure bool) (*corev1.Secret, error)
	// BuildGitCredentialsSecretFromCredentials synthesizes a git credentials
	// secret from raw credentials — used for org-level git integrations and
	// minted GitHub App tokens that have no backing hub secret.
	BuildGitCredentialsSecretFromCredentials(ctx context.Context, secretName string, creds gitclient.GitCredentials) (*corev1.Secret, error)
}

type secretBuilder struct{}

type SecretBuilderSpec struct{}

func NewSecretBuilder(spec SecretBuilderSpec) SecretBuilder {
	return &secretBuilder{}
}

func (b *secretBuilder) BuildDockerConfigJsonSecret(ctx context.Context, secretName, username, password, ref string, insecure bool) (*corev1.Secret, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required to build a docker config secret")
	}
	authURL, err := getAuthURLFromImage(ref, insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth URL from ref: %w", err)
	}

	dockerConfigJson, err := marshalDockerConfigJson(authURL, username, password)
	if err != nil {
		return nil, err
	}

	// Namespace is not added, the caller is supposed to add it
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJson,
		},
	}, nil
}

func (b *secretBuilder) BuildGitCredentialsSecretFromCredentials(ctx context.Context, secretName string, creds gitclient.GitCredentials) (*corev1.Secret, error) {
	username := creds.Username
	password := creds.Password
	if creds.Token != "" {
		if username == "" {
			// Token-based auth doesn't require a username; mirror the
			// placeholder used for token-backed hub secrets.
			username = tokenAuthPlaceholderUsername
		}
		password = creds.Token
	}
	if username == "" || password == "" {
		return nil, fmt.Errorf("git credentials are required to build a credentials secret")
	}

	// Namespace is not added, the caller is supposed to add it
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secretName,
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		StringData: map[string]string{
			models.UsernameSecretKey: username,
			models.PasswordSecretKey: password,
		},
	}, nil
}

func marshalDockerConfigJson(authURL, username, password string) ([]byte, error) {
	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			authURL: map[string]interface{}{
				"username": username,
				"password": password,
				"auth":     base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
			},
		},
	}

	dockerConfigJson, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config JSON: %w", err)
	}
	return dockerConfigJson, nil
}

func getAuthURLFromImage(imageURL string, insecure bool) (string, error) {
	var (
		ref name.Reference
		err error
	)
	if insecure {
		ref, err = name.ParseReference(imageURL, name.Insecure)
	} else {
		ref, err = name.ParseReference(imageURL)
	}
	if err != nil {
		return "", fmt.Errorf("failed to parse image URL %s: %w", imageURL, err)
	}

	// Get the registry hostname
	registry := ref.Context().RegistryStr()

	// Handle different registry types
	switch {
	case registry == "index.docker.io":
		return "https://index.docker.io/v1/", nil
	case registry == "ghcr.io":
		return "https://ghcr.io", nil
	case strings.HasSuffix(registry, ".gcr.io") || strings.HasSuffix(registry, ".pkg.dev"):
		return fmt.Sprintf("https://%s", registry), nil
	case strings.Contains(registry, ".amazonaws.com") && strings.Contains(registry, ".ecr."):
		return fmt.Sprintf("https://%s", registry), nil
	case strings.HasSuffix(registry, ".svc.cluster.local"):
		// For in-cluster registries, use HTTP if it's a .local domain
		return fmt.Sprintf("http://%s", registry), nil
	default:
		// Default to HTTPS for generic registries
		return fmt.Sprintf("https://%s", registry), nil
	}
}
