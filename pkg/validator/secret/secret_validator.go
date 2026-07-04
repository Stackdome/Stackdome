package secret

import (
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
)

type secretValidator struct{}

func NewSecretValidator() validator.SecretValidator {
	return &secretValidator{}
}

func (s *secretValidator) ValidateSecretData(secret *models.Secret) *errors.ServiceError {
	if secret.Name == "" {
		return errors.BadRequest("secret name cannot be empty")
	}

	if secret.Type == "" {
		return errors.BadRequest("secret type cannot be empty")
	}

	if secret.OrganisationID == "" {
		return errors.BadRequest("secret organisation ID cannot be empty")
	}

	if secret.UserID == "" {
		return errors.BadRequest("secret user ID cannot be empty")
	}

	if secret.Data == nil || len(secret.Data) == 0 {
		return errors.BadRequest("secret data cannot be empty")
	}

	return s.ValidateSecretType(secret.Type, secret)
}

func (s *secretValidator) ValidateSecretType(secretType models.SecretType, secret *models.Secret) *errors.ServiceError {
	switch secretType {
	case models.SecretTypeDockerRegistry:
		return s.validateDockerRegistrySecret(secret)
	case models.SecretTypeGitCredentials:
		return s.validateGitCredentialsSecret(secret)
	case models.SecretTypeUsernamePassword:
		return s.validateUsernamePasswordSecret(secret)
	case models.SecretTypeToken:
		return s.validateTokenSecret(secret)
	case models.SecretTypeSSHKey:
		return s.validateSSHKeySecret(secret)
	case models.SecretTypeGeneric:
		return s.validateGenericSecret(secret)
	default:
		return errors.BadRequest("unsupported secret type: %s", secretType)
	}
}

// validateDockerRegistrySecret validates Docker registry secret data
func (s *secretValidator) validateDockerRegistrySecret(secret *models.Secret) *errors.ServiceError {
	requiredFields := []string{models.UsernameSecretKey, models.PasswordSecretKey}

	for _, field := range requiredFields {
		if value, exists := secret.Data[field]; !exists || strings.TrimSpace(value) == "" {
			return errors.BadRequest("docker registry secret requires field: %s", field)
		}
	}

	// Validate registry URL format if provided
	registry := secret.Data[models.RegistrySecretKey]
	if registry != "" {
		// Basic validation - should not contain protocol if it's a domain
		if strings.HasPrefix(registry, "http://") || strings.HasPrefix(registry, "https://") {
			return errors.BadRequest("registry should be a domain name without protocol (e.g., 'docker.io', 'gcr.io')")
		}
	}

	return nil
}

// validateGitCredentialsSecret validates Git credentials secret data
func (s *secretValidator) validateGitCredentialsSecret(secret *models.Secret) *errors.ServiceError {
	// Git credentials can be one of: username/password, token.
	hasUsernamePassword := secret.Data[models.UsernameSecretKey] != "" && secret.Data[models.PasswordSecretKey] != ""
	hasToken := secret.Data[models.TokenSecretKey] != ""
	hasSSHKey := secret.Data[models.SshSecretKey] != ""

	if hasSSHKey {
		return errors.BadRequest("git credentials secret doesnt support ssh_private_key")
	}

	if !hasUsernamePassword && !hasToken {
		return errors.BadRequest("git credentials secret must contain either username/password, token")
	}

	return nil
}

func (s *secretValidator) ValidateGitCredentialsSecret(secret *models.Secret) *errors.ServiceError {
	return s.validateGitCredentialsSecret(secret)
}

// validateUsernamePasswordSecret validates username/password secret data
func (s *secretValidator) validateUsernamePasswordSecret(secret *models.Secret) *errors.ServiceError {
	requiredFields := []string{models.UsernameSecretKey, models.PasswordSecretKey}

	for _, field := range requiredFields {
		if value, exists := secret.Data[field]; !exists || strings.TrimSpace(value) == "" {
			return errors.BadRequest("username/password secret requires field: %s", field)
		}
	}

	return nil
}

// validateTokenSecret validates token secret data
func (s *secretValidator) validateTokenSecret(secret *models.Secret) *errors.ServiceError {
	if value, exists := secret.Data[models.TokenSecretKey]; !exists || strings.TrimSpace(value) == "" {
		return errors.BadRequest("token secret requires field: token")
	}

	token := secret.Data[models.TokenSecretKey]
	if len(token) < 8 {
		return errors.BadRequest("token must be at least 8 characters long")
	}

	return nil
}

// validateSSHKeySecret validates SSH key secret data
func (s *secretValidator) validateSSHKeySecret(secret *models.Secret) *errors.ServiceError {
	if value, exists := secret.Data[models.SshSecretKey]; !exists || strings.TrimSpace(value) == "" {
		return errors.BadRequest("SSH key secret requires field: ssh_private_key")
	}

	sshKey := secret.Data[models.SshSecretKey]
	if !strings.Contains(sshKey, "-----BEGIN") || !strings.Contains(sshKey, "-----END") {
		return errors.BadRequest("ssh_private_key must be in PEM format")
	}
	return nil
}

// validateGenericSecret validates generic secret data
func (s *secretValidator) validateGenericSecret(secret *models.Secret) *errors.ServiceError {
	// Generic secrets just need to have some data
	if len(secret.Data) == 0 {
		return errors.BadRequest("generic secret must contain at least one key-value pair")
	}

	// Validate that keys are not empty strings
	for key, value := range secret.Data {
		if strings.TrimSpace(key) == "" {
			return errors.BadRequest("secret keys cannot be empty")
		}
		if strings.TrimSpace(value) == "" {
			return errors.BadRequest("secret value for key '%s' cannot be empty", key)
		}
	}

	return nil
}
