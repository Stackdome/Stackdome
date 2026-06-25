package stackfile

import (
	"testing"

	openapi "github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
)

func TestToStack_BasicImageResource(t *testing.T) {
	sf := &Stackfile{
		Name: "my-stack",
		Resources: map[string]Resource{
			"web": {
				Image: "nginx:latest",
				Ports: []PortDef{
					{Name: "http", Port: 80, Public: true, Subdomain: "web"},
				},
			},
		},
	}

	stack := sf.ToStack()

	if stack.Name != "my-stack" {
		t.Errorf("expected name 'my-stack', got %q", stack.Name)
	}
	if len(stack.Spec.StackResources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(stack.Spec.StackResources))
	}

	res := stack.Spec.StackResources[0]
	if res.Name != "web" {
		t.Errorf("expected resource name 'web', got %q", res.Name)
	}
	if res.ImageSpec == nil || res.ImageSpec.Image != "nginx:latest" {
		t.Errorf("expected image 'nginx:latest', got %v", res.ImageSpec)
	}
	if res.BuildSpec != nil {
		t.Error("expected no build spec for image resource")
	}
	if len(res.Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(res.Ports))
	}
	if res.Ports[0].Name != "http" || res.Ports[0].Number != 80 || !res.Ports[0].ExposedToPublic {
		t.Errorf("unexpected port config: %+v", res.Ports[0])
	}
	if res.Ports[0].SubdomainPrefix == nil || *res.Ports[0].SubdomainPrefix != "web" {
		t.Errorf("expected subdomain 'web', got %v", res.Ports[0].SubdomainPrefix)
	}
}

func TestToStack_BuildFromSource(t *testing.T) {
	sf := &Stackfile{
		Name: "build-stack",
		Resources: map[string]Resource{
			"api": {
				Build: &BuildConfig{
					Repo:       "https://github.com/myorg/myapp.git",
					Branch:     "develop",
					Dockerfile: "docker/Dockerfile.prod",
					Context:    "./backend",
				},
			},
		},
	}

	stack := sf.ToStack()
	res := stack.Spec.StackResources[0]

	if res.ImageSpec != nil {
		t.Error("expected no image spec for build resource")
	}
	if res.BuildSpec == nil {
		t.Fatal("expected build spec")
	}
	if res.BuildSpec.SourceContext.GitRepo == nil {
		t.Fatal("expected git repo source context")
	}
	if res.BuildSpec.SourceContext.GitRepo.RepoUrl != "https://github.com/myorg/myapp.git" {
		t.Errorf("unexpected repo url: %s", res.BuildSpec.SourceContext.GitRepo.RepoUrl)
	}
	if res.BuildSpec.ContextPathWithinSource != "./backend" {
		t.Errorf("expected context './backend', got %q", res.BuildSpec.ContextPathWithinSource)
	}
	if res.BuildSpec.DockerfilePath != "docker/Dockerfile.prod" {
		t.Errorf("expected dockerfile 'docker/Dockerfile.prod', got %q", res.BuildSpec.DockerfilePath)
	}
	if res.BuildSpec.SourceRevision.GitRepoRevision == nil {
		t.Fatal("expected git repo revision")
	}
	if res.BuildSpec.SourceRevision.GitRepoRevision.Branch == nil || *res.BuildSpec.SourceRevision.GitRepoRevision.Branch.Name != "develop" {
		t.Errorf("expected branch 'develop', got %v", res.BuildSpec.SourceRevision.GitRepoRevision.Branch)
	}
}

