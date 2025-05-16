package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type dbDomainsStore struct {
	sessionFactory db.SessionFactory
}

type DomainsStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewDomainsStore(spec DomainsStoreSpec) stores.DomainsStore {
	return &dbDomainsStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbDomainsStore) DeleteForOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Delete(&models.Domain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return errors.GeneralError("failed to delete domains for owner: %s", err.Error())
	}
	return nil
}

func (d dbDomainsStore) BulkCreate(ctx context.Context, domains []*models.Domain) ([]*models.Domain, *errors.ServiceError) {
	if len(domains) == 0 {
		return nil, errors.BadRequest("domains list is empty")
	}
	for _, domain := range domains {
		if len(domain.Fqdn) == 0 {
			return nil, errors.BadRequest("domain fqdn is required")
		}
		if len(domain.OwnerID) == 0 {
			return nil, errors.BadRequest("domain owner id is required")
		}
		if domain.OwnerType != models.OwnerTypeStackResource && domain.OwnerType != models.OwnerTypeOrganisation {
			return nil, errors.BadRequest("domain owner type must be either user or organisation")
		}
	}
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create domains: %s", err.Error())
	}
	return domains, nil
}

func (d dbDomainsStore) Create(ctx context.Context, domain *models.Domain) (*models.Domain, *errors.ServiceError) {
	if len(domain.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}
	if len(domain.OwnerID) == 0 {
		return nil, errors.BadRequest("domain owner id is required")
	}
	if domain.OwnerType != models.OwnerTypeStackResource && domain.OwnerType != models.OwnerTypeOrganisation {
		return nil, errors.BadRequest("domain owner type must be either user or organisation")
	}

	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domain).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbDomainsStore) CreateWithTx(ctx context.Context, domain *models.Domain) (*models.Domain, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.GeneralError("transaction not found in context")
	}
	if err := tx.Create(&domain).Error; err != nil {
		return nil, errors.GeneralError("failed to create domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbDomainsStore) Get(ctx context.Context, id string) (*models.Domain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.Domain
	err := grm.Model(&models.Domain{}).Where("id = ?", id).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("domain with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch domain: %s", err.Error())
	}
	return &domain, nil
}

func (d dbDomainsStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("id = ?", id).Delete(&models.Domain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("domain with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete domain: %s", err.Error())
	}
	return nil
}

func (d dbDomainsStore) DeleteForOwnerWithTx(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	err := tx.Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Delete(&models.Domain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return errors.GeneralError("failed to delete domains for owner: %s", err.Error())
	}
	return nil
}

func (d dbDomainsStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.GeneralError("transaction not found in context")
	}
	err := tx.Where("id = ?", id).Delete(&models.Domain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("domain with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete domain: %s", err.Error())
	}
	return nil
}

func (d dbDomainsStore) Update(ctx context.Context, id string, domain *models.Domain) (*models.Domain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.Domain{}).Where("id = ?", id).Updates(domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("domain with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update domain: %s", err.Error())
	}
	return d.Get(ctx, id)
}

func (d dbDomainsStore) ListByOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) ([]*models.Domain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.Domain
	err := grm.Model(&models.Domain{}).Where("owner_id = ? AND owner_type = ?", ownerID, ownerType).Find(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list domains: %s", err.Error())
	}
	return domains, nil
}

func (d dbDomainsStore) GetByFqdn(ctx context.Context, fqdn string) (*models.Domain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.Domain
	err := grm.Model(&models.Domain{}).Where("fqdn = ?", fqdn).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("domain with fqdn '%s' not found", fqdn)
		}
		return nil, errors.GeneralError("failed to fetch domain by fqdn: %s", err.Error())
	}
	return &domain, nil
}

func (d dbDomainsStore) ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.Domain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.Domain
	err := grm.Model(&models.Domain{}).Where("fqdn LIKE ?", prefix+"%").Find(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to list domains by fqdn prefix: %s", err.Error())
	}

	return domains, nil
}
