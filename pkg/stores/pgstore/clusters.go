package pgstore

import (
	"context"

	"github.com/ashishmax31/soradev-api-server/pkg/db"
	"github.com/ashishmax31/soradev-api-server/pkg/errors"
	"github.com/ashishmax31/soradev-api-server/pkg/models"
	"github.com/ashishmax31/soradev-api-server/pkg/stores"
	"gorm.io/gorm"
)

type dbClusterStore struct {
	sessionFactory db.SessionFactory
}

type ClusterStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewClusterStore(spec ClusterStoreSpec) stores.ClusterStore {
	return &dbClusterStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbClusterStore) Create(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&cluster).Error
	if err != nil {
		return nil, errors.GeneralError("failed to add cluster: %s", err.Error())
	}
	return d.Get(ctx, cluster.ID)
}

func (d dbClusterStore) Get(ctx context.Context, id string) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("id = ?", id).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("cluster with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) GetClusterForOrg(ctx context.Context, orgID int) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("organisation_id = ?", orgID).First(&res).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("cluster for organisation '%d' not found", orgID)
		}
		return nil, errors.GeneralError("failed to fetch cluster: %s", err.Error())
	}
	return &res, nil
}

func (d dbClusterStore) GetDefaultCluster(ctx context.Context) (*models.Cluster, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var res models.Cluster
	err := grm.Model(&models.Cluster{}).Where("\"default\" = ?", true).First(&res).Error
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
