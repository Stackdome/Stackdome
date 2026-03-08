package shared

import (
	"context"
	"net/http"

	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

// PostgreSQL addon CRUD operations for Ginkgo tests
func CreatePostgresAddon(client *openapi.APIClient, orgID string, addon *openapi.PostgresAddon) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresPost(ctx, orgID).PostgresAddon(*addon).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func GetPostgresAddon(client *openapi.APIClient, orgID, addonID string) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func ListPostgresAddons(client *openapi.APIClient, orgID string) *openapi.PostgresAddonList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresGet(ctx, orgID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list postgres addons")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon list response")

	return resp
}

func UpdatePostgresAddon(client *openapi.APIClient, orgID, addonID string, addon *openapi.PostgresAddon) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdPut(ctx, orgID, addonID).PostgresAddon(*addon).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func DeletePostgresAddon(client *openapi.APIClient, orgID, addonID string) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdDelete(ctx, orgID, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent), "unexpected status code")
}

// Error testing helpers for Ginkgo
func CreatePostgresAddonExpectError(client *openapi.APIClient, orgID string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresPost(ctx, orgID).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdatePostgresAddonExpectError(client *openapi.APIClient, orgID, addonID string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdPut(ctx, orgID, addonID).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func GetPostgresAddonExpectError(client *openapi.APIClient, orgID, addonID string, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

// Assertion helpers for OpenAPI models using Ginkgo matchers
func ExpectPostgresAddonEqual(expected, actual *openapi.PostgresAddon) {
	Expect(actual.GetName()).To(Equal(expected.GetName()), "addon name mismatch")

	// Compare version
	Expect(actual.Spec.Version.GetMajor()).To(Equal(expected.Spec.Version.GetMajor()), "version major mismatch")

	// Compare instances
	Expect(actual.Spec.Instances.GetCount()).To(Equal(expected.Spec.Instances.GetCount()), "instance count mismatch")

	// Compare storage
	Expect(actual.Spec.Storage.GetSize()).To(Equal(expected.Spec.Storage.GetSize()), "storage size mismatch")
	Expect(actual.Spec.Storage.GetStorageClass()).To(Equal(expected.Spec.Storage.GetStorageClass()), "storage class mismatch")

	// Compare resources if present
	if expected.Spec.HasResources() && actual.Spec.HasResources() {
		expectedRes := expected.Spec.GetResources()
		actualRes := actual.Spec.GetResources()

		if expectedRes.HasCpu() && actualRes.HasCpu() {
			Expect(actualRes.Cpu.GetRequest()).To(Equal(expectedRes.Cpu.GetRequest()), "CPU request mismatch")
			Expect(actualRes.Cpu.GetLimit()).To(Equal(expectedRes.Cpu.GetLimit()), "CPU limit mismatch")
		}

		if expectedRes.HasMemory() && actualRes.HasMemory() {
			Expect(actualRes.Memory.GetRequest()).To(Equal(expectedRes.Memory.GetRequest()), "memory request mismatch")
			Expect(actualRes.Memory.GetLimit()).To(Equal(expectedRes.Memory.GetLimit()), "memory limit mismatch")
		}
	}

	// Compare databases if present
	if expected.Spec.HasDatabases() && actual.Spec.HasDatabases() {
		expectedDBs := expected.Spec.GetDatabases()
		actualDBs := actual.Spec.GetDatabases()

		Expect(len(actualDBs)).To(Equal(len(expectedDBs)), "database count mismatch")

		// Create maps for easier comparison
		expectedDBMap := make(map[string]openapi.PostgresDatabase)
		actualDBMap := make(map[string]openapi.PostgresDatabase)

		for _, db := range expectedDBs {
			expectedDBMap[db.GetName()] = db
		}

		for _, db := range actualDBs {
			actualDBMap[db.GetName()] = db
		}

		for name, expectedDB := range expectedDBMap {
			actualDB, found := actualDBMap[name]
			Expect(found).To(BeTrue(), "database %s not found in actual", name)
			Expect(actualDB.GetExtensions()).To(Equal(expectedDB.GetExtensions()), "database %s extensions mismatch", name)
		}
	}
}
