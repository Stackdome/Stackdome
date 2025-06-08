package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type stackVolumeStore struct {
	sessionFactory db.SessionFactory
	volumeStore    stores.VolumeStore
}

type StackVolumeStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewStackVolumeStore(spec StackVolumeStoreSpec) stores.StackVolumeStore {
	return &stackVolumeStore{
		sessionFactory: spec.SessionFactory,
		volumeStore:    NewVolumeStore(VolumeStoreSpec{SessionFactory: spec.SessionFactory}),
	}
}

func (s *stackVolumeStore) Create(ctx context.Context, sv *models.StackVolume) *errors.ServiceError {
	if len(sv.StackID) == 0 || len(sv.VolumeID) == 0 {
		return errors.BadRequest("stack_id and volume_id are required")
	}
	grm := s.sessionFactory.New(ctx)
	err := grm.Create(&sv).Error
	if err != nil {
		return errors.GeneralError("failed to create stack_volume: %s", err.Error())
	}
	return nil
}

func (s *stackVolumeStore) CreateWithTx(ctx context.Context, sv *models.StackVolume) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Create(&sv).Error; err != nil {
		return errors.GeneralError("failed to create stack_volume: %s", err.Error())
	}
	return nil
}

func (s *stackVolumeStore) Get(ctx context.Context, stackID, volumeID string) (*models.StackVolume, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var sv models.StackVolume
	err := grm.Model(&models.StackVolume{}).Where("stack_id = ? AND volume_id = ?", stackID, volumeID).First(&sv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack_volume with stack_id '%s' and volume_id '%s' not found", stackID, volumeID)
		}
		return nil, errors.GeneralError("failed to fetch stack_volume: %s", err.Error())
	}
	return &sv, nil
}

func (s *stackVolumeStore) Delete(ctx context.Context, stackID, volumeID string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	err := grm.Where("stack_id = ? AND volume_id = ?", stackID, volumeID).Delete(&models.StackVolume{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack_volume with stack_id '%s' and volume_id '%s' not found", stackID, volumeID)
		}
		return errors.GeneralError("failed to delete stack_volume: %s", err.Error())
	}
	return nil
}

func (s *stackVolumeStore) DeleteWithTx(ctx context.Context, stackID, volumeID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	err := tx.Where("stack_id = ? AND volume_id = ?", stackID, volumeID).Delete(&models.StackVolume{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack_volume with stack_id '%s' and volume_id '%s' not found", stackID, volumeID)
		}
		return errors.GeneralError("failed to delete stack_volume: %s", err.Error())
	}
	return nil
}

func (s *stackVolumeStore) GetByVolumeID(ctx context.Context, volumeID string) (*models.StackVolume, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var sv models.StackVolume
	err := grm.Model(&models.StackVolume{}).Where("volume_id = ?", volumeID).First(&sv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack_volume with volume_id '%s' not found", volumeID)
		}
		return nil, errors.GeneralError("failed to fetch stack_volume: %s", err.Error())
	}
	return &sv, nil
}

func (s *stackVolumeStore) ListVolumesByStackID(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var svs []*models.StackVolume
	err := grm.Model(&models.StackVolume{}).Where("stack_id = ?", stackID).Find(&svs).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list stack_volumes: %s", err.Error())
	}
	if len(svs) == 0 {
		return nil, nil
	}
	volumeIDs := make([]string, len(svs))
	for i, sv := range svs {
		volumeIDs[i] = sv.VolumeID
	}
	volumes, serr := s.volumeStore.InternalList(ctx, volumeIDs)
	if serr != nil {
		return nil, errors.GeneralError("failed to list volumes: %s", serr.Error())
	}

	return volumes, nil
}
