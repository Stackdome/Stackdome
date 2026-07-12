package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=stack_release_store.go -destination=../mocks/mock_stack_release_store.go -package=mocks
type StackReleaseStore interface {
	// Create inserts a new release with sequence = max(sequence)+1.
	Create(ctx context.Context, release *models.StackRelease) (*models.StackRelease, *errors.ServiceError)

	GetByID(ctx context.Context, id string) (*models.StackRelease, *errors.ServiceError)

	ListByStackID(ctx context.Context, stackID string, params ListParams) (*PaginatedResult[*models.StackRelease], *errors.ServiceError)

	// ListActive returns all non-terminal releases across all stacks.
	ListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError)

	// GetActiveByStackID returns the single active (non-terminal) release for a stack,
	// or (nil, nil) if none exists.
	GetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)

	// GetLatestByStackID returns the highest-sequence release for a stack, any state,
	// or (nil, nil) if the stack has no releases.
	GetLatestByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)

	// GetLatestByStackIDs returns the highest-sequence release for each of the given
	// stack IDs, keyed by stack ID. Stacks with no releases are absent from the map.
	GetLatestByStackIDs(ctx context.Context, stackIDs []string) (map[string]*models.StackRelease, *errors.ServiceError)

	// GetByIDs returns releases keyed by release ID.
	GetByIDs(ctx context.Context, ids []string) (map[string]*models.StackRelease, *errors.ServiceError)

	// CAS state transitions. The bool return indicates whether THIS caller won
	// the compare-and-swap (i.e., the row was in the expected state).

	MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError)
	SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError)
	MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError)
	MarkFailedWithValidationErrors(ctx context.Context, id, message string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError)
	Cancel(ctx context.Context, id string) (bool, *errors.ServiceError)
	MarkCancelled(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)
	MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError)

	// AppendImageDigests merges image digests into the release pins.
	AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError

	// ListTerminalSummariesByStackID returns lightweight release summaries for GC decisions.
	// Only terminal-state releases are returned — active releases are never GC candidates.
	ListTerminalSummariesByStackID(ctx context.Context, stackID string) ([]models.ReleaseSummary, *errors.ServiceError)

	// DeleteByIDs removes releases by their IDs.
	DeleteByIDs(ctx context.Context, ids []string) *errors.ServiceError

	// ListStackIDsExceedingRetention returns stack IDs that have more than `cap`
	// terminal releases — candidates for periodic GC.
	ListStackIDsExceedingRetention(ctx context.Context, cap int) ([]string, *errors.ServiceError)

	AtomicExecutor
}
