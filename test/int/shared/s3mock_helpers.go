package shared

import "github.com/Stackdome/stackdome/pkg/testutil"

// Re-export MinIO constants from testutil for use in test specs and fixtures.
const (
	MinIONamespace   = testutil.MinIONamespace
	MinIOName        = testutil.MinIOName
	MinIOServicePort = testutil.MinIOServicePort
	MinIOAccessKey   = testutil.MinIOAccessKey
	MinIOSecretKey   = testutil.MinIOSecretKey
)

// MinIOEndpoint returns the in-cluster endpoint for the MinIO service.
func MinIOEndpoint() string {
	return testutil.MinIOEndpoint()
}
