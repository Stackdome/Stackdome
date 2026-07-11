package int

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

var _ = Describe("ObjectStore", func() {
	var client *openapi.APIClient
	var orgID string
	projectName := models.DefaultProjectName

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
			s3Secret = shared.CreateSecret(client, orgID, projectName, secret)
		})

		It("should create an object store with S3 credentials", func() {
			By("Creating object store with S3 configuration")
			store := shared.CreateObjectStoreWithS3("test-s3-store", s3Secret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

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

			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			s3Creds := createdStore.Spec.Configuration.GetS3Credentials()
			Expect(s3Creds.GetEndpointUrl()).To(Equal("https://minio.example.com"))
		})

		It("should create an object store with custom retention policy", func() {
			By("Creating object store with 30-day retention")
			store := shared.CreateObjectStoreWithRetention("test-retention-store", s3Secret.GetId(), "30d")

			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			Expect(createdStore.GetId()).NotTo(BeEmpty())
			Expect(createdStore.Spec.GetRetentionPolicy()).To(Equal("30d"))
		})

		It("should retrieve an object store by ID", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-get-store", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			By("Retrieving the object store by ID")
			retrievedStore := shared.GetObjectStore(client, orgID, projectName, createdStore.GetId())

			Expect(retrievedStore.GetId()).To(Equal(createdStore.GetId()))
			Expect(retrievedStore.GetName()).To(Equal(createdStore.GetName()))
			Expect(retrievedStore.Spec.GetDestinationPath()).To(Equal(createdStore.Spec.GetDestinationPath()))
		})

		It("should list object stores", func() {
			By("Creating multiple object stores")
			store1 := shared.CreateObjectStoreWithS3("test-list-store-1", s3Secret.GetId())
			store2 := shared.CreateObjectStoreWithS3("test-list-store-2", s3Secret.GetId())

			shared.CreateObjectStore(client, orgID, projectName, store1)
			shared.CreateObjectStore(client, orgID, projectName, store2)

			By("Listing all object stores")
			storeList := shared.ListObjectStores(client, orgID, projectName)

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
			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			By("Updating the object store with new retention policy")
			updateStore := shared.CreateObjectStoreWithRetention("test-update-store", s3Secret.GetId(), "90d")

			updatedStore := shared.UpdateObjectStore(client, orgID, projectName, createdStore.GetId(), updateStore)

			Expect(updatedStore.GetId()).To(Equal(createdStore.GetId()))
			Expect(updatedStore.Spec.GetRetentionPolicy()).To(Equal("90d"))
		})

		It("should delete an object store", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-delete-store", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			By("Deleting the object store")
			shared.DeleteObjectStore(client, orgID, projectName, createdStore.GetId())

			By("Verifying the object store is deleted")
			_ = shared.GetObjectStoreExpectError(client, orgID, projectName, createdStore.GetId(), 404)
		})
	})

	Context("CRUD Operations with Azure credentials", func() {
		var azureSecret *openapi.Secret

		BeforeEach(func() {
			By("Creating an Azure credentials secret")
			secret := shared.CreateAzureCredentialsSecret("test-azure-creds")
			azureSecret = shared.CreateSecret(client, orgID, projectName, secret)
		})

		It("should create an object store with Azure credentials", func() {
			By("Creating object store with Azure configuration")
			store := shared.CreateObjectStoreWithAzure("test-azure-store", azureSecret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

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
			gcsSecret = shared.CreateSecret(client, orgID, projectName, secret)
		})

		It("should create an object store with GCS credentials", func() {
			By("Creating object store with GCS configuration")
			store := shared.CreateObjectStoreWithGCS("test-gcs-store", gcsSecret.GetId())

			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

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
			s3Secret = shared.CreateSecret(client, orgID, projectName, secret)
		})

		It("should reject object store with invalid name", func() {
			By("Creating object store with invalid name (uppercase)")
			store := shared.CreateObjectStoreWithS3("Test-Invalid-Name", s3Secret.GetId())

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject object store with invalid retention policy format", func() {
			By("Creating object store with invalid retention policy")
			store := shared.CreateObjectStoreWithRetention("test-invalid-retention", s3Secret.GetId(), "invalid")

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject object store with non-existent secret reference", func() {
			By("Creating object store with non-existent secret ID")
			store := shared.CreateObjectStoreWithS3("test-bad-secret", "non-existent-secret-id")

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
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

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject object store with no credentials", func() {
			By("Creating object store with empty configuration")
			config := openapi.NewObjectStoreConfiguration()
			spec := openapi.NewObjectStoreSpec(*config, "s3://my-bucket/backups")
			store := openapi.NewObjectStore("test-no-creds", *spec)

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject object store name change on update", func() {
			By("Creating an object store first")
			store := shared.CreateObjectStoreWithS3("test-immutable-name", s3Secret.GetId())
			createdStore := shared.CreateObjectStore(client, orgID, projectName, store)

			By("Attempting to change the name")
			updateStore := shared.CreateObjectStoreWithS3("test-new-name", s3Secret.GetId())

			_ = shared.UpdateObjectStoreExpectError(client, orgID, projectName, createdStore.GetId(), updateStore, 400)
		})

		It("should reject S3 object store with invalid destination path prefix", func() {
			By("Creating S3 object store with non-s3:// prefix")
			store := shared.CreateObjectStoreWithS3("test-s3-bad-prefix", s3Secret.GetId())
			store.Spec.SetDestinationPath("http://my-bucket/backups")

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject S3 object store with missing bucket name", func() {
			By("Creating S3 object store with empty bucket")
			store := shared.CreateObjectStoreWithS3("test-s3-no-bucket", s3Secret.GetId())
			store.Spec.SetDestinationPath("s3://")

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject object store with empty destination path", func() {
			By("Creating S3 object store with empty destination path")
			store := shared.CreateObjectStoreWithS3("test-empty-path", s3Secret.GetId())
			store.Spec.SetDestinationPath("")

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject Azure object store with invalid destination path", func() {
			By("Creating Azure credentials secret")
			azureSecretData := shared.CreateAzureCredentialsSecret("test-azure-validation-creds")
			azureSecret := shared.CreateSecret(client, orgID, projectName, azureSecretData)

			By("Creating Azure object store with invalid destination path")
			connStringRef := *openapi.NewSecretReference(azureSecret.GetId(), "connection_string")
			azureCreds := openapi.NewAzureCredentials(connStringRef)
			azureCreds.SetStorageAccountName("teststorageaccount")

			config := openapi.NewObjectStoreConfiguration()
			config.SetAzureCredentials(*azureCreds)

			spec := openapi.NewObjectStoreSpec(*config, "https://invalid-path/container")
			store := openapi.NewObjectStore("test-azure-bad-path", *spec)

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject GCS object store with invalid destination path prefix", func() {
			By("Creating GCS credentials secret")
			gcsSecretData := shared.CreateGCSCredentialsSecret("test-gcs-validation-creds")
			gcsSecret := shared.CreateSecret(client, orgID, projectName, gcsSecretData)

			By("Creating GCS object store with non-gs:// prefix")
			saCredsRef := *openapi.NewSecretReference(gcsSecret.GetId(), "service_account_credentials")
			gcsCreds := openapi.NewGCSCredentials(saCredsRef)

			config := openapi.NewObjectStoreConfiguration()
			config.SetGcsCredentials(*gcsCreds)

			spec := openapi.NewObjectStoreSpec(*config, "s3://wrong-prefix/backups")
			store := openapi.NewObjectStore("test-gcs-bad-prefix", *spec)

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})

		It("should reject GCS object store with missing bucket name", func() {
			By("Creating GCS credentials secret")
			gcsSecretData := shared.CreateGCSCredentialsSecret("test-gcs-no-bucket-creds")
			gcsSecret := shared.CreateSecret(client, orgID, projectName, gcsSecretData)

			By("Creating GCS object store with empty bucket")
			saCredsRef := *openapi.NewSecretReference(gcsSecret.GetId(), "service_account_credentials")
			gcsCreds := openapi.NewGCSCredentials(saCredsRef)

			config := openapi.NewObjectStoreConfiguration()
			config.SetGcsCredentials(*gcsCreds)

			spec := openapi.NewObjectStoreSpec(*config, "gs://")
			store := openapi.NewObjectStore("test-gcs-no-bucket", *spec)

			_ = shared.CreateObjectStoreExpectError(client, orgID, projectName, store, 400)
		})
	})
})
