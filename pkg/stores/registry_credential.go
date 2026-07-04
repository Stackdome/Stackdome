package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=registry_credential.go -destination=../mocks/mock_registry_credential_store.go -package=mocks
type RegistryCredentialStore interface {
	Create(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.RegistryCredential, *errors.ServiceError)
	// GetByOrgAndHost returns all credentials for the (org, normalized host)
	// pair; there can be up to one per purpose.
	GetByOrgAndHost(ctx context.Context, organisationID, host string) ([]*models.RegistryCredential, *errors.ServiceError)
	ListByOrgID(ctx context.Context, organisationID string) ([]*models.RegistryCredential, *errors.ServiceError)
	Update(ctx context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
}
