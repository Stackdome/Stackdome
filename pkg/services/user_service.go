package services

import (
	"context"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/utils/ptr"
)

type UserService interface {
	Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
	GetUserFromContext(ctx context.Context) (*models.User, []*models.TeamMembership, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string, params stores.PaginationParams) (*stores.PaginatedResult[*models.User], *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.User, *errors.ServiceError)
	InternalGetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error)
	Signup(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError)
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
		teamService:             spec.TeamService,
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
	TeamService                 TeamService
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
	teamService             TeamService
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

func (u usersService) ListByOrgID(ctx context.Context, orgID string, params stores.PaginationParams) (*stores.PaginatedResult[*models.User], *errors.ServiceError) {
	if permErr := u.permissions.Check(ctx, orgID, auth.ResourceUsers, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return u.userStore.ListByOrgID(ctx, orgID, params)
}

func (u usersService) InternalGet(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) GetUserFromContext(ctx context.Context) (*models.User, []*models.TeamMembership, *errors.ServiceError) {
	identity := auth.GetIdentityFromCtx(ctx)
	if identity == nil || identity.UserID == "" {
		return nil, nil, errors.Unauthorized("failed to fetch user")
	}
	user, serr := u.userStore.GetByID(ctx, identity.UserID)
	if serr != nil {
		return nil, nil, serr
	}
	memberships, membErr := u.teamService.InternalListUserTeams(ctx, user.ID, user.OrganisationID)
	if membErr != nil {
		u.logger.Errorf("failed to list user teams: %s", membErr.Error())
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

func (u usersService) Signup(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError) {
	u.logger.Infof("Creating user with email: %s", user.Email)
	if len(user.Password) < 8 {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Errorf("failed to hash password, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.Password = string(hashedPassword)

	if user.Organisation == nil {
		return nil, errors.BadRequest("organisation is required")
	}
	createdOrganisation, err := u.organisationService.InternalCreate(ctx, user.Organisation)
	if err != nil {
		u.logger.Errorf("failed to create organisation, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.OrganisationID = createdOrganisation.ID
	user.Role = models.OrgAdminRole

	if _, teamErr := u.teamService.InternalCreateDefaultTeam(ctx, createdOrganisation.ID); teamErr != nil {
		u.logger.Errorf("failed to create default team: %s", teamErr.Error())
		return nil, errors.GeneralError("failed to create default team")
	}

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, serr
	}

	if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgAdminRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		u.logger.Errorf("failed to add OrgAdmin policy for user: %s", policyAddErr.Error())
		return nil, errors.GeneralError("failed to create user")
	}

	// Org admin is also an org member, so they get both policies.
	// This is needed because another orgadmin can demote the user from OrgAdmin.
	if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		u.logger.Errorf("failed to add OrgMember policy for user: %s", policyAddErr.Error())
	}
	expirationTime := time.Now().UTC().Add(auth.JwtTokenExpiry)
	claims := u.jwtClaimsBuilder.BuildClaims(createdUser, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}
	refreshToken, refreshErr := auth.CreateRefreshToken(ctx, u.refreshTokenStore, createdUser.ID)
	if refreshErr != nil {
		u.logger.Errorf("failed to create refresh token: %s", refreshErr.Error())
		return nil, errors.GeneralError("failed to generate refresh token")
	}

	// All signup users are org admins, And org admins dont belong to any team.
	// Hence team membetship is not included in the response.
	res := openapi.UserSignupResponse{
		User:         ptr.To(presenters.PresentUser(createdUser)),
		JwtToken:     &tokenString,
		RefreshToken: &refreshToken,
	}

	return &res, nil
}

func (u usersService) InternalCreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error) {
	org := &models.Organisation{Name: models.UserOrgNameFromOauth(name)}
	createdOrg, serr := u.organisationService.InternalCreate(ctx, org)
	if serr != nil {
		return nil, fmt.Errorf("failed to create organisation for oauth user: %w", serr)
	}

	if u.teamService != nil {
		if _, teamErr := u.teamService.InternalCreateDefaultTeam(ctx, createdOrg.ID); teamErr != nil {
			return nil, fmt.Errorf("failed to create default team for oauth user: %w", teamErr)
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
		u.logger.Errorf("failed to add OrgAdmin policy for oauth user: %s", policyErr.Error())
		return nil, fmt.Errorf("failed to add access policy: %w", policyErr)
	}

	if policyErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyErr != nil {
		u.logger.Errorf("failed to add OrgMember policy for oauth user: %s", policyErr.Error())
	}

	return createdUser, nil
}

func (u usersService) Login(ctx context.Context, loginRequest *openapi.LoginRequest) (*openapi.LoginResponse, *errors.ServiceError) {
	userInDB, err := u.userStore.GetByEmail(ctx, loginRequest.GetEmail())
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userInDB.Password), []byte(loginRequest.Password)); err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, errors.BadRequest("invalid email or password")
		}
		u.logger.Errorf("failed to login: %s", err.Error())
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
		u.logger.Errorf("failed to create refresh token: %s", refreshErr.Error())
		return nil, errors.GeneralError("failed to generate refresh token")
	}

	memberships, membErr := u.teamService.InternalListUserTeams(ctx, userInDB.ID, userInDB.OrganisationID)
	if membErr != nil {
		u.logger.Errorf("failed to list user teams: %s", membErr.Error())
		return nil, errors.GeneralError("failed to list user teams")
	}

	res := openapi.NewLoginResponse()
	res.SetToken(tokenString)
	res.SetRefreshToken(refreshToken)
	res.SetUser(presenters.PresentUserWithTeams(userInDB, memberships))
	return res, nil
}
