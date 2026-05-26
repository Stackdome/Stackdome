package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

//go:generate mockgen -source=resource_usage_service.go -destination=../mocks/mock_resource_usage_service.go -package=mocks
type ResourceUsageService interface {
	Create(ctx context.Context, usage *models.ResourceUsage) error
	IsResourceInUse(ctx context.Context, resourceType, resourceID string) (bool, error)
	GetByStackID(ctx context.Context, stackID string) ([]*models.ResourceUsage, error)
	DeleteByStackID(ctx context.Context, stackID string) error
	Delete(ctx context.Context, resourceType, resourceID, stackID string) error
}

type ResourceUsageServiceSpec struct {
	SessionFactory db.SessionFactory
}

type resourceUsageService struct {
	store stores.ResourceUsageStore
}

func NewResourceUsageService(spec ResourceUsageServiceSpec) ResourceUsageService {
	return &resourceUsageService{
		store: pgstore.NewResourceUsageStore(pgstore.ResourceUsageStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
	}
}

func (s *resourceUsageService) Create(ctx context.Context, usage *models.ResourceUsage) error {
	return s.store.Create(ctx, usage)
}

func (s *resourceUsageService) IsResourceInUse(ctx context.Context, resourceType, resourceID string) (bool, error) {
	return s.store.IsResourceInUse(ctx, resourceType, resourceID)
}

func (s *resourceUsageService) GetByStackID(ctx context.Context, stackID string) ([]*models.ResourceUsage, error) {
	return s.store.GetByStackID(ctx, stackID)
}

func (s *resourceUsageService) DeleteByStackID(ctx context.Context, stackID string) error {
	return s.store.DeleteByStackID(ctx, stackID)
}

func (s *resourceUsageService) Delete(ctx context.Context, resourceType, resourceID, stackID string) error {
	return s.store.Delete(ctx, resourceType, resourceID, stackID)
}
