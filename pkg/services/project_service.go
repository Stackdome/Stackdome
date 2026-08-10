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

//go:generate mockgen -destination=../mocks/mock_project_service.go -package=mocks github.com/Stackdome/stackdome/pkg/services ProjectService

type ProjectService interface {
	CreateProject(ctx context.Context, orgID string, project *models.Project) (*models.Project, *errors.ServiceError)
	GetProject(ctx context.Context, orgID, projectID string) (*models.Project, *errors.ServiceError)
	GetProjectByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError)
	ListProjects(ctx context.Context, orgID string) ([]*models.Project, *errors.ServiceError)
	UpdateProject(ctx context.Context, id string, project *models.Project) (*models.Project, *errors.ServiceError)
	DeleteProject(ctx context.Context, id string) *errors.ServiceError
	InternalCreateDefaultProject(ctx context.Context, orgID string) (*models.Project, *errors.ServiceError)
	InternalGetProjectByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError)
	InternalAddMember(ctx context.Context, projectID, userID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError)

	AddMember(ctx context.Context, projectID, userID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError)
	RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError
	UpdateMemberRole(ctx context.Context, membershipID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError)
	ListMembers(ctx context.Context, projectID string) ([]*models.ProjectMembership, *errors.ServiceError)
	ListUserProjects(ctx context.Context, userID string) ([]*models.ProjectMembership, *errors.ServiceError)
	InternalListUserProjects(ctx context.Context, userID, orgID string) ([]*models.ProjectMembership, *errors.ServiceError)
}

