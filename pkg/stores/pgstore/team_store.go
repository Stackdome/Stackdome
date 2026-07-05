package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type dbTeamStore struct {
	sessionFactory db.SessionFactory
}

type TeamStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewTeamStore(spec TeamStoreSpec) stores.TeamStore {
	return &dbTeamStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbTeamStore) Create(ctx context.Context, team *models.Team) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(team).Error; err != nil {
		return nil, errors.GeneralError("failed to create team: %s", err.Error())
	}
	return s.GetByID(ctx, team.ID)
}

func (s *dbTeamStore) GetByID(ctx context.Context, id string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("id = ?", id).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch team: %s", err.Error())
	}
	return &team, nil
}

func (s *dbTeamStore) GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("organisation_id = ? AND name = ?", orgID, name).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team '%s' not found in organisation", name)
		}
		return nil, errors.GeneralError("failed to fetch team: %s", err.Error())
	}
	return &team, nil
}

func (s *dbTeamStore) ListByOrgID(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var teams []*models.Team
	if err := grm.Where("organisation_id = ?", orgID).Order("created_at ASC").Find(&teams).Error; err != nil {
		return nil, errors.GeneralError("failed to list teams: %s", err.Error())
	}
	return teams, nil
}

func (s *dbTeamStore) Update(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Model(&models.Team{}).Where("id = ?", id).Updates(team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update team: %s", err.Error())
	}
	return s.GetByID(ctx, id)
}

func (s *dbTeamStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.Team{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("team with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete team: %s", err.Error())
	}
	return nil
}

func (s *dbTeamStore) GetDefaultTeamForOrg(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("organisation_id = ? AND default_team = ?", orgID, true).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("default team not found for organisation '%s'", orgID)
		}
		return nil, errors.GeneralError("failed to fetch default team: %s", err.Error())
	}
	return &team, nil
}
