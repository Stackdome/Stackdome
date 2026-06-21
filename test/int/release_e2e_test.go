package int

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Release E2E", Ordered, func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeAll(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil())
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Release Lifecycle", func() {
		It("should deploy via release and reach Released state", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating a stack (draft)")
			stack := shared.CreateSimpleStack("test-release-lifecycle")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()
			stackName := created.GetName()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Creating a release (deploy)")
			release := shared.CreateRelease(client, orgID, teamName, stackID)
			Expect(release.GetSequence()).To(Equal(int32(1)))
			Expect(string(release.GetState())).To(Equal("Pending"))

			By("Waiting for release to reach Released state")
			released := shared.WaitForReleaseReleased(client, orgID, teamName, stackID, release.GetId(), 5*time.Minute)
			Expect(released.GetRenderedAt()).NotTo(BeZero())
			Expect(released.GetCompletedAt()).NotTo(BeZero())

			By("Verifying Stack CR exists in cluster")
			shared.WaitForStackCRExists(ctx, clusterClient, stackName, namespace, 1*time.Minute)

			By("Verifying StackResource CR is Available")
			shared.WaitForStackResourceCRAvailable(ctx, clusterClient, "web", namespace, 3*time.Minute)

			By("Verifying stack is Ready via API")
			shared.WaitForStackReady(client, orgID, teamName, stackID, 2*time.Minute)

			By("Listing releases — should have exactly 1")
			list := shared.ListReleases(client, orgID, teamName, stackID)
			Expect(list.GetItems()).To(HaveLen(1))
		})

		It("should deploy a second release with updated image", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating stack + deploying release #1")
			stack := shared.CreateSimpleStack("test-release-update")
			created, rel1 := shared.CreateStackAndDeploy(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Waiting for release #1 to converge")
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel1.GetId(), 5*time.Minute)
			shared.WaitForStackReady(client, orgID, teamName, stackID, 2*time.Minute)

			By("Updating the stack resource image")
			updateStack := shared.CreateSimpleStack("test-release-update")
			updateStack.Spec.StackResources[0].ImageSpec.SetImage("nginx:1.26-alpine")
			shared.UpdateStack(client, orgID, teamName, stackID, updateStack)

			By("Creating release #2")
			rel2 := shared.CreateRelease(client, orgID, teamName, stackID)
			Expect(rel2.GetSequence()).To(Equal(int32(2)))

			By("Waiting for release #2 to converge")
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel2.GetId(), 5*time.Minute)

			By("Verifying StackResource CR has updated image")
			srCR, err := shared.GetStackResourceCR(ctx, clusterClient, "web", namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(srCR.Spec.ImageSpec.Image).To(Equal("nginx:1.26-alpine"))

			By("Listing releases — should have 2")
			list := shared.ListReleases(client, orgID, teamName, stackID)
			Expect(list.GetItems()).To(HaveLen(2))
		})

		It("should rollback to a previous release", func() {
			testEnv := GetEnvironment()
			clusterClient := testEnv.Cluster.GetClient()
			ctx := context.Background()

			By("Creating stack + deploying release #1 with nginx:1.25")
			stack := shared.CreateSimpleStack("test-release-rollback")
			created, rel1 := shared.CreateStackAndDeploy(client, orgID, teamName, stack)
			stackID := created.GetId()
			namespace := created.GetNamespace()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel1.GetId(), 5*time.Minute)
			shared.WaitForStackReady(client, orgID, teamName, stackID, 2*time.Minute)

			By("Updating image to nginx:1.26 + deploying release #2")
			updateStack := shared.CreateSimpleStack("test-release-rollback")
			updateStack.Spec.StackResources[0].ImageSpec.SetImage("nginx:1.26-alpine")
			shared.UpdateStack(client, orgID, teamName, stackID, updateStack)
			rel2 := shared.CreateRelease(client, orgID, teamName, stackID)
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel2.GetId(), 5*time.Minute)

			By("Verifying new image is deployed")
			srCR, err := shared.GetStackResourceCR(ctx, clusterClient, "web", namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(srCR.Spec.ImageSpec.Image).To(Equal("nginx:1.26-alpine"))

			By("Rolling back to release #1")
			rel3 := shared.RollbackRelease(client, orgID, teamName, stackID, rel1.GetId())
			Expect(rel3.GetSequence()).To(Equal(int32(3)))
			cause := rel3.GetCause()
			Expect(string(cause.GetKind())).To(Equal("rollback"))

			By("Waiting for rollback release to converge")
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel3.GetId(), 5*time.Minute)

			By("Verifying original image is restored")
			srCR, err = shared.GetStackResourceCR(ctx, clusterClient, "web", namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(srCR.Spec.ImageSpec.Image).To(Equal("nginx:1.25-alpine"))

			By("Listing releases — should have 3")
			list := shared.ListReleases(client, orgID, teamName, stackID)
			Expect(list.GetItems()).To(HaveLen(3))
		})

		It("should supersede a pending release with a new one", func() {
			By("Creating a stack")
			stack := shared.CreateSimpleStack("test-release-supersede")
			created := shared.CreateStack(client, orgID, teamName, stack)
			stackID := created.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Creating release #1")
			rel1 := shared.CreateRelease(client, orgID, teamName, stackID)

			By("Immediately creating release #2 (should supersede #1)")
			rel2 := shared.CreateRelease(client, orgID, teamName, stackID)
			Expect(rel2.GetSequence()).To(Equal(int32(2)))

			By("Waiting for release #2 to reach Released")
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, rel2.GetId(), 5*time.Minute)

			By("Verifying release #1 was superseded")
			r1 := shared.GetRelease(client, orgID, teamName, stackID, rel1.GetId())
			Expect(string(r1.GetState())).To(Equal("Superseded"))
		})
	})
})
