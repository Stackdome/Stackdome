package stackfile

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/Stackdome/stackdome/pkg/models"
)

const (
	PostgresAddonType = "postgres"

	sourceSelf = "self"

	defaultAccessMode = string(models.READ_WRITE_ONCE)

	connectionKindEnv         = string(models.ConnectionKindEnv)
	connectionKindVolumeMount = string(models.ConnectionKindVolumeMount)
	nodeTypeStackResource     = string(models.TopologyNodeTypeStackResource)
	nodeTypeSecret            = string(models.TopologyNodeTypeSecret)
	nodeTypeVolume            = string(models.TopologyNodeTypeVolume)
	nodeTypePostgresAddon     = string(models.TopologyNodeTypePostgresAddon)
	targetTypeEnv             = string(models.ConnectionTargetTypeEnv)
)

type Stackfile struct {
	Name      string               `yaml:"name"`
	Resources map[string]Resource  `yaml:"resources"`
	Volumes   map[string]VolumeDef `yaml:"volumes,omitempty"`
}

type Resource struct {
	Image   string            `yaml:"image,omitempty"`
	Build   *BuildConfig      `yaml:"build,omitempty"`
	Command []string          `yaml:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
	Ports   []PortDef         `yaml:"ports,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	// Secret Name -> Mapping of secret keys to env var names
	Secrets map[string]SecretMapping `yaml:"secrets,omitempty"`
	// Addon Name -> Connection Config
	Addons       map[string]AddonConnectionConfig `yaml:"addons,omitempty"`
	Volumes      []VolumeMountDef                 `yaml:"volumes,omitempty"`
	DependsOn    []string                         `yaml:"depends_on,omitempty"`
	WorkloadType string                           `yaml:"workload_type,omitempty"`
	Schedule     string                           `yaml:"schedule,omitempty"`
	Replicas     *int32                           `yaml:"replicas,omitempty"`
}

type BuildConfig struct {
	Repo       string `yaml:"repo,omitempty"`
	Branch     string `yaml:"branch,omitempty"`
	Tag        string `yaml:"tag,omitempty"`
	Commit     string `yaml:"commit,omitempty"`
	Dockerfile string `yaml:"dockerfile,omitempty"`
	Context    string `yaml:"context,omitempty"`
	GitSecret  string `yaml:"git_secret,omitempty"`
}

type PortDef struct {
	Name      string `yaml:"name"`
	Port      int32  `yaml:"port"`
	Protocol  string `yaml:"protocol,omitempty"`
	Public    bool   `yaml:"public,omitempty"`
	Subdomain string `yaml:"subdomain,omitempty"`
}

type VolumeDef struct {
	Size       string `yaml:"size"`
	AccessMode string `yaml:"access_mode,omitempty"`
}

type VolumeMountDef struct {
	Name string `yaml:"name"`
	Path string `yaml:"path"`
}

// SecretMapping maps secret keys to env var names.
// Key = env var name, Value = secret key
type SecretMapping map[string]string

type AddonConnectionConfig struct {
	Type string `yaml:"type"`
	// Env vars templated from the addon's outputs. Output names are bare —
	// no addon prefix (the block already names the addon):
	// env:
	//   POSTGRES_HOST: "{{ host }}"
	//   POSTGRES_URI: "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}"
	Env      map[string]string    `yaml:"env"`
	Postgres *PostgresAddonConfig `yaml:"-"`

	rawNode yaml.Node `yaml:"-"`
}

type PostgresAddonConfig struct {
	Database  string `yaml:"database,omitempty"`
	Superuser bool   `yaml:"superuser,omitempty"`
}

// addonConfigFields are the keys allowed in an addon block. The custom
// unmarshaler bypasses the decoder's KnownFields check, so unknown keys
// are rejected here instead.
var addonConfigFields = map[string]bool{
	"type": true, "env": true, "database": true, "superuser": true, //nolint:goconst // yaml field names, not a shared constant

}

func (a *AddonConnectionConfig) UnmarshalYAML(node *yaml.Node) error {
	a.rawNode = *node

	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content)-1; i += 2 {
			if key := node.Content[i].Value; !addonConfigFields[key] {
				return fmt.Errorf("unknown field %q in addon config", key)
			}
		}
	}

	var base struct {
		Type string            `yaml:"type"`
		Env  map[string]string `yaml:"env"`
	}
	if err := node.Decode(&base); err != nil {
		return err
	}
	a.Type = base.Type
	a.Env = base.Env

	switch a.Type {
	case PostgresAddonType:
		var pg PostgresAddonConfig
		if err := node.Decode(&pg); err != nil {
			return err
		}
		a.Postgres = &pg
	}

	return nil
}
