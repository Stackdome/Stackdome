package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/computequota"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/go-logr/stdr"
)

const (
	CloudTestBaseDomain     = "sd.localhost"
	CloudTestStorageClass   = "standard"
	cloudTestMaxStorageSize = "2Gi"
)

// SetupCloud starts the integration environment with the real Stackdome Cloud
// runtime wiring and deliberately small limits. It is separate from Setup so
// the self-hosted suite remains unchanged.
func SetupCloud(env *Environment, ctx context.Context) (retErr error) {
	logger := stdr.New(log.New(os.Stderr, "", log.LstdFlags))
	logger.Info("Starting Stackdome Cloud integration test bootstrap")
	env.logger = logger

	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error(fmt.Errorf("panic during bootstrap: %v", recovered), "Bootstrap failed with panic, attempting cleanup")
			env.Cleanup()
			panic(recovered)
		}
		if retErr != nil {
			logger.Error(retErr, "Bootstrap failed, cleaning up")
			env.Cleanup()
		}
	}()

	dbManager := NewDatabaseManager()
	env.Database = dbManager
	if err := dbManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("database bootstrap failed: %w", err)
	}

	clusterManager := NewClusterManager()
	env.clusterManager = clusterManager
	if err := clusterManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("cluster bootstrap failed: %w", err)
	}
	env.Cluster = clusterManager.GetCluster()

	if err := exportSharedComputeProvisioningEnv(ctx, env.Cluster); err != nil {
		return fmt.Errorf("failed to export shared-compute provisioning env: %w", err)
	}
	if err := os.Setenv(config.EnvPlatformBaseDomain.Name, CloudTestBaseDomain); err != nil {
		return fmt.Errorf("failed to set Cloud test platform base domain: %w", err)
	}

	serverManager := NewServerManager(dbManager.GetSessionFactory(), dbManager.GetConfig(), logger)
	serverManager.config.RuntimeMode = config.RuntimeModeStackdomeCloud
	serverManager.config.ComputeMode = config.ComputeModeShared
	serverManager.config.StackdomeCloud = cloudTestConfig(serverManager.config.Server.Hostname)
	env.serverManager = serverManager
	if err := serverManager.Bootstrap(ctx); err != nil {
		return fmt.Errorf("server bootstrap failed: %w", err)
	}

	var sharedComputeCluster models.Cluster
	if err := dbManager.GetSessionFactory().New(ctx).
		Where("shared_compute = ?", true).
		First(&sharedComputeCluster).Error; err != nil {
		return fmt.Errorf("failed to resolve shared-compute cluster: %w", err)
	}

	clientManager := NewClientManager(serverManager.GetBaseURL(), logger)
	env.clientManager = clientManager
	if err := clientManager.Bootstrap(ctx, sharedComputeCluster.ID, dbManager.GetSessionFactory()); err != nil {
		return fmt.Errorf("client bootstrap failed: %w", err)
	}

	env.Client = clientManager.GetClient()
	env.ClusterID = clientManager.GetClusterID()
	env.OrgID = clientManager.GetOrgID()
	env.UserToken = clientManager.GetUserToken()
	env.RegistryName = clientManager.GetRegistryName()

	logger.Info("Stackdome Cloud integration environment initialized",
		"orgID", env.OrgID,
		"clusterID", env.ClusterID,
		"serverURL", serverManager.GetBaseURL())
	return nil
}

func cloudTestConfig(expectedHostname string) *config.StackdomeCloudConfig {
	return &config.StackdomeCloudConfig{
		Access: config.StackdomeCloudComputeAccessConfig{
			MaxActiveSharedComputeLeases: 1,
			TrialEntitlementDuration:     config.ConfigDuration(6 * time.Hour),
		},
		Limits: computequota.ComputeLimits{
			MaxStacksPerOrganization:         2,
			MaxStackResourcesPerOrganization: 3,
			ReplicasPerStackResource:         1,
			MaxVolumesPerOrganization:        2,
			MaxVolumeSize:                    cloudTestMaxStorageSize,
			VolumeStorageClass:               CloudTestStorageClass,
			MaxPostgresAddonsPerOrganization: 1,
			PostgresInstances:                1,
			MaxPostgresStorageSize:           cloudTestMaxStorageSize,
			PostgresCPURequest:               "50m",
			PostgresCPULimit:                 "500m",
			PostgresMemoryRequest:            "128Mi",
			PostgresMemoryLimit:              "1Gi",
			ConcurrentBuilds:                 1,
		},
		Registry: config.StackdomeCloudRegistryConfig{
			MaxActiveRegistries: 1,
			StorageClass:        CloudTestStorageClass,
			StorageSize:         cloudTestMaxStorageSize,
		},
		Features: config.StackdomeCloudFeaturesConfig{},
		Signup: config.StackdomeCloudSignupConfig{
			ClientIPSource: config.StackdomeCloudClientIPSourceRemoteAddr,
			Turnstile: config.StackdomeCloudTurnstileConfig{
				Enabled:             true,
				SiteKey:             "cloud-e2e-site-key",
				ExpectedHostname:    expectedHostname,
				ExpectedAction:      "signup",
				VerificationTimeout: config.ConfigDuration(5 * time.Second),
			},
			Throttle: config.StackdomeCloudThrottleConfig{
				IP: config.StackdomeCloudIPThrottleConfig{
					MaxTrackedClients: 100,
					MaxAttempts:       20,
					Window:            config.ConfigDuration(time.Minute),
				},
				Email: config.StackdomeCloudEmailThrottleConfig{
					MaxTrackedAddresses: 100,
					MaxAttempts:         20,
					Window:              config.ConfigDuration(time.Minute),
				},
			},
		},
	}
}
