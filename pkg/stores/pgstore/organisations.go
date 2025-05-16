package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
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

func (d dbOrganisationStore) Get(ctx context.Context, id string) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var org models.Organisation
	err := grm.Model(&models.Organisation{}).Where("id = ?", id).First(&org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("organisation with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch organisation: %s", err.Error())
	}
	return &org, nil
}

func (d dbOrganisationStore) GetDefaultOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var org models.Organisation
	err := grm.Model(&models.Organisation{}).Where("\"default\" = ?", true).First(&org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("default organisation not found")
		}
		return nil, errors.GeneralError("failed to fetch default organisation: %s", err.Error())
	}
	return &org, nil
}

func (d dbOrganisationStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	tx := grm.Begin()
	err := tx.Where("id = ?", id).Delete(&models.Organisation{}).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("organisation with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete organisation: %s", err.Error())
	}

	err = tx.Where("owner_id = ? AND owner_type = ?", id, models.OwnerTypeOrganisation).Delete(&models.Domain{}).Error
	if err != nil {
		tx.Rollback()
		return errors.GeneralError("failed to delete domains for organisation: %s", err.Error())
	}
	tx.Commit()
	return nil
}

func (d dbOrganisationStore) Update(ctx context.Context, id string, org *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.Organisation{}).Where("id = ?", id).Updates(org).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("organisation with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update organisation: %s", err.Error())
	}
	return d.Get(ctx, id)
}
