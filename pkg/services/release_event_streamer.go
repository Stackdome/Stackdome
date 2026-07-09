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
	events       stores.ReleaseEventStore
	releases     stores.StackReleaseStore
	releaseID    string
	afterSeq     int
	pollInterval time.Duration
	presentEvent func(*models.ReleaseEvent) ([]byte, error)
}

func (s *releaseEventStreamer) Stream(ctx context.Context) (<-chan interfaces.StreamObject, error) {
	out := make(chan interfaces.StreamObject)
	go func() {
		defer close(out)

		cursor := s.afterSeq
		ticker := time.NewTicker(s.pollInterval)
		defer ticker.Stop()

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

			// A poll that returned nothing new is the only moment it is safe to
			// close: everything persisted so far has been flushed. If the release
			// has also reached a terminal state, no more events can arrive.
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
					return
				}
			}

			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}
