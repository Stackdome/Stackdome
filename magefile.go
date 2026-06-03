//go:build mage
// +build mage

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/joho/godotenv"
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
	DefaultClusterName = "stackdome-dev"

	// Timeouts
	ClusterCreateTimeout = "10m"
	OperatorReadyTimeout = "5m"

	// Versions
	KindVersion      = "v0.20.0"
	YqVersion        = "v4.44.1"
	HelmVersion      = "v3.14.0"
	KubectlVersion   = "v1.29.0"
	GoimportsVersion = "latest"
	PnpmVersion      = "v10.33.2"

	// Stackdome agent Helm chart
	DefaultStackdomeChartVersion = "0.5.9-alpha"
	StackdomeChartRepo           = "oci://quay.io/stackdome/charts/stackdome-agent"
	StackdomeChartReleaseName    = "stackdome-agent"
	StackdomeChartNamespace      = "stackdome-control-plane"

	// Dev environment
	PGContainerName = "psql-stackdome-dev"
	DevConfigFile   = "dev_env.yaml"

	// RBAC resource names
	DevSAName      = "stackdome-api-server-account"
	DevSANamespace = "stackdome-control-plane"
	DevRoleName    = "stackdome-api-server-role"
	DevSecretName  = "stackdome-api-server-account-secret"
)

