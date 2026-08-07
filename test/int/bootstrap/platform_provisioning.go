package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/testutil"
)

// PlatformProvisioningBaseDomain is the platform fallback domain used by the
// integration environment for newly allocated public hostnames.
const PlatformProvisioningBaseDomain = "sd.test"

const (
	platformProvisioningEmail           = "ops@stackdome.io"
	platformProvisioningCloudflareToken = "integration-dummy-cloudflare-token"
	platformProvisioningRegistrySize    = "10Gi"
)

// Suite tenant org: signed up by the client bootstrap; signup seeds its registry
// on the platform cluster via the read-time fallback.
const (
	suiteOrgName      = "Integration Tests"
	suiteUserName     = "Integration Tester"
	suiteUserEmail    = "int-tests@stackdome.io"
	suiteUserPassword = "int-welcome@123"
)

// exportPlatformProvisioningEnv provisions API-server credentials on the test
// cluster and exports the PLATFORM_* environment variables, so that when the
// server environment initializes it creates the Platform=true cluster, platform
// org, and wildcard TLS resources via pkg/bootstrap.
func exportPlatformProvisioningEnv(ctx context.Context, cluster *testutil.TestCluster) error {
	if err := deployAPIServerServiceAccount(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy api-server service account: %w", err)
	}

	clusterURL, caData, saToken, err := ExtractAPIServerClusterCredentials(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to extract cluster credentials: %w", err)
	}

	envs := map[string]string{
		config.EnvPlatformClusterAPIURL.Name:          clusterURL,
		config.EnvPlatformClusterCAData.Name:          caData,
		config.EnvPlatformClusterToken.Name:           saToken,
		config.EnvPlatformBaseDomain.Name:             PlatformProvisioningBaseDomain,
		config.EnvPlatformEmail.Name:                  platformProvisioningEmail,
		config.EnvPlatformDNSCloudflareAPIToken.Name:  platformProvisioningCloudflareToken,
		config.EnvPlatformACMEEnvironment.Name:        config.ACMEEnvironmentStaging,
		config.EnvPlatformOrgRegistryStorageSize.Name: platformProvisioningRegistrySize,
	}
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}
