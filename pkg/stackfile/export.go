package stackfile

import (
	"fmt"
	"strings"

	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
)

// FromStack converts an API Stack document back into a Stackfile.
//
// It covers exactly the subset a stackfile can express. Constructs outside
// that subset (init containers, file targets, build-artifact connections,
// volume sources, registry credentials, ...) fail with an error naming the
// construct — never silently dropped. Connections must carry name-based
// node refs; id-only refs (a stack fetched after secret/addon resolution)
// need the names restored first.
func FromStack(stack *openapi.Stack) (*Stackfile, error) {
	sf := &Stackfile{
		Name:      stack.Name,
		Resources: make(map[string]Resource, len(stack.Spec.StackResources)),
	}

	for _, v := range stack.Spec.Volumes {
		vol, err := exportVolume(v)
		if err != nil {
			return nil, err
		}
		if sf.Volumes == nil {
			sf.Volumes = make(map[string]VolumeDef)
		}
		sf.Volumes[v.Name] = vol
	}

	for _, sr := range stack.Spec.StackResources {
		res, err := exportResource(sr)
		if err != nil {
			return nil, err
		}
		sf.Resources[sr.Name] = res
	}

	for _, conn := range stack.Spec.Connections {
		if err := applyConnection(sf, conn); err != nil {
			return nil, err
		}
	}

	if err := Validate(sf); err != nil {
		return nil, fmt.Errorf("exported stackfile failed validation: %w", err)
	}
	return sf, nil
}

func exportVolume(v openapi.Volume) (VolumeDef, error) {
	spec := v.Spec
	switch {
	case spec.StorageClass != nil && *spec.StorageClass != "":
		return VolumeDef{}, unsupported("volume %q: storage_class", v.Name)
	case spec.NeedsSyncBeforeUse:
		return VolumeDef{}, unsupported("volume %q: needs_sync_before_use", v.Name)
	case spec.Source != nil:
		return VolumeDef{}, unsupported("volume %q: volume source", v.Name)
	}
	vol := VolumeDef{Size: spec.Size}
	if string(spec.AccessMode) != defaultAccessMode {
		vol.AccessMode = string(spec.AccessMode)
	}
	return vol, nil
}

func exportResource(sr openapi.StackResource) (Resource, error) {
	res := Resource{DependsOn: sr.DependsOn}

	switch {
	case sr.InitSpec != nil:
		return res, unsupported("resource %q: init_spec", sr.Name)
	case sr.LifecycleConfig != nil:
		return res, unsupported("resource %q: lifecycle_config", sr.Name)
	case len(sr.Labels) > 0 || len(sr.Annotations) > 0:
		return res, unsupported("resource %q: labels/annotations", sr.Name)
	}

	if err := exportSource(sr.Name, sr.Source, &res); err != nil {
		return res, err
	}

	if sr.WorkloadType != nil {
		res.WorkloadType = *sr.WorkloadType
	}
	if sr.Schedule != nil {
		res.Schedule = *sr.Schedule
	}
	res.Replicas = sr.Replicas

	for _, p := range sr.Ports {
		port := PortDef{Name: p.Name, Port: p.Number, Public: p.ExposedToPublic}
		if p.Protocol != nil {
			port.Protocol = *p.Protocol
		}
		if p.SubdomainPrefix != nil {
			port.Subdomain = *p.SubdomainPrefix
		}
		res.Ports = append(res.Ports, port)
	}

	if err := exportExecutionConfig(sr.Name, sr.ExecutionConfig, &res); err != nil {
		return res, err
	}

	for _, vm := range sr.VolumeMounts {
		if vm.SourceSubPath != nil && *vm.SourceSubPath != "" {
			return res, unsupported("resource %q: volume mount source_sub_path", sr.Name)
		}
		res.Volumes = append(res.Volumes, VolumeMountDef{Name: vm.SourceVolumeName, Path: vm.TargetPath})
	}

	return res, nil
}

