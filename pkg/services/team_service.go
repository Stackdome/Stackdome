package services

import (
	"context"
	"regexp"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

var teamNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

type TeamService interface {
	CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError)
	GetTeam(ctx context.Context, id string) (*models.Team, *errors.ServiceError)
	GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError)
	ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError)
	UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError)
	DeleteTeam(ctx context.Context, id string) *errors.ServiceError
	CreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError)

	AddMember(ctx context.Context, teamID, userID string, role models.Role) (*models.TeamMembership, *errors.ServiceError)
	RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError
	UpdateMemberRole(ctx context.Context, membershipID string, role models.Role) (*models.TeamMembership, *errors.ServiceError)
	ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError)

	PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	DemoteOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError)
}

type TeamServiceSpec struct {
	SessionFactory db.SessionFactory
	PolicyManager  resourceaccess.ResourceAccessPolicyManager
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

type teamService struct {
	teamStore       stores.TeamStore
	membershipStore stores.TeamMembershipStore
	userStore       stores.UserStore
	policyMgr       resourceaccess.ResourceAccessPolicyManager
	permissions     auth.PermissionService
	logger          logger.Logger
}

func NewTeamService(spec TeamServiceSpec) TeamService {
	return &teamService{
		teamStore: pgstore.NewTeamStore(pgstore.TeamStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		membershipStore: pgstore.NewTeamMembershipStore(pgstore.TeamMembershipStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		userStore: pgstore.NewUserStore(pgstore.UserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		policyMgr:   spec.PolicyManager,
		permissions: spec.Permissions,
		logger:      spec.Logger,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if err := validateTeamName(team.Name); err != nil {
		return nil, err
	}

	team.OrganisationID = orgID
	created, serr := s.teamStore.Create(ctx, team)
	if serr != nil {
		return nil, serr
	}
	return created, nil
}

func (s *teamService) CreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError) {
	team := &models.Team{
		Name:           "default",
		OrganisationID: orgID,
		DefaultTeam:    true,
	}
	return s.teamStore.Create(ctx, team)
}

func (s *teamService) GetTeam(ctx context.Context, id string) (*models.Team, *errors.ServiceError) {
	return s.teamStore.GetByID(ctx, id)
}

func (s *teamService) GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError) {
	return s.teamStore.GetByOrgAndName(ctx, orgID, name)
}

func (s *teamService) ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError) {
	return s.teamStore.ListByOrgID(ctx, orgID)
}

func (s *teamService) UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError) {
	existing, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, existing.OrganisationID, auth.ResourceOrgs, existing.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if team.Name != "" {
		if err := validateTeamName(team.Name); err != nil {
			return nil, err
		}
	}
	return s.teamStore.Update(ctx, id, team)
}

func (s *teamService) DeleteTeam(ctx context.Context, id string) *errors.ServiceError {
	team, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return permErr
	}
	if team.DefaultTeam {
		return errors.BadRequest("cannot delete the default team")
	}
	return s.teamStore.Delete(ctx, id)
}

func (s *teamService) AddMember(ctx context.Context, teamID, userID string, role models.Role) (*models.TeamMembership, *errors.ServiceError) {
	team, serr := s.teamStore.GetByID(ctx, teamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("team membership role must be Developer or Viewer")
	}

	membership := &models.TeamMembership{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	created, serr := s.membershipStore.Create(ctx, membership)
	if serr != nil {
		return nil, serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(role), teamID); err != nil {
		s.logger.Errorf("failed to add team role grouping: %s", err.Error())
	}

	s.ensureOrgMemberGrouping(ctx, userID, team.OrganisationID)

	return created, nil
}

func (s *teamService) RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError {
	membership, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return serr
	}
	team, serr := s.teamStore.GetByID(ctx, membership.TeamID)
	if serr != nil {
		return serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if err := s.membershipStore.Delete(ctx, membershipID); err != nil {
		return err
	}

	if rmErr := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.TeamID); rmErr != nil {
		s.logger.Errorf("failed to remove team role grouping: %s", rmErr.Error())
	}

	s.cleanupOrgMemberGrouping(ctx, membership.UserID, team.OrganisationID)

	return nil
}

func (s *teamService) UpdateMemberRole(ctx context.Context, membershipID string, role models.Role) (*models.TeamMembership, *errors.ServiceError) {
	existing, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return nil, serr
	}
	team, serr := s.teamStore.GetByID(ctx, existing.TeamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("team membership role must be Developer or Viewer")
	}

	_ = s.policyMgr.RemoveGroupingPolicy(existing.UserID, string(existing.Role), existing.TeamID)
	if err := s.policyMgr.AddGroupingPolicy(existing.UserID, string(role), existing.TeamID); err != nil {
		s.logger.Errorf("failed to update team role grouping: %s", err.Error())
	}

	return s.membershipStore.Update(ctx, membershipID, &models.TeamMembership{Role: role})
}

func (s *teamService) ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError) {
	return s.membershipStore.ListByTeamID(ctx, teamID)
}

func (s *teamService) ListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError) {
	return s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
}

func (s *teamService) PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}
	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	user.Role = models.OrgAdminRole
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
		s.logger.Errorf("failed to add OrgAdmin grouping: %s", err.Error())
	}

	return nil
}

func (s *teamService) DemoteOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	admins, serr := s.ListOrgAdmins(ctx, orgID)
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
	user.Role = ""
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	_ = s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID)

	return nil
}

func (s *teamService) ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError) {
	return s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
}

func (s *teamService) ensureOrgMemberGrouping(ctx context.Context, userID, orgID string) {
	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID); err != nil {
		s.logger.Errorf("failed to add OrgMember grouping: %s", err.Error())
	}
}

func (s *teamService) cleanupOrgMemberGrouping(ctx context.Context, userID, orgID string) {
	memberships, err := s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
	if err != nil {
		s.logger.Errorf("failed to check remaining memberships: %s", err.Error())
		return
	}
	if len(memberships) == 0 {
		_ = s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgMemberRole), orgID)
	}
}

func validateTeamName(name string) *errors.ServiceError {
	if len(name) == 0 {
		return errors.BadRequest("team name is required")
	}
	if len(name) > 63 {
		return errors.BadRequest("team name must be at most 63 characters")
	}
	if !teamNameRegex.MatchString(name) {
		return errors.BadRequest("team name must be lowercase alphanumeric with hyphens (e.g. 'backend', 'prod-infra')")
	}
	return nil
}
