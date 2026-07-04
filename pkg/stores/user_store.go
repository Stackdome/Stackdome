package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.User, *errors.ServiceError)
	GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	Update(ctx context.Context, id string, user *models.User) (*models.User, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string, params ListParams) (*PaginatedResult[*models.User], *errors.ServiceError)
	ListByOrgAndRole(ctx context.Context, orgID string, role models.UserRole) ([]*models.User, *errors.ServiceError)
}