// GetClusterName returns the cluster name from K3D_CLUSTER_NAME or KIND_CLUSTER_NAME env var, or the default.
func GetClusterName() string {
	if name := os.Getenv("K3D_CLUSTER_NAME"); name != "" {
		return name
	}
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

// BuildFrontend builds the Vite SPA into pkg/web/dist for //go:embed.
// Skips `tsc -b` so unrelated type errors don't block the build.
func BuildFrontend(ctx context.Context) error {
	fmt.Println("Building frontend (pnpm install + vite build)...")
	if err := installDep(ctx, "pnpm", PnpmVersion); err != nil {
		return fmt.Errorf("failed to ensure pnpm: %w", err)
	}
	if _, err := exec.LookPath("node"); err != nil {
		return fmt.Errorf("node not found on PATH (Node >=20.12 required, see frontend/package.json engines)")
	}
	pnpmBin := filepath.Join(binDir, "pnpm")
	if err := sh.RunV(pnpmBin, "--prefix", "frontend", "install", "--frozen-lockfile"); err != nil {
		return fmt.Errorf("pnpm install failed: %w", err)
	}
	if err := sh.RunV(pnpmBin, "--prefix", "frontend", "exec", "vite", "build"); err != nil {
		return fmt.Errorf("vite build failed: %w", err)
	}
	// Vite's emptyOutDir wipes .gitkeep; restore so go:embed has a non-empty target.
	if err := os.WriteFile("pkg/web/dist/.gitkeep", nil, 0644); err != nil {
		return fmt.Errorf("restoring .gitkeep failed: %w", err)
	}
	return nil
}

func Build() error {
	mg.Deps(BuildFrontend)
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
		{"pnpm", PnpmVersion},
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
	case "pnpm":
		return installPnpm(ctx, version)
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

func installPnpm(ctx context.Context, version string) error {
	osName := goOs
	if osName == "darwin" {
		osName = "macos"
	}
	arch := goArch
	if arch == "amd64" {
		arch = "x64"
	}
	binaryName := fmt.Sprintf("pnpm-%s-%s", osName, arch)
	url := fmt.Sprintf("https://github.com/pnpm/pnpm/releases/download/%s/%s",
		version, binaryName)

	binPath := filepath.Join(binDir, "pnpm")
	if err := downloadFile(ctx, url, binPath); err != nil {
		return fmt.Errorf("failed to download pnpm: %w", err)
	}

	if err := os.Chmod(binPath, 0755); err != nil {
		return fmt.Errorf("failed to make pnpm executable: %w", err)
	}

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

	cmd := exec.CommandContext(ctx, "go", "test", "./pkg/...", "-v", "-race", "-count=1")
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

// =============================================================================
// Dev Namespace
// =============================================================================

// Dev namespace for local development environment management
type Dev mg.Namespace

// DevDBConfig holds PostgreSQL connection configuration for the dev environment.
type DevDBConfig struct {
	Host     string
	Port     int
	Name     string
	Username string
	Password string
}

// DevCredentials holds the extracted cluster credentials.
type DevCredentials struct {
	ClusterURL string
	CAData     string
	SAToken    string
}

// Setup bootstraps a complete local development environment:
// k3d cluster with stackdome-agent, RBAC for cluster registration,
// and a PostgreSQL container. Safe to run multiple times.
func (Dev) Setup(ctx context.Context) error {
	fmt.Println("========================================")
	fmt.Println(" Setting up Stackdome dev environment")
	fmt.Println("========================================")
	fmt.Println()

	// Step 1: Load .env and get DB config
	dbConfig := loadDevDBConfig()

	// Step 2: Start PostgreSQL container
	fmt.Println("[1/4] Starting PostgreSQL...")
	if err := startPostgresContainer(ctx, dbConfig); err != nil {
		return fmt.Errorf("failed to start PostgreSQL: %w", err)
	}

	// Step 3: Create k3d cluster + install stackdome-agent
	fmt.Println()
	fmt.Println("[2/4] Setting up k3d cluster with stackdome-agent...")

	clusterName := GetClusterName()

	if k3dClusterExists(ctx, clusterName) {
		fmt.Printf("k3d cluster '%s' already exists. Reusing.\n", clusterName)
	} else {
		fmt.Println("Creating k3d cluster...")
		createCmd := exec.CommandContext(ctx, "k3d", "cluster", "create", clusterName,
			"--port", "80:80@loadbalancer",
			"--port", "443:443@loadbalancer",
			"--k3s-arg", "--disable=traefik@server:0",
			"--agents", "2",
			"--wait",
		)
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		if err := createCmd.Run(); err != nil {
			return fmt.Errorf("failed to create k3d cluster: %w", err)
		}
		fmt.Println("✅ Cluster created.")
	}

	kubeconfig, err := k3dKubeconfig(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	if err := ensureStackdomeAgent(ctx, kubeconfig); err != nil {
		return fmt.Errorf("failed to ensure stackdome-agent: %w", err)
	}

	// Step 4: Deploy RBAC resources
	fmt.Println()
	fmt.Println("[3/4] Deploying RBAC resources...")
	if err := deployDevRBAC(ctx, kubeconfig); err != nil {
		return fmt.Errorf("failed to deploy RBAC: %w", err)
	}

	// Step 5: Extract credentials
	fmt.Println()
	fmt.Println("[4/4] Extracting cluster credentials...")
	creds, err := extractDevCredentials(ctx, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to extract credentials: %w", err)
	}

	// Step 6: Write config file and print summary
	configPath := filepath.Join(projectRoot, DevConfigFile)
	if err := writeDevConfig(configPath, clusterName, kubeconfig, creds, dbConfig); err != nil {
		return fmt.Errorf("failed to write dev config: %w", err)
	}

	printDevSummary(clusterName, kubeconfig, creds, dbConfig, configPath)
	return nil
}

// k3dClusterExists checks whether a k3d cluster with the given name exists.
func k3dClusterExists(ctx context.Context, name string) bool {
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), fmt.Sprintf("\"name\":\"%s\"", name))
}

// k3dKubeconfig writes the k3d kubeconfig to a temp file and returns its path.
func k3dKubeconfig(ctx context.Context, clusterName string) (string, error) {
	cmd := exec.CommandContext(ctx, "k3d", "kubeconfig", "get", clusterName)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("k3d kubeconfig get: %w", err)
	}
	kubeconfigPath := filepath.Join(cacheDir, "dev-env", "kubeconfig.yaml")
	if err := os.MkdirAll(filepath.Dir(kubeconfigPath), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(kubeconfigPath, output, 0o600); err != nil {
		return "", err
	}
	return kubeconfigPath, nil
}

// ensureStackdomeAgent runs helm upgrade --install so the chart is
// present even when the cluster was already running.
func ensureStackdomeAgent(ctx context.Context, kubeconfig string) error {
	chartVersion := GetStackdomeChartVersion()
	args := []string{
		"upgrade", "--install", StackdomeChartReleaseName, StackdomeChartRepo,
		"--version", chartVersion,
		"--namespace", StackdomeChartNamespace,
		"--create-namespace",
	}
	logger.Info("Ensuring stackdome-agent chart", "version", chartVersion)
	cmd := exec.CommandContext(ctx, "helm", args...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeconfig)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("helm upgrade --install failed: %w", err)
	}
	return nil
}

// Teardown removes the local development environment. Safe to run
// even if some components were already removed.
func (Dev) Teardown(ctx context.Context) error {
	fmt.Println("Tearing down Stackdome dev environment...")

	// Remove postgres container
	if containerRuntime == "" {
		containerRuntime, _ = detectContainerRuntime()
	}
	if containerRuntime != "" {
		fmt.Println("Stopping PostgreSQL container...")
		cmd := exec.CommandContext(ctx, containerRuntime, "rm", "-f", PGContainerName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
	} else {
		fmt.Println("Warning: no container runtime found, skipping postgres cleanup")
	}

	// Delete k3d cluster (handles missing cluster gracefully)
	clusterName := GetClusterName()
	fmt.Printf("Deleting k3d cluster '%s'...\n", clusterName)
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "delete", clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Warning: failed to delete k3d cluster: %v\n", err)
	}

	// Remove config file
	configPath := filepath.Join(projectRoot, DevConfigFile)
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: failed to remove %s: %v\n", DevConfigFile, err)
	}

	fmt.Println("✅ Dev environment torn down")
	return nil
}

