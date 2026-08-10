package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
)

type ComputeUsage struct {
	StackCount         int64
	StackResourceCount int64
	VolumeCount        int64
	PostgresAddonCount int64
}

type ComputeUsageStore interface {
	// LockOrganisationAndGetUsageWithTx serializes capacity-changing mutations
	// for one organisation and returns persisted usage. excludeStackID is used
	// when a stack's desired resources replace its current resources.
	LockOrganisationAndGetUsageWithTx(ctx context.Context, organisationID, excludeStackID string) (ComputeUsage, *errors.ServiceError)
}
