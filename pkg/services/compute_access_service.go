package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

// ComputeAccessService manages the entitlement and shared-compute lease that
// back an organisation's use of managed compute. The alpha grants trials;
// future billing or licence grants can reuse the same boundary.
//
//go:generate mockgen -source=compute_access_service.go -destination=compute_access_service_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type ComputeAccessService interface {
	Activate(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError)
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

func (s *computeAccessService) Activate(ctx context.Context, organisationID string) (*models.ComputeAccess, *errors.ServiceError) {
	now := s.now()
	var expiresAt *time.Time
	if s.defaultEntitlementDuration > 0 {
		expires := now.Add(s.defaultEntitlementDuration)
		expiresAt = &expires
	}
	return s.store.Activate(ctx, stores.ComputeAccessActivation{
		OrganisationID:    organisationID,
		EntitlementSource: s.defaultEntitlementSource,
		StartsAt:          now,
		ExpiresAt:         expiresAt,
	})
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
