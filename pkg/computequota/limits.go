package computequota

// ComputeLimits is runtime configuration for an organisation's use of managed
// compute. It is a value object and is not persisted.
type ComputeLimits struct {
	MaxStacksPerOrganization         int64  `yaml:"maxStacksPerOrganization" json:"max_stacks_per_organization"`
	MaxStackResourcesPerOrganization int64  `yaml:"maxStackResourcesPerOrganization" json:"max_stack_resources_per_organization"`
	ReplicasPerStackResource         int32  `yaml:"replicasPerStackResource" json:"replicas_per_stack_resource"`
	MaxVolumesPerOrganization        int64  `yaml:"maxVolumesPerOrganization" json:"max_volumes_per_organization"`
	MaxVolumeSize                    string `yaml:"maxVolumeSize" json:"max_volume_size"`
	VolumeStorageClass               string `yaml:"volumeStorageClass" json:"volume_storage_class"`
	MaxPostgresAddonsPerOrganization int64  `yaml:"maxPostgresAddonsPerOrganization" json:"max_postgres_addons_per_organization"`
	PostgresInstances                int    `yaml:"postgresInstances" json:"postgres_instances"`
	MaxPostgresStorageSize           string `yaml:"maxPostgresStorageSize" json:"max_postgres_storage_size"`
	// ConcurrentBuilds is informational here; request quota policy does not
	// enforce build scheduling.
	ConcurrentBuilds int `yaml:"concurrentBuilds" json:"concurrent_builds"`
}
