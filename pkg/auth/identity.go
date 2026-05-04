package auth

import "context"

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
}

func (i *Identity) IsOrgAdmin() bool {
	return i.Role == "OrgAdmin"
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
