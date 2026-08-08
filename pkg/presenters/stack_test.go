package presenters_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

func TestPresentStackIncludesConnections(t *testing.T) {
	stack := &models.Stack{
		Name: "demo",
		Connections: models.StackConnections{
			{
				ID:   "conn-1",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{
					Type: models.TopologyNodeTypePostgresAddon,
					Id:   "pg-1",
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
							Output: models.OutputNameURL,
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

	out := presenters.PresentStack(stack, models.StackReleaseRefs{})

	if len(out.Spec.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(out.Spec.Connections))
	}
	if out.Spec.Connections[0].Kind != string(models.ConnectionKindEnv) {
		t.Fatalf("expected env connection kind, got %q", out.Spec.Connections[0].Kind)
	}
	if out.Spec.Connections[0].From.GetId() != "pg-1" {
		t.Fatalf("expected from pg-1, got %q", out.Spec.Connections[0].From.GetId())
	}
	if out.Spec.Connections[0].Mappings[0].Value.GetOutput() != models.OutputNameURL {
		t.Fatalf("expected mapping output %s, got %q", models.OutputNameURL, out.Spec.Connections[0].Mappings[0].Value.GetOutput())
	}
	config := out.Spec.Connections[0].GetConfig()
	if config.PostgresEnvConfig == nil {
		t.Fatal("expected PostgresEnvConfig to be set")
	}
	if config.PostgresEnvConfig.GetDatabase() != "app" {
		t.Fatalf("expected database app, got %q", config.PostgresEnvConfig.GetDatabase())
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
	mapping.Value.SetOutput(models.OutputNameURL)
	conn.SetMappings([]openapi.ConnectionMapping{*mapping})
	db := "app"
	scope := "owner"
	conn.SetConfig(openapi.StackConnectionConfig{
		PostgresEnvConfig: &openapi.PostgresEnvConfig{
			Database:        &db,
			CredentialScope: &scope,
		},
	})

	spec := openapi.NewStackSpec()
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
	if out.Connections[0].Mappings[0].Value.Output != models.OutputNameURL {
		t.Fatalf("expected mapping output %s, got %q", models.OutputNameURL, out.Connections[0].Mappings[0].Value.Output)
	}
	if out.Connections[0].Config["database"] != "app" {
		t.Fatalf("expected connection config database app, got %#v", out.Connections[0].Config["database"])
	}
}

func TestPresentPostgresEnvConnectionWithUnrecognizedConfigReturnsNilConfig(t *testing.T) {
	stack := &models.Stack{
		Name: "demo",
		Connections: models.StackConnections{
			{
				ID:   "conn-1",
				Kind: models.ConnectionKindEnv,
				From: models.TopologyNodeRef{
					Type: models.TopologyNodeTypePostgresAddon,
					Id:   "pg-1",
				},
				To: models.TopologyNodeRef{
					Type: models.TopologyNodeTypeStackResource,
					Name: "web",
				},
				Config: map[string]interface{}{
					"unknown_key": "value",
				},
			},
		},
	}

	out := presenters.PresentStack(stack, models.StackReleaseRefs{})
	if len(out.Spec.Connections) != 1 {
		t.Fatalf("expected one connection, got %d", len(out.Spec.Connections))
	}
	cfg, ok := out.Spec.Connections[0].GetConfigOk()
	if ok && cfg != nil && cfg.PostgresEnvConfig != nil {
		t.Fatal("expected nil config when only unrecognized keys are present")
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
						{Name: "PUBLIC_URL", SelfOutput: models.OutputNamePublicURL},
					},
				},
			},
		},
	}

	presented := presenters.PresentStack(stack, models.StackReleaseRefs{})
	if len(presented.Spec.StackResources) != 1 || len(presented.Spec.StackResources[0].GetExecutionConfig().EnvironmentVariables) != 1 {
		t.Fatalf("expected one presented env var")
	}
	presentedEnv := presented.Spec.StackResources[0].GetExecutionConfig().EnvironmentVariables[0]
	if presentedEnv.GetSelfOutput() != models.OutputNamePublicURL {
		t.Fatalf("expected presented self_output, got %q", presentedEnv.GetSelfOutput())
	}

	spec := openapi.NewStackSpec()
	spec.SetStackResources([]openapi.StackResource{
		{
			Name:   "web",
			Source: &openapi.SourceSpec{Image: openapi.NewImageSource("nginx:latest")},
			Ports: []openapi.Port{
				{Name: "http", Number: 8080, Protocol: openapi.PtrString("http"), ExposedToPublic: true},
			},
			ExecutionConfig: &openapi.ExecutionConfig{
				EnvironmentVariables: []openapi.EnvVar{
					{
						Name:       "PUBLIC_URL",
						SelfOutput: openapi.PtrString(models.OutputNamePublicURL),
					},
				},
			},
		},
	})
	converted := presenters.ConvertStack(openapi.NewStack("demo", *spec))
	if converted.StackResources[0].ExecutionConfig.Env[0].SelfOutput != models.OutputNamePublicURL {
		t.Fatalf("expected converted self_output, got %q", converted.StackResources[0].ExecutionConfig.Env[0].SelfOutput)
	}
}

