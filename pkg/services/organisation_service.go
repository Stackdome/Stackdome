package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type OrganisationService interface {
	InternalCreate(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)

	PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	DemoteOrgAdmin(ctx context.Context, orgID, userID, teamID string, role models.TeamRole) *errors.ServiceError
	ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError)
}

type organisationService struct {
	organisationStore         stores.OrganisationStore
	organisationDomainService OrganisationDomainsService
	stackQueryService         StackQueryService
	userStore                 stores.UserStore
	teamService               TeamService
	atomicExecutor            stores.AtomicExecutor
	policyMgr                 resourceaccess.ResourceAccessPolicyManager
	permissions               auth.PermissionService
	logger                    logger.Logger
}

func NewOrganisationService(spec OrganisationServiceSpec) OrganisationService {
	return &organisationService{
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		userStore: pgstore.NewUserStore(pgstore.UserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackQueryService:         spec.StackQueryService,
		organisationDomainService: spec.OrganisationDomainService,
		teamService:               spec.TeamService,
		atomicExecutor:            pgstore.NewAtomicExecutor(spec.SessionFactory),
		policyMgr:                 spec.PolicyManager,
		permissions:               spec.Permissions,
		logger:                    spec.Logger,
	}
}

type OrganisationServiceSpec struct {
	SessionFactory            db.SessionFactory
	OrganisationDomainService OrganisationDomainsService
	StackQueryService         StackQueryService
	TeamService               TeamService
	PolicyManager             resourceaccess.ResourceAccessPolicyManager
	Permissions               auth.PermissionService
	Logger                    logger.Logger
}

func (s *organisationService) InternalCreate(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	if len(spec.Name) == 0 {
		return nil, errors.BadRequest("organisation name is required")
	}

	org, err := s.organisationStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create organisation: %v", err)
		return nil, err
	}

	for _, domain := range spec.Domains {
		domain.OrganisationID = org.ID
		if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
			return nil, err
		}
	}

	return s.organisationStore.Get(ctx, org.ID)
}

func (s *organisationService) Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	org, err := s.organisationStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return nil, err
	}
	return org, nil
}

// TODO: Org is the root of almost everything, so we need to be careful when deleting it.
func (s *organisationService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	stacks, err := s.stackQueryService.GetStacksByOrganisationID(ctx, ID)
	if err != nil {
		return err
	}
	if len(stacks) > 0 {
		return errors.BadRequest("cannot delete organisation with stacks")
	}
	err = s.organisationStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete organisation: %v", err)
		return err
	}
	return nil
}

// Update updates an organisation
func (s *organisationService) Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, ID, auth.ResourceOrgs, ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	existing, err := s.Get(ctx, ID)
	if err != nil {
		return nil, err
	}

	if spec.Name != "" && existing.Name != spec.Name {
		nameExists, err := s.organisationStore.OrganisationNameExists(ctx, spec.Name)
		if err != nil {
			s.logger.Errorf("failed to check if organisation name exists: %v", err)
			return nil, errors.GeneralError("failed to update organisation")
		}
		if nameExists {
			return nil, errors.Conflict("organisation with the same name already exists")
		}
	}

	if len(spec.Name) == 0 {
		spec.Name = existing.Name
	}
	org, err := s.organisationStore.Update(ctx, ID, spec)
	if err != nil {
		s.logger.Errorf("failed to update organisation: %v", err)
		return nil, err
	}

	existingDomains, err := s.organisationDomainService.ListByOrganisationID(ctx, ID)
	if err != nil {
		return nil, err
	}

	existingDomainsMap := make(map[string]struct{})
	for _, domain := range existingDomains {
		existingDomainsMap[domain.Domain] = struct{}{}
	}

	for _, domain := range spec.Domains {
		if _, exists := existingDomainsMap[domain.Domain]; !exists {
			domain.OrganisationID = org.ID
			if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
				return nil, err
			}
		}
	}

	currentDomainsMap := make(map[string]struct{})
	for _, domain := range spec.Domains {
		currentDomainsMap[domain.Domain] = struct{}{}
	}
	for _, domain := range existingDomains {
		if _, exists := currentDomainsMap[domain.Domain]; !exists {
			if err := s.organisationDomainService.Delete(ctx, domain.ID); err != nil {
				return nil, err
			}
		}
	}
	return s.Get(ctx, org.ID)
}

func (s *organisationService) PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}

	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	if user.IsOrgAdmin() {
		return errors.BadRequest("user is already an OrgAdmin")
	}

	user.Role = models.OrgAdminRole
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
		s.logger.Errorf("failed to add OrgAdmin grouping: %s", err.Error())
		return errors.InternalServerError("failed to add OrgAdmin grouping")
	}

	return nil
}

func (s *organisationService) DemoteOrgAdmin(ctx context.Context, orgID, userID, teamID string, role models.TeamRole) *errors.ServiceError {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	admins, serr := s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
	if serr != nil {
		return serr
	}
	if len(admins) <= 1 {
		return errors.BadRequest("cannot demote the last OrgAdmin")
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}

	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	if !user.IsOrgAdmin() {
		return errors.BadRequest("user is not an OrgAdmin")
	}

	return s.atomicExecutor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		user.Role = models.NoRole
		if _, serr := s.userStore.Update(txCtx, userID, user); serr != nil {
			return serr
		}

		if _, serr := s.teamService.InternalAddMember(txCtx, teamID, userID, role); serr != nil {
			s.logger.Errorf("failed to add demoted user to team: %s", serr.Error())
			return serr
		}

		if err := s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
			s.logger.Errorf("failed to remove OrgAdmin grouping: %s", err.Error())
			return errors.InternalServerError("failed to remove OrgAdmin grouping")
		}

		return nil
	})
}

func (s *organisationService) ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
}
