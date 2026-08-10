package stores

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

// ComputeAccessActivation describes the entitlement an organisation receives.
// Platform capacity is deliberately store configuration, not caller input.
type ComputeAccessActivation struct {
	OrganisationID    string
	EntitlementSource models.ComputeEntitlementSource
	StartsAt          time.Time
	ExpiresAt         *time.Time
}

// ComputeAccessStore atomically persists entitlements and shared-compute leases
// with the capacity-consuming resource that first activates them.
//
//go:generate mockgen -source=compute_access_store.go -destination=../mocks/mock_compute_access_store.go -package=mocks
type ComputeAccessStore interface {
	Activate(ctx context.Context, activation ComputeAccessActivation) (*models.ComputeAccess, *errors.ServiceError)
	HasSharedComputeLease(ctx context.Context, organisationID string) (bool, *errors.ServiceError)
}
