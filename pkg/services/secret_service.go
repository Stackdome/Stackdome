package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator/secret"
)

type SecretService interface {
	Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError)
	Update(ctx context.Context, id string, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Secret, *errors.ServiceError)
	ListByUser(ctx context.Context, organisationID, userID string) ([]*models.Secret, *errors.ServiceError)
	ListByType(ctx context.Context, organisationID, secretType models.SecretType) ([]*models.Secret, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	GetSecretKeys(ctx context.Context, secretID string) ([]string, *errors.ServiceError)
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
}

type SecretServiceSpec struct {
	SessionFactory    db.SessionFactory
	EncryptionService EncryptionService
	Logger            logger.Logger
}

type secretService struct {
	secretStore       stores.SecretStore
	encryptionService EncryptionService
	validator         validator.SecretValidator
	logger            logger.Logger
}

func NewSecretService(spec SecretServiceSpec) SecretService {
	return &secretService{
		secretStore: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		validator:         secret.NewSecretValidator(),
		encryptionService: spec.EncryptionService,
		logger:            spec.Logger,
	}
}

func (s *secretService) Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	if err := s.validator.ValidateSecretData(secret); err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(secret.Data))
	for key := range secret.Data {
		keys = append(keys, key)
	}
	secret.Keys = keys

	if err := s.encryptSecretData(secret); err != nil {
		return nil, err
	}

	secret.Data = nil

	createdSecret, err := s.secretStore.Create(ctx, secret)
	if err != nil {
		return nil, err
	}

	return createdSecret, nil
}

func (s *secretService) GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError) {
	secret, err := s.secretStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *secretService) GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError) {
	secret, err := s.secretStore.GetByName(ctx, organisationID, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (s *secretService) Update(ctx context.Context, id string, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	existingSecret, err := s.secretStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	secret.ID = existingSecret.ID
	secret.OrganisationID = existingSecret.OrganisationID
	secret.UserID = existingSecret.UserID
	// If data is provided, validate and re-encrypt
	if secret.Data != nil {
		if err := s.validator.ValidateSecretData(secret); err != nil {
			return nil, err
		}
		keys := make([]string, 0, len(secret.Data))
		for key := range secret.Data {
			keys = append(keys, key)
		}
		secret.Keys = keys
		if err := s.encryptSecretData(secret); err != nil {
			return nil, err
		}
		secret.Data = nil
	} else {
		secret.EncryptedData = existingSecret.EncryptedData
		secret.Keys = existingSecret.Keys
		secret.DataHash = existingSecret.DataHash
	}

	updatedSecret, err := s.secretStore.Update(ctx, secret)
	if err != nil {
		return nil, err
	}

	return updatedSecret, nil
}

func (s *secretService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := s.secretStore.Delete(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *secretService) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Secret, *errors.ServiceError) {
	secrets, err := s.secretStore.ListByOrganisation(ctx, organisationID)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (s *secretService) ListByUser(ctx context.Context, organisationID, userID string) ([]*models.Secret, *errors.ServiceError) {
	secrets, err := s.secretStore.ListByUser(ctx, organisationID, userID)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *secretService) ListByType(ctx context.Context, organisationID, secretType models.SecretType) ([]*models.Secret, *errors.ServiceError) {
	secrets, err := s.secretStore.ListByType(ctx, organisationID, secretType)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (s *secretService) ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError) {
	return s.secretStore.ValidateSecretExists(ctx, secretID)
}

func (s *secretService) GetSecretKeys(ctx context.Context, secretID string) ([]string, *errors.ServiceError) {
	return s.secretStore.GetSecretKeys(ctx, secretID)
}

func (s *secretService) ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError) {
	return s.secretStore.ValidateSecretHasKeys(ctx, secretID, requiredKeys)
}

func (s *secretService) encryptSecretData(secret *models.Secret) *errors.ServiceError {
	jsonData, err := json.Marshal(secret.Data)
	if err != nil {
		return errors.GeneralError("failed to serialize secret data: %s", err.Error())
	}

	encryptedData, err := s.encryptionService.EncryptData(jsonData)
	if err != nil {
		return errors.GeneralError("failed to encrypt secret data: %s", err.Error())
	}

	secret.EncryptedData = encryptedData
	secret.DataHash = s.generateDataHash(secret.Data)

	return nil
}

func (s *secretService) decryptSecretData(secret *models.Secret) *errors.ServiceError {
	if secret.EncryptedData == "" {
		return errors.GeneralError("no encrypted data found")
	}

	decryptedBytes, err := s.encryptionService.DecryptData(secret.EncryptedData)
	if err != nil {
		return errors.GeneralError("failed to decrypt secret data: %s", err.Error())
	}

	var data map[string]string
	if err := json.Unmarshal(decryptedBytes, &data); err != nil {
		return errors.GeneralError("failed to deserialize secret data: %s", err.Error())
	}

	secret.Data = data

	// Verify data integrity
	if s.generateDataHash(data) != secret.DataHash {
		return errors.GeneralError("data integrity check failed for secret %s", secret.ID)
	}

	return nil
}

func (s *secretService) generateDataHash(data map[string]string) string {
	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (s *secretService) ValidateDockerRegistrySecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError {
	requiredKeys := []string{"registry", "username", "password"}
	hasKeys, missingKeys, err := s.ValidateSecretHasKeys(ctx, secretID, requiredKeys)

	if err != nil {
		return err
	}

	if !hasKeys {
		return errors.BadRequest("docker registry secret missing required keys: %v", missingKeys)
	}

	return nil
}

func (s *secretService) ValidateGitSecretForStackResource(ctx context.Context, secretID string) *errors.ServiceError {
	keys, err := s.GetSecretKeys(ctx, secretID)
	if err != nil {
		return err
	}

	// Git secrets must have either token OR username/password.
	hasToken := s.containsKey(keys, "token")
	hasUsernamePassword := s.containsKey(keys, "username") && s.containsKey(keys, "password")

	if !hasToken && !hasUsernamePassword {
		return errors.BadRequest("git secret must contain either 'token' or 'username'+'password'")
	}

	return nil
}

func (s *secretService) containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
