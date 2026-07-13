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

type dbProjectStore struct {
	sessionFactory db.SessionFactory
}

type ProjectStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewProjectStore(spec ProjectStoreSpec) stores.ProjectStore {
	return &dbProjectStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbProjectStore) Create(ctx context.Context, project *models.Project) (*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(project).Error; err != nil {
		return nil, errors.GeneralError("failed to create project: %s", err.Error())
	}
	return s.GetByID(ctx, project.ID)
}

func (s *dbProjectStore) GetByID(ctx context.Context, id string) (*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var project models.Project
	if err := grm.Where("id = ?", id).First(&project).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("project with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch project: %s", err.Error())
	}
	return &project, nil
}

func (s *dbProjectStore) GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var project models.Project
	if err := grm.Where("organisation_id = ? AND name = ?", orgID, name).First(&project).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("project '%s' not found in organisation", name)
		}
		return nil, errors.GeneralError("failed to fetch project: %s", err.Error())
	}
	return &project, nil
}

func (s *dbProjectStore) ListByOrgID(ctx context.Context, orgID string) ([]*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var projects []*models.Project
	if err := grm.Where("organisation_id = ?", orgID).Order("created_at ASC").Find(&projects).Error; err != nil {
		return nil, errors.GeneralError("failed to list projects: %s", err.Error())
	}
	return projects, nil
}

func (s *dbProjectStore) Update(ctx context.Context, id string, project *models.Project) (*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Model(&models.Project{}).Where("id = ?", id).Updates(project).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("project with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update project: %s", err.Error())
	}
	return s.GetByID(ctx, id)
}

func (s *dbProjectStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.Project{}).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("project with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete project: %s", err.Error())
	}
	return nil
}

func (s *dbProjectStore) GetDefaultProjectForOrg(ctx context.Context, orgID string) (*models.Project, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var project models.Project
	if err := grm.Where("organisation_id = ? AND default_project = ?", orgID, true).First(&project).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("default project not found for organisation '%s'", orgID)
		}
		return nil, errors.GeneralError("failed to fetch default project: %s", err.Error())
	}
	return &project, nil
}
