package services

import (
	"context"
	"regexp"
	"strings"

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
	logger            logger.Logger
}

func NewOrganisationService(spec OrganisationServiceSpec) OrganisationService {
	return &organisationService{
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type OrganisationServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
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
	org, err := s.organisationStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create organisation: %v", err)
		return nil, err
	}
	return org, nil
}

func (s *organisationService) Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError) {
	org, err := s.organisationStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return nil, err
	}
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
	if len(spec.DomainName) == 0 || !IsValidDomain(spec.DomainName) {
		return nil, errors.BadRequest("domain name is required and must be a valid domain")
	}

	updatedOrg, err := s.Get(ctx, ID)
	if err != nil {
		return nil, err
	}
	updatedOrg.DomainName = spec.DomainName
	updatedOrg.Name = spec.Name
	// We dont want to set the default organisation by orgs created through the external API.
	updatedOrg.Default = false

	org, err := s.organisationStore.Update(ctx, ID, updatedOrg)
	if err != nil {
		s.logger.Errorf("failed to update organisation: %v", err)
		return nil, err
	}
	return org, nil
}

func IsValidDomain(domain string) bool {
	if strings.TrimSpace(domain) == "" {
		return false
	}

	// Match RFC standards for DNS domain names
	pattern := `^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z0-9][a-zA-Z0-9-]{0,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`

	regex := regexp.MustCompile(pattern)
	return regex.MatchString(domain)
}
