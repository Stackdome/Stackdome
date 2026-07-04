package pgstore

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm/clause"
)

type oauthStateStore struct {
	sessionFactory db.SessionFactory
}

type OAuthStateStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewOAuthStateStore(spec OAuthStateStoreSpec) stores.OAuthStateStore {
	return &oauthStateStore{sessionFactory: spec.SessionFactory}
}

func (s *oauthStateStore) Create(ctx context.Context, state *models.OAuthState) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	if err := session.Create(state).Error; err != nil {
		return errors.GeneralError("failed to store oauth state: %s", err.Error())
	}
	return nil
}

func (s *oauthStateStore) Consume(ctx context.Context, state, provider string) (*models.OAuthState, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	record := &models.OAuthState{}
	result := session.Clauses(clause.Returning{}).
		Where("state = ? AND provider = ?", state, provider).
		Delete(record)
	if result.Error != nil {
		return nil, errors.GeneralError("failed to consume oauth state: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, errors.BadRequest("invalid state parameter")
	}
	return record, nil
}

func (s *oauthStateStore) DeleteExpired(ctx context.Context, maxAge time.Duration) error {
	session := s.sessionFactory.New(ctx)
	return session.Where("created_at < ?", time.Now().UTC().Add(-maxAge)).Delete(&models.OAuthState{}).Error
}
