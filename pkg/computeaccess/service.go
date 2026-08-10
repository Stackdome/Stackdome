package computeaccess

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
)

// Service manages the entitlement and shared-compute lease that back an
// organisation's use of managed compute.
type Service interface {
	Activate(ctx context.Context, organisationID string) (*ComputeAccess, *errors.ServiceError)
}

type ServiceSpec struct {
	Store                      Store
	DefaultEntitlementSource   ComputeEntitlementSource
	DefaultEntitlementDuration time.Duration
	Now                        func() time.Time
}

type service struct {
	store                      Store
	defaultEntitlementSource   ComputeEntitlementSource
	defaultEntitlementDuration time.Duration
	now                        func() time.Time
}

// NewService configures the default entitlement issued when managed compute is
// activated for the first time.
func NewService(spec ServiceSpec) Service {
	now := spec.Now
	if now == nil {
		now = time.Now
	}
	return &service{
		store:                      spec.Store,
		defaultEntitlementSource:   spec.DefaultEntitlementSource,
		defaultEntitlementDuration: spec.DefaultEntitlementDuration,
		now:                        now,
	}
}

func (s *service) Activate(ctx context.Context, organisationID string) (*ComputeAccess, *errors.ServiceError) {
	now := s.now()
	var expiresAt *time.Time
	if s.defaultEntitlementDuration > 0 {
		expires := now.Add(s.defaultEntitlementDuration)
		expiresAt = &expires
	}
	return s.store.Activate(ctx, ComputeAccessActivation{
		OrganisationID:    organisationID,
		EntitlementSource: s.defaultEntitlementSource,
		StartsAt:          now,
		ExpiresAt:         expiresAt,
	})
}
