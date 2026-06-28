package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

//go:generate mockgen -source=preview_stack_store.go -destination=../mocks/mock_preview_stack_store.go -package=mocks
type PreviewStackStore interface {
	AtomicExecutor
	Create(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.PreviewStack, *errors.ServiceError)
	Update(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError)
	GetByConfigAndPR(ctx context.Context, configID string, prNumber string) (*models.PreviewStack, *errors.ServiceError)
	CountActiveByConfigID(ctx context.Context, configID string) (int64, *errors.ServiceError)
	ListByConfigID(ctx context.Context, configID string, params ListParams) (*PaginatedResult[*models.PreviewStack], *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string, params ListParams) (*PaginatedResult[*models.PreviewStack], *errors.ServiceError)
	ListNeedingReconciliation(ctx context.Context, page, pageSize int) ([]*models.PreviewStack, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
}
