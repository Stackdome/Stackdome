package pgstore

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type StackPreviewConfigStoreSpec struct {
	SessionFactory db.SessionFactory
}

type stackPreviewConfigStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewStackPreviewConfigStore(spec StackPreviewConfigStoreSpec) stores.StackPreviewConfigStore {
	return &stackPreviewConfigStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *stackPreviewConfigStore) Create(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Create(config).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("preview config with name '%s' already exists in this team", config.Name)
		}
		return nil, errors.GeneralError("failed to create preview config: %s", err.Error())
	}
	return s.GetByID(ctx, config.ID)
}

func (s *stackPreviewConfigStore) GetByID(ctx context.Context, id string) (*models.StackPreviewConfig, *errors.ServiceError) {
	var config models.StackPreviewConfig
	if err := s.sessionFactory.New(ctx).First(&config, "id = ?", id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("preview config with id %s not found", id)
		}
		return nil, errors.GeneralError("failed to get preview config: %s", err.Error())
	}
	return &config, nil
}

func (s *stackPreviewConfigStore) GetByTeamAndName(ctx context.Context, teamID, name string) (*models.StackPreviewConfig, *errors.ServiceError) {
	var config models.StackPreviewConfig
	if err := s.sessionFactory.New(ctx).First(&config, "team_id = ? AND name = ?", teamID, name).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("preview config with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to get preview config: %s", err.Error())
	}
	return &config, nil
}

func (s *stackPreviewConfigStore) Update(ctx context.Context, config *models.StackPreviewConfig) (*models.StackPreviewConfig, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Model(&models.StackPreviewConfig{}).Where("id = ?", config.ID).Select("*").Updates(config).Error; err != nil {
		return nil, errors.GeneralError("failed to update preview config: %s", err.Error())
	}
	return s.GetByID(ctx, config.ID)
}

func (s *stackPreviewConfigStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Where("id = ?", id).Delete(&models.StackPreviewConfig{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete preview config: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("preview config with id %s not found", id)
	}
	return nil
}

func (s *stackPreviewConfigStore) ListByTeamID(ctx context.Context, teamID string, params stores.ListParams) (*stores.PaginatedResult[*models.StackPreviewConfig], *errors.ServiceError) {
	params = params.WithDefaultOrder("created_at DESC")

	var total int64
	countQuery := s.sessionFactory.New(ctx).Model(&models.StackPreviewConfig{}).Where("team_id = ?", teamID)
	if err := params.ApplyFiltersOnly(countQuery).Count(&total).Error; err != nil {
		return nil, errors.GeneralError("failed to count preview configs: %s", err.Error())
	}

	var configs []*models.StackPreviewConfig
	dataQuery := s.sessionFactory.New(ctx).Where("team_id = ?", teamID)
	if err := params.Apply(dataQuery).Find(&configs).Error; err != nil {
		return nil, errors.GeneralError("failed to list preview configs: %s", err.Error())
	}

	return &stores.PaginatedResult[*models.StackPreviewConfig]{
		Items:      configs,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.Limit(),
		TotalPages: params.TotalPages(total),
	}, nil
}
