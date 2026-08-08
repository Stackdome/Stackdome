package services

import (
	"context"
	"net/mail"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/utils/ptr"
)

type SignupService interface {
	Signup(ctx context.Context, user *models.User, inviteToken string) (*openapi.UserSignupResponse, *errors.ServiceError)
}

type SignupServiceSpec struct {
	UserService         UserService
	OrgInviteService    OrgInviteService
	OrganisationService OrganisationService
	ProjectService      ProjectService
	PolicyManager       resourceaccess.ResourceAccessPolicyManager
	RefreshTokenStore   stores.RefreshTokenStore
	AtomicExecutor      stores.AtomicExecutor
	JWTSecretKey        string
	JWTClaimsBuilder    jwtClaimsBuilder
	Logger              logger.Logger
}

type signupService struct {
	userService         UserService
	orgInviteService    OrgInviteService
	organisationService OrganisationService
	projectService      ProjectService
	policyManager       resourceaccess.ResourceAccessPolicyManager
	refreshTokenStore   stores.RefreshTokenStore
	atomicExecutor      stores.AtomicExecutor
	jwtSecretKey        string
	jwtClaimsBuilder    jwtClaimsBuilder
	logger              logger.Logger
}

func NewSignupService(spec SignupServiceSpec) SignupService {
	return &signupService{
		userService:         spec.UserService,
		orgInviteService:    spec.OrgInviteService,
		organisationService: spec.OrganisationService,
		projectService:      spec.ProjectService,
		policyManager:       spec.PolicyManager,
		refreshTokenStore:   spec.RefreshTokenStore,
		atomicExecutor:      spec.AtomicExecutor,
		jwtSecretKey:        spec.JWTSecretKey,
		jwtClaimsBuilder:    spec.JWTClaimsBuilder,
		logger:              spec.Logger,
	}
}

func (s *signupService) Signup(ctx context.Context, user *models.User, inviteToken string) (*openapi.UserSignupResponse, *errors.ServiceError) {
	s.logger.Info(ctx, "Creating user with email: %s", user.Email)
	if len(user.Password) < 8 {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}

	_, addrErr := mail.ParseAddress(user.Email)
	if addrErr != nil {
		return nil, errors.BadRequest("invalid email address")
	}

	_, gerr := s.userService.InternalGetByEmail(ctx, user.Email)
	if gerr == nil {
		return nil, errors.Conflict("user with this email already exists! Please login instead.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error(ctx, "failed to hash password, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.Password = string(hashedPassword)

	if inviteToken != "" {
		invite, invErr := s.orgInviteService.ValidateAndConsume(ctx, inviteToken, user.Email)
		if invErr != nil {
			return nil, invErr
		}

		createdUser, createErr := s.userService.InternalCreateInvitedUser(ctx, user, invite)
		if createErr != nil {
			return nil, createErr
		}

		return s.buildSignupResponse(ctx, createdUser)
	}

	if user.Organisation == nil {
		return nil, errors.BadRequest("organisation is required")
	}
	if nameErr := validateOrganisationName(user.Organisation.Name); nameErr != nil {
		return nil, nameErr
	}

	var createdUser *models.User
	var granted bool

	txErr := s.atomicExecutor.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		createdOrganisation, orgErr := s.organisationService.InternalCreate(txCtx, user.Organisation)
		if orgErr != nil {
			s.logger.Error(ctx, "failed to create organisation, %s", orgErr.Error())
			return errors.GeneralError("failed to create user")
		}
		user.OrganisationID = createdOrganisation.ID
		user.Role = models.OrgAdminRole

		if _, projectErr := s.projectService.InternalCreateDefaultProject(txCtx, createdOrganisation.ID); projectErr != nil {
			s.logger.Error(ctx, "failed to create default project: %s", projectErr.Error())
			return errors.GeneralError("failed to create default project")
		}

		newUser, serr := s.userService.InternalCreate(txCtx, user)
		if serr != nil {
			return serr
		}
		createdUser = newUser

		granted = true
		if policyErr := grantOrgAdminRoles(s.policyManager, createdUser.ID, createdUser.OrganisationID); policyErr != nil {
			s.logger.Error(ctx, "failed to add org role policies for user: %s", policyErr.Error())
			return errors.GeneralError("failed to create user")
		}
		return nil
	})
	if txErr != nil {
		if granted {
			revokeOrgAdminRoles(s.policyManager, createdUser.ID, createdUser.OrganisationID)
		}
		return nil, txErr
	}

	return s.buildSignupResponse(ctx, createdUser)
}

// Casbin writes cannot join the DB transaction: the enforcer owns its own
// connection and its API takes no context. So the grants are made inside the
// transaction and undone by hand when it fails, leaving no orphan grouping.
func grantOrgAdminRoles(policyManager resourceaccess.ResourceAccessPolicyManager, userID, orgID string) error {
	if err := policyManager.AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
		return err
	}
	return policyManager.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID)
}

func revokeOrgAdminRoles(policyManager resourceaccess.ResourceAccessPolicyManager, userID, orgID string) {
	_ = policyManager.RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID)
	_ = policyManager.RemoveGroupingPolicy(userID, string(models.OrgMemberRole), orgID)
}

func (s *signupService) buildSignupResponse(ctx context.Context, createdUser *models.User) (*openapi.UserSignupResponse, *errors.ServiceError) {
	expirationTime := time.Now().UTC().Add(auth.JwtTokenExpiry)
	claims := s.jwtClaimsBuilder.BuildClaims(createdUser, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(s.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}
	refreshToken, refreshErr := auth.CreateRefreshToken(ctx, s.refreshTokenStore, createdUser.ID)
	if refreshErr != nil {
		s.logger.Error(ctx, "failed to create refresh token: %s", refreshErr.Error())
		return nil, errors.GeneralError("failed to generate refresh token")
	}
	return &openapi.UserSignupResponse{
		User:         ptr.To(presenters.PresentUser(createdUser)),
		JwtToken:     &tokenString,
		RefreshToken: &refreshToken,
	}, nil
}
