package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=team_store.go -destination=../mocks/mock_team_store.go -package=mocks
type TeamStore interface {
	Create(ctx context.Context, team *models.Team) (*models.Team, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Team, *errors.ServiceError)
	GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError)
	Update(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	GetDefaultTeamForOrg(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError)
}
