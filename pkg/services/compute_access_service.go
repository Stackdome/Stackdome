package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// ComputeAccessService hides how access was granted from release and worker
// code. The alpha grants trials; future billing or licence grants can reuse the
// same admission boundary.
//
//go:generate mockgen -source=compute_access_service.go -destination=compute_access_service_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type ComputeAccessService interface {
	ActivateWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError)
	RequireWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError)
	AdmitComputeMutationWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError)
	RequireAccess(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError)
	EnsureNoLease(ctx context.Context, organisationID string) *errors.ServiceError
}

type ComputeAccessServiceSpec struct {
	Store                      stores.ComputeAccessStore
	DefaultEntitlementSource   models.ComputeEntitlementSource
	DefaultEntitlementDuration time.Duration
	Now                        func() time.Time
}

type computeAccessService struct {
	store                      stores.ComputeAccessStore
	defaultEntitlementSource   models.ComputeEntitlementSource
	defaultEntitlementDuration time.Duration
	now                        func() time.Time
}

// NewComputeAccessService configures the default entitlement issued when
// managed compute is activated for the first time.
func NewComputeAccessService(spec ComputeAccessServiceSpec) ComputeAccessService {
	now := spec.Now
	if now == nil {
		now = time.Now
	}
	return &computeAccessService{
		store:                      spec.Store,
		defaultEntitlementSource:   spec.DefaultEntitlementSource,
		defaultEntitlementDuration: spec.DefaultEntitlementDuration,
		now:                        now,
	}
}

func (s *computeAccessService) ActivateWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError) {
	now := s.now()
	var expiresAt *time.Time
	if s.defaultEntitlementDuration > 0 {
		expires := now.Add(s.defaultEntitlementDuration)
		expiresAt = &expires
	}
	return s.store.ActivateWithTx(ctx, stores.ComputeAccessActivation{
		OrganisationID:    organisationID,
		EntitlementSource: s.defaultEntitlementSource,
		StartsAt:          now,
		ExpiresAt:         expiresAt,
	})
}

func (s *computeAccessService) RequireWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError) {
	return s.store.RequireWithTx(ctx, organisationID, s.now())
}

func (s *computeAccessService) AdmitComputeMutationWithTx(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError) {
	return s.store.AdmitComputeMutationWithTx(ctx, organisationID, s.now())
}

func (s *computeAccessService) RequireAccess(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError) {
	return s.store.GetActiveByOrganisationID(ctx, organisationID, s.now())
}

func (s *computeAccessService) EnsureNoLease(ctx context.Context, organisationID string) *errors.ServiceError {
	hasLease, serr := s.store.HasSharedComputeLease(ctx, organisationID)
	if serr != nil {
		return serr
	}
	if hasLease {
		return errors.BadRequest("cannot delete organisation while its shared compute lease exists")
	}
	return nil
}
