package bootstrap

import (
	"context"
	"fmt"
	"os"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/testutil"
	"github.com/Stackdome/stackdome/test/int/shared"
)

// DefaultProvisioningBaseDomain is the platform base domain the integration
// bootstrap seeds. Per-org subdomains derive as <slug>.<DefaultProvisioningBaseDomain>.
// Kept distinct from the stack-test organisation domain so the globally-unique
// organisation_domains.domain constraint never collides at boot.
const DefaultProvisioningBaseDomain = "sd.test"

const (
	defaultProvisioningClusterName   = "default-cluster"
	defaultProvisioningAdminEmail    = "platform-admin@stackdome.io"
	defaultProvisioningAdminName     = "Platform Admin"
	defaultProvisioningAdminPassword = "platform-welcome@123"
	defaultProvisioningRegistrySize  = "10Gi"
)

// exportDefaultProvisioningEnv provisions API-server credentials on the test
// cluster and exports the DEFAULT_* environment variables, so that when the
// server environment initializes it seeds a Default=true cluster, platform org
// and base domain via pkg/bootstrap — mirroring a real env-bootstrapped install.
func exportDefaultProvisioningEnv(ctx context.Context, cluster *testutil.TestCluster) error {
	if err := deployAPIServerServiceAccount(ctx, cluster); err != nil {
		return fmt.Errorf("failed to deploy api-server service account: %w", err)
	}

	clusterURL, caData, saToken, err := extractAPIServerClusterCredentials(ctx, cluster)
	if err != nil {
		return fmt.Errorf("failed to extract cluster credentials: %w", err)
	}

	envs := map[string]string{
		config.EnvDefaultClusterName.Name:          defaultProvisioningClusterName,
		config.EnvDefaultClusterAPIURL.Name:        clusterURL,
		config.EnvDefaultClusterCAData.Name:        caData,
		config.EnvDefaultClusterToken.Name:         saToken,
		config.EnvDefaultBaseDomain.Name:           DefaultProvisioningBaseDomain,
		config.EnvDefaultUserEmail.Name:            defaultProvisioningAdminEmail,
		config.EnvDefaultUserName.Name:             defaultProvisioningAdminName,
		config.EnvDefaultUserPassword.Name:         defaultProvisioningAdminPassword,
		config.EnvDefaultRegistryStorageSize.Name: defaultProvisioningRegistrySize,
		config.EnvDefaultClusterRegistryName.Name: shared.TestRegistryName,
	}
	for name, value := range envs {
		if err := os.Setenv(name, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}
