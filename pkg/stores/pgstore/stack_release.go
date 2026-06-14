package pgstore

import (
	"context"
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

// CreateSuperseding atomically supersedes Pending/Rendering releases, assigns the
// next sequence number, and inserts the new release.
func (s *stackReleaseStore) CreateSuperseding(ctx context.Context, release *models.StackRelease) (*models.StackRelease, *errors.ServiceError) {
	var result *models.StackRelease

	if err := s.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		tx := db.TxFromContext(txCtx)

		// Supersede any Pending or Rendering releases for this stack.
		// Never supersede Applying — that release owns the cluster.
		now := time.Now().UTC()
		if err := tx.Model(&models.StackRelease{}).
			Where("stack_id = ? AND state IN ?", release.StackID, []models.StackReleaseState{
				models.ReleaseStatePending,
				models.ReleaseStateRendering,
			}).
			Updates(map[string]interface{}{
				"state":        models.ReleaseStateSuperseded,
				"message":      "superseded by newer release",
				"updated_at":   now,
				"completed_at": now,
			}).Error; err != nil {
			return errors.GeneralError("failed to supersede existing releases: %s", err.Error())
		}

		// Get the next sequence number.
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
func (s *stackReleaseStore) ListByStackID(ctx context.Context, stackID string) ([]*models.StackRelease, *errors.ServiceError) {
	var releases []*models.StackRelease
	if err := s.sessionFactory.New(ctx).
		Select("id, stack_id, sequence, state, message, cause, snapshot_revision, manifest_revision, pins, renderer_version, created_by, created_at, updated_at, rendered_at, completed_at").
		Where("stack_id = ?", stackID).
		Order("sequence DESC").
		Find(&releases).Error; err != nil {
		return nil, errors.GeneralError("failed to list releases: %s", err.Error())
	}
	return releases, nil
}

// ListActive returns all non-terminal releases across all stacks.
func (s *stackReleaseStore) ListActive(ctx context.Context) ([]*models.StackRelease, *errors.ServiceError) {
	var releases []*models.StackRelease
	activeStates := []models.StackReleaseState{
		models.ReleaseStatePending,
		models.ReleaseStateRendering,
		models.ReleaseStateApplying,
	}
	if err := s.sessionFactory.New(ctx).
		Where("state IN ?", activeStates).
		Order("created_at ASC").
		Find(&releases).Error; err != nil {
		return nil, errors.GeneralError("failed to list active releases: %s", err.Error())
	}
	return releases, nil
}

// GetActiveByStackID returns the single active release for a stack, or (nil, nil) if none.
func (s *stackReleaseStore) GetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError) {
	var release models.StackRelease
	activeStates := []models.StackReleaseState{
		models.ReleaseStatePending,
		models.ReleaseStateRendering,
		models.ReleaseStateApplying,
	}
	if err := s.sessionFactory.New(ctx).
		Where("stack_id = ? AND state IN ?", stackID, activeStates).
		Order("sequence DESC").
		First(&release).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, errors.GeneralError("failed to get active release: %s", err.Error())
	}
	return &release, nil
}

// MarkRendering transitions Pending -> Rendering via conditional UPDATE.
func (s *stackReleaseStore) MarkRendering(ctx context.Context, id string) (bool, *errors.ServiceError) {
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ? AND manifest IS NULL", id, models.ReleaseStatePending).
		Updates(map[string]interface{}{
			"state":      models.ReleaseStateRendering,
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark rendering: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// MarkApplyingDirect transitions Pending -> Applying (for releases that skip rendering).
func (s *stackReleaseStore) MarkApplyingDirect(ctx context.Context, id string) (bool, *errors.ServiceError) {
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ? AND manifest IS NOT NULL", id, models.ReleaseStatePending).
		Updates(map[string]interface{}{
			"state":      models.ReleaseStateApplying,
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark applying: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// SaveManifest writes the rendered manifest and transitions Rendering -> Applying.
func (s *stackReleaseStore) SaveManifest(ctx context.Context, id string, m *models.ReleaseManifest, rev string, pins models.ReleasePins, rendererVersion string) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ? AND manifest IS NULL", id, models.ReleaseStateRendering).
		Updates(map[string]interface{}{
			"state":             models.ReleaseStateApplying,
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

// MarkReleased transitions Applying -> Released.
func (s *stackReleaseStore) MarkReleased(ctx context.Context, id string, outcome models.ReleaseOutcome) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state = ?", id, models.ReleaseStateApplying).
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

// MarkFailed transitions any active state -> Failed.
func (s *stackReleaseStore) MarkFailed(ctx context.Context, id string, message string, outcome *models.ReleaseOutcome) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	activeStates := []models.StackReleaseState{
		models.ReleaseStatePending,
		models.ReleaseStateRendering,
		models.ReleaseStateApplying,
	}

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
		Where("id = ? AND state IN ?", id, activeStates).
		Updates(updates)

	if result.Error != nil {
		return false, errors.GeneralError("failed to mark failed: %s", result.Error.Error())
	}
	return result.RowsAffected > 0, nil
}

// Cancel transitions Pending or Rendering -> Cancelled.
func (s *stackReleaseStore) Cancel(ctx context.Context, id string) (bool, *errors.ServiceError) {
	now := time.Now().UTC()
	cancellableStates := []models.StackReleaseState{
		models.ReleaseStatePending,
		models.ReleaseStateRendering,
	}

	result := s.sessionFactory.New(ctx).
		Model(&models.StackRelease{}).
		Where("id = ? AND state IN ?", id, cancellableStates).
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
