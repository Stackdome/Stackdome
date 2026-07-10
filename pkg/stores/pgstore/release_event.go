package pgstore

import (
	stderrors "errors"

	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReleaseEventStoreSpec struct {
	SessionFactory db.SessionFactory
}

type releaseEventStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewReleaseEventStore(spec ReleaseEventStoreSpec) stores.ReleaseEventStore {
	return &releaseEventStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *releaseEventStore) Insert(ctx context.Context, event *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError) {
	var created *models.ReleaseEvent
	if serr := s.WithTransaction(ctx, func(txCtx context.Context) *errors.ServiceError {
		var e *errors.ServiceError
		created, e = s.InsertWithTx(txCtx, event)
		return e
	}); serr != nil {
		return nil, serr
	}
	return created, nil
}

func (s *releaseEventStore) InsertWithTx(ctx context.Context, event *models.ReleaseEvent) (*models.ReleaseEvent, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	// Lock the release row before computing MAX(sequence): two concurrent
	// inserts for the same release would otherwise read the same max and
	// collide on the (release_id, sequence) unique index. The lock also
	// guarantees the parent release still exists for the FK insert below.
	var release models.StackRelease
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").
		First(&release, "id = ?", event.ReleaseID).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("release %s not found", event.ReleaseID)
		}
		return nil, errors.GeneralError("failed to lock release for event insert: %v", err)
	}

	var maxSeq int
	if err := tx.Model(&models.ReleaseEvent{}).
		Where("release_id = ?", event.ReleaseID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&maxSeq).Error; err != nil {
		return nil, errors.GeneralError("failed to compute next event sequence: %v", err)
	}
	event.Sequence = maxSeq + 1

	// Use ON CONFLICT DO NOTHING instead of letting a duplicate-key error
	// abort the INSERT: in Postgres a failed statement poisons the
	// surrounding transaction, which would break InsertWithTx callers who
	// keep using their transaction after a dedupe no-op.
	res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(event)
	if res.Error != nil {
		return nil, errors.GeneralError("failed to insert release event: %v", res.Error)
	}
	if res.RowsAffected == 0 {
		// Parent row is locked, so a conflict can only be the dedupe key.
		return nil, nil
	}
	return event, nil
}

func (s *releaseEventStore) ListByReleaseID(ctx context.Context, releaseID string, afterSequence int, limit int) ([]*models.ReleaseEvent, *errors.ServiceError) {
	var events []*models.ReleaseEvent
	if err := s.sessionFactory.New(ctx).
		Where("release_id = ? AND sequence > ?", releaseID, afterSequence).
		Order("sequence ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		return nil, errors.GeneralError("failed to list release events: %v", err)
	}
	return events, nil
}
