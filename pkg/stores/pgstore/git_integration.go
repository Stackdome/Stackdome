package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type GitIntegrationStoreSpec struct {
	SessionFactory db.SessionFactory
}

type gitIntegrationStore struct {
	sessionFactory db.SessionFactory
}

func NewGitIntegrationStore(spec GitIntegrationStoreSpec) stores.GitIntegrationStore {
	return &gitIntegrationStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *gitIntegrationStore) Create(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Create(integration).Error; err != nil {
		return nil, errors.GeneralError("failed to create git integration: %s", err.Error())
	}
	return s.GetByID(ctx, integration.ID)
}

func (s *gitIntegrationStore) GetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError) {
	var integration models.GitIntegration
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&integration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("git integration with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get git integration: %s", err.Error())
	}
	return &integration, nil
}

func (s *gitIntegrationStore) GetByOrgAndHost(ctx context.Context, organisationID, host string) (*models.GitIntegration, *errors.ServiceError) {
	var integration models.GitIntegration
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND host = ? AND type = ?", organisationID, host, models.GitIntegrationTypeGitCredentials).
		First(&integration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("no git integration for host '%s'", host)
		}
		return nil, errors.GeneralError("failed to get git integration for host: %s", err.Error())
	}
	return &integration, nil
}

func (s *gitIntegrationStore) GetGitHubAppForOrg(ctx context.Context, organisationID string) (*models.GitIntegration, *errors.ServiceError) {
	var integration models.GitIntegration
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND type = ?", organisationID, models.GitIntegrationTypeGitHubApp).
		First(&integration).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("no GitHub App integration for organisation '%s'", organisationID)
		}
		return nil, errors.GeneralError("failed to get GitHub App integration: %s", err.Error())
	}
	return &integration, nil
}

func (s *gitIntegrationStore) ListByOrgID(ctx context.Context, organisationID string) ([]*models.GitIntegration, *errors.ServiceError) {
	var integrations []*models.GitIntegration
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ?", organisationID).
		Order("type, host").
		Find(&integrations).Error; err != nil {
		return nil, errors.GeneralError("failed to list git integrations: %s", err.Error())
	}
	return integrations, nil
}

func (s *gitIntegrationStore) Update(ctx context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).Save(integration).Error; err != nil {
		return nil, errors.GeneralError("failed to update git integration: %s", err.Error())
	}
	return s.GetByID(ctx, integration.ID)
}

func (s *gitIntegrationStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Where("id = ?", ID).Delete(&models.GitIntegration{})
	if result.Error != nil {
		return errors.GeneralError("failed to delete git integration: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("git integration with id '%s' not found", ID)
	}
	return nil
}

func (s *gitIntegrationStore) ListGitHubApps(ctx context.Context) ([]*models.GitIntegration, *errors.ServiceError) {
	var integrations []*models.GitIntegration
	if err := s.sessionFactory.New(ctx).
		Where("type = ?", models.GitIntegrationTypeGitHubApp).
		Find(&integrations).Error; err != nil {
		return nil, errors.GeneralError("failed to list GitHub App integrations: %s", err.Error())
	}
	return integrations, nil
}
