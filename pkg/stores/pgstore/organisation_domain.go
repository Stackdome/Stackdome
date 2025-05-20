package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type dbOrganisationDomainStore struct {
	sessionFactory db.SessionFactory
}

type OrganisationDomainStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewOrganisationDomainStore(spec OrganisationDomainStoreSpec) stores.OrganisationDomainStore {
	return &dbOrganisationDomainStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbOrganisationDomainStore) BulkCreate(ctx context.Context, domains []*models.OrganisationDomain) ([]*models.OrganisationDomain, *errors.ServiceError) {
	if len(domains) == 0 {
		return nil, errors.BadRequest("domains list is empty")
	}
	for _, domain := range domains {
		if len(domain.Domain) == 0 {
			return nil, errors.BadRequest("domain url is required")
		}
		if len(domain.OrganisationID) == 0 {
			return nil, errors.BadRequest("organisation id is required")
		}
	}
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create organisation domains: %s", err.Error())
	}
	return domains, nil
}

func (d dbOrganisationDomainStore) Create(ctx context.Context, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	if len(domain.Domain) == 0 {
		return nil, errors.BadRequest("domain url is required")
	}
	if len(domain.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}

	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domain).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create organisation domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbOrganisationDomainStore) CreateWithTx(ctx context.Context, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if len(domain.Domain) == 0 {
		return nil, errors.BadRequest("domain url is required")
	}
	if len(domain.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}

	if err := tx.Create(&domain).Error; err != nil {
		return nil, errors.GeneralError("failed to create organisation domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbOrganisationDomainStore) Get(ctx context.Context, id string) (*models.OrganisationDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.OrganisationDomain
	err := grm.Model(&models.OrganisationDomain{}).Where("id = ?", id).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("organisation domain with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch organisation domain: %s", err.Error())
	}
	return &domain, nil
}

func (d dbOrganisationDomainStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("id = ?", id).Delete(&models.OrganisationDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("organisation domain with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete organisation domain: %s", err.Error())
	}
	return nil
}

func (d dbOrganisationDomainStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	err := tx.Where("id = ?", id).Delete(&models.OrganisationDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("organisation domain with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete organisation domain: %s", err.Error())
	}
	return nil
}

func (d dbOrganisationDomainStore) Update(ctx context.Context, id string, domain *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.OrganisationDomain{}).Where("id = ?", id).Updates(domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("organisation domain with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update organisation domain: %s", err.Error())
	}
	return d.Get(ctx, id)
}

func (d dbOrganisationDomainStore) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.OrganisationDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.OrganisationDomain
	err := grm.Model(&models.OrganisationDomain{}).Where("organisation_id = ?", organisationID).Find(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list organisation domains: %s", err.Error())
	}
	return domains, nil
}
