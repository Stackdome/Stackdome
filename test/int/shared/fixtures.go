package shared

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

// Shared test image used across all stack fixtures
const TestImage = "nginx:1.25-alpine"

// Cluster registration fixture values
const TestRegistryName = "test-registry"

// InitContainer fixture values
const (
	InitImage   = "busybox:1.36"
	InitCommand = "echo init-done"
)

// Multi-resource stack fixture values
const (
	MultiResourceBackendName  = "backend"
	MultiResourceFrontendName = "frontend"
	MultiResourceBackendPort  = 8080
	MultiResourceFrontendPort = 80
)

// Env and ports stack fixture values
const (
	EnvPortsResourceName = "app"
	EnvPortsAppEnvKey    = "APP_ENV"
	EnvPortsAppEnvVal    = "test"
	EnvPortsAppPortKey   = "APP_PORT"
	EnvPortsAppPortVal   = "8080"
	EnvPortsLogLevelKey  = "LOG_LEVEL"
	EnvPortsLogLevelVal  = "debug"
	EnvPortsPort1        = 8080
	EnvPortsPort2        = 9090
)

// Postgres addon env mapping keys used in CreateStackWithPostgresAddon
var PostgresEnvMapping = map[string]string{
	"host":             "PG_HOST",
	"port":             "PG_PORT",
	"username":         "PG_USER",
	"password":         "PG_PASSWORD",
	"database":         "PG_DATABASE",
	"connectionString": "DATABASE_URL",
}

// Build from source fixture values
const (
	BuildSourceRepoURL      = "https://github.com/ashishmax31/test-private-repo.git"
	BuildSourceBranch       = "main"
	BuildSourceDockerfile   = "Dockerfile"
	BuildSourceContextPath  = "."
	BuildSourcePort         = 3000
	BuildSourceResourceName = "todo-app"
	BuildSourceSecretName   = "test-git-creds"
)

