package testutil

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/mt-sre/devkube/dev"
	"golang.org/x/oauth2"
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
		client, _ := github.NewClient()
		return client
	}

	// Create authenticated client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client, _ := github.NewClient(github.WithHTTPClient(tc))
	return client
}

// checkTagExists verifies if a tag exists in the repository
func (g *githubLoader) checkTagExists(ctx context.Context) (bool, error) {
	// List tags
	tags, _, err := g.client.Repositories.ListTags(ctx, g.owner, g.repo, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
		// Try to list releases as an alternative
		releases, _, releaseErr := g.client.Repositories.ListReleases(ctx, g.owner, g.repo, &github.ListOptions{
			PerPage: 100,
		})
		if releaseErr != nil {
			return false, fmt.Errorf("listing tags failed: %w, listing releases failed: %w", err, releaseErr)
		}

		// Check releases
		for _, release := range releases {
			if release.TagName != nil && *release.TagName == g.tag {
				return true, nil
			}
		}
		return false, nil
	}

	// Check tags
	for _, t := range tags {
		if t.Name != nil && *t.Name == g.tag {
			return true, nil
		}
	}

	// If tag not found, also check if it exists as a branch
	_, _, err = g.client.Repositories.GetBranch(ctx, g.owner, g.repo, strings.TrimPrefix(g.tag, "v"), 1)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// findLatestTag finds the latest available tag in the repository
func (g *githubLoader) findLatestTag(ctx context.Context) (string, error) {
	// Try to get the latest release first
	release, _, err := g.client.Repositories.GetLatestRelease(ctx, g.owner, g.repo)
	if err == nil && release.TagName != nil {
		return *release.TagName, nil
	}

	// If no releases, list tags
	tags, _, err := g.client.Repositories.ListTags(ctx, g.owner, g.repo, &github.ListOptions{
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

// githubLoader implements dev.ClusterInitializer to load manifests from GitHub
type githubLoader struct {
	repo        string
	owner       string
	tag         string
	cacheDir    string
	client      *github.Client
	pathsToLoad []string
}

func WithRepoOwner(owner string) func(*githubLoader) {
	return func(g *githubLoader) {
		g.owner = owner
	}
}

func WithRepoName(repo string) func(*githubLoader) {
	return func(g *githubLoader) {
		g.repo = repo
	}
}

func WithRepoTag(tag string) func(*githubLoader) {
	return func(g *githubLoader) {
		g.tag = tag
	}
}

func WithPathsToLoad(paths []string) func(*githubLoader) {
	return func(g *githubLoader) {
		g.pathsToLoad = paths
	}
}

func WithCacheDir(cacheDir string) func(*githubLoader) {
	return func(g *githubLoader) {
		g.cacheDir = cacheDir
	}
}

func NewGitHubLoader(ctx context.Context, options ...func(*githubLoader)) *githubLoader {
	loader := &githubLoader{
		client: createGitHubClient(ctx),
	}

	for _, option := range options {
		option(loader)
	}

	return loader
}

func (g *githubLoader) Init(ctx context.Context, cluster *dev.Cluster) error {
	// Create GitHub client if not already created
	if g.client == nil {
		return fmt.Errorf("GitHub client not initialized")
	}

	// Validate required fields
	if g.owner == "" {
		return fmt.Errorf("repository owner not specified")
	}
	if g.repo == "" {
		return fmt.Errorf("repository name not specified")
	}
	if g.tag == "" {
		return fmt.Errorf("repository tag not specified")
	}

	if len(g.pathsToLoad) == 0 {
		return fmt.Errorf("no paths specified to load from GitHub")
	}

	// Verify tag exists
	exists, err := g.checkTagExists(ctx)
	if err != nil {
		return fmt.Errorf("checking if tag %s exists: %w", g.tag, err)
	}
	if !exists {
		// Try to find the latest available tag
		latestTag, err := g.findLatestTag(ctx)
		if err != nil {
			return fmt.Errorf("tag %s not found and unable to find latest tag: %w", g.tag, err)
		}
		return fmt.Errorf("tag %s not found in repository, latest available tag is %s", g.tag, latestTag)
	}

	// Prepare temporary directory for CRD files
	tempDir := filepath.Join(g.cacheDir, "temp-files", g.tag)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}

	directoriesToLoad := []string{}
	filesToLoad := []string{}

	// If the path ends with a slash, treat it as a directory, otherwise as a file
	for _, path := range g.pathsToLoad {
		if strings.HasSuffix(path, "/") {
			directoriesToLoad = append(directoriesToLoad, path)
		} else {
			filesToLoad = append(filesToLoad, path)
		}
	}

	directoryContentsToFetch, err := g.fetchDirectoryContents(ctx, directoriesToLoad)
	if err != nil {
		return fmt.Errorf("fetching directory contents: %w", err)
	}

	fileContentsToFetch := append(filesToLoad, directoryContentsToFetch...)

	// Download and write CRD files to temporary directory
	var filePaths []string
	for _, path := range fileContentsToFetch {
		manifest, err := g.fetchManifest(ctx, path)
		if err != nil {
			return fmt.Errorf("fetching manifest %s: %w", path, err)
		}
		manifestFilepath := filepath.Join(tempDir, filepath.Base(path))
		if err := os.WriteFile(manifestFilepath, manifest, 0644); err != nil {
			return fmt.Errorf("writing manifest file %s: %w", filepath.Base(path), err)
		}
		filePaths = append(filePaths, manifestFilepath)
	}

	// Use cluster's CreateAndWaitFromFiles method
	if err := cluster.CreateAndWaitFromFiles(ctx, filePaths); err != nil {
		return fmt.Errorf("creating CRDs from github: %w", err)
	}

	return nil
}

func (g *githubLoader) fetchDirectoryContents(ctx context.Context, directories []string) ([]string, error) {
	var allFilePaths []string
	for _, dir := range directories {
		_, contents, _, err := g.client.Repositories.GetContents(
			ctx,
			g.owner,
			g.repo,
			dir,
			&github.RepositoryContentGetOptions{
				Ref: g.tag,
			})
		if err != nil {
			return nil, fmt.Errorf("listing directory contents for %s: %w", dir, err)
		}

		for _, content := range contents {
			fileName := content.GetName()
			if strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml") {
				allFilePaths = append(allFilePaths, dir+fileName)
			}
		}
	}
	return allFilePaths, nil
}

// FetchManifest fetches a single file from the configured GitHub repository.
func (g *githubLoader) FetchManifest(ctx context.Context, path string) ([]byte, error) {
	return g.fetchManifest(ctx, path)
}

func (g *githubLoader) fetchManifest(ctx context.Context, path string) ([]byte, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s/%s", g.tag, path)
	if cached, ok := manifestCache[cacheKey]; ok {
		return cached, nil
	}

	// Create GitHub client if not already created
	if g.client == nil {
		g.client = createGitHubClient(ctx)
	}

	// Try GitHub API first
	fileContent, _, resp, err := g.client.Repositories.GetContents(ctx, g.owner, g.repo, path, &github.RepositoryContentGetOptions{
		Ref: g.tag,
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
			cacheFile := filepath.Join(g.cacheDir, "manifests", g.tag, path)
			if err := os.MkdirAll(filepath.Dir(cacheFile), 0755); err == nil {
				_ = os.WriteFile(cacheFile, contentBytes, 0644)
			}
		}

		return contentBytes, nil
	}

	// Return original error
	return nil, fmt.Errorf("fetching file from GitHub: %w: resp error code: %d", err, resp.StatusCode)
}
