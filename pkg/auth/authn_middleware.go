package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=authn_middleware.go -destination=authn_middleware_mock.go -package=auth

type authnMiddleware struct {
	authenticators []authenticator
}

type AuthnMiddleware interface {
	AuthenticateUser(next http.Handler) http.Handler
	GetAvailableAuthenticators() []authenticator
}

type UserGetter interface {
	InternalGet(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
}

type authenticator interface {
	// Returns true if the current authenticator can be run.
	Select(requestCtx context.Context) bool

	// Runs the current authenticator's auth handler.
	AuthenticaticationHandler(w http.ResponseWriter, r *http.Request, next http.Handler)
}

var _ AuthnMiddleware = &authnMiddleware{}

func NewAuthMiddleware(userGetter UserGetter) AuthnMiddleware {
	middleware := &authnMiddleware{
		authenticators: []authenticator{
			&jwtAuthenticator{userGetter: userGetter},
		},
	}
	return middleware
}

func (a *authnMiddleware) GetAvailableAuthenticators() []authenticator {
	return a.authenticators
}

// Middleware handler to validate JWT tokens and authenticate users
func (a *authnMiddleware) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCtx := r.Context()
		for _, auth := range a.GetAvailableAuthenticators() {
			ok := auth.Select(requestCtx)
			if ok {
				auth.AuthenticaticationHandler(w, r, next)
				return
			}
		}
		// If no authenticator matches.
		handleError(w, errors.ErrorUnauthorized, "No available authenticator for the passed token")
	})
}

type jwtAuthenticator struct {
	userGetter UserGetter
}

func (a *jwtAuthenticator) AuthenticaticationHandler(w http.ResponseWriter, r *http.Request, next http.Handler) {
	ctx := r.Context()
	payload, err := GetJwtAuthPayloadFromContext(ctx)
	if err != nil {
		handleError(w, errors.ErrorUnauthorized, fmt.Sprintf("Unable to get payload details from JWT token: %s", err))
		return
	}
	user, serr := a.userGetter.InternalGet(ctx, payload.UserID)
	if serr != nil {
		handleError(w, errors.ErrorUnauthorized, "Failed to fetch user")
		return
	}
	// Append the username to the request context
	ctx = SetUserInContext(ctx, user)
	ctx = SetIdentityInContext(ctx, &Identity{
		UserID:     user.ID,
		OrgID:      user.OrganisationID,
		Role:       string(user.Role),
		AuthMethod: AuthMethodJWT,
	})

	// Stamp identity onto the context and the per-request logger so every
	// downstream log line is attributable to a user/org.
	ctx = logger.WithUserID(ctx, user.ID)
	ctx = logger.WithOrgID(ctx, user.OrganisationID)
	reqLogger := logger.GetLoggerFromContext(ctx).WithFields(map[string]interface{}{
		logger.FieldUserID: user.ID,
		logger.FieldOrgID:  user.OrganisationID,
	})
	ctx = logger.AddLoggerToContext(ctx, reqLogger)
	*r = *r.WithContext(ctx)

	next.ServeHTTP(w, r)
}

func (a *jwtAuthenticator) Select(requestCtx context.Context) bool {
	return IsJwtTokenInCtx(requestCtx)
}
