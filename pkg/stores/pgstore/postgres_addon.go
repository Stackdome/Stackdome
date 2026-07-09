package pgstore

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostgresAddonStoreSpec struct {
	SessionFactory db.SessionFactory
}

type postgresAddonStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

// List of mutable fields that can be updated
var mutableFields = []string{
	"Labels", "Annotations", "Instances", "Resources", "Storage",
	"Configuration", "BackupConfig", "LifecycleConfig", "UpdatedAt",
}

func NewPostgresAddonStore(spec PostgresAddonStoreSpec) stores.PostgresAddonStore {
	return &postgresAddonStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *postgresAddonStore) Create(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingCount int64
	if err := tx.Model(&models.PostgresAddon{}).
		Where("organisation_id = ? AND name = ?", addon.OrganisationID, addon.Name).
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to check duplicate postgres addon name: %s", err.Error())
	}

	if existingCount > 0 {
		tx.Rollback()
		return nil, errors.BadRequest("postgres addon with name '%s' already exists", addon.Name)
	}

	if err := tx.Model(&models.PostgresAddon{}).Omit(clause.Associations).Create(addon).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create postgres addon: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres addon creation: %s", err.Error())
	}

	return s.GetByID(ctx, addon.ID)
}

func (s *postgresAddonStore) CreateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	var existingCount int64
	if err := tx.Model(&models.PostgresAddon{}).
		Where("organisation_id = ? AND name = ?", addon.OrganisationID, addon.Name).
		Count(&existingCount).Error; err != nil {
		return nil, errors.GeneralError("failed to check duplicate postgres addon name: %s", err.Error())
	}

	if existingCount > 0 {
		return nil, errors.BadRequest("postgres addon with name '%s' already exists", addon.Name)
	}

	if err := tx.Model(&models.PostgresAddon{}).Omit(clause.Associations).Create(addon).Error; err != nil {
		return nil, errors.GeneralError("failed to create postgres addon: %s", err.Error())
	}

	return addon, nil
}

func (s *postgresAddonStore) GetByID(ctx context.Context, ID string) (*models.PostgresAddon, *errors.ServiceError) {
	var addon models.PostgresAddon
	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("id = ?", ID).
		First(&addon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get postgres addon: %s", err.Error())
	}
	return &addon, nil
}

func (s *postgresAddonStore) GetByName(ctx context.Context, organisationID, name string) (*models.PostgresAddon, *errors.ServiceError) {
	var addon models.PostgresAddon
	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("organisation_id = ? AND name = ?", organisationID, name).
		First(&addon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to get postgres addon by name: %s", err.Error())
	}
	return &addon, nil
}

func (s *postgresAddonStore) Update(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingAddon models.PostgresAddon
	if err := tx.Where("id = ?", addon.ID).First(&existingAddon).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon with id '%s' not found", addon.ID)
		}
		return nil, errors.GeneralError("failed to find postgres addon for update: %s", err.Error())
	}

	// Update only mutable fields
	if err := tx.Model(&existingAddon).
		Select(mutableFields).
		Omit(clause.Associations).
		Updates(addon).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update postgres addon: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit postgres addon update: %s", err.Error())
	}

	return s.GetByID(ctx, addon.ID)
}

func (s *postgresAddonStore) UpdateWithTx(ctx context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	var existingAddon models.PostgresAddon
	if err := tx.Where("id = ?", addon.ID).First(&existingAddon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("postgres addon with id '%s' not found", addon.ID)
		}
		return nil, errors.GeneralError("failed to find postgres addon for update: %s", err.Error())
	}

	// Update only mutable fields
	if err := tx.Model(&existingAddon).
		Select(mutableFields).
		Omit(clause.Associations).
		Updates(addon).Error; err != nil {
		return nil, errors.GeneralError("failed to update postgres addon: %s", err.Error())
	}

	return addon, nil
}

func (s *postgresAddonStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Delete(&models.PostgresAddon{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete postgres addon: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon with id '%s' not found", ID)
	}

	return nil
}

func (s *postgresAddonStore) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	var addons []*models.PostgresAddon

	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("organisation_id = ?", organisationID).
		Order("created_at DESC").
		Find(&addons).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addons: %s", err.Error())
	}

	return addons, nil
}

func (s *postgresAddonStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	var addons []*models.PostgresAddon

	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&addons).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addons by team: %s", err.Error())
	}

	return addons, nil
}

func (s *postgresAddonStore) ListByTeamIDs(ctx context.Context, teamIDs []string) ([]*models.PostgresAddon, *errors.ServiceError) {
	if len(teamIDs) == 0 {
		return []*models.PostgresAddon{}, nil
	}
	var addons []*models.PostgresAddon
	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("team_id IN ?", teamIDs).
		Order("created_at DESC").
		Find(&addons).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addons by teams: %s", err.Error())
	}
	return addons, nil
}

func (s *postgresAddonStore) ListByCluster(ctx context.Context, clusterID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	var addons []*models.PostgresAddon

	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where("cluster_id = ?", clusterID).
		Order("created_at DESC").
		Find(&addons).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addons by cluster: %s", err.Error())
	}

	return addons, nil
}

func (s *postgresAddonStore) ValidateAddonExists(ctx context.Context, addonID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("id = ?", addonID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate postgres addon existence: %s", err.Error())
	}
	return count > 0, nil
}

func (s *postgresAddonStore) ValidateAddonNameUnique(ctx context.Context, organisationID, name, excludeID string) (bool, *errors.ServiceError) {
	var count int64
	query := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("organisation_id = ? AND name = ?", organisationID, name)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate postgres addon name uniqueness: %s", err.Error())
	}
	return count == 0, nil
}

func (s *postgresAddonStore) UpdateStatus(ctx context.Context, id string, status *models.PostgresAddonStatus) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return errors.GeneralError("failed to update postgres addon status: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon with id '%s' not found", id)
	}

	return nil
}

func (s *postgresAddonStore) UpdateBackupRequestedAt(ctx context.Context, id string, timestamp *time.Time) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("id = ?", id).
		Update("backup_requested_at", timestamp)

	if result.Error != nil {
		return errors.GeneralError("failed to update postgres addon backup requested timestamp: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon with id '%s' not found", id)
	}

	return nil
}

func (s *postgresAddonStore) UpdateDeletionTimestamp(ctx context.Context, id string, timestamp *time.Time) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("id = ?", id).
		Update("deletion_timestamp", timestamp)

	if result.Error != nil {
		return errors.GeneralError("failed to update postgres addon deletion timestamp: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("postgres addon with id '%s' not found", id)
	}

	return nil
}

func (s *postgresAddonStore) InternalList(ctx context.Context, query string, args ...any) ([]*models.PostgresAddon, *errors.ServiceError) {
	var addons []*models.PostgresAddon

	if err := s.sessionFactory.New(ctx).
		Preload("Databases").
		Preload("Backups").
		Where(query, args...).
		Find(&addons).Error; err != nil {
		return nil, errors.GeneralError("failed to list postgres addons: %s", err.Error())
	}

	return addons, nil
}
