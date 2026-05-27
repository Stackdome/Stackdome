package int

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Stack", func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("CRUD Operations", func() {
		It("should create a minimal stack", func() {
			stack := shared.CreateSimpleStack("test-create")
			created := shared.CreateStack(client, orgID, teamName, stack)

			Expect(created.GetId()).NotTo(BeEmpty(), "stack should have an ID")
			Expect(created.GetName()).To(Equal("test-create"))
			Expect(created.GetNamespace()).NotTo(BeEmpty(), "stack should have a namespace")
			Expect(created.GetOrganisationId()).To(Equal(orgID))

			Expect(created.Spec.StackResources).To(HaveLen(1))
			Expect(created.Spec.StackResources[0].GetName()).To(Equal("web"))
		})

		It("should get a stack by ID", func() {
			stack := shared.CreateSimpleStack("test-get")
			created := shared.CreateStack(client, orgID, teamName, stack)

			fetched := shared.GetStack(client, orgID, teamName, created.GetId())

			Expect(fetched.GetId()).To(Equal(created.GetId()))
			Expect(fetched.GetName()).To(Equal("test-get"))
			Expect(fetched.GetNamespace()).To(Equal(created.GetNamespace()))
			Expect(fetched.Spec.StackResources).To(HaveLen(1))
			Expect(fetched.Spec.StackResources[0].GetName()).To(Equal("web"))
		})

		It("should manage explicit connections independently and project them into topology", func() {
			api := openapi.NewStackResource("api")
			api.SetImageSpec(*openapi.NewImageSpec("nginx:1.25-alpine"))
			api.SetPorts([]openapi.Port{
				*openapi.NewPort("http", 8080, false),
			})

			web := openapi.NewStackResource("web")
			web.SetImageSpec(*openapi.NewImageSpec("nginx:1.25-alpine"))

			spec := openapi.NewStackSpec([]openapi.StackResource{*api, *web})
			stack := openapi.NewStack("test-connections", *spec)

			created := shared.CreateStack(client, orgID, teamName, stack)

			from := openapi.NewTopologyNodeRef("stack_resource")
			from.SetName("api")
			to := openapi.NewTopologyNodeRef("stack_resource")
			to.SetName("web")
			connection := openapi.NewStackConnection("env", *from, *to)

			target := openapi.NewConnectionTarget("env")
			target.SetName("API_HOST")
			value := openapi.NewValueRef()
			value.SetOutput("host")
			mapping := openapi.NewConnectionMapping(*target, *value)
			connection.SetMappings([]openapi.ConnectionMapping{*mapping})

			createdConnection := shared.CreateStackConnection(client, orgID, teamName, created.GetId(), connection)
			Expect(createdConnection.GetId()).NotTo(BeEmpty())
			Expect(createdConnection.GetMappings()).To(HaveLen(1))
			connectionID := createdConnection.GetId()

			fetched := shared.GetStack(client, orgID, teamName, created.GetId())
			Expect(fetched.Spec.GetConnections()).To(HaveLen(1))
			Expect(fetched.Spec.GetConnections()[0].GetId()).To(Equal(connectionID))

			listedConnections := shared.ListStackConnections(client, orgID, teamName, created.GetId())
			Expect(listedConnections).To(HaveLen(1))
			Expect(listedConnections[0].GetMappings()[0].Target.GetName()).To(Equal("API_HOST"))

			topology := shared.GetStackTopology(client, orgID, teamName, created.GetId())
			Expect(topology.GetEdges()).To(HaveLen(1))
			Expect(topology.GetEdges()[0].GetId()).To(Equal(connectionID))
			Expect(topology.GetEdges()[0].GetSourceOfTruth()).To(Equal("connection"))

			updatedTarget := openapi.NewConnectionTarget("env")
			updatedTarget.SetName("UPSTREAM_HOST")
			updatedMapping := openapi.NewConnectionMapping(*updatedTarget, *value)
			connection.SetMappings([]openapi.ConnectionMapping{*updatedMapping})

			updatedConnection := shared.UpdateStackConnection(client, orgID, teamName, created.GetId(), connectionID, connection)
			Expect(updatedConnection.GetMappings()).To(HaveLen(1))
			Expect(updatedConnection.GetMappings()[0].Target.GetName()).To(Equal("UPSTREAM_HOST"))

			listedConnections = shared.ListStackConnections(client, orgID, teamName, created.GetId())
			Expect(listedConnections).To(HaveLen(1))
			Expect(listedConnections[0].GetMappings()[0].Target.GetName()).To(Equal("UPSTREAM_HOST"))

			shared.DeleteStackConnection(client, orgID, teamName, created.GetId(), connectionID)

			Expect(shared.ListStackConnections(client, orgID, teamName, created.GetId())).To(BeEmpty())
			Expect(shared.GetStackTopology(client, orgID, teamName, created.GetId()).GetEdges()).To(BeEmpty())
		})

		It("should list stacks by organization", func() {
			stack1 := shared.CreateSimpleStack("test-list-1")
			stack2 := shared.CreateSimpleStack("test-list-2")
			shared.CreateStack(client, orgID, teamName, stack1)
			shared.CreateStack(client, orgID, teamName, stack2)

			list := shared.ListStacks(client, orgID, teamName)
			Expect(len(list.GetItems())).To(BeNumerically(">=", 2))
		})

		It("should update a stack", func() {
			stack := shared.CreateSimpleStack("test-update")
			created := shared.CreateStack(client, orgID, teamName, stack)

			updateStack := shared.CreateSimpleStack("test-update")
			exec := openapi.NewExecutionConfig()
			exec.SetEnvironmentVariables([]openapi.EnvVar{
				func() openapi.EnvVar { v := openapi.NewEnvVar("NEW_VAR"); v.SetValue("new-value"); return *v }(),
			})
			updateStack.Spec.StackResources[0].SetExecutionConfig(*exec)
			updateStack.Spec.StackResources[0].SetPorts([]openapi.Port{
				*openapi.NewPort("http", 80, false),
				*openapi.NewPort("https", 443, false),
			})

			updated := shared.UpdateStack(client, orgID, teamName, created.GetId(), updateStack)

			Expect(updated.GetId()).To(Equal(created.GetId()))
			Expect(updated.Spec.StackResources[0].ExecutionConfig).NotTo(BeNil())
			Expect(updated.Spec.StackResources[0].ExecutionConfig.GetEnvironmentVariables()).To(HaveLen(1))
			Expect(updated.Spec.StackResources[0].GetPorts()).To(HaveLen(2))
		})

		It("should delete a stack", func() {
			stack := shared.CreateSimpleStack("test-delete")
			created := shared.CreateStack(client, orgID, teamName, stack)

			deleted := shared.DeleteStack(client, orgID, teamName, created.GetId())

			status, ok := deleted.GetStatusOk()
			Expect(ok).To(BeTrue())
			state, stateOk := status.GetStateOk()
			Expect(stateOk).To(BeTrue())
			Expect(*state).To(Equal("Deleting"))
		})
	})

	Context("Validation", func() {
		It("should reject a stack with empty name", func() {
			resource := openapi.NewStackResource("web")
			resource.SetImageSpec(*openapi.NewImageSpec("nginx:1.25-alpine"))
			spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
			stack := openapi.NewStack("", *spec)

			shared.CreateStackExpectError(client, orgID, teamName, stack, http.StatusBadRequest)
		})

		It("should reject a stack with no resources", func() {
			spec := openapi.NewStackSpec([]openapi.StackResource{})
			stack := openapi.NewStack("test-no-resources", *spec)

			shared.CreateStackExpectError(client, orgID, teamName, stack, http.StatusBadRequest)
		})

		It("should reject a resource with neither build nor image spec", func() {
			resource := openapi.NewStackResource("web")
			spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
			stack := openapi.NewStack("test-no-image", *spec)

			shared.CreateStackExpectError(client, orgID, teamName, stack, http.StatusBadRequest)
		})

		It("should reject duplicate resource names", func() {
			resource1 := openapi.NewStackResource("web")
			resource1.SetImageSpec(*openapi.NewImageSpec("nginx:1.25-alpine"))
			resource2 := openapi.NewStackResource("web")
			resource2.SetImageSpec(*openapi.NewImageSpec("nginx:1.25-alpine"))

			spec := openapi.NewStackSpec([]openapi.StackResource{*resource1, *resource2})
			stack := openapi.NewStack("test-dup-resources", *spec)

			shared.CreateStackExpectError(client, orgID, teamName, stack, http.StatusBadRequest)
		})

		It("should reject duplicate stack names", func() {
			stack := shared.CreateSimpleStack("test-dup-name")
			shared.CreateStack(client, orgID, teamName, stack)

			duplicate := shared.CreateSimpleStack("test-dup-name")
			shared.CreateStackExpectError(client, orgID, teamName, duplicate, http.StatusConflict)
		})
	})
})