type ProjectServiceSpec struct {
	SessionFactory db.SessionFactory
	PolicyManager  resourceaccess.ResourceAccessPolicyManager
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

type projectService struct {
	projectStore       stores.ProjectStore
	membershipStore    stores.ProjectMembershipStore
	userStore          stores.UserStore
	stackStore         stores.StackStore
	secretStore        stores.SecretStore
	volumeStore        stores.VolumeStore
	postgresAddonStore stores.PostgresAddonStore
	objectStoreStore   stores.ObjectStoreStore
	policyMgr          resourceaccess.ResourceAccessPolicyManager
	permissions        auth.PermissionService
	logger             logger.Logger
}

func NewProjectService(spec ProjectServiceSpec) ProjectService {
	return &projectService{
		projectStore: pgstore.NewProjectStore(pgstore.ProjectStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		membershipStore: pgstore.NewProjectMembershipStore(pgstore.ProjectMembershipStoreSpec{
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
		policyMgr:   spec.PolicyManager,
		permissions: spec.Permissions,
		logger:      spec.Logger,
	}
}

func (s *projectService) CreateProject(ctx context.Context, orgID string, project *models.Project) (*models.Project, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceProjects, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

	if err := validateProjectName(project.Name); err != nil {
		return nil, err
	}

	// Alpha ships one project per organisation: the default one made at signup.
	existing, serr := s.projectStore.ListByOrgID(ctx, orgID)
	if serr != nil {
		return nil, serr
	}
	if len(existing) > 0 {
		return nil, errors.Conflict("alpha organisations use the %q project; additional projects are not available yet", models.DefaultProjectName)
	}

	project.OrganisationID = orgID
	created, serr := s.projectStore.Create(ctx, project)
	if serr != nil {
		return nil, serr
	}
	return created, nil
}

func (s *projectService) InternalCreateDefaultProject(ctx context.Context, orgID string) (*models.Project, *errors.ServiceError) {
	project := &models.Project{
		Name:           models.DefaultProjectName,
		OrganisationID: orgID,
		DefaultProject: true,
	}
	return s.projectStore.Create(ctx, project)
}

func (s *projectService) GetProject(ctx context.Context, orgID, id string) (*models.Project, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceProjects, id, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.projectStore.GetByID(ctx, id)
}

func (s *projectService) GetProjectByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError) {
	res, err := s.projectStore.GetByOrgAndName(ctx, orgID, name)
	if err != nil {
		return nil, err
	}

	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceProjects, res.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return res, nil
}

// Name -> ID lookup with no permission check. Callers use it to resolve a URL
// path segment before authorizing the resource actually being requested; an
// API token scoped to that resource must not need projects:read to get there.
func (s *projectService) InternalGetProjectByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError) {
	return s.projectStore.GetByOrgAndName(ctx, orgID, name)
}

func (s *projectService) ListProjects(ctx context.Context, orgID string) ([]*models.Project, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, orgID, auth.ResourceProjects, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.projectStore.ListByOrgID(ctx, orgID)
}

func (s *projectService) UpdateProject(ctx context.Context, id string, project *models.Project) (*models.Project, *errors.ServiceError) {
	existing, serr := s.projectStore.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, existing.OrganisationID, auth.ResourceProjects, existing.ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if project.Name != "" {
		if err := validateProjectName(project.Name); err != nil {
			return nil, err
		}
	}
	if existing.Name != project.Name {
		// check for name conflict
		conflict, serr := s.projectStore.GetByOrgAndName(ctx, existing.OrganisationID, project.Name)
		if serr != nil && serr.Code != errors.ErrorNotFound {
			return nil, serr
		}
		if conflict != nil {
			return nil, errors.Conflict("another project with the same name already exists in this organisation")
		}
	}

	return s.projectStore.Update(ctx, id, project)
}

func (s *projectService) DeleteProject(ctx context.Context, id string) *errors.ServiceError {
	project, serr := s.projectStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, project.OrganisationID, auth.ResourceProjects, project.ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	if project.DefaultProject {
		return errors.BadRequest("cannot delete the default project")
	}

	if depErr := s.checkProjectDependencies(ctx, id); depErr != nil {
		return depErr
	}

	// Pre-fetch memberships before DB delete since CASCADE will remove them
	memberships, err := s.membershipStore.ListByProjectID(ctx, id)
	if err != nil {
		return errors.InternalServerError("failed to list project memberships")
	}

	// Delete project from DB (cascades project_memberships via FK ON DELETE CASCADE)
	if serr := s.projectStore.Delete(ctx, id); serr != nil {
		return serr
	}

	// Clean up Casbin policies using pre-fetched membership data
	for _, membership := range memberships {
		if rmErr := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.ProjectID); rmErr != nil {
			s.logger.Error(ctx, "failed to remove project role grouping: %s", rmErr.Error())
		}
	}

	// For each of the users who were members of this project,
	// check if they are still a member of any other project in the same org.
	// If not, remove the OrgMember grouping for that user.
	for _, membership := range memberships {
		if err := s.cleanupOrgMemberGrouping(ctx, membership.UserID, project.OrganisationID); err != nil {
			s.logger.Error(ctx, "failed to cleanup org member grouping for user %s: %s", membership.UserID, err.Error())
		}
	}

	return nil
}

func (s *projectService) checkProjectDependencies(ctx context.Context, projectID string) *errors.ServiceError {
	var blocking []string

	stacks, err := s.stackStore.ListByProjectID(ctx, projectID)
	if err != nil {
		return errors.InternalServerError("failed to check project dependencies: %s", err.Reason)
	}
	if len(stacks) > 0 {
		blocking = append(blocking, fmt.Sprintf("stacks (%d)", len(stacks)))
	}

	secrets, err := s.secretStore.ListByProjectID(ctx, projectID, stores.ListParams{})
	if err != nil {
		return errors.InternalServerError("failed to check project dependencies: %s", err.Reason)
	}
	if len(secrets) > 0 {
		blocking = append(blocking, fmt.Sprintf("secrets (%d)", len(secrets)))
	}

	volumes, err := s.volumeStore.ListByProjectID(ctx, projectID)
	if err != nil {
		return errors.InternalServerError("failed to check project dependencies: %s", err.Reason)
	}
	if len(volumes) > 0 {
		blocking = append(blocking, fmt.Sprintf("volumes (%d)", len(volumes)))
	}

	addons, err := s.postgresAddonStore.ListByProjectID(ctx, projectID)
	if err != nil {
		return errors.InternalServerError("failed to check project dependencies: %s", err.Reason)
	}
	if len(addons) > 0 {
		blocking = append(blocking, fmt.Sprintf("postgres addons (%d)", len(addons)))
	}

	objectStores, err := s.objectStoreStore.ListByProjectID(ctx, projectID)
	if err != nil {
		return errors.InternalServerError("failed to check project dependencies: %s", err.Reason)
	}
	if len(objectStores) > 0 {
		blocking = append(blocking, fmt.Sprintf("object stores (%d)", len(objectStores)))
	}

	if len(blocking) > 0 {
		return errors.BadRequest(
			"cannot delete project with existing resources: %s",
			strings.Join(blocking, ", "),
		)
	}

	return nil
}

func (s *projectService) InternalAddMember(ctx context.Context, projectID, userID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError) {
	exists, serr := s.membershipStore.GetByProjectAndUser(ctx, projectID, userID)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		return nil, serr
	}
	if exists != nil {
		return exists, nil
	}

	project, serr := s.projectStore.GetByID(ctx, projectID)
	if serr != nil {
		return nil, serr
	}

	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("project membership role must be Developer or Viewer")
	}

	membership := &models.ProjectMembership{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	created, serr := s.membershipStore.Create(ctx, membership)
	if serr != nil {
		return nil, serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(role), projectID); err != nil {
		s.logger.Error(ctx, "failed to add project role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to add project role grouping")
	}

	if err := s.ensureOrgMemberGrouping(ctx, userID, project.OrganisationID); err != nil {
		s.logger.Error(ctx, "failed to ensure org member grouping: %s", err.Error())
		return nil, err
	}

	return created, nil
}

