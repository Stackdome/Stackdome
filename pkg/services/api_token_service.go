package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type APITokenService interface {
	Create(ctx context.Context, name string, scopes []string, resourceIDs []string, expiresAt *time.Time) (*models.APIToken, string, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.APIToken, *errors.ServiceError)
	List(ctx context.Context) ([]*models.APIToken, *errors.ServiceError)
	Revoke(ctx context.Context, id string) *errors.ServiceError
	ValidateToken(ctx context.Context, rawToken string) (*models.APIToken, *errors.ServiceError)
}

type APITokenServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

type apiTokenService struct {
	store  stores.APITokenStore
	logger logger.Logger
}

func NewAPITokenService(spec APITokenServiceSpec) APITokenService {
	return &apiTokenService{
		store: pgstore.NewAPITokenStore(pgstore.APITokenStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *apiTokenService) Create(ctx context.Context, name string, scopes []string, resourceIDs []string, expiresAt *time.Time) (*models.APIToken, string, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, "", errors.Unauthorized("not authenticated")
	}

	for _, scope := range scopes {
		if !auth.ValidateScope(scope) {
			return nil, "", errors.BadRequest("invalid scope: %s", scope)
		}
	}

	rawBytes := make([]byte, auth.ApiTokenByteLen)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, "", errors.GeneralError("failed to generate token")
	}
	rawToken := auth.ApiTokenPrefix + hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token := &models.APIToken{
		Name:        name,
		UserID:      identity.UserID,
		TokenHash:   tokenHash,
		TokenPrefix: rawToken[:8],
		Scopes:      scopes,
		ResourceIDs: resourceIDs,
		OrgID:       identity.OrgID,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now().UTC(),
	}

	created, serr := s.store.Create(ctx, token)
	if serr != nil {
		return nil, "", serr
	}

	return created, rawToken, nil
}

func (s *apiTokenService) GetByID(ctx context.Context, id string) (*models.APIToken, *errors.ServiceError) {
	token, serr := s.store.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}

	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}

	if token.UserID != identity.UserID && !identity.IsOrgAdmin() {
		return nil, errors.Forbidden("cannot view another user's token")
	}

	return token, nil
}

func (s *apiTokenService) List(ctx context.Context) ([]*models.APIToken, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("not authenticated")
	}
	return s.store.ListByUserID(ctx, identity.UserID)
}

func (s *apiTokenService) Revoke(ctx context.Context, id string) *errors.ServiceError {
	token, serr := s.store.GetByID(ctx, id)
	if serr != nil {
		return serr
	}

	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return errors.Unauthorized("not authenticated")
	}

	if token.UserID != identity.UserID && !identity.IsOrgAdmin() {
		return errors.Forbidden("cannot revoke another user's token")
	}

	return s.store.Revoke(ctx, id)
}

func (s *apiTokenService) ValidateToken(ctx context.Context, rawToken string) (*models.APIToken, *errors.ServiceError) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, serr := s.store.GetByTokenHash(ctx, tokenHash)
	if serr != nil {
		return nil, serr
	}

	if token.RevokedAt != nil {
		return nil, errors.Unauthorized("token has been revoked")
	}

	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.Unauthorized("token has expired")
	}

	go func() {
		bgCtx := context.Background()
		if err := s.store.UpdateLastUsed(bgCtx, token.ID); err != nil {
			s.logger.Errorf("failed to update token last used: %s", err.Error())
		}
	}()

	return token, nil
}
