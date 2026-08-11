package pgstore

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/Stackdome/stackdome/pkg/computeaccess"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// sharedComputeCapacityAdvisoryLockKey is the stable PostgreSQL advisory-lock
// namespace key for shared-compute allocation; its eight bytes spell STACKDOM.
const sharedComputeCapacityAdvisoryLockKey int64 = 0x535441434b444f4d

type ComputeAccessStoreSpec struct {
	SessionFactory db.SessionFactory
	// MaxActiveSharedComputeLeases is platform-wide and fixed for this store.
	MaxActiveSharedComputeLeases int
}

type computeAccessStore struct {
	sessionFactory               db.SessionFactory
	maxActiveSharedComputeLeases int
}

// NewComputeAccessStore fixes the platform capacity ceiling at construction so
// individual activation callers cannot weaken the platform-wide limit.
func NewComputeAccessStore(spec ComputeAccessStoreSpec) computeaccess.Store {
	return &computeAccessStore{
		sessionFactory:               spec.SessionFactory,
		maxActiveSharedComputeLeases: spec.MaxActiveSharedComputeLeases,
	}
}

func (s *computeAccessStore) Activate(ctx context.Context, activation computeaccess.ComputeAccessActivation) (*computeaccess.ComputeAccess, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if s.maxActiveSharedComputeLeases <= 0 {
		return nil, errors.GeneralError("maximum active shared compute leases must be greater than zero")
	}

	// Return an existing grant so retrying the first capacity-consuming request
	// is safe.
	access, err := findCurrentComputeAccessForUpdate(tx, activation.OrganisationID)
	if err == nil {
		return validateComputeAccess(access, activation.StartsAt)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get compute access: %v", err)
	}

	// Serialize first-time grants before checking the platform-wide ceiling.
	if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", sharedComputeCapacityAdvisoryLockKey).Error; err != nil {
		return nil, errors.GeneralError("failed to lock shared compute capacity: %v", err)
	}

	// A concurrent activation may have completed while this transaction waited
	// for the advisory lock.
	access, err = findCurrentComputeAccessForUpdate(tx, activation.OrganisationID)
	if err == nil {
		return validateComputeAccess(access, activation.StartsAt)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get compute access: %v", err)
	}

	// With no current lease, an existing entitlement is a previously consumed
	// grant and must not be issued again.
	entitlementPreviouslyGranted, err := hasComputeEntitlementForUpdate(tx, activation.OrganisationID, activation.EntitlementSource)
	if err != nil {
		return nil, errors.GeneralError("failed to get compute entitlement: %v", err)
	}
	if entitlementPreviouslyGranted {
		return nil, errors.ComputeAccessInactive()
	}

	// Capacity remains reserved until runtime cleanup marks the lease cleaned.
	var activeLeaseCount int64
	if err := tx.Model(&computeaccess.SharedComputeLease{}).
		Where("state <> ?", computeaccess.SharedComputeLeaseStateCleaned).
		Count(&activeLeaseCount).Error; err != nil {
		return nil, errors.GeneralError("failed to count shared compute leases: %v", err)
	}
	if activeLeaseCount >= int64(s.maxActiveSharedComputeLeases) {
		return nil, errors.SharedComputeCapacityReached()
	}

	entitlement := &computeaccess.ComputeEntitlement{
		ID:             uuid.NewString(),
		OrganisationID: activation.OrganisationID,
		Source:         activation.EntitlementSource,
		Status:         computeaccess.ComputeEntitlementStatusActive,
		StartsAt:       activation.StartsAt,
		ExpiresAt:      activation.ExpiresAt,
	}
	if err := tx.Create(entitlement).Error; err != nil {
		return nil, errors.GeneralError("failed to create compute entitlement: %v", err)
	}

	lease := &computeaccess.SharedComputeLease{
		ID:             uuid.NewString(),
		OrganisationID: activation.OrganisationID,
		EntitlementID:  entitlement.ID,
		State:          computeaccess.SharedComputeLeaseStateActive,
		ActivatedAt:    activation.StartsAt,
	}
	if err := tx.Create(lease).Error; err != nil {
		return nil, errors.GeneralError("failed to create shared compute lease: %v", err)
	}

	return &computeaccess.ComputeAccess{Entitlement: entitlement, Lease: lease}, nil
}

func findCurrentComputeAccessForUpdate(tx *gorm.DB, organisationID string) (*computeaccess.ComputeAccess, error) {
	var lease computeaccess.SharedComputeLease
	if err := tx.Clauses(clause.Locking{Strength: rowLockStrengthUpdate}).
		Where("organisation_id = ? AND state <> ?", organisationID, computeaccess.SharedComputeLeaseStateCleaned).
		First(&lease).Error; err != nil {
		return nil, err
	}

	var entitlement computeaccess.ComputeEntitlement
	if err := tx.Clauses(clause.Locking{Strength: rowLockStrengthUpdate}).
		Where("id = ?", lease.EntitlementID).
		First(&entitlement).Error; err != nil {
		return nil, err
	}

	return &computeaccess.ComputeAccess{Entitlement: &entitlement, Lease: &lease}, nil
}

func hasComputeEntitlementForUpdate(tx *gorm.DB, organisationID string, source computeaccess.ComputeEntitlementSource) (bool, error) {
	var entitlement computeaccess.ComputeEntitlement
	err := tx.Clauses(clause.Locking{Strength: rowLockStrengthUpdate}).
		Select("id").
		Where("organisation_id = ? AND source = ?", organisationID, source).
		First(&entitlement).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func validateComputeAccess(access *computeaccess.ComputeAccess, now time.Time) (*computeaccess.ComputeAccess, *errors.ServiceError) {
	if access == nil || access.Entitlement == nil || access.Lease == nil {
		return nil, errors.ComputeAccessInactive()
	}
	entitlement := access.Entitlement
	if entitlement.Status != computeaccess.ComputeEntitlementStatusActive {
		return nil, errors.ComputeAccessInactive()
	}
	if entitlement.StartsAt.After(now) {
		return nil, errors.ComputeAccessInactive()
	}
	if entitlement.ExpiresAt != nil && !entitlement.ExpiresAt.After(now) {
		return nil, errors.ComputeAccessInactive()
	}
	if access.Lease.State != computeaccess.SharedComputeLeaseStateActive {
		return nil, errors.ComputeAccessInactive()
	}
	return access, nil
}