func (s *projectService) AddMember(ctx context.Context, projectID, userID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError) {
	// Check if membership already exists
	exists, serr := s.membershipStore.GetByProjectAndUser(ctx, projectID, userID)
	if serr != nil && serr.Code != errors.ErrorNotFound {
		return nil, serr
	}
	if exists != nil {
		return nil, errors.Conflict("user is already a member of this project")
	}

	project, serr := s.projectStore.GetByID(ctx, projectID)
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

	if project.OrganisationID != user.OrganisationID {
		return nil, errors.BadRequest("user and project belong to different organisations")
	}

	if permErr := s.permissions.Check(ctx, project.OrganisationID, auth.ResourceProjects, project.ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("project membership role must be Developer or Viewer")
	}

	membership := &models.ProjectMembership{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	created, serr := s.membershipStore.Create(ctx, membership)
	if serr != nil {
		return nil, serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(role), projectID); err != nil {
		s.logger.Error(ctx, "failed to add project role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to add project role grouping")
	}

	if err := s.ensureOrgMemberGrouping(ctx, userID, project.OrganisationID); err != nil {
		s.logger.Error(ctx, "failed to ensure org member grouping: %s", err.Error())
		return nil, err
	}

	return created, nil
}

func (s *projectService) RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError {
	membership, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return serr
	}
	project, serr := s.projectStore.GetByID(ctx, membership.ProjectID)
	if serr != nil {
		return serr
	}
	if permErr := s.permissions.Check(ctx, project.OrganisationID, auth.ResourceProjects, project.ID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if err := s.membershipStore.Delete(ctx, membershipID); err != nil {
		return err
	}

	if rmErr := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.ProjectID); rmErr != nil {
		s.logger.Error(ctx, "failed to remove project role grouping: %s", rmErr.Error())
		return errors.InternalServerError("failed to remove project role grouping: %s", rmErr.Error())
	}

	if err := s.cleanupOrgMemberGrouping(ctx, membership.UserID, project.OrganisationID); err != nil {
		s.logger.Error(ctx, "failed to cleanup org member grouping: %s", err.Error())
		return errors.InternalServerError("failed to cleanup org member grouping")
	}

	return nil
}

