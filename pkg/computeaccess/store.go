package computeaccess

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
)

// ComputeAccessActivation describes the entitlement an organisation receives.
// Platform capacity is store configuration, not activation input.
type ComputeAccessActivation struct {
	OrganisationID    string
	EntitlementSource ComputeEntitlementSource
	StartsAt          time.Time
	ExpiresAt         *time.Time
}

// Store atomically persists an entitlement and its shared-compute lease with
// the capacity-consuming resource that first activates them.
type Store interface {
	Activate(ctx context.Context, activation ComputeAccessActivation) (*ComputeAccess, *errors.ServiceError)
}
