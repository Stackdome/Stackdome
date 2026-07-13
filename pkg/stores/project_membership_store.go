package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type ProjectMembershipStore interface {
	Create(ctx context.Context, membership *models.ProjectMembership) (*models.ProjectMembership, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.ProjectMembership, *errors.ServiceError)
	GetByProjectAndUser(ctx context.Context, projectID, userID string) (*models.ProjectMembership, *errors.ServiceError)
	ListByProjectID(ctx context.Context, projectID string) ([]*models.ProjectMembership, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.ProjectMembership, *errors.ServiceError)
	ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.ProjectMembership, *errors.ServiceError)
	Update(ctx context.Context, id string, membership *models.ProjectMembership) (*models.ProjectMembership, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
}
