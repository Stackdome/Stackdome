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
	GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError)
	Create(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError)
	CreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error)
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
}

func (u usersService) Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(u.permissions, ctx, "", auth.ResourceUsers, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) GetByEmail(ctx context.Context, email string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByEmail(ctx, email)
}

func (u usersService) Create(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError) {
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

	if user.OrganisationID == "" {
		if user.Organisation == nil {
			return nil, errors.BadRequest("organisation is required")
		}
		createdOrganisation, err := u.organisationService.Create(ctx, user.Organisation)
		if err != nil {
			u.logger.Errorf("failed to create organisation, %s", err.Error())
			return nil, errors.GeneralError("failed to create user")
		}
		user.OrganisationID = createdOrganisation.ID
		user.Role = models.OrgAdminRole

		if u.teamService != nil {
			if _, teamErr := u.teamService.CreateDefaultTeam(ctx, createdOrganisation.ID); teamErr != nil {
				u.logger.Errorf("failed to create default team: %s", teamErr.Error())
				return nil, errors.GeneralError("failed to create default team")
			}
		}
	}

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, serr
	}

	if createdUser.Role == models.OrgAdminRole {
		if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
			createdUser.ID,
			string(models.OrgAdminRole),
			createdUser.OrganisationID,
		); policyAddErr != nil {
			u.logger.Errorf("failed to add OrgAdmin policy for user: %s", policyAddErr.Error())
			return nil, errors.GeneralError("failed to create user")
		}
	}

	if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		u.logger.Errorf("failed to add OrgMember policy for user: %s", policyAddErr.Error())
	}
	expirationTime := time.Now().UTC().Add(10 * 24 * time.Hour)
	claims := u.jwtClaimsBuilder.BuildClaims(createdUser, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}
	res := openapi.UserSignupResponse{
		User:     ptr.To(presenters.PresentUser(createdUser)),
		JwtToken: &tokenString,
	}

	return &res, nil
}

func (u usersService) CreateOAuthUser(ctx context.Context, email, name, githubID, avatarURL string) (*models.User, error) {
	org := &models.Organisation{Name: fmt.Sprintf("%s-org", name)}
	createdOrg, serr := u.organisationService.Create(ctx, org)
	if serr != nil {
		return nil, fmt.Errorf("failed to create organisation for oauth user: %w", serr)
	}

	if u.teamService != nil {
		if _, teamErr := u.teamService.CreateDefaultTeam(ctx, createdOrg.ID); teamErr != nil {
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
	expirationTime := time.Now().UTC().Add(10 * 24 * time.Hour)
	claims := u.jwtClaimsBuilder.BuildClaims(userInDB, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}

	res := openapi.NewLoginResponse()
	res.SetToken(tokenString)
	res.SetUser(presenters.PresentUser(userInDB))
	return res, nil
}
