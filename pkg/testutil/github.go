package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/mt-sre/devkube/dev"
	"golang.org/x/oauth2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// Default version should match the version in go.mod
	defaultClusterAgentVersion = "v0.4.7-alpha"
	defaultRepoOwner           = "ashishmax31"
	defaultRepoName            = "cluster-agent"

	// GitHub raw content URL pattern
	githubRawURLPattern = "https://raw.githubusercontent.com/%s/%s/%s/%s"
)

// manifestCache holds cached manifest files
var manifestCache = make(map[string][]byte)

// createGitHubClient creates a GitHub client with optional authentication
func createGitHubClient(ctx context.Context) *github.Client {
	// Check for GitHub token - try GHACCESS_TOKEN first (from .env), then GITHUB_TOKEN
	token := os.Getenv("GHACCESS_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		// Return unauthenticated client
		return github.NewClient(nil)
	}

	// Create authenticated client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

// checkTagExists verifies if a tag exists in the repository
func checkTagExists(ctx context.Context, client *github.Client, tag string) (bool, error) {
	// List tags
	tags, _, err := client.Repositories.ListTags(ctx, defaultRepoOwner, defaultRepoName, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		// Try to list releases as an alternative
		releases, _, releaseErr := client.Repositories.ListReleases(ctx, defaultRepoOwner, defaultRepoName, &github.ListOptions{
			PerPage: 100,
		})
		if releaseErr != nil {
			return false, fmt.Errorf("listing tags failed: %w, listing releases failed: %w", err, releaseErr)
		}

		// Check releases
		for _, release := range releases {
			if release.TagName != nil && *release.TagName == tag {
				return true, nil
			}
		}
		return false, nil
	}

	// Check tags
	for _, t := range tags {
		if t.Name != nil && *t.Name == tag {
			return true, nil
		}
	}

	// If tag not found, also check if it exists as a branch
	_, _, err = client.Repositories.GetBranch(ctx, defaultRepoOwner, defaultRepoName, strings.TrimPrefix(tag, "v"), true)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// findLatestTag finds the latest available tag in the repository
func findLatestTag(ctx context.Context, client *github.Client) (string, error) {
	// Try to get the latest release first
	release, _, err := client.Repositories.GetLatestRelease(ctx, defaultRepoOwner, defaultRepoName)
	if err == nil && release.TagName != nil {
		return *release.TagName, nil
	}

	// If no releases, list tags
	tags, _, err := client.Repositories.ListTags(ctx, defaultRepoOwner, defaultRepoName, &github.ListOptions{
		PerPage: 10, // Get only recent tags
	})
	if err != nil {
		return "", fmt.Errorf("listing tags: %w", err)
	}

	if len(tags) > 0 && tags[0].Name != nil {
		return *tags[0].Name, nil
	}

	return "", fmt.Errorf("no tags found in repository")
}

// githubCRDLoader implements dev.ClusterInitializer to load CRDs from GitHub
type githubCRDLoader struct {
	version  string
	cacheDir string
	client   *github.Client
}

func (g *githubCRDLoader) Init(ctx context.Context, cluster *dev.Cluster) error {
	// Create GitHub client if not already created
	if g.client == nil {
		g.client = createGitHubClient(ctx)
	}

	// Verify tag exists
	exists, err := checkTagExists(ctx, g.client, g.version)
	if err != nil {
		return fmt.Errorf("checking if tag %s exists: %w", g.version, err)
	}
	if !exists {
		// Try to find the latest available tag
		latestTag, err := findLatestTag(ctx, g.client)
		if err != nil {
			return fmt.Errorf("tag %s not found and unable to find latest tag: %w", g.version, err)
		}
		return fmt.Errorf("tag %s not found in repository, latest available tag is %s", g.version, latestTag)
	}

	// Prepare temporary directory for CRD files
	tempDir := filepath.Join(g.cacheDir, "temp-crds", g.version)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	// Get CRD file names
	crdFiles := []string{
		"addons.stackdome.io_postgresclusters.yaml",
		"builds.stackdome.io_imagebuilds.yaml",
		"core.stackdome.io_stackresources.yaml",
		"core.stackdome.io_stacks.yaml",
		"registry.stackdome.io_clusterregistries.yaml",
		"storage.stackdome.io_nfsservers.yaml",
		"storage.stackdome.io_storages.yaml",
		"storage.stackdome.io_volumes.yaml",
		"users.stackdome.io_users.yaml",
	}

	// Download and write CRD files to temporary directory
	var filePaths []string
	for _, filename := range crdFiles {
		content, err := g.fetchManifest(ctx, "config/deploy/crds/"+filename)
		if err != nil {
			return fmt.Errorf("fetching CRD %s: %w", filename, err)
		}

		filePath := filepath.Join(tempDir, filename)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("writing CRD file %s: %w", filename, err)
		}
		filePaths = append(filePaths, filePath)
	}

	// Use cluster's CreateAndWaitFromFiles method
	if err := cluster.CreateAndWaitFromFiles(ctx, filePaths); err != nil {
		return fmt.Errorf("creating CRDs from files: %w", err)
	}

	return nil
}

func (g *githubCRDLoader) fetchManifest(ctx context.Context, path string) ([]byte, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s/%s", g.version, path)
	if cached, ok := manifestCache[cacheKey]; ok {
		return cached, nil
	}

	// Create GitHub client if not already created
	if g.client == nil {
		g.client = createGitHubClient(ctx)
	}

	// Try GitHub API first
	fileContent, _, resp, err := g.client.Repositories.GetContents(ctx, defaultRepoOwner, defaultRepoName, path, &github.RepositoryContentGetOptions{
		Ref: g.version,
	})

	// If successful and it's a file
	if err == nil && fileContent != nil && fileContent.Content != nil {
		content, err := fileContent.GetContent()
		if err != nil {
			return nil, fmt.Errorf("decoding content: %w", err)
		}

		contentBytes := []byte(content)

		// Cache content
		manifestCache[cacheKey] = contentBytes

		// Also save to disk cache if cache directory is provided
		if g.cacheDir != "" {
			cacheFile := filepath.Join(g.cacheDir, "manifests", g.version, path)
			if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err == nil {
				os.WriteFile(cacheFile, contentBytes, 0644)
			}
		}

		return contentBytes, nil
	}

	// If API failed with 404 or other error, try raw URL as fallback
	if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 403) {
		// Build GitHub raw URL
		url := fmt.Sprintf(githubRawURLPattern, defaultRepoOwner, defaultRepoName, g.version, path)

		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Add GitHub token if available for private repos
		token := os.Getenv("GHACCESS_TOKEN")
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token != "" {
			req.Header.Set("Authorization", "token "+token)
		}

		// Make request
		httpResp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetching %s: %w", url, err)
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d fetching %s", httpResp.StatusCode, url)
		}

		// Read content
		content, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response: %w", err)
		}

		// Cache content
		manifestCache[cacheKey] = content

		// Also save to disk cache if cache directory is provided
		if g.cacheDir != "" {
			cacheFile := filepath.Join(g.cacheDir, "manifests", g.version, path)
			if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err == nil {
				os.WriteFile(cacheFile, content, 0644)
			}
		}

		return content, nil
	}

	// Return original error
	return nil, fmt.Errorf("fetching file from GitHub: %w", err)
}

