package pgstore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type StackReleaseStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackReleaseStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewStackReleaseStore(spec StackReleaseStoreSpec) stores.StackReleaseStore {
	return &stackReleaseStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

var activeReleaseStates = []models.StackReleaseState{
	models.ReleaseStatePending,
	models.ReleaseStateInProgress,
}

// Create assigns the next sequence number and inserts the new release.
// Supersession of older active releases is handled by the gatekeeper reconciler.
func (s *stackReleaseStore) Create(ctx context.Context, release *models.StackRelease) (*models.StackRelease, *errors.ServiceError) {
	var result *models.StackRelease

	if err := s.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		tx := db.TxFromContext(txCtx)

		var maxSeq *int
		if err := tx.Model(&models.StackRelease{}).
			Where("stack_id = ?", release.StackID).
			Select("MAX(sequence)").
			Scan(&maxSeq).Error; err != nil {
			return errors.GeneralError("failed to get max sequence: %s", err.Error())
		}

		if maxSeq != nil {
			release.Sequence = *maxSeq + 1
		} else {
			release.Sequence = 1
		}

		if err := tx.Create(release).Error; err != nil {
			if stderrors.Is(err, gorm.ErrDuplicatedKey) {
				return errors.Conflict("a concurrent deploy is already in progress for this stack, please retry")
			}
			return errors.GeneralError("failed to create release: %s", err.Error())
		}

		result = release
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *stackReleaseStore) GetByID(ctx context.Context, id string) (*models.StackRelease, *errors.ServiceError) {
	var release models.StackRelease
	if err := s.sessionFactory.New(ctx).First(&release, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("release with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get release: %s", err.Error())
	}
	return &release, nil
}

// ListByStackID returns releases for a stack, excluding heavy JSONB columns for performance.
func (s *stackReleaseStore) ListByStackID(ctx context.Context, stackID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackRelease], *errors.ServiceError) {
	params = params.WithDefaultOrder("sequence DESC")
	selectCols := "id, stack_id, sequence, state, message, cause, snapshot_revision, manifest_revision, pins, renderer_version, created_by, created_at, updated_at, rendered_at, completed_at"

	var total int64
	countQuery := s.sessionFactory.New(ctx).Model(&models.StackRelease{}).Where("stack_id = ?", stackID)
	if err := params.ApplyFiltersOnly(countQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count releases: %s", err.Error())
	}

	var releases []*models.StackRelease
	dataQuery := s.sessionFactory.New(ctx).Select(selectCols).Where("stack_id = ?", stackID)
	if err := params.Apply(dataQuery).Find(&releases).Error; err != nil {
		return nil, errors.GeneralError("failed to list releases: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.StackRelease]{
		Items:      releases,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}

// ListActive returns all non-terminal releases across all stacks.
func (s *stackReleaseStore) ListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError) {
	var releases []*models.StackRelease
	if err := s.sessionFactory.New(ctx).
		Where("state IN ?", activeReleaseStates).
		Order("created_at ASC").
		Find(&releases).Error; err != nil {
		return nil, errors.GeneralError("failed to list active releases: %s", err.Error())
	}
	return releases, nil
}

// GetActiveByStackID returns the highest-sequence active release for a stack, or (nil, nil) if none.
func (s *stackReleaseStore) GetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	var release models.StackRelease
	if err := s.sessionFactory.New(ctx).
		Where("stack_id = ? AND state IN ?", stackID, activeReleaseStates).
		Order("sequence DESC").
		First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.GeneralError("failed to get active release: %s", err.Error())
	}
	return &release, nil
}

