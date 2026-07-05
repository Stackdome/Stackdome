package stackfile

import (
	"testing"

	openapi "github.com/Stackdome/stackdome/pkg/api/openapi"
)

// --- buildEnvRefConnections ---

func TestBuildEnvRefConnections_ExactRef(t *testing.T) {
	env := map[string]string{
		"DB_HOST": "{{ db.host }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	c := conns[0]
	assertConnection(t, c, "env", "stack_resource", "db", "stack_resource", "app")

	if len(c.Mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(c.Mappings))
	}
	assertDirectMapping(t, c.Mappings[0], "DB_HOST", "host")
}

func TestBuildEnvRefConnections_TemplateRef(t *testing.T) {
	env := map[string]string{
		"REDIS_URL": "redis://{{ redis.host }}:6379",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	c := conns[0]
	assertConnection(t, c, "env", "stack_resource", "redis", "stack_resource", "app")

	m := c.Mappings[0]
	if m.Value.Template == nil {
		t.Fatal("expected template value")
	}
	if *m.Value.Template != "redis://{{ host }}:6379" {
		t.Errorf("expected template 'redis://{{ host }}:6379', got %q", *m.Value.Template)
	}
	assertTemplateVar(t, m, "host", "host")
}

func TestBuildEnvRefConnections_TemplateWithDottedOutput(t *testing.T) {
	env := map[string]string{
		"CALLBACK": "https://{{ api.public.http.url }}/callback",
	}
	conns := buildEnvRefConnections("worker", env)

	m := conns[0].Mappings[0]
	if *m.Value.Template != "https://{{ public_http_url }}/callback" {
		t.Errorf("expected dotted output converted to underscore var, got %q", *m.Value.Template)
	}
	assertTemplateVar(t, m, "public_http_url", "public.http.url")
}

func TestBuildEnvRefConnections_MultipleRefsFromSameSource(t *testing.T) {
	env := map[string]string{
		"DB_HOST": "{{ db.host }}",
		"DB_PORT": "{{ db.port.postgres }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection (grouped), got %d", len(conns))
	}
	if len(conns[0].Mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(conns[0].Mappings))
	}
}

func TestBuildEnvRefConnections_DifferentSourcesCreateSeparateConnections(t *testing.T) {
	env := map[string]string{
		"DB_HOST":    "{{ db.host }}",
		"CACHE_HOST": "{{ redis.host }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}

	sources := map[string]bool{}
	for _, c := range conns {
		sources[*c.From.Name] = true
	}
	if !sources["db"] || !sources["redis"] {
		t.Errorf("expected connections from db and redis, got %v", sources)
	}
}

func TestBuildEnvRefConnections_SkipsSelfRefs(t *testing.T) {
	env := map[string]string{
		"SITE_URL": "{{ self.public.http.url }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 0 {
		t.Errorf("expected 0 connections for self refs, got %d", len(conns))
	}
}

func TestBuildEnvRefConnections_SkipsLiterals(t *testing.T) {
	env := map[string]string{
		"LOG_LEVEL": "info",
		"PORT":      "3000",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 0 {
		t.Errorf("expected 0 connections for literals, got %d", len(conns))
	}
}

func TestBuildEnvRefConnections_MixedLiteralsAndRefs(t *testing.T) {
	env := map[string]string{
		"DB_HOST":   "{{ db.host }}",
		"LOG_LEVEL": "info",
		"SITE_URL":  "{{ self.public.http.url }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection (only db ref), got %d", len(conns))
	}
	if *conns[0].From.Name != "db" {
		t.Errorf("expected connection from db, got %s", *conns[0].From.Name)
	}
}

func TestBuildEnvRefConnections_TemplateMultipleRefsFromSameSource(t *testing.T) {
	env := map[string]string{
		"DB_URL": "postgres://{{ db.host }}:{{ db.port.postgres }}",
	}
	conns := buildEnvRefConnections("app", env)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}

	m := conns[0].Mappings[0]
	if m.Value.Template == nil {
		t.Fatal("expected template")
	}
	if *m.Value.Template != "postgres://{{ host }}:{{ port_postgres }}" {
		t.Errorf("unexpected template: %q", *m.Value.Template)
	}
	assertTemplateVar(t, m, "host", "host")
	assertTemplateVar(t, m, "port_postgres", "port.postgres")
}

