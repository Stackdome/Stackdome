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

type dbTeamMembershipStore struct {
	sessionFactory db.SessionFactory
}

type TeamMembershipStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewTeamMembershipStore(spec TeamMembershipStoreSpec) stores.TeamMembershipStore {
	return &dbTeamMembershipStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbTeamMembershipStore) Create(ctx context.Context, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(membership).Error; err != nil {
		return nil, errors.GeneralError("failed to create team membership: %s", err.Error())
	}
	return s.GetByID(ctx, membership.ID)
}

func (s *dbTeamMembershipStore) GetByID(ctx context.Context, id string) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.TeamMembership
	if err := grm.Preload("Team").Preload("User").Where("id = ?", id).First(&membership).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("team membership with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch team membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbTeamMembershipStore) GetByTeamAndUser(ctx context.Context, teamID, userID string) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.TeamMembership
	if err := grm.Preload("Team").Preload("User").Where("team_id = ? AND user_id = ?", teamID, userID).First(&membership).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("membership not found for user in team")
		}
		return nil, errors.GeneralError("failed to fetch team membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbTeamMembershipStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("User").Preload("Team").Where("team_id = ?", teamID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list team memberships: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) ListByUserID(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("Team").Preload("User").Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("Team").Preload("User").
		Joins("JOIN teams ON teams.id = team_memberships.team_id").
		Where("team_memberships.user_id = ? AND teams.organisation_id = ?", userID, orgID).
		Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user in org: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) Update(ctx context.Context, id string, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	result := grm.Model(&models.TeamMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"role": membership.Role,
	})
	if result.Error != nil {
		return nil, errors.GeneralError("failed to update team membership: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return nil, errors.NotFound("team membership with id '%s' not found", id)
	}
	return s.GetByID(ctx, id)
}

func (s *dbTeamMembershipStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	result := grm.Where("id = ?", id).Delete(&models.TeamMembership{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete team membership: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("team membership with id '%s' not found", id)
	}
	return nil
}
