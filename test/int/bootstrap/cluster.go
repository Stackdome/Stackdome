package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/testutil"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

type ClusterManager struct {
	cluster *testutil.TestCluster
	logger  logr.Logger
}

func NewClusterManager() *ClusterManager {
	logger := stdr.New(log.New(os.Stdout, "", log.LstdFlags))

	return &ClusterManager{
		logger: logger,
	}
}

func NewClusterManagerWithLogger(logger logr.Logger) *ClusterManager {
	return &ClusterManager{
		logger: logger,
	}
}

func (cm *ClusterManager) Bootstrap(ctx context.Context) error {
	cm.logger.Info("Starting cluster bootstrap")

	// Get kubeconfig from environment - this is required
	kubeconfig := os.Getenv("TEST_KUBECONFIG")
	if kubeconfig == "" {
		return fmt.Errorf("TEST_KUBECONFIG not set - cluster must be created before running tests. Run 'mage cluster:create' first")
	}

	// Verify the kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file does not exist at %s - ensure cluster is created. Run 'mage cluster:create' first", kubeconfig)
	}

	cm.logger.Info("Using Mage-managed test cluster", "kubeconfig", kubeconfig)

	// Create cluster configuration using DefaultClusterConfig
	config := testutil.DefaultClusterConfig("stackdome-int-test", cm.logger)

	// Create test cluster instance
	cm.cluster = testutil.NewTestCluster(config)

	// Deploy CRDs and cluster agent
	bootstrapCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Verify cluster is accessible
	cm.logger.Info("Verifying cluster accessibility")
	if _, err := cm.cluster.GetKubeClient(); err != nil {
		return fmt.Errorf("failed to connect to cluster - ensure cluster is running: %w", err)
	}

	// Deploy cluster agent
	cm.logger.Info("Deploying cluster agent")
	imageTag := getClusterAgentImageTag()
	if err := cm.cluster.DeployClusterAgent(bootstrapCtx, imageTag); err != nil {
		return fmt.Errorf("failed to deploy cluster agent: %w", err)
	}

	// Wait for cluster agent to be ready
	cm.logger.Info("Waiting for cluster agent to be ready")
	if err := cm.waitForClusterAgentReady(bootstrapCtx); err != nil {
		return fmt.Errorf("cluster agent not ready: %w", err)
	}

	cm.logger.Info("Cluster bootstrap completed successfully")
	return nil
}

func (cm *ClusterManager) GetCluster() *testutil.TestCluster {
	return cm.cluster
}

func (cm *ClusterManager) Cleanup(ctx context.Context) error {
	if cm.cluster != nil {
		// Always skip cluster teardown - cluster is managed by Mage
		cm.logger.Info("Skipping cluster teardown - cluster is managed by Mage. Use 'mage cluster:delete' to remove cluster")

		// We could optionally clean up deployed test resources here
		// but for now, let the Mage cluster management handle everything
	}
	return nil
}

func (cm *ClusterManager) waitForClusterAgentReady(ctx context.Context) error {
	// Get kubernetes client
	kubeClient, err := cm.cluster.GetKubeClient()
	if err != nil {
		return fmt.Errorf("failed to get kube client: %w", err)
	}

	// TODO: Add proper readiness check for cluster agent deployment
	// For now, just wait a bit to ensure the deployment is ready
	time.Sleep(30 * time.Second)

	cm.logger.Info("Cluster agent is ready", "client", kubeClient != nil)
	return nil
}

func getClusterAgentImageTag() string {
	return testutil.GetClusterAgentVersion()
}
