package auth

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/golang-jwt/jwt"
)

const (
	// 30 days.
	RefreshTokenExpiry = 30 * 24 * time.Hour
	// 1 hour.
	JwtTokenExpiry = 1 * time.Hour
)

type RefreshHandlerSpec struct {
	RefreshTokenStore stores.RefreshTokenStore
	UserGetter        UserGetter
	JWTSecret         []byte
	JWTClaimsBuilder  JWTClaimsBuilder
}

type refreshHandler struct {
	refreshStore     stores.RefreshTokenStore
	userGetter       UserGetter
	jwtSecret        []byte
	jwtClaimsBuilder JWTClaimsBuilder
}

func NewRefreshHandler(spec RefreshHandlerSpec) *refreshHandler {
	return &refreshHandler{
		refreshStore:     spec.RefreshTokenStore,
		userGetter:       spec.UserGetter,
		jwtSecret:        spec.JWTSecret,
		jwtClaimsBuilder: spec.JWTClaimsBuilder,
	}
}

func (h *refreshHandler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	var req openapi.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handleError(w, errors.ErrorBadRequest, "invalid request body")
		return
	}

	if req.GetRefreshToken() == "" {
		handleError(w, errors.ErrorBadRequest, "refreshToken is required")
		return
	}

	resp, serr := h.processRefresh(r.Context(), req.GetRefreshToken())
	if serr != nil {
		writeJSONResponse(w, serr.HttpCode, serr.AsOpenapiError())
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

func (h *refreshHandler) processRefresh(ctx context.Context, rawRefreshToken string) (*openapi.RefreshTokenResponse, *errors.ServiceError) {
	storedToken, err := h.validateAndRevokeRefreshToken(ctx, rawRefreshToken)
	if err != nil {
		return nil, err
	}

	user, userErr := h.userGetter.InternalGet(ctx, storedToken.UserID)
	if userErr != nil {
		return nil, errors.Unauthenticated("user not found")
	}

	accessToken, err := h.issueAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, createErr := CreateRefreshToken(ctx, h.refreshStore, user.ID)
	if createErr != nil {
		return nil, errors.GeneralError("failed to generate refresh token")
	}

	resp := openapi.NewRefreshTokenResponse()
	resp.SetToken(accessToken)
	resp.SetRefreshToken(newRefreshToken)
	return resp, nil
}

func (h *refreshHandler) validateAndRevokeRefreshToken(ctx context.Context, rawToken string) (*models.RefreshToken, *errors.ServiceError) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	storedToken, serr := h.refreshStore.GetByTokenHash(ctx, tokenHash)
	if serr != nil {
		return nil, errors.Unauthenticated("invalid refresh token")
	}

	if storedToken.ExpiresAt.Before(time.Now().UTC()) {
		return nil, errors.Unauthenticated("refresh token expired")
	}

	if serr := h.refreshStore.Revoke(ctx, storedToken.ID); serr != nil {
		return nil, errors.GeneralError("failed to process refresh")
	}

	return storedToken, nil
}

func (h *refreshHandler) issueAccessToken(user *models.User) (string, *errors.ServiceError) {
	expirationTime := time.Now().UTC().Add(JwtTokenExpiry)
	claims := h.jwtClaimsBuilder.BuildClaims(user, expirationTime)
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(h.jwtSecret)
	if err != nil {
		return "", errors.GeneralError("failed to generate token")
	}
	return tokenString, nil
}

func CreateRefreshToken(ctx context.Context, store stores.RefreshTokenStore, userID string) (string, error) {
	rawBytes := make([]byte, 32)
	if _, err := cryptorand.Read(rawBytes); err != nil {
		return "", err
	}
	rawToken := hex.EncodeToString(rawBytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	refreshToken := &models.RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().UTC().Add(RefreshTokenExpiry),
		CreatedAt: time.Now().UTC(),
	}

	if _, serr := store.Create(ctx, refreshToken); serr != nil {
		return "", serr.AsError()
	}

	return rawToken, nil
}
