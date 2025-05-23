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
	organisationStore         stores.OrganisationStore
	organisationDomainService OrganisationDomainsService
	stackQueryService         StackQueryService
	logger                    logger.Logger
}

func NewOrganisationService(spec OrganisationServiceSpec) OrganisationService {
	return &organisationService{
		organisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackQueryService:         spec.StackQueryService,
		organisationDomainService: spec.OrganisationDomainService,
		logger:                    spec.Logger,
	}
}

type OrganisationServiceSpec struct {
	SessionFactory            db.SessionFactory
	OrganisationDomainService OrganisationDomainsService
	StackQueryService         StackQueryService
	Logger                    logger.Logger
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
		domain.OrganisationID = org.ID
		if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
			return nil, err
		}
	}

	return s.Get(ctx, org.ID)
}

func (s *organisationService) Get(ctx context.Context, ID string) (*models.Organisation, *errors.ServiceError) {
	org, err := s.organisationStore.Get(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get organisation: %v", err)
		return nil, err
	}
	return org, nil
}

// TODO: Org is the root of almost everything, so we need to be careful when deleting it.
func (s *organisationService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	stacks, err := s.stackQueryService.GetStacksByOrganisationID(ctx, ID)
	if err != nil {
		return err
	}
	if len(stacks) > 0 {
		return errors.BadRequest("cannot delete organisation with stacks")
	}
	err = s.organisationStore.Delete(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to delete organisation: %v", err)
		return err
	}
	return nil
}

// Update updates an organisation
func (s *organisationService) Update(ctx context.Context, ID string, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
	existing, err := s.Get(ctx, ID)
	if err != nil {
		return nil, err
	}

	if spec.Name != "" && existing.Name != spec.Name {
		return nil, errors.BadRequest("organisation name cannot be updated")
	}

	spec.Name = existing.Name
	org, err := s.organisationStore.Update(ctx, ID, spec)
	if err != nil {
		s.logger.Errorf("failed to update organisation: %v", err)
		return nil, err
	}

	existingDomains, err := s.organisationDomainService.ListByOrganisationID(ctx, ID)
	if err != nil {
		return nil, err
	}

	existingDomainsMap := make(map[string]struct{})
	for _, domain := range existingDomains {
		existingDomainsMap[domain.Domain] = struct{}{}
	}

	for _, domain := range spec.Domains {
		if _, exists := existingDomainsMap[domain.Domain]; !exists {
			domain.OrganisationID = org.ID
			if _, err := s.organisationDomainService.Create(ctx, domain); err != nil {
				return nil, err
			}
		}
	}

	currentDomainsMap := make(map[string]struct{})
	for _, domain := range spec.Domains {
		currentDomainsMap[domain.Domain] = struct{}{}
	}
	for _, domain := range existingDomains {
		if _, exists := currentDomainsMap[domain.Domain]; !exists {
			if err := s.organisationDomainService.Delete(ctx, domain.ID); err != nil {
				return nil, err
			}
		}
	}
	return s.Get(ctx, org.ID)
}
