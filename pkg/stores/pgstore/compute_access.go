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

const postgresDialectName = "postgres"

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
func NewComputeAccessStore(spec ComputeAccessStoreSpec) stores.ComputeAccessStore {
	return &computeAccessStore{
		sessionFactory:               spec.SessionFactory,
		maxActiveSharedComputeLeases: spec.MaxActiveSharedComputeLeases,
	}
}

func (s *computeAccessStore) Activate(ctx context.Context, activation stores.ComputeAccessActivation) (*models.ComputeAccess, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if s.maxActiveSharedComputeLeases <= 0 {
		return nil, errors.GeneralError("maximum active shared compute leases must be greater than zero")
	}

	// Return the existing grant so retrying the first capacity-consuming request is safe.
	access, err := findCurrentComputeAccessForUpdate(tx, activation.OrganisationID)
	if err == nil {
		return validateComputeAccess(access, activation.StartsAt)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get compute access: %s", err.Error())
	}

	if tx.Name() == postgresDialectName {
		// Serialize first-time grants before checking the platform-wide ceiling.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", sharedComputeCapacityAdvisoryLockKey).Error; err != nil {
			return nil, errors.GeneralError("failed to lock shared compute capacity: %s", err.Error())
		}
	}

	access, err = findCurrentComputeAccessForUpdate(tx, activation.OrganisationID)
	if err == nil {
		return validateComputeAccess(access, activation.StartsAt)
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.GeneralError("failed to get compute access: %s", err.Error())
	}

	// With no current lease, an existing entitlement is a previously consumed grant.
	entitlementPreviouslyGranted, err := hasComputeEntitlementForUpdate(tx, activation.OrganisationID, activation.EntitlementSource)
	if err != nil {
		return nil, errors.GeneralError("failed to get compute entitlement: %s", err.Error())
	}
	if entitlementPreviouslyGranted {
		return nil, errors.ComputeAccessInactive()
	}

	// Capacity remains reserved until runtime cleanup marks the lease cleaned.
	var allocated int64
	if err := tx.Model(&models.SharedComputeLease{}).
		Where("state <> ?", models.SharedComputeLeaseStateCleaned).
		Count(&allocated).Error; err != nil {
		return nil, errors.GeneralError("failed to count shared compute leases: %s", err.Error())
	}
	if allocated >= int64(s.maxActiveSharedComputeLeases) {
		return nil, errors.CapacityReached()
	}

	entitlement := &models.ComputeEntitlement{
		ID:             uuid.NewString(),
		OrganisationID: activation.OrganisationID,
		Source:         activation.EntitlementSource,
		Status:         models.ComputeEntitlementStatusActive,
		StartsAt:       activation.StartsAt,
		ExpiresAt:      activation.ExpiresAt,
	}
	if err := tx.Create(entitlement).Error; err != nil {
		return nil, errors.GeneralError("failed to create compute entitlement: %s", err.Error())
	}

	lease := &models.SharedComputeLease{
		ID:             uuid.NewString(),
		OrganisationID: activation.OrganisationID,
		EntitlementID:  entitlement.ID,
		State:          models.SharedComputeLeaseStateActive,
		ActivatedAt:    activation.StartsAt,
	}
	if err := tx.Create(lease).Error; err != nil {
		return nil, errors.GeneralError("failed to create shared compute lease: %s", err.Error())
	}
	return &models.ComputeAccess{Entitlement: entitlement, Lease: lease}, nil
}

func (s *computeAccessStore) HasSharedComputeLease(ctx context.Context, organisationID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.SharedComputeLease{}).
		Where("organisation_id = ?", organisationID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to check shared compute lease: %s", err.Error())
	}
	return count > 0, nil
}

func findCurrentComputeAccessForUpdate(tx *gorm.DB, organisationID string) (*models.ComputeAccess, error) {
	var lease models.SharedComputeLease
	if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("organisation_id = ? AND state <> ?", organisationID, models.SharedComputeLeaseStateCleaned).
		First(&lease).Error; err != nil {
		return nil, err
	}

	var entitlement models.ComputeEntitlement
	if err := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Where("id = ?", lease.EntitlementID).
		First(&entitlement).Error; err != nil {
		return nil, err
	}
	return &models.ComputeAccess{Entitlement: &entitlement, Lease: &lease}, nil
}

func hasComputeEntitlementForUpdate(tx *gorm.DB, organisationID string, source models.ComputeEntitlementSource) (bool, error) {
	query := tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	var entitlement models.ComputeEntitlement
	err := query.Select("id").Where("organisation_id = ? AND source = ?", organisationID, source).First(&entitlement).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func validateComputeAccess(access *models.ComputeAccess, now time.Time) (*models.ComputeAccess, *errors.ServiceError) {
	if access == nil || access.Entitlement == nil || access.Lease == nil {
		return nil, errors.ComputeAccessInactive()
	}
	entitlement := access.Entitlement
	if entitlement.Status != models.ComputeEntitlementStatusActive {
		return nil, errors.ComputeAccessInactive()
	}
	if entitlement.StartsAt.After(now) {
		return nil, errors.ComputeAccessInactive()
	}
	if entitlement.ExpiresAt != nil && !entitlement.ExpiresAt.After(now) {
		return nil, errors.ComputeAccessInactive()
	}
	if access.Lease.State != models.SharedComputeLeaseStateActive {
		return nil, errors.ComputeAccessInactive()
	}
	return access, nil
}
