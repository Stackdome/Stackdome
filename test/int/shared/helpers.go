package shared

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

// ShouldSkipCleanup returns true when KEEP_RESOURCES_ON_FAILURE is set and the current spec failed.
// Use inside DeferCleanup callbacks to preserve cluster resources for debugging.
func ShouldSkipCleanup() bool {
	return os.Getenv("KEEP_RESOURCES_ON_FAILURE") == "true" && ginkgo.CurrentSpecReport().Failed()
}

// DeferResourceCleanup wraps ginkgo.DeferCleanup to skip resource deletion when
// KEEP_RESOURCES_ON_FAILURE is set and the current spec failed.
func DeferResourceCleanup(fn func()) {
	ginkgo.DeferCleanup(func() {
		if ShouldSkipCleanup() {
			fmt.Fprintf(ginkgo.GinkgoWriter, "KEEP_RESOURCES_ON_FAILURE: skipping cleanup for failed spec %q\n",
				ginkgo.CurrentSpecReport().FullText())
			return
		}
		fn()
	})
}

// PostgreSQL addon CRUD operations for Ginkgo tests
func CreatePostgresAddon(client *openapi.APIClient, orgID, teamName string, addon *openapi.PostgresAddon) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost(ctx, orgID, teamName).PostgresAddon(*addon).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func GetPostgresAddon(client *openapi.APIClient, orgID, teamName, addonID string) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(ctx, orgID, teamName, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func ListPostgresAddons(client *openapi.APIClient, orgID, teamName string) *openapi.PostgresAddonList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresGet(ctx, orgID, teamName).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list postgres addons")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon list response")

	return resp
}

func UpdatePostgresAddon(client *openapi.APIClient, orgID, teamName, addonID string, addon *openapi.PostgresAddon) *openapi.PostgresAddon {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut(ctx, orgID, teamName, addonID).PostgresAddon(*addon).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected postgres addon response")

	return resp
}

func DeletePostgresAddon(client *openapi.APIClient, orgID, teamName, addonID string) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdDelete(ctx, orgID, teamName, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete postgres addon")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
}

// Error testing helpers for Ginkgo
func CreatePostgresAddonExpectError(client *openapi.APIClient, orgID, teamName string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost(ctx, orgID, teamName).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdatePostgresAddonExpectError(client *openapi.APIClient, orgID, teamName, addonID string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut(ctx, orgID, teamName, addonID).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func GetPostgresAddonExpectError(client *openapi.APIClient, orgID, teamName, addonID string, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(ctx, orgID, teamName, addonID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

// Secret CRUD operations for Ginkgo tests

func CreateSecret(client *openapi.APIClient, orgID, teamName string, secret *openapi.Secret) *openapi.Secret {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsPost(ctx, orgID, teamName).Secret(*secret).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create secret")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected secret response")

	return resp
}

func GetSecret(client *openapi.APIClient, orgID, teamName, secretID string) *openapi.Secret {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdGet(ctx, orgID, teamName, secretID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get secret")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected secret response")

	return resp
}

func DeleteSecret(client *openapi.APIClient, orgID, teamName, secretID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete(ctx, orgID, teamName, secretID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete secret")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent), "unexpected status code")
}

// ObjectStore CRUD operations for Ginkgo tests

func CreateObjectStore(client *openapi.APIClient, orgID, teamName string, store *openapi.ObjectStore) *openapi.ObjectStore {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost(ctx, orgID, teamName).ObjectStore(*store).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create object store")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected object store response")

	return resp
}

func GetObjectStore(client *openapi.APIClient, orgID, teamName, storeID string) *openapi.ObjectStore {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet(ctx, orgID, teamName, storeID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get object store")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected object store response")

	return resp
}

func ListObjectStores(client *openapi.APIClient, orgID, teamName string) *openapi.ObjectStoreList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresGet(ctx, orgID, teamName).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list object stores")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected object store list response")

	return resp
}

func UpdateObjectStore(client *openapi.APIClient, orgID, teamName, storeID string, store *openapi.ObjectStore) *openapi.ObjectStore {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut(ctx, orgID, teamName, storeID).ObjectStore(*store).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update object store")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected object store response")

	return resp
}

