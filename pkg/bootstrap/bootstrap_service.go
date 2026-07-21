package bootstrap

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/services"
	"golang.org/x/crypto/bcrypt"
)

type Spec struct {
	UserService               services.UserService
	OrganisationService       services.OrganisationService
	ProjectService            services.ProjectService
	ClusterService            services.ClusterService
	ImageRegistryService      services.ImageRegistryService
	OrganisationDomainService services.OrganisationDomainsService
	PolicyManager             resourceaccess.ResourceAccessPolicyManager
	BootstrapConfig           *config.BootstrapConfig
	ClusterConfig             *config.ClusterConfig
	Logger                    logger.Logger
}

type Service struct {
	userService               services.UserService
	organisationService       services.OrganisationService
	projectService            services.ProjectService
	clusterService            services.ClusterService
	imageRegistryService      services.ImageRegistryService
	organisationDomainService services.OrganisationDomainsService
	policyManager             resourceaccess.ResourceAccessPolicyManager
	bootstrapCfg              *config.BootstrapConfig
	clusterCfg                *config.ClusterConfig
	logger                    logger.Logger
}

func NewService(spec Spec) *Service {
	return &Service{
		userService:               spec.UserService,
		organisationService:       spec.OrganisationService,
		projectService:            spec.ProjectService,
		clusterService:            spec.ClusterService,
		imageRegistryService:      spec.ImageRegistryService,
		organisationDomainService: spec.OrganisationDomainService,
		policyManager:             spec.PolicyManager,
		bootstrapCfg:              spec.BootstrapConfig,
		clusterCfg:                spec.ClusterConfig,
		logger:                    spec.Logger,
	}
}

func (s *Service) Run(ctx context.Context) error {
	if !s.clusterCfg.IsSet() {
		s.logger.Info(ctx, "no default cluster configured; skipping platform bootstrap")
		return nil
	}

	admin, org, err := s.upsertAdminAndOrg(ctx)
	if err != nil {
		return err
	}

	sysCtx := auth.SetUserInContext(
		auth.SetIdentityInContext(ctx, &auth.Identity{IsSystem: true, UserID: admin.ID, OrgID: org.ID}),
		admin,
	)

	cluster, cErr := s.clusterService.InternalUpsertDefaultCluster(sysCtx, &models.Cluster{
		Name:           s.clusterCfg.Name,
		OrganisationID: org.ID,
		Default:        true,
		ClusterURL:     s.clusterCfg.ClusterURL,
		ClusterCAData:  s.clusterCfg.ClusterCAData,
		Token:          s.clusterCfg.Token,
	})
	if cErr != nil {
		return fmt.Errorf("failed to upsert default cluster: %w", cErr)
	}

	if err := s.ensurePlatformRegistry(sysCtx, org.ID, cluster.ID); err != nil {
		return err
	}

	return s.ensurePlatformDomain(sysCtx, org.ID)
}

func (s *Service) upsertAdminAndOrg(ctx context.Context) (*models.User, *models.Organisation, error) {
	existing, gErr := s.userService.InternalGetByEmail(ctx, s.bootstrapCfg.DefaultUser.Email)
	if gErr != nil && gErr.Code != errors.ErrorNotFound {
		return nil, nil, fmt.Errorf("failed to look up platform admin: %w", gErr)
	}
	if gErr == nil {
		org, oErr := s.organisationService.InternalGetDefaultOrg(ctx)
		if oErr != nil {
			return nil, nil, fmt.Errorf("default org missing for existing admin: %w", oErr)
		}
		if bcrypt.CompareHashAndPassword([]byte(existing.Password), []byte(s.bootstrapCfg.DefaultUser.Password)) != nil {
			hashed, hErr := bcrypt.GenerateFromPassword([]byte(s.bootstrapCfg.DefaultUser.Password), bcrypt.DefaultCost)
			if hErr != nil {
				return nil, nil, fmt.Errorf("failed to hash admin password: %w", hErr)
			}
			if uErr := s.userService.InternalUpdatePassword(ctx, existing.ID, string(hashed)); uErr != nil {
				return nil, nil, fmt.Errorf("failed to update admin password: %w", uErr)
			}
		}
		return existing, org, nil
	}

	org, oErr := s.organisationService.InternalCreate(ctx, &models.Organisation{
		Name:    s.bootstrapCfg.PlatformOrgName,
		Default: true,
	})
	if oErr != nil {
		return nil, nil, fmt.Errorf("failed to create platform org: %w", oErr)
	}
	if _, pErr := s.projectService.InternalCreateDefaultProject(ctx, org.ID); pErr != nil {
		return nil, nil, fmt.Errorf("failed to create platform default project: %w", pErr)
	}
	hashed, hErr := bcrypt.GenerateFromPassword([]byte(s.bootstrapCfg.DefaultUser.Password), bcrypt.DefaultCost)
	if hErr != nil {
		return nil, nil, fmt.Errorf("failed to hash admin password: %w", hErr)
	}
	user, uErr := s.userService.InternalCreate(ctx, &models.User{
		Email:          s.bootstrapCfg.DefaultUser.Email,
		Name:           s.bootstrapCfg.DefaultUser.Name,
		Password:       string(hashed),
		Role:           models.OrgAdminRole,
		OrganisationID: org.ID,
	})
	if uErr != nil {
		return nil, nil, fmt.Errorf("failed to create platform admin: %w", uErr)
	}
	if pErr := s.policyManager.AddGroupingPolicy(user.ID, string(models.OrgAdminRole), org.ID); pErr != nil {
		return nil, nil, fmt.Errorf("failed to add OrgAdmin policy: %w", pErr)
	}
	if pErr := s.policyManager.AddGroupingPolicy(user.ID, string(models.OrgMemberRole), org.ID); pErr != nil {
		return nil, nil, fmt.Errorf("failed to add OrgMember policy: %w", pErr)
	}
	return user, org, nil
}

func (s *Service) ensurePlatformRegistry(ctx context.Context, orgID, clusterID string) error {
	if s.bootstrapCfg.RegistryName == "" {
		return nil
	}
	if _, err := s.imageRegistryService.GetForOrg(ctx, orgID); err == nil {
		return nil
	} else if err.Code != errors.ErrorNotFound {
		return fmt.Errorf("failed to check platform registry: %w", err)
	}
	if _, cErr := s.imageRegistryService.Create(ctx, &models.ClusterImageRegistry{
		ClusterID:           clusterID,
		OrganisationID:      orgID,
		Name:                s.bootstrapCfg.RegistryName,
		BackendStorageSize:  s.bootstrapCfg.RegistryStorageSize,
		BackendStorageClass: s.bootstrapCfg.RegistryStorageClass,
	}); cErr != nil {
		return fmt.Errorf("failed to create platform registry: %w", cErr)
	}
	return nil
}

func (s *Service) ensurePlatformDomain(ctx context.Context, orgID string) error {
	existing, err := s.organisationDomainService.GetDefaultDomainForOrganisation(ctx, orgID)
	if err == nil && existing.Domain == s.bootstrapCfg.BaseDomain {
		return nil
	}
	if err != nil && err.Code != errors.ErrorNotFound {
		return fmt.Errorf("failed to check platform domain: %w", err)
	}
	if _, cErr := s.organisationDomainService.Create(ctx, &models.OrganisationDomain{
		OrganisationID: orgID,
		Domain:         s.bootstrapCfg.BaseDomain,
	}); cErr != nil && cErr.Code != errors.ErrorConflict {
		return fmt.Errorf("failed to create platform domain: %w", cErr)
	}
	return nil
}
