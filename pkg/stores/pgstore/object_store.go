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

type ObjectStoreStoreSpec struct {
	SessionFactory db.SessionFactory
}

type objectStoreStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewObjectStoreStore(spec ObjectStoreStoreSpec) stores.ObjectStoreStore {
	return &objectStoreStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (s *objectStoreStore) Create(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingCount int64
	if err := tx.Model(&models.ObjectStore{}).
		Where("organisation_id = ? AND name = ?", objectStore.OrganisationID, objectStore.Name).
		Count(&existingCount).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to check duplicate object store name: %s", err.Error())
	}

	if existingCount > 0 {
		tx.Rollback()
		return nil, errors.BadRequest("object store with name '%s' already exists", objectStore.Name)
	}

	if err := tx.Model(&models.ObjectStore{}).Omit(clause.Associations).Create(objectStore).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create object store: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit object store creation: %s", err.Error())
	}

	return s.GetByID(ctx, objectStore.ID)
}

func (s *objectStoreStore) GetByID(ctx context.Context, ID string) (*models.ObjectStore, *errors.ServiceError) {
	var objectStore models.ObjectStore
	if err := s.sessionFactory.New(ctx).Where("id = ?", ID).First(&objectStore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("object store with id '%s' not found", ID)
		}
		return nil, errors.GeneralError("failed to get object store: %s", err.Error())
	}
	return &objectStore, nil
}

func (s *objectStoreStore) GetByName(ctx context.Context, organisationID, name string) (*models.ObjectStore, *errors.ServiceError) {
	var objectStore models.ObjectStore
	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ? AND name = ?", organisationID, name).
		First(&objectStore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("object store with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to get object store by name: %s", err.Error())
	}
	return &objectStore, nil
}

func (s *objectStoreStore) Update(ctx context.Context, objectStore *models.ObjectStore) (*models.ObjectStore, *errors.ServiceError) {
	tx := s.sessionFactory.New(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingObjectStore models.ObjectStore
	if err := tx.Where("id = ?", objectStore.ID).First(&existingObjectStore).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("object store with id '%s' not found", objectStore.ID)
		}
		return nil, errors.GeneralError("failed to find object store for update: %s", err.Error())
	}

	if err := tx.Model(&existingObjectStore).
		Omit(clause.Associations, "id", "organisation_id", "name", "created_at").
		Updates(objectStore).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update object store: %s", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.GeneralError("failed to commit object store update: %s", err.Error())
	}

	return s.GetByID(ctx, objectStore.ID)
}

func (s *objectStoreStore) UpdateStatus(ctx context.Context, id string, status models.ObjectStoreStatus) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Model(&models.ObjectStore{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return errors.GeneralError("failed to update object store status: %s", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("object store with id '%s' not found", id)
	}
	return nil
}

func (s *objectStoreStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	result := s.sessionFactory.New(ctx).Delete(&models.ObjectStore{}, "id = ?", ID)
	if result.Error != nil {
		return errors.GeneralError("failed to delete object store: %s", result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return errors.NotFound("object store with id '%s' not found", ID)
	}

	return nil
}

func (s *objectStoreStore) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.ObjectStore, *errors.ServiceError) {
	var objectStores []*models.ObjectStore

	if err := s.sessionFactory.New(ctx).
		Where("organisation_id = ?", organisationID).
		Order("created_at DESC").
		Find(&objectStores).Error; err != nil {
		return nil, errors.GeneralError("failed to list object stores: %s", err.Error())
	}

	return objectStores, nil
}

func (s *objectStoreStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.ObjectStore, *errors.ServiceError) {
	var objectStores []*models.ObjectStore

	if err := s.sessionFactory.New(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&objectStores).Error; err != nil {
		return nil, errors.GeneralError("failed to list object stores by team: %s", err.Error())
	}

	return objectStores, nil
}

func (s *objectStoreStore) ValidateObjectStoreExists(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.ObjectStore{}).
		Where("id = ?", objectStoreID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate object store existence: %s", err.Error())
	}
	return count > 0, nil
}

func (s *objectStoreStore) ValidateObjectStoreNameUnique(ctx context.Context, organisationID, name, excludeID string) (bool, *errors.ServiceError) {
	var count int64
	query := s.sessionFactory.New(ctx).Model(&models.ObjectStore{}).
		Where("organisation_id = ? AND name = ?", organisationID, name)

	if excludeID != "" {
		query = query.Where("id != ?", excludeID)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to validate object store name uniqueness: %s", err.Error())
	}
	return count == 0, nil
}

// TODO: Replace JSONB cross-table query with a dedicated object_store_references table
// (see docs/plans/postgres-addon-improvements.md #7).
func (s *objectStoreStore) IsReferencedByAddon(ctx context.Context, objectStoreID string) (bool, *errors.ServiceError) {
	var count int64
	if err := s.sessionFactory.New(ctx).Model(&models.PostgresAddon{}).
		Where("backup_config->>'objectStoreId' = ? OR initialization->'restoreFromObjectStore'->>'objectStoreId' = ?",
			objectStoreID, objectStoreID).
		Count(&count).Error; err != nil {
		return false, errors.GeneralError("failed to check object store references: %s", err.Error())
	}
	return count > 0, nil
}
