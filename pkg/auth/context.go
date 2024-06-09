package auth

import (
	"context"
	"fmt"

	"github.com/ashishmax31/soradev-api-server/pkg/models"
	"github.com/golang-jwt/jwt"
)

type AuthenticationToken string
type ContextUser string

const (
	AuthenticationTokenKey AuthenticationToken = "AuthToken"
	ContextUserKey         ContextUser         = "UserInContext"
)

type JwtAuthPayload struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

func GetJwtAuthPayloadFromContext(ctx context.Context) (*JwtAuthPayload, error) {
	token, ok := ctx.Value(AuthenticationTokenKey).(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("token missing from context")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	authPayload := &JwtAuthPayload{
		UserID: claims.UserID,
		Role:   claims.Role,
	}

	return authPayload, nil
}

func IsJwtTokenInCtx(ctx context.Context) bool {
	switch ctx.Value(AuthenticationTokenKey).(type) {
	case nil:
		return false
	case *jwt.Token:
		return true
	default:
		return false
	}
}

func SetUserInContext(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, ContextUserKey, user)
}

func GetCurrentUserFromCtx(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value(ContextUserKey).(*models.User)
	if !ok {
		return nil, fmt.Errorf("missing user in context")
	}
	return user, nil
}
