package builders

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/google/go-containerregistry/pkg/name"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SecretBuilder interface {
	BuildDockerConfigJsonSecretForImage(ctx context.Context, secret *models.Secret, imageUrl string) (*corev1.Secret, error)
	BuildDockerConfigJsonSecretForRepository(ctx context.Context, secret *models.Secret, imageUrl string) (*corev1.Secret, error)
	BuildGitCredentialsSecret(ctx context.Context, secret *models.Secret, repoUrl string) (*corev1.Secret, error)
}

type secretBuilder struct {
	secretFetcher secretFetcher
}

type secretFetcher interface {
	InternalGetByID(ctx context.Context, secretID string) (*models.Secret, *errors.ServiceError)
}

type SecretBuilderSpec struct {
	SecretFetcher secretFetcher
}

func NewSecretBuilder(spec SecretBuilderSpec) SecretBuilder {
	return &secretBuilder{
		secretFetcher: spec.SecretFetcher,
	}
}

func (b *secretBuilder) BuildDockerConfigJsonSecretForImage(ctx context.Context, secret *models.Secret, imageUrl string) (*corev1.Secret, error) {
	secretID := secret.ID
	if secret.Type != models.SecretTypeDockerRegistry {
		return nil, fmt.Errorf("secret with ID %s is not a Docker Registry secret", secretID)
	}
	authURL, err := getAuthURLFromImage(imageUrl, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth URL from image: %w", err)
	}

	userName, ok := secret.Data[models.UsernameSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain username", secretID)
	}
	password, ok := secret.Data[models.PasswordSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain password", secretID)
	}
	if len(userName) == 0 || len(password) == 0 {
		return nil, fmt.Errorf("username or password is empty in secret with ID %s", secretID)
	}

	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			authURL: map[string]interface{}{
				"username": userName,
				"password": password,
				"auth":     base64.StdEncoding.EncodeToString([]byte(userName + ":" + password)),
			},
		},
	}

	dockerConfigJson, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config JSON: %w", err)
	}

	// Namespace is not added, the caller is supposed to add it
	res := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.ClusterSecretName(),
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJson,
		},
	}
	return res, nil
}

func (b *secretBuilder) BuildDockerConfigJsonSecretForRepository(ctx context.Context, secret *models.Secret, repoUrl string) (*corev1.Secret, error) {
	secretID := secret.ID
	if secret.Type != models.SecretTypeDockerRegistry {
		return nil, fmt.Errorf("secret with ID %s is not a Docker Registry secret", secretID)
	}
	authURL, err := getAuthURLFromRepository(repoUrl, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth URL from image: %w", err)
	}

	userName, ok := secret.Data[models.UsernameSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain username", secretID)
	}
	password, ok := secret.Data[models.PasswordSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain password", secretID)
	}
	if len(userName) == 0 || len(password) == 0 {
		return nil, fmt.Errorf("username or password is empty in secret with ID %s", secretID)
	}

	dockerConfig := map[string]interface{}{
		"auths": map[string]interface{}{
			authURL: map[string]interface{}{
				"username": userName,
				"password": password,
				"auth":     base64.StdEncoding.EncodeToString([]byte(userName + ":" + password)),
			},
		},
	}

	dockerConfigJson, err := json.Marshal(dockerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal docker config JSON: %w", err)
	}

	// Namespace is not added, the caller is supposed to add it
	res := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.ClusterSecretName(),
			Annotations: map[string]string{},
			Labels:      map[string]string{},
		},
		Type: corev1.SecretTypeDockerConfigJson,
		Data: map[string][]byte{
			corev1.DockerConfigJsonKey: dockerConfigJson,
		},
	}
	return res, nil
}

func (b *secretBuilder) BuildGitCredentialsSecret(ctx context.Context, secret *models.Secret, repoUrl string) (*corev1.Secret, error) {
	secretID := secret.ID

	if secret.Type != models.SecretTypeGitCredentials {
		return nil, fmt.Errorf("secret with ID %s is not a Git Credentials secret", secretID)
	}

	token, ok := secret.Data[models.TokenSecretKey]
	if ok && len(token) > 0 {
		// If token is present, use it for authentication
		userName := "username" // Placeholder, as token-based auth doesn't require username
		res := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:        secret.ClusterSecretName(),
				Annotations: map[string]string{},
				Labels:      map[string]string{},
			},
			StringData: map[string]string{
				models.UsernameSecretKey: userName,
				models.PasswordSecretKey: token,
			},
		}
		return res, nil
	}

	userName, ok := secret.Data[models.UsernameSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain username", secretID)
	}
	password, ok := secret.Data[models.PasswordSecretKey]
	if !ok {
		return nil, fmt.Errorf("secret with ID %s does not contain password", secretID)
	}
	if len(userName) == 0 || len(password) == 0 {
		return nil, fmt.Errorf("username or password is empty in secret with ID %s", secretID)
	}

	res := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        secret.ClusterSecretName(),
			Annotations: map[string]string{},
		},
		StringData: map[string]string{
			models.UsernameSecretKey: userName,
			models.PasswordSecretKey: password,
		},
	}
	return res, nil
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

func getAuthURLFromRepository(repoUrl string, inSecure bool) (string, error) {
	var opts name.Option
	if inSecure {
		opts = name.Insecure
	}
	repo, err := name.NewRepository(repoUrl, opts)
	if err != nil {
		return "", fmt.Errorf("failed to parse repo url: %w", err)
	}

	registry := repo.Registry.RegistryStr()

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
