package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type volumeMountStore struct {
	sessionFactory db.SessionFactory
}

type VolumeMountStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewVolumeMountStore(spec VolumeMountStoreSpec) stores.VolumeMountStore {
	return &volumeMountStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (v *volumeMountStore) ListByStackResourceID(ctx context.Context, stackResourceID string) ([]*models.VolumeMount, *errors.ServiceError) {
	var volumeMounts []*models.VolumeMount
	if err := v.sessionFactory.New(ctx).Where("stack_resource_id = ?", stackResourceID).Find(&volumeMounts).Error; err != nil {
		return nil, errors.GeneralError("failed to get volume mounts: %s", err.Error())
	}
	return volumeMounts, nil
}

func (v *volumeMountStore) ListBySourceVolumeID(ctx context.Context, sourceVolumeID string) ([]*models.VolumeMount, *errors.ServiceError) {
	var volumeMounts []*models.VolumeMount
	if err := v.sessionFactory.New(ctx).Where("source_volume_id = ?", sourceVolumeID).Find(&volumeMounts).Error; err != nil {
		return nil, errors.GeneralError("failed to get volume mounts: %s", err.Error())
	}
	return volumeMounts, nil
}

func (v *volumeMountStore) ListByStackID(ctx context.Context, stackID string) ([]*models.VolumeMount, *errors.ServiceError) {
	var volumeMounts []*models.VolumeMount
	if err := v.sessionFactory.New(ctx).Where("stack_id = ?", stackID).Find(&volumeMounts).Error; err != nil {
		return nil, errors.GeneralError("failed to get volume mounts: %s", err.Error())
	}
	return volumeMounts, nil
}

func (v *volumeMountStore) Create(ctx context.Context, volumeMount *models.VolumeMount) (*models.VolumeMount, *errors.ServiceError) {
	if err := v.sessionFactory.New(ctx).Create(&volumeMount).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume mount: %s", err.Error())
	}
	return volumeMount, nil
}

// CreateWithTx(ctx context.Context, resource *models.StackResource) (*models.StackResource, *errors.ServiceError)

func (v *volumeMountStore) CreateWithTx(ctx context.Context, volumeMount *models.VolumeMount) (*models.VolumeMount, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.VolumeMount{}).Create(volumeMount).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume mount: %s", err.Error())
	}
	return volumeMount, nil
}

// BulkCreateWithTx(ctx context.Context, volumeMounts []*models.VolumeMount) ([]*models.VolumeMount, *errors.ServiceError)

func (v *volumeMountStore) BulkCreateWithTx(ctx context.Context, volumeMounts []*models.VolumeMount) ([]*models.VolumeMount, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.VolumeMount{}).Create(volumeMounts).Error; err != nil {
		return nil, errors.GeneralError("failed to create volume mounts: %s", err.Error())
	}
	return volumeMounts, nil
}

// DeleteForStackResourceWithTx(ctx context.Context, stackResourceID string)

func (v *volumeMountStore) DeleteForStackResourceWithTx(ctx context.Context, stackResourceID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("stack_resource_id = ?", stackResourceID).Delete(&models.VolumeMount{}).Error; err != nil {
		return errors.GeneralError("failed to delete volume mounts for stack resource: %s", err.Error())
	}
	return nil
}

func (v *volumeMountStore) DeleteForStackResource(ctx context.Context, stackResourceID string) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Where("stack_resource_id = ?", stackResourceID).Delete(&models.VolumeMount{}).Error; err != nil {
		return errors.GeneralError("failed to delete volume mounts for stack resource: %s", err.Error())
	}
	return nil
}

func (v *volumeMountStore) DeleteForStackWithTx(ctx context.Context, stackID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("stack_id = ?", stackID).Delete(&models.VolumeMount{}).Error; err != nil {
		return errors.GeneralError("failed to delete volume mounts for stack: %s", err.Error())
	}
	return nil
}

func (v *volumeMountStore) DeleteForStack(ctx context.Context, stackID string) *errors.ServiceError {
	if err := v.sessionFactory.New(ctx).Where("stack_id = ?", stackID).Delete(&models.VolumeMount{}).Error; err != nil {
		return errors.GeneralError("failed to delete volume mounts for stack: %s", err.Error())
	}
	return nil
}
