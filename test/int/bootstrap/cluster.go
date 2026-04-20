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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"
)

const (
	testClusterName = "stackdome-api-server-test"
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
		return fmt.Errorf("kubeconfig file does not exist at %s\n\nTo create cluster, run:\n  cd /Users/asnaraya/projects/skysync/cluster-agent\n  ./mage Dev.Deploy\n\nOr unset TEST_KUBECONFIG to create temporary cluster", kubeconfig)
	}

	cm.logger.Info("Kubeconfig file found, connecting to external cluster", "path", kubeconfig)

	// Load REST config from kubeconfig to verify cluster is accessible
	restConfig, err := cm.loadKubeconfig(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Verify cluster connectivity
	kubeClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	version, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to external cluster\n\nCheck if cluster is running:\n  kubectl cluster-info --kubeconfig=%s\n\nError: %w", kubeconfig, err)
	}

	cm.logger.Info("Successfully connected to external cluster", "kubernetes-version", version.String())

	// Create TestCluster using testutil but pointing to external cluster
	// We'll use a special config that references the existing kubeconfig
	config := &testutil.ClusterConfig{
		Name:             "external-cluster",
		CacheDir:         ".cache/test-clusters",
		NodeCount:        1,     // Not used for external cluster
		InstallOperators: false, // Don't install, assume already there
		ContainerRuntime: "docker",
		Logger:           cm.logger,
	}

	cm.cluster = testutil.NewTestCluster(config)

	// Verify cluster agent is deployed
	cm.logger.Info("Verifying cluster agent deployment")

	// Check if cluster-agent-manager deployment exists in stackdome-system namespace
	deploymentsClient := kubeClient.AppsV1().Deployments("stackdome-system")
	deployment, err := deploymentsClient.Get(ctx, "cluster-agent-manager", metav1.GetOptions{})
	if err != nil {
		cm.logger.Info("Cluster agent not found - this is OK if it will be deployed by mage", "error", err.Error())
		// Don't fail here - assume mage will deploy it or it's already there with different name
	} else {
		cm.logger.Info("Found cluster agent deployment", "replicas", deployment.Status.ReadyReplicas)
	}

	cm.logger.Info("External cluster verification complete")
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

	// Delete existing cluster if it exists to ensure clean state
	cm.logger.Info("Checking for existing cluster", "name", testClusterName)
	if err := cm.deleteClusterIfExists(ctx); err != nil {
		cm.logger.Info("Warning: failed to delete existing cluster", "error", err.Error())
		// Continue anyway - cluster might not exist
	}

	// Create cluster configuration
	config := testutil.DefaultClusterConfig(testClusterName, cm.logger)

	// Create test cluster instance
	cm.cluster = testutil.NewTestCluster(config)

	// Bootstrap context with timeout
	bootstrapCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Clear the cache directory
	if err := os.RemoveAll(config.CacheDir); err != nil {
		cm.logger.Info("Warning: failed to clear cache directory", "path", config.CacheDir, "error", err.Error())
	}
	cm.logger.Info("Cache directory cleared", "path", config.CacheDir)

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
		cm.logger.Info("To delete later, run: kind delete cluster --name", "name", testClusterName)
		return nil
	}

	// Cleanup cluster we created
	cm.logger.Info("Cleaning up test cluster")
	return cm.cluster.Teardown(ctx)
}

func (cm *ClusterManager) deleteClusterIfExists(ctx context.Context) error {
	provider := kindcluster.NewProvider()
	clusters, err := provider.List()
	if err != nil {
		return fmt.Errorf("listing kind clusters: %w", err)
	}

	for _, name := range clusters {
		if name == testClusterName {
			cm.logger.Info("Deleting existing Kind cluster", "name", testClusterName)
			if err := provider.Delete(testClusterName, ""); err != nil {
				return fmt.Errorf("deleting kind cluster: %w", err)
			}
			cm.logger.Info("Existing cluster deleted")
			return nil
		}
	}

	cm.logger.Info("No existing cluster found", "name", testClusterName)
	return nil
}

func getClusterAgentImageTag() string {
	return testutil.GetClusterAgentVersion()
}
