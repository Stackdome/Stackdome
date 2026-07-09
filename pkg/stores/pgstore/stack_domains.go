package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
)

type dbStackDomainsStore struct {
	sessionFactory db.SessionFactory
}

type StackDomainsStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewStackDomainsStore(spec StackDomainsStoreSpec) stores.StackDomainsStore {
	return &dbStackDomainsStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (d dbStackDomainsStore) BulkCreate(ctx context.Context, domains []*models.StackDomain) (models.StackDomainList, *errors.ServiceError) {
	if len(domains) == 0 {
		return nil, errors.BadRequest("domains list is empty")
	}
	for _, domain := range domains {
		if len(domain.Fqdn) == 0 {
			return nil, errors.BadRequest("domain fqdn is required")
		}
		if len(domain.StackID) == 0 {
			return nil, errors.BadRequest("stack id is required")
		}
		if len(domain.StackResourceID) == 0 {
			return nil, errors.BadRequest("stack resource id is required")
		}
		if len(domain.StackResourceName) == 0 {
			return nil, errors.BadRequest("stack resource name is required")
		}
		if len(domain.OrganisationID) == 0 {
			return nil, errors.BadRequest("organisation id is required")
		}
	}
	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domains).Error
	if err != nil {
		return nil, errors.GeneralError("failed to create stack domains: %s", err.Error())
	}
	return domains, nil
}

func (d dbStackDomainsStore) ListByOrganisationID(ctx context.Context, organisationID string) (models.StackDomainList, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("organisation_id = ?", organisationID).Find(&domains).Error
	if err != nil {
		return nil, errors.InternalServerError("failed to list stack domains: %s", err.Error())
	}
	return domains, nil
}

// GetByStackResourceAndPort(ctx context.Context, stackResourceID string, port int) (*models.StackDomain, *errors.ServiceError)
func (d dbStackDomainsStore) GetByStackResourceAndPort(ctx context.Context, stackResourceID string, port int) (*models.StackDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("stack_resource_id = ? AND target_port = ?", stackResourceID, port).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack domain with stack resource id '%s' and port '%d' not found", stackResourceID, port)
		}
		return nil, errors.InternalServerError("failed to fetch stack domain by stack resource id and port: %s", err.Error())
	}
	return &domain, nil
}

// DeleteForStackResourceAndPort(ctx context.Context, stackResourceID string, port int) *errors.ServiceError
func (d dbStackDomainsStore) DeleteForStackResourceAndPort(ctx context.Context, stackResourceID string, port int) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("stack_resource_id = ? AND target_port = ?", stackResourceID, port).Delete(&models.StackDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack domain with stack resource id '%s' and port '%d' not found", stackResourceID, port)
		}
		return errors.InternalServerError("failed to delete stack domain: %s", err.Error())
	}
	return nil
}

// DeleteForStackResourceAndPortWithTx(ctx context.Context, stackResourceID string, port int) *errors.ServiceError

func (d dbStackDomainsStore) DeleteForStackResourceAndPortWithTx(ctx context.Context, stackResourceID string, port int) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.InternalServerError("transaction not found in context")
	}
	err := tx.Where("stack_resource_id = ? AND target_port = ?", stackResourceID, port).Delete(&models.StackDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack domain with stack resource id '%s' and port '%d' not found", stackResourceID, port)
		}
		return errors.InternalServerError("failed to delete stack domain: %s", err.Error())
	}
	return nil
}

func (d dbStackDomainsStore) Create(ctx context.Context, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	if len(domain.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}
	if len(domain.StackID) == 0 {
		return nil, errors.BadRequest("stack id is required")
	}
	if len(domain.StackResourceID) == 0 {
		return nil, errors.BadRequest("stack resource id is required")
	}
	if len(domain.StackResourceName) == 0 {
		return nil, errors.BadRequest("stack resource name is required")
	}
	if len(domain.OrganisationID) == 0 {
		return nil, errors.BadRequest("organisation id is required")
	}

	grm := d.sessionFactory.New(ctx)
	err := grm.Create(&domain).Error
	if err != nil {
		return nil, errors.InternalServerError("failed to create stack domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbStackDomainsStore) CreateWithTx(ctx context.Context, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return nil, errors.InternalServerError("transaction not found in context")
	}
	if err := tx.Create(&domain).Error; err != nil {
		return nil, errors.InternalServerError("failed to create stack domain: %s", err.Error())
	}
	return d.Get(ctx, domain.ID)
}

func (d dbStackDomainsStore) Get(ctx context.Context, id string) (*models.StackDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("id = ?", id).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack domain with id '%s' not found", id)
		}
		return nil, errors.InternalServerError("failed to fetch stack domain: %s", err.Error())
	}
	return &domain, nil
}

func (d dbStackDomainsStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := d.sessionFactory.New(ctx)
	err := grm.Where("id = ?", id).Delete(&models.StackDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack domain with id '%s' not found", id)
		}
		return errors.InternalServerError("failed to delete stack domain: %s", err.Error())
	}
	return nil
}

func (d dbStackDomainsStore) DeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return errors.InternalServerError("transaction not found in context")
	}
	err := tx.Where("id = ?", id).Delete(&models.StackDomain{}).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("stack domain with id '%s' not found", id)
		}
		return errors.InternalServerError("failed to delete stack domain: %s", err.Error())
	}
	return nil
}

func (d dbStackDomainsStore) Update(ctx context.Context, id string, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	err := grm.Model(&models.StackDomain{}).Where("id = ?", id).Updates(domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack domain with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update stack domain: %s", err.Error())
	}
	return d.Get(ctx, id)
}

func (d dbStackDomainsStore) ListByStackID(ctx context.Context, stackID string) (models.StackDomainList, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("stack_id = ?", stackID).Find(&domains).Error
	if err != nil {
		return nil, errors.InternalServerError("failed to list stack domains: %s", err.Error())
	}
	return domains, nil
}

func (d dbStackDomainsStore) ListByStackResourceID(ctx context.Context, stackResourceID string) (models.StackDomainList, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("stack_resource_id = ?", stackResourceID).Find(&domains).Error
	if err != nil {
		return nil, errors.InternalServerError("failed to list stack domains by resource id: %s", err.Error())
	}
	return domains, nil
}

func (d dbStackDomainsStore) GetByFqdn(ctx context.Context, fqdn string) (*models.StackDomain, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domain models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("fqdn = ?", fqdn).First(&domain).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("stack domain with fqdn '%s' not found", fqdn)
		}
		return nil, errors.InternalServerError("failed to fetch stack domain by fqdn: %s", err.Error())
	}
	return &domain, nil
}

func (d dbStackDomainsStore) ListByFqdnPrefix(ctx context.Context, prefix string) (models.StackDomainList, *errors.ServiceError) {
	grm := d.sessionFactory.New(ctx)
	var domains []*models.StackDomain
	err := grm.Model(&models.StackDomain{}).Where("fqdn LIKE ?", prefix+"%").Find(&domains).Error
	if err != nil {
		return nil, errors.InternalServerError("failed to list stack domains by fqdn prefix: %s", err.Error())
	}
	return domains, nil
}
