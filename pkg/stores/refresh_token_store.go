package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type RefreshTokenStore interface {
	Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, *errors.ServiceError)
	GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, *errors.ServiceError)
	Revoke(ctx context.Context, id string) *errors.ServiceError
	RevokeAllForUser(ctx context.Context, userID string) *errors.ServiceError
}
