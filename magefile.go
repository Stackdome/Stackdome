//go:build mage
// +build mage

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mt-sre/devkube/dev"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Global configuration
var (
	workingDir       string
	projectRoot      string
	cacheDir         string
	binDir           string
	goOs             string
	goArch           string
	logger           logr.Logger
	devEnvironment   *dev.Environment
	containerRuntime string
)

func init() {
	var err error
	workingDir, err = os.Getwd()
	if err != nil {
		panic(fmt.Errorf("failed to get working directory: %w", err))
	}

	projectRoot = workingDir

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get home directory: %w", err))
	}

	cacheDir = path.Join(homeDir, ".cache", "stackdome-api-server")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create cache directory: %w", err))
	}

	binDir = path.Join(cacheDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create bin directory: %w", err))
	}

	goOs = runtime.GOOS
	goArch = runtime.GOARCH

	// Add bin directory to PATH
	currentPath := os.Getenv("PATH")
	os.Setenv("PATH", fmt.Sprintf("%s:%s", binDir, currentPath))

	// Initialize logger
	logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags)).WithName("mage")

	// Detect container runtime
	containerRuntime, err = detectContainerRuntime()
	if err != nil {
		// Don't panic here, let individual commands handle the error
		containerRuntime = ""
	}
}

// Configuration constants
const (
	// Test cluster configuration
	DefaultClusterName = "stackdome-dev-cluster"

	// Timeouts
	ClusterCreateTimeout = "10m"
	OperatorReadyTimeout = "5m"

	// Versions
	KindVersion      = "v0.20.0"
	YqVersion        = "v4.44.1"
	HelmVersion      = "v3.14.0"
	KubectlVersion   = "v1.29.0"
	GoimportsVersion = "latest"

	// Stackdome agent Helm chart
	DefaultStackdomeChartVersion = "0.5.1-alpha"
	StackdomeChartRepo           = "oci://quay.io/stackdome/charts/stackdome-agent"
	StackdomeChartReleaseName    = "stackdome-agent"
	StackdomeChartNamespace      = "stackdome-control-plane"
)

// GetClusterName returns the cluster name from KIND_CLUSTER_NAME env var or the default.
func GetClusterName() string {
	if name := os.Getenv("KIND_CLUSTER_NAME"); name != "" {
		return name
	}
	return DefaultClusterName
}

// GetStackdomeChartVersion returns the chart version from env or the default.
func GetStackdomeChartVersion() string {
	if v := os.Getenv("STACKDOME_CHART_VERSION"); v != "" {
		return v
	}
	return DefaultStackdomeChartVersion
}

// Environment variable helpers
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Errorf("required environment variable %s not set", key))
	}
	return value
}

// =============================================================================
// Core Build Functions (Global Namespace)
// =============================================================================

// Build builds the API server binary
func Build() error {
	fmt.Println("Building API server...")
	return sh.Run("go", "build", "-o", "bin/api-server", "./cmd")
}

// Run runs the API server locally
func Run() error {
	mg.Deps(Build)
	return sh.Run("./bin/api-server", "serve")
}

// Generate regenerates OpenAPI client code
func Generate() error {
	fmt.Println("Generating OpenAPI client code...")

	// Run OpenAPI generator
	cmd := exec.Command("make", "generate-client")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate OpenAPI client: %w", err)
	}

	// Format generated code
	return sh.Run("go", "fmt", "./pkg/api/openapi/...")
}

// Fmt formats the code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Lint runs the linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "./...")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return os.RemoveAll("bin")
}

// Migrate runs database migrations
func Migrate() error {
	fmt.Println("Running database migrations...")
	return sh.Run("go", "run", "./cmd", "migrate")
}

// Default target
var Default = Build

// =============================================================================
// Dependencies Namespace
// =============================================================================

// Deps namespace for dependency management
type Deps mg.Namespace

