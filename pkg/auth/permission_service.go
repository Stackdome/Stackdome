package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/stores"
)

//go:generate mockgen -source=permission_service.go -destination=../mocks/mock_permission_service.go -package=mocks
type PermissionService interface {
	Check(ctx context.Context, domain, resource, resourceID, action string) *errors.ServiceError
}

type permissionService struct {
	policyMgr resourceaccess.ResourceAccessPolicyManager
	projectStore stores.ProjectStore
	logger    logger.Logger
}

type PermissionServiceSpec struct {
	PolicyManager resourceaccess.ResourceAccessPolicyManager
	ProjectStore     stores.ProjectStore
	Logger        logger.Logger
}

func NewPermissionService(spec PermissionServiceSpec) PermissionService {
	return &permissionService{
		policyMgr: spec.PolicyManager,
		projectStore: spec.ProjectStore,
		logger:    spec.Logger,
	}
}

func (p *permissionService) Check(ctx context.Context, domain, resource, resourceID, action string) *errors.ServiceError {
	identity := GetIdentityFromCtx(ctx)
	if identity == nil {
		return errors.Unauthenticated("no identity found in context")
	}
	if identity.IsSystem {
		return nil
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

	// OrgAdmin fallback: OrgAdmin grouping is on orgID, but project resources use projectID as domain.
	if p.isOrgAdminViaPolicy(ctx, identity.UserID, domain) {
		return nil
	}

	return errors.Forbidden("insufficient permissions")
}

func (p *permissionService) isOrgAdminViaPolicy(ctx context.Context, userID, projectID string) bool {
	if projectID == "" {
		return false
	}
	project, err := p.projectStore.GetByID(ctx, projectID)
	if err != nil {
		p.logger.Errorf("failed to fetch project for OrgAdmin check: %s", err.Error())
		return false
	}
	has, policyErr := p.policyMgr.HasGroupingPolicy(userID, models.OrgAdminRole.String(), project.OrganisationID)
	if policyErr != nil {
		p.logger.Errorf("failed to check OrgAdmin grouping policy: %s", policyErr.Error())
		return false
	}
	return has
}
