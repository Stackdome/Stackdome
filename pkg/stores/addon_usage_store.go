package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type AddonUsageStore interface {
	Create(ctx context.Context, usage *models.AddonUsage) error
	Delete(ctx context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error
	GetByAddonID(ctx context.Context, addonType models.AddonType, addonID string) ([]*models.AddonUsage, error)
	GetByStackID(ctx context.Context, stackID string) ([]*models.AddonUsage, error)
	IsAddonInUse(ctx context.Context, addonType models.AddonType, addonID string) (bool, error)
	DeleteByStackID(ctx context.Context, stackID string) error
}
