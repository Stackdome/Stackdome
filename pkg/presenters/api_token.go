package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

func PresentAPIToken(in *models.APIToken) openapi.APIToken {
	res := openapi.APIToken{}
	res.SetId(in.ID)
	res.SetName(in.Name)
	res.SetUserId(in.UserID)
	res.SetTokenPrefix(in.TokenPrefix)
	res.SetScopes([]string(in.Scopes))
	res.SetOrgId(in.OrgID)
	res.SetCreatedAt(in.CreatedAt)
	if len(in.ResourceIDs) > 0 {
		res.SetResourceIds([]string(in.ResourceIDs))
	}
	if in.ExpiresAt != nil {
		res.SetExpiresAt(*in.ExpiresAt)
	}
	if in.LastUsedAt != nil {
		res.SetLastUsedAt(*in.LastUsedAt)
	}
	if in.RevokedAt != nil {
		res.SetRevokedAt(*in.RevokedAt)
	}
	return res
}

func PresentAPITokenList(in []*models.APIToken) []openapi.APIToken {
	res := make([]openapi.APIToken, len(in))
	for i, t := range in {
		res[i] = PresentAPIToken(t)
	}
	return res
}

func PresentAPITokenCreateResponse(token *models.APIToken, rawToken string) openapi.APITokenCreateResponse {
	res := openapi.APITokenCreateResponse{}
	res.SetToken(rawToken)
	res.SetId(token.ID)
	res.SetName(token.Name)
	res.SetTokenPrefix(token.TokenPrefix)
	if token.ExpiresAt != nil {
		res.SetExpiresAt(*token.ExpiresAt)
	}
	return res
}
