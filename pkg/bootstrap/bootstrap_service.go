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
	OrganisationService services.OrganisationService
	ClusterService      services.ClusterService
	BootstrapConfig     *config.BootstrapConfig
	ClusterConfig       *config.ClusterConfig
	Logger              logger.Logger
}

type Service struct {
	organisationService services.OrganisationService
	clusterService      services.ClusterService
	bootstrapCfg        *config.BootstrapConfig
	clusterCfg          *config.ClusterConfig
	logger              logger.Logger
}

func NewService(spec Spec) *Service {
	return &Service{
		organisationService: spec.OrganisationService,
		clusterService:      spec.ClusterService,
		bootstrapCfg:        spec.BootstrapConfig,
		clusterCfg:          spec.ClusterConfig,
		logger:              spec.Logger,
	}
}

// Run provisions the platform org and shared-compute cluster, then conditionally
// reconciles platform wildcard TLS.
// The platform org is infrastructure-only: no users, no projects.
func (s *Service) Run(ctx context.Context) error {
	if !s.clusterCfg.IsSet() {
		s.logger.Info(ctx, "no shared-compute cluster configured; skipping shared-compute bootstrap")
		return nil
	}

	org, err := s.ensurePlatformOrg(ctx)
	if err != nil {
		return err
	}

	systemIdentity := &auth.Identity{IsSystem: true, OrgID: org.ID}
	if s.bootstrapCfg.PlatformTLSEnabled {
		systemIdentity.ContactEmail = s.bootstrapCfg.Email
	}
	sysCtx := auth.SetIdentityInContext(ctx, systemIdentity)

	cluster, cErr := s.clusterService.InternalUpsertSharedComputeCluster(sysCtx, &models.Cluster{
		Name:           models.SharedComputeClusterName,
		OrganisationID: org.ID,
		SharedCompute:  true,
		ClusterURL:     s.clusterCfg.ClusterURL,
		ClusterCAData:  s.clusterCfg.ClusterCAData,
		Token:          s.clusterCfg.Token,
	})
	if cErr != nil {
		return fmt.Errorf("failed to upsert shared-compute cluster: %w", cErr)
	}

	if s.bootstrapCfg.PlatformTLSEnabled {
		if err := s.clusterService.InternalEnsurePlatformWildcardTLS(sysCtx, cluster, s.bootstrapCfg); err != nil {
			return fmt.Errorf("failed to create or update platform wildcard TLS: %w", err)
		}
	}

	return nil
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
