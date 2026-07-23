package bootstrap

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
)

type Spec struct {
	OrganisationService       services.OrganisationService
	ClusterService            services.ClusterService
	OrganisationDomainService services.OrganisationDomainsService
	BootstrapConfig           *config.BootstrapConfig
	ClusterConfig             *config.ClusterConfig
	Logger                    logger.Logger
}

type Service struct {
	organisationService       services.OrganisationService
	clusterService            services.ClusterService
	organisationDomainService services.OrganisationDomainsService
	bootstrapCfg              *config.BootstrapConfig
	clusterCfg                *config.ClusterConfig
	logger                    logger.Logger
}

func NewService(spec Spec) *Service {
	return &Service{
		organisationService:       spec.OrganisationService,
		clusterService:            spec.ClusterService,
		organisationDomainService: spec.OrganisationDomainService,
		bootstrapCfg:              spec.BootstrapConfig,
		clusterCfg:                spec.ClusterConfig,
		logger:                    spec.Logger,
	}
}

// Run provisions the platform org, cluster and base domain.
// The platform org is infrastructure-only: no users, no projects.
func (s *Service) Run(ctx context.Context) error {
	if !s.clusterCfg.IsSet() {
		s.logger.Info(ctx, "no platform cluster configured; skipping platform bootstrap")
		return nil
	}

	org, err := s.ensurePlatformOrg(ctx)
	if err != nil {
		return err
	}

	sysCtx := auth.SetIdentityInContext(ctx, &auth.Identity{
		IsSystem:     true,
		OrgID:        org.ID,
		ContactEmail: s.bootstrapCfg.Email,
	})

	if _, cErr := s.clusterService.InternalUpsertPlatformCluster(sysCtx, &models.Cluster{
		Name:           s.clusterCfg.Name,
		OrganisationID: org.ID,
		Platform:       true,
		ClusterURL:     s.clusterCfg.ClusterURL,
		ClusterCAData:  s.clusterCfg.ClusterCAData,
		Token:          s.clusterCfg.Token,
	}); cErr != nil {
		return fmt.Errorf("failed to upsert platform cluster: %w", cErr)
	}

	return s.ensurePlatformDomain(sysCtx, org.ID)
}

func (s *Service) ensurePlatformOrg(ctx context.Context) (*models.Organisation, error) {
	org, gErr := s.organisationService.InternalGetPlatformOrg(ctx)
	if gErr == nil {
		return org, nil
	}
	if gErr.Code != errors.ErrorNotFound {
		return nil, fmt.Errorf("failed to look up platform org: %w", gErr)
	}

	created, cErr := s.organisationService.InternalCreate(ctx, &models.Organisation{
		Name:     models.PlatformOrganisationName,
		Platform: true,
	})
	if cErr != nil {
		return nil, fmt.Errorf("failed to create platform org: %w", cErr)
	}
	return created, nil
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
