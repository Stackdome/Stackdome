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

type PostgresBackupStoreSpec struct {
	SessionFactory db.SessionFactory
}

type postgresBackupStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewPostgresBackupStore(spec PostgresBackupStoreSpec) stores.PostgresBackupStore {
	return &postgresBackupStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *postgresBackupStore) Create(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&models.PostgresBackup{}).Omit(clause.Associations).Create(backup).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create postgres backup: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres backup creation: %s", err.Error())
	}

	return s.GetByID(ctx, backup.ID)
}

func (s *postgresBackupStore) GetByID(ctx context.Context, ID string) (*models.PostgresBackup, *errors.ServiceError) {
	var backup models.PostgresBackup
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&backup).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("postgres backup with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get postgres backup: %s", err.Error())
	}
	return &backup, nil
}

func (s *postgresBackupStore) GetByName(ctx context.Context, postgresAddonID string, name string) (*models.PostgresBackup, *errors.ServiceError) {
	var backup models.PostgresBackup
	if err := s.sessionFactory.New(ctx).
		Where("postgres_addon_id = ? AND name = ?", postgresAddonID, name).
		First(&backup).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("postgres backup '%s' not found for addon '%s'", name, postgresAddonID)
		}
		return nil, errors.GeneralError("failed to get postgres backup by name: %s", err.Error())
	}
	return &backup, nil
}

func (s *postgresBackupStore) Update(ctx context.Context, backup *models.PostgresBackup) (*models.PostgresBackup, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingBackup models.PostgresBackup
	if err := tx.Where("id = ?", backup.ID).First(&existingBackup).Error; err != nil {
		tx.Rollback()
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("postgres backup with id '%s' not found", backup.ID)
		}
		return nil, errors.GeneralError("failed to find postgres backup for update: %s", err.Error())
	}

	if err := tx.Model(&existingBackup).
		Omit(clause.Associations, "id", "postgres_addon_id", "created_at").
		Updates(backup).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update postgres backup: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres backup update: %s", err.Error())
	}

	return s.GetByID(ctx, backup.ID)
}

func (s *postgresBackupStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Delete(&models.PostgresBackup{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete postgres backup: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres backup with id '%s' not found", ID)
	}

	return nil
}

func (s *postgresBackupStore) ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresBackup, *errors.ServiceError) {
	var backups []*models.PostgresBackup

	if err := s.sessionFactory.New(ctx).
		Where("postgres_addon_id = ?", postgresAddonID).
		Order("created_at DESC").
		Find(&backups).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres backups: %s", err.Error())
	}

	return backups, nil
}

func (s *postgresBackupStore) ValidateBackupExists(ctx context.Context, backupID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.PostgresBackup{}).
		Where("id = ?", backupID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate postgres backup existence: %s", err.Error())
	}
	return count > 0, nil
}