func exportSource(name string, source *openapi.SourceSpec, res *Resource) error {
	if source == nil {
		return unsupported("resource %q: missing source", name)
	}
	switch {
	case source.Image != nil:
		if source.Image.RegistryCredentialsId != nil {
			return unsupported("resource %q: image registry_credentials_id", name)
		}
		res.Image = source.Image.Ref

	case source.Git != nil:
		git := source.Git
		if git.IntegrationId != nil {
			return unsupported("resource %q: git integration_id", name)
		}
		if git.Push != nil {
			return unsupported("resource %q: git push target", name)
		}
		res.Build = &BuildConfig{
			Repo:       git.RepoUrl,
			Branch:     git.GetBranch(),
			Tag:        git.GetTag(),
			Commit:     git.GetCommit(),
			Dockerfile: git.GetDockerfilePath(),
			Context:    git.GetBuildContext(),
		}

	case source.Volume != nil:
		return unsupported("resource %q: volume build source", name)

	default:
		return unsupported("resource %q: empty source", name)
	}
	return nil
}

func exportExecutionConfig(name string, cfg *openapi.ExecutionConfig, res *Resource) error {
	if cfg == nil {
		return nil
	}
	res.Command = cfg.Command
	res.Args = cfg.Args
	for _, ev := range cfg.EnvironmentVariables {
		switch {
		case ev.SelfOutput != nil:
			setEnv(res, ev.Name, "{{ "+sourceSelf+"."+*ev.SelfOutput+" }}")
		case ev.Value != nil:
			setEnv(res, ev.Name, *ev.Value)
		default:
			return unsupported("resource %q: env var %q has neither value nor self_output", name, ev.Name)
		}
	}
	return nil
}

// applyConnection folds a topology edge back into the target resource's
// env / secrets / addons / volumes blocks.
func applyConnection(sf *Stackfile, conn openapi.StackConnection) error {
	to, err := refName(conn.To)
	if err != nil {
		return err
	}
	res, ok := sf.Resources[to]
	if !ok {
		return fmt.Errorf("connection targets unknown resource %q", to)
	}

	switch conn.Kind {
	case connectionKindVolumeMount:
		// Volume mounts are already exported from the resource's volume_mounts;
		// the connection is the same fact restated. Just check they agree.
		return checkVolumeMountConnection(res, conn)
	case connectionKindEnv:
	default:
		return unsupported("connection kind %q", conn.Kind)
	}

	from, err := refName(conn.From)
	if err != nil {
		return err
	}

	switch {
	case conn.From.Type == nodeTypeStackResource:
		err = applyResourceEnvConnection(&res, from, conn.Mappings)
	case conn.From.Type == nodeTypeSecret:
		err = applySecretConnection(&res, from, conn.Mappings)
	case strings.HasPrefix(conn.From.Type, "addon/"):
		err = applyAddonConnection(&res, from, strings.TrimPrefix(conn.From.Type, "addon/"), conn.Mappings, conn.Config)
	default:
		err = unsupported("connection from node type %q", conn.From.Type)
	}
	if err != nil {
		return err
	}

	sf.Resources[to] = res
	return nil
}

func applyResourceEnvConnection(res *Resource, from string, mappings []openapi.ConnectionMapping) error {
	for _, m := range mappings {
		envName, err := envTargetName(m.Target)
		if err != nil {
			return err
		}
		value, err := valueRefToDSL(from, m.Value)
		if err != nil {
			return err
		}
		if err := setEnvUnique(res, envName, value); err != nil {
			return err
		}
	}
	return nil
}

func applySecretConnection(res *Resource, secretName string, mappings []openapi.ConnectionMapping) error {
	for _, m := range mappings {
		envName, err := envTargetName(m.Target)
		if err != nil {
			return err
		}
		if m.Value.Output == nil {
			return unsupported("secret %q: templated secret values", secretName)
		}
		if res.Secrets == nil {
			res.Secrets = make(map[string]SecretMapping)
		}
		if res.Secrets[secretName] == nil {
			res.Secrets[secretName] = make(SecretMapping)
		}
		res.Secrets[secretName][envName] = *m.Value.Output
	}
	return nil
}