// deployManifestsFromGitHub deploys manifests fetched from GitHub
func deployManifestsFromGitHub(ctx context.Context, cluster *dev.Cluster, files []string, version, cacheDir string) error {
	loader := &githubCRDLoader{
		version:  version,
		cacheDir: cacheDir,
		client:   createGitHubClient(ctx),
	}

	// Prepare temporary directory for manifest files
	tempDir := filepath.Join(cacheDir, "temp-manifests", version)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	// Download and write manifest files to temporary directory
	var filePaths []string
	for _, file := range files {
		content, err := loader.fetchManifest(ctx, "config/deploy/"+file)
		if err != nil {
			return fmt.Errorf("fetching manifest %s: %w", file, err)
		}

		filePath := filepath.Join(tempDir, file)
		if err := os.WriteFile(filePath, content, 0644); err != nil {
			return fmt.Errorf("writing manifest file %s: %w", file, err)
		}
		filePaths = append(filePaths, filePath)
	}

	// Use cluster's CreateAndWaitFromFiles method
	if err := cluster.CreateAndWaitFromFiles(ctx, filePaths); err != nil {
		return fmt.Errorf("creating manifests from files: %w", err)
	}

	return nil
}

// getClusterAgentDeployment creates a deployment for the cluster agent manager
func getClusterAgentDeployment(imageTag string) (*appsv1.Deployment, error) {
	var image string
	if imageTag == "" {
		image = "quay.io/stackdome/cluster-agent/cluster-agent-manager:latest"
	} else {
		image = "quay.io/stackdome/cluster-agent/cluster-agent-manager:" + imageTag
	}

	replicas := int32(1)
	runAsNonRoot := true
	allowPrivilegeEscalation := false
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stackdome-operator-manager",
			Namespace: "stackdome-control-plane",
			Labels: map[string]string{
				"app.kubernetes.io/name": "stackdome-operator",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "stackdome-operator",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "stackdome-operator",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "stackdome-operator",
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: &runAsNonRoot,
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "manager",
							Image: image,
							Args: []string{
								"--leader-elect",
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromInt(8081),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readyz",
										Port: intstr.FromInt(8081),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}

// GetCRDsFromGitHub returns all CRD YAML content as a map fetched from GitHub
func GetCRDsFromGitHub(ctx context.Context, version, cacheDir string) (map[string][]byte, error) {
	if version == "" {
		version = GetClusterAgentVersion()
	}

	loader := &githubCRDLoader{
		version:  version,
		cacheDir: cacheDir,
		client:   createGitHubClient(ctx),
	}

	crds := make(map[string][]byte)
	crdFiles := []string{
		"addons.stackdome.io_postgresclusters.yaml",
		"builds.stackdome.io_imagebuilds.yaml",
		"core.stackdome.io_stackresources.yaml",
		"core.stackdome.io_stacks.yaml",
		"registry.stackdome.io_clusterregistries.yaml",
		"storage.stackdome.io_nfsservers.yaml",
		"storage.stackdome.io_storages.yaml",
		"storage.stackdome.io_volumes.yaml",
		"users.stackdome.io_users.yaml",
	}

	for _, filename := range crdFiles {
		content, err := loader.fetchManifest(ctx, "config/deploy/crds/"+filename)
		if err != nil {
			return nil, fmt.Errorf("fetching CRD %s: %w", filename, err)
		}
		crds[filename] = content
	}

	return crds, nil
}

// getClusterAgentVersion returns the version to use, checking environment variables
func GetClusterAgentVersion() string {
	if version := os.Getenv("CLUSTER_AGENT_VERSION"); version != "" {
		return version
	}
	return defaultClusterAgentVersion
}