func TestBuildEnvRefConnections_EmptyEnv(t *testing.T) {
	conns := buildEnvRefConnections("app", nil)
	if len(conns) != 0 {
		t.Errorf("expected 0, got %d", len(conns))
	}
	conns = buildEnvRefConnections("app", map[string]string{})
	if len(conns) != 0 {
		t.Errorf("expected 0, got %d", len(conns))
	}
}

// --- buildSecretConnections ---

func TestBuildSecretConnections_SingleSecret(t *testing.T) {
	secrets := map[string]SecretMapping{
		"api-keys": {
			"API_KEY":    "api_key",
			"API_SECRET": "api_secret",
		},
	}
	conns := buildSecretConnections("app", secrets)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	c := conns[0]
	assertConnection(t, c, "env", "secret", "api-keys", "stack_resource", "app")

	if len(c.Mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(c.Mappings))
	}
	mappings := mappingMap(c.Mappings)
	if mappings["API_KEY"] != "api_key" {
		t.Errorf("expected API_KEY->api_key, got %s", mappings["API_KEY"])
	}
	if mappings["API_SECRET"] != "api_secret" {
		t.Errorf("expected API_SECRET->api_secret, got %s", mappings["API_SECRET"])
	}
}

func TestBuildSecretConnections_MultipleSecrets(t *testing.T) {
	secrets := map[string]SecretMapping{
		"db-creds":   {"DB_PASS": "password"},
		"smtp-creds": {"SMTP_PASS": "password"},
	}
	conns := buildSecretConnections("app", secrets)

	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}

	names := map[string]bool{}
	for _, c := range conns {
		names[*c.From.Name] = true
		if c.From.Type != "secret" {
			t.Errorf("expected from.type 'secret', got %q", c.From.Type)
		}
		if *c.To.Name != "app" {
			t.Errorf("expected to.name 'app', got %q", *c.To.Name)
		}
	}
	if !names["db-creds"] || !names["smtp-creds"] {
		t.Errorf("expected both secrets, got %v", names)
	}
}

func TestBuildSecretConnections_UsesNameNotId(t *testing.T) {
	secrets := map[string]SecretMapping{
		"my-secret": {"KEY": "value"},
	}
	conns := buildSecretConnections("app", secrets)

	c := conns[0]
	if c.From.Name == nil || *c.From.Name != "my-secret" {
		t.Errorf("expected from.name 'my-secret', got %v", c.From.Name)
	}
	if c.From.Id != nil {
		t.Error("expected from.id to be nil (resolved at deploy time)")
	}
}

func TestBuildSecretConnections_Empty(t *testing.T) {
	conns := buildSecretConnections("app", nil)
	if len(conns) != 0 {
		t.Errorf("expected 0, got %d", len(conns))
	}
}

// --- buildAddonConnections ---

func TestBuildAddonConnections_DirectOutput(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"main-db": {
			Type:     "postgres",
			Env:      map[string]string{"DB_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{Database: "mydb"},
		},
	}
	conns := buildAddonConnections("app", addons)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	c := conns[0]
	assertConnection(t, c, "env", "addon/postgres", "main-db", "stack_resource", "app")
	assertDirectMapping(t, c.Mappings[0], "DB_HOST", "host")
}

func TestBuildAddonConnections_Template(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"main-db": {
			Type: "postgres",
			Env: map[string]string{
				"DATABASE_URL": "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}",
			},
			Postgres: &PostgresAddonConfig{Database: "mydb"},
		},
	}
	conns := buildAddonConnections("app", addons)

	m := conns[0].Mappings[0]
	if m.Value.Template == nil {
		t.Fatal("expected template")
	}
	if m.Value.Output != nil {
		t.Error("template mapping should not have direct output")
	}

	vals := *m.Value.Values
	for _, key := range []string{"username", "password", "host", "port", "database"} {
		if _, ok := vals[key]; !ok {
			t.Errorf("missing template var %q", key)
		}
	}
}

