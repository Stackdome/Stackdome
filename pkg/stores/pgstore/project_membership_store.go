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

type dbProjectMembershipStore struct {
	sessionFactory db.SessionFactory
}

type ProjectMembershipStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewProjectMembershipStore(spec ProjectMembershipStoreSpec) stores.ProjectMembershipStore {
	return &dbProjectMembershipStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbProjectMembershipStore) Create(ctx context.Context, membership *models.ProjectMembership) (*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(membership).Error; err != nil {
		return nil, errors.GeneralError("failed to create project membership: %s", err.Error())
	}
	return s.GetByID(ctx, membership.ID)
}

func (s *dbProjectMembershipStore) GetByID(ctx context.Context, id string) (*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.ProjectMembership
	if err := grm.Preload("Project").Preload("User").Where("id = ?", id).First(&membership).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("project membership with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch project membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbProjectMembershipStore) GetByProjectAndUser(ctx context.Context, projectID, userID string) (*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.ProjectMembership
	if err := grm.Preload("Project").Preload("User").Where("project_id = ? AND user_id = ?", projectID, userID).First(&membership).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("membership not found for user in project")
		}
		return nil, errors.GeneralError("failed to fetch project membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbProjectMembershipStore) ListByProjectID(ctx context.Context, projectID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.ProjectMembership
	if err := grm.Preload("User").Preload("Project").Where("project_id = ?", projectID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list project memberships: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbProjectMembershipStore) ListByUserID(ctx context.Context, userID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.ProjectMembership
	if err := grm.Preload("Project").Preload("User").Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbProjectMembershipStore) ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.ProjectMembership
	if err := grm.Preload("Project").Preload("User").
		Joins("JOIN projects ON projects.id = project_memberships.project_id").
		Where("project_memberships.user_id = ? AND projects.organisation_id = ?", userID, orgID).
		Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user in org: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbProjectMembershipStore) Update(ctx context.Context, id string, membership *models.ProjectMembership) (*models.ProjectMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	result := grm.Model(&models.ProjectMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"role": membership.Role,
	})
	if result.Error != nil {
		return nil, errors.GeneralError("failed to update project membership: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, errors.NotFound("project membership with id '%s' not found", id)
	}
	return s.GetByID(ctx, id)
}

func (s *dbProjectMembershipStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	result := grm.Where("id = ?", id).Delete(&models.ProjectMembership{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete project membership: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("project membership with id '%s' not found", id)
	}
	return nil
}
