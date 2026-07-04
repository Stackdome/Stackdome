package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=git_installation.go -destination=../mocks/mock_git_installation_store.go -package=mocks
type GitInstallationStore interface {
	// Upsert creates or updates the installation keyed by
	// (git_integration_id, installation_id).
	Upsert(ctx context.Context, installation *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError)
	ListByIntegrationID(ctx context.Context, integrationID string) ([]*models.GitInstallation, *errors.ServiceError)
	GetByIntegrationAndAccount(ctx context.Context, integrationID, accountLogin string) (*models.GitInstallation, *errors.ServiceError)
	DeleteByInstallationID(ctx context.Context, integrationID string, installationID int64) *errors.ServiceError
	DeleteByIntegrationID(ctx context.Context, integrationID string) *errors.ServiceError
}
