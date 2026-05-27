package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type ResourceUsageStore interface {
	Create(ctx context.Context, usage *models.ResourceUsage) error
	IsResourceInUse(ctx context.Context, resourceType, resourceID string) (bool, error)
	GetByStackID(ctx context.Context, stackID string) ([]*models.ResourceUsage, error)
	DeleteByStackID(ctx context.Context, stackID string) error
	Delete(ctx context.Context, resourceType, resourceID, stackID string) error
}
