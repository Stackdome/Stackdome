package pgstore

import (
	"context"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type workspaceUserStore struct {
	sessionFactory          db.SessionFactory
	workspaceNamespaceStore stores.WorkspaceNamespaceStore
	atomicExecutor
}

type WorkspaceUserStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewWorkspaceUserStore(spec WorkspaceUserStoreSpec) stores.WorkspaceUserStore {
	return &workspaceUserStore{
		sessionFactory: spec.SessionFactory,
		workspaceNamespaceStore: NewWorkspaceNamespaceStore(WorkspaceNamespaceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (w *workspaceUserStore) CreateWithTx(ctx context.Context, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError) {
	_, fErr := w.GetByUserID(ctx, spec.UserID)
	if fErr == nil {
		return nil, errors.Conflict("workspace provision request for user_id '%s' already exists", spec.UserID)
	}
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create workspace provision request: %s", err.Error())
	}
	for _, wn := range spec.WorkspaceNamespaces {
		wn.UserID = spec.UserID
		wn.WorkspaceUserID = spec.ID
		fmt.Printf("setting workspace user id %s\n", wn.WorkspaceUserID)
		fmt.Printf("setting user id %s\n", wn.UserID)
	}
	if _, err := w.workspaceNamespaceStore.CreateBatchWithTx(ctx, spec.WorkspaceNamespaces); err != nil {
		return nil, err
	}
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceUserStore) Create(ctx context.Context, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError) {
	_, fErr := w.GetByUserID(ctx, spec.UserID)
	if fErr == nil {
		return nil, errors.Conflict("workspace provision request for user_id '%s' already exists", spec.UserID)
	}
	tx := w.sessionFactory.New(ctx)
	tx.Begin()
	if err := tx.Omit(clause.Associations).Create(&spec).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to create workspace provision request: %s", err.Error())
	}
	for _, wn := range spec.WorkspaceNamespaces {
		wn.UserID = spec.UserID
		wn.WorkspaceUserID = spec.ID
	}
	if _, err := w.workspaceNamespaceStore.CreateBatch(ctx, spec.WorkspaceNamespaces); err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return w.GetByID(ctx, spec.ID)
}

func (w *workspaceUserStore) GetByID(ctx context.Context, id string) (*models.WorkspaceUser, *errors.ServiceError) {
	var grm *gorm.DB
	grm = db.TxFromContext(ctx)
	if grm == nil {
		grm = w.sessionFactory.New(ctx)
	}
	var res models.WorkspaceUser
	if err := grm.Model(&models.WorkspaceUser{}).Preload(clause.Associations).Where("id = ?", id).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace provision request with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch workspace provision request: %s", err.Error())
	}
	return &res, nil
}

func (w *workspaceUserStore) InternalList(ctx context.Context, query string, args ...any) ([]*models.WorkspaceUser, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var requests []*models.WorkspaceUser
	if err := grm.Model(&models.WorkspaceUser{}).
		Where(query, args...).
		Preload(clause.Associations).
		Find(&requests).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch workspace provision requests: %s", err.Error())
	}
	return requests, nil
}

func (w *workspaceUserStore) GetByUserID(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var res models.WorkspaceUser
	if err := grm.Model(&models.WorkspaceUser{}).Preload(clause.Associations).Where("user_id = ?", userID).First(&res).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("workspace provision request with user_id '%s' not found", userID)
		}
		return nil, errors.GeneralError("failed to fetch workspace provision request: %s", err.Error())
	}
	return &res, nil
}

func (w *workspaceUserStore) ListByOrgID(ctx context.Context, orgID string) ([]*models.WorkspaceUser, *errors.ServiceError) {
	grm := w.sessionFactory.New(ctx)
	var res []*models.WorkspaceUser

	if err := grm.Model(&models.WorkspaceUser{}).Preload(clause.Associations).Where("organisation_id = ?", orgID).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspace provision requests for orgID '%s': %s", orgID, err.Error())
	}
	return res, nil
}

func (w *workspaceUserStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.WorkspaceUser, *errors.ServiceError) {
	var workspaceUsers []*models.WorkspaceUser
	if err := w.sessionFactory.New(ctx).
		Model(&models.WorkspaceUser{}).
		Preload(clause.Associations).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&workspaceUsers).Error; err != nil {
		return nil, errors.GeneralError("failed to list workspace users by team: %s", err.Error())
	}
	return workspaceUsers, nil
}

