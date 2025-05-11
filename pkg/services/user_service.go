package services

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
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
	GetDefaultUser(ctx context.Context) (*models.User, *errors.ServiceError)
	Create(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError)
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
	}
}

type UserServiceSpec struct {
	SessionFactory              db.SessionFactory
	Logger                      logger.Logger
	JwtSecretKey                string
	ResourceAccessPolicyManager resourceaccess.ResourceAccessPolicyManager
	JWTClaimsBuilder            jwtClaimsBuilder
	OrganisationService         OrganisationService
}

type usersService struct {
	userStore               stores.UserStore
	organisationService     OrganisationService
	logger                  logger.Logger
	jwtSecretKey            string
	resourceAccessPolicyMgr resourceaccess.ResourceAccessPolicyManager
	jwtClaimsBuilder        jwtClaimsBuilder
}

func (u usersService) Get(ctx context.Context, ID string) (*models.User, *errors.ServiceError) {
	return u.userStore.GetByID(ctx, ID)
}

func (u usersService) GetDefaultUser(ctx context.Context) (*models.User, *errors.ServiceError) {
	return u.userStore.GetDefaultUser(ctx)
}

func (u usersService) Create(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError) {
	if len(user.Password) < 8 {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Errorf("failed to hash password, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.Password = string(hashedPassword)
	if len(user.Role) == 0 {
		user.Role = models.UserRole
	}

	if user.OrganisationID == "" {
		// We create an organisation for the user if organisation ID is not provided.
		createdOrganisation, err := u.organisationService.Create(ctx, user.Organisation)
		if err != nil {
			u.logger.Errorf("failed to create organisation, %s", err.Error())
			return nil, errors.GeneralError("failed to create user")
		}
		user.OrganisationID = createdOrganisation.ID
		// That mean this user is the organisation admin.
		user.Role = models.OrganisationAdminRole
	}
	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, serr
	}

	// TODO: Wrap user creation in a transaction so that we can rollback if policyAddErr is not nil.
	if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		createdUser.Role.String(),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		u.logger.Errorf("failed to add policy for user: %s", policyAddErr.Error())
		return nil, errors.GeneralError("failed to create user")
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