func TestToStack_BuildDefaults(t *testing.T) {
	sf := &Stackfile{
		Name: "build-defaults",
		Resources: map[string]Resource{
			"app": {
				Build: &BuildConfig{
					Repo: "https://github.com/myorg/myapp.git",
				},
			},
		},
	}

	stack := sf.ToStack()
	res := stack.Spec.StackResources[0]

	if res.BuildSpec.ContextPathWithinSource != "." {
		t.Errorf("expected default context '.', got %q", res.BuildSpec.ContextPathWithinSource)
	}
	if res.BuildSpec.DockerfilePath != "Dockerfile" {
		t.Errorf("expected default dockerfile 'Dockerfile', got %q", res.BuildSpec.DockerfilePath)
	}
	if res.BuildSpec.SourceRevision.GitRepoRevision.Branch == nil || *res.BuildSpec.SourceRevision.GitRepoRevision.Branch.Name != "main" {
		t.Error("expected default branch 'main'")
	}
	if res.BuildSpec.ImageRepository.UseInternalRegistry == nil || !*res.BuildSpec.ImageRepository.UseInternalRegistry {
		t.Error("expected internal registry to be true by default")
	}
}

func TestToStack_BuildWithTag(t *testing.T) {
	sf := &Stackfile{
		Name: "tag-build",
		Resources: map[string]Resource{
			"app": {Build: &BuildConfig{Repo: "https://github.com/x/y.git", Tag: "v1.0.0"}},
		},
	}

	stack := sf.ToStack()
	rev := stack.Spec.StackResources[0].BuildSpec.SourceRevision.GitRepoRevision
	if rev.Tag == nil || *rev.Tag != "v1.0.0" {
		t.Errorf("expected tag 'v1.0.0', got %v", rev.Tag)
	}
	if rev.Branch != nil {
		t.Error("expected no branch when tag is set")
	}
}

func TestToStack_BuildWithCommit(t *testing.T) {
	sf := &Stackfile{
		Name: "commit-build",
		Resources: map[string]Resource{
			"app": {Build: &BuildConfig{Repo: "https://github.com/x/y.git", Commit: "abc123"}},
		},
	}

	stack := sf.ToStack()
	rev := stack.Spec.StackResources[0].BuildSpec.SourceRevision.GitRepoRevision
	if rev.Commit == nil || *rev.Commit != "abc123" {
		t.Errorf("expected commit 'abc123', got %v", rev.Commit)
	}
}

func TestToStack_EnvLiterals(t *testing.T) {
	sf := &Stackfile{
		Name: "env-test",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Env: map[string]string{
					"FOO": "bar",
					"NUM": "42",
				},
			},
		},
	}

	stack := sf.ToStack()
	res := stack.Spec.StackResources[0]

	if res.ExecutionConfig == nil {
		t.Fatal("expected execution config")
	}

	envMap := envVarsToMap(res.ExecutionConfig.EnvironmentVariables)
	if v, ok := envMap["FOO"]; !ok || *v.Value != "bar" {
		t.Errorf("expected FOO=bar, got %v", v)
	}
	if v, ok := envMap["NUM"]; !ok || *v.Value != "42" {
		t.Errorf("expected NUM=42, got %v", v)
	}
}

func TestToStack_SelfOutputEnv(t *testing.T) {
	sf := &Stackfile{
		Name: "self-output",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Ports: []PortDef{
					{Name: "http", Port: 8080, Public: true, Subdomain: "app"},
				},
				Env: map[string]string{
					"SITE_URL": "{{ self.public.http.url }}",
				},
			},
		},
	}

	stack := sf.ToStack()
	res := stack.Spec.StackResources[0]

	envMap := envVarsToMap(res.ExecutionConfig.EnvironmentVariables)
	v := envMap["SITE_URL"]
	if v.SelfOutput == nil || *v.SelfOutput != "public.http.url" {
		t.Errorf("expected self output 'public.http.url', got %v", v.SelfOutput)
	}
	if v.Value != nil {
		t.Error("self output should not have a literal value")
	}
}