// Install installs all required dependencies for integration testing
func (Deps) Install(ctx context.Context) error {
	fmt.Println("Installing integration test dependencies...")

	deps := []struct {
		name    string
		version string
	}{
		{"kind", KindVersion},
		{"yq", YqVersion},
		{"helm", HelmVersion},
		{"kubectl", KubectlVersion},
	}

	for _, dep := range deps {
		if err := installDep(ctx, dep.name, dep.version); err != nil {
			return fmt.Errorf("failed to install %s: %w", dep.name, err)
		}
	}

	// Install Go dependencies
	goDeps := []string{
		"golang.org/x/tools/cmd/goimports@" + GoimportsVersion,
	}

	for _, dep := range goDeps {
		if err := installGoDep(ctx, dep); err != nil {
			return err
		}
	}

	fmt.Println("✅ All dependencies installed successfully")
	return nil
}

// Clean removes all installed dependencies
func (Deps) Clean() error {
	fmt.Println("Cleaning installed dependencies...")

	if err := os.RemoveAll(binDir); err != nil {
		return fmt.Errorf("failed to remove bin directory: %w", err)
	}

	// Recreate bin directory
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate bin directory: %w", err)
	}

	fmt.Println("✅ Dependencies cleaned")
	return nil
}

// installDep installs a binary dependency if not already present
func installDep(ctx context.Context, name, version string) error {
	binPath := filepath.Join(binDir, name)

	// Check if already installed
	if _, err := os.Stat(binPath); err == nil {
		// Verify it's executable
		cmd := exec.CommandContext(ctx, binPath, "--version")
		if err := cmd.Run(); err == nil {
			fmt.Printf("✓ %s already installed\n", name)
			return nil
		}
	}

	fmt.Printf("Installing %s %s...\n", name, version)

	// Download and install based on the tool
	switch name {
	case "kind":
		return installKind(ctx, version)
	case "yq":
		return installYq(ctx, version)
	case "helm":
		return installHelm(ctx, version)
	case "kubectl":
		return installKubectl(ctx, version)
	default:
		return fmt.Errorf("unknown dependency: %s", name)
	}
}

func installKind(ctx context.Context, version string) error {
	url := fmt.Sprintf("https://github.com/kubernetes-sigs/kind/releases/download/%s/kind-%s-%s",
		version, goOs, goArch)

	binPath := filepath.Join(binDir, "kind")
	if err := downloadFile(ctx, url, binPath); err != nil {
		return fmt.Errorf("failed to download kind: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to make kind executable: %w", err)
	}

	return nil
}

func installYq(ctx context.Context, version string) error {
	binaryName := fmt.Sprintf("yq_%s_%s", goOs, goArch)
	url := fmt.Sprintf("https://github.com/mikefarah/yq/releases/download/%s/%s",
		version, binaryName)

	binPath := filepath.Join(binDir, "yq")
	if err := downloadFile(ctx, url, binPath); err != nil {
		return fmt.Errorf("failed to download yq: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to make yq executable: %w", err)
	}

	return nil
}

func installHelm(ctx context.Context, version string) error {
	archiveName := fmt.Sprintf("helm-%s-%s-%s.tar.gz", version, goOs, goArch)
	url := fmt.Sprintf("https://get.helm.sh/%s", archiveName)

	tempFile := filepath.Join(cacheDir, archiveName)
	if err := downloadFile(ctx, url, tempFile); err != nil {
		return fmt.Errorf("failed to download helm: %w", err)
	}
	defer os.Remove(tempFile)

	// Extract helm binary
	cmd := exec.CommandContext(ctx, "tar", "-xzf", tempFile, "-C", cacheDir,
		fmt.Sprintf("%s-%s/helm", goOs, goArch))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract helm: %w", err)
	}

	// Move to bin directory
	src := filepath.Join(cacheDir, fmt.Sprintf("%s-%s", goOs, goArch), "helm")
	dst := filepath.Join(binDir, "helm")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move helm to bin: %w", err)
	}

	// Cleanup
	os.RemoveAll(filepath.Join(cacheDir, fmt.Sprintf("%s-%s", goOs, goArch)))

	return nil
}

