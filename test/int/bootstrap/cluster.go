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
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return fmt.Errorf("kubeconfig file does not exist at %s\n\nTo create cluster, run:\n  mage cluster:setup\n\nOr unset TEST_KUBECONFIG to create temporary cluster", kubeconfig)
	}

	cm.logger.Info("Kubeconfig file found, connecting to external cluster", "path", kubeconfig)

	cluster, err := testutil.NewTestClusterFromKubeconfig(kubeconfig, cm.logger)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Verify cluster connectivity
	kubeClient, err := cluster.GetKubeClient()
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	version, err := kubeClient.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to connect to external cluster\n\nCheck if cluster is running:\n  kubectl cluster-info --kubeconfig=%s\n\nError: %w", kubeconfig, err)
	}

	cm.logger.Info("Successfully connected to external cluster", "kubernetes-version", version.String())

	// Verify cluster agent is deployed
	deploymentsClient := kubeClient.AppsV1().Deployments("stackdome-system")
	deployment, err := deploymentsClient.Get(ctx, "cluster-agent-manager", metav1.GetOptions{})
	if err != nil {
		cm.logger.Info("Cluster agent not found - this is OK if it will be deployed by mage", "error", err.Error())
	} else {
		cm.logger.Info("Found cluster agent deployment", "replicas", deployment.Status.ReadyReplicas)
	}

	cm.cluster = cluster
	cm.logger.Info("External cluster verification complete")
	return nil
}

func (cm *ClusterManager) createNewCluster(ctx context.Context) error {
	cm.logger.Info("Creating new Kind cluster for integration tests", "name", testClusterName)

	if err := cm.deleteClusterIfExists(ctx); err != nil {
		cm.logger.Info("Warning: failed to delete existing cluster", "error", err.Error())
	}

	config := testutil.DefaultClusterConfig(testClusterName, cm.logger)
	tc := testutil.NewTestCluster(config)

	bootstrapCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cm.logger.Info("Setting up test cluster with helm chart (this may take 5-10 minutes)")
	if err := tc.Setup(bootstrapCtx); err != nil {
		return fmt.Errorf("failed to setup test cluster: %w", err)
	}

	// Get kubeconfig from kind rather than relying on devkube's internal paths
	provider := kindcluster.NewProvider()
	kubeconfigData, err := provider.KubeConfig(testClusterName, false)
	if err != nil {
		return fmt.Errorf("getting kubeconfig from kind: %w", err)
	}

	restCfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfigData))
	if err != nil {
		return fmt.Errorf("parsing kubeconfig: %w", err)
	}

	cm.cluster = testutil.NewTestClusterFromRESTConfig(restCfg, cm.logger)
	cm.createdCluster = true
	cm.logger.Info("Cluster bootstrap completed successfully")
	return nil
}

func (cm *ClusterManager) GetCluster() *testutil.TestCluster {
	return cm.cluster
}

func (cm *ClusterManager) Cleanup(ctx context.Context) error {
	if !cm.createdCluster {
		cm.logger.Info("Skipping cluster cleanup - using external cluster")
		return nil
	}

	if os.Getenv("KEEP_CLUSTER") == "true" {
		cm.logger.Info("KEEP_CLUSTER=true, preserving test cluster for debugging")
		cm.logger.Info("To delete later, run: kind delete cluster --name", "name", testClusterName)
		return nil
	}

	cm.logger.Info("Cleaning up test cluster", "name", testClusterName)
	provider := kindcluster.NewProvider()
	return provider.Delete(testClusterName, "")
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