func TestBuildAddonConnections_MixedDirectAndTemplate(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"pg": {
			Type: "postgres",
			Env: map[string]string{
				"DB_HOST":      "{{ host }}",
				"DATABASE_URL": "postgres://{{ username }}:{{ password }}@{{ host }}:{{ port }}/{{ database }}",
			},
			Postgres: &PostgresAddonConfig{},
		},
	}
	conns := buildAddonConnections("app", addons)

	if len(conns[0].Mappings) != 2 {
		t.Fatalf("expected 2 mappings, got %d", len(conns[0].Mappings))
	}

	for _, m := range conns[0].Mappings {
		switch *m.Target.Name {
		case "DB_HOST":
			if m.Value.Output == nil || *m.Value.Output != "host" {
				t.Error("DB_HOST should be direct output 'host'")
			}
		case "DATABASE_URL":
			if m.Value.Template == nil {
				t.Error("DATABASE_URL should be template")
			}
		}
	}
}

func TestBuildAddonConnections_WithDatabaseConfig(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"pg": {
			Type:     "postgres",
			Env:      map[string]string{"DB_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{Database: "appdb"},
		},
	}
	conns := buildAddonConnections("app", addons)

	c := conns[0]
	if c.Config == nil || c.Config.PostgresEnvConfig == nil {
		t.Fatal("expected postgres config")
	}
	if *c.Config.PostgresEnvConfig.Database != "appdb" {
		t.Errorf("expected database 'appdb', got %q", *c.Config.PostgresEnvConfig.Database)
	}
}

func TestBuildAddonConnections_WithSuperuser(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"pg": {
			Type:     "postgres",
			Env:      map[string]string{"DB_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{Database: "appdb", Superuser: true},
		},
	}
	conns := buildAddonConnections("app", addons)

	pgCfg := conns[0].Config.PostgresEnvConfig
	if pgCfg.Superuser == nil || !*pgCfg.Superuser {
		t.Error("expected superuser=true")
	}
}

func TestBuildAddonConnections_NoConfig(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"pg": {
			Type:     "postgres",
			Env:      map[string]string{"DB_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{},
		},
	}
	conns := buildAddonConnections("app", addons)

	if conns[0].Config != nil {
		t.Error("expected nil config when no database or superuser set")
	}
}

func TestBuildAddonConnections_UsesNameNotId(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"my-postgres": {
			Type:     "postgres",
			Env:      map[string]string{"DB_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{},
		},
	}
	conns := buildAddonConnections("app", addons)

	c := conns[0]
	if c.From.Name == nil || *c.From.Name != "my-postgres" {
		t.Errorf("expected from.name 'my-postgres', got %v", c.From.Name)
	}
	if c.From.Id != nil {
		t.Error("expected from.id to be nil")
	}
}

func TestBuildAddonConnections_MultipleAddons(t *testing.T) {
	addons := map[string]AddonConnectionConfig{
		"primary": {
			Type:     "postgres",
			Env:      map[string]string{"PRIMARY_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{Database: "primary"},
		},
		"analytics": {
			Type:     "postgres",
			Env:      map[string]string{"ANALYTICS_HOST": "{{ host }}"},
			Postgres: &PostgresAddonConfig{Database: "analytics"},
		},
	}
	conns := buildAddonConnections("app", addons)

	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}

	dbs := map[string]string{}
	for _, c := range conns {
		dbs[*c.From.Name] = *c.Config.PostgresEnvConfig.Database
	}
	if dbs["primary"] != "primary" || dbs["analytics"] != "analytics" {
		t.Errorf("unexpected addon configs: %v", dbs)
	}
}

func TestBuildAddonConnections_Empty(t *testing.T) {
	conns := buildAddonConnections("app", nil)
	if len(conns) != 0 {
		t.Errorf("expected 0, got %d", len(conns))
	}
}

// --- buildVolumeMountConnections ---

func TestBuildVolumeMountConnections_Single(t *testing.T) {
	mounts := []VolumeMountDef{
		{Name: "pg-data", Path: "/var/lib/postgresql/data"},
	}
	conns := buildVolumeMountConnections("db", mounts)

	if len(conns) != 1 {
		t.Fatalf("expected 1 connection, got %d", len(conns))
	}
	c := conns[0]
	assertConnection(t, c, "volume_mount", "volume", "pg-data", "stack_resource", "db")

	if c.Config == nil || c.Config.VolumeMountConfig == nil {
		t.Fatal("expected volume mount config")
	}
	if c.Config.VolumeMountConfig.MountPath != "/var/lib/postgresql/data" {
		t.Errorf("expected mount path '/var/lib/postgresql/data', got %q", c.Config.VolumeMountConfig.MountPath)
	}
}

