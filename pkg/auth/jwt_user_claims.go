package auth

import (
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
	jwt.StandardClaims
}
