package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type RegistryCredentialStoreSpec struct {
	SessionFactory db.SessionFactory
}

type registryCredentialStore struct {
	sessionFactory db.SessionFactory
}

func NewRegistryCredentialStore(spec RegistryCredentialStoreSpec) stores.RegistryCredentialStore {
	return &registryCredentialStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *registryCredentialStore) Create(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Create(credential).Error; err != nil {
		return nil, errors.GeneralError("failed to create registry credential: %s", err.Error())
	}
	return s.GetByID(ctx, credential.ID)
}

func (s *registryCredentialStore) GetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError) {
	var credential models.RegistryCredential
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&credential).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("registry credential with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get registry credential: %s", err.Error())
	}
	return &credential, nil
}

func (s *registryCredentialStore) GetByOrgAndHost(ctx context.Context, organisationID, host string) ([]*models.RegistryCredential, *errors.ServiceError) {
	var credentials []*models.RegistryCredential
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND host = ?", organisationID, host).
		Find(&credentials).Error; err != nil {
		return nil, errors.GeneralError("failed to list registry credentials for host: %s", err.Error())
	}
	return credentials, nil
}

func (s *registryCredentialStore) ListByOrgID(ctx context.Context, organisationID string) ([]*models.RegistryCredential, *errors.ServiceError) {
	var credentials []*models.RegistryCredential
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ?", organisationID).
		Order("host, purpose").
		Find(&credentials).Error; err != nil {
		return nil, errors.GeneralError("failed to list registry credentials: %s", err.Error())
	}
	return credentials, nil
}

func (s *registryCredentialStore) Update(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Save(credential).Error; err != nil {
		return nil, errors.GeneralError("failed to update registry credential: %s", err.Error())
	}
	return s.GetByID(ctx, credential.ID)
}

func (s *registryCredentialStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Where("id = ?", ID).Delete(&models.RegistryCredential{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete registry credential: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("registry credential with id '%s' not found", ID)
	}
	return nil
}
