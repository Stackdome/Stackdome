package shared

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

// PostgreSQL addon factory functions using OpenAPI models
func CreateMinimalPostgresAddon(name string) *openapi.PostgresAddon {
	version := openapi.NewPostgresVersion(16)

	instances := openapi.NewPostgresInstances(1)

	storage := openapi.NewPostgresStorage("5Gi", "standard")

	spec := openapi.NewPostgresAddonSpec(*version, *instances, *storage)

	addon := openapi.NewPostgresAddon(name, *spec)

	return addon
}

func CreatePostgresAddonWithResources(name string) *openapi.PostgresAddon {
	addon := CreateMinimalPostgresAddon(name)

	// Add resources
	cpuRes := openapi.NewPostgresResourcesCpu()
	cpuRes.SetRequest("250m")
	cpuRes.SetLimit("1")

	memRes := openapi.NewPostgresResourcesMemory()
	memRes.SetRequest("512Mi")
	memRes.SetLimit("2Gi")

	resources := openapi.NewPostgresResources()
	resources.SetCpu(*cpuRes)
	resources.SetMemory(*memRes)

	addon.Spec.SetResources(*resources)

	// Add database
	db := openapi.NewPostgresDatabase("testdb")
	db.SetExtensions([]string{})

	addon.Spec.SetDatabases([]openapi.PostgresDatabase{*db})

	return addon
}

func CreatePostgresAddonForUpdate(name string) *openapi.PostgresAddon {
	addon := CreateMinimalPostgresAddon(name)

	// Scale to 3 instances
	instances := openapi.NewPostgresInstances(3)
	addon.Spec.SetInstances(*instances)

	// Add resources
	cpuRes := openapi.NewPostgresResourcesCpu()
	cpuRes.SetRequest("500m")
	cpuRes.SetLimit("2")

	memRes := openapi.NewPostgresResourcesMemory()
	memRes.SetRequest("1Gi")
	memRes.SetLimit("4Gi")

	resources := openapi.NewPostgresResources()
	resources.SetCpu(*cpuRes)
	resources.SetMemory(*memRes)

	addon.Spec.SetResources(*resources)

	// Add multiple databases
	db1 := openapi.NewPostgresDatabase("app")
	db1.SetExtensions([]string{})

	db2 := openapi.NewPostgresDatabase("analytics")
	db2.SetExtensions([]string{"vector"})

	addon.Spec.SetDatabases([]openapi.PostgresDatabase{*db1, *db2})

	// Add configuration
	config := openapi.NewPostgresConfiguration()
	config.SetEnableSuperuserAccess(true)
	params := map[string]string{
		"max_connections": "200",
		"shared_buffers":  "256MB",
	}
	config.SetParameters(params)

	addon.Spec.SetConfiguration(*config)

	return addon
}

func CreatePostgresAddonWithBackup(name string) *openapi.PostgresAddon {
	addon := CreateMinimalPostgresAddon(name)

	// Add backup configuration
	backup := openapi.NewPostgresBackupConfig()
	backup.SetEnabled(true)
	backup.SetSchedule("0 0 0 * * *")
	backup.SetWalArchiving(false)

	addon.Spec.SetBackup(*backup)

	return addon
}

func CreatePostgresAddonWithHA(name string) *openapi.PostgresAddon {
	addon := CreateMinimalPostgresAddon(name)

	// Scale to 3 instances for HA
	instances := openapi.NewPostgresInstances(3)
	addon.Spec.SetInstances(*instances)

	// Add placement configuration for multi-AZ
	placement := openapi.NewPostgresInstancesPlacement()
	placement.SetTopologyKey("kubernetes.io/zone")

	addon.Spec.Instances.SetPlacement(*placement)

	return addon
}

// Secret factory functions

// CreateGenericSecret creates a generic secret with key-value pairs
func CreateGenericSecret(name string, data map[string]string) *openapi.Secret {
	secretData := make([]openapi.SecretData, 0, len(data))
	for k, v := range data {
		secretData = append(secretData, *openapi.NewSecretData(k, v))
	}
	return openapi.NewSecret(name, openapi.GENERIC, secretData)
}

