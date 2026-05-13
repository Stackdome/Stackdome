package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

const (
	ApiTokenPrefix  = "sdm_"
	ApiTokenByteLen = 32
)

// TokenLookup validates a raw token and returns the stored token record.
type TokenLookup interface {
	ValidateToken(ctx context.Context, rawToken string) (*models.APIToken, *errors.ServiceError)
}

type apiTokenAuthnHandler struct {
	tokenLookup TokenLookup
	userGetter  UserGetter
	next        http.Handler
}

type ApiTokenAuthnHandlerSpec struct {
	TokenLookup TokenLookup
	UserGetter  UserGetter
}

func NewAPITokenHandler(next http.Handler, spec ApiTokenAuthnHandlerSpec) http.Handler {
	return &apiTokenAuthnHandler{
		next:        next,
		tokenLookup: spec.TokenLookup,
		userGetter:  spec.UserGetter,
	}
}

func (v *apiTokenAuthnHandler) ValidateAndResolve(r *http.Request, rawToken string) (*Identity, error) {
	token, err := v.tokenLookup.ValidateToken(r.Context(), rawToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	user, serr := v.userGetter.InternalGet(r.Context(), token.UserID)
	if serr != nil {
		if serr.Code == errors.ErrorNotFound {
			return nil, fmt.Errorf("token user not found: %w", serr)
		}
		return nil, fmt.Errorf("failed to resolve token owner: %s", serr.Error())
	}

	// Set the legacy user context so existing handlers still work
	ctx := SetUserInContext(r.Context(), user)
	*r = *r.WithContext(ctx)

	return &Identity{
		UserID:      token.UserID,
		OrgID:       token.OrgID,
		Role:        string(user.Role),
		AuthMethod:  AuthMethodAPIToken,
		TokenID:     token.ID,
		TokenScopes: token.Scopes,
		ResourceIDs: token.ResourceIDs,
	}, nil
}

func CanAuthenticateWithAPIToken(r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}
	return strings.HasPrefix(parts[1], ApiTokenPrefix)
}

func (h *apiTokenAuthnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		handleError(w, errors.ErrorUnauthenticated, "invalid authorization header")
		return
	}
	rawToken := parts[1]

	identity, err := h.ValidateAndResolve(r, rawToken)
	if err != nil {
		handleError(w, errors.ErrorUnauthenticated, "invalid api token")
		return
	}

	ctx := r.Context()
	ctx = SetIdentityInContext(ctx, identity)
	*r = *r.WithContext(ctx)

	h.next.ServeHTTP(w, r)
}
