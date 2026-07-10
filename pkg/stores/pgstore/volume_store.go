package pgstore

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type volumeStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

type VolumeStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewVolumeStore(v VolumeStoreSpec) stores.VolumeStore {
	return &volumeStore{
		sessionFactory: v.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: v.SessionFactory},
	}
}

func (v *volumeStore) Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume: %s", err.Error())
	}
	return v.GetByID(ctx, spec.ID)
}

func (v *volumeStore) InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError) {
	var res []*models.Volume
	if err := v.sessionFactory.New(ctx).Where(ids).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch volumes: %s", err.Error())
	}
	return res, nil
}

func (v *volumeStore) InternalListNotReady(ctx context.Context) ([]*models.Volume, *errors.ServiceError) {
	var res []*models.Volume
	if err := v.sessionFactory.New(ctx).
		Where("status IS NULL OR status->>'phase' IS NULL OR status->>'phase' <> ?", models.VolumePhaseReady).
		Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch not-ready volumes: %s", err.Error())
	}
	return res, nil
}

func (v *volumeStore) GetByVolumeNameAndNamespace(ctx context.Context, name string, namespace string) (*models.Volume, *errors.ServiceError) {
	var res models.Volume
	if err := v.sessionFactory.New(ctx).Where("name = ? AND namespace = ?", name, namespace).First(&res).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("volume with name '%s' and namespace '%s' not found", name, namespace)
		}
		return nil, errors.GeneralError("failed to fetch volume: %s", err.Error())
	}
	return &res, nil
}

func (v *volumeStore) CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Create(&spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume: %s", err.Error())
	}
	return v.GetByID(ctx, spec.ID)
}

func (v *volumeStore) GetByID(ctx context.Context, id string) (*models.Volume, *errors.ServiceError) {
	var res models.Volume
	if err := v.sessionFactory.New(ctx).Where("id = ?", id).First(&res).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("volume with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch volume: %s", err.Error())
	}
	return &res, nil
}

func (v *volumeStore) UpdateGitRepoSourceRevision(ctx context.Context, id string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	existing, err := v.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.VolumeSource.GitRepoSource == nil {
		return nil, errors.GeneralError("volume source is not git repo source")
	}

	existing.VolumeSource.GitRepoSource.Revision = revision
	if err := v.sessionFactory.New(ctx).Model(&existing).Update("VolumeSource", existing.VolumeSource).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume source revision: %s", err.Error())
	}
	return existing, nil
}

func (v *volumeStore) UpdateGitRepoSourceRevisionWithTx(ctx context.Context, id string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	existing, err := v.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.VolumeSource.GitRepoSource == nil {
		return nil, errors.GeneralError("volume source is not git repo source")
	}
	existing.VolumeSource.GitRepoSource.Revision = revision
	if err := tx.Model(&existing).Update("VolumeSource", existing.VolumeSource).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume source revision: %s", err.Error())
	}
	return existing, nil
}

func (v *volumeStore) UpdateRemoteDirSourceHash(ctx context.Context, id string, remoteDirHash string) (*models.Volume, *errors.ServiceError) {
	existing, err := v.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if existing.VolumeSource.RemoteDirSource == nil {
		return nil, errors.GeneralError("volume source is not remote dir source")
	}

	existing.VolumeSource.RemoteDirSource.CurrentDirectoryHash = remoteDirHash

	if err := v.sessionFactory.New(ctx).Model(&existing).Update("VolumeSource", existing.VolumeSource).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume source revision: %s", err.Error())
	}
	return existing, nil
}

func (v *volumeStore) UpdateRemoteDirSourceHashWithTx(ctx context.Context, id string, remoteDirHash string) (*models.Volume, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	existing, err := v.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.VolumeSource.RemoteDirSource == nil {
		return nil, errors.GeneralError("volume source is not remote dir source")
	}
	existing.VolumeSource.RemoteDirSource.CurrentDirectoryHash = remoteDirHash
	if err := tx.Model(&existing).Update("VolumeSource", existing.VolumeSource).Error; err != nil {
		return nil, errors.GeneralError("failed to update volume source revision: %s", err.Error())
	}
	return existing, nil
}

func (v *volumeStore) UpdateStatus(ctx context.Context, id string, status *models.VolumeStatus) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Model(&models.Volume{}).
		Where("id = ?", id).
		UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update volume status: %s", err.Error())
	}
	return nil
}

func (v *volumeStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Delete(&models.Volume{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete volume: %s", err.Error())
	}
	return nil
}

func (v *volumeStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Delete(&models.Volume{}, "id = ?", id).Error; err != nil {
		return errors.GeneralError("failed to delete volume: %s", err.Error())
	}
	return nil
}

func (v *volumeStore) GetByUserID(ctx context.Context, userID string) ([]*models.Volume, *errors.ServiceError) {
	var res []*models.Volume
	if err := v.sessionFactory.New(ctx).Where("user_id = ?", userID).Find(&res).Error; err != nil {
		return nil, errors.GeneralError("failed to fetch volumes for user: %s", err.Error())
	}
	return res, nil
}

func (v *volumeStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.Volume, *errors.ServiceError) {
	var volumes []*models.Volume
	if err := v.sessionFactory.New(ctx).
		Where("team_id = ?", teamID).
		Order("created_at DESC").
		Find(&volumes).Error; err != nil {
		return nil, errors.GeneralError("failed to list volumes by team: %s", err.Error())
	}
	return volumes, nil
}
