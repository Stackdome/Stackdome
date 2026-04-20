package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresAddonDatabaseStoreSpec struct {
	SessionFactory db.SessionFactory
}

type postgresAddonDatabaseStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewPostgresAddonDatabaseStore(spec PostgresAddonDatabaseStoreSpec) stores.PostgresAddonDatabaseStore {
	return &postgresAddonDatabaseStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *postgresAddonDatabaseStore) Create(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingCount int64
	if err := tx.Model(&models.PostgresAddonDatabase{}).
		Where("postgres_addon_id = ? AND name = ?", database.PostgresAddonID, database.Name).
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to check duplicate database name: %s", err.Error())
	}

	if existingCount > 0 {
		tx.Rollback()
		return nil, errors.BadRequest("database with name '%s' already exists in this postgres addon", database.Name)
	}

	if err := tx.Model(&models.PostgresAddonDatabase{}).Omit(clause.Associations).Create(database).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create postgres addon database: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres addon database creation: %s", err.Error())
	}

	return s.GetByID(ctx, database.ID)
}

func (s *postgresAddonDatabaseStore) CreateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	var existingCount int64
	if err := tx.Model(&models.PostgresAddonDatabase{}).
		Where("postgres_addon_id = ? AND name = ?", database.PostgresAddonID, database.Name).
		Count(&existingCount).Error; err != nil {
		return nil, errors.GeneralError("failed to check duplicate database name: %s", err.Error())
	}

	if existingCount > 0 {
		return nil, errors.BadRequest("database with name '%s' already exists in this postgres addon", database.Name)
	}

	if err := tx.Model(&models.PostgresAddonDatabase{}).Omit(clause.Associations).Create(database).Error; err != nil {
		return nil, errors.GeneralError("failed to create postgres addon database: %s", err.Error())
	}

	return database, nil
}

func (s *postgresAddonDatabaseStore) GetByID(ctx context.Context, ID string) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	var database models.PostgresAddonDatabase
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&database).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon database with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get postgres addon database: %s", err.Error())
	}
	return &database, nil
}

func (s *postgresAddonDatabaseStore) GetByName(ctx context.Context, postgresAddonID, name string) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	var database models.PostgresAddonDatabase
	if err := s.sessionFactory.New(ctx).
		Where("postgres_addon_id = ? AND name = ?", postgresAddonID, name).
		First(&database).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon database with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to get postgres addon database by name: %s", err.Error())
	}
	return &database, nil
}

func (s *postgresAddonDatabaseStore) Update(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingDatabase models.PostgresAddonDatabase
	if err := tx.Where("id = ?", database.ID).First(&existingDatabase).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon database with id '%s' not found", database.ID)
		}
		return nil, errors.GeneralError("failed to find postgres addon database for update: %s", err.Error())
	}

	if database.Name != existingDatabase.Name {
		var nameCount int64
		if err := tx.Model(&models.PostgresAddonDatabase{}).
			Where("postgres_addon_id = ? AND name = ? AND id != ?",
				database.PostgresAddonID, database.Name, database.ID).
			Count(&nameCount).Error; err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to check duplicate name: %s", err.Error())
		}

		if nameCount > 0 {
			tx.Rollback()
			return nil, errors.BadRequest("database with name '%s' already exists in this postgres addon", database.Name)
		}
	}

	if err := tx.Model(&existingDatabase).
		Omit(clause.Associations, "id", "postgres_addon_id", "created_at").
		Updates(database).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update postgres addon database: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres addon database update: %s", err.Error())
	}

	return s.GetByID(ctx, database.ID)
}

func (s *postgresAddonDatabaseStore) UpdateWithTx(ctx context.Context, database *models.PostgresAddonDatabase) (*models.PostgresAddonDatabase, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	var existingDatabase models.PostgresAddonDatabase
	if err := tx.Where("id = ?", database.ID).First(&existingDatabase).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon database with id '%s' not found", database.ID)
		}
		return nil, errors.GeneralError("failed to find postgres addon database for update: %s", err.Error())
	}

	if database.Name != existingDatabase.Name {
		var nameCount int64
		if err := tx.Model(&models.PostgresAddonDatabase{}).
			Where("postgres_addon_id = ? AND name = ? AND id != ?",
				database.PostgresAddonID, database.Name, database.ID).
			Count(&nameCount).Error; err != nil {
			return nil, errors.GeneralError("failed to check duplicate name: %s", err.Error())
		}

		if nameCount > 0 {
			return nil, errors.BadRequest("database with name '%s' already exists in this postgres addon", database.Name)
		}
	}

	if err := tx.Model(&existingDatabase).
		Omit(clause.Associations, "id", "postgres_addon_id", "created_at").
		Updates(database).Error; err != nil {
		return nil, errors.GeneralError("failed to update postgres addon database: %s", err.Error())
	}

	return database, nil
}

func (s *postgresAddonDatabaseStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Delete(&models.PostgresAddonDatabase{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete postgres addon database: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon database with id '%s' not found", ID)
	}

	return nil
}

func (s *postgresAddonDatabaseStore) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}

	result := tx.Delete(&models.PostgresAddonDatabase{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete postgres addon database: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon database with id '%s' not found", ID)
	}

	return nil
}

func (s *postgresAddonDatabaseStore) ListByPostgresAddon(ctx context.Context, postgresAddonID string) ([]*models.PostgresAddonDatabase, *errors.ServiceError) {
	var databases []*models.PostgresAddonDatabase

	if err := s.sessionFactory.New(ctx).
		Where("postgres_addon_id = ?", postgresAddonID).
		Order("created_at DESC").
		Find(&databases).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addon databases: %s", err.Error())
	}

	return databases, nil
}

func (s *postgresAddonDatabaseStore) ValidateDatabaseExists(ctx context.Context, databaseID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.PostgresAddonDatabase{}).
		Where("id = ?", databaseID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate postgres addon database existence: %s", err.Error())
	}
	return count > 0, nil
}

func (s *postgresAddonDatabaseStore) ValidateDatabaseNameUnique(ctx context.Context, postgresAddonID, name, excludeID string) (bool, *errors.ServiceError) {
	var count int64
	query := s.sessionFactory.New(ctx).Model(&models.PostgresAddonDatabase{}).
		Where("postgres_addon_id = ? AND name = ?", postgresAddonID, name)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate postgres addon database name uniqueness: %s", err.Error())
	}
	return count == 0, nil
}
