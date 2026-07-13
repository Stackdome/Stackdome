package services

import (
	"context"
	stderrors "errors"
	"fmt"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
	GetUserFromContext(ctx context.Context) (*models.User, []*models.ProjectMembership, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string, params stores.ListParams) (*stores.PaginatedResult[*models.User], *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
	InternalGetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	InternalCreate(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError)
	InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error)
	InternalCreateInvitedUser(ctx context.Context, user *models.User, invite *models.OrgInvite) (*models.User, *errors.ServiceError)
	InternalCreateInvitedOAuthUser(ctx context.Context, email, name, githubID, avatarURL string, invite *models.OrgInvite) (*models.User, error)
	Login(ctx context.Context, loginRequest *openapi.LoginRequest) (*openapi.LoginResponse, *errors.ServiceError)
}

type jwtClaimsBuilder interface {
	BuildClaims(user *models.User, expirationTime time.Time) jwt.Claims
}

var _ UserService = &usersService{}

func NewUserService(spec UserServiceSpec) UserService {
	return &usersService{
		userStore: pgstore.NewUserStore(
			pgstore.UserStoreSpec{
				SessionFactory: spec.SessionFactory,
			},
		),
		logger:                  spec.Logger,
		jwtSecretKey:            spec.JwtSecretKey,
		resourceAccessPolicyMgr: spec.ResourceAccessPolicyManager,
		jwtClaimsBuilder:        spec.JWTClaimsBuilder,
		organisationService:     spec.OrganisationService,
		permissions:             spec.Permissions,
		projectService:          spec.ProjectService,
		refreshTokenStore:       spec.RefreshTokenStore,
	}
}

type UserServiceSpec struct {
	SessionFactory              db.SessionFactory
	Logger                      logger.Logger
	JwtSecretKey                string
	ResourceAccessPolicyManager resourceaccess.ResourceAccessPolicyManager
	JWTClaimsBuilder            jwtClaimsBuilder
	OrganisationService         OrganisationService
	Permissions                 auth.PermissionService
	ProjectService              ProjectService
	RefreshTokenStore           stores.RefreshTokenStore
}

type usersService struct {
	userStore               stores.UserStore
	organisationService     OrganisationService
	logger                  logger.Logger
	jwtSecretKey            string
	resourceAccessPolicyMgr resourceaccess.ResourceAccessPolicyManager
	jwtClaimsBuilder        jwtClaimsBuilder
	permissions             auth.PermissionService
	projectService          ProjectService
	refreshTokenStore       stores.RefreshTokenStore
}

func (u usersService) Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("failed to fetch user")
	}

	if permErr := u.permissions.Check(ctx, identity.OrgID, auth.ResourceUsers, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) ListByOrgID(ctx context.Context, orgID string, params stores.ListParams) (*stores.PaginatedResult[*models.User], *errors.ServiceError) {
	if permErr := u.permissions.Check(ctx, orgID, auth.ResourceUsers, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return u.userStore.ListByOrgID(ctx, orgID, params)
}

func (u usersService) InternalGet(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) GetUserFromContext(ctx context.Context) (*models.User, []*models.ProjectMembership, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil || identity.UserID == "" {
		return nil, nil, errors.Unauthorized("failed to fetch user")
	}
	user, serr := u.userStore.GetByID(ctx, identity.UserID)
	if serr != nil {
		return nil, nil, serr
	}
	memberships, membErr := u.projectService.InternalListUserProjects(ctx, user.ID, user.OrganisationID)
	if membErr != nil {
		u.logger.Error(ctx, "failed to list user projects: %s", membErr.Error())
	}
	return user, memberships, nil
}

func (u usersService) GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil {
		return nil, errors.Unauthorized("failed to fetch user")
	}

	if permErr := u.permissions.Check(ctx, identity.OrgID, auth.ResourceUsers, "", auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return u.userStore.GetByEmail(ctx, email)
}

func (u usersService) InternalGetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByEmail(ctx, email)
}

func (u usersService) InternalCreate(ctx context.Context, user *models.User) (*models.User, *errors.ServiceError) {
	return u.userStore.Create(ctx, user)
}

