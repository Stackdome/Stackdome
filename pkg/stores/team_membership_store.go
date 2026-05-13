package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type TeamMembershipStore interface {
	Create(ctx context.Context, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.TeamMembership, *errors.ServiceError)
	GetByTeamAndUser(ctx context.Context, teamID, userID string) (*models.TeamMembership, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError)
	Update(ctx context.Context, id string, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
}
