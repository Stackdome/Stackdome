package stackfile

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/Stackdome/stackdome/pkg/models"
)

const MaxStackfileSize = 1 << 20 // 1MB

// Load parses raw YAML bytes into a Stackfile and validates it.
func Load(content []byte) (*Stackfile, error) {
	if len(content) > MaxStackfileSize {
		return nil, fmt.Errorf("stackfile too large (%d bytes, max %d)", len(content), MaxStackfileSize)
	}
	var sf Stackfile
	if err := yaml.Unmarshal(content, &sf); err != nil {
		return nil, fmt.Errorf("failed to parse stackfile: %w", err)
	}

	if err := Validate(&sf); err != nil {
		return nil, err
	}

	return &sf, nil
}

func Validate(sf *Stackfile) error {
	if sf.Name == "" {
		return fmt.Errorf("stackfile missing required field: name")
	}
	if len(sf.Resources) == 0 {
		return fmt.Errorf("stackfile must define at least one resource")
	}
	if len(sf.Resources) > 50 {
		return fmt.Errorf("stackfile defines too many resources (%d, max 50)", len(sf.Resources))
	}
	for name, res := range sf.Resources {
		if res.Image == "" && res.Build == nil {
			return fmt.Errorf("resource '%s' must have either 'image' or 'build'", name)
		}
		if res.Image != "" && res.Build != nil {
			return fmt.Errorf("resource '%s' cannot have both 'image' and 'build'", name)
		}
		if res.Build != nil {
			if res.Build.Repo == "" {
				return fmt.Errorf("resource '%s' build config missing 'repo'", name)
			}
			set := 0
			if res.Build.Branch != "" {
				set++
			}
			if res.Build.Tag != "" {
				set++
			}
			if res.Build.Commit != "" {
				set++
			}
			if set > 1 {
				return fmt.Errorf("resource '%s' build config: only one of 'branch', 'tag', or 'commit' can be set", name)
			}
			if res.Build.Dockerfile != "" && !strings.HasSuffix(res.Build.Dockerfile, "Dockerfile") && !strings.Contains(res.Build.Dockerfile, "Dockerfile.") && !strings.Contains(res.Build.Dockerfile, "dockerfile") {
				return fmt.Errorf("resource '%s' build config: 'dockerfile' should be a path to a Dockerfile", name)
			}
		}
		if len(res.Ports) > 20 {
			return fmt.Errorf("resource '%s' defines too many ports (%d, max 20)", name, len(res.Ports))
		}
		for _, p := range res.Ports {
			if p.Name == "" {
				return fmt.Errorf("resource '%s' has a port without a name", name)
			}
			if p.Port <= 0 || p.Port > 65535 {
				return fmt.Errorf("resource '%s' port '%s' has invalid port number", name, p.Name)
			}
		}
		for _, vm := range res.Volumes {
			if _, ok := sf.Volumes[vm.Name]; !ok {
				return fmt.Errorf("resource '%s' references undefined volume '%s'", name, vm.Name)
			}
		}

		for _, dep := range res.DependsOn {
			if _, ok := sf.Resources[dep]; !ok {
				return fmt.Errorf("resource '%s' depends_on unknown resource '%s'", name, dep)
			}
			if dep == name {
				return fmt.Errorf("resource '%s' cannot depend on itself", name)
			}
		}

		if err := validateEnvRefs(name, res.Env, res.Ports, sf.Resources); err != nil {
			return err
		}

		for addonName, addon := range res.Addons {
			if err := validateAddonEnv(name, addonName, addon); err != nil {
				return err
			}
		}
	}

	if len(sf.Volumes) > 50 {
		return fmt.Errorf("stackfile defines too many volumes (%d, max 50)", len(sf.Volumes))
	}
	validAccessModes := map[string]bool{
		"":                             true,
		string(models.READ_WRITE_ONCE): true,
		string(models.READ_ONLY_MANY):  true,
		string(models.READ_WRITE_MANY): true,
	}
	for volName, vol := range sf.Volumes {
		if !validAccessModes[vol.AccessMode] {
			return fmt.Errorf("volume '%s' has invalid access_mode '%s'; valid values: ReadWriteOnce, ReadOnlyMany, ReadWriteMany", volName, vol.AccessMode)
		}
		if vol.Size == "" {
			return fmt.Errorf("volume '%s' missing required field: size", volName)
		}
	}

	return nil
}

