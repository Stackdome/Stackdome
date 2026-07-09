package services

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

var _ = Describe("releaseEventStreamer", func() {
	const streamReleaseID = "rel-stream-1"

	var (
		ctrl         *gomock.Controller
		eventStore   *mocks.MockReleaseEventStore
		releaseStore *mocks.MockStackReleaseStore
		streamer     *releaseEventStreamer
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		eventStore = mocks.NewMockReleaseEventStore(ctrl)
		releaseStore = mocks.NewMockStackReleaseStore(ctrl)
		streamer = &releaseEventStreamer{
			events:       eventStore,
			releases:     releaseStore,
			releaseID:    streamReleaseID,
			afterSeq:     0,
			pollInterval: 10 * time.Millisecond,
			presentEvent: defaultPresentReleaseEvent,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// collect drains the stream into a slice, guarded by a deadline so a bug
	// that fails to close the channel fails the test instead of hanging forever.
	collect := func(ch <-chan interfaces.StreamObject) []interfaces.StreamObject {
		GinkgoHelper()
		var got []interfaces.StreamObject
		timeout := time.After(2 * time.Second)
		for {
			select {
			case obj, ok := <-ch:
				if !ok {
					return got
				}
				got = append(got, obj)
			case <-timeout:
				Fail("stream did not close within the deadline")
			}
		}
	}

	It("replays persisted events after the cursor in order, with id=sequence and data=event JSON, then closes on terminal drain", func() {
		ctx := context.Background()
		streamer.afterSeq = 2

		ev3 := &models.ReleaseEvent{ID: "e3", ReleaseID: streamReleaseID, StackID: "stack-1", Sequence: 3, Message: "third"}
		ev4 := &models.ReleaseEvent{ID: "e4", ReleaseID: streamReleaseID, StackID: "stack-1", Sequence: 4, Message: "fourth"}

		gomock.InOrder(
			// First poll from the requested cursor returns the two new events.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 2, releaseEventsMaxLimit).
				Return([]*models.ReleaseEvent{ev3, ev4}, nil),
			// Next poll advances to the last emitted sequence and drains empty.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 4, releaseEventsMaxLimit).
				Return(nil, nil),
		)
		// Empty poll checks release state; terminal → close.
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateReleased}, nil)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		got := collect(ch)
		Expect(got).To(HaveLen(2))

		for i, ev := range []*models.ReleaseEvent{ev3, ev4} {
			withID, ok := got[i].(interfaces.StreamObjectWithID)
			Expect(ok).To(BeTrue(), "frame should expose an SSE id")
			Expect(withID.ID()).To(Equal(strconv.Itoa(ev.Sequence)))
			Expect(got[i].Error()).ToNot(HaveOccurred())

			var decoded map[string]interface{}
			Expect(json.Unmarshal([]byte(got[i].Data()), &decoded)).To(Succeed())

			expected, mErr := json.Marshal(presenters.PresentReleaseEvent(ev))
			Expect(mErr).ToNot(HaveOccurred())
			Expect(got[i].Data()).To(Equal(string(expected)))
		}
	})

	It("closes the channel once the release is terminal and a poll returns no new events", func() {
		ctx := context.Background()

		eventStore.EXPECT().
			ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, nil)
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateFailed}, nil)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(collect(ch)).To(BeEmpty())
	})

	It("closes the channel promptly when the context is cancelled", func() {
		ctx, cancel := context.WithCancel(context.Background())

		// Non-terminal release: without cancellation the streamer would poll
		// forever. AnyTimes covers the race between cancel and the poll loop.
		eventStore.EXPECT().
			ListByReleaseID(gomock.Any(), streamReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, nil).
			AnyTimes()
		releaseStore.EXPECT().
			GetByID(gomock.Any(), streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStatePending}, nil).
			AnyTimes()

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		cancel()

		Eventually(func() bool {
			select {
			case _, ok := <-ch:
				return !ok
			default:
				return false
			}
		}, time.Second, 5*time.Millisecond).Should(BeTrue(), "channel should close after context cancellation")
	})

	It("emits an error frame then closes when the event store returns an error", func() {
		ctx := context.Background()
		storeErr := errors.GeneralError("boom")

		eventStore.EXPECT().
			ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, storeErr)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		got := collect(ch)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Error()).To(HaveOccurred())
	})

	It("surfaces a presenter error as an error frame then closes", func() {
		ctx := context.Background()
		presentErr := stderrors.New("marshal failed")
		streamer.presentEvent = func(*models.ReleaseEvent) ([]byte, error) { return nil, presentErr }

		eventStore.EXPECT().
			ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
			Return([]*models.ReleaseEvent{{Sequence: 1}}, nil)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		got := collect(ch)
		Expect(got).To(HaveLen(1))
		Expect(got[0].Error()).To(MatchError(presentErr))
	})
})