// loadDevDBConfig loads PostgreSQL config from .env file and environment,
// then writes any missing values back to .env so that subsequent commands
// (mage migrate, mage run) work without manual configuration.
func loadDevDBConfig() *DevDBConfig {
	envFile := filepath.Join(projectRoot, ".env")
	_ = godotenv.Load(envFile)

	defaults := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_NAME":     "stackdome_dev",
		"DB_USERNAME": "postgres",
		"DB_PASSWORD": "foobar-bizz-buzz",
	}

	port := 5432
	if v := os.Getenv("DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	config := &DevDBConfig{
		Host:     getEnv("DB_HOST", defaults["DB_HOST"]),
		Port:     port,
		Name:     getEnv("DB_NAME", defaults["DB_NAME"]),
		Username: getEnv("DB_USERNAME", defaults["DB_USERNAME"]),
		Password: getEnv("DB_PASSWORD", defaults["DB_PASSWORD"]),
	}

	ensureDotEnvDefaults(envFile, defaults)
	return config
}

// ensureDotEnvDefaults reads the .env file and appends any missing keys
// with their default values. Creates the file from .env_template if it
// doesn't exist.
func ensureDotEnvDefaults(envFile string, defaults map[string]string) {
	// If .env doesn't exist, seed it from .env_template
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		templateFile := filepath.Join(projectRoot, ".env_template")
		if tmpl, err := os.ReadFile(templateFile); err == nil {
			_ = os.WriteFile(envFile, tmpl, 0600)
			fmt.Println("Created .env from .env_template")
		} else {
			_ = os.WriteFile(envFile, []byte(""), 0600)
			fmt.Println("Created empty .env")
		}
	}

	existing, err := godotenv.Read(envFile)
	if err != nil {
		existing = map[string]string{}
	}

	// Append only missing keys to preserve file ordering and comments
	var lines []string
	for key, defaultVal := range defaults {
		if _, exists := existing[key]; !exists {
			lines = append(lines, fmt.Sprintf("%s=%q", key, defaultVal))
		}
	}

	if len(lines) > 0 {
		f, err := os.OpenFile(envFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			fmt.Printf("Warning: failed to update .env: %v\n", err)
			return
		}
		defer f.Close()

		content := "\n" + strings.Join(lines, "\n") + "\n"
		if _, err := f.WriteString(content); err != nil {
			fmt.Printf("Warning: failed to write to .env: %v\n", err)
			return
		}

		var keys []string
		for _, line := range lines {
			keys = append(keys, strings.SplitN(line, "=", 2)[0])
		}
		fmt.Printf("Added missing defaults to .env: %s\n", strings.Join(keys, ", "))
	}
}

