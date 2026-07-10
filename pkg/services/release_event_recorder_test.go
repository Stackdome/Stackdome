package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func newTestRelease() *models.StackRelease {
	return &models.StackRelease{ID: "rel-1", StackID: "stack-1", Sequence: 7}
}

var _ = Describe("ReleaseEventRecorder", func() {
	var (
		ctrl  *gomock.Controller
		store *mocks.MockReleaseEventStore
		rec   ReleaseEventRecorder
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockReleaseEventStore(ctrl)
		rec = NewReleaseEventRecorder(ReleaseEventRecorderSpec{Store: store})
	})

	expectInsert := func(assert func(ev *models.ReleaseEvent)) {
		store.EXPECT().Insert(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, ev *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError) {
				assert(ev)
				return ev, nil
			})
	}

	Describe("RecordReleaseCreated", func() {
		It("uses the caller's ambient transaction", func() {
			store.EXPECT().InsertWithTx(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, ev *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError) {
					Expect(ev.Type).To(Equal(models.ReleaseEventTypeReleaseCreated))
					Expect(ev.Scope).To(Equal(models.ReleaseEventScopeRelease))
					Expect(ev.Level).To(Equal(models.ReleaseEventLevelInfo))
					Expect(ev.ReleaseID).To(Equal("rel-1"))
					Expect(ev.StackID).To(Equal("stack-1"))
					Expect(ev.DedupeKey).To(Equal("release:created"))
					return ev, nil
				})

			Expect(rec.RecordReleaseCreated(context.Background(), newTestRelease())).To(BeNil())
		})
	})

	Describe("RecordReleaseCheckFailed", func() {
		It("records a resource-scoped event with check metadata", func() {
			store.EXPECT().Insert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, ev *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError) {
					Expect(ev.Type).To(Equal(models.ReleaseEventTypeReleaseCheckFailed))
					Expect(ev.Scope).To(Equal(models.ReleaseEventScopeResource))
					Expect(*ev.ResourceName).To(Equal("api"))
					Expect(ev.DedupeKey).To(Equal("release:check_failed:api:image_pull"))
					Expect(ev.Metadata[models.ReleaseEventMetaCheck]).To(Equal("image_pull"))
					Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("image not found"))
					return ev, nil
				})

			Expect(rec.RecordReleaseCheckFailed(context.Background(), newTestRelease(), "api", "image_pull", "image not found")).To(BeNil())
		})
	})

	Describe("RecordBuildEvent", func() {
		const (
			resourceName = "api"
			buildID      = "build-1"
		)

		It("records build_failed preferring the failure message and keeping the reason in metadata", func() {
			failure := &models.BuildFailureDetail{Reason: "OOMKilled", Message: "container ran out of memory"}
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Type).To(Equal(models.ReleaseEventTypeBuildFailed))
				Expect(ev.Scope).To(Equal(models.ReleaseEventScopeResource))
				Expect(ev.Level).To(Equal(models.ReleaseEventLevelError))
				Expect(*ev.ResourceName).To(Equal(resourceName))
				Expect(ev.Message).To(Equal("Build failed for api: container ran out of memory"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("OOMKilled"))
				Expect(ev.Metadata[models.ReleaseEventMetaBuildID]).To(Equal(buildID))
				Expect(ev.Links).To(HaveLen(1))
				Expect(ev.Links[0].Kind).To(Equal(models.ReleaseEventLinkKindBuildLogs))
			})

			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildFailed, buildID, "", failure)).To(BeNil())
		})

		It("falls back to the failure reason in the build_failed message when no message was captured", func() {
			failure := &models.BuildFailureDetail{Reason: "OOMKilled"}
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Message).To(Equal("Build failed for api: OOMKilled"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("OOMKilled"))
			})

			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildFailed, buildID, "", failure)).To(BeNil())
		})

		It("omits failure detail and reason metadata when build_failed has no failure", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Message).To(Equal("Build failed for api"))
				Expect(ev.Metadata).NotTo(HaveKey(models.ReleaseEventMetaReason))
			})

			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildFailed, buildID, "", nil)).To(BeNil())
		})

		It("records build_attempt_failed as a resource-scoped warning with retry wording", func() {
			failure := &models.BuildFailureDetail{Reason: "StepFailed", Message: "step 3 exited with code 1"}
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Type).To(Equal(models.ReleaseEventTypeBuildAttemptFailed))
				Expect(ev.Scope).To(Equal(models.ReleaseEventScopeResource))
				Expect(ev.Level).To(Equal(models.ReleaseEventLevelWarning))
				Expect(ev.Message).To(Equal("Build attempt failed for api (will retry): step 3 exited with code 1"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("StepFailed"))
			})

			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildAttemptFailed, buildID, "", failure)).To(BeNil())
		})

		It("records attribution metadata when provided", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Metadata[models.ReleaseEventMetaAttribution]).To(Equal("dependency"))
			})

			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildStarted, buildID, "dependency", nil)).To(BeNil())
		})

		It("rejects a non-build event type", func() {
			Expect(rec.RecordBuildEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceReady, buildID, "", nil)).NotTo(BeNil())
		})
	})

	Describe("RecordResourceEvent", func() {
		const resourceName = "api"

		It("prefers the message in resource_failed text and keeps the reason in metadata", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Type).To(Equal(models.ReleaseEventTypeResourceFailed))
				Expect(ev.Level).To(Equal(models.ReleaseEventLevelError))
				Expect(*ev.ResourceName).To(Equal(resourceName))
				Expect(ev.Message).To(Equal("api failed to start: back-off restarting container"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("CrashLoopBackOff"))
			})

			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceFailed, "CrashLoopBackOff", "back-off restarting container")).To(BeNil())
		})

		It("falls back to the reason in resource_failed text when the message is empty", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Message).To(Equal("api failed to start: CrashLoopBackOff"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("CrashLoopBackOff"))
			})

			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceFailed, "CrashLoopBackOff", "")).To(BeNil())
		})

		It("records resource_waiting with the detail and reason metadata", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Type).To(Equal(models.ReleaseEventTypeResourceWaiting))
				Expect(ev.Level).To(Equal(models.ReleaseEventLevelWarning))
				Expect(ev.Message).To(Equal("api is waiting: waiting for volume to bind"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("VolumesPending"))
			})

			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceWaiting, "VolumesPending", "waiting for volume to bind")).To(BeNil())
		})

		It("records resource_deploying with the convergence detail", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Type).To(Equal(models.ReleaseEventTypeResourceDeploying))
				Expect(ev.Level).To(Equal(models.ReleaseEventLevelInfo))
				Expect(ev.Message).To(Equal("Deploying api: previous revision still serving; rollout not converged: 1/2 updated"))
				Expect(ev.Metadata[models.ReleaseEventMetaReason]).To(Equal("NotConverged"))
			})

			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceDeploying, "NotConverged", "previous revision still serving; rollout not converged: 1/2 updated")).To(BeNil())
		})

		It("omits reason metadata when the reason is empty", func() {
			expectInsert(func(ev *models.ReleaseEvent) {
				Expect(ev.Message).To(Equal("api is ready"))
				Expect(ev.Metadata).To(BeNil())
			})

			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeResourceReady, "", "")).To(BeNil())
		})

		It("rejects a non-resource event type", func() {
			Expect(rec.RecordResourceEvent(context.Background(), newTestRelease(), resourceName, models.ReleaseEventTypeBuildFailed, "", "")).NotTo(BeNil())
		})
	})

	Describe("Record", func() {
		It("rejects an unknown event type", func() {
			_, serr := rec.Record(context.Background(), ReleaseEventInput{
				Release: newTestRelease(),
				Type:    models.ReleaseEventType("nonsense"),
			})
			Expect(serr).NotTo(BeNil())
		})

		It("requires a resource name for resource-scoped types", func() {
			_, serr := rec.Record(context.Background(), ReleaseEventInput{
				Release: newTestRelease(),
				Type:    models.ReleaseEventTypeBuildStarted, // resource scope
			})
			Expect(serr).NotTo(BeNil())
		})
	})

})
