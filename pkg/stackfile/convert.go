package stackfile

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/samber/lo"
	"k8s.io/utils/ptr"
)

var (
	// Matches a full-value ref: the entire string is {{ source.output }}
	exactRefPattern = regexp.MustCompile(`^\{\{\s*([\w-]+(?:\.[\w-]+)+)\s*\}\}$`)
	// Matches embedded refs within a larger string
	embeddedRefPattern = regexp.MustCompile(`\{\{\s*([\w-]+(?:\.[\w-]+)+)\s*\}\}`)
	// addonVarPattern matches {{ varname }} in addon env templates.
	// Addon vars are plain output names (host, port, username) — no source prefix.
	addonVarPattern = regexp.MustCompile(`\{\{\s*([\w-]+)\s*\}\}`)
)

func (sf *Stackfile) ToStack() (openapi.Stack, error) {
	resources, err := sf.buildResources()
	if err != nil {
		return openapi.Stack{}, err
	}
	spec := openapi.StackSpec{
		StackResources: resources,
		Volumes:        sf.buildVolumes(),
		Connections:    sf.buildConnections(),
	}
	return openapi.Stack{
		Name: sf.Name,
		Spec: spec,
	}, nil
}

func (sf *Stackfile) buildResources() ([]openapi.StackResource, error) {
	resources := make([]openapi.StackResource, 0, len(sf.Resources))
	names := make([]string, 0, len(sf.Resources))
	for name := range sf.Resources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		res := sf.Resources[name]
		sr := openapi.StackResource{
			Name:      name,
			DependsOn: res.DependsOn,
		}

		if res.Image != "" {
			sr.Source = &openapi.SourceSpec{Image: openapi.NewImageSource(res.Image)}
		}

		if res.Build != nil {
			source, err := buildGitSource(res.Build)
			if err != nil {
				return nil, fmt.Errorf("resource %q: %w", name, err)
			}
			sr.Source = &openapi.SourceSpec{Git: source}
		}

		if res.WorkloadType != "" {
			sr.WorkloadType = ptr.To(res.WorkloadType)
		}
		if res.Schedule != "" {
			sr.Schedule = ptr.To(res.Schedule)
		}
		sr.Replicas = res.Replicas

		sr.Ports = buildPorts(res.Ports)
		sr.ExecutionConfig = buildExecutionConfig(res.Env, res.Command, res.Args)
		sr.VolumeMounts = buildVolumeMounts(res.Volumes)

		resources = append(resources, sr)
	}
	return resources, nil
}

func buildGitSource(b *BuildConfig) (*openapi.GitSource, error) {
	if b.GitSecret != "" {
		return nil, fmt.Errorf("git_secret is no longer supported; configure clone credentials on the preview config or an org-level git integration")
	}

	source := openapi.NewGitSource(b.Repo)
	if b.Context != "" {
		source.SetBuildContext(b.Context)
	}
	if b.Dockerfile != "" {
		source.SetDockerfilePath(b.Dockerfile)
	}
	switch {
	case b.Branch != "":
		source.SetBranch(b.Branch)
		if b.Commit != "" {
			source.SetCommit(b.Commit)
		}
	case b.Tag != "":
		source.SetTag(b.Tag)
		if b.Commit != "" {
			source.SetCommit(b.Commit)
		}
	case b.Commit != "":
		source.SetCommit(b.Commit)
	}
	// No branch or tag: the server resolves the repository's default branch.
	return source, nil
}

func buildPorts(ports []PortDef) []openapi.Port {
	if len(ports) == 0 {
		return nil
	}
	out := make([]openapi.Port, len(ports))
	for i, p := range ports {
		port := openapi.Port{
			Name:            p.Name,
			Number:          p.Port,
			ExposedToPublic: p.Public,
		}
		if p.Protocol != "" {
			port.Protocol = ptr.To(p.Protocol)
		}
		if p.Subdomain != "" {
			port.SubdomainPrefix = ptr.To(p.Subdomain)
		}
		out[i] = port
	}
	return out
}

