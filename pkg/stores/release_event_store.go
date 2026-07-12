package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=release_event_store.go -destination=../mocks/mock_release_event_store.go -package=mocks

type ReleaseEventStore interface {
	// Insert assigns the next release-local sequence and inserts the event in
	// its own transaction. A duplicate (release_id, dedupe_key) is a no-op and
	// returns (nil, nil).
	Insert(ctx context.Context, event *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError)
	// InsertWithTx is Insert joining the caller's ambient transaction.
	// Must run inside a transaction (db.TxFromContext must be non-nil).
	InsertWithTx(ctx context.Context, event *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError)
	// ListByReleaseID returns events with sequence > afterSequence ordered by
	// sequence ASC, capped at limit.
	ListByReleaseID(ctx context.Context, releaseID string, afterSequence int, limit int) ([]*models.ReleaseEvent, *errors.ServiceError)
	AtomicExecutor
}