// MarkInProgress transitions Pending -> InProgress via conditional UPDATE.
func (s *stackReleaseStore) MarkInProgress(ctx context.Context, id string) (bool, *errors.ServiceError) {
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStatePending).
		Updates(map[string]interface{}{
			"state":      models.ReleaseStateInProgress,
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark in progress: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// SaveManifest writes the rendered manifest while the release is InProgress and manifest is unset.
func (s *stackReleaseStore) SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ? AND manifest IS NULL", id, models.ReleaseStateInProgress).
		Updates(map[string]interface{}{
			"manifest":          m,
			"manifest_revision": rev,
			"pins":              pins,
			"renderer_version":  rendererVersion,
			"rendered_at":       now,
			"updated_at":        now,
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to save manifest: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

func (s *stackReleaseStore) MarkCancelled(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStatePending).
		Updates(map[string]interface{}{
			"state":      models.ReleaseStateCancelled,
			"message":    reason,
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return false, errors.GeneralError("failed to mark cancelled: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// MarkSuperseded transitions any active state -> Superseded (worker self-superseding).
func (s *stackReleaseStore) MarkSuperseded(ctx context.Context, id string, reason string) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state IN ?", id, activeReleaseStates).
		Updates(map[string]interface{}{
			"state":        models.ReleaseStateSuperseded,
			"message":      reason,
			"completed_at": now,
			"updated_at":   now,
		})
	if result.Error != nil {
		return false, errors.GeneralError("failed to mark superseded: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// MarkReleased transitions InProgress -> Released.
func (s *stackReleaseStore) MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStateInProgress).
		Updates(map[string]interface{}{
			"state":        models.ReleaseStateReleased,
			"outcome":      outcome,
			"completed_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark released: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// MarkFailed transitions InProgress -> Failed.
func (s *stackReleaseStore) MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError) {
	now := time.Now().UTC()

	updates := map[string]interface{}{
		"state":        models.ReleaseStateFailed,
		"message":      message,
		"completed_at": now,
		"updated_at":   now,
	}
	if outcome != nil {
		updates["outcome"] = outcome
	}

	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStateInProgress).
		Updates(updates)

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark failed: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// Cancel transitions Pending -> Cancelled (user-initiated).
func (s *stackReleaseStore) Cancel(ctx context.Context, id string) (bool, *errors.ServiceError) {
	now := time.Now().UTC()

	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStatePending).
		Updates(map[string]interface{}{
			"state":        models.ReleaseStateCancelled,
			"completed_at": now,
			"updated_at":   now,
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to cancel release: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// ListTerminalSummariesByStackID returns lightweight release summaries for GC decisions.
// Only terminal-state releases are returned — active releases are never GC candidates.
func (s *stackReleaseStore) ListTerminalSummariesByStackID(ctx context.Context, stackID string) ([]models.ReleaseSummary, *errors.ServiceError) {
	var out []models.ReleaseSummary
	if err := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Select("id", "sequence", "state").
		Where("stack_id = ? AND state IN ?", stackID, models.TerminalStackReleaseStates()).
		Order("sequence DESC").
		Scan(&out).Error; err != nil {
		return nil, errors.GeneralError("failed to list terminal release summaries: %s", err.Error())
	}
	return out, nil
}

// DeleteByIDs removes releases by their IDs.
func (s *stackReleaseStore) DeleteByIDs(ctx context.Context, ids []string) *errors.ServiceError {
	if len(ids) == 0 {
		return nil
	}
	if err := s.sessionFactory.New(ctx).
		Where("id IN ?", ids).
		Delete(&models.StackRelease{}).Error; err != nil {
		return errors.GeneralError("failed to delete releases: %s", err.Error())
	}
	return nil
}

func (s *stackReleaseStore) ListStackIDsExceedingRetention(ctx context.Context, cap int) ([]string, *errors.ServiceError) {
	var stackIDs []string
	if err := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Select("stack_id").
		Where("state IN ?", models.TerminalStackReleaseStates()).
		Group("stack_id").
		Having("COUNT(*) > ?", cap).
		Pluck("stack_id", &stackIDs).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks exceeding retention: %s", err.Error())
	}
	return stackIDs, nil
}

// AppendImageDigests merges image digests into the release's pins.
func (s *stackReleaseStore) AppendImageDigests(ctx context.Context, id string, digests map[string]string) *errors.ServiceError {
	release, svcErr := s.GetByID(ctx, id)
	if svcErr != nil {
		return svcErr
	}

	if release.Pins.Resources == nil {
		release.Pins.Resources = make(map[string]models.ResourcePins)
	}

	for resourceName, digest := range digests {
		pin := release.Pins.Resources[resourceName]
		if pin.ImageDigest == "" {
			pin.ImageDigest = digest
			release.Pins.Resources[resourceName] = pin
		}
	}

	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"pins":       release.Pins,
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return errors.GeneralError("failed to append image digests: %s", result.Error.Error())
	}
	return nil
}
