package stack

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
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
	resultRequeue      = subReconcilerResult{resultRequeue: true}
	resultRequeueAfter = func(t time.Duration) subReconcilerResult {
		return subReconcilerResult{resultRequeueAfter: &t}
	}
)

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

type secretService interface {
	InternalGetByID(ctx context.Context, id string) (*models.Secret, *errors.ServiceError)
	CreateSecretUsage(ctx context.Context, secretID string, stackID string) *errors.ServiceError
	GetSecretUsageBySecretIDAndStackID(ctx context.Context, secretID, stackID string) (*models.SecretUsage, *errors.ServiceError)
	GetSecretUsageByStackID(ctx context.Context, stackID string) ([]*models.SecretUsage, *errors.ServiceError)
	DeleteSecretUsage(ctx context.Context, secretID, stackID string) *errors.ServiceError
}

type namespaceService interface {
	Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError)
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
}

type volumeService interface {
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError
}

type postgresAddonService interface {
	InternalGetPostgresAddon(ctx context.Context, id string) (*models.PostgresAddon, *errors.ServiceError)
	InternalGetCredentials(ctx context.Context, addonID string, database string, superuser bool) (*models.PostgresCredentials, *errors.ServiceError)
}

type addonUsageService interface {
	Create(ctx context.Context, usage *models.AddonUsage) error
	Delete(ctx context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error
	GetByStackID(ctx context.Context, stackID string) ([]*models.AddonUsage, error)
	ExistsByStackResourceAndAddon(ctx context.Context, stackID, resourceID, addonID string) (bool, error)
	DeleteByStackID(ctx context.Context, stackID string) error
}
