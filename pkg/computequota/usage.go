package computequota

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

// UsageStore serializes capacity-changing mutations for one organisation and
// returns its persisted usage. Implementations require a transaction in ctx.
// When replacedStackID is set, the replaced stack remains in StackCount while
// its old resources are omitted from StackResourceCount.
type UsageStore interface {
	LockOrganisationAndGetUsage(ctx context.Context, organisationID, replacedStackID string) (ComputeUsage, *errors.ServiceError)
}
