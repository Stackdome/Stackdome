package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
)

type TeamService interface {
	CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError)
	GetTeam(ctx context.Context, orgID, teamID string) (*models.Team, *errors.ServiceError)
	GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError)
	ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError)
	UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError)
	DeleteTeam(ctx context.Context, id string) *errors.ServiceError
	InternalCreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError)
	InternalAddMember(ctx context.Context, teamID, userID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError)

	AddMember(ctx context.Context, teamID, userID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError)
	RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError
	UpdateMemberRole(ctx context.Context, membershipID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError)
	ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListUserTeams(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError)
	InternalListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError)
}

type TeamServiceSpec struct {
	SessionFactory db.SessionFactory
	PolicyManager  resourceaccess.ResourceAccessPolicyManager
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

type teamService struct {
	teamStore          stores.TeamStore
	membershipStore    stores.TeamMembershipStore
	userStore          stores.UserStore
	stackStore         stores.StackStore
	secretStore        stores.SecretStore
	volumeStore        stores.VolumeStore
	postgresAddonStore stores.PostgresAddonStore
	objectStoreStore   stores.ObjectStoreStore
	workspaceUserStore stores.WorkspaceUserStore
	policyMgr          resourceaccess.ResourceAccessPolicyManager
	permissions        auth.PermissionService
	logger             logger.Logger
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
		stackStore: pgstore.NewStackStore(&pgstore.StackStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		secretStore: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		volumeStore: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		postgresAddonStore: pgstore.NewPostgresAddonStore(pgstore.PostgresAddonStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		objectStoreStore: pgstore.NewObjectStoreStore(pgstore.ObjectStoreStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceUserStore: pgstore.NewWorkspaceUserStore(pgstore.WorkspaceUserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		policyMgr:   spec.PolicyManager,
		permissions: spec.Permissions,
		logger:      spec.Logger,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceTeams, "", auth.ActionCreate); permErr != nil {
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

func (s *teamService) InternalCreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError) {
	team := &models.Team{
		Name:           "default",
		OrganisationID: orgID,
		DefaultTeam:    true,
	}
	return s.teamStore.Create(ctx, team)
}

func (s *teamService) GetTeam(ctx context.Context, orgID, id string) (*models.Team, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceTeams, id, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.teamStore.GetByID(ctx, id)
}

func (s *teamService) GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError) {
	res, err := s.teamStore.GetByOrgAndName(ctx, orgID, name)
	if err != nil {
		return nil, err
	}

	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceTeams, res.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return res, nil
}

func (s *teamService) ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceTeams, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.teamStore.ListByOrgID(ctx, orgID)
}

func (s *teamService) UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError) {
	existing, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, existing.OrganisationID, auth.ResourceTeams, existing.ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if team.Name != "" {
		if err := validateTeamName(team.Name); err != nil {
			return nil, err
		}
	}
	if existing.Name != team.Name {
		// check for name conflict
		conflict, serr := s.teamStore.GetByOrgAndName(ctx, existing.OrganisationID, team.Name)
		if serr != nil && serr.Code != errors.ErrorNotFound {
			return nil, serr
		}
		if conflict != nil {
			return nil, errors.Conflict("another team with the same name already exists in this organisation")
		}
	}

	return s.teamStore.Update(ctx, id, team)
}

func (s *teamService) DeleteTeam(ctx context.Context, id string) *errors.ServiceError {
	team, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, team.OrganisationID, auth.ResourceTeams, team.ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	if team.DefaultTeam {
		return errors.BadRequest("cannot delete the default team")
	}

	if depErr := s.checkTeamDependencies(ctx, id); depErr != nil {
		return depErr
	}

	// Pre-fetch memberships before DB delete since CASCADE will remove them
	memberships, err := s.membershipStore.ListByTeamID(ctx, id)
	if err != nil {
		return errors.InternalServerError("failed to list team memberships")
	}

	// Delete team from DB (cascades team_memberships via FK ON DELETE CASCADE)
	if serr := s.teamStore.Delete(ctx, id); serr != nil {
		return serr
	}

	// Clean up Casbin policies using pre-fetched membership data
	for _, membership := range memberships {
		if rmErr := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.TeamID); rmErr != nil {
			s.logger.Errorf("failed to remove team role grouping: %s", rmErr.Error())
		}
	}

	// For each of the users who were members of this team,
	// check if they are still a member of any other team in the same org.
	// If not, remove the OrgMember grouping for that user.
	for _, membership := range memberships {
		if err := s.cleanupOrgMemberGrouping(ctx, membership.UserID, team.OrganisationID); err != nil {
			s.logger.Errorf("failed to cleanup org member grouping for user %s: %s", membership.UserID, err.Error())
		}
	}

	return nil
}

func (s *teamService) checkTeamDependencies(ctx context.Context, teamID string) *errors.ServiceError {
	var blocking []string

	stacks, err := s.stackStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(stacks) > 0 {
		blocking = append(blocking, fmt.Sprintf("stacks (%d)", len(stacks)))
	}

	secrets, err := s.secretStore.ListByTeamID(ctx, teamID, stores.ListParams{})
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(secrets) > 0 {
		blocking = append(blocking, fmt.Sprintf("secrets (%d)", len(secrets)))
	}

	volumes, err := s.volumeStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(volumes) > 0 {
		blocking = append(blocking, fmt.Sprintf("volumes (%d)", len(volumes)))
	}

	addons, err := s.postgresAddonStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(addons) > 0 {
		blocking = append(blocking, fmt.Sprintf("postgres addons (%d)", len(addons)))
	}

	objectStores, err := s.objectStoreStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(objectStores) > 0 {
		blocking = append(blocking, fmt.Sprintf("object stores (%d)", len(objectStores)))
	}

	workspaceUsers, err := s.workspaceUserStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return errors.InternalServerError("failed to check team dependencies: %s", err.Reason)
	}
	if len(workspaceUsers) > 0 {
		blocking = append(blocking, fmt.Sprintf("workspace users (%d)", len(workspaceUsers)))
	}

	if len(blocking) > 0 {
		return errors.BadRequest(
			"cannot delete team with existing resources: %s",
			strings.Join(blocking, ", "),
		)
	}

	return nil
}

