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
			events:        eventStore,
			releases:      releaseStore,
			releaseID:     streamReleaseID,
			afterSeq:      0,
			pollInterval:  10 * time.Millisecond,
			graceInterval: 2 * time.Millisecond,
			presentEvent:  defaultPresentReleaseEvent,
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
			// Subsequent polls advance to the last emitted sequence and drain
			// empty. Terminal-and-empty triggers one grace re-poll, so this is
			// hit twice before the stream closes.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 4, releaseEventsMaxLimit).
				Return(nil, nil).
				MinTimes(2),
		)
		// Empty polls check release state; terminal → close after the grace round.
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateReleased}, nil).
			MinTimes(2)

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

	It("closes the channel once the release is terminal and polls return no new events", func() {
		ctx := context.Background()

		// Terminal with genuinely no events: the first empty+terminal poll grace
		// re-polls, the second closes. Two empty polls, two state checks.
		eventStore.EXPECT().
			ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, nil).
			MinTimes(2)
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateFailed}, nil).
			MinTimes(2)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(collect(ch)).To(BeEmpty())
	})

	It("flushes the terminal event that lands after the state CAS, then closes (the race)", func() {
		ctx := context.Background()

		// The producer CAS'd the release to terminal but has not yet inserted the
		// terminal event. The first poll therefore sees an empty event list while
		// GetByID already reports terminal. The streamer must NOT close here: it
		// grace re-polls, the terminal event lands, gets flushed, and only then
		// does the stream close. Against a streamer that closes on the first
		// empty+terminal poll, this event is dropped and the assertion fails.
		terminalEvent := &models.ReleaseEvent{
			ID:        "e1",
			ReleaseID: streamReleaseID,
			StackID:   "stack-1",
			Sequence:  1,
			Type:      models.ReleaseEventTypeReleaseReleased,
			Message:   "released",
		}

		gomock.InOrder(
			// Race poll: empty list while the release is already terminal.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
				Return(nil, nil),
			// Grace re-poll: the terminal event insert has now landed.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
				Return([]*models.ReleaseEvent{terminalEvent}, nil),
			// After flushing, drains empty again and closes after the grace round.
			eventStore.EXPECT().
				ListByReleaseID(ctx, streamReleaseID, 1, releaseEventsMaxLimit).
				Return(nil, nil).
				MinTimes(2),
		)
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateReleased}, nil).
			MinTimes(3)

		ch, err := streamer.Stream(ctx)
		Expect(err).ToNot(HaveOccurred())

		got := collect(ch)
		Expect(got).To(HaveLen(1), "the terminal event must be flushed before close")
		Expect(got[0].Error()).ToNot(HaveOccurred())
		withID, ok := got[0].(interfaces.StreamObjectWithID)
		Expect(ok).To(BeTrue())
		Expect(withID.ID()).To(Equal(strconv.Itoa(terminalEvent.Sequence)))
	})

	It("closes (bounded) when the release is terminal but its terminal event never lands (producer crash)", func() {
		ctx := context.Background()

		// Degenerate case: the producer crashed between the state CAS and the
		// terminal event insert, so polls stay empty forever while the release
		// reports terminal. The grace re-poll approach still closes after the
		// grace round rather than hanging the stream.
		eventStore.EXPECT().
			ListByReleaseID(ctx, streamReleaseID, 0, releaseEventsMaxLimit).
			Return(nil, nil).
			AnyTimes()
		releaseStore.EXPECT().
			GetByID(ctx, streamReleaseID).
			Return(&models.StackRelease{ID: streamReleaseID, State: models.ReleaseStateFailed}, nil).
			AnyTimes()

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
