package testutil

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/mt-sre/devkube/dev"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	// Import all CRD schemes from cluster-agent dependency
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
)

// ClusterConfig holds configuration for test cluster setup
type ClusterConfig struct {
	Name             string
	CacheDir         string
	NodeCount        int
	InstallOperators bool
	ContainerRuntime string
	Logger           logr.Logger
}

// DefaultClusterConfig returns a default configuration for test clusters
func DefaultClusterConfig(name string, logger logr.Logger) *ClusterConfig {
	return &ClusterConfig{
		Name:             name,
		CacheDir:         ".cache/test-clusters",
		NodeCount:        2, // 1 control plane + 1 worker
		InstallOperators: true,
		ContainerRuntime: "docker",
		Logger:           logger,
	}
}

// TestCluster manages a test Kubernetes cluster
type TestCluster struct {
	config      *ClusterConfig
	environment *dev.Environment
	logger      logr.Logger
	restConfig  *rest.Config
}

// NewTestCluster creates a new test cluster manager
func NewTestCluster(config *ClusterConfig) *TestCluster {
	// Use provided logger or create a default one that outputs to stdout
	if config.Logger.GetSink() == nil {
		config.Logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	}
	ctrl.SetLogger(config.Logger)

	return &TestCluster{
		config: config,
		logger: config.Logger,
	}
}

// NewTestClusterFromKubeconfig creates a TestCluster backed by an existing kubeconfig file.
func NewTestClusterFromKubeconfig(kubeconfigPath string, logger logr.Logger) (*TestCluster, error) {
	if logger.GetSink() == nil {
		logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	}

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig %s: %w", kubeconfigPath, err)
	}

	return NewTestClusterFromRESTConfig(restCfg, logger), nil
}

// NewTestClusterFromRESTConfig creates a TestCluster from an existing REST config.
func NewTestClusterFromRESTConfig(cfg *rest.Config, logger logr.Logger) *TestCluster {
	if logger.GetSink() == nil {
		logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	}

	return &TestCluster{
		config: &ClusterConfig{
			Logger: logger,
		},
		logger:     logger,
		restConfig: cfg,
	}
}

// Setup creates and initializes the test cluster
func (tc *TestCluster) Setup(ctx context.Context) error {
	tc.logger.Info("Setting up test cluster", "name", tc.config.Name)

	helmBin, err := EnsureHelm(ctx)
	if err != nil {
		return fmt.Errorf("ensuring helm is available: %w", err)
	}
	tc.logger.Info("Using helm", "path", helmBin)

	chartVersion := GetChartVersion()
	tc.logger.Info("Using stackdome-agent chart", "version", chartVersion)

	clusterInitializers := []dev.ClusterInitializer{
		dev.ClusterInitFn(func(ctx context.Context, cluster *dev.Cluster) error {
			args := []string{
				"upgrade", "--install", ChartReleaseName, ChartRepo,
				"--version", chartVersion,
				"--namespace", ChartNamespace,
				"--create-namespace",
			}

			tc.logger.Info("Installing stackdome-agent chart", "args", args)
			cmd := exec.CommandContext(ctx, helmBin, args...)
			cmd.Env = append(os.Environ(), "KUBECONFIG="+cluster.Kubeconfig())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("installing stackdome-agent chart: %w", err)
			}
			return nil
		}),
	}

	kindConfig := kindv1alpha4.Cluster{
		Nodes: tc.getNodeConfig(),
	}

	containerRuntime := tc.config.ContainerRuntime
	if containerRuntime == "auto" {
		cr, err := dev.DetectContainerRuntime()
		if err != nil {
			return fmt.Errorf("detecting container runtime: %w", err)
		}
		containerRuntime = string(cr)
		tc.logger.Info("Detected container runtime", "runtime", containerRuntime)
	}

	tc.environment = dev.NewEnvironment(
		tc.config.Name,
		path.Join(tc.config.CacheDir, tc.config.Name),
		dev.WithClusterOptions([]dev.ClusterOption{
			dev.WithWaitOptions([]dev.WaitOption{
				dev.WithTimeout(10 * time.Minute),
			}),
			dev.WithSchemeBuilder(tc.getSchemeBuilder()),
		}),
		dev.WithContainerRuntime(containerRuntime),
		dev.WithKindClusterConfig(kindConfig),
		dev.WithClusterInitializers(clusterInitializers),
	)

	if err := tc.environment.Init(ctx); err != nil {
		return fmt.Errorf("initializing test environment: %w", err)
	}

	tc.logger.Info("Test cluster setup complete", "name", tc.config.Name)
	return nil
}

// GetClient returns a Kubernetes client for the test cluster
func (tc *TestCluster) GetClient() client.Client {
	cfg := tc.GetRESTConfig()
	if cfg == nil {
		return nil
	}

	scheme := runtime.NewScheme()
	schemeBuilder := tc.getSchemeBuilder()
	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		tc.logger.Error(err, "Failed to add schemes")
		return nil
	}

	c, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		tc.logger.Error(err, "Failed to create client")
		return nil
	}
	return c
}

// GetRESTConfig returns the REST config for the test cluster
func (tc *TestCluster) GetRESTConfig() *rest.Config {
	if tc.restConfig != nil {
		return tc.restConfig
	}
	if tc.environment == nil || tc.environment.Cluster == nil {
		return nil
	}
	return tc.environment.Cluster.RestConfig
}

// GetKubeClient returns a kubernetes.Interface client for the test cluster
func (tc *TestCluster) GetKubeClient() (*kubernetes.Clientset, error) {
	cfg := tc.GetRESTConfig()
	if cfg == nil {
		return nil, fmt.Errorf("cluster not initialized")
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetKubeConfig returns the REST config for the test cluster (alias for GetRESTConfig)
func (tc *TestCluster) GetKubeConfig() (*rest.Config, error) {
	cfg := tc.GetRESTConfig()
	if cfg == nil {
		return nil, fmt.Errorf("cluster not initialized")
	}
	return cfg, nil
}

// LoadImage loads a container image into the cluster
func (tc *TestCluster) LoadImage(imagePath string) error {
	if tc.environment == nil {
		return fmt.Errorf("cluster not initialized")
	}
	return tc.environment.LoadImageFromTar(imagePath)
}

// Teardown destroys the test cluster
func (tc *TestCluster) Teardown(ctx context.Context) error {
	if tc.environment != nil {
		tc.logger.Info("Tearing down test cluster", "name", tc.config.Name)
		return tc.environment.Destroy(ctx)
	}
	return nil
}

// getNodeConfig returns Kind node configuration based on NodeCount
func (tc *TestCluster) getNodeConfig() []kindv1alpha4.Node {
	nodes := []kindv1alpha4.Node{
		{Role: kindv1alpha4.ControlPlaneRole},
	}

	// Add worker nodes
	for i := 1; i < tc.config.NodeCount; i++ {
		nodes = append(nodes, kindv1alpha4.Node{
			Role: kindv1alpha4.WorkerRole,
		})
	}

	return nodes
}

// getSchemeBuilder returns the scheme builder with all CRDs
func (tc *TestCluster) getSchemeBuilder() runtime.SchemeBuilder {
	return runtime.SchemeBuilder{
		addonsv1alpha1.AddToScheme,
		buildsv1alpha1.AddToScheme,
		corev1alpha1.AddToScheme,
		registryv1alpha1.AddToScheme,
		storagev1alpha1.AddToScheme,
		cmv1.AddToScheme,
		clientgoscheme.AddToScheme,
	}
}
