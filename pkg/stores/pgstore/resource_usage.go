package pgstore

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type ResourceUsageStoreSpec struct {
	SessionFactory db.SessionFactory
}

type resourceUsageStore struct {
	sessionFactory db.SessionFactory
}

func NewResourceUsageStore(spec ResourceUsageStoreSpec) stores.ResourceUsageStore {
	return &resourceUsageStore{sessionFactory: spec.SessionFactory}
}

func (s *resourceUsageStore) Create(ctx context.Context, usage *models.ResourceUsage) error {
	if err := s.sessionFactory.New(ctx).Create(usage).Error; err != nil {
		return fmt.Errorf("failed to create resource usage: %w", err)
	}
	return nil
}

func (s *resourceUsageStore) IsResourceInUse(ctx context.Context, resourceType, resourceID string) (bool, error) {
	var count int64
	if err := s.sessionFactory.New(ctx).
		Model(&models.ResourceUsage{}).
		Where("resource_type = ? AND resource_id = ?", resourceType, resourceID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check resource usage: %w", err)
	}
	return count > 0, nil
}

func (s *resourceUsageStore) GetByStackID(ctx context.Context, stackID string) ([]*models.ResourceUsage, error) {
	var usages []*models.ResourceUsage
	if err := s.sessionFactory.New(ctx).
		Where("stack_id = ?", stackID).
		Find(&usages).Error; err != nil {
		return nil, fmt.Errorf("failed to list resource usages: %w", err)
	}
	return usages, nil
}

func (s *resourceUsageStore) DeleteByStackID(ctx context.Context, stackID string) error {
	if err := s.sessionFactory.New(ctx).
		Where("stack_id = ?", stackID).
		Delete(&models.ResourceUsage{}).Error; err != nil {
		return fmt.Errorf("failed to delete resource usages by stack: %w", err)
	}
	return nil
}

func (s *resourceUsageStore) Delete(ctx context.Context, resourceType, resourceID, stackID string) error {
	if err := s.sessionFactory.New(ctx).
		Where("resource_type = ? AND resource_id = ? AND stack_id = ?", resourceType, resourceID, stackID).
		Delete(&models.ResourceUsage{}).Error; err != nil {
		return fmt.Errorf("failed to delete resource usage: %w", err)
	}
	return nil
}
