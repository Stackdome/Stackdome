package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=project_store.go -destination=../mocks/mock_project_store.go -package=mocks
type ProjectStore interface {
	Create(ctx context.Context, project *models.Project) (*models.Project, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Project, *errors.ServiceError)
	GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string) ([]*models.Project, *errors.ServiceError)
	Update(ctx context.Context, id string, project *models.Project) (*models.Project, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	GetDefaultProjectForOrg(ctx context.Context, orgID string) (*models.Project, *errors.ServiceError)
}