func installKubectl(ctx context.Context, version string) error {
	url := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/%s/%s/kubectl",
		version, goOs, goArch)

	binPath := filepath.Join(binDir, "kubectl")
	if err := downloadFile(ctx, url, binPath); err != nil {
		return fmt.Errorf("failed to download kubectl: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to make kubectl executable: %w", err)
	}

	return nil
}

func downloadFile(ctx context.Context, url, filepath string) error {
	cmd := exec.CommandContext(ctx, "curl", "-L", "-o", filepath, url)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("download failed: %w\nOutput: %s", err, output)
	}
	return nil
}

// installGoDep installs a Go-based tool using go install
func installGoDep(ctx context.Context, pkg string) error {
	fmt.Printf("Installing %s...\n", pkg)

	cmd := exec.CommandContext(ctx, "go", "install", pkg)
	cmd.Env = append(os.Environ(), fmt.Sprintf("GOBIN=%s", binDir))

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to install %s: %w\nOutput: %s", pkg, err, output)
	}

	return nil
}

// detectContainerRuntime detects the available container runtime
func detectContainerRuntime() (string, error) {
	// Check for Docker first
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker", nil
	}

	// Check for Podman
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman", nil
	}

	return "", fmt.Errorf("no container runtime found (docker or podman required)")
}

// =============================================================================
// Cluster Namespace
// =============================================================================

// Cluster namespace for cluster management
type Cluster mg.Namespace

// TestCluster represents a test cluster instance
type TestCluster struct {
	Name       string
	ConfigPath string
	CacheDir   string
	Runtime    string
}

// ClusterState represents the persisted state of a test cluster
type ClusterState struct {
	Name               string    `json:"name"`
	CreatedAt          time.Time `json:"created_at"`
	Runtime            string    `json:"runtime"`
	ConfigPath         string    `json:"config_path"`
	OperatorsInstalled bool      `json:"operators_installed"`
	CRDsInstalled      bool      `json:"crds_installed"`
	AgentDeployed      bool      `json:"agent_deployed"`
}

// Setup creates a Kind cluster and installs the stackdome-agent Helm chart with all dependencies
func (Cluster) Setup(ctx context.Context) error {
	mg.Deps(mg.F(Deps.Install))

	if err := initDevEnvironment(); err != nil {
		return err
	}

	if err := devEnvironment.Init(ctx); err != nil {
		return fmt.Errorf("initializing dev environment: %w", err)
	}

	fmt.Printf("\n✅ Test cluster ready!\n")
	fmt.Printf("KUBECONFIG=%s\n", filepath.Join(devEnvironment.WorkDir, "kubeconfig.yaml"))

	return nil
}

// Delete removes the test cluster
func (Cluster) Delete(ctx context.Context) error {
	if devEnvironment == nil {
		if err := initDevEnvironment(); err != nil {
			return err
		}
	}

	if err := devEnvironment.Destroy(ctx); err != nil {
		return fmt.Errorf("tearing down dev environment: %w", err)
	}

	fmt.Println("✅ Test cluster deleted successfully")
	return nil
}

// Status shows the status of the test cluster
func (Cluster) Status(ctx context.Context) error {
	if err := initDevEnvironment(); err != nil {
		return err
	}

	// Check if cluster exists using kind
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("❌ Error checking clusters: %v\n", err)
		return nil
	}

	clusterName := GetClusterName()
	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	clusterExists := false
	for _, cluster := range clusters {
		if cluster == clusterName {
			clusterExists = true
			break
		}
	}

	if !clusterExists {
		fmt.Printf("❌ Cluster %s does not exist\n", clusterName)
		return nil
	}

	fmt.Printf("✅ Cluster %s exists\n", clusterName)
	fmt.Printf("Runtime: %s\n", containerRuntime)
	fmt.Printf("KUBECONFIG: %s\n", filepath.Join(devEnvironment.WorkDir, "kubeconfig.yaml"))

	return nil
}