func TestToStack_ResourceRefSimpleOutput(t *testing.T) {
	sf := &Stackfile{
		Name: "ref-test",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Env: map[string]string{
					"DB_HOST":       "{{ db.host }}",
					"LITERAL_VALUE": "hello",
				},
			},
			"db": {
				Image: "postgres:14",
			},
		},
	}

	stack := sf.ToStack()

	for _, res := range stack.Spec.StackResources {
		if res.Name == "app" && res.ExecutionConfig != nil {
			for _, ev := range res.ExecutionConfig.EnvironmentVariables {
				if ev.Name == "DB_HOST" {
					t.Error("DB_HOST should not be in execution config")
				}
			}
		}
	}

	found := false
	for _, conn := range stack.Spec.Connections {
		if conn.Kind == "env" && conn.From.Type == "stack_resource" && *conn.From.Name == "db" && *conn.To.Name == "app" {
			found = true
			if len(conn.Mappings) != 1 {
				t.Errorf("expected 1 mapping, got %d", len(conn.Mappings))
			}
			m := conn.Mappings[0]
			if *m.Target.Name != "DB_HOST" || *m.Value.Output != "host" {
				t.Errorf("expected DB_HOST->host, got %s->%s", *m.Target.Name, *m.Value.Output)
			}
		}
	}
	if !found {
		t.Error("expected connection from db to app")
	}
}

func TestToStack_ResourceRefTemplate(t *testing.T) {
	sf := &Stackfile{
		Name: "template-ref",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Env: map[string]string{
					"DB_URL": "postgres://user:pass@{{ db.host }}:5432/mydb",
				},
			},
			"db": {
				Image: "postgres:14",
			},
		},
	}

	stack := sf.ToStack()

	for _, res := range stack.Spec.StackResources {
		if res.Name == "app" && res.ExecutionConfig != nil {
			for _, ev := range res.ExecutionConfig.EnvironmentVariables {
				if ev.Name == "DB_URL" {
					t.Error("DB_URL should not be in execution config (handled by connection)")
				}
			}
		}
	}

	found := false
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "stack_resource" && *conn.From.Name == "db" {
			found = true
			m := conn.Mappings[0]
			if m.Value.Template == nil {
				t.Fatal("expected template value")
			}
			if *m.Value.Template != "postgres://user:pass@{{ host }}:5432/mydb" {
				t.Errorf("expected template with {{ host }}, got %q", *m.Value.Template)
			}
			if m.Value.Values == nil {
				t.Fatal("expected values map")
			}
			vals := *m.Value.Values
			if vals["host"].Output != "host" {
				t.Errorf("expected host output, got %q", vals["host"].Output)
			}
		}
	}
	if !found {
		t.Error("expected connection from db to app")
	}
}

func TestToStack_MultipleResourceRefs(t *testing.T) {
	sf := &Stackfile{
		Name: "multi-ref",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Env: map[string]string{
					"DB_HOST":   "{{ db.host }}",
					"REDIS_URL": "redis://{{ redis.host }}:6379",
				},
			},
			"db":    {Image: "postgres:14"},
			"redis": {Image: "redis:latest"},
		},
	}

	stack := sf.ToStack()

	sources := make(map[string]bool)
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "stack_resource" && conn.From.Name != nil {
			sources[*conn.From.Name] = true
		}
	}
	if !sources["db"] {
		t.Error("expected connection from db")
	}
	if !sources["redis"] {
		t.Error("expected connection from redis")
	}
}

func TestToStack_Secrets(t *testing.T) {
	sf := &Stackfile{
		Name: "secret-test",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Secrets: map[string]SecretMapping{
					"my-secret-id": {
						"API_KEY":    "api_key",
						"API_SECRET": "api_secret",
					},
				},
			},
		},
	}

	stack := sf.ToStack()

	found := false
	for _, conn := range stack.Spec.Connections {
		if conn.Kind == "env" && conn.From.Type == "secret" && *conn.From.Name == "my-secret-id" {
			found = true
			if *conn.To.Name != "app" {
				t.Errorf("expected target 'app', got %q", *conn.To.Name)
			}
			if len(conn.Mappings) != 2 {
				t.Errorf("expected 2 mappings, got %d", len(conn.Mappings))
			}
			mappingMap := make(map[string]string)
			for _, m := range conn.Mappings {
				mappingMap[*m.Target.Name] = *m.Value.Output
			}
			if mappingMap["API_KEY"] != "api_key" {
				t.Errorf("expected API_KEY->api_key, got %v", mappingMap)
			}
		}
	}
	if !found {
		t.Error("expected secret connection")
	}
}

