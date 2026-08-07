package stackfile

import (
	"fmt"
	"sort"

	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
	"k8s.io/utils/ptr"
)

// ToStack converts a validated Stackfile into an API Stack document,
// ready for the stack apply endpoint. Output is deterministic: resources,
// volumes, connections, and mappings are emitted in sorted order.
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
	for _, name := range sortedKeys(sf.Resources) {
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
	// Validate guarantees: branch and tag are exclusive, commit requires one
	// of them (matching the API's git_commit_requires_ref rule).
	if b.Branch != "" {
		source.SetBranch(b.Branch)
	}
	if b.Tag != "" {
		source.SetTag(b.Tag)
	}
	if b.Commit != "" {
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
			ev.SelfOutput = ptr.To(extractSelfOutput(value))
		case hasResourceRef(value):
			// Becomes an env connection; see buildEnvRefConnections.
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
	for _, name := range sortedKeys(sf.Volumes) {
		v := sf.Volumes[name]
		accessMode := openapi.VolumeAccessMode(defaultAccessMode)
		if v.AccessMode != "" {
			accessMode = openapi.VolumeAccessMode(v.AccessMode)
		}
		volumes = append(volumes, openapi.Volume{
			Name: name,
			Spec: openapi.VolumeSpec{
				Size:               v.Size,
				AccessMode:         accessMode,
				NeedsSyncBeforeUse: false,
			},
		})
	}
	return volumes
}
