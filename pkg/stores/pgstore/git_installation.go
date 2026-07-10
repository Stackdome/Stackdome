package pgstore

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GitInstallationStoreSpec struct {
	SessionFactory db.SessionFactory
}

type gitInstallationStore struct {
	sessionFactory db.SessionFactory
}

func NewGitInstallationStore(spec GitInstallationStoreSpec) stores.GitInstallationStore {
	return &gitInstallationStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *gitInstallationStore) Upsert(ctx context.Context, installation *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
	if err := s.sessionFactory.New(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "git_integration_id"}, {Name: "installation_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"account_login", "account_type", "repository_selection", colUpdatedAt,
			}),
		}).
		Create(installation).Error; err != nil {
		return nil, errors.GeneralError("failed to upsert git installation: %s", err.Error())
	}
	return installation, nil
}

func (s *gitInstallationStore) ListByIntegrationID(ctx context.Context, integrationID string) ([]*models.GitInstallation, *errors.ServiceError) {
	var installations []*models.GitInstallation
	if err := s.sessionFactory.New(ctx).
		Where("git_integration_id = ?", integrationID).
		Order("account_login").
		Find(&installations).Error; err != nil {
		return nil, errors.GeneralError("failed to list git installations: %s", err.Error())
	}
	return installations, nil
}

func (s *gitInstallationStore) GetByIntegrationAndAccount(ctx context.Context, integrationID, accountLogin string) (*models.GitInstallation, *errors.ServiceError) {
	var installation models.GitInstallation
	if err := s.sessionFactory.New(ctx).
		Where("git_integration_id = ? AND LOWER(account_login) = LOWER(?)", integrationID, accountLogin).
		First(&installation).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("no installation for account '%s'", accountLogin)
		}
		return nil, errors.GeneralError("failed to get git installation: %s", err.Error())
	}
	return &installation, nil
}

func (s *gitInstallationStore) DeleteByInstallationID(ctx context.Context, integrationID string, installationID int64) *errors.ServiceError {
	if err := s.sessionFactory.New(ctx).
		Where("git_integration_id = ? AND installation_id = ?", integrationID, installationID).
		Delete(&models.GitInstallation{}).Error; err != nil {
		return errors.GeneralError("failed to delete git installation: %s", err.Error())
	}
	return nil
}
