package pgstore

import (
	"context"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type dbOrganisationStore struct {
	sessionFactory db.SessionFactory
}

type OrganisationStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewOrganisationStore(spec OrganisationStoreSpec) stores.OrganisationStore {
	return &dbOrganisationStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbOrganisationStore) Create(ctx context.Context, org *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&org).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create organisation: %s", err.Error())
	}
	return d.Get(ctx, org.ID)
}

func (d dbOrganisationStore) OrganisationNameExists(ctx context.Context, name string) (bool, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var count int64
	err := grm.Model(&models.Organisation{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false, errors.GeneralError("failed to check if organisation name exists: %s", err.Error())
	}
	return count > 0, nil
}

func (d dbOrganisationStore) Get(ctx context.Context, id string) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var org models.Organisation
	err := grm.Model(&models.Organisation{}).Preload(clause.Associations).Where("id = ?", id).First(&org).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("organisation with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch organisation: %s", err.Error())
	}
	return &org, nil
}

func (d dbOrganisationStore) GetPlatformOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var org models.Organisation
	err := grm.Model(&models.Organisation{}).Preload(clause.Associations).Where("platform = ?", true).First(&org).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("platform organisation not found")
		}
		return nil, errors.GeneralError("failed to fetch platform organisation: %s", err.Error())
	}
	return &org, nil
}

func (d dbOrganisationStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	tx := grm.Begin()
	err := tx.Where("id = ?", id).Delete(&models.Organisation{}).Error
	if err != nil {
		tx.Rollback()
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("organisation with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete organisation: %s", err.Error())
	}
	return nil
}

func (d dbOrganisationStore) Update(ctx context.Context, id string, org *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.Organisation{}).Where("id = ?", id).Updates(org).Error
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("organisation with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update organisation: %s", err.Error())
	}
	return d.Get(ctx, id)
}
