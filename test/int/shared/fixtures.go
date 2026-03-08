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