func TestToStack_Volumes(t *testing.T) {
	sf := &Stackfile{
		Name: "vol-test",
		Resources: map[string]Resource{
			"db": {
				Image: "postgres:14",
				Volumes: []VolumeMountDef{
					{Name: "pg-data", Path: "/var/lib/postgresql/data"},
				},
				Stateful: true,
			},
		},
		Volumes: map[string]VolumeDef{
			"pg-data": {Size: "10Gi"},
		},
	}

	stack := sf.ToStack()

	// Check volume definition
	if len(stack.Spec.Volumes) != 1 {
		t.Fatalf("expected 1 volume, got %d", len(stack.Spec.Volumes))
	}
	vol := stack.Spec.Volumes[0]
	if vol.Name != "pg-data" {
		t.Errorf("expected volume name 'pg-data', got %q", vol.Name)
	}
	if vol.Spec.Size != "10Gi" {
		t.Errorf("expected size '10Gi', got %q", vol.Spec.Size)
	}
	if vol.Spec.AccessMode != "ReadWriteOnce" {
		t.Errorf("expected default access mode 'ReadWriteOnce', got %q", vol.Spec.AccessMode)
	}

	// Check volume mount on resource
	res := stack.Spec.StackResources[0]
	if len(res.VolumeMounts) != 1 {
		t.Fatalf("expected 1 volume mount, got %d", len(res.VolumeMounts))
	}
	if res.VolumeMounts[0].SourceVolumeName != "pg-data" {
		t.Errorf("expected source volume 'pg-data', got %q", res.VolumeMounts[0].SourceVolumeName)
	}
	if res.VolumeMounts[0].TargetPath != "/var/lib/postgresql/data" {
		t.Errorf("expected target path '/var/lib/postgresql/data', got %q", res.VolumeMounts[0].TargetPath)
	}

	// Check volume mount connection
	foundConn := false
	for _, conn := range stack.Spec.Connections {
		if conn.Kind == "volume_mount" && *conn.From.Name == "pg-data" && *conn.To.Name == "db" {
			foundConn = true
			if conn.Config == nil || conn.Config.VolumeMountConfig == nil {
				t.Fatal("expected volume mount config")
			}
			if conn.Config.VolumeMountConfig.MountPath != "/var/lib/postgresql/data" {
				t.Errorf("expected mount path '/var/lib/postgresql/data', got %q", conn.Config.VolumeMountConfig.MountPath)
			}
		}
	}
	if !foundConn {
		t.Error("expected volume_mount connection")
	}

	// Check stateful flag
	if res.Stateful == nil || !*res.Stateful {
		t.Error("expected stateful=true")
	}
}

func TestToStack_CustomAccessMode(t *testing.T) {
	sf := &Stackfile{
		Name: "access-mode",
		Resources: map[string]Resource{
			"app": {Image: "nginx:latest"},
		},
		Volumes: map[string]VolumeDef{
			"shared": {Size: "5Gi", AccessMode: "ReadWriteMany"},
		},
	}

	stack := sf.ToStack()
	if stack.Spec.Volumes[0].Spec.AccessMode != "ReadWriteMany" {
		t.Errorf("expected ReadWriteMany, got %q", stack.Spec.Volumes[0].Spec.AccessMode)
	}
}

func TestToStack_DependsOn(t *testing.T) {
	sf := &Stackfile{
		Name: "deps-test",
		Resources: map[string]Resource{
			"app": {
				Image:     "myapp:latest",
				DependsOn: []string{"db", "redis"},
			},
			"db":    {Image: "postgres:14"},
			"redis": {Image: "redis:latest"},
		},
	}

	stack := sf.ToStack()
	for _, res := range stack.Spec.StackResources {
		if res.Name == "app" {
			if len(res.DependsOn) != 2 {
				t.Errorf("expected 2 dependencies, got %d", len(res.DependsOn))
			}
		}
	}
}

