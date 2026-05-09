package auth

import (
	"fmt"
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/golang-jwt/jwt"
)

type jwtCookieAuthnHandler struct {
	next       http.Handler
	jwtSecret  []byte
	userGetter UserGetter
	cookieName string
}

type JWTCookieAuthnHandlerSpec struct {
	JWTSecret  []byte
	UserGetter UserGetter
}

const DefaultAuthCookieName = "auth_token"

func NewJwtCookieAuthnHandler(next http.Handler, spec JWTCookieAuthnHandlerSpec) http.Handler {
	return &jwtCookieAuthnHandler{
		next:       next,
		jwtSecret:  spec.JWTSecret,
		userGetter: spec.UserGetter,
		cookieName: DefaultAuthCookieName,
	}
}

func CanAuthenticateWithCookie(r *http.Request) bool {
	// Check if the request has the authentication cookie
	cookie, err := r.Cookie(DefaultAuthCookieName)
	if err != nil {
		return false // No cookie means no authentication
	}

	// If the cookie exists, we assume authentication is required
	if cookie != nil && cookie.Value != "" {
		return true
	}

	return false
}

func (h *jwtCookieAuthnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the JWT token from cookie
	cookie, err := r.Cookie(h.cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			handleError(w, errors.ErrorUnauthorized, "Authentication cookie missing")
			return
		}
		handleError(w, errors.ErrorUnauthorized, "Error reading authentication cookie")
		return
	}

	tokenString := cookie.Value
	if tokenString == "" {
		handleError(w, errors.ErrorUnauthorized, "Empty authentication token")
		return
	}

	// Parse and decode the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return h.jwtSecret, nil
	})
	if err != nil {
		handleError(w, errors.ErrorUnauthorized, fmt.Sprintf("token parse error: %s", err.Error()))
		return
	}

	// Check if the token is valid
	if token.Valid {
		userID, err := getUserIdFromToken(token)
		if err != nil {
			handleError(w, errors.ErrorUnauthorized, fmt.Sprintf("Unable to get user ID from JWT token: %s", err))
			return
		}
		user, serr := h.userGetter.InternalGet(r.Context(), userID)
		if serr != nil {
			handleError(w, errors.ErrorUnauthorized, "Failed to fetch user")
			return
		}
		ctx := SetUserInContext(r.Context(), user)
		ctx = SetIdentityInContext(ctx, &Identity{
			UserID:     user.ID,
			OrgID:      user.OrganisationID,
			Role:       string(user.Role),
			AuthMethod: AuthMethodJWT,
		})
		*r = *r.WithContext(ctx)
		h.next.ServeHTTP(w, r)
	} else {
		handleError(w, errors.ErrorUnauthorized, "Invalid token")
		return
	}
}
