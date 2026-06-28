package auth

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type IdentityKey string

const identityContextKey IdentityKey = "AuthIdentity"

type AuthMethod string

const (
	AuthMethodJWT      AuthMethod = "jwt"
	AuthMethodAPIToken AuthMethod = "api_token"
	AuthMethodOAuth    AuthMethod = "oauth"
)

type Identity struct {
	UserID      string
	OrgID       string
	Role        string
	AuthMethod  AuthMethod
	TokenID     string
	TokenScopes []string
	ResourceIDs []string
	IsSystem    bool
}

func (i *Identity) IsOrgAdmin() bool {
	return i.Role == models.OrgAdminRole.String()
}

func SetIdentityInContext(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

func GetIdentityFromCtx(ctx context.Context) *Identity {
	identity, ok := ctx.Value(identityContextKey).(*Identity)
	if !ok {
		return nil
	}
	return identity
}