func (s *projectService) UpdateMemberRole(ctx context.Context, membershipID string, role models.ProjectRole) (*models.ProjectMembership, *errors.ServiceError) {
	existing, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return nil, serr
	}
	project, serr := s.projectStore.GetByID(ctx, existing.ProjectID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, project.OrganisationID, auth.ResourceProjects, project.ID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("project membership role must be Developer or Viewer")
	}

	res, serr := s.membershipStore.Update(ctx, membershipID, &models.ProjectMembership{Role: role})
	if serr != nil {
		return nil, serr
	}
	// Remove old grouping policy and add new one
	err := s.policyMgr.RemoveGroupingPolicy(existing.UserID, string(existing.Role), existing.ProjectID)
	if err != nil {
		s.logger.Error(ctx, "failed to remove project role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to remove project role grouping")
	}

	if err := s.policyMgr.AddGroupingPolicy(existing.UserID, string(role), existing.ProjectID); err != nil {
		s.logger.Error(ctx, "failed to update project role grouping: %s", err.Error())
		return nil, errors.InternalServerError("failed to update project role grouping")
	}
	return res, nil
}

func (s *projectService) ListMembers(ctx context.Context, projectID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	project, serr := s.projectStore.GetByID(ctx, projectID)
	if serr != nil {
		return nil, serr
	}
	if permErr := s.permissions.Check(ctx, project.OrganisationID, auth.ResourceProjects, project.ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return s.membershipStore.ListByProjectID(ctx, projectID)
}

func (s *projectService) ListUserProjects(ctx context.Context, userID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		if serr.Code == errors.ErrorNotFound {
			return nil, errors.Unauthorized("user not found")
		}
		return nil, serr
	}

	if permErr := s.permissions.Check(ctx, user.OrganisationID, auth.ResourceProjects, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.listUserProjects(ctx, user)
}

func (s *projectService) InternalListUserProjects(ctx context.Context, userID, orgID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return nil, serr
	}
	return s.listUserProjects(ctx, user)
}

// Org admins have org-wide access without any membership row, so their list is
// projected from the org's projects at read time. The API contract only knows
// Developer and Viewer, so the projected role is Developer.
func (s *projectService) listUserProjects(ctx context.Context, user *models.User) ([]*models.ProjectMembership, *errors.ServiceError) {
	if !user.IsOrgAdmin() {
		return s.membershipStore.ListByUserIDAndOrgID(ctx, user.ID, user.OrganisationID)
	}

	projects, serr := s.projectStore.ListByOrgID(ctx, user.OrganisationID)
	if serr != nil {
		return nil, serr
	}
	memberships := make([]*models.ProjectMembership, len(projects))
	for i, project := range projects {
		memberships[i] = &models.ProjectMembership{
			ProjectID: project.ID,
			UserID:    user.ID,
			Role:      models.DeveloperRole,
			Project:   project,
		}
	}
	return memberships, nil
}

func (s *projectService) ensureOrgMemberGrouping(ctx context.Context, userID, orgID string) *errors.ServiceError {
	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID); err != nil {
		return errors.InternalServerError("failed to add OrgMember grouping")
	}
	return nil
}

func (s *projectService) cleanupOrgMemberGrouping(ctx context.Context, userID, orgID string) *errors.ServiceError {
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

var validProjectName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`)

func validateProjectName(name string) *errors.ServiceError {
	if len(name) == 0 {
		return errors.BadRequest("project name is required")
	}
	if len(name) > 63 {
		return errors.BadRequest("project name must be at most 63 characters")
	}
	if len(name) == 1 {
		if (name[0] < 'a' || name[0] > 'z') && (name[0] < '0' || name[0] > '9') {
			return errors.BadRequest("project name must contain only lowercase alphanumeric characters and hyphens, and must start and end with an alphanumeric character")
		}
		return nil
	}
	if !validProjectName.MatchString(name) {
		return errors.BadRequest("project name must contain only lowercase alphanumeric characters and hyphens, and must start and end with an alphanumeric character")
	}
	return nil
}