func applyAddonConnection(res *Resource, addonName, addonType string, mappings []openapi.ConnectionMapping, config *openapi.StackConnectionConfig) error {
	addon := AddonConnectionConfig{Type: addonType, Env: make(map[string]string)}

	if config != nil {
		if config.PostgresEnvConfig == nil {
			return unsupported("addon %q: non-postgres connection config", addonName)
		}
		pg := config.PostgresEnvConfig
		if pg.CredentialScope != nil {
			return unsupported("addon %q: credential_scope", addonName)
		}
		addon.Postgres = &PostgresAddonConfig{
			Database:  pg.GetDatabase(),
			Superuser: pg.GetSuperuser(),
		}
	}

	for _, m := range mappings {
		envName, err := envTargetName(m.Target)
		if err != nil {
			return err
		}
		value, err := addonValueRefToDSL(addonName, m.Value)
		if err != nil {
			return err
		}
		addon.Env[envName] = value
	}

	if res.Addons == nil {
		res.Addons = make(map[string]AddonConnectionConfig)
	}
	res.Addons[addonName] = addon
	return nil
}

// valueRefToDSL renders a mapping value back into env DSL, e.g.
// output "host" from "redis" -> "{{ redis.host }}", and a template
// "redis://{{ host_val }}" back into "redis://{{ redis.host }}".
func valueRefToDSL(from string, v openapi.ValueRef) (string, error) {
	if v.Output != nil {
		return "{{ " + from + "." + *v.Output + " }}", nil
	}
	if v.Template == nil || v.Values == nil {
		return "", unsupported("connection from %q: mapping without output or template", from)
	}
	rendered := *v.Template
	for key, ref := range *v.Values {
		rendered = templateVarPattern(key).ReplaceAllString(rendered, "{{ "+from+"."+ref.Output+" }}")
	}
	return rendered, nil
}

func addonValueRefToDSL(addonName string, v openapi.ValueRef) (string, error) {
	if v.Output != nil {
		return "{{ " + *v.Output + " }}", nil
	}
	if v.Template == nil || v.Values == nil {
		return "", unsupported("addon %q: mapping without output or template", addonName)
	}
	rendered := *v.Template
	for key, ref := range *v.Values {
		rendered = templateVarPattern(key).ReplaceAllString(rendered, "{{ "+ref.Output+" }}")
	}
	return rendered, nil
}

func checkVolumeMountConnection(res Resource, conn openapi.StackConnection) error {
	from, err := refName(conn.From)
	if err != nil {
		return err
	}
	if conn.Config == nil || conn.Config.VolumeMountConfig == nil {
		return fmt.Errorf("volume_mount connection from %q missing mount config", from)
	}
	cfg := conn.Config.VolumeMountConfig
	if cfg.SubPath != nil || cfg.ReadOnly != nil {
		return unsupported("volume mount %q: sub_path/read_only", from)
	}
	for _, vm := range res.Volumes {
		if vm.Name == from && vm.Path == cfg.MountPath {
			return nil
		}
	}
	return fmt.Errorf("volume_mount connection from %q has no matching volume mount on the resource", from)
}

func envTargetName(t openapi.ConnectionTarget) (string, error) {
	if t.Type != targetTypeEnv || t.Name == nil {
		return "", unsupported("connection target type %q", t.Type)
	}
	return *t.Name, nil
}

func refName(ref openapi.TopologyNodeRef) (string, error) {
	if ref.Name == nil || *ref.Name == "" {
		return "", fmt.Errorf("%s node ref has no name (id-only refs cannot be exported)", ref.Type)
	}
	return *ref.Name, nil
}

func setEnv(res *Resource, name, value string) {
	if res.Env == nil {
		res.Env = make(map[string]string)
	}
	res.Env[name] = value
}

func setEnvUnique(res *Resource, name, value string) error {
	if existing, ok := res.Env[name]; ok {
		return fmt.Errorf("env var %q set twice (%q and %q)", name, existing, value)
	}
	setEnv(res, name, value)
	return nil
}

func unsupported(format string, args ...any) error {
	return fmt.Errorf("not expressible in a stackfile: "+format, args...)
}