func TestToStack_PortProtocol(t *testing.T) {
	sf := &Stackfile{
		Name: "proto-test",
		Resources: map[string]Resource{
			"db": {
				Image: "postgres:14",
				Ports: []PortDef{
					{Name: "postgres", Port: 5432, Protocol: "TCP"},
				},
			},
		},
	}

	stack := sf.ToStack()
	port := stack.Spec.StackResources[0].Ports[0]
	if port.Protocol == nil || *port.Protocol != "TCP" {
		t.Errorf("expected protocol TCP, got %v", port.Protocol)
	}
	if port.ExposedToPublic {
		t.Error("expected not public")
	}
}

func TestToStack_NoEnv(t *testing.T) {
	sf := &Stackfile{
		Name: "no-env",
		Resources: map[string]Resource{
			"app": {Image: "nginx:latest"},
		},
	}

	stack := sf.ToStack()
	if stack.Spec.StackResources[0].ExecutionConfig != nil {
		t.Error("expected nil execution config when no env")
	}
}

func TestToStack_FullInfisicalExample(t *testing.T) {
	sf := &Stackfile{
		Name: "infisical",
		Resources: map[string]Resource{
			"infisical": {
				Image: "infisical/infisical:latest",
				Ports: []PortDef{
					{Name: "http", Port: 8080, Protocol: "HTTP", Public: true, Subdomain: "infisical"},
				},
				Env: map[string]string{
					"SITE_URL":          "{{ self.public.http.url }}",
					"DB_CONNECTION_URI": "postgres://infisical:infisical@{{ db.host }}:5432/infisical",
					"REDIS_URL":         "redis://{{ redis.host }}:6379",
					"ENCRYPTION_KEY":    "6c1fe4e407b8911c104518103505b218",
					"POSTGRES_USER":     "infisical",
					"POSTGRES_PASSWORD": "infisical",
					"POSTGRES_DB":       "infisical",
				},
				DependsOn: []string{"db", "redis"},
			},
			"db": {
				Image:    "postgres:14-alpine",
				Ports:    []PortDef{{Name: "postgres", Port: 5432, Protocol: "TCP"}},
				Env:      map[string]string{"POSTGRES_USER": "infisical", "POSTGRES_PASSWORD": "infisical", "POSTGRES_DB": "infisical"},
				Volumes:  []VolumeMountDef{{Name: "pg-data", Path: "/var/lib/postgresql/data"}},
				Stateful: true,
			},
			"redis": {
				Image:   "redis:latest",
				Ports:   []PortDef{{Name: "redis", Port: 6379, Protocol: "TCP"}},
				Env:     map[string]string{"ALLOW_EMPTY_PASSWORD": "yes"},
				Volumes: []VolumeMountDef{{Name: "redis-data", Path: "/data"}},
			},
		},
		Volumes: map[string]VolumeDef{
			"pg-data":    {Size: "5Gi"},
			"redis-data": {Size: "1Gi"},
		},
	}

	stack := sf.ToStack()

	if stack.Name != "infisical" {
		t.Errorf("expected name 'infisical', got %q", stack.Name)
	}
	if len(stack.Spec.StackResources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(stack.Spec.StackResources))
	}
	if len(stack.Spec.Volumes) != 2 {
		t.Errorf("expected 2 volumes, got %d", len(stack.Spec.Volumes))
	}

	envConns := 0
	volConns := 0
	for _, conn := range stack.Spec.Connections {
		switch conn.Kind {
		case "env":
			envConns++
			// Verify template-based connections
			if conn.From.Type == "stack_resource" {
				for _, m := range conn.Mappings {
					if m.Value.Template == nil {
						t.Errorf("expected template for env connection from %s", *conn.From.Name)
					}
				}
			}
		case "volume_mount":
			volConns++
		}
	}
	if envConns != 2 {
		t.Errorf("expected 2 env connections (db + redis), got %d", envConns)
	}
	if volConns != 2 {
		t.Errorf("expected 2 volume_mount connections, got %d", volConns)
	}

	for _, res := range stack.Spec.StackResources {
		if res.Name == "infisical" && res.ExecutionConfig != nil {
			envMap := envVarsToMap(res.ExecutionConfig.EnvironmentVariables)
			if v, ok := envMap["SITE_URL"]; !ok || v.SelfOutput == nil {
				t.Error("expected SITE_URL as self output")
			}
			if v, ok := envMap["ENCRYPTION_KEY"]; !ok || *v.Value != "6c1fe4e407b8911c104518103505b218" {
				t.Error("expected ENCRYPTION_KEY as literal")
			}
			if _, ok := envMap["DB_CONNECTION_URI"]; ok {
				t.Error("DB_CONNECTION_URI should not be in execution config (handled by connection)")
			}
			if _, ok := envMap["REDIS_URL"]; ok {
				t.Error("REDIS_URL should not be in execution config (handled by connection)")
			}
		}
	}
}

