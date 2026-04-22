package testutil

import (
	"fmt"
	"os"
)

const (
	// Cluster agent
	clusterAgentVersion                = "v0.4.8-alpha"
	clusterAgentRepo                   = "cluster-agent"
	clusterAgentOwner                  = "ashishmax31"
	clusterAgentImage                  = "quay.io/stackdome/cluster-agent/cluster-agent-manager"
	clusterAgentDeploymentManifestPath = "config/deploy/deployment.yaml"

	// Barman Cloud plugin
	barmanCloudVersion = "v0.5.0"

	// MinIO (S3-compatible object storage for backup tests)
	MinIOImage       = "minio/minio:latest"
	MinIOClientImage = "minio/mc:latest"
	MinIONamespace   = "stackdome-system"
	MinIOName        = "minio"
	MinIOServicePort = 9000
	MinIOAccessKey   = "minioadmin"
	MinIOSecretKey   = "minioadmin"
	MinIOBucket      = "backups"

	// Helm chart repositories
	TraefikChartRepo     = "https://traefik.github.io/charts"
	CNPGChartRepo        = "https://cloudnative-pg.github.io/charts"
	CertManagerChartRepo = "https://charts.jetstack.io"
)

// MinIOEndpoint returns the in-cluster endpoint for the MinIO service.
func MinIOEndpoint() string {
	return fmt.Sprintf("http://%s.%s.svc:%d", MinIOName, MinIONamespace, MinIOServicePort)
}

// GetClusterAgentVersion returns the version to use, checking environment variables
func GetClusterAgentVersion() string {
	if version := os.Getenv("CLUSTER_AGENT_VERSION"); version != "" {
		return version
	}
	return clusterAgentVersion
}

// GetBarmanCloudVersion returns the version to use, checking environment variables
func GetBarmanCloudVersion() string {
	if version := os.Getenv("BARMAN_CLOUD_VERSION"); version != "" {
		return version
	}
	return barmanCloudVersion
}
