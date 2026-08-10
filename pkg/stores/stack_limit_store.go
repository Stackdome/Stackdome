package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
)

type StackUsage struct {
	StackCount         int64
	StackResourceCount int64
}

type StackLimitStore interface {
	// LockOrganisationAndGetUsageWithTx serializes limit-changing mutations for
	// one organisation and returns persisted usage. excludeStackID is used by a
	// whole-stack replacement, which supplies that stack's desired count itself.
	LockOrganisationAndGetUsageWithTx(ctx context.Context, organisationID, excludeStackID string) (StackUsage, *errors.ServiceError)
}
