package int

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("ObjectStore", func() {
	var client *openapi.APIClient
	var orgID string

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("CRUD Operations with S3 credentials", func() {
		var s3Secret *openapi.Secret

		BeforeEach(func() {
			By("Creating an S3 credentials secret")
			secret := shared.CreateS3CredentialsSecret("test-s3-creds")
			s3Secret = shared.CreateSecret(client, orgID, secret)
		})

		It("should create an object store with S3 credentials", func() {
			By("Creating object store with S3 configuration")
			store := shared.CreateObjectStoreWithS3("test-s3-store", s3Secret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			Expect(createdStore.GetName()).To(Equal("test-s3-store"))
			Expect(createdStore.Spec.GetDestinationPath()).To(Equal("s3://my-bucket/backups"))
			Expect(createdStore.Spec.Configuration.HasS3Credentials()).To(BeTrue())

			s3Creds := createdStore.Spec.Configuration.GetS3Credentials()
			Expect(s3Creds.GetRegion()).To(Equal("us-west-2"))
			accessKeyRef := s3Creds.GetAccessKeyId()
			Expect(accessKeyRef.SecretId).To(Equal(s3Secret.GetId()))
			Expect(accessKeyRef.Key).To(Equal("access_key_id"))
		})

		It("should create an object store with S3-compatible endpoint", func() {
			By("Creating object store with MinIO endpoint")
			store := shared.CreateObjectStoreWithS3Endpoint("test-minio-store", s3Secret.GetId(), "https://minio.example.com")

			createdStore := shared.CreateObjectStore(client, orgID, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			s3Creds := createdStore.Spec.Configuration.GetS3Credentials()
			Expect(s3Creds.GetEndpointUrl()).To(Equal("https://minio.example.com"))
		})

		It("should create an object store with custom retention policy", func() {
			By("Creating object store with 30-day retention")
			store := shared.CreateObjectStoreWithRetention("test-retention-store", s3Secret.GetId(), "30d")

			createdStore := shared.CreateObjectStore(client, orgID, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			Expect(createdStore.Spec.GetRetentionPolicy()).To(Equal("30d"))
		})

		It("should retrieve an object store by ID", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-get-store", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, store)

			By("Retrieving the object store by ID")
			retrievedStore := shared.GetObjectStore(client, orgID, createdStore.GetId())

			Expect(retrievedStore.GetId()).To(Equal(createdStore.GetId()))
			Expect(retrievedStore.GetName()).To(Equal(createdStore.GetName()))
			Expect(retrievedStore.Spec.GetDestinationPath()).To(Equal(createdStore.Spec.GetDestinationPath()))
		})

		It("should list object stores", func() {
			By("Creating multiple object stores")
			store1 := shared.CreateObjectStoreWithS3("test-list-store-1", s3Secret.GetId())
			store2 := shared.CreateObjectStoreWithS3("test-list-store-2", s3Secret.GetId())

			shared.CreateObjectStore(client, orgID, store1)
			shared.CreateObjectStore(client, orgID, store2)

			By("Listing all object stores")
			storeList := shared.ListObjectStores(client, orgID)

			Expect(len(storeList.GetItems())).To(BeNumerically(">=", 2))

			var foundNames []string
			for _, store := range storeList.GetItems() {
				foundNames = append(foundNames, store.GetName())
			}
			Expect(foundNames).To(ContainElement("test-list-store-1"))
			Expect(foundNames).To(ContainElement("test-list-store-2"))
		})

		It("should update an object store", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-update-store", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, store)

			By("Updating the object store with new retention policy")
			updateStore := shared.CreateObjectStoreWithRetention("test-update-store", s3Secret.GetId(), "90d")

			updatedStore := shared.UpdateObjectStore(client, orgID, createdStore.GetId(), updateStore)

			Expect(updatedStore.GetId()).To(Equal(createdStore.GetId()))
			Expect(updatedStore.Spec.GetRetentionPolicy()).To(Equal("90d"))
		})

		It("should delete an object store", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-delete-store", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, store)

			By("Deleting the object store")
			shared.DeleteObjectStore(client, orgID, createdStore.GetId())

			By("Verifying the object store is deleted")
			shared.GetObjectStoreExpectError(client, orgID, createdStore.GetId(), 404)
		})
	})

	Context("CRUD Operations with Azure credentials", func() {
		var azureSecret *openapi.Secret

		BeforeEach(func() {
			By("Creating an Azure credentials secret")
			secret := shared.CreateAzureCredentialsSecret("test-azure-creds")
			azureSecret = shared.CreateSecret(client, orgID, secret)
		})

		It("should create an object store with Azure credentials", func() {
			By("Creating object store with Azure configuration")
			store := shared.CreateObjectStoreWithAzure("test-azure-store", azureSecret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			Expect(createdStore.GetName()).To(Equal("test-azure-store"))
			Expect(createdStore.Spec.Configuration.HasAzureCredentials()).To(BeTrue())

			azureCreds := createdStore.Spec.Configuration.GetAzureCredentials()
			connStringRef := azureCreds.GetConnectionString()
			Expect(connStringRef.SecretId).To(Equal(azureSecret.GetId()))
			Expect(connStringRef.Key).To(Equal("connection_string"))
			Expect(azureCreds.GetStorageAccountName()).To(Equal("teststorageaccount"))
		})
	})

	Context("CRUD Operations with GCS credentials", func() {
		var gcsSecret *openapi.Secret

		BeforeEach(func() {
			By("Creating a GCS credentials secret")
			secret := shared.CreateGCSCredentialsSecret("test-gcs-creds")
			gcsSecret = shared.CreateSecret(client, orgID, secret)
		})

		It("should create an object store with GCS credentials", func() {
			By("Creating object store with GCS configuration")
			store := shared.CreateObjectStoreWithGCS("test-gcs-store", gcsSecret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			Expect(createdStore.GetName()).To(Equal("test-gcs-store"))
			Expect(createdStore.Spec.Configuration.HasGcsCredentials()).To(BeTrue())

			gcsCreds := createdStore.Spec.Configuration.GetGcsCredentials()
			saCredsRef := gcsCreds.GetServiceAccountCredentials()
			Expect(saCredsRef.SecretId).To(Equal(gcsSecret.GetId()))
			Expect(saCredsRef.Key).To(Equal("service_account_credentials"))
		})
	})

	Context("Validation", func() {
		var s3Secret *openapi.Secret

		BeforeEach(func() {
			By("Creating an S3 credentials secret")
			secret := shared.CreateS3CredentialsSecret("test-validation-creds")
			s3Secret = shared.CreateSecret(client, orgID, secret)
		})

		It("should reject object store with invalid name", func() {
			By("Creating object store with invalid name (uppercase)")
			store := shared.CreateObjectStoreWithS3("Test-Invalid-Name", s3Secret.GetId())

			shared.CreateObjectStoreExpectError(client, orgID, store, 400)
		})

		It("should reject object store with invalid retention policy format", func() {
			By("Creating object store with invalid retention policy")
			store := shared.CreateObjectStoreWithRetention("test-invalid-retention", s3Secret.GetId(), "invalid")

			shared.CreateObjectStoreExpectError(client, orgID, store, 400)
		})

		It("should reject object store with non-existent secret reference", func() {
			By("Creating object store with non-existent secret ID")
			store := shared.CreateObjectStoreWithS3("test-bad-secret", "non-existent-secret-id")

			shared.CreateObjectStoreExpectError(client, orgID, store, 400)
		})

		It("should reject object store with invalid secret key reference", func() {
			By("Creating object store with wrong secret key")
			accessKeyRef := *openapi.NewSecretReference(s3Secret.GetId(), "wrong_key")
			secretKeyRef := *openapi.NewSecretReference(s3Secret.GetId(), "secret_access_key")

			s3Creds := openapi.NewS3Credentials(accessKeyRef, secretKeyRef, "us-west-2")

			config := openapi.NewObjectStoreConfiguration()
			config.SetS3Credentials(*s3Creds)

			spec := openapi.NewObjectStoreSpec(*config, "s3://my-bucket/backups")
			store := openapi.NewObjectStore("test-wrong-key", *spec)

			shared.CreateObjectStoreExpectError(client, orgID, store, 400)
		})

		It("should reject object store with no credentials", func() {
			By("Creating object store with empty configuration")
			config := openapi.NewObjectStoreConfiguration()
			spec := openapi.NewObjectStoreSpec(*config, "s3://my-bucket/backups")
			store := openapi.NewObjectStore("test-no-creds", *spec)

			shared.CreateObjectStoreExpectError(client, orgID, store, 400)
		})

		It("should reject object store name change on update", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-immutable-name", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, store)

			By("Attempting to change the name")
			updateStore := shared.CreateObjectStoreWithS3("test-new-name", s3Secret.GetId())

			shared.UpdateObjectStoreExpectError(client, orgID, createdStore.GetId(), updateStore, 400)
		})
	})
})
