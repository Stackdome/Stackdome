package pgstore

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type addonUsageStore struct {
	sessionFactory db.SessionFactory
}

type AddonUsageStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewAddonUsageStore(spec AddonUsageStoreSpec) stores.AddonUsageStore {
	return &addonUsageStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *addonUsageStore) Create(ctx context.Context, usage *models.AddonUsage) error {
	conn := s.sessionFactory.New(ctx)
	if err := conn.Create(usage).Error; err != nil {
		return fmt.Errorf("failed to create addon usage: %w", err)
	}
	return nil
}

func (s *addonUsageStore) Delete(ctx context.Context, addonType models.AddonType, addonID, stackID, resourceID string) error {
	conn := s.sessionFactory.New(ctx)
	if err := conn.Where(
		"addon_type = ? AND addon_id = ? AND stack_id = ? AND stack_resource_id = ?",
		addonType, addonID, stackID, resourceID,
	).Delete(&models.AddonUsage{}).Error; err != nil {
		return fmt.Errorf("failed to delete addon usage: %w", err)
	}
	return nil
}

func (s *addonUsageStore) GetByAddonID(ctx context.Context, addonType models.AddonType, addonID string) ([]*models.AddonUsage, error) {
	conn := s.sessionFactory.New(ctx)
	var usages []*models.AddonUsage
	if err := conn.Where("addon_type = ? AND addon_id = ?", addonType, addonID).Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to get addon usages by addon ID: %w", err)
	}
	return usages, nil
}

func (s *addonUsageStore) GetByStackID(ctx context.Context, stackID string) ([]*models.AddonUsage, error) {
	conn := s.sessionFactory.New(ctx)
	var usages []*models.AddonUsage
	if err := conn.Where("stack_id = ?", stackID).Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to get addon usages by stack ID: %w", err)
	}
	return usages, nil
}

func (s *addonUsageStore) ExistsByStackResourceAndAddon(ctx context.Context, stackID, resourceID, addonID string) (bool, error) {
	conn := s.sessionFactory.New(ctx)
	var count int64
	if err := conn.Model(&models.AddonUsage{}).Where(
		"stack_id = ? AND stack_resource_id = ? AND addon_id = ?", stackID, resourceID, addonID,
	).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check addon usage existence: %w", err)
	}
	return count > 0, nil
}

func (s *addonUsageStore) IsAddonInUse(ctx context.Context, addonType models.AddonType, addonID string) (bool, error) {
	conn := s.sessionFactory.New(ctx)
	var count int64
	if err := conn.Model(&models.AddonUsage{}).Where(
		"addon_type = ? AND addon_id = ?", addonType, addonID,
	).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check addon usage: %w", err)
	}
	return count > 0, nil
}

func (s *addonUsageStore) DeleteByStackID(ctx context.Context, stackID string) error {
	conn := s.sessionFactory.New(ctx)
	if err := conn.Where("stack_id = ?", stackID).Delete(&models.AddonUsage{}).Error; err != nil {
		return fmt.Errorf("failed to delete addon usages by stack ID: %w", err)
	}
	return nil
}
