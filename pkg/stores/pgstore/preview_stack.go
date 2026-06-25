package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type PreviewStackStoreSpec struct {
	SessionFactory db.SessionFactory
}

type previewStackStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewPreviewStackStore(spec PreviewStackStoreSpec) stores.PreviewStackStore {
	return &previewStackStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *previewStackStore) Create(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Create(preview).Error; err != nil {
		return nil, errors.GeneralError("failed to create preview stack: %s", err.Error())
	}
	return s.GetByID(ctx, preview.ID)
}

func (s *previewStackStore) GetByID(ctx context.Context, id string) (*models.PreviewStack, *errors.ServiceError) {
	var preview models.PreviewStack
	if err := s.sessionFactory.New(ctx).First(&preview, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("preview stack with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get preview stack: %s", err.Error())
	}
	return &preview, nil
}

func (s *previewStackStore) Update(ctx context.Context, preview *models.PreviewStack) (*models.PreviewStack, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Model(&models.PreviewStack{}).Where("id = ?", preview.ID).Updates(preview).Error; err != nil {
		return nil, errors.GeneralError("failed to update preview stack: %s", err.Error())
	}
	return s.GetByID(ctx, preview.ID)
}

func (s *previewStackStore) GetByConfigAndPR(ctx context.Context, configID string, prNumber string) (*models.PreviewStack, *errors.ServiceError) {
	var preview models.PreviewStack
	if err := s.sessionFactory.New(ctx).
		Where("stack_preview_config_id = ? AND pr_number = ?", configID, prNumber).
		First(&preview).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("preview stack for config %s and PR #%s not found", configID, prNumber)
		}
		return nil, errors.GeneralError("failed to get preview stack: %s", err.Error())
	}
	return &preview, nil
}

func (s *previewStackStore) CountActiveByConfigID(ctx context.Context, configID string) (int64, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).
		Model(&models.PreviewStack{}).
		Where("stack_preview_config_id = ?", configID).
		Count(&count).Error; err != nil {
		return 0, errors.GeneralError("failed to count active preview stacks: %s", err.Error())
	}
	return count, nil
}

func (s *previewStackStore) ListByConfigID(ctx context.Context, configID string, params stores.ListParams) (*stores.PaginatedResult[*models.PreviewStack], *errors.ServiceError) {
	params = params.WithDefaultOrder("created_at DESC")

	var total int64
	countQuery := s.sessionFactory.New(ctx).Model(&models.PreviewStack{}).
		Where("stack_preview_config_id = ?", configID)
	if err := params.ApplyFiltersOnly(countQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count preview stacks: %s", err.Error())
	}

	var previews []*models.PreviewStack
	dataQuery := s.sessionFactory.New(ctx).
		Where("stack_preview_config_id = ?", configID)
	if err := params.Apply(dataQuery).Find(&previews).Error; err != nil {
		return nil, errors.GeneralError("failed to list preview stacks: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.PreviewStack]{
		Items:      previews,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}

func (s *previewStackStore) ListByTeamID(ctx context.Context, teamID string, params stores.ListParams) (*stores.PaginatedResult[*models.PreviewStack], *errors.ServiceError) {
	params = params.WithDefaultOrder("created_at DESC")

	var total int64
	countQuery := s.sessionFactory.New(ctx).Model(&models.PreviewStack{}).
		Where("team_id = ?", teamID)
	if err := params.ApplyFiltersOnly(countQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count preview stacks: %s", err.Error())
	}

	var previews []*models.PreviewStack
	dataQuery := s.sessionFactory.New(ctx).
		Where("team_id = ?", teamID)
	if err := params.Apply(dataQuery).Find(&previews).Error; err != nil {
		return nil, errors.GeneralError("failed to list preview stacks: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.PreviewStack]{
		Items:      previews,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}

func (s *previewStackStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Where("id = ?", id).Delete(&models.PreviewStack{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete preview stack: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("preview stack with id %s not found", id)
	}
	return nil
}