func buildExecutionConfig(env map[string]string, command, args []string) *openapi.ExecutionConfig {
	if len(env) == 0 && len(command) == 0 && len(args) == 0 {
		return nil
	}

	cfg := &openapi.ExecutionConfig{}

	if len(command) > 0 {
		cfg.Command = command
	}
	if len(args) > 0 {
		cfg.Args = args
	}

	var envVars []openapi.EnvVar
	for name, value := range env {
		ev := openapi.EnvVar{Name: name}
		switch {
		case isSelfRef(value):
			output := extractSelfOutput(value)
			ev.SelfOutput = ptr.To(output)
		case hasResourceRef(value):
			continue
		default:
			ev.Value = ptr.To(value)
		}
		envVars = append(envVars, ev)
	}

	sort.Slice(envVars, func(i, j int) bool {
		return envVars[i].Name < envVars[j].Name
	})
	if len(envVars) > 0 {
		cfg.EnvironmentVariables = envVars
	}

	return cfg
}

type envRef struct {
	Source   string
	Output   string
	RawMatch string // the exact substring matched, e.g. "{{ redis.host }}"
}

func findRefs(value string) []envRef {
	matches := embeddedRefPattern.FindAllStringSubmatch(value, -1)
	var refs []envRef
	for _, m := range matches {
		parts := strings.SplitN(m[1], ".", 2)
		if len(parts) == 2 {
			refs = append(refs, envRef{Source: parts[0], Output: parts[1], RawMatch: m[0]})
		}
	}
	return refs
}

func isExactRef(value string) bool {
	return exactRefPattern.MatchString(value)
}

func isSelfRef(value string) bool {
	refs := findRefs(value)
	for _, r := range refs {
		if r.Source == "self" {
			return true
		}
	}
	return false
}

func extractSelfOutput(value string) string {
	refs := findRefs(value)
	for _, r := range refs {
		if r.Source == "self" {
			return r.Output
		}
	}
	return ""
}

func hasResourceRef(value string) bool {
	refs := findRefs(value)
	for _, r := range refs {
		if r.Source != "self" {
			return true
		}
	}
	return false
}

func outputToVarName(output string) string {
	return strings.ReplaceAll(output, ".", "_")
}

func buildVolumeMounts(mounts []VolumeMountDef) []openapi.VolumeMount {
	if len(mounts) == 0 {
		return nil
	}
	out := make([]openapi.VolumeMount, len(mounts))
	for i, m := range mounts {
		out[i] = openapi.VolumeMount{
			SourceVolumeName: m.Name,
			TargetPath:       m.Path,
		}
	}
	return out
}

func (sf *Stackfile) buildVolumes() []openapi.Volume {
	if len(sf.Volumes) == 0 {
		return nil
	}
	volumes := make([]openapi.Volume, 0, len(sf.Volumes))
	names := make([]string, 0, len(sf.Volumes))
	for name := range sf.Volumes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		v := sf.Volumes[name]
		accessMode := openapi.VolumeAccessMode("ReadWriteOnce")
		if v.AccessMode != "" {
			accessMode = openapi.VolumeAccessMode(v.AccessMode)
		}
		vol := openapi.Volume{
			Name: name,
			Spec: openapi.VolumeSpec{
				Size:               v.Size,
				AccessMode:         accessMode,
				NeedsSyncBeforeUse: false,
			},
		}
		volumes = append(volumes, vol)
	}
	return volumes
}

