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
	TeamService         TeamService
	PolicyManager       resourceaccess.ResourceAccessPolicyManager
	RefreshTokenStore   stores.RefreshTokenStore
	JWTSecretKey        string
	JWTClaimsBuilder    jwtClaimsBuilder
	Logger              logger.Logger
}

type signupService struct {
	userService         UserService
	orgInviteService    OrgInviteService
	organisationService OrganisationService
	teamService         TeamService
	policyManager       resourceaccess.ResourceAccessPolicyManager
	refreshTokenStore   stores.RefreshTokenStore
	jwtSecretKey        string
	jwtClaimsBuilder    jwtClaimsBuilder
	logger              logger.Logger
}

func NewSignupService(spec SignupServiceSpec) SignupService {
	return &signupService{
		userService:         spec.UserService,
		orgInviteService:    spec.OrgInviteService,
		organisationService: spec.OrganisationService,
		teamService:         spec.TeamService,
		policyManager:       spec.PolicyManager,
		refreshTokenStore:   spec.RefreshTokenStore,
		jwtSecretKey:        spec.JWTSecretKey,
		jwtClaimsBuilder:    spec.JWTClaimsBuilder,
		logger:              spec.Logger,
	}
}

func (s *signupService) Signup(ctx context.Context, user *models.User, inviteToken string) (*openapi.UserSignupResponse, *errors.ServiceError) {
	s.logger.Infof("Creating user with email: %s", user.Email)
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
		s.logger.Errorf("failed to hash password, %s", err.Error())
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
	createdOrganisation, orgErr := s.organisationService.InternalCreate(ctx, user.Organisation)
	if orgErr != nil {
		s.logger.Errorf("failed to create organisation, %s", orgErr.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.OrganisationID = createdOrganisation.ID
	user.Role = models.OrgAdminRole

	if _, teamErr := s.teamService.InternalCreateDefaultTeam(ctx, createdOrganisation.ID); teamErr != nil {
		s.logger.Errorf("failed to create default team: %s", teamErr.Error())
		return nil, errors.GeneralError("failed to create default team")
	}

	createdUser, serr := s.userService.InternalCreate(ctx, user)
	if serr != nil {
		return nil, serr
	}

	if policyAddErr := s.policyManager.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgAdminRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		s.logger.Errorf("failed to add OrgAdmin policy for user: %s", policyAddErr.Error())
		return nil, errors.GeneralError("failed to create user")
	}

	if policyAddErr := s.policyManager.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		s.logger.Errorf("failed to add OrgMember policy for user: %s", policyAddErr.Error())
	}

	return s.buildSignupResponse(ctx, createdUser)
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
		s.logger.Errorf("failed to create refresh token: %s", refreshErr.Error())
		return nil, errors.GeneralError("failed to generate refresh token")
	}
	return &openapi.UserSignupResponse{
		User:         ptr.To(presenters.PresentUser(createdUser)),
		JwtToken:     &tokenString,
		RefreshToken: &refreshToken,
	}, nil
}