func (w *workspaceUserStore) Update(ctx context.Context, id string, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError) {
	existingObj, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx := w.sessionFactory.New(ctx).Begin()

	if err := tx.Model(&existingObj).Updates(
		map[string]interface{}{
			"ssh_public_key": spec.SshPublicKey,
		}).Error; err != nil {
		tx.Rollback()
		return nil, errors.GeneralError("failed to update object: %s", err.Error())
	}

	existingWorkspaceNames := existingObj.WorkspaceNamespaces
	existingWSNameMap := make(map[string]bool)
	existingWSMap := make(map[string]*models.WorkspaceNamespace)

	for _, wn := range existingWorkspaceNames {
		existingWSNameMap[wn.Workspace] = false
		existingWSMap[wn.Workspace] = wn
	}

	for _, currentWS := range spec.WorkspaceNamespaces {
		if _, ok := existingWSMap[currentWS.Workspace]; !ok {
			// New workspaces.
			currentWS.UserID = existingObj.UserID
			currentWS.WorkspaceUserID = existingObj.ID
			if _, err := w.workspaceNamespaceStore.Create(ctx, currentWS); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		existingWSNameMap[currentWS.Workspace] = true
	}

	for workspaceName, inUse := range existingWSNameMap {
		if !inUse {
			// Disable workspace namespace objects not in the current patch.
			wn := existingWSMap[workspaceName]
			wn.Enabled = false
			if _, err := w.workspaceNamespaceStore.Update(ctx, wn.Workspace, wn.UserID, wn); err != nil {
				tx.Rollback()
				return nil, err
			}
		}
	}
	tx.Commit()
	return w.GetByID(ctx, id)
}

func (w *workspaceUserStore) UpdateWithTx(ctx context.Context, id string, spec *models.WorkspaceUser) (*models.WorkspaceUser, *errors.ServiceError) {
	existingObj, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}

	if err := tx.Model(&existingObj).Updates(
		map[string]interface{}{
			"ssh_public_key": spec.SshPublicKey,
		}).Error; err != nil {
		return nil, errors.GeneralError("failed to update object: %s", err.Error())
	}

	existingWorkspaceMap := make(map[string]*models.WorkspaceNamespace)
	currentWorkspaceMap := make(map[string]*models.WorkspaceNamespace)
	for _, wn := range existingObj.WorkspaceNamespaces {
		existingWorkspaceMap[wn.Workspace] = wn
	}

	for _, wn := range spec.WorkspaceNamespaces {
		currentWorkspaceMap[wn.Workspace] = wn
	}

	// Create new workspaces if they don't exist currently.
	for _, currentWS := range spec.WorkspaceNamespaces {
		if _, ok := existingWorkspaceMap[currentWS.Workspace]; !ok {
			// New workspaces.
			fmt.Printf("creating new workspace %s\n", currentWS.Workspace)
			currentWS.UserID = existingObj.UserID
			currentWS.WorkspaceUserID = existingObj.ID
			if _, err := w.workspaceNamespaceStore.Create(ctx, currentWS); err != nil {
				return nil, err
			}
		} else {
			existingWS := existingWorkspaceMap[currentWS.Workspace]
			// Enable workspace namespace objects in the current patch.
			fmt.Printf("enabling workspace %s\n", currentWS.Workspace)
			currentWS.Enabled = true
			if _, err := w.workspaceNamespaceStore.UpdateWithTx(ctx, existingWS.Workspace, existingWS.UserID, currentWS); err != nil {
				return nil, err
			}
		}
	}

	// Disable workspaces not in the current patch.
	for workspaceName, existingWN := range existingWorkspaceMap {
		if _, ok := currentWorkspaceMap[workspaceName]; !ok {
			// Disable workspace namespace objects not in the current patch.
			fmt.Printf("disabling workspace %s\n", workspaceName)
			existingWN.Enabled = false
			if _, err := w.workspaceNamespaceStore.UpdateWithTx(ctx, existingWN.Workspace, existingWN.UserID, existingWN); err != nil {
				return nil, err
			}
		}
	}

	return w.GetByID(ctx, id)
}

// We dont want to update updated_at field, as this is a status update.
func (w *workspaceUserStore) PatchStatus(ctx context.Context, id string, status *models.WorkspaceUserStatus) (*models.WorkspaceUser, *errors.ServiceError) {
	existingObj, err := w.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	grm := w.sessionFactory.New(ctx)
	if status == nil {
		return nil, errors.GeneralError("status is nil")
	}
	if err := grm.Model(&existingObj).UpdateColumn("status", status).Error; err != nil {
		return nil, errors.GeneralError("failed to update object: %s", err.Error())
	}
	return w.GetByID(ctx, id)
}

func (w *workspaceUserStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	if _, err := w.GetByID(ctx, id); err != nil {
		return err
	}
	grm := w.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.WorkspaceUser{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("workspace provision request with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete object: %s", err.Error())
	}
	return nil
}

func (w *workspaceUserStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	if _, err := w.GetByID(ctx, id); err != nil {
		return err
	}
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("id = ?", id).Delete(&models.WorkspaceUser{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("workspace provision request with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete object: %s", err.Error())
	}
	return nil
}
