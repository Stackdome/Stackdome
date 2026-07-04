package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
)

// Interface to run a function with a DB transaction.
//
//go:generate mockgen -source=common.go -destination=../mocks/mock_atomic_executor.go -package=mocks
type AtomicExecutor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}
