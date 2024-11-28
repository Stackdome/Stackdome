package pgstore

import (
	"context"
	"slices"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WorkspaceStorageStore
type workspaceStorageStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

type WorkspaceStorageStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewWorkspaceStorageStore(w WorkspaceStorageStoreSpec) stores.WorkspaceStorageStore {
	return &workspaceStorageStore{
		sessionFactory: w.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: w.SessionFactory},
	}
}

func (w *workspaceStorageStore) Create(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
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

func (w *workspaceStorageStore) CreateWithTx(ctx context.Context, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("no transaction found in context")
	}

	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create workspace storage: %s", err.Error())
	}
	for _, volume := range spec.Volumes {
		volume.WorkspaceStorageID = spec.ID
		if err := tx.Omit(clause.Associations).Create(&volume).Error; err != nil {
			return nil, errors.GeneralError("failed to create volume: %s", err.Error())
		}
	}
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

func (w *workspaceStorageStore) GetByIDorName(ctx context.Context, idOrName string, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	if isValidUUID(idOrName) {
		return w.GetByID(ctx, idOrName)
	}
	return w.getByName(ctx, idOrName, userID)
}

func (w *workspaceStorageStore) GetByWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	var res models.WorkspaceStorage
	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("workspace_name = ? AND user_id = ?", workspaceName, userID).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace storage with workspace name '%s' not found", workspaceName)
		}
		return nil, errors.GeneralError("failed to fetch workspace storage: %s", err.Error())
	}
	return &res, nil
}

func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func (w *workspaceStorageStore) getByName(ctx context.Context, name string, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	var res models.WorkspaceStorage
	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("name = ? AND user_id = ?", name, userID).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace storage with name '%s' not found", name)
		}
		return nil, errors.GeneralError("failed to fetch workspace storage: %s", err.Error())
	}
	return &res, nil
}

