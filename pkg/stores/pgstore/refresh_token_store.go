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

type refreshTokenStore struct {
	sessionFactory db.SessionFactory
}

type RefreshTokenStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewRefreshTokenStore(spec RefreshTokenStoreSpec) stores.RefreshTokenStore {
	return &refreshTokenStore{sessionFactory: spec.SessionFactory}
}

func (s *refreshTokenStore) Create(ctx context.Context, token *models.RefreshToken) (*models.RefreshToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	if err := session.Create(token).Error; err != nil {
		return nil, errors.GeneralError("failed to create refresh token: %s", err.Error())
	}
	return token, nil
}

func (s *refreshTokenStore) GetByTokenHash(ctx context.Context, hash string) (*models.RefreshToken, *errors.ServiceError) {
	session := s.sessionFactory.New(ctx)
	token := &models.RefreshToken{}
	if err := session.Where("token_hash = ? AND revoked_at IS NULL", hash).First(token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("refresh token not found")
		}
		return nil, errors.GeneralError("failed to get refresh token: %s", err.Error())
	}
	return token, nil
}

func (s *refreshTokenStore) Revoke(ctx context.Context, id string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	result := session.Model(&models.RefreshToken{}).Where("id = ?", id).Update("revoked_at", now)
	if result.Error != nil {
		return errors.GeneralError("failed to revoke refresh token: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("refresh token not found")
	}
	return nil
}

func (s *refreshTokenStore) RevokeAllForUser(ctx context.Context, userID string) *errors.ServiceError {
	session := s.sessionFactory.New(ctx)
	now := time.Now().UTC()
	if err := session.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error; err != nil {
		return errors.GeneralError("failed to revoke refresh tokens: %s", err.Error())
	}
	return nil
}