func TestToStack_MultipleSecrets(t *testing.T) {
	sf := &Stackfile{
		Name: "multi-secret",
		Resources: map[string]Resource{
			"app": {
				Image: "myapp:latest",
				Secrets: map[string]SecretMapping{
					"db-creds": {
						"DB_USER": "username",
						"DB_PASS": "password",
					},
					"api-keys": {
						"STRIPE_KEY":   "stripe_key",
						"SENDGRID_KEY": "sendgrid_key",
					},
				},
			},
		},
	}

	stack := sf.ToStack()

	secretConns := 0
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "secret" {
			secretConns++
		}
	}
	if secretConns != 2 {
		t.Errorf("expected 2 secret connections, got %d", secretConns)
	}

	// Verify each secret has correct mappings
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type != "secret" {
			continue
		}
		switch *conn.From.Name {
		case "db-creds":
			if len(conn.Mappings) != 2 {
				t.Errorf("db-creds: expected 2 mappings, got %d", len(conn.Mappings))
			}
		case "api-keys":
			if len(conn.Mappings) != 2 {
				t.Errorf("api-keys: expected 2 mappings, got %d", len(conn.Mappings))
			}
		default:
			t.Errorf("unexpected secret name: %s", *conn.From.Name)
		}
		if conn.To.Type != "stack_resource" || *conn.To.Name != "app" {
			t.Errorf("expected target app, got %s/%v", conn.To.Type, conn.To.Name)
		}
	}
}

func TestToStack_SecretOnMultipleResources(t *testing.T) {
	sf := &Stackfile{
		Name: "shared-secret",
		Resources: map[string]Resource{
			"api": {
				Image: "api:latest",
				Secrets: map[string]SecretMapping{
					"shared-creds": {"DB_URL": "url"},
				},
			},
			"worker": {
				Image: "worker:latest",
				Secrets: map[string]SecretMapping{
					"shared-creds": {"DB_URL": "url"},
				},
			},
		},
	}

	stack := sf.ToStack()

	targets := make(map[string]bool)
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "secret" && *conn.From.Name == "shared-creds" {
			targets[*conn.To.Name] = true
		}
	}
	if !targets["api"] || !targets["worker"] {
		t.Errorf("expected connections to both api and worker, got %v", targets)
	}
}

