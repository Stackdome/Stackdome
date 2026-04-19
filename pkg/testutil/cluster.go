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
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/yaml"

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

	// We use GitHubLoader and httpObjectApplier since github loader uses
	// a client which can use gh personal access token for authentication and thus can handle private repositories,
	// while httpObjectApplier is a simple initializer that applies manifests from given URLs and does not require authentication.
	// We use GitHubLoader for the cluster agent since it is a private repository, and httpObjectApplier for the Barman Cloud plugin since it is public.
	// Load cluster agent manifests from GitHub
	clusterAgentInitializer := NewGitHubLoader(
		ctx,
		WithCacheDir(tc.config.CacheDir),
		WithRepoOwner(clusterAgentOwner),
		WithRepoName(clusterAgentRepo),
		WithRepoTag(GetClusterAgentVersion()),
		WithPathsToLoad([]string{
			"config/deploy/00-namespace.yaml",
			"config/deploy/01-rbac.yaml",
			// Trailing slash is important as it tells the loader to load all manifests in the folder.
			"config/deploy/crds/",
		}),
	)
	clusterInitializers = append(clusterInitializers, clusterAgentInitializer)
	// install dependent operators, like cnpg, traefik, cert-manager
	if tc.config.InstallOperators {
		clusterInitializers = append(clusterInitializers, tc.getClusterDependencies()...)
	}

	// Load Barman Cloud plugin from GitHub
	clusterInitializers = append(clusterInitializers,
		httpObjectApplier(
			fmt.Sprintf(
				"https://github.com/cloudnative-pg/plugin-barman-cloud/releases/download/%s/manifest.yaml",
				GetBarmanCloudVersion(),
			),
		),
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
func (tc *TestCluster) GetKubeClient() (*kubernetes.Clientset, error) {
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

// DeployClusterAgent fetches the deployment manifest from the cluster-agent
// GitHub repository, patches the container image, and deploys it.
func (tc *TestCluster) DeployClusterAgent(ctx context.Context, imageTag string) error {
	if tc.environment == nil || tc.environment.Cluster == nil {
		return fmt.Errorf("cluster not initialized")
	}

	loader := NewGitHubLoader(
		ctx,
		WithCacheDir(tc.config.CacheDir),
		WithRepoOwner(clusterAgentOwner),
		WithRepoName(clusterAgentRepo),
		WithRepoTag(GetClusterAgentVersion()),
	)

	manifestBytes, err := loader.FetchManifest(ctx, clusterAgentDeploymentManifestPath)
	if err != nil {
		return fmt.Errorf("fetching deployment manifest: %w", err)
	}

	deployment := &appsv1.Deployment{}
	if err := yaml.Unmarshal(manifestBytes, deployment); err != nil {
		return fmt.Errorf("parsing deployment manifest: %w", err)
	}

	tag := imageTag
	if tag == "" {
		tag = GetClusterAgentVersion()
	}
	image := fmt.Sprintf("%s:%s", clusterAgentImage, tag)
	for i := range deployment.Spec.Template.Spec.Containers {
		if deployment.Spec.Template.Spec.Containers[i].Name == "manager" {
			deployment.Spec.Template.Spec.Containers[i].Image = image
			break
		}
	}

	if err := tc.environment.Cluster.CreateAndWaitForReadiness(ctx, deployment); err != nil {
		return fmt.Errorf("deploying cluster agent manager: %w", err)
	}

	tc.logger.Info("Cluster agent deployed", "image", image)
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
			RepoURL:     TraefikChartRepo,
			PackageName: "traefik",
			Namespace:   "traefik-v2",
			ReleaseName: "traefik",
		},
		dev.ClusterHelmInstall{
			RepoName:    "cnpg",
			RepoURL:     CNPGChartRepo,
			PackageName: "cloudnative-pg",
			Namespace:   "cnpg-system",
			ReleaseName: "cnpg",
		},
		dev.ClusterHelmInstall{
			RepoName:    "jetstack",
			RepoURL:     CertManagerChartRepo,
			PackageName: "cert-manager",
			Namespace:   "cert-manager",
			ReleaseName: "cert-manager",
			SetVars: []string{
				"crds.enabled=true",
			},
		},
	}
}
