package pgstore

// type stackStorageStore struct {
// 	sessionFactory db.SessionFactory
// 	atomicExecutor
// }

// type StackStorageStoreSpec struct {
// 	SessionFactory db.SessionFactory
// }

// func NewStackStorageStore(w StackStorageStoreSpec) stores.StorageStore {
// 	return &stackStorageStore{
// 		sessionFactory: w.SessionFactory,
// 		atomicExecutor: atomicExecutor{sessionFactory: w.SessionFactory},
// 	}
// }

// func (w *stackStorageStore) Create(ctx context.Context, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError) {
// 	tx := w.sessionFactory.New(ctx)
// 	tx.Begin()
// 	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
// 		tx.Rollback()
// 		return nil, errors.GeneralError("failed to create storage: %s", err.Error())
// 	}
// 	for _, volume := range spec.Volumes {
// 		// volume.StorageID = spec.ID
// 		if err := tx.Omit(clause.Associations).Create(&volume).Error; err != nil {
// 			tx.Rollback()
// 			return nil, errors.GeneralError("failed to create volume: %s", err.Error())
// 		}
// 	}
// 	tx.Commit()
// 	return w.GetByID(ctx, spec.ID)
// }

// func (w *stackStorageStore) CreateWithTx(ctx context.Context, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError) {
// 	tx := db.TxFromContext(ctx)
// 	if tx == nil {
// 		return nil, errors.GeneralError("no transaction found in context")
// 	}

// 	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
// 		return nil, errors.GeneralError("failed to create  storage: %s", err.Error())
// 	}
// 	for _, volume := range spec.Volumes {
// 		volume.StorageID = spec.ID
// 		if err := tx.Omit(clause.Associations).Create(&volume).Error; err != nil {
// 			return nil, errors.GeneralError("failed to create volume: %s", err.Error())
// 		}
// 	}
// 	return w.GetByID(ctx, spec.ID)
// }

// func (w *stackStorageStore) GetByID(ctx context.Context, id string) (*models.StackStorage, *errors.ServiceError) {
// 	var res models.StackStorage
// 	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("id = ?", id).First(&res).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, errors.NotFound("storage with id '%s' not found", id)
// 		}
// 		return nil, errors.GeneralError("failed to fetch storage: %s", err.Error())
// 	}
// 	return &res, nil
// }

// func (w *stackStorageStore) GetByIDorName(ctx context.Context, idOrName string, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	if isValidUUID(idOrName) {
// 		return w.GetByID(ctx, idOrName)
// 	}
// 	return w.getByName(ctx, idOrName, userID)
// }

// func (w *stackStorageStore) GetByWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	var res models.StackStorage
// 	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("workspace_name = ? AND user_id = ?", workspaceName, userID).First(&res).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, errors.NotFound("workspace storage not found for workspace '%s'", workspaceName)
// 		}
// 		return nil, errors.GeneralError("failed to fetch storage: %s", err.Error())
// 	}
// 	return &res, nil
// }

// func isValidUUID(id string) bool {
// 	_, err := uuid.Parse(id)
// 	return err == nil
// }

// func (w *stackStorageStore) getByName(ctx context.Context, name string, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	var res models.StackStorage
// 	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("name = ? AND user_id = ?", name, userID).First(&res).Error; err != nil {
// 		if err == gorm.ErrRecordNotFound {
// 			return nil, errors.NotFound("storage with name '%s' not found", name)
// 		}
// 		return nil, errors.GeneralError("failed to fetch storage: %s", err.Error())
// 	}
// 	return &res, nil
// }

// func (w *stackStorageStore) Update(ctx context.Context, id string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError) {
// 	existingWorkspaceStorage, serr := w.GetByID(ctx, id)
// 	if serr != nil {
// 		return nil, serr
// 	}
// 	db := w.sessionFactory.New(ctx).Begin()
// 	if err := db.Model(&models.StackStorage{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
// 		db.Rollback()
// 		return nil, errors.GeneralError("failed to update storage: %s", err.Error())
// 	}

// 	existingWorkspaceVolumesMap := make(map[string]*models.Volume)
// 	for _, volume := range existingWorkspaceStorage.Volumes {
// 		existingWorkspaceVolumesMap[volume.Name] = volume
// 	}

// 	currentVolumeNames := make([]string, len(spec.Volumes))
// 	for i, volume := range spec.Volumes {
// 		currentVolumeNames[i] = volume.Name
// 	}

// 	for _, volume := range spec.Volumes {
// 		if currentVolume, ok := existingWorkspaceVolumesMap[volume.Name]; ok {
// 			if err := db.Model(&models.Volume{}).Where("id = ? AND storage_id = ?", currentVolume.ID, id).Updates(&volume).Error; err != nil {
// 				db.Rollback()
// 				return nil, errors.GeneralError("failed to update volume: %s", err.Error())
// 			}
// 		} else {
// 			volume.StorageID = id
// 			if err := db.Omit(clause.Associations).Create(&volume).Error; err != nil {
// 				db.Rollback()
// 				return nil, errors.GeneralError("failed to create volume: %s", err.Error())
// 			}
// 		}
// 	}

