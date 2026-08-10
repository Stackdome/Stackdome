package worker

import (
	"context"
	"errors"
)

var ErrMutationNotAuthorized = errors.New("workload mutation is no longer authorized")

// MutationAuthorizer revalidates the persisted authority for a cluster write.
// Reconcilers must call it immediately before each Kubernetes mutation.
type MutationAuthorizer func(context.Context) error

// AllowMutation is used for self-hosted reconciliation and resource deletion.
func AllowMutation(context.Context) error {
	return nil
}
