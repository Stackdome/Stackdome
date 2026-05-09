package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
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
	if err := session.Where("state = ? AND provider = ?", state, provider).First(record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.BadRequest("invalid state parameter")
		}
		return nil, errors.GeneralError("failed to lookup oauth state: %s", err.Error())
	}

	if err := session.Delete(record).Error; err != nil {
		return nil, errors.GeneralError("failed to consume oauth state: %s", err.Error())
	}

	return record, nil
}
