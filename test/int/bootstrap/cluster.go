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

	// Check for external cluster first
	if kubeconfig := os.Getenv("TEST_KUBECONFIG"); kubeconfig != "" {
		cm.logger.Info("TEST_KUBECONFIG detected, using external cluster", "kubeconfig", kubeconfig)
		return cm.useExistingCluster(ctx, kubeconfig)
	}

	// Fall back to creating new cluster
	cm.logger.Info("TEST_KUBECONFIG not set, creating new test cluster")
	return cm.createNewCluster(ctx)
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
