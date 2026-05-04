package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.User, *errors.ServiceError)
	GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	Update(ctx context.Context, id string, user *models.User) (*models.User, *errors.ServiceError)
	ListByOrgAndRole(ctx context.Context, orgID string, role models.Role) ([]*models.User, *errors.ServiceError)
}