func validateEnvRefs(resourceName string, env map[string]string, ports []PortDef, allResources map[string]Resource) error {
	for envKey, envVal := range env {
		refs := findRefs(envVal)
		if len(refs) == 0 {
			continue
		}

		// Check that self refs are not mixed with resource refs
		hasSelf := false
		hasResource := false
		for _, ref := range refs {
			if ref.Source == sourceSelf {
				hasSelf = true
			} else {
				hasResource = true
			}
		}
		if hasSelf && hasResource {
			return fmt.Errorf("resource '%s' env var '%s': cannot mix self and resource references in the same value", resourceName, envKey)
		}

		// Self refs must be exact (entire value is the ref)
		if hasSelf {
			if !exactRefPattern.MatchString(envVal) {
				return fmt.Errorf("resource '%s' env var '%s': self-references must be the only content of the env var (e.g., '{{ self.port }}')", resourceName, envKey)
			}
			if err := validateSelfOutput(resourceName, envKey, refs[0].Output, ports); err != nil {
				return err
			}
			continue
		}

		// All resource refs in a single env value must reference the same source
		source := refs[0].Source
		for _, ref := range refs[1:] {
			if ref.Source != source {
				return fmt.Errorf("resource '%s' env var '%s': references multiple resources ('%s' and '%s'). Each env var can only reference one source resource", resourceName, envKey, source, ref.Source)
			}
		}

		// Validate each ref's output against the source resource
		for _, ref := range refs {
			targetRes, ok := allResources[ref.Source]
			if !ok {
				return fmt.Errorf("resource '%s' env var '%s': references resource '%s' which is not defined in the stackfile", resourceName, envKey, ref.Source)
			}
			if err := validateResourceOutput(resourceName, envKey, ref.Source, ref.Output, targetRes.Ports); err != nil {
				return err
			}
		}
	}
	return nil
}

var postgresAddonOutputs = map[string]bool{
	models.OutputNameHost:          true,
	models.OutputNamePort:          true,
	models.OutputNameDatabase:      true,
	models.OutputNameUsername:      true,
	models.OutputNamePassword:      true,
	models.OutputNameSSLMode:       true,
	models.OutputNameCACertificate: true,
	models.OutputNameURL:           true,
}

func addonOutputsForType(addonType string) map[string]bool {
	switch addonType {
	case PostgresAddonType:
		return postgresAddonOutputs
	default:
		return nil
	}
}

func validateAddonEnv(resourceName, addonName string, addon AddonConnectionConfig) error {
	validOutputs := addonOutputsForType(addon.Type)
	if validOutputs == nil {
		return fmt.Errorf("resource '%s' addon '%s': unsupported addon type '%s'", resourceName, addonName, addon.Type)
	}

	for envKey, envVal := range addon.Env {
		refs := findAddonRefs(envVal)
		if len(refs) == 0 {
			return fmt.Errorf("resource '%s' addon '%s' env var '%s': value must use {{ output }} references (e.g., '{{ host }}'), got bare string '%s'", resourceName, addonName, envKey, envVal)
		}
		for _, ref := range refs {
			if !validOutputs[ref.Output] {
				valid := make([]string, 0, len(validOutputs))
				for k := range validOutputs {
					valid = append(valid, k)
				}
				return fmt.Errorf("resource '%s' addon '%s' env var '%s': unknown %s output '%s'. Valid outputs: %s", resourceName, addonName, envKey, addon.Type, ref.Output, strings.Join(valid, ", "))
			}
		}
	}
	return nil
}

func validateResourceOutput(resourceName, envKey, sourceResource, output string, sourcePorts []PortDef) error {
	return validateOutputAgainstPorts(resourceName, envKey, sourceResource, output, sourcePorts)
}

func validateSelfOutput(resourceName, envKey, output string, ports []PortDef) error {
	return validateOutputAgainstPorts(resourceName, envKey, sourceSelf, output, ports)
}

func validateOutputAgainstPorts(resourceName, envKey, source, output string, ports []PortDef) error {
	valid := map[string]struct{}{models.OutputNameHost: {}}
	multiPort := len(ports) > 1
	key := func(base, portName string) string {
		if !multiPort {
			return base
		}
		return base + "." + portName
	}
	for _, p := range ports {
		valid[key(models.OutputNamePort, p.Name)] = struct{}{}
		valid[key(models.OutputNameURL, p.Name)] = struct{}{}
		if p.Public {
			valid[key(models.OutputNamePublicHost, p.Name)] = struct{}{}
			valid[key(models.OutputNamePublicURL, p.Name)] = struct{}{}
		}
	}

	if _, ok := valid[output]; ok {
		return nil
	}

	label := "resource '" + source + "'"
	if source == sourceSelf {
		label = "resource '" + resourceName + "'"
	}
	names := make([]string, 0, len(valid))
	for k := range valid {
		names = append(names, k)
	}
	sort.Strings(names)
	return fmt.Errorf("resource '%s' env var '%s': unknown output '%s' on %s. Valid outputs: %s",
		resourceName, envKey, output, label, strings.Join(names, ", "))
}
