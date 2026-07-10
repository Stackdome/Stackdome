//go:generate mockgen -source=types.go -destination=types_mock_test.go -package=preview
package preview

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
	resultNil          = subReconcilerResult{resultNil: true}
	resultStop         = subReconcilerResult{resultStop: true}
	resultRequeueAfter = func(t time.Duration) subReconcilerResult {
		return subReconcilerResult{resultRequeueAfter: &t}
	}
)

type subReconciler interface {
	Reconcile(ctx context.Context, preview *models.PreviewStack) (subReconcilerResult, error)
	Name() string
}

type previewStackStore interface {
	GetByID(ctx context.Context, id string) (*models.PreviewStack, *errors.ServiceError)
	Update(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	ListNeedingReconciliation(ctx context.Context, page, pageSize int) ([]*models.PreviewStack, *errors.ServiceError)
	WithTransaction(ctx context.Context, fn func(ctx context.Context) *errors.ServiceError) *errors.ServiceError
}

type previewStackService interface {
	InternalFetchStackfile(ctx context.Context, config *models.StackPreviewConfig, commitSHA string) ([]byte, string, *errors.OperationError)
	InternalBuildStackFromContent(ctx context.Context, config *models.StackPreviewConfig, preview *models.PreviewStack, content []byte) (*models.Stack, *errors.OperationError)
}

type configStore interface {
	GetByID(ctx context.Context, id string) (*models.StackPreviewConfig, *errors.ServiceError)
}

type releaseService interface {
	InternalGet(ctx context.Context, releaseID string) (*models.StackRelease, *errors.ServiceError)
	InternalCreateRelease(ctx context.Context, stackID string, cause models.ReleaseCause) (*models.StackRelease, *errors.ServiceError)
}

type stackService interface {
	InternalGetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	InternalCreateStack(ctx context.Context, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	InternalUpdateStack(ctx context.Context, ID string, spec *models.Stack) (*models.Stack, *errors.ServiceError)
	InternalDeleteStack(ctx context.Context, stack *models.Stack) (*models.Stack, *errors.ServiceError)
}
