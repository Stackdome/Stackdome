package testutil

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/mt-sre/devkube/dev"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

	// Import all CRD schemes from cluster-agent dependency
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"
	usersv1alpha1 "stackdome.io/cluster-agent/api/users/v1alpha1"
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

// Setup creates and initializes the test cluster
func (tc *TestCluster) Setup(ctx context.Context) error {
	tc.logger.Info("Setting up test cluster", "name", tc.config.Name)

	// Prepare cluster initializers
	var clusterInitializers []dev.ClusterInitializer

	// Always load CRDs from GitHub
	clusterInitializers = append(clusterInitializers, &githubCRDLoader{
		version:  GetClusterAgentVersion(),
		cacheDir: tc.config.CacheDir,
		client:   createGitHubClient(ctx),
	})

	// Optionally install operators
	if tc.config.InstallOperators {
		clusterInitializers = append(clusterInitializers, tc.getClusterDependencies()...)
	}

	clusterInitializers = append(clusterInitializers,
		HTTPObjectApplier("https://github.com/cloudnative-pg/plugin-barman-cloud/releases/download/v0.5.0/manifest.yaml"),
	)

	// Configure Kind cluster
	kindConfig := kindv1alpha4.Cluster{
		Nodes: tc.getNodeConfig(),
	}

	// Detect container runtime if set to auto
	containerRuntime := tc.config.ContainerRuntime
	if containerRuntime == "auto" {
		cr, err := dev.DetectContainerRuntime()
		if err != nil {
			return fmt.Errorf("detecting container runtime: %w", err)
		}
		containerRuntime = string(cr)
		tc.logger.Info("Detected container runtime", "runtime", containerRuntime)
	}

	// Create development environment
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

	// Initialize the cluster
	if err := tc.environment.Init(ctx); err != nil {
		return fmt.Errorf("initializing test environment: %w", err)
	}

	tc.logger.Info("Test cluster setup complete", "name", tc.config.Name)
	return nil
}

// GetClient returns a Kubernetes client for the test cluster
func (tc *TestCluster) GetClient() client.Client {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return nil
	}
	// Create client from REST config
	config := tc.environment.Cluster.RestConfig
	if config == nil {
		return nil
	}

	scheme := runtime.NewScheme()
	schemeBuilder := tc.getSchemeBuilder()
	if err := schemeBuilder.AddToScheme(scheme); err != nil {
		tc.logger.Error(err, "Failed to add schemes")
		return nil
	}

	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		tc.logger.Error(err, "Failed to create client")
		return nil
	}
	return c
}

// GetRESTConfig returns the REST config for the test cluster
func (tc *TestCluster) GetRESTConfig() *rest.Config {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return nil
	}
	return tc.environment.Cluster.RestConfig
}

// GetKubeClient returns a kubernetes.Interface client for the test cluster
func (tc *TestCluster) GetKubeClient() (interface{}, error) {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return nil, fmt.Errorf("cluster not initialized")
	}

	config := tc.environment.Cluster.RestConfig
	if config == nil {
		return nil, fmt.Errorf("rest config not available")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// GetKubeConfig returns the REST config for the test cluster (alias for GetRESTConfig)
func (tc *TestCluster) GetKubeConfig() (*rest.Config, error) {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return nil, fmt.Errorf("cluster not initialized")
	}
	return tc.environment.Cluster.RestConfig, nil
}

// LoadImage loads a container image into the cluster
func (tc *TestCluster) LoadImage(imagePath string) error {
	if tc.environment == nil {
		return fmt.Errorf("cluster not initialized")
	}
	return tc.environment.LoadImageFromTar(imagePath)
}

// DeployClusterAgent deploys the cluster agent to the test cluster
func (tc *TestCluster) DeployClusterAgent(ctx context.Context, imageTag string) error {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return fmt.Errorf("cluster not initialized")
	}

	cluster := tc.environment.Cluster

	// Deploy namespace and RBAC from GitHub
	version := GetClusterAgentVersion()

	// Create GitHub client to verify tag exists
	githubClient := createGitHubClient(ctx)
	exists, err := checkTagExists(ctx, githubClient, version)
	if err != nil {
		return fmt.Errorf("checking if tag %s exists: %w", version, err)
	}
	if !exists {
		// Try to find the latest available tag
		latestTag, err := findLatestTag(ctx, githubClient)
		if err != nil {
			return fmt.Errorf("tag %s not found and unable to find latest tag: %w", version, err)
		}
		return fmt.Errorf("tag %s not found in repository, latest available tag is %s. Please update the version in go.mod or set CLUSTER_AGENT_VERSION environment variable", version, latestTag)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := deployManifestsFromGitHub(ctxWithTimeout, cluster, []string{
		"00-namespace.yaml",
		"01-rbac.yaml",
	}, version, tc.config.CacheDir); err != nil {
		return fmt.Errorf("deploying cluster agent dependencies: %w", err)
	}

	// Wait a moment for namespace to be ready
	time.Sleep(5 * time.Second)

	// Deploy the cluster agent manager
	deployment, err := getClusterAgentDeployment(imageTag)
	if err != nil {
		return fmt.Errorf("creating deployment: %w", err)
	}

	if err := cluster.CreateAndWaitForReadiness(ctx, deployment); err != nil {
		return fmt.Errorf("deploying cluster agent manager: %w", err)
	}

	tc.logger.Info("Cluster agent deployed", "image", imageTag)
	return nil
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
		usersv1alpha1.AddToScheme,
		clientgoscheme.AddToScheme,
	}
}

func (tc *TestCluster) getClusterDependencies() []dev.ClusterInitializer {
	return []dev.ClusterInitializer{
		dev.ClusterHelmInstall{
			RepoName:    "traefik",
			RepoURL:     "https://traefik.github.io/charts",
			PackageName: "traefik",
			Namespace:   "traefik-v2",
			ReleaseName: "traefik",
		},
		dev.ClusterHelmInstall{
			RepoName:    "cnpg",
			RepoURL:     "https://cloudnative-pg.github.io/charts",
			PackageName: "cloudnative-pg",
			Namespace:   "cnpg-system",
			ReleaseName: "cnpg",
		},
		dev.ClusterHelmInstall{
			RepoName:    "jetstack",
			RepoURL:     "https://charts.jetstack.io",
			PackageName: "cert-manager",
			Namespace:   "cert-manager",
			ReleaseName: "cert-manager",
			SetVars: []string{
				"crds.enabled=true",
			},
		},
	}
}