func (s *teamService) InternalAddMember(ctx context.Context, teamID, userID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError) {
	exists, serr := s.membershipStore.GetByTeamAndUser(ctx, teamID, userID)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		return nil, serr
	}
	if exists != nil {
		return exists, nil
	}

	team, serr := s.teamStore.GetByID(ctx, teamID)
	if serr != nil {
		return nil, serr
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
		return nil, errors.InternalServerError("failed to add team role grouping")
	}

	if err := s.ensureOrgMemberGrouping(ctx, userID, team.OrganisationID); err != nil {
		s.logger.Errorf("failed to ensure org member grouping: %s", err.Error())
		return nil, err
	}

	return created, nil
}

func (s *teamService) AddMember(ctx context.Context, teamID, userID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError) {
	// Check if membership already exists
	exists, serr := s.membershipStore.GetByTeamAndUser(ctx, teamID, userID)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		return nil, serr
	}
	if exists != nil {
		return nil, errors.Conflict("user is already a member of this team")
	}

	team, serr := s.teamStore.GetByID(ctx, teamID)
	if serr != nil {
		return nil, serr
	}

	user, err := s.userStore.GetByID(ctx, userID)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return nil, errors.BadRequest("user not found")
		}
		return nil, err
	}

	if team.OrganisationID != user.OrganisationID {
		return nil, errors.BadRequest("user and team belong to different organisations")
	}

	if permErr := s.permissions.Check(ctx, team.OrganisationID, auth.ResourceTeams, team.ID, auth.ActionWrite); permErr != nil {
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
		return nil, errors.InternalServerError("failed to add team role grouping")
	}

	if err := s.ensureOrgMemberGrouping(ctx, userID, team.OrganisationID); err != nil {
		s.logger.Errorf("failed to ensure org member grouping: %s", err.Error())
		return nil, err
	}

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
	if permErr := s.permissions.Check(ctx, team.OrganisationID, auth.ResourceTeams, team.ID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if err := s.membershipStore.Delete(ctx, membershipID); err != nil {
		return err
	}

	if rmErr := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.TeamID); rmErr != nil {
		s.logger.Errorf("failed to remove team role grouping: %s", rmErr.Error())
		return errors.InternalServerError("failed to remove team role grouping: %s", rmErr.Error())
	}

	if err := s.cleanupOrgMemberGrouping(ctx, membership.UserID, team.OrganisationID); err != nil {
		s.logger.Errorf("failed to cleanup org member grouping: %s", err.Error())
		return errors.InternalServerError("failed to cleanup org member grouping")
	}

	return nil
}

