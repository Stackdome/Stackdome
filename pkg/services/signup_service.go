package services

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/slug"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"k8s.io/utils/ptr"
)

const (
	shortOrgIDLength      = 8
	maxDomainSeedAttempts = 5
)

type SignupService interface {
	Signup(ctx context.Context, user *models.User, inviteToken string) (*openapi.UserSignupResponse, *errors.ServiceError)
}

type SignupServiceSpec struct {
	UserService               UserService
	OrgInviteService          OrgInviteService
	OrganisationService       OrganisationService
	ProjectService            ProjectService
	ClusterService            ClusterService
	OrganisationDomainService OrganisationDomainsService
	ImageRegistryService      ImageRegistryService
	PolicyManager             resourceaccess.ResourceAccessPolicyManager
	RefreshTokenStore         stores.RefreshTokenStore
	JWTSecretKey              string
	JWTClaimsBuilder          jwtClaimsBuilder
	OrgRegistryDefaults       models.OrgRegistryDefaults
	Logger                    logger.Logger
}

type signupService struct {
	userService               UserService
	orgInviteService          OrgInviteService
	organisationService       OrganisationService
	projectService            ProjectService
	clusterService            ClusterService
	organisationDomainService OrganisationDomainsService
	imageRegistryService      ImageRegistryService
	policyManager             resourceaccess.ResourceAccessPolicyManager
	refreshTokenStore         stores.RefreshTokenStore
	jwtSecretKey              string
	jwtClaimsBuilder          jwtClaimsBuilder
	orgRegistryDefaults       models.OrgRegistryDefaults
	logger                    logger.Logger
}

func NewSignupService(spec SignupServiceSpec) SignupService {
	return &signupService{
		userService:               spec.UserService,
		orgInviteService:          spec.OrgInviteService,
		organisationService:       spec.OrganisationService,
		projectService:            spec.ProjectService,
		clusterService:            spec.ClusterService,
		organisationDomainService: spec.OrganisationDomainService,
		imageRegistryService:      spec.ImageRegistryService,
		policyManager:             spec.PolicyManager,
		refreshTokenStore:         spec.RefreshTokenStore,
		jwtSecretKey:              spec.JWTSecretKey,
		jwtClaimsBuilder:          spec.JWTClaimsBuilder,
		orgRegistryDefaults:       spec.OrgRegistryDefaults,
		logger:                    spec.Logger,
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
	createdOrganisation, orgErr := s.organisationService.InternalCreate(ctx, user.Organisation)
	if orgErr != nil {
		s.logger.Error(ctx, "failed to create organisation, %s", orgErr.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.OrganisationID = createdOrganisation.ID
	user.Role = models.OrgAdminRole

	if _, projectErr := s.projectService.InternalCreateDefaultProject(ctx, createdOrganisation.ID); projectErr != nil {
		s.logger.Error(ctx, "failed to create default project: %s", projectErr.Error())
		return nil, errors.GeneralError("failed to create default project")
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
		s.logger.Error(ctx, "failed to add OrgAdmin policy for user: %s", policyAddErr.Error())
		return nil, errors.GeneralError("failed to create user")
	}

	if policyAddErr := s.policyManager.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		s.logger.Error(ctx, "failed to add OrgMember policy for user: %s", policyAddErr.Error())
	}

	if seedErr := s.seedPlatformInfra(ctx, createdUser.OrganisationID, user.Organisation.Name); seedErr != nil {
		return nil, seedErr
	}

	return s.buildSignupResponse(ctx, createdUser)
}

func (s *signupService) seedPlatformInfra(ctx context.Context, orgID, orgName string) *errors.ServiceError {
	platformCluster, err := s.clusterService.GetPlatformCluster(ctx)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return nil
		}
		return err
	}
	ctx = auth.SetIdentityInContext(ctx, &auth.Identity{IsSystem: true, OrgID: orgID})

	platformOrg, pErr := s.organisationService.InternalGetPlatformOrg(ctx)
	if pErr != nil {
		return pErr
	}
	baseDomain, dErr := s.organisationDomainService.GetDefaultDomainForOrganisation(ctx, platformOrg.ID)
	if dErr != nil {
		return dErr
	}

	orgSlug := slug.FromOrgName(orgName)
	if sErr := s.seedOrgDomain(ctx, orgID, orgSlug, baseDomain.Domain); sErr != nil {
		return sErr
	}
	return s.seedOrgRegistry(ctx, orgID, orgSlug, platformCluster.ID)
}

func (s *signupService) seedOrgDomain(ctx context.Context, orgID, orgSlug, baseDomain string) *errors.ServiceError {
	candidate := fmt.Sprintf("%s.%s", orgSlug, baseDomain)
	for attempt := 0; attempt < maxDomainSeedAttempts; attempt++ {
		_, err := s.organisationDomainService.Create(ctx, &models.OrganisationDomain{
			OrganisationID: orgID,
			Domain:         candidate,
		})
		if err == nil {
			return nil
		}
		if err.Code != errors.ErrorConflict {
			return err
		}
		candidate = fmt.Sprintf("%s-%s.%s", orgSlug, slug.RandomSuffix(), baseDomain)
	}
	return errors.Conflict("could not allocate a unique domain for organisation")
}

func (s *signupService) seedOrgRegistry(ctx context.Context, orgID, orgSlug, clusterID string) *errors.ServiceError {
	shortID := strings.ReplaceAll(orgID, "-", "")
	if len(shortID) > shortOrgIDLength {
		shortID = shortID[:shortOrgIDLength]
	}
	_, err := s.imageRegistryService.Create(ctx, &models.ClusterImageRegistry{
		ClusterID:           clusterID,
		OrganisationID:      orgID,
		Name:                fmt.Sprintf("%s-%s", orgSlug, shortID),
		BackendStorageSize:  s.orgRegistryDefaults.StorageSize,
		BackendStorageClass: s.orgRegistryDefaults.StorageClass,
	})
	return err
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