func (w *workspaceStorageStore) Update(ctx context.Context, id string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	existingWorkspaceStorage, serr := w.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	db := w.sessionFactory.New(ctx).Begin()
	if err := db.Model(&models.WorkspaceStorage{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		db.Rollback()
		return nil, errors.GeneralError("failed to update workspace storage: %s", err.Error())
	}

	existingWorkspaceVolumesMap := make(map[string]*models.Volume)
	for _, volume := range existingWorkspaceStorage.Volumes {
		existingWorkspaceVolumesMap[volume.Name] = volume
	}

	currentVolumeNames := make([]string, len(spec.Volumes))
	for i, volume := range spec.Volumes {
		currentVolumeNames[i] = volume.Name
	}

	for _, volume := range spec.Volumes {
		if currentVolume, ok := existingWorkspaceVolumesMap[volume.Name]; ok {
			if err := db.Model(&models.Volume{}).Where("id = ? AND workspace_storage_id = ?", currentVolume.ID, id).Updates(&volume).Error; err != nil {
				db.Rollback()
				return nil, errors.GeneralError("failed to update volume: %s", err.Error())
			}
		} else {
			volume.WorkspaceStorageID = id
			if err := db.Omit(clause.Associations).Create(&volume).Error; err != nil {
				db.Rollback()
				return nil, errors.GeneralError("failed to create volume: %s", err.Error())
			}
		}
	}

	// Delete volumes not in this patch.
	for _, existingVolume := range existingWorkspaceStorage.Volumes {
		if !slices.Contains(currentVolumeNames, existingVolume.Name) {
			if err := db.Delete(&models.Volume{}, "id = ? AND workspace_storage_id = ?", existingVolume.ID, id).Error; err != nil {
				db.Rollback()
				return nil, errors.GeneralError("failed to delete volume: %s as part of workspace storage update", err.Error())
			}
		}
	}
	db.Commit()
	return w.GetByID(ctx, id)
}

func (w *workspaceStorageStore) UpdateWithTx(ctx context.Context, id string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	existingWorkspaceStorage, serr := w.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}

	db := db.TxFromContext(ctx)
	if db == nil {
		return nil, errors.GeneralError("no transaction found in context")
	}

	if err := db.Model(&models.WorkspaceStorage{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to update workspace storage: %s", err.Error())
	}

	existingWorkspaceVolumesMap := make(map[string]*models.Volume)
	for _, volume := range existingWorkspaceStorage.Volumes {
		existingWorkspaceVolumesMap[volume.Name] = volume
	}

	currentVolumeNames := make([]string, len(spec.Volumes))
	for i, volume := range spec.Volumes {
		currentVolumeNames[i] = volume.Name
	}

	for _, volume := range spec.Volumes {
		if currentVolume, ok := existingWorkspaceVolumesMap[volume.Name]; ok {
			if err := db.Model(&models.Volume{}).Where("id = ? AND workspace_storage_id = ?", currentVolume.ID, id).Updates(&volume).Error; err != nil {
				return nil, errors.GeneralError("failed to update volume: %s", err.Error())
			}
		} else {
			volume.WorkspaceStorageID = id
			if err := db.Omit(clause.Associations).Create(&volume).Error; err != nil {
				return nil, errors.GeneralError("failed to create volume: %s", err.Error())
			}
		}
	}

	// Delete volumes not in this patch.
	for _, existingVolume := range existingWorkspaceStorage.Volumes {
		// We only delete volumes that are not in the current patch and are not in use.
		if !slices.Contains(currentVolumeNames, existingVolume.Name) && !existingVolume.Status.InUse {
			if err := db.Delete(&models.Volume{}, "id = ? AND workspace_storage_id = ?", existingVolume.ID, id).Error; err != nil {
				return nil, errors.GeneralError("failed to delete volume: %s as part of workspace storage update", err.Error())
			}
		}
	}
	return w.GetByID(ctx, id)
}

// UpdateStatus should be used to update the status of a workspace storage. We dont want to update the updated_at field. (Hence using the UpdateColumn method)
func (w *workspaceStorageStore) UpdateStatus(ctx context.Context, id string, status *models.WorkspaceStorageStatus) (*models.WorkspaceStorage, *errors.ServiceError) {
	db := w.sessionFactory.New(ctx)
	if err := db.Model(&models.WorkspaceStorage{}).Where("id = ?", id).UpdateColumn("status", status).Error; err != nil {
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

func (w *workspaceStorageStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("no transaction found in context")
	}
	if err := tx.Delete(&models.WorkspaceStorage{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete workspace storage: %s", err.Error())
	}
	return nil
}

func (w *workspaceStorageStore) DeleteByNameWithTx(ctx context.Context, name string, userID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("no transaction found in context")
	}
	if err := tx.Delete(&models.WorkspaceStorage{}, "name = ? AND user_id = ?", name, userID).Error; err != nil {
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
	return v.GetByID(ctx, spec.ID, spec.WorkspaceStorageID)
}

func (v *volumeStore) GetByID(ctx context.Context, id string, workspaceStorageID string) (*models.Volume, *errors.ServiceError) {
	var res models.Volume
	if err := v.sessionFactory.New(ctx).Where("id = ? AND workspace_storage_id = ?", id, workspaceStorageID).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("volume with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch volume: %s", err.Error())
	}
	return &res, nil
}

func (v *volumeStore) Update(ctx context.Context, id string, workspaceStorageID string, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).
		Model(&models.Volume{}).
		Where("id = ? AND workspace_storage_id = ?", id, workspaceStorageID).
		Updates(map[string]interface{}{
			"Labels":        spec.Labels,
			"Annotations":   spec.Annotations,
			"VolumeSource":  spec.VolumeSource,
			"SyncBeforeUse": spec.SyncBeforeUse,
		}).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume: %s", err.Error())
	}
	return v.GetByID(ctx, id, workspaceStorageID)
}

func (v *volumeStore) UpdateStatus(ctx context.Context, id string, workspaceStorageID string, status *models.VolumeStatus) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Model(&models.Volume{}).
		Where("id = ? AND workspace_storage_id = ?", id, workspaceStorageID).
		UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update volume status: %s", err.Error())
	}
	return nil
}

func (v *volumeStore) Delete(ctx context.Context, id string, workspaceStorageID string) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Delete(&models.Volume{}, "id = ? AND workspace_storage_id = ?", id, workspaceStorageID).Error; err != nil {
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
