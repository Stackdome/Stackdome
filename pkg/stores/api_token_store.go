package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type APITokenStore interface {
	Create(ctx context.Context, token *models.APIToken) (*models.APIToken, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.APIToken, *errors.ServiceError)
	GetByTokenHash(ctx context.Context, hash string) (*models.APIToken, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.APIToken, *errors.ServiceError)
	Revoke(ctx context.Context, id string) *errors.ServiceError
	UpdateLastUsed(ctx context.Context, id string) *errors.ServiceError
}
