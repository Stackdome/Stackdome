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
// excludeStackID is used only when replacement of a stack's complete desired
// state makes its persisted resources irrelevant.
type UsageStore interface {
	LockOrganisationAndGetUsage(ctx context.Context, organisationID, excludeStackID string) (ComputeUsage, *errors.ServiceError)
}
