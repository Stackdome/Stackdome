package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/golang-jwt/jwt"
)

type jwtAuthPayload struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

type jwtAuthnHandler struct {
	next       http.Handler
	jwtSecret  []byte
	userGetter UserGetter
}

type JWTAuthnHandlerSpec struct {
	JWTSecret  []byte
	UserGetter UserGetter
}

func NewJwtAuthnHandler(next http.Handler, spec JWTAuthnHandlerSpec) http.Handler {
	return &jwtAuthnHandler{
		next:       next,
		jwtSecret:  spec.JWTSecret,
		userGetter: spec.UserGetter,
	}
}

func (h *jwtAuthnHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		handleError(w, errors.ErrorUnauthorized, "Authorization header missing")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		handleError(w, errors.ErrorUnauthorized, "Invalid Authorization header")
		return
	}

	tokenString := parts[1]

	// Parse and decode the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.jwtSecret, nil
	})

	if err != nil {
		handleError(w, errors.ErrorUnauthorized, "Invalid token")
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

func getUserIdFromToken(jwtToken *jwt.Token) (string, error) {
	claims, ok := jwtToken.Claims.(*Claims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	if claims.UserID == "" {
		return "", fmt.Errorf("user ID missing in token claims")
	}
	return claims.UserID, nil
}
