package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type OrganisationService interface {
	GetDefaultOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError)
	Create(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
	Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError)
}

type organisationService struct {
	organisationStore stores.OrganisationStore
	domainNameService DomainsService
	logger            logger.Logger
}

func NewOrganisationService(spec OrganisationServiceSpec) OrganisationService {
	return &organisationService{
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		domainNameService: spec.DomainNameService,
		logger:            spec.Logger,
	}
}

type OrganisationServiceSpec struct {
	SessionFactory    db.SessionFactory
	DomainNameService DomainsService
	Logger            logger.Logger
}

func (s *organisationService) GetDefaultOrg(ctx context.Context) (*models.Organisation, *errors.ServiceError) {
	org, err := s.organisationStore.GetDefaultOrg(ctx)
	if err != nil {
		s.logger.Errorf("failed to get default organisation: %v", err)
		return nil, err
	}
	return org, nil
}

func (s *organisationService) Create(ctx context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	if len(spec.Name) == 0 {
		return nil, errors.BadRequest("organisation name is required")
	}

	org, err := s.organisationStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create organisation: %v", err)
		return nil, err
	}

	for _, domain := range spec.Domains {
		domain.OwnerID = org.ID
		domain.OwnerType = models.OwnerTypeOrganisation
		createdDomain, err := s.domainNameService.Create(ctx, domain)
		if err != nil {
			s.logger.Errorf("failed to create domain: %v", err)
			return nil, err
		}
		org.Domains = append(org.Domains, createdDomain)
	}

	return org, nil
}

func (s *organisationService) Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError) {
	org, err := s.organisationStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return nil, err
	}
	domains, err := s.domainNameService.ListByOwner(ctx, ID, models.OwnerTypeOrganisation)
	if err != nil {
		s.logger.Errorf("failed to list domains: %v", err)
		return nil, err
	}
	org.Domains = domains
	return org, nil
}

func (s *organisationService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	err := s.organisationStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete organisation: %v", err)
		return err
	}
	return nil
}

// Update updates an organisation
func (s *organisationService) Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	updatedOrg, err := s.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	updatedOrg.Name = spec.Name
	// We dont want to set the default organisation by orgs created through the external API.
	updatedOrg.Default = false

	org, err := s.organisationStore.Update(ctx, ID, updatedOrg)
	if err != nil {
		s.logger.Errorf("failed to update organisation: %v", err)
		return nil, err
	}

	for _, domain := range spec.Domains {
		existing, err := s.domainNameService.GetByFqdn(ctx, domain.Fqdn)
		if err != nil && err.Code != errors.ErrorNotFound {
			s.logger.Errorf("failed to check if domain exists: %v", err)
			return nil, err
		}
		if existing != nil {
			continue
		}
		domain.OwnerID = ID
		domain.OwnerType = models.OwnerTypeOrganisation
		_, err = s.domainNameService.Create(ctx, domain)
		if err != nil {
			s.logger.Errorf("failed to create domain: %v", err)
			return nil, err
		}
	}
	domains, err := s.domainNameService.ListByOwner(ctx, ID, models.OwnerTypeOrganisation)
	if err != nil {
		s.logger.Errorf("failed to list domains: %v", err)
		return nil, err
	}
	org.Domains = domains
	return org, nil
}