func (u usersService) InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error) {
	org := &models.Organisation{Name: models.UserOrgNameFromOauth(name)}
	createdOrg, serr := u.organisationService.InternalCreate(ctx, org)
	if serr != nil {
		return nil, fmt.Errorf("failed to create organisation for oauth user: %w", serr)
	}

	if u.projectService != nil {
		if _, projectErr := u.projectService.InternalCreateDefaultProject(ctx, createdOrg.ID); projectErr != nil {
			return nil, fmt.Errorf("failed to create default project for oauth user: %w", projectErr)
		}
	}

	user := &models.User{
		Email:          email,
		Name:           name,
		Role:           models.OrgAdminRole,
		OrganisationID: createdOrg.ID,
		GithubID:       &githubID,
		AvatarURL:      &avatarURL,
	}

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, fmt.Errorf("failed to create oauth user: %w", serr)
	}

	if policyErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgAdminRole),
		createdUser.OrganisationID,
	); policyErr != nil {
		u.logger.Error(ctx, "failed to add OrgAdmin policy for oauth user: %s", policyErr.Error())
		return nil, fmt.Errorf("failed to add access policy: %w", policyErr)
	}

	if policyErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyErr != nil {
		u.logger.Error(ctx, "failed to add OrgMember policy for oauth user: %s", policyErr.Error())
	}

	return createdUser, nil
}

func (u usersService) InternalCreateInvitedUser(ctx context.Context, user *models.User, invite *models.OrgInvite) (*models.User, *errors.ServiceError) {
	user.OrganisationID = invite.OrganisationID
	user.Role = models.NoRole

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, serr
	}

	if _, projectErr := u.projectService.InternalAddMember(ctx, invite.ProjectID, createdUser.ID, invite.ProjectRole); projectErr != nil {
		u.logger.Error(ctx, "failed to add invited user to project: %s", projectErr.Error())
		return nil, errors.GeneralError("failed to add user to project")
	}

	return createdUser, nil
}

func (u usersService) InternalCreateInvitedOAuthUser(ctx context.Context, email, name, githubID, avatarURL string, invite *models.OrgInvite) (*models.User, error) {
	user := &models.User{
		Email:          email,
		Name:           name,
		Role:           models.NoRole,
		OrganisationID: invite.OrganisationID,
		GithubID:       &githubID,
		AvatarURL:      &avatarURL,
	}

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, fmt.Errorf("failed to create invited oauth user: %w", serr)
	}

	if _, projectErr := u.projectService.InternalAddMember(ctx, invite.ProjectID, createdUser.ID, invite.ProjectRole); projectErr != nil {
		u.logger.Error(ctx, "failed to add invited oauth user to project: %s", projectErr.Error())
		return nil, fmt.Errorf("failed to add user to project: %w", projectErr)
	}

	return createdUser, nil
}

func (u usersService) Login(ctx context.Context, loginRequest *openapi.LoginRequest) (*openapi.LoginResponse, *errors.ServiceError) {
	userInDB, err := u.userStore.GetByEmail(ctx, loginRequest.GetEmail())
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(loginRequest.Password)); err != nil {
		if stderrors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, errors.BadRequest("invalid email or password")
		}
		u.logger.Error(ctx, "failed to login: %s", err.Error())
		return nil, errors.GeneralError("failed to login")
	}
	expirationTime := time.Now().UTC().Add(auth.JwtTokenExpiry)
	claims := u.jwtClaimsBuilder.BuildClaims(userInDB, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}

	refreshToken, refreshErr := auth.CreateRefreshToken(ctx, u.refreshTokenStore, userInDB.ID)
	if refreshErr != nil {
		u.logger.Error(ctx, "failed to create refresh token: %s", refreshErr.Error())
		return nil, errors.GeneralError("failed to generate refresh token")
	}

	memberships, membErr := u.projectService.InternalListUserProjects(ctx, userInDB.ID, userInDB.OrganisationID)
	if membErr != nil {
		u.logger.Error(ctx, "failed to list user projects: %s", membErr.Error())
		return nil, errors.GeneralError("failed to list user projects")
	}

	res := openapi.NewLoginResponse()
	res.SetToken(tokenString)
	res.SetRefreshToken(refreshToken)
	res.SetUser(presenters.PresentUserWithProjects(userInDB, memberships))
	return res, nil
}
