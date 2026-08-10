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
// so a failed provisioning request cannot consume capacity without its
// database resource being created.
//
//go:generate mockgen -source=compute_access_store.go -destination=../mocks/mock_compute_access_store.go -package=mocks
type ComputeAccessStore interface {
	ActivateWithTx(ctx context.Context, activation ComputeAccessActivation) (*models.ComputeAccess, *errors.ServiceError)
	RequireWithTx(ctx context.Context, organisationID string, now time.Time) (*models.ComputeAccess, *errors.ServiceError)
	AdmitComputeMutationWithTx(ctx context.Context, organisationID string, now time.Time) (*models.ComputeAccess, *errors.ServiceError)
	HasSharedComputeLease(ctx context.Context, organisationID string) (bool, *errors.ServiceError)
}
