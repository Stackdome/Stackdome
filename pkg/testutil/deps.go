package testutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

const (
	// Stackdome agent Helm chart (bundles cluster-agent + all dependencies)
	DefaultChartVersion = "0.6.5-alpha"
	ChartRepo           = "oci://quay.io/stackdome/charts/stackdome-agent"
	ChartReleaseName    = "stackdome-agent"
	ChartNamespace      = "stackdome-control-plane"

	// MinIO (S3-compatible object storage for backup tests)
	MinIOImage       = "minio/minio:latest"
	MinIOClientImage = "minio/mc:latest"
	MinIONamespace   = "stackdome-system"
	MinIOName        = "minio"
	MinIOServicePort = 9000
	MinIOAccessKey   = "minioadmin"
	MinIOSecretKey   = "minioadmin"
	MinIOBucket      = "backups"
)

// MinIOEndpoint returns the in-cluster endpoint for the MinIO service.
func MinIOEndpoint() string {
	return fmt.Sprintf("http://%s.%s.svc:%d", MinIOName, MinIONamespace, MinIOServicePort)
}

const (
	HelmVersion = "v3.14.0"
)

// GetChartVersion returns the stackdome-agent chart version, checking environment variables.
func GetChartVersion() string {
	if version := os.Getenv("STACKDOME_CHART_VERSION"); version != "" {
		return version
	}
	return DefaultChartVersion
}

func cacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = os.Getenv("HOME")
	}
	return filepath.Join(homeDir, ".cache", "stackdome-api-server")
}

// EnsureHelm returns the path to a working helm binary, installing it if necessary.
func EnsureHelm(ctx context.Context) (string, error) {
	if p, err := exec.LookPath("helm"); err == nil {
		return p, nil
	}

	binDir := filepath.Join(cacheDir(), "bin")
	helmPath := filepath.Join(binDir, "helm")
	if _, err := os.Stat(helmPath); err == nil {
		return helmPath, nil
	}

	if err := installHelm(ctx, binDir); err != nil {
		return "", err
	}
	return helmPath, nil
}

func installHelm(ctx context.Context, binDir string) error {
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	cache := cacheDir()
	archiveName := fmt.Sprintf("helm-%s-%s-%s.tar.gz", HelmVersion, runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("https://get.helm.sh/%s", archiveName)
	tempFile := filepath.Join(cache, archiveName)

	dl := exec.CommandContext(ctx, "curl", "-L", "-o", tempFile, url)
	if output, err := dl.CombinedOutput(); err != nil {
		return fmt.Errorf("downloading helm: %w\n%s", err, output)
	}
	defer os.Remove(tempFile)

	extract := exec.CommandContext(ctx, "tar", "-xzf", tempFile, "-C", cache,
		fmt.Sprintf("%s-%s/helm", runtime.GOOS, runtime.GOARCH))
	if output, err := extract.CombinedOutput(); err != nil {
		return fmt.Errorf("extracting helm: %w\n%s", err, output)
	}

	src := filepath.Join(cache, fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH), "helm")
	dst := filepath.Join(binDir, "helm")
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("moving helm to bin: %w", err)
	}
	os.RemoveAll(filepath.Join(cache, fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)))

	return nil
}
