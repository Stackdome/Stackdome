package bootstrap

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/testutil"
)

// SharedComputeProvisioningBaseDomain is the platform routing domain used by
// the integration environment for newly allocated public hostnames.
const SharedComputeProvisioningBaseDomain = "sd.test"

const (
	sharedComputeProvisioningEmail           = "ops@stackdome.io"
	sharedComputeProvisioningCloudflareToken = "integration-dummy-cloudflare-token"
	sharedComputeProvisioningRegistrySize    = "10Gi"
)

// Suite tenant org: signed up by the client bootstrap; signup seeds its registry
// on the shared-compute cluster via the read-time fallback.
const (
	suiteOrgName      = "Integration Tests"
	suiteUserName     = "Integration Tester"
	suiteUserEmail    = "int-tests@stackdome.io"
	suiteUserPassword = "int-welcome@123"
)

// exportSharedComputeProvisioningEnv provisions API-server credentials on the
// test cluster and exports shared-compute and platform-routing environment
// variables. When the server environment initializes, it creates the
// shared-compute cluster, platform org, and wildcard TLS resources via
// pkg/bootstrap.
func exportSharedComputeProvisioningEnv(ctx context.Context, cluster *testutil.TestCluster) error {
	if err := deployAPIServerServiceAccount(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy api-server service account: %w", err)
	}

	clusterURL, caData, saToken, err := ExtractAPIServerClusterCredentials(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to extract cluster credentials: %w", err)
	}

	envs := map[string]string{
		config.EnvSharedComputeClusterAPIURL.Name:     clusterURL,
		config.EnvSharedComputeClusterCAData.Name:     caData,
		config.EnvSharedComputeClusterToken.Name:      saToken,
		config.EnvPlatformBaseDomain.Name:             SharedComputeProvisioningBaseDomain,
		config.EnvPlatformTLSEnabled.Name:             strconv.FormatBool(true),
		config.EnvPlatformEmail.Name:                  sharedComputeProvisioningEmail,
		config.EnvPlatformDNSCloudflareAPIToken.Name:  sharedComputeProvisioningCloudflareToken,
		config.EnvPlatformACMEEnvironment.Name:        config.ACMEEnvironmentStaging,
		config.EnvPlatformOrgRegistryStorageSize.Name: sharedComputeProvisioningRegistrySize,
	}
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}