// 	// Delete volumes not in this patch.
// 	for _, existingVolume := range existingWorkspaceStorage.Volumes {
// 		if !slices.Contains(currentVolumeNames, existingVolume.Name) {
// 			if err := db.Delete(&models.Volume{}, "id = ? AND storage_id = ?", existingVolume.ID, id).Error; err != nil {
// 				db.Rollback()
// 				return nil, errors.GeneralError("failed to delete volume: %s as part of storage update", err.Error())
// 			}
// 		}
// 	}
// 	db.Commit()
// 	return w.GetByID(ctx, id)
// }

// func (w *stackStorageStore) UpdateWithTx(ctx context.Context, id string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError) {
// 	existingWorkspaceStorage, serr := w.GetByID(ctx, id)
// 	if serr != nil {
// 		return nil, serr
// 	}

// 	db := db.TxFromContext(ctx)
// 	if db == nil {
// 		return nil, errors.GeneralError("no transaction found in context")
// 	}

// 	if err := db.Model(&models.StackStorage{}).Omit(clause.Associations).Where("id = ?", id).Updates(spec).Error; err != nil {
// 		return nil, errors.GeneralError("failed to update storage: %s", err.Error())
// 	}

// 	existingWorkspaceVolumesMap := make(map[string]*models.Volume)
// 	for _, volume := range existingWorkspaceStorage.Volumes {
// 		existingWorkspaceVolumesMap[volume.Name] = volume
// 	}

// 	currentVolumeNames := make([]string, len(spec.Volumes))
// 	for i, volume := range spec.Volumes {
// 		currentVolumeNames[i] = volume.Name
// 	}

// 	for _, volume := range spec.Volumes {
// 		if currentVolume, ok := existingWorkspaceVolumesMap[volume.Name]; ok {
// 			if err := db.Model(&models.Volume{}).Where("id = ? AND storage_id = ?", currentVolume.ID, id).Updates(&volume).Error; err != nil {
// 				return nil, errors.GeneralError("failed to update volume: %s", err.Error())
// 			}
// 		} else {
// 			volume.StorageID = id
// 			if err := db.Omit(clause.Associations).Create(&volume).Error; err != nil {
// 				return nil, errors.GeneralError("failed to create volume: %s", err.Error())
// 			}
// 		}
// 	}

// 	// Delete volumes not in this patch.
// 	for _, existingVolume := range existingWorkspaceStorage.Volumes {
// 		// We only delete volumes that are not in the current patch and are not in use.
// 		if !slices.Contains(currentVolumeNames, existingVolume.Name) && !existingVolume.Status.InUse {
// 			if err := db.Delete(&models.Volume{}, "id = ? AND storage_id = ?", existingVolume.ID, id).Error; err != nil {
// 				return nil, errors.GeneralError("failed to delete volume: %s as part of storage update", err.Error())
// 			}
// 		}
// 	}
// 	return w.GetByID(ctx, id)
// }

// // UpdateStatus should be used to update the status of a workspace storage. We dont want to update the updated_at field. (Hence using the UpdateColumn method)
// func (w *stackStorageStore) UpdateStatus(ctx context.Context, id string, status *models.StackStorageStatus) (*models.StackStorage, *errors.ServiceError) {
// 	db := w.sessionFactory.New(ctx)
// 	if err := db.Model(&models.StackStorage{}).Where("id = ?", id).UpdateColumn("status", status).Error; err != nil {
// 		return nil, errors.GeneralError("failed to update storage status: %s", err.Error())
// 	}
// 	return w.GetByID(ctx, id)
// }

// func (w *stackStorageStore) Delete(ctx context.Context, id string) *errors.ServiceError {
// 	if err := w.sessionFactory.New(ctx).Delete(&models.StackStorage{}, "id = ?", id).Error; err != nil {
// 		return errors.GeneralError("failed to delete storage: %s", err.Error())
// 	}
// 	return nil
// }

// func (w *stackStorageStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
// 	tx := db.TxFromContext(ctx)
// 	if tx == nil {
// 		return errors.GeneralError("no transaction found in context")
// 	}
// 	if err := tx.Delete(&models.StackStorage{}, "id = ?", id).Error; err != nil {
// 		return errors.GeneralError("failed to delete storage: %s", err.Error())
// 	}
// 	return nil
// }

// func (w *stackStorageStore) DeleteByNameWithTx(ctx context.Context, name string, userID string) *errors.ServiceError {
// 	tx := db.TxFromContext(ctx)
// 	if tx == nil {
// 		return errors.GeneralError("no transaction found in context")
// 	}
// 	if err := tx.Delete(&models.StackStorage{}, "name = ? AND user_id = ?", name, userID).Error; err != nil {
// 		return errors.GeneralError("failed to delete storage: %s", err.Error())
// 	}
// 	return nil
// }

// func (w *stackStorageStore) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.StackStorage, *errors.ServiceError) {
// 	var res []*models.StackStorage
// 	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("organisation_id = ?", organisationID).Find(&res).Error; err != nil {
// 		return nil, errors.GeneralError("failed to list storages: %s", err.Error())
// 	}
// 	return res, nil
// }

// func (w *stackStorageStore) ListByUserID(ctx context.Context, userID string) ([]*models.StackStorage, *errors.ServiceError) {
// 	var storages []*models.StackStorage
// 	if err := w.sessionFactory.New(ctx).Preload("Volumes").Where("user_id = ?", userID).Find(&storages).Error; err != nil {
// 		return nil, errors.GeneralError("failed to list storages by user ID: %s", err.Error())
// 	}
// 	return storages, nil
// }
