package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/Stackdome/stackdome/pkg/validator"
	"github.com/Stackdome/stackdome/pkg/validator/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretService interface {
	Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
	// InternalGetByID is used internally to get the secret with decrypted data.
	InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
	InternalGetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError)
	Update(ctx context.Context, id string, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError)
	ListSecretsForCurrentUser(ctx context.Context, orgID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError)
	ListByUser(ctx context.Context, organisationID, userID string) ([]*models.Secret, *errors.ServiceError)
	ListByType(ctx context.Context, organisationID, secretType models.SecretType) ([]*models.Secret, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	GetSecretKeys(ctx context.Context, secretID string) ([]string, *errors.ServiceError)
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
}

type ClusterClientGetter interface {
	GetClient(clusterID string) (client.Client, error)
}

type SecretServiceSpec struct {
	SessionFactory      db.SessionFactory
	EncryptionService   EncryptionService
	ReferenceService    ReferenceService
	ClusterClientGetter ClusterClientGetter
	TeamService         TeamService
	Logger              logger.Logger
	Permissions         auth.PermissionService
}

type secretService struct {
	secretStore         stores.SecretStore
	referenceService    ReferenceService
	encryptionService   EncryptionService
	validator           validator.SecretValidator
	clusterClientGetter ClusterClientGetter
	teamService         TeamService
	logger              logger.Logger
	permissions         auth.PermissionService
}

func NewSecretService(spec SecretServiceSpec) SecretService {
	s := &secretService{
		secretStore: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		referenceService:    spec.ReferenceService,
		validator:           secret.NewSecretValidator(),
		encryptionService:   spec.EncryptionService,
		clusterClientGetter: spec.ClusterClientGetter,
		teamService:         spec.TeamService,
		logger:              spec.Logger,
		permissions:         spec.Permissions,
	}
	return s
}

func (s *secretService) Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, secret.TeamID, auth.ResourceSecrets, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

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

	s.logger.WithFields(map[string]interface{}{
		"secret_id": createdSecret.ID,
		"team_id":   createdSecret.TeamID,
		"key_count": len(createdSecret.Keys),
	}).Info(ctx, "created secret")
	return createdSecret, nil
}

func (s *secretService) GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError) {
	secret, err := s.secretStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, secret.TeamID, auth.ResourceSecrets, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return secret, nil
}

func (s *secretService) InternalGetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError) {
	secret, err := s.secretStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	// Decrypt the secret data for internal use
	if err := s.decryptSecretData(secret); err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *secretService) InternalGetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError) {
	return s.secretStore.GetByName(ctx, organisationID, name)
}

func (s *secretService) GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError) {
	secret, err := s.secretStore.GetByName(ctx, organisationID, name)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, secret.TeamID, auth.ResourceSecrets, secret.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return secret, nil
}

func (s *secretService) Update(ctx context.Context, id string, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	existingSecret, err := s.secretStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, existingSecret.TeamID, auth.ResourceSecrets, id, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	secret.ID = existingSecret.ID
	secret.OrganisationID = existingSecret.OrganisationID
	secret.UserID = existingSecret.UserID
	secret.TeamID = existingSecret.TeamID
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

	s.logger.WithField("secret_id", updatedSecret.ID).Info(ctx, "updated secret")
	return updatedSecret, nil
}

func (s *secretService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	secret, sErr := s.secretStore.GetByID(ctx, ID)
	if sErr != nil {
		return sErr
	}
	if permErr := s.permissions.Check(ctx, secret.TeamID, auth.ResourceSecrets, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}

	inUse, refs, refErr := s.referenceService.IsReferentInUse(ctx, models.ReferentSecret, ID)
	if refErr != nil {
		return refErr
	}
	if inUse {
		return errors.Conflict("secret '%s' is in use by %s and cannot be deleted", ID, describeReferences(refs))
	}

	if err := s.secretStore.Delete(ctx, ID); err != nil {
		return err
	}

	s.logger.WithField("secret_id", ID).Info(ctx, "deleted secret")
	return nil
}

func (s *secretService) ListByOrganisation(ctx context.Context, organisationID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, organisationID, auth.ResourceSecrets, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	secrets, err := s.secretStore.ListByOrganisation(ctx, organisationID, params)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (s *secretService) ListByTeamID(ctx context.Context, teamID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceSecrets, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.secretStore.ListByTeamID(ctx, teamID, params)
}

func (s *secretService) ListSecretsForCurrentUser(ctx context.Context, orgID string, params stores.ListParams) ([]*models.Secret, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}

	if identity.IsOrgAdmin() {
		return s.secretStore.ListByOrganisation(ctx, orgID, params)
	}

	memberships, serr := s.teamService.InternalListUserTeams(ctx, identity.UserID, orgID)
	if serr != nil {
		return nil, serr
	}

	var allowedTeamIDs []string
	for _, m := range memberships {
		if permErr := s.permissions.Check(ctx, m.TeamID, auth.ResourceSecrets, "", auth.ActionList); permErr == nil {
			allowedTeamIDs = append(allowedTeamIDs, m.TeamID)
		}
	}

	return s.secretStore.ListByTeamIDs(ctx, allowedTeamIDs, params)
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