// startPostgresContainer starts a PostgreSQL Docker container for the dev environment.
func startPostgresContainer(ctx context.Context, config *DevDBConfig) error {
	if containerRuntime == "" {
		var err error
		containerRuntime, err = detectContainerRuntime()
		if err != nil {
			return err
		}
	}

	// Check if container is already running
	cmd := exec.CommandContext(ctx, containerRuntime, "ps", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), PGContainerName) {
		fmt.Printf("PostgreSQL container '%s' already running.\n", PGContainerName)
		return nil
	}

	// Check if container exists but is stopped
	cmd = exec.CommandContext(ctx, containerRuntime, "ps", "-a", "--format", "{{.Names}}")
	output, err = cmd.Output()
	if err == nil && strings.Contains(string(output), PGContainerName) {
		fmt.Printf("Starting existing PostgreSQL container '%s'...\n", PGContainerName)
		cmd = exec.CommandContext(ctx, containerRuntime, "start", PGContainerName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start existing container: %w", err)
		}
	} else {
		// Create new container
		fmt.Printf("Creating PostgreSQL container '%s'...\n", PGContainerName)
		cmd = exec.CommandContext(ctx, containerRuntime, "run",
			"--name", PGContainerName,
			"-e", fmt.Sprintf("POSTGRES_USER=%s", config.Username),
			"-e", fmt.Sprintf("POSTGRES_PASSWORD=%s", config.Password),
			"-e", fmt.Sprintf("POSTGRES_DB=%s", config.Name),
			"-p", fmt.Sprintf("%d:5432", config.Port),
			"-d", "postgres",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create postgres container: %w", err)
		}
	}

	// Wait for PostgreSQL to be ready
	fmt.Println("Waiting for PostgreSQL to be ready...")
	for i := 0; i < 30; i++ {
		cmd = exec.CommandContext(ctx, containerRuntime, "exec", PGContainerName,
			"pg_isready", "-U", config.Username)
		if err := cmd.Run(); err == nil {
			fmt.Printf("PostgreSQL ready (database: %s, port: %d)\n", config.Name, config.Port)
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("PostgreSQL failed to start within 30 seconds")
}

// deployDevRBAC creates the service account, roles, and bindings needed
// for the API server to manage the cluster.
func deployDevRBAC(ctx context.Context, kubeconfig string) error {
	// Ensure namespace exists (may already exist from Helm chart)
	if err := runKubectlApply(ctx, kubeconfig, fmt.Sprintf(`
apiVersion: v1
kind: Namespace
metadata:
  name: %s`, DevSANamespace)); err != nil {
		return fmt.Errorf("failed to ensure namespace: %w", err)
	}

	// Create ServiceAccount
	if err := runKubectlApply(ctx, kubeconfig, fmt.Sprintf(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: %s
  namespace: %s`, DevSAName, DevSANamespace)); err != nil {
		return fmt.Errorf("failed to create service account: %w", err)
	}

	// Create ClusterRole with full permissions for API server
	if err := runKubectlApply(ctx, kubeconfig, fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: %s
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["*"]`, DevRoleName)); err != nil {
		return fmt.Errorf("failed to create cluster role: %w", err)
	}

	// Create ClusterRoleBinding
	if err := runKubectlApply(ctx, kubeconfig, fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: %s-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: %s
subjects:
  - kind: ServiceAccount
    name: %s
    namespace: %s`, DevRoleName, DevRoleName, DevSAName, DevSANamespace)); err != nil {
		return fmt.Errorf("failed to create cluster role binding: %w", err)
	}

	// Create Secret for ServiceAccount token
	if err := runKubectlApply(ctx, kubeconfig, fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: %s
  namespace: %s
  annotations:
    kubernetes.io/service-account.name: %s
type: kubernetes.io/service-account-token`, DevSecretName, DevSANamespace, DevSAName)); err != nil {
		return fmt.Errorf("failed to create secret: %w", err)
	}

	// Wait for token to be populated
	fmt.Println("Waiting for service account token...")
	for i := 0; i < 30; i++ {
		cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
			"get", "secret", DevSecretName,
			"-n", DevSANamespace,
			"-o", "jsonpath={.data.token}")
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			fmt.Println("Service account token ready.")
			return nil
		}
		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for service account token (30s)")
}