// Kubeconfig prints the kubeconfig path
func (Cluster) Kubeconfig() error {
	if err := initDevEnvironment(); err != nil {
		return err
	}

	// Check if cluster exists using kind
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error checking clusters: %w", err)
	}

	clusterName := GetClusterName()
	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	clusterExists := false
	for _, cluster := range clusters {
		if cluster == clusterName {
			clusterExists = true
			break
		}
	}

	if !clusterExists {
		return fmt.Errorf("cluster %s does not exist", clusterName)
	}

	fmt.Println(filepath.Join(devEnvironment.WorkDir, "kubeconfig.yaml"))
	return nil
}

// initDevEnvironment initializes the devkube environment
func initDevEnvironment() error {
	if devEnvironment != nil {
		return nil
	}

	if containerRuntime == "" {
		var err error
		containerRuntime, err = detectContainerRuntime()
		if err != nil {
			return err
		}
	}

	ctrl.SetLogger(logger)

	chartVersion := GetStackdomeChartVersion()
	logger.Info("Using stackdome-agent chart", "version", chartVersion)

	clusterInitializers := []dev.ClusterInitializer{
		dev.ClusterInitFn(func(ctx context.Context, cluster *dev.Cluster) error {
			args := []string{
				"upgrade", "--install", StackdomeChartReleaseName, StackdomeChartRepo,
				"--version", chartVersion,
				"--namespace", StackdomeChartNamespace,
				"--create-namespace",
			}

			logger.Info("Installing stackdome-agent chart", "args", args)
			cmd := exec.CommandContext(ctx, "helm", args...)
			cmd.Env = append(os.Environ(), "KUBECONFIG="+cluster.Kubeconfig())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("installing stackdome-agent chart: %w", err)
			}
			return nil
		}),
	}

	devEnvironment = dev.NewEnvironment(
		GetClusterName(),
		path.Join(cacheDir, "dev-env"),
		dev.WithClusterOptions([]dev.ClusterOption{
			dev.WithWaitOptions([]dev.WaitOption{
				dev.WithTimeout(10 * time.Minute),
			}),
			dev.WithSchemeBuilder(k8sruntime.SchemeBuilder{
				clientgoscheme.AddToScheme,
			}),
		}),
		dev.WithContainerRuntime(containerRuntime),
		dev.WithKindClusterConfig(kindv1alpha4.Cluster{
			Nodes: []kindv1alpha4.Node{
				{
					Role: kindv1alpha4.ControlPlaneRole,
				},
				{
					Role: kindv1alpha4.WorkerRole,
				},
			},
		}),
		dev.WithClusterInitializers(clusterInitializers),
	)

	return nil
}

// newTestCluster creates a new test cluster instance (legacy for compatibility)
func newTestCluster() (*TestCluster, error) {
	if devEnvironment == nil {
		if err := initDevEnvironment(); err != nil {
			return nil, err
		}
	}

	clusterName := GetClusterName()
	return &TestCluster{
		Name:       clusterName,
		ConfigPath: filepath.Join(devEnvironment.WorkDir, "kubeconfig.yaml"),
		CacheDir:   filepath.Join(cacheDir, "clusters", clusterName),
		Runtime:    containerRuntime,
	}, nil
}

// exists checks if the cluster exists (compatibility method)
func (tc *TestCluster) exists(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, cluster := range clusters {
		if cluster == tc.Name {
			return true
		}
	}

	return false
}

// getKubeconfig returns the path to the kubeconfig file (compatibility method)
func (tc *TestCluster) getKubeconfig() string {
	if devEnvironment != nil {
		return filepath.Join(devEnvironment.WorkDir, "kubeconfig.yaml")
	}
	return tc.ConfigPath
}

// =============================================================================
// Test Namespace
// =============================================================================

// Test namespace for test-related tasks
type Test mg.Namespace

// Integration runs the integration tests
func (Test) Integration(ctx context.Context) error {
	mg.Deps(mg.F(Cluster.Setup))
	return runIntegrationTests(ctx, "", false)
}

// IntegrationVerbose runs the integration tests with verbose output
func (Test) IntegrationVerbose(ctx context.Context) error {
	mg.Deps(mg.F(Cluster.Setup))
	return runIntegrationTests(ctx, "", true)
}

