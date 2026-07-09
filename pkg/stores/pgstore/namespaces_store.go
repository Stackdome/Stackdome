package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type dbNamespacesStore struct {
	sessionFactory db.SessionFactory
}

type NamespacesStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewNamespacesStore(spec NamespacesStoreSpec) stores.NamespacesStore {
	return &dbNamespacesStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbNamespacesStore) Create(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	if len(ns.Name) == 0 {
		return nil, errors.BadRequest("namespace name is required")
	}
	if len(ns.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&ns).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create namespace: %s", err.Error())
	}
	return d.Get(ctx, ns.ID)
}

func (d dbNamespacesStore) CreateWithTx(ctx context.Context, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Create(&ns).Error; err != nil {
		return nil, errors.GeneralError("failed to create namespace: %s", err.Error())
	}
	return d.Get(ctx, ns.ID)
}

func (d dbNamespacesStore) Get(ctx context.Context, id string) (*models.Namespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var ns models.Namespace
	err := grm.Model(&models.Namespace{}).Where("id = ?", id).First(&ns).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("namespace with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch namespace: %s", err.Error())
	}
	return &ns, nil
}

func (d dbNamespacesStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("id = ?", id).Delete(&models.Namespace{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("namespace with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete namespace: %s", err.Error())
	}
	return nil
}

func (d dbNamespacesStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	err := tx.Where("id = ?", id).Delete(&models.Namespace{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("namespace with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete namespace: %s", err.Error())
	}
	return nil
}

func (d dbNamespacesStore) Update(ctx context.Context, id string, ns *models.Namespace) (*models.Namespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.Namespace{}).Where("id = ?", id).Updates(ns).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("namespace with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update namespace: %s", err.Error())
	}
	return d.Get(ctx, id)
}

func (d dbNamespacesStore) ListByOrganisation(ctx context.Context, organisationID string) ([]*models.Namespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var nss []*models.Namespace
	err := grm.Model(&models.Namespace{}).Where("organisation_id = ?", organisationID).Find(&nss).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list namespaces: %s", err.Error())
	}
	return nss, nil
}

func (d dbNamespacesStore) ListByStack(ctx context.Context, stackID string) ([]*models.Namespace, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var nss []*models.Namespace
	err := grm.Model(&models.Namespace{}).Where("stack_id = ?", stackID).Find(&nss).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list namespaces by stack: %s", err.Error())
	}
	return nss, nil
}