func (s *teamService) UpdateMemberRole(ctx context.Context, membershipID string, role models.TeamRole) (*models.TeamMembership, *errors.ServiceError) {
	existing, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return nil, serr
	}
	team, serr := s.teamStore.GetByID(ctx, existing.TeamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, team.OrganisationID, auth.ResourceTeams, team.ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("team membership role must be Developer or Viewer")
	}

	res, serr := s.membershipStore.Update(ctx, membershipID, &models.TeamMembership{Role: role})
	if serr != nil {
		return nil, serr
	}
	// Remove old grouping policy and add new one
	err := s.policyMgr.RemoveGroupingPolicy(existing.UserID, string(existing.Role), existing.TeamID)
	if err != nil {
		s.logger.Errorf("failed to remove team role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to remove team role grouping")
	}

	if err := s.policyMgr.AddGroupingPolicy(existing.UserID, string(role), existing.TeamID); err != nil {
		s.logger.Errorf("failed to update team role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to update team role grouping")
	}
	return res, nil
}

func (s *teamService) ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError) {
	team, serr := s.teamStore.GetByID(ctx, teamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, team.OrganisationID, auth.ResourceTeams, team.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.membershipStore.ListByTeamID(ctx, teamID)
}

func (s *teamService) ListUserTeams(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError) {
	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		if serr.Code == errors.ErrorNotFound {
			return nil, errors.Unauthorized("user not found")
		}
		return nil, serr
	}

	if permErr := s.permissions.Check(ctx, user.OrganisationID, auth.ResourceTeams, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.membershipStore.ListByUserIDAndOrgID(ctx, userID, user.OrganisationID)
}

func (s *teamService) InternalListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError) {
	return s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
}

func (s *teamService) ensureOrgMemberGrouping(ctx context.Context, userID, orgID string) *errors.ServiceError {
	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID); err != nil {
		return errors.InternalServerError("failed to add OrgMember grouping")
	}
	return nil
}

func (s *teamService) cleanupOrgMemberGrouping(ctx context.Context, userID, orgID string) *errors.ServiceError {
	memberships, err := s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
	if err != nil {
		return errors.InternalServerError("failed to check remaining memberships")
	}
	if len(memberships) == 0 {
		if err := s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgMemberRole), orgID); err != nil {
			return errors.InternalServerError("failed to remove OrgMember grouping")
		}
	}
	return nil
}

var validTeamName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func validateTeamName(name string) *errors.ServiceError {
	if len(name) == 0 {
		return errors.BadRequest("team name is required")
	}
	if len(name) > 63 {
		return errors.BadRequest("team name must be at most 63 characters")
	}
	if len(name) == 1 {
		if (name[0] < 'a' || name[0] > 'z') && (name[0] < '0' || name[0] > '9') {
			return errors.BadRequest("team name must contain only lowercase alphanumeric characters and hyphens, and must start and end with an alphanumeric character")
		}
		return nil
	}
	if !validTeamName.MatchString(name) {
		return errors.BadRequest("team name must contain only lowercase alphanumeric characters and hyphens, and must start and end with an alphanumeric character")
	}
	return nil
}
