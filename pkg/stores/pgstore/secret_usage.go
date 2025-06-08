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

type dbSecretUsageStore struct {
	sessionFactory db.SessionFactory
}

type SecretUsageStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewSecretUsageStore(spec SecretUsageStoreSpec) stores.SecretUsageStore {
	return &dbSecretUsageStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbSecretUsageStore) Create(ctx context.Context, secretUsage *models.SecretUsage) (*models.SecretUsage, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&secretUsage).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create secret usage: %s", err.Error())
	}
	return d.GetBySecretIDAndStackID(ctx, secretUsage.SecretID, secretUsage.StackID)
}

func (d dbSecretUsageStore) GetBySecretIDAndStackID(ctx context.Context, secretID, stackID string) (*models.SecretUsage, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var secretUsage models.SecretUsage
	err := grm.Model(&models.SecretUsage{}).
		Preload(clause.Associations).
		Where("secret_id = ? AND stack_id = ?", secretID, stackID).
		First(&secretUsage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("secret usage with secret_id '%s' and stack_id '%s' not found", secretID, stackID)
		}
		return nil, errors.GeneralError("failed to fetch secret usage: %s", err.Error())
	}
	return &secretUsage, nil
}

func (d dbSecretUsageStore) GetBySecretID(ctx context.Context, secretID string) ([]*models.SecretUsage, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var secretUsages []*models.SecretUsage
	err := grm.Model(&models.SecretUsage{}).
		Preload(clause.Associations).
		Where("secret_id = ?", secretID).
		Find(&secretUsages).Error
	if err != nil {
		return nil, errors.GeneralError("failed to fetch secret usages by secret_id: %s", err.Error())
	}
	return secretUsages, nil
}

func (d dbSecretUsageStore) GetByStackID(ctx context.Context, stackID string) ([]*models.SecretUsage, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var secretUsages []*models.SecretUsage
	err := grm.Model(&models.SecretUsage{}).
		Preload(clause.Associations).
		Where("stack_id = ?", stackID).
		Find(&secretUsages).Error
	if err != nil {
		return nil, errors.GeneralError("failed to fetch secret usages by stack_id: %s", err.Error())
	}
	return secretUsages, nil
}

func (d dbSecretUsageStore) Delete(ctx context.Context, secretID, stackID string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("secret_id = ? AND stack_id = ?", secretID, stackID).
		Delete(&models.SecretUsage{}).Error
	if err != nil {
		return errors.GeneralError("failed to delete secret usage: %s", err.Error())
	}
	return nil
}

func (d dbSecretUsageStore) DeleteBySecretID(ctx context.Context, secretID string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("secret_id = ?", secretID).
		Delete(&models.SecretUsage{}).Error
	if err != nil {
		return errors.GeneralError("failed to delete secret usages by secret_id: %s", err.Error())
	}
	return nil
}

func (d dbSecretUsageStore) DeleteByStackID(ctx context.Context, stackID string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("stack_id = ?", stackID).
		Delete(&models.SecretUsage{}).Error
	if err != nil {
		return errors.GeneralError("failed to delete secret usages by stack_id: %s", err.Error())
	}
	return nil
}
