package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type PermissionService interface {
	Check(ctx context.Context, domain, resource, resourceID, action string) *errors.ServiceError
}

type permissionService struct {
	policyMgr resourceaccess.ResourceAccessPolicyManager
	teamStore stores.TeamStore
	logger    logger.Logger
}

type PermissionServiceSpec struct {
	PolicyManager  resourceaccess.ResourceAccessPolicyManager
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewPermissionService(spec PermissionServiceSpec) PermissionService {
	return &permissionService{
		policyMgr: spec.PolicyManager,
		teamStore: pgstore.NewTeamStore(pgstore.TeamStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (p *permissionService) Check(ctx context.Context, domain, resource, resourceID, action string) *errors.ServiceError {
	identity := GetIdentityFromCtx(ctx)
	if identity == nil {
		return errors.Unauthenticated("no identity found in context")
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
			return errors.Forbidden("insufficient permissions: token scope does not cover %s:%s", resource, action)
		}
		if len(identity.ResourceIDs) > 0 && resourceID != "" {
			if !slices.Contains(identity.ResourceIDs, resourceID) {
				return errors.Forbidden("insufficient permissions: token not authorized for resource %s", resourceID)
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
		return errors.Forbidden("insufficient permissions")
	}
	if allowed {
		return nil
	}

	// OrgAdmin fallback: OrgAdmin grouping is on orgID, but team resources use teamID as domain.
	if p.isOrgAdminViaPolicy(ctx, identity.UserID, domain) {
		return nil
	}

	return errors.Forbidden("insufficient permissions")
}

func (p *permissionService) isOrgAdminViaPolicy(ctx context.Context, userID, teamID string) bool {
	if teamID == "" {
		return false
	}
	team, err := p.teamStore.GetByID(ctx, teamID)
	if err != nil {
		p.logger.Errorf("failed to fetch team for OrgAdmin check: %s", err.Error())
		return false
	}
	has, policyErr := p.policyMgr.HasGroupingPolicy(userID, models.OrgAdminRole.String(), team.OrganisationID)
	if policyErr != nil {
		p.logger.Errorf("failed to check OrgAdmin grouping policy: %s", policyErr.Error())
		return false
	}
	return has
}
