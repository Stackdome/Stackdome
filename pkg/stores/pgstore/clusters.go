package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dbClusterStore struct {
	sessionFactory db.SessionFactory
	atomicExecutor
}

type ClusterStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewClusterStore(spec ClusterStoreSpec) stores.ClusterStore {
	return &dbClusterStore{
		sessionFactory: spec.SessionFactory,
		atomicExecutor: atomicExecutor{sessionFactory: spec.SessionFactory},
	}
}

func (d dbClusterStore) GetByClusterUrl(ctx context.Context, clusterURL string) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("cluster_url = ?", clusterURL).Preload(clause.Associations).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("cluster with url '%s' not found", clusterURL)
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) ListAll(ctx context.Context) ([]*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res []*models.Cluster
	err := grm.Model(&models.Cluster{}).Preload(clause.Associations).Find(&res).Error
	if err != nil {
		return nil, errors.GeneralError("failed to fetch clusters: %s", err.Error())
	}
	return res, nil
}

func (d dbClusterStore) Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Omit(clause.Associations).Create(&cluster).Error
	if err != nil {
		return nil, errors.GeneralError("failed to add cluster: %s", err.Error())
	}
	return d.Get(ctx, cluster.ID)
}

func (d dbClusterStore) CreateWithTx(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	err := tx.Omit(clause.Associations).Create(&cluster).Error
	if err != nil {
		return nil, errors.GeneralError("failed to add cluster: %s", err.Error())
	}
	return d.Get(ctx, cluster.ID)
}

func (d dbClusterStore) PersistManagerState(ctx context.Context, id string, running bool) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.Cluster{}).Where("id = ?", id).Update("manager_running", running).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("cluster with id '%s' not found", id)
		}
		return errors.GeneralError("failed to update cluster: %s", err.Error())
	}
	return nil
}

func (d dbClusterStore) Get(ctx context.Context, id string) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Preload(clause.Associations).Where("id = ?", id).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("cluster with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("organisation_id = ?", orgID).Preload(clause.Associations).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("cluster for organisation '%s' not found", orgID)
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("\"default\" = ?", true).Preload(clause.Associations).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("default cluster not found")
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("id = ?", id).Delete(&models.Cluster{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("cluster with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete cluster: %s", err.Error())
	}
	return nil
}
