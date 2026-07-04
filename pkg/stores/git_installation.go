package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
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
