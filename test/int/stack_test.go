package int

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
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
			api.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})
			api.SetPorts([]openapi.Port{
				*openapi.NewPort("http", 8080, false),
			})

			web := openapi.NewStackResource("web")
			web.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})

			spec := openapi.NewStackSpec()
			spec.SetStackResources([]openapi.StackResource{*api, *web})
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

			lifecycle, ok := deleted.GetLifecycleOk()
			Expect(ok).To(BeTrue())
			Expect(*lifecycle).To(Equal(openapi.STACK_LIFECYCLE_DELETING))
		})
	})

	Context("Validation", func() {
		It("should reject a stack with empty name", func() {
			resource := openapi.NewStackResource("web")
			resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})
			spec := openapi.NewStackSpec()
			spec.SetStackResources([]openapi.StackResource{*resource})
			stack := openapi.NewStack("", *spec)

			_ = shared.CreateStackExpectError(client, orgID, teamName, stack, http.StatusBadRequest)
		})

		It("should allow creating a stack with no resources", func() {
			spec := openapi.NewStackSpec()
			stack := openapi.NewStack("test-no-resources", *spec)

			created := shared.CreateStack(client, orgID, teamName, stack)
			Expect(created.Spec.StackResources).To(BeEmpty())
		})

		It("should reject a resource with neither build nor image spec", func() {
			created := shared.CreateShellStack(client, orgID, teamName,
				openapi.NewStack("test-no-image", *openapi.NewStackSpec()))

			resource := openapi.NewStackResource("web")
			spec := openapi.NewStackSpec()
			spec.SetStackResources([]openapi.StackResource{*resource})
			badStack := openapi.NewStack("test-no-image", *spec)

			_ = shared.ApplyStackExpectError(client, orgID, teamName, created.GetId(), badStack, http.StatusBadRequest)
		})

		It("should reject duplicate resource names", func() {
			created := shared.CreateShellStack(client, orgID, teamName,
				openapi.NewStack("test-dup-resources", *openapi.NewStackSpec()))

			resource1 := openapi.NewStackResource("web")
			resource1.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})
			resource2 := openapi.NewStackResource("web")
			resource2.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})

			spec := openapi.NewStackSpec()
			spec.SetStackResources([]openapi.StackResource{*resource1, *resource2})
			badStack := openapi.NewStack("test-dup-resources", *spec)

			_ = shared.ApplyStackExpectError(client, orgID, teamName, created.GetId(), badStack, http.StatusBadRequest)
		})

		It("should reject duplicate stack names", func() {
			stack := shared.CreateSimpleStack("test-dup-name")
			shared.CreateStack(client, orgID, teamName, stack)

			duplicate := shared.CreateSimpleStack("test-dup-name")
			_ = shared.CreateStackExpectError(client, orgID, teamName, duplicate, http.StatusConflict)
		})
	})

	Context("Shell-only update", func() {
		It("PUT /stacks/{id} preserves children", func() {
			By("Creating a fat stack with two resources and a connection")
			created := shared.CreateStack(client, orgID, teamName, shared.CreateMultiResourceStack("shell-preserve"))
			stackID := created.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
			})

			By("PUT-ing a shell-only body with a new label and empty resources/connections")
			shellUpdate := openapi.NewStack("shell-preserve", *openapi.NewStackSpec())
			shellUpdate.SetLabels([]openapi.Label{*openapi.NewLabel("tier", "shell-only")})

			updated := shared.UpdateStackShellH(client, orgID, teamName, stackID, shellUpdate)

			By("Verifying the label change is reflected")
			hasLabel := false
			for _, l := range updated.GetLabels() {
				if l.GetKey() == "tier" && l.GetValue() == "shell-only" {
					hasLabel = true
					break
				}
			}
			Expect(hasLabel).To(BeTrue(), "shell PUT should apply the label change")

			By("Verifying children were preserved, not wiped")
			topology := shared.GetStackTopology(client, orgID, teamName, stackID)
			Expect(topology.GetNodes()).To(HaveLen(2), "shell PUT must not drop resources")
			Expect(topology.GetEdges()).To(HaveLen(1), "shell PUT must not drop the connection edge")
			Expect(shared.ListStackConnections(client, orgID, teamName, stackID)).To(HaveLen(1), "shell PUT must preserve connections")
		})

		It("POST /stacks/{id}/resources with invalid source returns 400", func() {
			shell := shared.CreateShellStack(client, orgID, teamName,
				openapi.NewStack("shell-bad-resource", *openapi.NewStackSpec()))
			stackID := shell.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
			})

			resource := openapi.NewStackResource("bad")
			_ = shared.CreateStackResourceExpectError(client, orgID, teamName, stackID, resource, http.StatusBadRequest)
		})
	})
})
