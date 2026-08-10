package stack

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

type subReconcilerResult struct {
	resultNil          bool
	resultStop         bool
	resultRequeue      bool
	resultRequeueAfter *time.Duration
}

var (
	resultNil  = subReconcilerResult{resultNil: true}
	resultStop = subReconcilerResult{resultStop: true}
)

//go:generate mockgen -source=types.go -destination=types_mock.go -package=stack
type subReconciler interface {
	Reconcile(ctx context.Context, stack *models.Stack) (subReconcilerResult, error)
	Name() string
}

type stackService interface {
	UpdateStackCrRevision(ctx context.Context, ID string, crRevision string) *errors.ServiceError
	InternalList(ctx context.Context, query string, args ...any) ([]*models.Stack, *errors.ServiceError)
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.StackStatus) *errors.ServiceError
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
}

type releaseService interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalGetActiveByStackID(ctx context.Context, stackID string) (*models.StackRelease, *errors.ServiceError)
}

type secretService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
}

type namespaceService interface {
	Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
}

type volumeService interface {
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError
}
