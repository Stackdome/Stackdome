package pgstore

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type apiTokenStore struct {
	sessionFactory db.SessionFactory
}

type APITokenStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewAPITokenStore(spec APITokenStoreSpec) stores.APITokenStore {
	return &apiTokenStore{sessionFactory: spec.SessionFactory}
}

func (s *apiTokenStore) Create(ctx context.Context, token *models.APIToken) (*models.APIToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	if err := session.Create(token).Error; err != nil {
		return nil, errors.GeneralError("failed to create api token: %s", err.Error())
	}
	return token, nil
}

func (s *apiTokenStore) GetByID(ctx context.Context, id string) (*models.APIToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	token := &models.APIToken{}
	if err := session.Where("id = ?", id).First(token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("api token not found")
		}
		return nil, errors.GeneralError("failed to get api token: %s", err.Error())
	}
	return token, nil
}

func (s *apiTokenStore) GetByTokenHash(ctx context.Context, hash string) (*models.APIToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	token := &models.APIToken{}
	if err := session.Where("token_hash = ? AND revoked_at IS NULL", hash).First(token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("api token not found")
		}
		return nil, errors.GeneralError("failed to get api token: %s", err.Error())
	}
	return token, nil
}

func (s *apiTokenStore) ListByUserID(ctx context.Context, userID string) ([]*models.APIToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	var tokens []*models.APIToken
	if err := session.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, errors.GeneralError("failed to list api tokens: %s", err.Error())
	}
	return tokens, nil
}

func (s *apiTokenStore) Revoke(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	result := session.Model(&models.APIToken{}).Where("id = ?", id).Update("revoked_at", now)
	if result.Error != nil {
		return errors.GeneralError("failed to revoke api token: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("api token not found")
	}
	return nil
}

func (s *apiTokenStore) UpdateLastUsed(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	if err := session.Model(&models.APIToken{}).Where("id = ?", id).Update("last_used_at", now).Error; err != nil {
		return errors.GeneralError("failed to update last used: %s", err.Error())
	}
	return nil
}
