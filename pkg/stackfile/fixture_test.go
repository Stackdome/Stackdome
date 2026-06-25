package stackfile

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testdataPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func loadFixture(t *testing.T, name string) *Stackfile {
	t.Helper()
	data, err := os.ReadFile(testdataPath(name))
	if err != nil {
		t.Fatalf("failed to read testdata file %s: %v", name, err)
	}
	sf, err := Load(data)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	return sf
}

func TestFixture_BasicImage(t *testing.T) {
	sf := loadFixture(t, "basic_image.yaml")

	stack := sf.ToStack()

	if stack.Name != "basic-stack" {
		t.Errorf("expected name 'basic-stack', got %q", stack.Name)
	}
	if len(stack.Spec.StackResources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(stack.Spec.StackResources))
	}
	res := stack.Spec.StackResources[0]
	if res.ImageSpec == nil || res.ImageSpec.Image != "nginx:latest" {
		t.Error("expected image nginx:latest")
	}
	if len(res.Ports) != 1 || !res.Ports[0].ExposedToPublic {
		t.Error("expected public port")
	}
}

func TestFixture_BuildFromSource(t *testing.T) {
	sf := loadFixture(t, "build_from_source.yaml")

	stack := sf.ToStack()
	res := stack.Spec.StackResources[0]

	if res.BuildSpec == nil {
		t.Fatal("expected build spec")
	}
	if res.BuildSpec.SourceContext.GitRepo.RepoUrl != "https://github.com/myorg/myapp.git" {
		t.Error("wrong repo url")
	}
	if res.BuildSpec.ContextPathWithinSource != "./backend" {
		t.Errorf("expected context './backend', got %q", res.BuildSpec.ContextPathWithinSource)
	}
	if *res.BuildSpec.SourceRevision.GitRepoRevision.Branch.Name != "develop" {
		t.Error("expected branch develop")
	}
}

func TestFixture_Infisical(t *testing.T) {
	sf := loadFixture(t, "infisical.yaml")

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
		case "volume_mount":
			volConns++
		}
	}
	if envConns != 2 {
		t.Errorf("expected 2 env connections, got %d", envConns)
	}
	if volConns != 2 {
		t.Errorf("expected 2 volume mount connections, got %d", volConns)
	}

	for _, res := range stack.Spec.StackResources {
		if res.Name != "infisical" || res.ExecutionConfig == nil {
			continue
		}
		envMap := envVarsToMap(res.ExecutionConfig.EnvironmentVariables)
		if v, ok := envMap["SITE_URL"]; !ok || v.SelfOutput == nil {
			t.Error("SITE_URL should be self output")
		}
		if v, ok := envMap["ENCRYPTION_KEY"]; !ok || *v.Value != "6c1fe4e407b8911c104518103505b218" {
			t.Error("ENCRYPTION_KEY should be literal")
		}
		if _, ok := envMap["DB_CONNECTION_URI"]; ok {
			t.Error("DB_CONNECTION_URI should be a connection, not in execution config")
		}
	}
}

func TestFixture_WithSecrets(t *testing.T) {
	sf := loadFixture(t, "with_secrets.yaml")

	stack := sf.ToStack()

	secretConns := 0
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "secret" {
			secretConns++
			if *conn.To.Name != "api" {
				t.Errorf("expected target 'api', got %q", *conn.To.Name)
			}
		}
	}
	if secretConns != 2 {
		t.Errorf("expected 2 secret connections, got %d", secretConns)
	}
}

func TestFixture_WithAddon(t *testing.T) {
	sf := loadFixture(t, "with_addon.yaml")

	stack := sf.ToStack()

	found := false
	for _, conn := range stack.Spec.Connections {
		if conn.From.Type == "addon/postgres" && *conn.From.Name == "pg-addon-id" {
			found = true
			if conn.Config == nil || conn.Config.PostgresEnvConfig == nil {
				t.Fatal("expected postgres config")
			}
			if *conn.Config.PostgresEnvConfig.Database != "appdb" {
				t.Errorf("expected database 'appdb', got %q", *conn.Config.PostgresEnvConfig.Database)
			}
			if len(conn.Mappings) != 2 {
				t.Errorf("expected 2 mappings, got %d", len(conn.Mappings))
			}

			for _, m := range conn.Mappings {
				switch *m.Target.Name {
				case "DATABASE_URL":
					if m.Value.Template == nil {
						t.Error("DATABASE_URL should use template")
					}
				case "DB_HOST":
					if m.Value.Output == nil || *m.Value.Output != "host" {
						t.Error("DB_HOST should be direct output 'host'")
					}
				}
			}
		}
	}
	if !found {
		t.Error("expected addon/postgres connection")
	}
}

func TestFixture_WithAddonSuperuser(t *testing.T) {
	sf := loadFixture(t, "with_addon_superuser.yaml")

	stack := sf.ToStack()

	for _, conn := range stack.Spec.Connections {
		if conn.From.Type != "addon/postgres" {
			continue
		}
		pgCfg := conn.Config.PostgresEnvConfig
		if pgCfg.Superuser == nil || !*pgCfg.Superuser {
			t.Error("expected superuser=true")
		}
		if pgCfg.Database == nil || *pgCfg.Database != "appdb" {
			t.Errorf("expected database 'appdb', got %v", pgCfg.Database)
		}
	}
}

func TestFixture_KitchenSink(t *testing.T) {
	sf := loadFixture(t, "kitchen_sink.yaml")

	stack := sf.ToStack()

	if len(stack.Spec.StackResources) != 3 {
		t.Errorf("expected 3 resources, got %d", len(stack.Spec.StackResources))
	}
	if len(stack.Spec.Volumes) != 1 {
		t.Errorf("expected 1 volume, got %d", len(stack.Spec.Volumes))
	}

	counts := map[string]int{}
	for _, conn := range stack.Spec.Connections {
		key := conn.Kind + ":" + conn.From.Type
		counts[key]++
	}

	if counts["env:stack_resource"] != 2 {
		t.Errorf("expected 2 resource env connections (redis->api, redis->worker), got %d", counts["env:stack_resource"])
	}
	if counts["env:secret"] != 3 {
		t.Errorf("expected 3 secret connections, got %d", counts["env:secret"])
	}
	if counts["env:addon/postgres"] != 2 {
		t.Errorf("expected 2 addon connections, got %d", counts["env:addon/postgres"])
	}
	if counts["volume_mount:volume"] != 1 {
		t.Errorf("expected 1 volume mount connection, got %d", counts["volume_mount:volume"])
	}
}
