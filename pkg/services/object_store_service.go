package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator"
	"github.com/ashishmax31/stackdome-api-server/pkg/validator/objectstore"
)

type ObjectStoreService interface {
	Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError)
	Update(ctx context.Context, id string, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError)
	ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError)
	TestConnection(ctx context.Context, objectStoreID string) *errors.ServiceError
	UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError
}

type ObjectStoreServiceSpec struct {
	SessionFactory db.SessionFactory
	SecretService  SecretService
	Logger         logger.Logger
}

type objectStoreService struct {
	objectStoreStore stores.ObjectStoreStore
	secretService    SecretService
	validator        validator.ObjectStoreValidator
	logger           logger.Logger
}

func NewObjectStoreService(spec ObjectStoreServiceSpec) ObjectStoreService {
	return &objectStoreService{
		objectStoreStore: pgstore.NewObjectStoreStore(pgstore.ObjectStoreStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		secretService: spec.SecretService,
		validator:     objectstore.NewObjectStoreValidator(),
		logger:        spec.Logger,
	}
}

func (s *objectStoreService) validateSecretReference(ctx context.Context, ref models.SecretReference, fieldName string) *errors.ServiceError {
	secret, err := s.secretService.GetByID(ctx, ref.SecretID)
	if err != nil {
		if err.Is404() {
			return errors.BadRequest("%s: secret with ID '%s' does not exist", fieldName, ref.SecretID)
		}
		return err
	}

	keyFound := false
	for _, k := range secret.Keys {
		if k == ref.Key {
			keyFound = true
			break
		}
	}
	if !keyFound {
		return errors.BadRequest("%s: key '%s' does not exist in secret '%s'", fieldName, ref.Key, secret.Name)
	}

	return nil
}

func (s *objectStoreService) validateSecretReferences(ctx context.Context, config models.ObjectStoreConfiguration) *errors.ServiceError {
	if config.S3Credentials != nil {
		if err := s.validateSecretReference(ctx, config.S3Credentials.AccessKeyID, "S3 access key ID"); err != nil {
			return err
		}
		if err := s.validateSecretReference(ctx, config.S3Credentials.SecretAccessKey, "S3 secret access key"); err != nil {
			return err
		}
	}
	if config.AzureCredentials != nil {
		if err := s.validateSecretReference(ctx, config.AzureCredentials.ConnectionString, "Azure connection string"); err != nil {
			return err
		}
	}
	if config.GCSCredentials != nil {
		if err := s.validateSecretReference(ctx, config.GCSCredentials.ServiceAccountCredentials, "GCS service account credentials"); err != nil {
			return err
		}
	}
	return nil
}

func (s *objectStoreService) Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	// Validate input
	if err := s.validator.ValidateForCreate(ctx, objectStore); err != nil {
		return nil, err
	}

	// Validate secret references exist
	if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
		return nil, err
	}

	// Set default retention policy if not provided
	if objectStore.RetentionPolicy == "" {
		objectStore.RetentionPolicy = "7d"
	}

	createdObjectStore, err := s.objectStoreStore.Create(ctx, objectStore)
	if err != nil {
		return nil, err
	}

	return createdObjectStore, nil
}

func (s *objectStoreService) GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError) {
	objectStore, err := s.objectStoreStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return objectStore, nil
}

func (s *objectStoreService) GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError) {
	objectStore, err := s.objectStoreStore.GetByName(ctx, organisationID, name)
	if err != nil {
		return nil, err
	}
	return objectStore, nil
}

func (s *objectStoreService) Update(ctx context.Context, id string, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	existingObjectStore, err := s.objectStoreStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate immutable field changes before overwriting
	if err := s.validator.ValidateForUpdate(ctx, existingObjectStore, objectStore); err != nil {
		return nil, err
	}

	// Preserve immutable fields (must be done AFTER validation check)
	objectStore.ID = existingObjectStore.ID
	objectStore.OrganisationID = existingObjectStore.OrganisationID
	objectStore.Name = existingObjectStore.Name

	// Validate the full spec including configuration
	if err := s.validator.ValidateForCreate(ctx, objectStore); err != nil {
		return nil, err
	}

	// Validate secret references exist
	if err := s.validateSecretReferences(ctx, objectStore.Configuration); err != nil {
		return nil, err
	}

	updatedObjectStore, err := s.objectStoreStore.Update(ctx, objectStore)
	if err != nil {
		return nil, err
	}

	return updatedObjectStore, nil
}

func (s *objectStoreService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	// TODO: Check if object store is in use by any PostgreSQL addons
	// This would require querying PostgreSQL addons that reference this object store

	if err := s.objectStoreStore.Delete(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *objectStoreService) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError) {
	objectStores, err := s.objectStoreStore.ListByOrganisation(ctx, organisationID)
	if err != nil {
		return nil, err
	}
	return objectStores, nil
}

func (s *objectStoreService) ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError) {
	return s.objectStoreStore.ValidateObjectStoreExists(ctx, objectStoreID)
}

func (s *objectStoreService) TestConnection(ctx context.Context, objectStoreID string) *errors.ServiceError {
	objectStore, err := s.objectStoreStore.GetByID(ctx, objectStoreID)
	if err != nil {
		return err
	}

	// TODO: Implement actual connection testing based on credential type
	// This would involve:
	// - For S3: Test ListObjects operation
	// - For Azure: Test blob container access
	// - For GCS: Test bucket access

	s.logger.Info(ctx, "Object store connection test not yet implemented", "objectStoreId", objectStore.ID)
	return nil
}

func (s *objectStoreService) UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError {
	return s.objectStoreStore.UpdateStatus(ctx, id, status)
}