// PostgreSQL addon factory functions using OpenAPI models
func CreateMinimalPostgresAddon(name string) *openapi.PostgresAddon {
	version := openapi.NewPostgresVersion(16)
	version.SetMinor(6)

	instances := openapi.NewPostgresInstances(2)

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

// CreateS3CredentialsSecret creates a secret with MinIO S3 access credentials
func CreateS3CredentialsSecret(name string) *openapi.Secret {
	data := []openapi.SecretData{
		*openapi.NewSecretData("access_key_id", MinIOAccessKey),
		*openapi.NewSecretData("secret_access_key", MinIOSecretKey),
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

// CreateGitCredentialsSecret creates a GitCredentials secret with a GitHub PAT token
func CreateGitCredentialsSecret(name string, token string) *openapi.Secret {
	data := []openapi.SecretData{
		*openapi.NewSecretData("token", token),
	}
	return openapi.NewSecret(name, openapi.GIT_CREDENTIALS, data)
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
	store.Spec.SetDestinationPath("s3://backups/")
	return store
}

// CreateObjectStoreWithAzure creates an ObjectStore with Azure credentials
func CreateObjectStoreWithAzure(name string, secretID string) *openapi.ObjectStore {
	connStringRef := *openapi.NewSecretReference(secretID, "connection_string")

	azureCreds := openapi.NewAzureCredentials(connStringRef)
	azureCreds.SetStorageAccountName("teststorageaccount")

	config := openapi.NewObjectStoreConfiguration()
	config.SetAzureCredentials(*azureCreds)

	spec := openapi.NewObjectStoreSpec(*config, "https://teststorageaccount.blob.core.windows.net/backups/data")

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

// Stack factory functions using OpenAPI models

func CreateSimpleStack(name string) *openapi.Stack {
	resource := openapi.NewStackResource("web")
	imageSpec := openapi.NewImageSpec("nginx:1.25-alpine")
	resource.SetImageSpec(*imageSpec)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort(80, false),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateMultiResourceStack(name string) *openapi.Stack {
	backend := openapi.NewStackResource(MultiResourceBackendName)
	backendImage := openapi.NewImageSpec(TestImage)
	backend.SetImageSpec(*backendImage)
	backend.SetPorts([]openapi.Port{
		*openapi.NewPort(MultiResourceBackendPort, false),
	})
	backendExec := openapi.NewExecutionConfig()
	backendExec.SetEnvironmentVariables([]openapi.EnvVar{
		*openapi.NewEnvVar("APP_ROLE", MultiResourceBackendName),
	})
	backend.SetExecutionConfig(*backendExec)

	frontend := openapi.NewStackResource(MultiResourceFrontendName)
	frontendImage := openapi.NewImageSpec(TestImage)
	frontend.SetImageSpec(*frontendImage)
	frontend.SetPorts([]openapi.Port{
		*openapi.NewPort(MultiResourceFrontendPort, false),
	})
	frontendExec := openapi.NewExecutionConfig()
	frontendExec.SetEnvironmentVariables([]openapi.EnvVar{
		*openapi.NewEnvVar("BACKEND_URL", "{{ STACKDOME_BACKEND_INTERNAL }}"),
	})
	frontend.SetExecutionConfig(*frontendExec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*backend, *frontend})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithDependencies(name string) *openapi.Stack {
	resourceA := openapi.NewStackResource("database")
	imageA := openapi.NewImageSpec("nginx:1.25-alpine")
	resourceA.SetImageSpec(*imageA)
	resourceA.SetPorts([]openapi.Port{
		*openapi.NewPort(5432, false),
	})

	resourceB := openapi.NewStackResource("app")
	imageB := openapi.NewImageSpec("nginx:1.25-alpine")
	resourceB.SetImageSpec(*imageB)
	resourceB.SetPorts([]openapi.Port{
		*openapi.NewPort(8080, false),
	})
	resourceB.SetDependsOn([]string{"database"})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resourceA, *resourceB})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithEnvAndPorts(name string) *openapi.Stack {
	resource := openapi.NewStackResource(EnvPortsResourceName)
	image := openapi.NewImageSpec(TestImage)
	resource.SetImageSpec(*image)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort(EnvPortsPort1, false),
		*openapi.NewPort(EnvPortsPort2, false),
	})
	exec := openapi.NewExecutionConfig()
	exec.SetEnvironmentVariables([]openapi.EnvVar{
		*openapi.NewEnvVar(EnvPortsAppEnvKey, EnvPortsAppEnvVal),
		*openapi.NewEnvVar(EnvPortsAppPortKey, EnvPortsAppPortVal),
		*openapi.NewEnvVar(EnvPortsLogLevelKey, EnvPortsLogLevelVal),
	})
	resource.SetExecutionConfig(*exec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithInitContainer(name string) *openapi.Stack {
	resource := openapi.NewStackResource("app")
	image := openapi.NewImageSpec(TestImage)
	resource.SetImageSpec(*image)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort(80, false),
	})

	initSpec := openapi.NewInitSpec()
	initImage := openapi.NewImageSpec(InitImage)
	initSpec.SetImageSpec(*initImage)
	initSpec.Command = []string{"sh", "-c", InitCommand}
	resource.SetInitSpec(*initSpec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithPostgresAddon(name string, addonID string, database string) *openapi.Stack {
	resource := openapi.NewStackResource("app")
	image := openapi.NewImageSpec("nginx:1.25-alpine")
	resource.SetImageSpec(*image)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort(8080, false),
	})

	pgEnvSource := openapi.NewPostgresAddonEnvSource(addonID, database, PostgresEnvMapping)
	addonEnvSource := openapi.NewAddonEnvSource()
	addonEnvSource.SetPostgres(*pgEnvSource)

	exec := openapi.NewExecutionConfig()
	exec.SetEnvFromAddons([]openapi.AddonEnvSource{*addonEnvSource})
	resource.SetExecutionConfig(*exec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithBuildSource(name string, repoURL string, secretID string) *openapi.Stack {
	resource := openapi.NewStackResource(BuildSourceResourceName)

	// Build source context — git repo with secret for private access
	gitRepo := openapi.NewBuildSourceContextGitRepo(repoURL)
	gitSecret := openapi.NewSecretRef(secretID)
	gitRepo.SetGitSecret(*gitSecret)

	sourceContext := openapi.NewBuildSourceContext()
	sourceContext.SetGitRepo(*gitRepo)

	// Source revision — branch
	branchRevision := openapi.GitRepoRevisionBranch{
		Name: openapi.PtrString(BuildSourceBranch),
	}
	gitRepoRevision := openapi.NewGitRepoRevision()
	gitRepoRevision.SetBranch(branchRevision)

	sourceRevision := openapi.NewBuildSourceRevision()
	sourceRevision.SetGitRepoRevision(*gitRepoRevision)

	// Image repository — use internal (in-cluster Zot) registry
	imageRepo := openapi.NewImageRepository()
	imageRepo.SetUseInternalRegistry(true)

	// Assemble build spec
	buildSpec := openapi.NewStackResourceBuildSpec(
		*sourceContext,
		BuildSourceContextPath,
		BuildSourceDockerfile,
		*sourceRevision,
		*imageRepo,
	)
	resource.SetBuildSpec(*buildSpec)

	// Port 3000 exposed to public
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort(int32(BuildSourcePort), true),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}
