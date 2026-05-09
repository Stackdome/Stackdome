package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type OAuthStateStore interface {
	Create(ctx context.Context, state *models.OAuthState) *errors.ServiceError
	Consume(ctx context.Context, state, provider string) (*models.OAuthState, *errors.ServiceError)
}
