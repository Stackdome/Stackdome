package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/testutil"
	"github.com/Stackdome/stackdome/test/int/shared"
)

// PlatformProvisioningBaseDomain is the platform base domain the integration
// bootstrap seeds. Per-org subdomains derive as <slug>.<PlatformProvisioningBaseDomain>.
// Kept distinct from the stack-test organisation domain so the globally-unique
// organisation_domains.domain constraint never collides at boot.
const PlatformProvisioningBaseDomain = "sd.test"

const (
	platformProvisioningClusterName   = "platform-cluster"
	platformProvisioningAdminEmail    = "platform-admin@stackdome.io"
	platformProvisioningAdminName     = "Platform Admin"
	platformProvisioningAdminPassword = "platform-welcome@123"
	platformProvisioningRegistrySize  = "10Gi"
)

// exportPlatformProvisioningEnv provisions API-server credentials on the test
// cluster and exports the PLATFORM_* environment variables, so that when the
// server environment initializes it seeds a Platform=true cluster, platform org
// and base domain via pkg/bootstrap — mirroring a real env-bootstrapped install.
func exportPlatformProvisioningEnv(ctx context.Context, cluster *testutil.TestCluster) error {
	if err := deployAPIServerServiceAccount(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy api-server service account: %w", err)
	}

	clusterURL, caData, saToken, err := extractAPIServerClusterCredentials(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to extract cluster credentials: %w", err)
	}

	envs := map[string]string{
		config.EnvPlatformClusterName.Name:         platformProvisioningClusterName,
		config.EnvPlatformClusterAPIURL.Name:       clusterURL,
		config.EnvPlatformClusterCAData.Name:       caData,
		config.EnvPlatformClusterToken.Name:        saToken,
		config.EnvPlatformBaseDomain.Name:          PlatformProvisioningBaseDomain,
		config.EnvPlatformAdminEmail.Name:          platformProvisioningAdminEmail,
		config.EnvPlatformAdminName.Name:           platformProvisioningAdminName,
		config.EnvPlatformAdminPassword.Name:       platformProvisioningAdminPassword,
		config.EnvPlatformRegistryStorageSize.Name: platformProvisioningRegistrySize,
		config.EnvPlatformClusterRegistryName.Name: shared.TestRegistryName,
	}
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}
