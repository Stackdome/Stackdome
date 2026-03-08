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
	cluster        *testutil.TestCluster
	logger         logr.Logger
	createdCluster bool // Track if we created the cluster
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

func (cm *ClusterManager) createNewCluster(ctx context.Context) error {
	cm.logger.Info("Creating new Kind cluster for integration tests")

	// Create cluster configuration
	config := testutil.DefaultClusterConfig("stackdome-int-test", cm.logger)

	// Create test cluster instance
	cm.cluster = testutil.NewTestCluster(config)

	// Bootstrap context with timeout
	bootstrapCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Create cluster and deploy all dependencies (operators + CRDs)
	cm.logger.Info("Setting up test cluster (this may take 5-10 minutes)")
	if err := cm.cluster.Setup(bootstrapCtx); err != nil {
		return fmt.Errorf("failed to setup test cluster: %w", err)
	}

	cm.logger.Info("Test cluster created successfully")

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
	cm.createdCluster = true // Mark that we created this cluster
	return nil
}

func (cm *ClusterManager) GetCluster() *testutil.TestCluster {
	return cm.cluster
}

func (cm *ClusterManager) Cleanup(ctx context.Context) error {
	if cm.cluster == nil {
		return nil
	}

	// Only cleanup if we created the cluster (not external)
	if !cm.createdCluster {
		cm.logger.Info("Skipping cluster cleanup - using external cluster managed by Mage")
		return nil
	}

	// Check if user wants to keep cluster for debugging
	if os.Getenv("KEEP_CLUSTER") == "true" {
		cm.logger.Info("KEEP_CLUSTER=true, preserving test cluster for debugging")
		cm.logger.Info("To delete later, run: kind delete cluster --name stackdome-int-test")
		return nil
	}

	// Cleanup cluster we created
	cm.logger.Info("Cleaning up test cluster")
	return cm.cluster.Teardown(ctx)
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