func TestToStack_MultipleVolumeMountsOnOneResource(t *testing.T) {
	sf := &Stackfile{
		Name: "multi-mount",
		Resources: map[string]Resource{
			"db": {
				Image: "postgres:16",
				Volumes: []VolumeMountDef{
					{Name: "pg-data", Path: "/var/lib/postgresql/data"},
					{Name: "pg-wal", Path: "/var/lib/postgresql/wal"},
				},
				Stateful: true,
			},
		},
		Volumes: map[string]VolumeDef{
			"pg-data": {Size: "20Gi"},
			"pg-wal":  {Size: "5Gi"},
		},
	}

	stack := sf.ToStack()

	res := stack.Spec.StackResources[0]
	if len(res.VolumeMounts) != 2 {
		t.Fatalf("expected 2 volume mounts, got %d", len(res.VolumeMounts))
	}

	mountPaths := make(map[string]string)
	for _, vm := range res.VolumeMounts {
		mountPaths[vm.SourceVolumeName] = vm.TargetPath
	}
	if mountPaths["pg-data"] != "/var/lib/postgresql/data" {
		t.Errorf("expected pg-data -> /var/lib/postgresql/data, got %q", mountPaths["pg-data"])
	}
	if mountPaths["pg-wal"] != "/var/lib/postgresql/wal" {
		t.Errorf("expected pg-wal -> /var/lib/postgresql/wal, got %q", mountPaths["pg-wal"])
	}

	volConns := 0
	for _, conn := range stack.Spec.Connections {
		if conn.Kind == "volume_mount" {
			volConns++
			if conn.Config == nil || conn.Config.VolumeMountConfig == nil {
				t.Error("expected volume mount config")
			}
		}
	}
	if volConns != 2 {
		t.Errorf("expected 2 volume_mount connections, got %d", volConns)
	}
}

func TestToStack_SharedVolumeBetweenResources(t *testing.T) {
	sf := &Stackfile{
		Name: "shared-vol",
		Resources: map[string]Resource{
			"writer": {
				Image:   "writer:latest",
				Volumes: []VolumeMountDef{{Name: "shared-data", Path: "/data/output"}},
			},
			"reader": {
				Image:   "reader:latest",
				Volumes: []VolumeMountDef{{Name: "shared-data", Path: "/data/input"}},
			},
		},
		Volumes: map[string]VolumeDef{
			"shared-data": {Size: "10Gi", AccessMode: "ReadWriteMany"},
		},
	}

	stack := sf.ToStack()

	if len(stack.Spec.Volumes) != 1 {
		t.Fatalf("expected 1 volume definition, got %d", len(stack.Spec.Volumes))
	}

	// Should create 2 volume_mount connections (one per resource)
	volConns := make(map[string]string)
	for _, conn := range stack.Spec.Connections {
		if conn.Kind == "volume_mount" && *conn.From.Name == "shared-data" {
			volConns[*conn.To.Name] = conn.Config.VolumeMountConfig.MountPath
		}
	}
	if volConns["writer"] != "/data/output" {
		t.Errorf("expected writer mount at /data/output, got %q", volConns["writer"])
	}
	if volConns["reader"] != "/data/input" {
		t.Errorf("expected reader mount at /data/input, got %q", volConns["reader"])
	}
}

func TestToStack_VolumeMountConnectionHasCorrectNodeTypes(t *testing.T) {
	sf := &Stackfile{
		Name: "vol-types",
		Resources: map[string]Resource{
			"app": {
				Image:   "app:latest",
				Volumes: []VolumeMountDef{{Name: "cache", Path: "/cache"}},
			},
		},
		Volumes: map[string]VolumeDef{
			"cache": {Size: "1Gi"},
		},
	}

	stack := sf.ToStack()

	for _, conn := range stack.Spec.Connections {
		if conn.Kind != "volume_mount" {
			continue
		}
		if conn.From.Type != "volume" {
			t.Errorf("expected from.type 'volume', got %q", conn.From.Type)
		}
		if conn.From.Name == nil || *conn.From.Name != "cache" {
			t.Errorf("expected from.name 'cache', got %v", conn.From.Name)
		}
		if conn.To.Type != "stack_resource" {
			t.Errorf("expected to.type 'stack_resource', got %q", conn.To.Type)
		}
		if conn.To.Name == nil || *conn.To.Name != "app" {
			t.Errorf("expected to.name 'app', got %v", conn.To.Name)
		}
	}
}

func envVarsToMap(vars []openapi.EnvVar) map[string]openapi.EnvVar {
	m := make(map[string]openapi.EnvVar)
	for _, v := range vars {
		m[v.Name] = v
	}
	return m
}