func DeleteObjectStore(client *openapi.APIClient, orgID, teamName, storeID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdDelete(ctx, orgID, teamName, storeID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete object store")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent), "unexpected status code")
}

// Error testing helpers for ObjectStore

func CreateObjectStoreExpectError(client *openapi.APIClient, orgID, teamName string, store *openapi.ObjectStore, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresPost(ctx, orgID, teamName).ObjectStore(*store).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func GetObjectStoreExpectError(client *openapi.APIClient, orgID, teamName, storeID string, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet(ctx, orgID, teamName, storeID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdateObjectStoreExpectError(client *openapi.APIClient, orgID, teamName, storeID string, store *openapi.ObjectStore, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut(ctx, orgID, teamName, storeID).ObjectStore(*store).Execute()
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

// Stack CRUD operations for Ginkgo tests

func CreateStack(client *openapi.APIClient, orgID, teamName string, stack *openapi.Stack) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(ctx, orgID, teamName).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

func GetStack(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdGet(ctx, orgID, teamName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

func GetStackTopology(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.StackTopology {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet(ctx, orgID, teamName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get stack topology")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack topology response")

	return resp
}

func ListStackConnections(client *openapi.APIClient, orgID, teamName, stackID string) []openapi.StackConnection {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet(ctx, orgID, teamName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list stack connections")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack connection list response")

	items := resp.GetItems()
	if items == nil {
		return []openapi.StackConnection{}
	}
	return items
}

func CreateStackConnection(client *openapi.APIClient, orgID, teamName, stackID string, connection *openapi.StackConnection) *openapi.StackConnection {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost(ctx, orgID, teamName, stackID).
		StackConnection(*connection).
		Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create stack connection")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack connection response")

	return resp
}

func UpdateStackConnection(client *openapi.APIClient, orgID, teamName, stackID, connectionID string, connection *openapi.StackConnection) *openapi.StackConnection {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut(ctx, orgID, teamName, stackID, connectionID).
		StackConnection(*connection).
		Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update stack connection")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack connection response")

	return resp
}

func DeleteStackConnection(client *openapi.APIClient, orgID, teamName, stackID, connectionID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete(ctx, orgID, teamName, stackID, connectionID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete stack connection")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent), "unexpected status code")
}

func ListStacks(client *openapi.APIClient, orgID, teamName string) *openapi.StackList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet(ctx, orgID, teamName).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list stacks")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack list response")

	return resp
}

func UpdateStack(client *openapi.APIClient, orgID, teamName, stackID string, stack *openapi.Stack) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

func DeleteStack(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete(ctx, orgID, teamName, stackID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusAccepted), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

func CreateStackExpectError(client *openapi.APIClient, orgID, teamName string, stack *openapi.Stack, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(ctx, orgID, teamName).Stack(*stack).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdateStackExpectError(client *openapi.APIClient, orgID, teamName, stackID string, stack *openapi.Stack, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func CreateStackConnectionExpectError(client *openapi.APIClient, orgID, teamName, stackID string, connection *openapi.StackConnection, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost(ctx, orgID, teamName, stackID).
		StackConnection(*connection).
		Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdateStackConnectionExpectError(client *openapi.APIClient, orgID, teamName, stackID, connectionID string, connection *openapi.StackConnection, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdPut(ctx, orgID, teamName, stackID, connectionID).
		StackConnection(*connection).
		Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	apiErr, ok := err.(*openapi.GenericOpenAPIError)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func DeleteStackConnectionExpectError(client *openapi.APIClient, orgID, teamName, stackID, connectionID string, expectedStatus int) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete(ctx, orgID, teamName, stackID, connectionID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")
}

func GetStackTopologyExpectError(client *openapi.APIClient, orgID, teamName, stackID string, expectedStatus int) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdTopologyGet(ctx, orgID, teamName, stackID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")
}

func ListStackConnectionsExpectError(client *openapi.APIClient, orgID, teamName, stackID string, expectedStatus int) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsGet(ctx, orgID, teamName, stackID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")
}
