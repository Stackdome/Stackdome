package presenters_test

import (
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
)

func TestPresentStackIncludesConnections(t *testing.T) {
	stack := &models.Stack{
		Name: "demo",
		Connections: models.StackConnections{
			{
				ID:   "conn-1",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{
					Type: models.TopologyNodeTypeStackResource,
					Name: "redis",
				},
				To: models.TopologyNodeRef{
					Type: models.TopologyNodeTypeStackResource,
					Name: "web",
				},
				Mappings: []models.ConnectionMapping{
					{
						Target: models.ConnectionTarget{
							Type: models.ConnectionTargetTypeEnv,
							Name: "REDIS_URL",
						},
						Value: models.ValueRef{
							Output: "url.http",
						},
					},
				},
				Config: map[string]interface{}{
					"database":         "app",
					"credential_scope": "owner",
				},
			},
		},
	}

	out := presenters.PresentStack(stack)

	if len(out.Spec.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(out.Spec.Connections))
	}
	if out.Spec.Connections[0].Kind != string(models.ConnectionKindEnv) {
		t.Fatalf("expected env connection kind, got %q", out.Spec.Connections[0].Kind)
	}
	if out.Spec.Connections[0].From.GetName() != "redis" {
		t.Fatalf("expected from redis, got %q", out.Spec.Connections[0].From.GetName())
	}
	if out.Spec.Connections[0].Mappings[0].Value.GetOutput() != "url.http" {
		t.Fatalf("expected mapping output url.http, got %q", out.Spec.Connections[0].Mappings[0].Value.GetOutput())
	}
	if out.Spec.Connections[0].GetConfig()["database"] != "app" {
		t.Fatalf("expected connection config database app, got %#v", out.Spec.Connections[0].GetConfig()["database"])
	}
}

func TestConvertStackIncludesConnections(t *testing.T) {
	conn := openapi.NewStackConnection(
		"env",
		*openapi.NewTopologyNodeRef("stack_resource"),
		*openapi.NewTopologyNodeRef("stack_resource"),
	)
	conn.SetId("conn-1")
	conn.From.SetName("redis")
	conn.To.SetName("web")

	mapping := openapi.NewConnectionMapping(
		*openapi.NewConnectionTarget("env"),
		*openapi.NewValueRef(),
	)
	mapping.Target.SetName("REDIS_URL")
	mapping.Value.SetOutput("url.http")
	conn.SetMappings([]openapi.ConnectionMapping{*mapping})
	conn.SetConfig(map[string]interface{}{
		"database":         "app",
		"credential_scope": "owner",
	})

	spec := openapi.NewStackSpec([]openapi.StackResource{})
	spec.SetConnections([]openapi.StackConnection{*conn})
	stack := openapi.NewStack("demo", *spec)

	out := presenters.ConvertStack(stack)

	if len(out.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(out.Connections))
	}
	if out.Connections[0].Kind != models.ConnectionKindEnv {
		t.Fatalf("expected env connection kind, got %q", out.Connections[0].Kind)
	}
	if out.Connections[0].From.Name != "redis" {
		t.Fatalf("expected from redis, got %q", out.Connections[0].From.Name)
	}
	if out.Connections[0].Mappings[0].Value.Output != "url.http" {
		t.Fatalf("expected mapping output url.http, got %q", out.Connections[0].Mappings[0].Value.Output)
	}
	if out.Connections[0].Config["database"] != "app" {
		t.Fatalf("expected connection config database app, got %#v", out.Connections[0].Config["database"])
	}
}

func TestPresentAndConvertStackPreservesEnvVarSelfOutput(t *testing.T) {
	stack := &models.Stack{
		Name: "demo",
		StackResources: []*models.StackResource{
			{
				Name:        "web",
				ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"},
				Ports:       models.Ports{{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true}},
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{
						{Name: "PUBLIC_URL", SelfOutput: "public.http.url"},
					},
				},
			},
		},
	}

	presented := presenters.PresentStack(stack)
	if len(presented.Spec.StackResources) != 1 || len(presented.Spec.StackResources[0].GetExecutionConfig().EnvironmentVariables) != 1 {
		t.Fatalf("expected one presented env var")
	}
	presentedEnv := presented.Spec.StackResources[0].GetExecutionConfig().EnvironmentVariables[0]
	if presentedEnv.GetSelfOutput() != "public.http.url" {
		t.Fatalf("expected presented self_output, got %q", presentedEnv.GetSelfOutput())
	}

	spec := openapi.NewStackSpec([]openapi.StackResource{
		{
			Name:      "web",
			ImageSpec: &openapi.ImageSpec{Image: "nginx:latest"},
			Ports: []openapi.Port{
				{Name: "http", Number: 8080, Protocol: openapi.PtrString("http"), ExposedToPublic: true},
			},
			ExecutionConfig: &openapi.ExecutionConfig{
				EnvironmentVariables: []openapi.EnvVar{
					{
						Name:       "PUBLIC_URL",
						SelfOutput: openapi.PtrString("public.http.url"),
					},
				},
			},
		},
	})
	converted := presenters.ConvertStack(openapi.NewStack("demo", *spec))
	if converted.StackResources[0].ExecutionConfig.Env[0].SelfOutput != "public.http.url" {
		t.Fatalf("expected converted self_output, got %q", converted.StackResources[0].ExecutionConfig.Env[0].SelfOutput)
	}
}
