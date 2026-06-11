package shared

import (
	"fmt"

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

// PostgresEnvMapping maps output accessor names to env var names for postgres connections.
var PostgresEnvMapping = map[string]string{
	"host":     "PG_HOST",
	"port":     "PG_PORT",
	"username": "PG_USER",
	"password": "PG_PASSWORD",
	"database": "PG_DATABASE",
	"url":      "DATABASE_URL",
}

// Full-stack e2e fixture values (2 resources, postgres, secret, volume, depends_on)
const (
	FullStackAPIName    = "api"
	FullStackWorkerName = "worker"
	FullStackAPIPort    = 8080
	FullStackVolumeName = "uploads"
	FullStackMountPath  = "/data/uploads"
	FullStackSubPath    = "api"
	FullStackSecretName = "tls-certs"
)

// FullStackSecretEnvMapping maps secret output keys to env var names.
var FullStackSecretEnvMapping = map[string]string{
	"tls_cert": "TLS_CERT",
	"tls_key":  "TLS_KEY",
}

// Crash detection fixture values
const (
	CrashResourceName = "crash-app"
	CrashImage        = "busybox:1.36"
	CrashMessage      = "application crashed with fatal error"
)

// Build from source fixture values
const (
	BuildSourceRepoURL      = "https://github.com/ashishmax31/test-private-repo.git"
	BuildSourceBranch       = "main"
	BuildSourceDockerfile   = "Dockerfile"
	BuildSourceContextPath  = "."
	BuildSourcePort         = 3000
	BuildSourceResourceName = "todo-app"
	BuildSourceSecretName   = "test-git-creds"

	// BrokenBuildResourceName is the resource name for the broken-build fixture.
	// It points to a branch that contains a Dockerfile with an invalid command so the build fails.
	BrokenBuildResourceName     = "broken-app"
	BrokenBuildSourceBranch     = "broken-dockerfile"
	BrokenBuildSourceDockerfile = "Dockerfile"
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

func CreatePostgresAddonWithSuperuser(name string) *openapi.PostgresAddon {
	addon := CreatePostgresAddonWithResources(name)

	config := openapi.NewPostgresConfiguration()
	config.SetEnableSuperuserAccess(true)
	addon.Spec.SetConfiguration(*config)

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

// CreateCrashingStack creates a stack whose sole resource immediately exits with a non-zero
// code, triggering CrashLoopBackOff so the cluster-agent populates LastFailureDetails.
func CreateCrashingStack(name string) *openapi.Stack {
	resource := openapi.NewStackResource(CrashResourceName)
	imageSpec := openapi.NewImageSpec(CrashImage)
	resource.SetImageSpec(*imageSpec)
	exec := openapi.NewExecutionConfig()
	exec.SetCommand([]string{"sh", "-c", fmt.Sprintf("echo '%s'; exit 1", CrashMessage)})
	resource.SetExecutionConfig(*exec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateSimpleStack(name string) *openapi.Stack {
	resource := openapi.NewStackResource("web")
	imageSpec := openapi.NewImageSpec("nginx:1.25-alpine")
	resource.SetImageSpec(*imageSpec)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort("http", 80, false),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func CreateMultiResourceStack(name string) *openapi.Stack {
	backend := openapi.NewStackResource(MultiResourceBackendName)
	backendImage := openapi.NewImageSpec(TestImage)
	backend.SetImageSpec(*backendImage)
	backend.SetPorts([]openapi.Port{
		*openapi.NewPort("http", MultiResourceBackendPort, false),
	})
	backendExec := openapi.NewExecutionConfig()
	backendExec.SetEnvironmentVariables([]openapi.EnvVar{
		func() openapi.EnvVar {
			v := openapi.NewEnvVar("APP_ROLE")
			v.SetValue(MultiResourceBackendName)
			return *v
		}(),
	})
	backend.SetExecutionConfig(*backendExec)

	frontend := openapi.NewStackResource(MultiResourceFrontendName)
	frontendImage := openapi.NewImageSpec(TestImage)
	frontend.SetImageSpec(*frontendImage)
	frontend.SetPorts([]openapi.Port{
		*openapi.NewPort("http", MultiResourceFrontendPort, false),
	})

	from := openapi.NewTopologyNodeRef("stack_resource")
	from.SetName(MultiResourceBackendName)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(MultiResourceFrontendName)
	conn := openapi.NewStackConnection("env", *from, *to)
	target := openapi.NewConnectionTarget("env")
	target.SetName("BACKEND_URL")
	value := openapi.NewValueRef()
	value.SetOutput("host")
	conn.SetMappings([]openapi.ConnectionMapping{*openapi.NewConnectionMapping(*target, *value)})

	spec := openapi.NewStackSpec([]openapi.StackResource{*backend, *frontend})
	spec.SetConnections([]openapi.StackConnection{*conn})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithDependencies(name string) *openapi.Stack {
	resourceA := openapi.NewStackResource("database")
	imageA := openapi.NewImageSpec("nginx:1.25-alpine")
	resourceA.SetImageSpec(*imageA)
	resourceA.SetPorts([]openapi.Port{
		*openapi.NewPort("postgres", 5432, false),
	})

	resourceB := openapi.NewStackResource("app")
	imageB := openapi.NewImageSpec("nginx:1.25-alpine")
	resourceB.SetImageSpec(*imageB)
	resourceB.SetPorts([]openapi.Port{
		*openapi.NewPort("http", 8080, false),
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
		*openapi.NewPort("http", EnvPortsPort1, false),
		*openapi.NewPort("metrics", EnvPortsPort2, false),
	})
	exec := openapi.NewExecutionConfig()
	exec.SetEnvironmentVariables([]openapi.EnvVar{
		func() openapi.EnvVar {
			v := openapi.NewEnvVar(EnvPortsAppEnvKey)
			v.SetValue(EnvPortsAppEnvVal)
			return *v
		}(),
		func() openapi.EnvVar {
			v := openapi.NewEnvVar(EnvPortsAppPortKey)
			v.SetValue(EnvPortsAppPortVal)
			return *v
		}(),
		func() openapi.EnvVar {
			v := openapi.NewEnvVar(EnvPortsLogLevelKey)
			v.SetValue(EnvPortsLogLevelVal)
			return *v
		}(),
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
		*openapi.NewPort("http", 80, false),
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
		*openapi.NewPort("http", 8080, false),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	spec.SetConnections([]openapi.StackConnection{
		postgresEnvConnection(addonID, "app", database, false),
	})
	return openapi.NewStack(name, *spec)
}

func CreateStackWithPostgresAddonSuperuser(name string, addonID string) *openapi.Stack {
	resource := openapi.NewStackResource("app")
	image := openapi.NewImageSpec("nginx:1.25-alpine")
	resource.SetImageSpec(*image)
	resource.SetPorts([]openapi.Port{
		*openapi.NewPort("http", 8080, false),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	spec.SetConnections([]openapi.StackConnection{
		postgresEnvConnection(addonID, "app", "", true),
	})
	return openapi.NewStack(name, *spec)
}

func postgresEnvConnection(addonID, targetResource, database string, superuser bool) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef("addon/postgres")
	from.SetId(addonID)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(targetResource)
	conn := openapi.NewStackConnection("env", *from, *to)

	pgConfig := openapi.NewPostgresEnvConfig()
	hasConfig := false
	if database != "" {
		pgConfig.Database = &database
		hasConfig = true
	}
	if superuser {
		scope := "superuser"
		pgConfig.CredentialScope = &scope
		hasConfig = true
	}
	if hasConfig {
		conn.SetConfig(openapi.StackConnectionConfig{PostgresEnvConfig: pgConfig})
	}

	var mappings []openapi.ConnectionMapping
	for output, envName := range PostgresEnvMapping {
		target := openapi.NewConnectionTarget("env")
		target.SetName(envName)
		value := openapi.NewValueRef()
		value.SetOutput(output)
		mappings = append(mappings, *openapi.NewConnectionMapping(*target, *value))
	}
	conn.SetMappings(mappings)
	return *conn
}

// CreateFullStack creates a stack with 2 resources (api + worker), a volume mount,
// postgres addon connection, secret connection, resource-to-resource connection,
// and depends_on. This exercises all connection kinds and topology features.
func CreateFullStack(name string, addonID string, database string, secretID string) *openapi.Stack {
	api := openapi.NewStackResource(FullStackAPIName)
	apiImage := openapi.NewImageSpec(TestImage)
	api.SetImageSpec(*apiImage)
	api.SetPorts([]openapi.Port{
		*openapi.NewPort("http", int32(FullStackAPIPort), false),
	})
	api.SetDependsOn([]string{FullStackWorkerName})

	worker := openapi.NewStackResource(FullStackWorkerName)
	workerImage := openapi.NewImageSpec(TestImage)
	worker.SetImageSpec(*workerImage)

	volumeSpec := openapi.NewVolumeSpec("1Gi", false, openapi.READ_WRITE_ONCE)
	volume := openapi.NewVolume(FullStackVolumeName, *volumeSpec)

	connections := []openapi.StackConnection{
		postgresEnvConnection(addonID, FullStackAPIName, database, false),
		postgresEnvConnection(addonID, FullStackWorkerName, database, false),
		secretEnvConnection(secretID, FullStackAPIName),
		volumeMountConnection(FullStackVolumeName, FullStackAPIName, FullStackMountPath, FullStackSubPath),
		resourceToResourceEnvConnection(FullStackAPIName, FullStackWorkerName),
	}

	spec := openapi.NewStackSpec([]openapi.StackResource{*api, *worker})
	spec.SetVolumes([]openapi.Volume{*volume})
	spec.SetConnections(connections)
	return openapi.NewStack(name, *spec)
}

func secretEnvConnection(secretID, targetResource string) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef("secret")
	from.SetId(secretID)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(targetResource)
	conn := openapi.NewStackConnection("env", *from, *to)

	var mappings []openapi.ConnectionMapping
	for output, envName := range FullStackSecretEnvMapping {
		target := openapi.NewConnectionTarget("env")
		target.SetName(envName)
		value := openapi.NewValueRef()
		value.SetOutput(output)
		mappings = append(mappings, *openapi.NewConnectionMapping(*target, *value))
	}
	conn.SetMappings(mappings)
	return *conn
}

func volumeMountConnection(volumeName, targetResource, mountPath, subPath string) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef("volume")
	from.SetName(volumeName)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(targetResource)
	conn := openapi.NewStackConnection("volume_mount", *from, *to)
	vc := openapi.NewVolumeMountConfig(mountPath)
	if subPath != "" {
		vc.SubPath = &subPath
	}
	conn.SetConfig(openapi.StackConnectionConfig{VolumeMountConfig: vc})
	return *conn
}

func resourceToResourceEnvConnection(sourceResource, targetResource string) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef("stack_resource")
	from.SetName(sourceResource)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(targetResource)
	conn := openapi.NewStackConnection("env", *from, *to)
	conn.SetMappings([]openapi.ConnectionMapping{
		func() openapi.ConnectionMapping {
			target := openapi.NewConnectionTarget("env")
			target.SetName("API_HOST")
			value := openapi.NewValueRef()
			value.SetOutput("host")
			return *openapi.NewConnectionMapping(*target, *value)
		}(),
		func() openapi.ConnectionMapping {
			target := openapi.NewConnectionTarget("env")
			target.SetName("API_URL")
			value := openapi.NewValueRef()
			value.SetOutput("url.http")
			return *openapi.NewConnectionMapping(*target, *value)
		}(),
	})
	return *conn
}

// CreateStackWithBrokenBuildSource creates a stack pointing at a branch with a broken
// Dockerfile so the kaniko build fails, triggering last_build_failure_detail.
func CreateStackWithBrokenBuildSource(name string, repoURL string, secretID string) *openapi.Stack {
	resource := openapi.NewStackResource(BrokenBuildResourceName)

	gitRepo := openapi.NewBuildSourceContextGitRepo(repoURL)
	gitSecret := openapi.NewSecretRef(secretID)
	gitRepo.SetGitSecret(*gitSecret)

	sourceContext := openapi.NewBuildSourceContext()
	sourceContext.SetGitRepo(*gitRepo)

	branchRevision := openapi.GitRepoRevisionBranch{
		Name: openapi.PtrString(BrokenBuildSourceBranch),
	}
	gitRepoRevision := openapi.NewGitRepoRevision()
	gitRepoRevision.SetBranch(branchRevision)

	sourceRevision := openapi.NewBuildSourceRevision()
	sourceRevision.SetGitRepoRevision(*gitRepoRevision)

	imageRepo := openapi.NewImageRepository()
	imageRepo.SetUseInternalRegistry(true)

	buildSpec := openapi.NewStackResourceBuildSpec(
		*sourceContext,
		BuildSourceContextPath,
		BrokenBuildSourceDockerfile,
		*sourceRevision,
		*imageRepo,
	)
	resource.SetBuildSpec(*buildSpec)

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

// SkipProvisioningAnnotation is the annotation value that causes the stack worker
// to skip cluster provisioning and mark the stack as Ready immediately.
const SkipProvisioningAnnotationKey = "stack.stackdome.io/skip-cluster-provisioning"

func skipProvisioningAnnotation() []openapi.Annotation {
	return []openapi.Annotation{
		*openapi.NewAnnotation(SkipProvisioningAnnotationKey, "true"),
	}
}

func CreateSkipProvisioningStack(name string, resources []openapi.StackResource) *openapi.Stack {
	spec := openapi.NewStackSpec(resources)
	stack := openapi.NewStack(name, *spec)
	stack.SetAnnotations(skipProvisioningAnnotation())
	return stack
}

func CreateSkipProvisioningStackWithVolumes(name string, resources []openapi.StackResource, volumes []openapi.Volume) *openapi.Stack {
	spec := openapi.NewStackSpec(resources)
	spec.SetVolumes(volumes)
	stack := openapi.NewStack(name, *spec)
	stack.SetAnnotations(skipProvisioningAnnotation())
	return stack
}

func CreateSkipProvisioningStackWithConnections(name string, resources []openapi.StackResource, connections []openapi.StackConnection) *openapi.Stack {
	spec := openapi.NewStackSpec(resources)
	spec.SetConnections(connections)
	stack := openapi.NewStack(name, *spec)
	stack.SetAnnotations(skipProvisioningAnnotation())
	return stack
}

func CreateSkipProvisioningStackFull(name string, resources []openapi.StackResource, volumes []openapi.Volume, connections []openapi.StackConnection) *openapi.Stack {
	spec := openapi.NewStackSpec(resources)
	spec.SetVolumes(volumes)
	spec.SetConnections(connections)
	stack := openapi.NewStack(name, *spec)
	stack.SetAnnotations(skipProvisioningAnnotation())
	return stack
}

func SimpleResource(name string) openapi.StackResource {
	r := openapi.NewStackResource(name)
	r.SetImageSpec(*openapi.NewImageSpec(TestImage))
	return *r
}

func ResourceWithPort(name string, port int32) openapi.StackResource {
	r := openapi.NewStackResource(name)
	r.SetImageSpec(*openapi.NewImageSpec(TestImage))
	r.SetPorts([]openapi.Port{*openapi.NewPort("http", port, false)})
	return *r
}

func ResourceWithDependsOn(name string, deps []string) openapi.StackResource {
	r := openapi.NewStackResource(name)
	r.SetImageSpec(*openapi.NewImageSpec(TestImage))
	r.SetDependsOn(deps)
	return *r
}

func EnvConnection(fromType, fromName, toType, toName string, mappings []openapi.ConnectionMapping) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef(fromType)
	from.SetName(fromName)
	to := openapi.NewTopologyNodeRef(toType)
	to.SetName(toName)
	conn := openapi.NewStackConnection("env", *from, *to)
	if len(mappings) > 0 {
		conn.SetMappings(mappings)
	}
	return *conn
}

func EnvConnectionWithID(fromType, fromID, toType, toName string, mappings []openapi.ConnectionMapping) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef(fromType)
	from.SetId(fromID)
	to := openapi.NewTopologyNodeRef(toType)
	to.SetName(toName)
	conn := openapi.NewStackConnection("env", *from, *to)
	if len(mappings) > 0 {
		conn.SetMappings(mappings)
	}
	return *conn
}

func VolumeMountConn(volumeName, targetResource, mountPath, subPath string) openapi.StackConnection {
	from := openapi.NewTopologyNodeRef("volume")
	from.SetName(volumeName)
	to := openapi.NewTopologyNodeRef("stack_resource")
	to.SetName(targetResource)
	conn := openapi.NewStackConnection("volume_mount", *from, *to)
	vc := openapi.NewVolumeMountConfig(mountPath)
	if subPath != "" {
		vc.SubPath = &subPath
	}
	conn.SetConfig(openapi.StackConnectionConfig{VolumeMountConfig: vc})
	return *conn
}

func PostgresMapping(output, envName string) openapi.ConnectionMapping {
	target := openapi.NewConnectionTarget("env")
	target.SetName(envName)
	value := openapi.NewValueRef()
	value.SetOutput(output)
	return *openapi.NewConnectionMapping(*target, *value)
}

func HostMapping() openapi.ConnectionMapping {
	target := openapi.NewConnectionTarget("env")
	target.SetName("API_HOST")
	value := openapi.NewValueRef()
	value.SetOutput("host")
	return *openapi.NewConnectionMapping(*target, *value)
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
		*openapi.NewPort("http", int32(BuildSourcePort), true),
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}
