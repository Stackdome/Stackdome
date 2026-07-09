package stores

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=oauth_state_store.go -destination=../mocks/mock_oauth_state_store.go -package=mocks
type OAuthStateStore interface {
	Create(ctx context.Context, state *models.OAuthState) *errors.ServiceError
	Consume(ctx context.Context, state, provider string) (*models.OAuthState, *errors.ServiceError)
	DeleteExpired(ctx context.Context, maxAge time.Duration) error
}
