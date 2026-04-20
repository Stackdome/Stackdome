package shared

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"

	// postgres driver for connectivity check
	_ "github.com/lib/pq"
)

// WaitForCRExists polls the cluster until a PostgresCluster CR exists with the given name and namespace.
func WaitForCRExists(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) *addonsv1alpha1.PostgresCluster {
	var cr addonsv1alpha1.PostgresCluster
	Eventually(func(g Gomega) {
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).NotTo(HaveOccurred(), "PostgresCluster CR should exist")
	}, timeout, 2*time.Second).Should(Succeed())
	return &cr
}

// WaitForAddonReady polls the API until the addon status is "Ready".
func WaitForAddonReady(apiClient *openapi.APIClient, orgID, addonID string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue(), "addon should have status")

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue(), "status should have state")
		g.Expect(*state).To(Equal("Ready"), "addon should be Ready, got: %s", *state)
		addon = resp
	}, timeout, 5*time.Second).Should(Succeed())
	return addon
}

// WaitForAddonState polls the API until the addon status matches the expected state.
func WaitForAddonState(apiClient *openapi.APIClient, orgID, addonID, expectedState string, timeout time.Duration) *openapi.PostgresAddon {
	var addon *openapi.PostgresAddon
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(200))

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue())

		state, stateOk := status.GetStateOk()
		g.Expect(stateOk).To(BeTrue())
		g.Expect(*state).To(Equal(expectedState))
		addon = resp
	}, timeout, 5*time.Second).Should(Succeed())
	return addon
}

// WaitForConditionTrue polls the API until a specific condition on the addon is True.
func WaitForConditionTrue(apiClient *openapi.APIClient, orgID, addonID, conditionType string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		resp, _, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).NotTo(HaveOccurred())

		status, ok := resp.GetStatusOk()
		g.Expect(ok).To(BeTrue())

		conditions, condOk := status.GetConditionsOk()
		g.Expect(condOk).To(BeTrue())

		found := false
		for _, c := range conditions {
			if c.GetType() == conditionType && c.GetStatus() == "True" {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(), "condition %s should be True", conditionType)
	}, timeout, 5*time.Second).Should(Succeed())
}

// WaitForAddonDeleted polls the API until the addon returns 404.
func WaitForAddonDeleted(apiClient *openapi.APIClient, orgID, addonID string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		ctx := context.Background()
		_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdGet(ctx, orgID, addonID).Execute()
		g.Expect(err).To(HaveOccurred())
		g.Expect(httpResp.StatusCode).To(Equal(404))
	}, timeout, 5*time.Second).Should(Succeed())
}

// WaitForCRDeleted polls the cluster until the PostgresCluster CR no longer exists.
func WaitForCRDeleted(ctx context.Context, clusterClient client.Client, name, namespace string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		var cr addonsv1alpha1.PostgresCluster
		err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr)
		g.Expect(err).To(HaveOccurred(), "PostgresCluster CR should be deleted")
	}, timeout, 2*time.Second).Should(Succeed())
}

// GetPostgresClusterCR retrieves the PostgresCluster CR from the cluster.
func GetPostgresClusterCR(ctx context.Context, clusterClient client.Client, name, namespace string) (*addonsv1alpha1.PostgresCluster, error) {
	var cr addonsv1alpha1.PostgresCluster
	if err := clusterClient.Get(ctx, client.ObjectKey{Name: name, Namespace: namespace}, &cr); err != nil {
		return nil, err
	}
	return &cr, nil
}

// VerifyCRSpec asserts that the PostgresCluster CR spec matches expected values.
func VerifyCRSpec(cr *addonsv1alpha1.PostgresCluster, expectedInstances int, expectedMajorVersion int, expectedStorage string) {
	Expect(cr.Spec.Instances).To(Equal(expectedInstances), "CR instances mismatch")
	Expect(cr.Spec.PostgreSQLSpec).NotTo(BeNil(), "PostgreSQLSpec should be set")
	Expect(cr.Spec.PostgreSQLSpec.PostgreSQLMajorVersion).To(Equal(expectedMajorVersion), "CR PostgreSQL major version mismatch")
	Expect(cr.Spec.StorageSpec).NotTo(BeNil(), "StorageSpec should be set")
	Expect(cr.Spec.StorageSpec.Size).To(Equal(expectedStorage), "CR storage size mismatch")
}

// VerifyCRLabel checks that the CR has the expected addon ID label.
func VerifyCRLabel(cr *addonsv1alpha1.PostgresCluster, addonID string) {
	labels := cr.GetLabels()
	Expect(labels).To(HaveKeyWithValue(models.PostgresAddonIDLabel, addonID), "CR should have addon ID label")
}

// GetCredentials fetches JIT credentials for an addon database via the API.
func GetCredentials(apiClient *openapi.APIClient, orgID, addonID, database string) *openapi.PostgresCredentials {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdCredentialsDatabaseGet(ctx, orgID, addonID, database).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get credentials")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp
}

// ConnectToPostgres opens a real PostgreSQL connection and runs SELECT 1 to verify connectivity.
// TODO: Exercise this in an e2e test (see docs/plans/postgres-addon-e2e-tests-enhancement.md #4).
func ConnectToPostgres(host string, port int32, username, password, dbName, sslMode string) *sql.DB {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, username, password, dbName, sslMode)

	db, err := sql.Open("postgres", connStr)
	Expect(err).NotTo(HaveOccurred(), "failed to open postgres connection")

	err = db.Ping()
	Expect(err).NotTo(HaveOccurred(), "failed to ping postgres")

	return db
}

// CRNameForAddon returns the expected PostgresCluster CR name for a given addon.
// The CR name matches the addon name directly (set in postgres_cluster_builder.go).
func CRNameForAddon(addonName string) string {
	return addonName
}

// TriggerBackup triggers an immediate backup for a postgres addon.
func TriggerBackup(apiClient *openapi.APIClient, orgID, addonID string) {
	ctx := context.Background()
	req := openapi.NewApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPostRequest()
	req.SetDescription("e2e test backup")
	_, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPost(ctx, orgID, addonID).
		ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPostRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to trigger backup")
	Expect(httpResp.StatusCode).To(Equal(202))
}

// ListBackups returns all backups for a postgres addon.
func ListBackups(apiClient *openapi.APIClient, orgID, addonID string) []openapi.PostgresBackup {
	ctx := context.Background()
	resp, httpResp, err := apiClient.DefaultApi.ApiV1OrganizationsOrgIdAddonsPostgresIdBackupsGet(ctx, orgID, addonID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list backups")
	Expect(httpResp.StatusCode).To(Equal(200))
	Expect(resp).NotTo(BeNil())
	return resp.GetItems()
}

// WaitForBackupPhase polls until at least one backup reaches the expected phase.
func WaitForBackupPhase(apiClient *openapi.APIClient, orgID, addonID, expectedPhase string, timeout time.Duration) {
	Eventually(func(g Gomega) {
		backups := ListBackups(apiClient, orgID, addonID)
		g.Expect(len(backups)).To(BeNumerically(">=", 1), "should have at least one backup")

		found := false
		for _, b := range backups {
			if b.GetPhase() == expectedPhase {
				found = true
				break
			}
		}
		g.Expect(found).To(BeTrue(), "no backup in phase %s", expectedPhase)
	}, timeout, 10*time.Second).Should(Succeed())
}
