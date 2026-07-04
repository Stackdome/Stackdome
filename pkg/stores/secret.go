package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=secret.go -destination=../mocks/mock_secret_store.go -package=mocks
type SecretStore interface {
	Create(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.Secret, *errors.ServiceError)
	GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError)
	Update(ctx context.Context, secret *models.Secret) (*models.Secret, *errors.ServiceError)
	UpdateEncryptedData(ctx context.Context, secretID string, encryptedData string, keys []string, dataHash string) *errors.ServiceError
	Delete(ctx context.Context, ID string) *errors.ServiceError
	ListByOrganisation(ctx context.Context, organisationID string, params ListParams) ([]*models.Secret, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string, params ListParams) ([]*models.Secret, *errors.ServiceError)
	ListByTeamIDs(ctx context.Context, teamIDs []string, params ListParams) ([]*models.Secret, *errors.ServiceError)
	ListByUser(ctx context.Context, organisationID, userID string) ([]*models.Secret, *errors.ServiceError)
	ListByType(ctx context.Context, organisationID, secretType models.SecretType) ([]*models.Secret, *errors.ServiceError)
	ValidateSecretExists(ctx context.Context, secretID string) (bool, *errors.ServiceError)
	GetSecretKeys(ctx context.Context, secretID string) ([]string, *errors.ServiceError)
	ValidateSecretHasKeys(ctx context.Context, secretID string, requiredKeys []string) (bool, []string, *errors.ServiceError)
	GetManagedByOwner(ctx context.Context, ownerKind, ownerID string, slot models.ManagedSecretSlot) (*models.Secret, *errors.ServiceError)
	DeleteManagedByOwner(ctx context.Context, ownerKind, ownerID string, slots []models.ManagedSecretSlot) *errors.ServiceError
}
