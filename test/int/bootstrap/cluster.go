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
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

func (cm *ClusterManager) useExistingCluster(ctx context.Context, kubeconfig string) error {
	// Verify kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file does not exist at %s - ensure cluster is created. Run 'cd /Users/asnaraya/projects/skysync/cluster-agent && ./mage Dev.Deploy' first", kubeconfig)
	}

	cm.logger.Info("Kubeconfig file found, connecting to external cluster")

	// Load REST config from kubeconfig
	restConfig, err := cm.loadKubeconfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Create TestCluster wrapper pointing to external cluster
	// We create the config but don't call Setup() since cluster already exists
	config := testutil.DefaultClusterConfig("stackdome-int-test-external", cm.logger)
	cm.cluster = testutil.NewTestCluster(config)

	// Manually set the REST config to point to external cluster
	// Note: This requires accessing internal fields - we'll verify connectivity instead
	cm.logger.Info("Verifying external cluster connectivity")

	// Create a temporary client to verify connectivity
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Verify cluster is accessible
	_, err = kubeClient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to external cluster - ensure cluster is running: kubectl cluster-info. Error: %w", err)
	}

	cm.logger.Info("Successfully connected to external cluster")

	// For now, assume cluster agent is already deployed by mage
	// In the future, we could add verification logic here
	cm.logger.Info("Assuming cluster agent is deployed (deployed by mage Dev.Deploy)")

	return nil
}

func (cm *ClusterManager) loadKubeconfig(kubeconfig string) (*rest.Config, error) {
	// Use client-go to load kubeconfig from file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig

	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	return kubeConfig.ClientConfig()
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
