package int

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
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

			By("Fetching the release event timeline")
			events := shared.ListReleaseEvents(client, orgID, teamName, stackID, release.GetId()).GetItems()
			Expect(events).NotTo(BeEmpty(), "expected a non-empty release event timeline")

			By("Asserting the first event is release_created with sequence 1")
			Expect(events[0].GetType()).To(Equal(string(models.ReleaseEventTypeReleaseCreated)))
			Expect(events[0].GetSequence()).To(Equal(int32(1)))

			By("Asserting sequences are strictly ascending (gaps allowed)")
			for i := 1; i < len(events); i++ {
				Expect(events[i].GetSequence()).To(BeNumerically(">", events[i-1].GetSequence()),
					"event %d sequence %d must be strictly greater than previous %d",
					i, events[i].GetSequence(), events[i-1].GetSequence())
			}

			By("Asserting every event type is within the known 18-constant vocabulary")
			vocabulary := map[string]struct{}{
				string(models.ReleaseEventTypeReleaseCreated):       {},
				string(models.ReleaseEventTypeReleaseChecksStarted): {},
				string(models.ReleaseEventTypeReleaseCheckFailed):   {},
				string(models.ReleaseEventTypeReleaseChecksPassed):  {},
				string(models.ReleaseEventTypeReleaseStarted):       {},
				string(models.ReleaseEventTypeBuildQueued):          {},
				string(models.ReleaseEventTypeBuildStarted):         {},
				string(models.ReleaseEventTypeBuildSucceeded):       {},
				string(models.ReleaseEventTypeBuildAttemptFailed):   {},
				string(models.ReleaseEventTypeBuildFailed):          {},
				string(models.ReleaseEventTypeResourceWaiting):      {},
				string(models.ReleaseEventTypeResourceDeploying):    {},
				string(models.ReleaseEventTypeResourceReady):        {},
				string(models.ReleaseEventTypeResourceFailed):       {},
				string(models.ReleaseEventTypeReleaseReleased):      {},
				string(models.ReleaseEventTypeReleaseFailed):        {},
				string(models.ReleaseEventTypeReleaseSuperseded):    {},
				string(models.ReleaseEventTypeReleaseCancelled):     {},
			}
			Expect(vocabulary).To(HaveLen(18), "the release event vocabulary must contain exactly 18 types")

			counts := map[string]int{}
			indexOf := map[string]int{}
			for i, ev := range events {
				Expect(vocabulary).To(HaveKey(ev.GetType()),
					"event type %q at index %d is outside the known vocabulary", ev.GetType(), i)
				counts[ev.GetType()]++
				if _, seen := indexOf[ev.GetType()]; !seen {
					indexOf[ev.GetType()] = i
				}
			}

			By("Asserting checks_passed, started, and released each appear exactly once")
			Expect(counts[string(models.ReleaseEventTypeReleaseChecksPassed)]).To(Equal(1),
				"expected exactly one release_checks_passed event")
			Expect(counts[string(models.ReleaseEventTypeReleaseStarted)]).To(Equal(1),
				"expected exactly one release_started event")
			Expect(counts[string(models.ReleaseEventTypeReleaseReleased)]).To(Equal(1),
				"expected exactly one release_released event")

			By("Asserting release_created opens and release_released closes the lifecycle")
			// The release worker runs its sub-reconcilers in a fixed sequence
			// (gatekeeper -> simulator -> validation -> ...). The gatekeeper marks the
			// release InProgress and records release_started; the later validation stage
			// records release_checks_started/release_checks_passed. That fixed ordering
			// means release_started ALWAYS precedes the validation events, and
			// release_created / release_released bookend the timeline.
			Expect(indexOf[string(models.ReleaseEventTypeReleaseCreated)]).To(Equal(0),
				"release_created must be the first event")
			Expect(indexOf[string(models.ReleaseEventTypeReleaseStarted)]).To(
				BeNumerically("<", indexOf[string(models.ReleaseEventTypeReleaseChecksPassed)]),
				"release_started (gatekeeper) must precede release_checks_passed (validation)")
			if checksStartedIdx, ok := indexOf[string(models.ReleaseEventTypeReleaseChecksStarted)]; ok {
				Expect(indexOf[string(models.ReleaseEventTypeReleaseStarted)]).To(
					BeNumerically("<", checksStartedIdx),
					"release_started (gatekeeper) must precede release_checks_started (validation)")
			}
			Expect(indexOf[string(models.ReleaseEventTypeReleaseReleased)]).To(
				BeNumerically(">", indexOf[string(models.ReleaseEventTypeReleaseStarted)]),
				"release_released must come after release_started")
			Expect(indexOf[string(models.ReleaseEventTypeReleaseReleased)]).To(
				BeNumerically(">", indexOf[string(models.ReleaseEventTypeReleaseChecksPassed)]),
				"release_released must come after release_checks_passed")
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
			updateStack.Spec.StackResources[0].Source.Image.Ref = "nginx:1.26-alpine"
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
			updateStack.Spec.StackResources[0].Source.Image.Ref = "nginx:1.26-alpine"
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