// IntegrationFocus runs specific integration tests matching the given pattern
func (t Test) IntegrationFocus(ctx context.Context, pattern string) error {
	mg.Deps(mg.F(Cluster.Setup))
	return runIntegrationTests(ctx, pattern, true)
}

// Unit runs the unit tests
func (Test) Unit(ctx context.Context) error {
	fmt.Println("Running unit tests...")

	cmd := exec.CommandContext(ctx, "go", "test", "./pkg/...", "-v", "-race")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = projectRoot

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	fmt.Println("✅ Unit tests passed")
	return nil
}

// All runs all tests (unit and integration)
func (t Test) All(ctx context.Context) error {
	// Run unit tests first
	if err := t.Unit(ctx); err != nil {
		return err
	}

	// Then run integration tests
	return t.Integration(ctx)
}

// Coverage runs tests with coverage reporting
func (Test) Coverage(ctx context.Context) error {
	fmt.Println("Running tests with coverage...")

	coverageDir := filepath.Join(projectRoot, "coverage")
	if err := os.MkdirAll(coverageDir, 0755); err != nil {
		return fmt.Errorf("failed to create coverage directory: %w", err)
	}

	// Run tests with coverage
	coverFile := filepath.Join(coverageDir, "coverage.out")
	cmd := exec.CommandContext(ctx, "go", "test", "./pkg/...", "-coverprofile", coverFile, "-covermode=atomic")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = projectRoot

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("coverage tests failed: %w", err)
	}

	// Generate HTML report
	htmlFile := filepath.Join(coverageDir, "coverage.html")
	cmd = exec.CommandContext(ctx, "go", "tool", "cover", "-html", coverFile, "-o", htmlFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate coverage report: %w", err)
	}

	// Show coverage summary
	cmd = exec.CommandContext(ctx, "go", "tool", "cover", "-func", coverFile)
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to show coverage summary: %w", err)
	}

	fmt.Printf("\n✅ Coverage report generated: %s\n", htmlFile)
	return nil
}

// Clean removes test artifacts
func (Test) Clean() error {
	fmt.Println("Cleaning test artifacts...")

	// Remove coverage directory
	coverageDir := filepath.Join(projectRoot, "coverage")
	if err := os.RemoveAll(coverageDir); err != nil {
		return fmt.Errorf("failed to remove coverage directory: %w", err)
	}

	// Clean test cache
	cmd := exec.Command("go", "clean", "-testcache")
	cmd.Dir = projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean test cache: %w", err)
	}

	fmt.Println("✅ Test artifacts cleaned")
	return nil
}

// runIntegrationTests executes the integration tests
func runIntegrationTests(ctx context.Context, focus string, verbose bool) error {
	// Get cluster instance (cluster creation handled by dependencies)
	cluster, err := newTestCluster()
	if err != nil {
		return err
	}

	// Verify cluster exists (should exist due to dependency)
	if !cluster.exists(ctx) {
		return fmt.Errorf("test cluster does not exist - this should not happen due to dependency management")
	}

	// Set environment variables for tests
	testEnv := os.Environ()
	testEnv = append(testEnv, fmt.Sprintf("TEST_KUBECONFIG=%s", cluster.getKubeconfig()))

	// Keep cluster after tests if requested
	if keepCluster := os.Getenv("KEEP_CLUSTER"); keepCluster == "true" {
		testEnv = append(testEnv, "KEEP_CLUSTER=true")
	}

	// Build test command
	args := []string{"test", "./test/int/..."}

	if verbose {
		args = append(args, "-v")
	}

	if focus != "" {
		args = append(args, "-run", focus)
	}

	// Add test timeout
	args = append(args, "-timeout", "30m")

	// Run tests
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = testEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = projectRoot

	fmt.Printf("Running: go %s\n", strings.Join(args, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}

	fmt.Println("✅ Integration tests passed")

	// Clean up cluster if not keeping it
	if os.Getenv("KEEP_CLUSTER") != "true" {
		fmt.Println("\nCleaning up test cluster...")
		if err := (Cluster{}).Delete(ctx); err != nil {
			fmt.Printf("Warning: failed to delete test cluster: %v\n", err)
		}
	}

	return nil
}
