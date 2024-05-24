package stores

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/pkg/errors"
	"github.com/ashishmax31/soradev-api-server/pkg/models"
)

type UserStore interface {
	Create(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.User, *errors.ServiceError)
	GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
}
