package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
)

// Interface to run a function with a DB transaction.
type AtomicExecutor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}
