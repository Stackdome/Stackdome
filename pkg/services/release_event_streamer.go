package services

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/Stackdome/stackdome/pkg/interfaces"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// releaseEventStreamPollInterval is how often the streamer polls the event
// store for events newer than its cursor. Tests override the per-streamer
// pollInterval field to a much smaller value.
const releaseEventStreamPollInterval = 2 * time.Second

// releaseEventStreamTerminalGraceInterval is the shorter delay used for the one
// extra "grace" re-poll after the streamer first observes an empty poll while
// the release is already terminal. Producers write the terminal state CAS
// *before* inserting the terminal event, so a poll landing between the two sees
// terminal state with the terminal event not yet persisted. The grace re-poll
// gives that insert a brief window to land so it gets flushed before close.
// Tests override the per-streamer graceInterval field to a much smaller value.
const releaseEventStreamTerminalGraceInterval = 300 * time.Millisecond

// defaultPresentReleaseEvent marshals a release event into the same wire shape
// the list endpoint returns for a single event, so SSE `data:` frames and the
// list response are byte-for-byte identical per event.
func defaultPresentReleaseEvent(e *models.ReleaseEvent) ([]byte, error) {
	return json.Marshal(presenters.PresentReleaseEvent(e))
}

// releaseEventStreamObject is a single SSE frame. It carries either a data
// payload (with its sequence as the SSE id) or a terminal error.
type releaseEventStreamObject struct {
	data string
	id   string
	err  error
}

func (o releaseEventStreamObject) Data() string { return o.data }
func (o releaseEventStreamObject) Error() error { return o.err }
func (o releaseEventStreamObject) ID() string   { return o.id }

// releaseEventStreamer implements interfaces.ServerSideStreamable by polling the
// release event store and emitting each new event as an SSE frame. It closes the
// stream once the release is terminal and a poll returns no further events.
type releaseEventStreamer struct {
	events        stores.ReleaseEventStore
	releases      stores.StackReleaseStore
	releaseID     string
	afterSeq      int
	pollInterval  time.Duration
	graceInterval time.Duration
	presentEvent  func(*models.ReleaseEvent) ([]byte, error)
}

func (s *releaseEventStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	out := make(chan interfaces.StreamObject)
	go func() {
		defer close(out)

		cursor := s.afterSeq

		// terminalSeen records that the previous poll was empty while the release
		// was already terminal. Producers CAS the terminal state *before*
		// inserting the terminal event, so the first such poll may race ahead of
		// that insert. Rather than close and drop the terminal event, we grace
		// re-poll once; only a *second* consecutive empty poll under a terminal
		// state closes the stream. Any events flushed in between reset the flag,
		// so the terminal event always gets emitted before close. If the producer
		// crashed between the two writes and no terminal event ever lands, the
		// stream still closes after the grace round — bounded, never hanging.
		terminalSeen := false

		for {
			events, serr := s.events.ListByReleaseID(ctx, s.releaseID, cursor, releaseEventsMaxLimit)
			if serr != nil {
				select {
				case out <- releaseEventStreamObject{err: serr}:
				case <-ctx.Done():
				}
				return
			}

			for _, ev := range events {
				payload, err := s.presentEvent(ev)
				if err != nil {
					select {
					case out <- releaseEventStreamObject{err: err}:
					case <-ctx.Done():
					}
					return
				}
				select {
				case out <- releaseEventStreamObject{data: string(payload), id: strconv.Itoa(ev.Sequence)}:
					cursor = ev.Sequence
				case <-ctx.Done():
					return
				}
			}

			// A poll that returned events means work is still landing; reset the
			// grace flag so a fresh empty+terminal round is required before close.
			if len(events) > 0 {
				terminalSeen = false
			}

			// The delay before the next poll. A terminal-and-empty poll shortens
			// it to the grace interval so the pending terminal event is flushed
			// promptly rather than after a full poll interval.
			delay := s.pollInterval

			// A poll that returned nothing new is the only moment it is safe to
			// consider closing: everything persisted so far has been flushed. If
			// the release has also reached a terminal state, no more events can
			// arrive — but we grace re-poll once to let a just-CAS'd terminal
			// event's insert land before closing.
			if len(events) == 0 {
				release, serr := s.releases.GetByID(ctx, s.releaseID)
				if serr != nil {
					select {
					case out <- releaseEventStreamObject{err: serr}:
					case <-ctx.Done():
					}
					return
				}
				if release.State.Terminal() {
					if terminalSeen {
						return
					}
					terminalSeen = true
					delay = s.graceInterval
				}
			}

			timer := time.NewTimer(delay)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
	}()
	return out, nil
}