func (sf *Stackfile) buildConnections() []openapi.StackConnection {
	var connections []openapi.StackConnection

	names := make([]string, 0, len(sf.Resources))
	for name := range sf.Resources {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, resourceName := range names {
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

	for envName, value := range env {
		refsInCurrentEnv := findRefs(value)
		if len(refsInCurrentEnv) == 0 {
			continue
		}

		_, currentEnvIsSelfRef := lo.Find(refsInCurrentEnv, func(r envRef) bool { return r.Source == "self" })
		if currentEnvIsSelfRef {
			// skip self refs, they will be handled in execution config
			continue
		}

		source := refsInCurrentEnv[0].Source

		var vr openapi.ValueRef
		if isExactRef(value) && len(refsInCurrentEnv) == 1 {
			vr.Output = ptr.To(refsInCurrentEnv[0].Output)
		} else {
			tmpl := value
			values := make(map[string]openapi.OutputValueRef)
			for _, r := range refsInCurrentEnv {
				varName := outputToVarName(r.Output)
				tmpl = strings.Replace(tmpl, r.RawMatch, "{{ "+varName+" }}", 1)
				values[varName] = openapi.OutputValueRef{Output: r.Output}
			}
			vr.Template = ptr.To(tmpl)
			vr.Values = &values
		}

		mapping := openapi.ConnectionMapping{
			Target: openapi.ConnectionTarget{
				Type: "env",
				Name: ptr.To(envName),
			},
			Value: vr,
		}
		grouped[source] = append(grouped[source], mapping)
	}

	var connections []openapi.StackConnection
	for source, mappings := range grouped {
		conn := openapi.StackConnection{
			Kind: "env",
			From: openapi.TopologyNodeRef{
				Type: "stack_resource",
				Name: ptr.To(source),
			},
			To: openapi.TopologyNodeRef{
				Type: "stack_resource",
				Name: ptr.To(targetResource),
			},
			Mappings: mappings,
		}
		connections = append(connections, conn)
	}
	return connections
}

func buildSecretConnections(targetResource string, secrets map[string]SecretMapping) []openapi.StackConnection {
	var connections []openapi.StackConnection

	for secretName, mapping := range secrets {
		var mappings []openapi.ConnectionMapping
		for envName, secretKey := range mapping {
			mappings = append(mappings, openapi.ConnectionMapping{
				Target: openapi.ConnectionTarget{
					Type: "env",
					Name: ptr.To(envName),
				},
				Value: openapi.ValueRef{
					Output: ptr.To(secretKey),
				},
			})
		}

		conn := openapi.StackConnection{
			Kind: "env",
			From: openapi.TopologyNodeRef{
				Type: "secret",
				Name: ptr.To(secretName),
			},
			To: openapi.TopologyNodeRef{
				Type: "stack_resource",
				Name: ptr.To(targetResource),
			},
			Mappings: mappings,
		}
		connections = append(connections, conn)
	}
	return connections
}

func buildAddonConnections(targetResource string, addons map[string]AddonConnectionConfig) []openapi.StackConnection {
	var connections []openapi.StackConnection

	for addonName, addon := range addons {
		mappings := buildAddonMappings(addon.Env)

		conn := openapi.StackConnection{
			Kind: "env",
			From: openapi.TopologyNodeRef{
				Type: "addon/" + addon.Type,
				Name: ptr.To(addonName),
			},
			To: openapi.TopologyNodeRef{
				Type: "stack_resource",
				Name: ptr.To(targetResource),
			},
			Mappings: mappings,
			Config:   buildAddonConfig(addon),
		}

		connections = append(connections, conn)
	}
	return connections
}

type addonRef struct {
	Output   string
	RawMatch string
}

func findAddonRefs(value string) []addonRef {
	matches := addonVarPattern.FindAllStringSubmatch(value, -1)
	var refs []addonRef
	for _, m := range matches {
		refs = append(refs, addonRef{Output: m[1], RawMatch: m[0]})
	}
	return refs
}

func buildAddonMappings(env map[string]string) []openapi.ConnectionMapping {
	var mappings []openapi.ConnectionMapping
	for envName, envValue := range env {
		refs := findAddonRefs(envValue)

		var vr openapi.ValueRef
		switch {
		case len(refs) == 1 && refs[0].RawMatch == envValue:
			vr.Output = ptr.To(refs[0].Output)
		default:
			values := make(map[string]openapi.OutputValueRef)
			for _, r := range refs {
				values[r.Output] = openapi.OutputValueRef{Output: r.Output}
			}
			vr.Template = ptr.To(envValue)
			vr.Values = &values
		}

		mappings = append(mappings, openapi.ConnectionMapping{
			Target: openapi.ConnectionTarget{
				Type: "env",
				Name: ptr.To(envName),
			},
			Value: vr,
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
		conn := openapi.StackConnection{
			Kind: "volume_mount",
			From: openapi.TopologyNodeRef{
				Type: "volume",
				Name: ptr.To(m.Name),
			},
			To: openapi.TopologyNodeRef{
				Type: "stack_resource",
				Name: ptr.To(targetResource),
			},
			Config: &openapi.StackConnectionConfig{
				VolumeMountConfig: &openapi.VolumeMountConfig{
					MountPath: m.Path,
				},
			},
		}
		connections = append(connections, conn)
	}
	return connections
}
