package int

import (
	"sort"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

var _ = Describe("Stack dual-flow (thin vs fat)", func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	// buildThin drives the thin flow: POST a shell, then add each resource and
	// connection from the multi-resource fixture incrementally. It uses the
	// fixture only as a source of resource/connection objects and registers its
	// own cleanup. Returns the created shell stack.
	buildThin := func(name string) *openapi.Stack {
		fat := shared.CreateMultiResourceStack(name)

		shell := shared.CreateShellStack(client, orgID, teamName, openapi.NewStack(name, *openapi.NewStackSpec()))
		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, shell.GetId())
			shared.WaitForStackDeleted(client, orgID, teamName, shell.GetId(), 2*time.Minute)
		})

		// Resources must exist before connections that reference them.
		for i := range fat.Spec.StackResources {
			res := fat.Spec.StackResources[i]
			shared.AddStackResource(client, orgID, teamName, shell.GetId(), &res)
		}
		for i := range fat.Spec.Connections {
			conn := fat.Spec.Connections[i]
			shared.CreateStackConnection(client, orgID, teamName, shell.GetId(), &conn)
		}

		return shell
	}

	nodeLabels := func(topology *openapi.StackTopology) []string {
		labels := []string{}
		for _, n := range topology.GetNodes() {
			labels = append(labels, n.GetLabel())
		}
		sort.Strings(labels)
		return labels
	}

	It("thin flow: shell + incremental children deploys to Ready", func() {
		shell := buildThin("dual-thin")

		By("Releasing the incrementally-built stack")
		shared.CreateRelease(client, orgID, teamName, shell.GetId())

		By("Waiting for the stack to become Ready")
		ready := shared.WaitForStackReady(client, orgID, teamName, shell.GetId(), 5*time.Minute)

		status, ok := ready.GetStatusOk()
		Expect(ok).To(BeTrue(), "ready stack should have status")
		Expect(status.GetState()).To(Equal(string(models.StackReady)))
		Expect(status.GetConditions()).NotTo(BeEmpty(), "ready stack should have conditions")
	})

	It("fat flow: apply full document deploys to Ready", func() {
		created, _ := shared.CreateStackAndDeploy(client, orgID, teamName, shared.CreateMultiResourceStack("dual-fat"))
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
		})

		By("Waiting for the stack to become Ready")
		ready := shared.WaitForStackReady(client, orgID, teamName, stackID, 5*time.Minute)

		status, ok := ready.GetStatusOk()
		Expect(ok).To(BeTrue(), "ready stack should have status")
		Expect(status.GetState()).To(Equal(string(models.StackReady)))
	})

	It("both flows produce the same topology", func() {
		By("Building a stack via the thin flow")
		thin := buildThin("dual-thin-topo")

		By("Building an equivalent stack via the fat flow")
		fat := shared.CreateStack(client, orgID, teamName, shared.CreateMultiResourceStack("dual-fat2"))
		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, fat.GetId())
			shared.WaitForStackDeleted(client, orgID, teamName, fat.GetId(), 2*time.Minute)
		})

		thinTopology := shared.GetStackTopology(client, orgID, teamName, thin.GetId())
		fatTopology := shared.GetStackTopology(client, orgID, teamName, fat.GetId())

		By("Comparing node identities")
		Expect(nodeLabels(thinTopology)).To(Equal(nodeLabels(fatTopology)), "both flows should yield the same node set")

		By("Comparing edge counts")
		Expect(thinTopology.GetEdges()).To(HaveLen(len(fatTopology.GetEdges())), "both flows should yield the same number of edges")
	})

	It("thin update: PUT one resource leaves others intact", func() {
		thin := buildThin("dual-thin-upd")

		By("Updating only the backend resource")
		updated := openapi.NewStackResource(shared.MultiResourceBackendName)
		updated.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource(shared.TestImage)})
		updated.SetPorts([]openapi.Port{
			*openapi.NewPort("http", shared.MultiResourceBackendPort, false),
			*openapi.NewPort("metrics", shared.MultiResourceBackendPort+1, false),
		})
		exec := openapi.NewExecutionConfig()
		exec.SetEnvironmentVariables([]openapi.EnvVar{
			func() openapi.EnvVar { v := openapi.NewEnvVar("EXTRA"); v.SetValue("value"); return *v }(),
		})
		updated.SetExecutionConfig(*exec)
		shared.UpdateStackResourceH(client, orgID, teamName, thin.GetId(), shared.MultiResourceBackendName, updated)

		By("Verifying both resource nodes are still present")
		topology := shared.GetStackTopology(client, orgID, teamName, thin.GetId())
		Expect(topology.GetNodes()).To(HaveLen(2), "updating one resource must not drop the others")
		Expect(nodeLabels(topology)).To(Equal([]string{shared.MultiResourceBackendName, shared.MultiResourceFrontendName}))
	})

	It("fat update: apply with a resource removed deletes it", func() {
		fatDoc := shared.CreateMultiResourceStack("dual-fat-del")
		created := shared.CreateStack(client, orgID, teamName, fatDoc)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
		})

		By("Confirming the full stack has two resources and one connection")
		Expect(shared.GetStackTopology(client, orgID, teamName, stackID).GetNodes()).To(HaveLen(2))
		Expect(shared.ListStackConnections(client, orgID, teamName, stackID)).To(HaveLen(1))

		By("Applying a document containing only the backend resource and no connections")
		backend := fatDoc.Spec.StackResources[0]
		reducedSpec := openapi.NewStackSpec()
		reducedSpec.SetStackResources([]openapi.StackResource{backend})
		reduced := openapi.NewStack("dual-fat-del", *reducedSpec)
		shared.UpdateStack(client, orgID, teamName, stackID, reduced)

		By("Verifying the frontend node and the connection were delete-missing'd")
		topology := shared.GetStackTopology(client, orgID, teamName, stackID)
		Expect(topology.GetNodes()).To(HaveLen(1), "frontend resource should be removed")
		Expect(topology.GetNodes()[0].GetLabel()).To(Equal(shared.MultiResourceBackendName))
		Expect(shared.ListStackConnections(client, orgID, teamName, stackID)).To(BeEmpty(), "connection referencing the removed resource should be gone")
	})
})