func TestBuildVolumeMountConnections_Multiple(t *testing.T) {
	mounts := []VolumeMountDef{
		{Name: "data", Path: "/data"},
		{Name: "logs", Path: "/var/log"},
		{Name: "config", Path: "/etc/app"},
	}
	conns := buildVolumeMountConnections("app", mounts)

	if len(conns) != 3 {
		t.Fatalf("expected 3 connections, got %d", len(conns))
	}

	paths := map[string]string{}
	for _, c := range conns {
		paths[*c.From.Name] = c.Config.VolumeMountConfig.MountPath
	}
	if paths["data"] != "/data" {
		t.Errorf("expected data->/data, got %q", paths["data"])
	}
	if paths["logs"] != "/var/log" {
		t.Errorf("expected logs->/var/log, got %q", paths["logs"])
	}
	if paths["config"] != "/etc/app" {
		t.Errorf("expected config->/etc/app, got %q", paths["config"])
	}
}

func TestBuildVolumeMountConnections_NodeTypes(t *testing.T) {
	mounts := []VolumeMountDef{{Name: "vol", Path: "/mnt"}}
	conns := buildVolumeMountConnections("app", mounts)

	c := conns[0]
	if c.From.Type != "volume" {
		t.Errorf("expected from.type 'volume', got %q", c.From.Type)
	}
	if c.To.Type != "stack_resource" {
		t.Errorf("expected to.type 'stack_resource', got %q", c.To.Type)
	}
	if c.Kind != "volume_mount" {
		t.Errorf("expected kind 'volume_mount', got %q", c.Kind)
	}
	if c.Mappings != nil {
		t.Error("volume_mount connections should have no mappings")
	}
}

func TestBuildVolumeMountConnections_Empty(t *testing.T) {
	conns := buildVolumeMountConnections("app", nil)
	if len(conns) != 0 {
		t.Errorf("expected 0, got %d", len(conns))
	}
}

// --- helpers ---

func assertConnection(t *testing.T, c interface{}, kind, fromType, fromName, toType, toName string) {
	t.Helper()
	conn, ok := c.(openapi.StackConnection)
	if !ok {
		t.Fatal("expected StackConnection")
	}
	if conn.Kind != kind {
		t.Errorf("expected kind %q, got %q", kind, conn.Kind)
	}
	if conn.From.Type != fromType {
		t.Errorf("expected from.type %q, got %q", fromType, conn.From.Type)
	}
	if conn.From.Name == nil || *conn.From.Name != fromName {
		t.Errorf("expected from.name %q, got %v", fromName, conn.From.Name)
	}
	if conn.To.Type != toType {
		t.Errorf("expected to.type %q, got %q", toType, conn.To.Type)
	}
	if conn.To.Name == nil || *conn.To.Name != toName {
		t.Errorf("expected to.name %q, got %v", toName, conn.To.Name)
	}
}

func assertDirectMapping(t *testing.T, m openapi.ConnectionMapping, envName, output string) {
	t.Helper()
	if m.Target.Type != "env" {
		t.Errorf("expected target.type 'env', got %q", m.Target.Type)
	}
	if *m.Target.Name != envName {
		t.Errorf("expected target.name %q, got %q", envName, *m.Target.Name)
	}
	if m.Value.Output == nil || *m.Value.Output != output {
		t.Errorf("expected direct output %q, got %v", output, m.Value.Output)
	}
	if m.Value.Template != nil {
		t.Error("direct mapping should not have template")
	}
}

func assertTemplateVar(t *testing.T, m openapi.ConnectionMapping, varName, output string) {
	t.Helper()
	if m.Value.Values == nil {
		t.Fatalf("expected values map for template var %q", varName)
	}
	vals := *m.Value.Values
	ref, ok := vals[varName]
	if !ok {
		t.Errorf("missing template var %q", varName)
		return
	}
	if ref.Output != output {
		t.Errorf("expected var %q -> output %q, got %q", varName, output, ref.Output)
	}
}

func mappingMap(mappings []openapi.ConnectionMapping) map[string]string {
	m := make(map[string]string)
	for _, mapping := range mappings {
		if mapping.Value.Output != nil {
			m[*mapping.Target.Name] = *mapping.Value.Output
		}
	}
	return m
}
