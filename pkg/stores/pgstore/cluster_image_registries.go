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

type ClusterImageRegistryStoreSpec struct {
	SessionFactory db.SessionFactory
}

type clusterImageRegistryStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

func NewClusterImageRegistryStore(spec ClusterImageRegistryStoreSpec) stores.ClusterImageRegistryStore {
	return &clusterImageRegistryStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (c *clusterImageRegistryStore) Create(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	if err := c.sessionFactory.New(ctx).Model(&models.ClusterImageRegistry{}).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create cluster image registry: %s", err.Error())
	}
	return c.GetByID(ctx, spec.ID)
}

func (c *clusterImageRegistryStore) CreateWithTx(ctx context.Context, spec *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Model(&models.ClusterImageRegistry{}).Create(spec).Error; err != nil {
		return nil, errors.GeneralError("failed to create cluster image registry: %s", err.Error())
	}
	return c.GetByID(ctx, spec.ID)
}

func (c *clusterImageRegistryStore) GetForOrg(ctx context.Context, orgID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	var registry models.ClusterImageRegistry
	if err := c.sessionFactory.New(ctx).Where("organisation_id = ?", orgID).First(&registry).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("cluster image registry not found")
		}
		return nil, errors.GeneralError("failed to get cluster image registry: %s", err.Error())
	}
	return &registry, nil
}

func (c *clusterImageRegistryStore) GetByID(ctx context.Context, ID string) (*models.ClusterImageRegistry, *errors.ServiceError) {
	var registry models.ClusterImageRegistry
	if err := c.sessionFactory.New(ctx).Where("id = ?", ID).First(&registry).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("cluster image registry not found")
		}
		return nil, errors.GeneralError("failed to get cluster image registry: %s", err.Error())
	}
	return &registry, nil
}

func (c *clusterImageRegistryStore) ListForOrg(ctx context.Context, orgID string) ([]*models.ClusterImageRegistry, *errors.ServiceError) {
	var registries []*models.ClusterImageRegistry
	if err := c.sessionFactory.New(ctx).Where("organisation_id = ?", orgID).Find(&registries).Error; err != nil {
		return nil, errors.GeneralError("failed to list cluster image registries: %s", err.Error())
	}
	return registries, nil
}

func (c *clusterImageRegistryStore) ListByClusterID(ctx context.Context, orgID, clusterID string) ([]*models.ClusterImageRegistry, *errors.ServiceError) {
	var registries []*models.ClusterImageRegistry
	if err := c.sessionFactory.New(ctx).Where("organisation_id = ? AND cluster_id = ?", orgID, clusterID).Find(&registries).Error; err != nil {
		return nil, errors.GeneralError("failed to list cluster image registries: %s", err.Error())
	}
	return registries, nil
}

func (c *clusterImageRegistryStore) UpdateStatus(ctx context.Context, ID string, status *models.ClusterImageRegistryStatus) *errors.ServiceError {
	if err := c.sessionFactory.New(ctx).Model(&models.ClusterImageRegistry{}).Where("id = ?", ID).UpdateColumn("status", status).Error; err != nil {
		return errors.GeneralError("failed to update cluster image registry status: %s", err.Error())
	}
	return nil
}

func (c *clusterImageRegistryStore) Delete(ctx context.Context, ID string) *errors.ServiceError {
	if err := c.sessionFactory.New(ctx).Where("id = ?", ID).Delete(&models.ClusterImageRegistry{}).Error; err != nil {
		return errors.GeneralError("failed to delete cluster image registry: %s", err.Error())
	}
	return nil
}

func (c *clusterImageRegistryStore) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	if err := tx.Where("id = ?", ID).Delete(&models.ClusterImageRegistry{}).Error; err != nil {
		return errors.GeneralError("failed to delete cluster image registry: %s", err.Error())
	}
	return nil
}
