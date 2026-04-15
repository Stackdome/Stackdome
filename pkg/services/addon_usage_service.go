package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type AddonUsageService interface {
	Create(ctx context.Context, usage *models.AddonUsage) error
	Delete(ctx context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error
	GetByAddonID(ctx context.Context, addonType models.AddonType, addonID string) ([]*models.AddonUsage, error)
	GetByStackID(ctx context.Context, stackID string) ([]*models.AddonUsage, error)
	IsAddonInUse(ctx context.Context, addonType models.AddonType, addonID string) (bool, error)
	ExistsByStackResourceAndAddon(ctx context.Context, stackID, resourceID, addonID string) (bool, error)
	DeleteByStackID(ctx context.Context, stackID string) error
}

type AddonUsageServiceSpec struct {
	SessionFactory db.SessionFactory
}

type addonUsageService struct {
	store stores.AddonUsageStore
}

func NewAddonUsageService(spec AddonUsageServiceSpec) AddonUsageService {
	return &addonUsageService{
		store: pgstore.NewAddonUsageStore(pgstore.AddonUsageStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
	}
}

func (s *addonUsageService) Create(ctx context.Context, usage *models.AddonUsage) error {
	return s.store.Create(ctx, usage)
}

func (s *addonUsageService) Delete(ctx context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error {
	return s.store.Delete(ctx, addonType, addonID, stackID, resourceID)
}

func (s *addonUsageService) GetByAddonID(ctx context.Context, addonType models.AddonType, addonID string) ([]*models.AddonUsage, error) {
	return s.store.GetByAddonID(ctx, addonType, addonID)
}

func (s *addonUsageService) GetByStackID(ctx context.Context, stackID string) ([]*models.AddonUsage, error) {
	return s.store.GetByStackID(ctx, stackID)
}

func (s *addonUsageService) IsAddonInUse(ctx context.Context, addonType models.AddonType, addonID string) (bool, error) {
	return s.store.IsAddonInUse(ctx, addonType, addonID)
}

func (s *addonUsageService) ExistsByStackResourceAndAddon(ctx context.Context, stackID, resourceID, addonID string) (bool, error) {
	return s.store.ExistsByStackResourceAndAddon(ctx, stackID, resourceID, addonID)
}

func (s *addonUsageService) DeleteByStackID(ctx context.Context, stackID string) error {
	return s.store.DeleteByStackID(ctx, stackID)
}
