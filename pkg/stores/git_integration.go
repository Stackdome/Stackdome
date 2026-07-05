package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=git_integration.go -destination=../mocks/mock_git_integration_store.go -package=mocks
type GitIntegrationStore interface {
	Create(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError)
	// GetByOrgAndHost returns the git_credentials integration for the (org,
	// host) pair.
	GetByOrgAndHost(ctx context.Context, organisationID, host string) (*models.GitIntegration, *errors.ServiceError)
	// GetGitHubAppForOrg returns the org's github_app integration.
	GetGitHubAppForOrg(ctx context.Context, organisationID string) (*models.GitIntegration, *errors.ServiceError)
	ListByOrgID(ctx context.Context, organisationID string) ([]*models.GitIntegration, *errors.ServiceError)
	// ListGitHubApps returns every github_app integration across orgs (used
	// to match webhook deliveries by app ID).
	ListGitHubApps(ctx context.Context) ([]*models.GitIntegration, *errors.ServiceError)
	Update(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
}
