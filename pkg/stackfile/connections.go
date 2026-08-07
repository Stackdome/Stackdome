package stackfile

import (
	"strings"

	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
	"k8s.io/utils/ptr"
)

// buildConnections synthesizes the stack's topology edges:
//   - env vars referencing another resource -> env connections
//   - secret mappings -> env connections from secrets
//   - addon blocks -> env connections from addons (with per-type config)
//   - volume mounts -> volume_mount connections
func (sf *Stackfile) buildConnections() []openapi.StackConnection {
	var connections []openapi.StackConnection
	for _, resourceName := range sortedKeys(sf.Resources) {
		res := sf.Resources[resourceName]
		connections = append(connections, buildEnvRefConnections(resourceName, res.Env)...)
		connections = append(connections, buildSecretConnections(resourceName, res.Secrets)...)
		connections = append(connections, buildAddonConnections(resourceName, res.Addons)...)
		connections = append(connections, buildVolumeMountConnections(resourceName, res.Volumes)...)
	}
	return connections
}

func buildEnvRefConnections(targetResource string, env map[string]string) []openapi.StackConnection {
	grouped := make(map[string][]openapi.ConnectionMapping)

	for _, envName := range sortedKeys(env) {
		value := env[envName]
		refs := findRefs(value)
		if len(refs) == 0 || isSelfRef(value) {
			// Literals and self refs live in execution config, not connections.
			continue
		}

		// Validate guarantees all refs in one value share a source.
		source := refs[0].Source

		var vr openapi.ValueRef
		if isExactRef(value) && len(refs) == 1 {
			vr.Output = ptr.To(refs[0].Output)
		} else {
			tmpl := value
			values := make(map[string]openapi.OutputValueRef)
			for _, r := range refs {
				varName := outputToVarName(r.Output)
				tmpl = strings.Replace(tmpl, r.RawMatch, "{{ "+varName+" }}", 1)
				values[varName] = openapi.OutputValueRef{Output: r.Output}
			}
			vr.Template = ptr.To(tmpl)
			vr.Values = &values
		}

		grouped[source] = append(grouped[source], openapi.ConnectionMapping{
			Target: envTarget(envName),
			Value:  vr,
		})
	}

	var connections []openapi.StackConnection
	for _, source := range sortedKeys(grouped) {
		connections = append(connections, openapi.StackConnection{
			Kind:     connectionKindEnv,
			From:     nodeRef(nodeTypeStackResource, source),
			To:       nodeRef(nodeTypeStackResource, targetResource),
			Mappings: grouped[source],
		})
	}
	return connections
}

func buildSecretConnections(targetResource string, secrets map[string]SecretMapping) []openapi.StackConnection {
	var connections []openapi.StackConnection

	for _, secretName := range sortedKeys(secrets) {
		mapping := secrets[secretName]
		mappings := make([]openapi.ConnectionMapping, 0, len(mapping))
		for _, envName := range sortedKeys(mapping) {
			mappings = append(mappings, openapi.ConnectionMapping{
				Target: envTarget(envName),
				Value:  openapi.ValueRef{Output: ptr.To(mapping[envName])},
			})
		}

		connections = append(connections, openapi.StackConnection{
			Kind:     connectionKindEnv,
			From:     nodeRef(nodeTypeSecret, secretName),
			To:       nodeRef(nodeTypeStackResource, targetResource),
			Mappings: mappings,
		})
	}
	return connections
}

func buildAddonConnections(targetResource string, addons map[string]AddonConnectionConfig) []openapi.StackConnection {
	var connections []openapi.StackConnection

	for _, addonName := range sortedKeys(addons) {
		addon := addons[addonName]
		connections = append(connections, openapi.StackConnection{
			Kind:     connectionKindEnv,
			From:     nodeRef("addon/"+addon.Type, addonName),
			To:       nodeRef(nodeTypeStackResource, targetResource),
			Mappings: buildAddonMappings(addon.Env),
			Config:   buildAddonConfig(addon),
		})
	}
	return connections
}

func buildAddonMappings(env map[string]string) []openapi.ConnectionMapping {
	mappings := make([]openapi.ConnectionMapping, 0, len(env))
	for _, envName := range sortedKeys(env) {
		envValue := env[envName]
		refs := findAddonRefs(envValue)

		var vr openapi.ValueRef
		if len(refs) == 1 && refs[0].RawMatch == envValue {
			vr.Output = ptr.To(refs[0].Output)
		} else {
			values := make(map[string]openapi.OutputValueRef)
			for _, r := range refs {
				values[r.Output] = openapi.OutputValueRef{Output: r.Output}
			}
			vr.Template = ptr.To(envValue)
			vr.Values = &values
		}

		mappings = append(mappings, openapi.ConnectionMapping{
			Target: envTarget(envName),
			Value:  vr,
		})
	}
	return mappings
}

func buildAddonConfig(addon AddonConnectionConfig) *openapi.StackConnectionConfig {
	if addon.Postgres == nil {
		return nil
	}
	pg := addon.Postgres
	if pg.Database == "" && !pg.Superuser {
		return nil
	}
	pgConfig := &openapi.PostgresEnvConfig{}
	if pg.Database != "" {
		pgConfig.Database = ptr.To(pg.Database)
	}
	if pg.Superuser {
		pgConfig.Superuser = ptr.To(true)
	}
	return &openapi.StackConnectionConfig{
		PostgresEnvConfig: pgConfig,
	}
}

func buildVolumeMountConnections(targetResource string, mounts []VolumeMountDef) []openapi.StackConnection {
	var connections []openapi.StackConnection
	for _, m := range mounts {
		connections = append(connections, openapi.StackConnection{
			Kind: connectionKindVolumeMount,
			From: nodeRef(nodeTypeVolume, m.Name),
			To:   nodeRef(nodeTypeStackResource, targetResource),
			Config: &openapi.StackConnectionConfig{
				VolumeMountConfig: &openapi.VolumeMountConfig{
					MountPath: m.Path,
				},
			},
		})
	}
	return connections
}

func nodeRef(nodeType, name string) openapi.TopologyNodeRef {
	return openapi.TopologyNodeRef{Type: nodeType, Name: ptr.To(name)}
}

func envTarget(envName string) openapi.ConnectionTarget {
	return openapi.ConnectionTarget{Type: targetTypeEnv, Name: ptr.To(envName)}
}
