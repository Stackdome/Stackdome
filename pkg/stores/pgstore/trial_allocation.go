package pgstore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const trialCapacityAdvisoryLockKey int64 = 0x535441434b444f4d

type TrialAllocationStoreSpec struct {
	SessionFactory db.SessionFactory
}

type trialAllocationStore struct {
	sessionFactory db.SessionFactory
}

func NewTrialAllocationStore(spec TrialAllocationStoreSpec) stores.TrialAllocationStore {
	return &trialAllocationStore{sessionFactory: spec.SessionFactory}
}

func (s *trialAllocationStore) AcquireWithTx(ctx context.Context, organisationID string, now, expiresAt time.Time, capacity int) (*models.TrialAllocation, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	existing, err := findTrialAllocationForUpdate(tx, organisationID)
	if err == nil {
		return validateExistingTrialAllocation(existing, now)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get trial allocation: %s", err.Error())
	}

	if tx.Name() == "postgres" {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", trialCapacityAdvisoryLockKey).Error; err != nil {
			return nil, errors.GeneralError("failed to lock trial allocation capacity: %s", err.Error())
		}
	}

	existing, err = findTrialAllocationForUpdate(tx, organisationID)
	if err == nil {
		return validateExistingTrialAllocation(existing, now)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get trial allocation: %s", err.Error())
	}

	var allocated int64
	if err := tx.Model(&models.TrialAllocation{}).
		Where("state <> ?", models.TrialAllocationStateCleaned).
		Count(&allocated).Error; err != nil {
		return nil, errors.GeneralError("failed to count trial allocations: %s", err.Error())
	}
	if allocated >= int64(capacity) {
		return nil, errors.CapacityReached()
	}

	allocation := &models.TrialAllocation{
		ID:             uuid.NewString(),
		OrganisationID: organisationID,
		State:          models.TrialAllocationStateActive,
		ActivatedAt:    now,
		ExpiresAt:      expiresAt,
	}
	if err := tx.Create(allocation).Error; err != nil {
		return nil, errors.GeneralError("failed to create trial allocation: %s", err.Error())
	}
	return allocation, nil
}

func findTrialAllocationForUpdate(tx *gorm.DB, organisationID string) (*models.TrialAllocation, error) {
	query := tx
	if tx.Name() == "postgres" {
		query = query.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	var allocation models.TrialAllocation
	if err := query.Where("organisation_id = ?", organisationID).First(&allocation).Error; err != nil {
		return nil, err
	}
	return &allocation, nil
}

func validateExistingTrialAllocation(allocation *models.TrialAllocation, now time.Time) (*models.TrialAllocation, *errors.ServiceError) {
	if allocation.State == models.TrialAllocationStateActive && allocation.ExpiresAt.After(now) {
		return allocation, nil
	}
	return nil, errors.TrialInactive()
}

func (s *trialAllocationStore) RevalidateWithTx(ctx context.Context, organisationID string, now time.Time) (*models.TrialAllocation, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	allocation, err := findTrialAllocationForUpdate(tx, organisationID)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.TrialInactive()
	}
	if err != nil {
		return nil, errors.GeneralError("failed to get trial allocation: %s", err.Error())
	}
	return validateExistingTrialAllocation(allocation, now)
}

func (s *trialAllocationStore) GetActiveByOrganisationID(ctx context.Context, organisationID string, now time.Time) (*models.TrialAllocation, *errors.ServiceError) {
	allocation, serr := s.GetByOrganisationID(ctx, organisationID)
	if serr != nil {
		if serr.Is404() {
			return nil, errors.TrialInactive()
		}
		return nil, serr
	}
	return validateExistingTrialAllocation(allocation, now)
}

func (s *trialAllocationStore) GetByOrganisationID(ctx context.Context, organisationID string) (*models.TrialAllocation, *errors.ServiceError) {
	var allocation models.TrialAllocation
	if err := s.sessionFactory.New(ctx).Where("organisation_id = ?", organisationID).First(&allocation).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("trial allocation not found")
		}
		return nil, errors.GeneralError("failed to get trial allocation: %s", err.Error())
	}
	return &allocation, nil
}
