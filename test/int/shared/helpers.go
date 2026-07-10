package shared

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

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

// CreateGitCredentialsIntegration registers an org-level git_credentials
// integration (basic username/password). The credential resolver auto-attaches
// it to clones whose host matches, so private-repo builds authenticate without
// any per-source credential input.
func CreateGitCredentialsIntegration(client *openapi.APIClient, orgID, host, username, password string) *openapi.GitIntegration {
	ctx := context.Background()
	gi := openapi.NewGitIntegration(host)
	gi.SetType(openapi.GIT_INTEGRATION_TYPE_CREDENTIALS)
	auth := openapi.NewGitIntegrationAuth()
	auth.SetBasic(*openapi.NewGitIntegrationBasicAuth(username, password))
	gi.SetAuth(*auth)

	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsPost(ctx, orgID).GitIntegration(*gi).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create git integration")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected git integration response")

	return resp
}

func DeleteGitIntegration(client *openapi.APIClient, orgID, integrationID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdGitIntegrationsIdDelete(ctx, orgID, integrationID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete git integration")
	Expect(httpResp.StatusCode).To(BeElementOf(http.StatusOK, http.StatusNoContent), "unexpected status code")
}

// Error testing helpers for Ginkgo
func CreatePostgresAddonExpectError(client *openapi.APIClient, orgID, teamName string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresPost(ctx, orgID, teamName).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdatePostgresAddonExpectError(client *openapi.APIClient, orgID, teamName, addonID string, addon *openapi.PostgresAddon, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdPut(ctx, orgID, teamName, addonID).PostgresAddon(*addon).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func GetPostgresAddonExpectError(client *openapi.APIClient, orgID, teamName, addonID string, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(ctx, orgID, teamName, addonID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
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
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	Expect(err).NotTo(HaveOccurred(), "failed to delete secret")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent), "unexpected status code")
}

func DeleteSecretRaw(client *openapi.APIClient, orgID, teamName, secretID string) (*http.Response, error) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsIdDelete(ctx, orgID, teamName, secretID).Execute()
	return httpResp, err
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

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func GetObjectStoreExpectError(client *openapi.APIClient, orgID, teamName, storeID string, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdGet(ctx, orgID, teamName, storeID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdateObjectStoreExpectError(client *openapi.APIClient, orgID, teamName, storeID string, store *openapi.ObjectStore, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameObjectStoresIdPut(ctx, orgID, teamName, storeID).ObjectStore(*store).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
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

	// Thin create: server creates a shell, ignoring inline children.
	postResp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(ctx, orgID, teamName).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(postResp).NotTo(BeNil(), "expected stack response")

	// Fat apply: submit the same full stack so children are declared.
	applied, httpResp, err := client.DefaultApi.ApplyStack(ctx, orgID, teamName, postResp.GetId()).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to apply stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(applied).NotTo(BeNil(), "expected applied stack response")

	return applied
}

func CreateShellStack(client *openapi.APIClient, orgID, teamName string, stack *openapi.Stack) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksPost(ctx, orgID, teamName).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create shell stack")
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
	resp, httpResp, err := client.DefaultApi.ApplyStack(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to apply stack")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

// UpdateStackShellH updates a stack via the shell-only PUT endpoint. The server
// updates shell fields (name, labels, etc.) and ignores any inline children, so
// existing resources/connections are preserved.
func UpdateStackShellH(client *openapi.APIClient, orgID, teamName, stackID string, stack *openapi.Stack) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update stack shell")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack response")

	return resp
}

func DeleteStack(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.Stack {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdDelete(ctx, orgID, teamName, stackID).Execute()
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return nil
	}
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

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func UpdateStackExpectError(client *openapi.APIClient, orgID, teamName, stackID string, stack *openapi.Stack, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdPut(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func ApplyStackExpectError(client *openapi.APIClient, orgID, teamName, stackID string, stack *openapi.Stack, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApplyStack(ctx, orgID, teamName, stackID).Stack(*stack).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

// Thin stack sub-resource operations for Ginkgo tests

func UpdateStackResourceH(client *openapi.APIClient, orgID, teamName, stackID, resourceName string, resource *openapi.StackResource) *openapi.StackResource {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.UpdateStackResource(ctx, orgID, teamName, stackID, resourceName).StackResource(*resource).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to update stack resource")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack resource response")

	return resp
}

func DeleteStackResourceH(client *openapi.APIClient, orgID, teamName, stackID, resourceName string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.DeleteStackResource(ctx, orgID, teamName, stackID, resourceName).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to delete stack resource")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK), "unexpected status code")
}

func CreateStackConnectionExpectError(client *openapi.APIClient, orgID, teamName, stackID string, connection *openapi.StackConnection, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsPost(ctx, orgID, teamName, stackID).
		StackConnection(*connection).
		Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
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

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

func DeleteStackConnectionExpectError(client *openapi.APIClient, orgID, teamName, stackID, connectionID string, expectedStatus int) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksIdConnectionsConnectionIdDelete(ctx, orgID, teamName, stackID, connectionID).Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")
}

// CreateStackResource adds a resource to an existing stack via the thin
// resource-create endpoint (POST /stacks/{id}/resources).
func CreateStackResource(client *openapi.APIClient, orgID, teamName, stackID string, resource *openapi.StackResource) *openapi.StackResource {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.CreateStackResource(ctx, orgID, teamName, stackID).
		StackResource(*resource).
		Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create stack resource")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated), "unexpected status code")
	Expect(resp).NotTo(BeNil(), "expected stack resource response")

	return resp
}

func CreateStackResourceExpectError(client *openapi.APIClient, orgID, teamName, stackID string, resource *openapi.StackResource, expectedStatus int) *openapi.GenericOpenAPIError {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.CreateStackResource(ctx, orgID, teamName, stackID).
		StackResource(*resource).
		Execute()
	Expect(err).To(HaveOccurred(), "expected error")
	Expect(httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	Expect(ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
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

// Release CRUD operations for Ginkgo tests

func CreateRelease(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.StackRelease {
	release, httpResp, err := client.ReleasesApi.CreateRelease(
		context.Background(), orgID, teamName, stackID,
	).CreateReleaseRequest(openapi.CreateReleaseRequest{}).Execute()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to create release, status: %d", httpResp.StatusCode)
	return release
}

func RollbackRelease(client *openapi.APIClient, orgID, teamName, stackID, fromReleaseID string) *openapi.StackRelease {
	release, httpResp, err := client.ReleasesApi.CreateRelease(
		context.Background(), orgID, teamName, stackID,
	).CreateReleaseRequest(openapi.CreateReleaseRequest{
		FromReleaseId: openapi.PtrString(fromReleaseID),
	}).Execute()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to rollback release, status: %d", httpResp.StatusCode)
	return release
}

func ListReleases(client *openapi.APIClient, orgID, teamName, stackID string) *openapi.StackReleaseList {
	list, httpResp, err := client.ReleasesApi.ListReleases(
		context.Background(), orgID, teamName, stackID,
	).Execute()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to list releases, status: %d", httpResp.StatusCode)
	return list
}

func GetRelease(client *openapi.APIClient, orgID, teamName, stackID, releaseID string) *openapi.StackReleaseDetail {
	release, httpResp, err := client.ReleasesApi.GetRelease(
		context.Background(), orgID, teamName, stackID, releaseID,
	).Execute()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to get release, status: %d", httpResp.StatusCode)
	return release
}

// CreateReleaseExpectError creates a release expecting the request to be rejected
// (e.g. sync validation failures surfaced as 400 from resolvePins).
func CreateReleaseExpectError(client *openapi.APIClient, orgID, teamName, stackID string, expectedStatus int) *openapi.GenericOpenAPIError {
	_, httpResp, err := client.ReleasesApi.CreateRelease(
		context.Background(), orgID, teamName, stackID,
	).CreateReleaseRequest(openapi.CreateReleaseRequest{}).Execute()
	ExpectWithOffset(1, err).To(HaveOccurred(), "expected error creating release")
	ExpectWithOffset(1, httpResp.StatusCode).To(Equal(expectedStatus), "unexpected status code")

	var apiErr *openapi.GenericOpenAPIError
	ok := errors.As(err, &apiErr)
	ExpectWithOffset(1, ok).To(BeTrue(), "expected GenericOpenAPIError")

	return apiErr
}

// ErrorValidationCodes decodes an aggregated validation error's body and
// returns the codes carried in details.errors[].code. Works for any endpoint
// that returns errors.ValidationFailed (details: {"errors": [{field, code, message}]}),
// whether or not the generated client populated apiErr.Model() for the route.
func ErrorValidationCodes(apiErr *openapi.GenericOpenAPIError) []string {
	var errObj openapi.Error
	ExpectWithOffset(1, json.Unmarshal(apiErr.Body(), &errObj)).To(Succeed(), "failed to decode error response body")

	raw, ok := errObj.Details["errors"]
	ExpectWithOffset(1, ok).To(BeTrue(), "expected details.errors in error response, got: %+v", errObj.Details)

	entries, ok := raw.([]interface{})
	ExpectWithOffset(1, ok).To(BeTrue(), "expected details.errors to be an array")

	codes := make([]string, 0, len(entries))
	for _, e := range entries {
		m, ok := e.(map[string]interface{})
		ExpectWithOffset(1, ok).To(BeTrue(), "expected each details.errors entry to be an object")
		code, _ := m["code"].(string)
		codes = append(codes, code)
	}
	return codes
}

func CancelRelease(client *openapi.APIClient, orgID, teamName, stackID, releaseID string) {
	httpResp, err := client.ReleasesApi.CancelRelease(
		context.Background(), orgID, teamName, stackID, releaseID,
	).Execute()
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to cancel release, status: %d", httpResp.StatusCode)
}

func WaitForReleaseState(client *openapi.APIClient, orgID, teamName, stackID, releaseID, expectedState string, timeout time.Duration) *openapi.StackReleaseDetail {
	var result *openapi.StackReleaseDetail
	Eventually(func(g Gomega) {
		release := GetRelease(client, orgID, teamName, stackID, releaseID)
		g.Expect(string(release.GetState())).To(Equal(expectedState),
			"release state should be %s, got %s (message: %s)", expectedState, release.GetState(), release.GetMessage())
		result = release
	}, timeout, 5*time.Second).Should(Succeed())
	return result
}

func WaitForReleaseReleased(client *openapi.APIClient, orgID, teamName, stackID, releaseID string, timeout time.Duration) *openapi.StackReleaseDetail {
	return WaitForReleaseState(client, orgID, teamName, stackID, releaseID, string(models.ReleaseStateReleased), timeout)
}

// CreateStackAndDeploy creates a stack and immediately creates a release to deploy it.
// Use this instead of CreateStack when the test expects the stack to reach Ready state.
func CreateStackAndDeploy(client *openapi.APIClient, orgID, teamName string, stack *openapi.Stack) (*openapi.Stack, *openapi.StackRelease) {
	created := CreateStack(client, orgID, teamName, stack)
	release := CreateRelease(client, orgID, teamName, created.GetId())
	return created, release
}
