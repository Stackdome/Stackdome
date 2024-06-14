package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/golang-jwt/jwt"
)

type jwtAuthHandler struct {
	next      http.Handler
	JWTSecret []byte
}

func NewJwtAuthHandler(next http.Handler, JWTsecret []byte) http.Handler {
	return &jwtAuthHandler{
		next:      next,
		JWTSecret: JWTsecret,
	}
}

func (h *jwtAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return h.JWTSecret, nil
	})

	if err != nil {
		handleError(w, errors.ErrorUnauthorized, "Invalid token")
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
