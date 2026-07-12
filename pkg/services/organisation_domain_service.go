package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
)

type OrganisationDomainsService interface {
	Create(ctx context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.OrganisationDomain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	Update(ctx context.Context, id string, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.OrganisationDomain, *errors.ServiceError)
	// TODO
	GetDefaultDomainForOrganisation(ctx context.Context, organisationID string) (*models.OrganisationDomain, *errors.ServiceError)
}

type organisationDomainService struct {
	organisationDomainStore stores.OrganisationDomainStore
	stackDomainStore        stores.StackDomainsStore
	logger                  logger.Logger
}

type OrganisationDomainsServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewOrganisationDomainsService(spec OrganisationDomainsServiceSpec) OrganisationDomainsService {
	return &organisationDomainService{
		organisationDomainStore: pgstore.NewOrganisationDomainStore(pgstore.OrganisationDomainStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackDomainStore: pgstore.NewStackDomainsStore(pgstore.StackDomainsStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *organisationDomainService) InternalCreateWithTx(ctx context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	if len(spec.Domain) == 0 {
		return nil, errors.BadRequest("domain is required")
	}
	if !isValidDomain(spec.Domain) {
		return nil, errors.BadRequest("domain is invalid")
	}
	if spec.OrganisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}

	// Verify no duplicate domain exists
	domains, err := s.organisationDomainStore.ListByOrganisationID(ctx, spec.OrganisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to list domains for organisation %s: %v", spec.OrganisationID, err)
		return nil, err
	}

	for _, domain := range domains {
		if domain.Domain == spec.Domain {
			return nil, errors.Conflict("domain '%s' already exists for organisation", spec.Domain)
		}
	}

	createdDomain, err := s.organisationDomainStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to create organisation domain: %v", err)
		return nil, err
	}
	return createdDomain, nil
}

func (s *organisationDomainService) InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	domain, err := s.organisationDomainStore.Get(ctx, id)
	if err != nil {
		return err
	}
	domainsInUse, err := s.stackDomainStore.ListByOrganisationID(ctx, domain.OrganisationID)
	if err != nil {
		return err
	}
	if len(domainsInUse) > 0 {
		return errors.Conflict("cannot delete domain '%s' as it is in use by stacks", domain.Domain)
	}
	return s.organisationDomainStore.DeleteWithTx(ctx, id)
}

func (s *organisationDomainService) Create(ctx context.Context, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	if len(spec.Domain) == 0 {
		return nil, errors.BadRequest("domain is required")
	}
	if !isValidDomain(spec.Domain) {
		return nil, errors.BadRequest("domain is invalid")
	}
	if spec.OrganisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}

	domains, err := s.organisationDomainStore.ListByOrganisationID(ctx, spec.OrganisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to list domains for organisation %s: %v", spec.OrganisationID, err)
		return nil, err
	}

	for _, domain := range domains {
		if domain.Domain == spec.Domain {
			return nil, errors.Conflict("domain '%s' already exists for organisation", spec.Domain)
		}
	}

	return s.organisationDomainStore.Create(ctx, spec)
}

func (s *organisationDomainService) Get(ctx context.Context, id string) (*models.OrganisationDomain, *errors.ServiceError) {
	return s.organisationDomainStore.Get(ctx, id)
}

func (s *organisationDomainService) Delete(ctx context.Context, id string) *errors.ServiceError {
	domain, err := s.organisationDomainStore.Get(ctx, id)
	if err != nil {
		return err
	}
	domainsInUse, err := s.stackDomainStore.ListByOrganisationID(ctx, domain.OrganisationID)
	if err != nil {
		return err
	}
	if len(domainsInUse) > 0 {
		return errors.Conflict("cannot delete domain '%s' as it is in use by stacks", domain.Domain)
	}

	return s.organisationDomainStore.Delete(ctx, id)
}

func (s *organisationDomainService) Update(ctx context.Context, id string, spec *models.OrganisationDomain) (*models.OrganisationDomain, *errors.ServiceError) {
	if len(spec.Domain) == 0 {
		return nil, errors.BadRequest("domain is required")
	}
	if !isValidDomain(spec.Domain) {
		return nil, errors.BadRequest("domain is invalid")
	}

	// Get the existing domain to check if it exists
	existingDomain, err := s.organisationDomainStore.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// If the domain name is changing, check for duplicates
	if existingDomain.Domain != spec.Domain {
		domains, err := s.organisationDomainStore.ListByOrganisationID(ctx, existingDomain.OrganisationID)
		if err != nil {
			s.logger.Error(ctx, "failed to list domains for organisation %s: %v", existingDomain.OrganisationID, err)
			return nil, err
		}

		for _, domain := range domains {
			if domain.Domain == spec.Domain && domain.ID != id {
				return nil, errors.Conflict("domain '%s' already exists for organisation", spec.Domain)
			}
		}
	}

	// Ensure organisation ID cannot be changed
	spec.OrganisationID = existingDomain.OrganisationID

	return s.organisationDomainStore.Update(ctx, id, spec)
}

func (s *organisationDomainService) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.OrganisationDomain, *errors.ServiceError) {
	if organisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}

	return s.organisationDomainStore.ListByOrganisationID(ctx, organisationID)
}

func (s *organisationDomainService) GetDefaultDomainForOrganisation(ctx context.Context, organisationID string) (*models.OrganisationDomain, *errors.ServiceError) {
	if organisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}

	domains, err := s.organisationDomainStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to list domains for organisation %s: %v", organisationID, err)
		return nil, err
	}

	if len(domains) == 0 {
		return nil, errors.NotFound("no domains found for organisation %s", organisationID)
	}

	// In the future we can implement a mechanism to designate a default domain
	// For now we just return the first one
	return domains[0], nil
}
