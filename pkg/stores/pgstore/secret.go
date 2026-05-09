package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SecretStoreSpec struct {
	SessionFactory db.SessionFactory
}

type secretStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewSecretStore(spec SecretStoreSpec) stores.SecretStore {
	return &secretStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *secretStore) Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingCount int64
	if err := tx.Model(&models.Secret{}).
		Where("organisation_id = ? AND user_id = ? AND name = ?",
			secret.OrganisationID, secret.UserID, secret.Name).
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to check duplicate secret name: %s", err.Error())
	}

	if existingCount > 0 {
		tx.Rollback()
		return nil, errors.BadRequest("secret with name '%s' already exists", secret.Name)
	}

	if err := tx.Model(&models.Secret{}).Omit(clause.Associations).Create(secret).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create secret: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit secret creation: %s", err.Error())
	}

	return s.GetByID(ctx, secret.ID)
}

func (s *secretStore) GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError) {
	var secret models.Secret
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&secret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("secret with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get secret: %s", err.Error())
	}
	return &secret, nil
}

func (s *secretStore) GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError) {
	var secret models.Secret
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND name = ?", organisationID, name).
		First(&secret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("secret with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to get secret by name: %s", err.Error())
	}
	return &secret, nil
}

func (s *secretStore) Update(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingSecret models.Secret
	if err := tx.Where("id = ?", secret.ID).First(&existingSecret).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("secret with id '%s' not found", secret.ID)
		}
		return nil, errors.GeneralError("failed to find secret for update: %s", err.Error())
	}

	if secret.Name != existingSecret.Name {
		var nameCount int64
		if err := tx.Model(&models.Secret{}).
			Where("organisation_id = ? AND name = ? AND id != ?",
				secret.OrganisationID, secret.Name, secret.ID).
			Count(&nameCount).Error; err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to check duplicate name: %s", err.Error())
		}

		if nameCount > 0 {
			tx.Rollback()
			return nil, errors.BadRequest("secret with name '%s' already exists", secret.Name)
		}
	}

	if err := tx.Model(&existingSecret).
		Omit(clause.Associations, "id", "organisation_id", "user_id").
		Updates(secret).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update secret: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit secret update: %s", err.Error())
	}

	return s.GetByID(ctx, secret.ID)
}

func (s *secretStore) UpdateEncryptedData(ctx context.Context, secretID string, encryptedData string, keys []string, dataHash string) *errors.ServiceError {
	updates := map[string]interface{}{
		"encrypted_data": encryptedData,
		"keys":           keys,
		"data_hash":      dataHash,
	}

	if err := s.sessionFactory.New(ctx).
		Model(&models.Secret{}).
		Where("id = ?", secretID).
		Updates(updates).Error; err != nil {
		return errors.GeneralError("failed to update secret encrypted data: %s", err.Error())
	}

	return nil
}

func (s *secretStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Delete(&models.Secret{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete secret: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("secret with id '%s' not found", ID)
	}

	return nil
}

func (s *secretStore) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Secret, *errors.ServiceError) {
	var secrets []*models.Secret

	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ?", organisationID).
		Order("created_at DESC").
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list secrets: %s", err.Error())
	}

	return secrets, nil
}

func (s *secretStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.Secret, *errors.ServiceError) {
	var secrets []*models.Secret

	if err := s.sessionFactory.New(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list secrets by team: %s", err.Error())
	}

	return secrets, nil
}

func (s *secretStore) ListByTeamIDs(ctx context.Context, teamIDs []string) ([]*models.Secret, *errors.ServiceError) {
	if len(teamIDs) == 0 {
		return []*models.Secret{}, nil
	}
	var secrets []*models.Secret
	if err := s.sessionFactory.New(ctx).
		Where("team_id IN ?", teamIDs).
		Order("created_at DESC").
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list secrets by teams: %s", err.Error())
	}
	return secrets, nil
}

func (s *secretStore) ListByUser(ctx context.Context, organisationID, userID string) ([]*models.Secret, *errors.ServiceError) {
	var secrets []*models.Secret

	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND user_id = ?", organisationID, userID).
		Order("created_at DESC").
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list user secrets: %s", err.Error())
	}

	return secrets, nil
}

func (s *secretStore) ListByType(ctx context.Context, organisationID, secretType models.SecretType) ([]*models.Secret, *errors.ServiceError) {
	var secrets []*models.Secret

	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND type = ?", organisationID, secretType).
		Order("created_at DESC").
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list secrets by type: %s", err.Error())
	}

	return secrets, nil
}

func (s *secretStore) ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.Secret{}).
		Where("id = ?", secretID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate secret existence: %s", err.Error())
	}
	return count > 0, nil
}

func (s *secretStore) GetSecretKeys(ctx context.Context, secretID string) ([]string, *errors.ServiceError) {
	var secret models.Secret
	if err := s.sessionFactory.New(ctx).Select("keys").Where("id = ?", secretID).First(&secret).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("secret with id '%s' not found", secretID)
		}
		return nil, errors.GeneralError("failed to get secret keys: %s", err.Error())
	}

	return secret.Keys, nil
}

func (s *secretStore) ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError) {
	keys, err := s.GetSecretKeys(ctx, secretID)
	if err != nil {
		return false, nil, err
	}

	availableKeys := make(map[string]bool)
	for _, key := range keys {
		availableKeys[key] = true
	}

	var missingKeys []string
	for _, requiredKey := range requiredKeys {
		if !availableKeys[requiredKey] {
			missingKeys = append(missingKeys, requiredKey)
		}
	}

	return len(missingKeys) == 0, missingKeys, nil
}

func (s *secretStore) GetSecretsByIDs(ctx context.Context, secretIDs []string) ([]*models.Secret, *errors.ServiceError) {
	if len(secretIDs) == 0 {
		return []*models.Secret{}, nil
	}

	var secrets []*models.Secret
	if err := s.sessionFactory.New(ctx).
		Where("id IN ?", secretIDs).
		Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to get secrets by IDs: %s", err.Error())
	}

	return secrets, nil
}
