package auth

import (
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.StandardClaims
}

type JWTClaimsBuilder interface {
	BuildClaims(user *models.User, expirationTime time.Time) jwt.Claims
}

type JWTClaimsBuilderImpl struct{}

func (c *JWTClaimsBuilderImpl) BuildClaims(user *models.User, expirationTime time.Time) jwt.Claims {
	return &Claims{
		UserID: user.ID,
		Role:   user.Role.String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
}

// NewJWTClaimsBuilder returns a new instance of JWTClaimsBuilder.
func NewJWTClaimsBuilder() *JWTClaimsBuilderImpl {
	return &JWTClaimsBuilderImpl{}
}
