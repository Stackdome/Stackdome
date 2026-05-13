package int

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("PostgresAddon E2E", Ordered, func() {
	var client *openapi.APIClient
	var orgID string
	var teamName = models.DefaultTeamName

	BeforeAll(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Basic CR Verification", func() {
		It("should create a PostgresCluster CR in the cluster when addon is created via API", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon via API")
			addon := shared.CreateMinimalPostgresAddon("test-cr-verify")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)

			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			namespace := createdAddon.GetNamespace()

			Expect(addonID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			By("Waiting for the PostgresCluster CR to appear in the cluster")
			crName := shared.CRNameForAddon(addonName)
			cr := shared.WaitForCRExists(ctx, clusterClient, crName, namespace, 2*time.Minute)

			By("Verifying CR spec matches the addon spec")
			shared.VerifyCRSpec(cr, int(addon.Spec.Instances.Count), int(addon.Spec.Version.Major), addon.Spec.Storage.Size)

			By("Verifying CR has the addon ID label")
			shared.VerifyCRLabel(cr, addonID)

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})
		})
	})

	Context("Full Lifecycle", func() {
		It("should create addon, reach Ready, verify CR, and return credentials", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon with a database")
			addon := shared.CreatePostgresAddonWithResources("test-lifecycle")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			namespace := createdAddon.GetNamespace()

			Expect(addonID).NotTo(BeEmpty())
			Expect(namespace).NotTo(BeEmpty())

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for addon to become Ready")
			readyAddon := shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Verifying ConnectionInfo is populated")
			status, ok := readyAddon.GetStatusOk()
			Expect(ok).To(BeTrue())

			connInfo, connOk := status.GetConnectionInfoOk()
			Expect(connOk).To(BeTrue(), "ConnectionInfo should be populated")
			Expect(connInfo.GetHost()).NotTo(BeEmpty(), "Host should be set")
			Expect(connInfo.GetPort()).To(BeNumerically(">", 0), "Port should be positive")

			By("Verifying conditions include ClusterReady")
			shared.WaitForConditionTrue(client, orgID, teamName, addonID, string(models.PostgresAddonConditionClusterReady), 30*time.Second)

			By("Waiting for databases to be applied")
			shared.WaitForConditionTrue(client, orgID, teamName, addonID, string(models.PostgresAddonConditionDatabasesApplied), 2*time.Minute)

			By("Verifying CR exists with correct spec in the cluster")
			cr, err := shared.GetPostgresClusterCR(ctx, clusterClient, shared.CRNameForAddon(addonName), namespace)
			Expect(err).NotTo(HaveOccurred())
			shared.VerifyCRSpec(cr, int(addon.Spec.Instances.Count), int(addon.Spec.Version.Major), addon.Spec.Storage.Size)
			shared.VerifyCRLabel(cr, addonID)

			By("Fetching JIT credentials for the database")
			creds := shared.GetCredentials(client, orgID, teamName, addonID, "testdb")
			Expect(creds.GetHost()).NotTo(BeEmpty())
			Expect(creds.GetPort()).To(BeNumerically(">", 0))
			Expect(creds.GetUsername()).NotTo(BeEmpty())
			Expect(creds.GetPassword()).NotTo(BeEmpty())

			By("Port-forwarding to the primary postgres pod")
			clientset, err := testEnv.Cluster.GetKubeClient()
			Expect(err).NotTo(HaveOccurred())

			cnpgName := shared.CnpgClusterName(addonName, int(addon.Spec.Version.Major))
			localPort, stopChan := shared.PortForwardPostgres(ctx, testEnv.Cluster.GetRESTConfig(), clientset, namespace, cnpgName)
			defer close(stopChan)

			By("Connecting to postgres and running a query")
			db := shared.ConnectToPostgres("127.0.0.1", localPort, creds.GetUsername(), creds.GetPassword(), "testdb", "disable")
			defer db.Close()

			var result int
			err = db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
			Expect(err).NotTo(HaveOccurred(), "SELECT 1 query should succeed")
			Expect(result).To(Equal(1))
		})
	})

	Context("Update Propagation", func() {
		It("should propagate spec updates to the cluster CR", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon")
			addon := shared.CreateMinimalPostgresAddon("test-update-prop")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			namespace := createdAddon.GetNamespace()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for CR to appear")
			shared.WaitForCRExists(ctx, clusterClient, shared.CRNameForAddon(addonName), namespace, 2*time.Minute)

			By("Waiting for addon to become Ready")
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Updating addon to 3 instances")
			updateAddon := shared.CreateMinimalPostgresAddon("test-update-prop")
			updateAddon.Spec.Instances.SetCount(3)
			shared.UpdatePostgresAddon(client, orgID, teamName, addonID, updateAddon)

			By("Waiting for CR to reflect the updated instance count")
			Eventually(func(g Gomega) {
				cr, err := shared.GetPostgresClusterCR(ctx, clusterClient, shared.CRNameForAddon(addonName), namespace)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(cr.Spec.Instances).To(Equal(3))
			}, 2*time.Minute, 5*time.Second).Should(Succeed())

			By("Waiting for addon to return to Ready state")
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)
		})
	})

	Context("Deletion Cleanup", func() {
		It("should delete the PostgresCluster CR when addon is deleted via API", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a postgres addon")
			addon := shared.CreateMinimalPostgresAddon("test-delete-cleanup")
			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()
			addonName := createdAddon.GetName()
			namespace := createdAddon.GetNamespace()

			By("Waiting for addon to become Ready")
			shared.WaitForCRExists(ctx, clusterClient, shared.CRNameForAddon(addonName), namespace, 2*time.Minute)
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Deleting the addon via API")
			shared.DeletePostgresAddon(client, orgID, teamName, addonID)

			By("Verifying the CR is deleted from the cluster")
			shared.WaitForCRDeleted(ctx, clusterClient, shared.CRNameForAddon(addonName), namespace, 2*time.Minute)

			By("Verifying the addon is gone from the API")
			shared.WaitForAddonDeleted(client, orgID, teamName, addonID, 30*time.Second)
		})
	})

	Context("Failure Reporting", func() {
		It("should report non-Ready status for invalid configuration", func() {
			By("Creating a postgres addon with a non-existent storage class")
			addon := shared.CreateMinimalPostgresAddon("test-failure-report")
			addon.Spec.Storage.SetStorageClass("nonexistent-storage-class")

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for addon status to reflect the failure")
			Eventually(func(g Gomega) {
				ctx := context.Background()
				resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdGet(ctx, orgID, teamName, addonID).Execute()
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(httpResp.StatusCode).To(Equal(200))

				status, ok := resp.GetStatusOk()
				g.Expect(ok).To(BeTrue())

				state, stateOk := status.GetStateOk()
				g.Expect(stateOk).To(BeTrue())
				g.Expect(*state).NotTo(Equal(string(models.PostgresAddonStateReady)), "addon should not be Ready with invalid storage class")
			}, 3*time.Minute, 10*time.Second).Should(Succeed())
		})
	})

	Context("Backup and WAL Archiving", func() {
		It("should trigger a backup and record it via the backup controller", func() {
			By("Creating an S3 credentials secret")
			secret := shared.CreateS3CredentialsSecret("minio-creds")
			createdSecret := shared.CreateSecret(client, orgID, teamName, secret)

			By("Creating an ObjectStore pointing to MinIO")
			store := shared.CreateObjectStoreWithS3Endpoint(
				"minio-store",
				createdSecret.GetId(),
				shared.MinIOEndpoint(),
			)
			createdStore := shared.CreateObjectStore(client, orgID, teamName, store)

			By("Creating a postgres addon with backup config referencing the ObjectStore")
			addon := shared.CreateMinimalPostgresAddon("test-backup-e2e")

			db := openapi.NewPostgresDatabase("appdb")
			db.SetExtensions([]string{})
			addon.Spec.SetDatabases([]openapi.PostgresDatabase{*db})

			backup := openapi.NewPostgresBackupConfig()
			backup.SetEnabled(true)
			backup.SetWalArchiving(true)
			backup.SetObjectStoreId(createdStore.GetId())
			addon.Spec.SetBackup(*backup)

			createdAddon := shared.CreatePostgresAddon(client, orgID, teamName, addon)
			addonID := createdAddon.GetId()

			DeferCleanup(func() {
				shared.DeletePostgresAddon(client, orgID, teamName, addonID)
			})

			By("Waiting for addon to become Ready")
			shared.WaitForAddonReady(client, orgID, teamName, addonID, 10*time.Minute)

			By("Verifying ContinuousWalArchivingSuccess condition becomes True")
			shared.WaitForConditionTrue(client, orgID, teamName, addonID, string(models.PostgresAddonConditionWalArchivingSuccess), 5*time.Minute)

			By("Triggering an immediate backup")
			shared.TriggerBackup(client, orgID, teamName, addonID)

			By("Waiting for backup to complete")
			shared.WaitForBackupPhase(client, orgID, teamName, addonID, "completed", 10*time.Minute)

			By("Verifying backup records exist")
			backups := shared.ListBackups(client, orgID, teamName, addonID)
			Expect(len(backups)).To(BeNumerically(">=", 1), "should have at least one backup record")
		})
	})
})
