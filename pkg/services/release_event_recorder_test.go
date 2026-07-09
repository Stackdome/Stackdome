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

// Ginkgo supports exactly one RunSpecs call per test binary; this package's
// suite is bootstrapped by TestAESEncryptionService in
// encryption_service_test.go, which discovers every Describe below.

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

	Describe("NewReleaseEventRecorder", func() {
		It("panics when constructed with a nil store", func() {
			Expect(func() { NewReleaseEventRecorder(ReleaseEventRecorderSpec{}) }).To(Panic())
		})
	})
})
