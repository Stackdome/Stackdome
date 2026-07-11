package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=stack_preview_config_store.go -destination=../mocks/mock_stack_preview_config_store.go -package=mocks
type StackPreviewConfigStore interface {
	AtomicExecutor
	Create(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.StackPreviewConfig, *errors.ServiceError)
	GetByProjectAndName(ctx context.Context, projectID, name string) (*models.StackPreviewConfig, *errors.ServiceError)
	Update(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	ListByProjectID(ctx context.Context, projectID string, params ListParams) (*PaginatedResult[*models.StackPreviewConfig], *errors.ServiceError)
}