var _ = Describe("PresentStack release-centric fields", func() {
	It("presents lifecycle active and null releases for a fresh stack", func() {
		s := &models.Stack{ID: "s1", Name: "app"}
		out := presenters.PresentStack(s, models.StackReleaseRefs{})
		Expect(*out.Lifecycle).To(Equal(openapi.STACK_LIFECYCLE_ACTIVE))
		Expect(out.ConvergedRelease).To(BeNil())
		Expect(out.LatestRelease).To(BeNil())
	})

	It("presents lifecycle deleting when deletion timestamp set", func() {
		now := time.Now()
		s := &models.Stack{ID: "s1", DeletionTimestamp: &now}
		out := presenters.PresentStack(s, models.StackReleaseRefs{})
		Expect(*out.Lifecycle).To(Equal(openapi.STACK_LIFECYCLE_DELETING))
	})

	It("embeds current and latest release summaries with health", func() {
		current := &models.StackRelease{ID: "r1", Sequence: 3, State: models.ReleaseStateReleased, Message: "Release is live"}
		latest := &models.StackRelease{ID: "r2", Sequence: 4, State: models.ReleaseStateInProgress}
		s := &models.Stack{
			ID: "s1",
			Status: &models.StackStatus{
				LastConverged: &models.StackConvergenceRecord{ReleaseID: "r1"},
				Conditions: []models.Condition{
					{Type: string(models.StackConditionConverged), Status: string(models.ConditionTrue)},
					{Type: string(models.StackConditionAvailable), Status: string(models.ConditionTrue)},
				},
			},
			StackResources: []*models.StackResource{
				{Name: "web", Status: &models.StackResourceStatus{State: models.StackResourcePhaseReady}},
			},
		}
		out := presenters.PresentStack(s, models.StackReleaseRefs{Converged: current, Latest: latest})
		Expect(*out.ConvergedRelease.Id).To(Equal("r1"))
		Expect(*out.ConvergedRelease.Health).To(Equal(openapi.RELEASE_HEALTH_OK))
		Expect(*out.LatestRelease.Id).To(Equal("r2"))
		Expect(*out.LatestRelease.Health).To(Equal(openapi.RELEASE_HEALTH_PROGRESSING))
	})
})

// The shape a client needs to answer "is my deploy live, and where is it?".
var _ = Describe("released stack API contract", func() {
	const publicURL = "https://app.example.com"

	released := &models.StackRelease{ID: "r1", Sequence: 1, State: models.ReleaseStateReleased}

	stackWith := func(resources ...*models.StackResource) *models.Stack {
		return &models.Stack{
			ID: "s1",
			Status: &models.StackStatus{
				LastConverged: &models.StackConvergenceRecord{ReleaseID: released.ID},
				Conditions: []models.Condition{
					{Type: string(models.StackConditionConverged), Status: string(models.ConditionTrue)},
					{Type: string(models.StackConditionAvailable), Status: string(models.ConditionTrue)},
				},
			},
			StackResources: resources,
		}
	}

	webResource := func() *models.StackResource {
		return &models.StackResource{
			Name: "web",
			Status: &models.StackResourceStatus{
				State:               models.StackResourcePhaseReady,
				PublicIngresses:     []models.Ingress{{URL: publicURL, TargetPort: 8080}},
				InternalServiceName: ptr.To("web-svc"),
			},
		}
	}

	workerResource := func() *models.StackResource {
		return &models.StackResource{
			Name:   "worker",
			Status: &models.StackResourceStatus{State: models.StackResourcePhaseReady},
		}
	}

	liveStatusFor := func(resources ...*models.StackResource) *models.ReleaseLiveStatus {
		byName := map[string]*models.StackResourceStatus{}
		for _, r := range resources {
			byName[r.Name] = r.Status
		}
		return &models.ReleaseLiveStatus{Health: models.ReleaseHealthOK, Resources: byName}
	}

	It("reports the released and converged release with runtime health", func() {
		out := presenters.PresentStack(stackWith(webResource()), models.StackReleaseRefs{
			Converged: released,
			Latest:    released,
		})

		Expect(*out.LatestRelease.Id).To(Equal(released.ID))
		Expect(*out.LatestRelease.State).To(Equal(openapi.StackReleaseState(models.ReleaseStateReleased)))
		Expect(*out.ConvergedRelease.Id).To(Equal(released.ID))
		Expect(*out.ConvergedRelease.Health).To(Equal(openapi.RELEASE_HEALTH_OK))
	})

	It("reports the release id, Released state and health on the release itself", func() {
		out := presenters.PresentStackRelease(released, liveStatusFor(webResource()))

		Expect(*out.Id).To(Equal(released.ID))
		Expect(*out.State).To(Equal(openapi.StackReleaseState(models.ReleaseStateReleased)))
		Expect(*out.LiveStatus.Health).To(Equal(openapi.RELEASE_HEALTH_OK))
	})

	It("exposes the public ingress URL only for the resource with an exposed port", func() {
		out := presenters.PresentStackRelease(released, liveStatusFor(webResource(), workerResource()))

		resources := *out.LiveStatus.Resources
		Expect(resources["web"].PublicIngress).To(HaveLen(1))
		Expect(*resources["web"].PublicIngress[0].Url).To(Equal(publicURL))
		Expect(resources["worker"].PublicIngress).To(BeEmpty())
	})
})