// extractDevCredentials retrieves the cluster URL, CA data, and SA token.
func extractDevCredentials(ctx context.Context, kubeconfig string) (*DevCredentials, error) {
	// Get cluster API URL
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"config", "view", "--raw", "--minify", "--flatten",
		"-o", "jsonpath={.clusters[0].cluster.server}")
	urlOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster URL: %w", err)
	}

	// Get CA data (already base64 encoded in the secret)
	cmd = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "secret", DevSecretName,
		"-n", DevSANamespace,
		"-o", "jsonpath={.data.ca\\.crt}")
	caOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get CA data: %w", err)
	}

	// Get SA token (base64 encoded in the secret, needs decoding)
	cmd = exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig,
		"get", "secret", DevSecretName,
		"-n", DevSANamespace,
		"-o", "jsonpath={.data.token}")
	tokenB64, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get SA token: %w", err)
	}

	token, err := base64.StdEncoding.DecodeString(string(tokenB64))
	if err != nil {
		return nil, fmt.Errorf("failed to decode SA token: %w", err)
	}

	return &DevCredentials{
		ClusterURL: string(urlOutput),
		CAData:     string(caOutput),
		SAToken:    string(token),
	}, nil
}

// writeDevConfig writes the dev environment configuration to a YAML file.
func writeDevConfig(path, clusterName, kubeconfig string, creds *DevCredentials, db *DevDBConfig) error {
	content := fmt.Sprintf(`kubeconfig: %s
cluster_name: %s
cluster_url: %s
cluster_ca_data: %s
cluster_sa_token: %s
kubectl_context: k3d-%s
db_host: %s
db_port: %d
db_name: %s
db_username: %s
db_password: %s
`, kubeconfig, clusterName, creds.ClusterURL, creds.CAData, creds.SAToken,
		clusterName, db.Host, db.Port, db.Name, db.Username, db.Password)

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// printDevSummary prints the dev environment details and next steps.
func printDevSummary(clusterName, kubeconfig string, creds *DevCredentials, db *DevDBConfig, configPath string) {
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println(" Stackdome dev environment is ready!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Cluster:")
	fmt.Printf("  Name:        %s\n", clusterName)
	fmt.Printf("  API URL:     %s\n", creds.ClusterURL)
	fmt.Printf("  Kubeconfig:  %s\n", kubeconfig)
	fmt.Printf("  Context:     kind-%s\n", clusterName)
	fmt.Println()
	fmt.Println("Service Account:")
	fmt.Printf("  Token:       %s...%s\n", creds.SAToken[:10], creds.SAToken[len(creds.SAToken)-10:])
	fmt.Printf("  CA Data:     %s...\n", creds.CAData[:40])
	fmt.Println()
	fmt.Println("PostgreSQL:")
	fmt.Printf("  Host:        %s\n", db.Host)
	fmt.Printf("  Port:        %d\n", db.Port)
	fmt.Printf("  Database:    %s\n", db.Name)
	fmt.Printf("  Username:    %s\n", db.Username)
	fmt.Println()
	fmt.Printf("Config written to: %s\n", configPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println()
	fmt.Println("  # Run database migrations")
	fmt.Println("  mage migrate")
	fmt.Println()
	fmt.Println("  # Start the API server")
	fmt.Println("  mage run")
	fmt.Println()
	fmt.Println("  # Watch cluster pods")
	fmt.Printf("  kubectl --context kind-%s get pods -A -w\n", clusterName)
	fmt.Println()
	fmt.Println("  # Tear down the environment")
	fmt.Println("  mage dev:teardown")
	fmt.Println()
}

// runKubectlApply applies a YAML manifest via kubectl apply.
func runKubectlApply(ctx context.Context, kubeconfig, manifest string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply failed: %w\nOutput: %s", err, output)
	}
	return nil
}
