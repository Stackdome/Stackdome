package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/golang-jwt/jwt"
)

type jwtCookieAuthnHandler struct {
	next       http.Handler
	JWTSecret  []byte
	CookieName string
}

const DefaultAuthCookieName = "auth_token"

func NewJwtCookieAuthnHandler(next http.Handler, JWTsecret []byte) http.Handler {
	return &jwtCookieAuthnHandler{
		next:       next,
		JWTSecret:  JWTsecret,
		CookieName: DefaultAuthCookieName,
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
	cookie, err := r.Cookie(h.CookieName)
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

		return h.JWTSecret, nil
	})
	if err != nil {
		handleError(w, errors.ErrorUnauthorized, fmt.Sprintf("token parse error: %s", err.Error()))
		return
	}

	// Check if the token is valid
	if token.Valid {
		ctx := context.WithValue(r.Context(), AuthenticationTokenKey, token)

		// Pass the new context to the next handler
		h.next.ServeHTTP(w, r.WithContext(ctx))
	} else {
		handleError(w, errors.ErrorUnauthorized, "Invalid token")
		return
	}
}