// CreateS3CredentialsSecret creates a secret with S3 access credentials
func CreateS3CredentialsSecret(name string) *openapi.Secret {
	data := []openapi.SecretData{
		*openapi.NewSecretData("access_key_id", "AKIAIOSFODNN7EXAMPLE"),
		*openapi.NewSecretData("secret_access_key", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"),
	}
	return openapi.NewSecret(name, openapi.GENERIC, data)
}

// CreateAzureCredentialsSecret creates a secret with Azure connection string
func CreateAzureCredentialsSecret(name string) *openapi.Secret {
	data := []openapi.SecretData{
		*openapi.NewSecretData("connection_string", "DefaultEndpointsProtocol=https;AccountName=test;AccountKey=testkey;EndpointSuffix=core.windows.net"),
	}
	return openapi.NewSecret(name, openapi.GENERIC, data)
}

// CreateGCSCredentialsSecret creates a secret with GCS service account credentials
func CreateGCSCredentialsSecret(name string) *openapi.Secret {
	data := []openapi.SecretData{
		*openapi.NewSecretData("service_account_credentials", `{"type":"service_account","project_id":"test-project"}`),
	}
	return openapi.NewSecret(name, openapi.GENERIC, data)
}

// ObjectStore factory functions

// CreateObjectStoreWithS3 creates an ObjectStore with S3 credentials
func CreateObjectStoreWithS3(name string, secretID string) *openapi.ObjectStore {
	accessKeyRef := *openapi.NewSecretReference(secretID, "access_key_id")
	secretKeyRef := *openapi.NewSecretReference(secretID, "secret_access_key")

	s3Creds := openapi.NewS3Credentials(accessKeyRef, secretKeyRef, "us-west-2")

	config := openapi.NewObjectStoreConfiguration()
	config.SetS3Credentials(*s3Creds)

	spec := openapi.NewObjectStoreSpec(*config, "s3://my-bucket/backups")

	return openapi.NewObjectStore(name, *spec)
}

// CreateObjectStoreWithS3Endpoint creates an ObjectStore with S3-compatible endpoint
func CreateObjectStoreWithS3Endpoint(name string, secretID string, endpoint string) *openapi.ObjectStore {
	store := CreateObjectStoreWithS3(name, secretID)
	s3Creds := store.Spec.Configuration.GetS3Credentials()
	s3Creds.SetEndpointUrl(endpoint)
	store.Spec.Configuration.SetS3Credentials(s3Creds)
	return store
}

// CreateObjectStoreWithAzure creates an ObjectStore with Azure credentials
func CreateObjectStoreWithAzure(name string, secretID string) *openapi.ObjectStore {
	connStringRef := *openapi.NewSecretReference(secretID, "connection_string")

	azureCreds := openapi.NewAzureCredentials(connStringRef)
	azureCreds.SetStorageAccountName("teststorageaccount")

	config := openapi.NewObjectStoreConfiguration()
	config.SetAzureCredentials(*azureCreds)

	spec := openapi.NewObjectStoreSpec(*config, "https://teststorageaccount.blob.core.windows.net/backups")

	return openapi.NewObjectStore(name, *spec)
}

// CreateObjectStoreWithGCS creates an ObjectStore with GCS credentials
func CreateObjectStoreWithGCS(name string, secretID string) *openapi.ObjectStore {
	saCredsRef := *openapi.NewSecretReference(secretID, "service_account_credentials")

	gcsCreds := openapi.NewGCSCredentials(saCredsRef)

	config := openapi.NewObjectStoreConfiguration()
	config.SetGcsCredentials(*gcsCreds)

	spec := openapi.NewObjectStoreSpec(*config, "gs://my-bucket/backups")

	return openapi.NewObjectStore(name, *spec)
}

// CreateObjectStoreWithRetention creates an ObjectStore with custom retention policy
func CreateObjectStoreWithRetention(name string, secretID string, retention string) *openapi.ObjectStore {
	store := CreateObjectStoreWithS3(name, secretID)
	store.Spec.SetRetentionPolicy(retention)
	return store
}
