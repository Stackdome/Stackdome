package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/davecgh/go-spew/spew"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WorkspaceStorageStore
type workspaceStorageStore struct {
	sessionFactory db.SessionFactory
}

type WorkspaceStorageStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewWorkspaceStorageStore(w WorkspaceStorageStoreSpec) stores.WorkspaceStorageStore {
	return &workspaceStorageStore{
		sessionFactory: w.SessionFactory,
	}
}

func (w *workspaceStorageStore) Create(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	spew.Dump(spec)
	tx := w.sessionFactory.New(ctx)
	tx.Begin()
	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace storage: %s", err.Error())
	}
	for _, volume := range spec.Volumes {
		volume.WorkspaceStorageID = spec.ID
		if err := tx.Omit(clause.Associations).Create(&volume).Error; err != nil {
			tx.Rollback()
			return nil, errors.GeneralError("failed to create volume: %s", err.Error())
		}
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceStorageStore) GetByID(ctx context.Context, id string) (*models.WorkspaceStorage, *errors.ServiceError) {
	var res models.WorkspaceStorage
	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("id = ?", id).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace storage with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch workspace storage: %s", err.Error())
	}
	return &res, nil
}

func (w *workspaceStorageStore) Update(ctx context.Context, id string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	db := w.sessionFactory.New(ctx)
	if err := db.Model(&models.WorkspaceStorage{}).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace storage: %s", err.Error())
	}
	return w.GetByID(ctx, id)
}

func (w *workspaceStorageStore) UpsertStatus(ctx context.Context, id string, status *models.WorkspaceStorageStatus) (*models.WorkspaceStorage, *errors.ServiceError) {
	db := w.sessionFactory.New(ctx)
	if err := db.Model(&models.WorkspaceStorage{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace storage status: %s", err.Error())
	}
	return w.GetByID(ctx, id)
}

func (w *workspaceStorageStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	if err := w.sessionFactory.New(ctx).Delete(&models.WorkspaceStorage{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete workspace storage: %s", err.Error())
	}
	return nil
}

func (w *workspaceStorageStore) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	var res []*models.WorkspaceStorage
	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("organisation_id = ?", organisationID).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspace storages: %s", err.Error())
	}
	return res, nil
}

func (w *workspaceStorageStore) ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	var storages []*models.WorkspaceStorage
	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("user_id = ?", userID).Find(&storages).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspace storages by user ID: %s", err.Error())
	}
	return storages, nil
}

type volumeStore struct {
	sessionFactory db.SessionFactory
}

type VolumeStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewVolumeStore(v VolumeStoreSpec) stores.VolumeStore {
	return &volumeStore{
		sessionFactory: v.SessionFactory,
	}
}

func (v *volumeStore) Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume: %s", err.Error())
	}
	return v.GetByID(ctx, spec.ID)
}

func (v *volumeStore) GetByID(ctx context.Context, id string) (*models.Volume, *errors.ServiceError) {
	var res models.Volume
	if err := v.sessionFactory.New(ctx).Where("id = ?", id).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("volume with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch volume: %s", err.Error())
	}
	return &res, nil
}

func (v *volumeStore) Update(ctx context.Context, id string, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).Model(&models.Volume{}).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume: %s", err.Error())
	}
	return v.GetByID(ctx, id)
}

func (v *volumeStore) UpsertStatus(ctx context.Context, id string, status *models.VolumeStatus) (*models.Volume, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).Model(&models.Volume{}).Where("id = ?", id).Update("volume_status", status).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume status: %s", err.Error())
	}
	return v.GetByID(ctx, id)
}

func (v *volumeStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Delete(&models.Volume{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete volume: %s", err.Error())
	}
	return nil
}

func (v *volumeStore) GetByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.Volume, *errors.ServiceError) {
	var res []*models.Volume
	if err := v.sessionFactory.New(ctx).Where("workspace_storage_id = ?", workspaceStorageID).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch volumes for workspace storage: %s", err.Error())
	}
	return res, nil
}

func (v *volumeStore) ListByIDs(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError) {
	var res []*models.Volume
	if err := v.sessionFactory.New(ctx).Where("id IN ?", ids).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to list volumes by IDs: %s", err.Error())
	}
	return res, nil
}
