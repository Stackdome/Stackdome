package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type PermissionService interface {
	Check(ctx context.Context, domain, resource, resourceID, action string) error
}

var ErrAccessDenied = fmt.Errorf("access denied")
var ErrUnauthenticated = fmt.Errorf("unauthenticated")

type permissionService struct {
	policyMgr resourceaccess.ResourceAccessPolicyManager
	teamStore stores.TeamStore
	logger    logger.Logger
}

type PermissionServiceConfig struct {
	PolicyManager resourceaccess.ResourceAccessPolicyManager
	TeamStore     stores.TeamStore
	Logger        logger.Logger
}

func NewPermissionService(cfg PermissionServiceConfig) PermissionService {
	return &permissionService{
		policyMgr: cfg.PolicyManager,
		teamStore: cfg.TeamStore,
		logger:    cfg.Logger,
	}
}

func (p *permissionService) Check(ctx context.Context, domain, resource, resourceID, action string) error {
	identity := GetIdentityFromCtx(ctx)
	if identity == nil {
		return ErrUnauthenticated
	}

	if identity.AuthMethod == AuthMethodAPIToken {
		scopeAllowed := false
		for _, scope := range identity.TokenScopes {
			if ScopeCovers(scope, resource, action) {
				scopeAllowed = true
				break
			}
		}
		if !scopeAllowed {
			return ErrAccessDenied
		}
		if len(identity.ResourceIDs) > 0 && resourceID != "" {
			if !slices.Contains(identity.ResourceIDs, resourceID) {
				return ErrAccessDenied
			}
		}
	}

	casbinResource := resource
	if resourceID != "" {
		casbinResource = fmt.Sprintf("%s/%s", resource, resourceID)
	}

	allowed, err := p.policyMgr.CheckPermission(identity.UserID, domain, casbinResource, action)
	if err != nil {
		p.logger.Errorf("permission check failed: %s", err.Error())
		return ErrAccessDenied
	}
	if allowed {
		return nil
	}

	// OrgAdmin fallback: OrgAdmin grouping is on orgID, but team resources use teamID as domain.
	if identity.IsOrgAdmin() && p.teamBelongsToOrg(ctx, domain, identity.OrgID) {
		return nil
	}

	return ErrAccessDenied
}

func (p *permissionService) teamBelongsToOrg(ctx context.Context, teamID, orgID string) bool {
	if teamID == "" || teamID == orgID {
		return false
	}
	if p.teamStore == nil {
		return false
	}
	team, err := p.teamStore.GetByID(ctx, teamID)
	if err != nil {
		return false
	}
	return team.OrganisationID == orgID
}
