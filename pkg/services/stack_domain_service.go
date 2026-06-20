package services

import (
	"context"
	"regexp"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

var domainRegex = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

//go:generate mockgen -source=stack_domain_service.go -destination=../mocks/mock_stack_domains_service.go -package=mocks

type StackDomainsService interface {
	Create(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.StackDomain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	GetByFqdn(ctx context.Context, fqdn string) (*models.StackDomain, *errors.ServiceError)
	ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.StackDomain, *errors.ServiceError)
	DomainToUseForStack(ctx context.Context, stack *models.Stack) (*models.OrganisationDomain, *errors.ServiceError)
	PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError
	InternalDeleteDomainsForResourceWithTx(ctx context.Context, resourceID string) *errors.ServiceError
}

type stackDomainService struct {
	domainsStore            stores.StackDomainsStore
	organisationDomainStore stores.OrganisationDomainStore
	logger                  logger.Logger
}

type StackDomainsServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewStackDomainsService(spec StackDomainsServiceSpec) StackDomainsService {
	return &stackDomainService{
		domainsStore: pgstore.NewStackDomainsStore(pgstore.StackDomainsStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		organisationDomainStore: pgstore.NewOrganisationDomainStore(pgstore.OrganisationDomainStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *stackDomainService) InternalCreateWithTx(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	if len(spec.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}
	if !isValidDomain(spec.Fqdn) {
		return nil, errors.BadRequest("domain fqdn is invalid")
	}
	existing, err := s.domainsStore.GetByFqdn(ctx, spec.Fqdn)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to check if domain exists: %v", err)
		return nil, err
	}
	if existing != nil {
		return nil, errors.Conflict("domain with fqdn '%s' already exists", spec.Fqdn)
	}
	if spec.OrganisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}
	if spec.StackID == "" {
		return nil, errors.BadRequest("stack id is required")
	}
	if spec.StackResourceID == "" {
		return nil, errors.BadRequest("stack resource id is required")
	}
	if spec.StackResourceName == "" {
		return nil, errors.BadRequest("stack resource name is required")
	}
	if spec.TargetPort == 0 {
		return nil, errors.BadRequest("target port is required")
	}
	createdDomain, err := s.domainsStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create domain: %v", err)
		return nil, err
	}
	return createdDomain, nil
}

func (s *stackDomainService) InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	return s.domainsStore.DeleteWithTx(ctx, id)
}

func (s *stackDomainService) DomainToUseForStack(ctx context.Context, stack *models.Stack) (*models.OrganisationDomain, *errors.ServiceError) {
	domains, err := s.organisationDomainStore.ListByOrganisationID(ctx, stack.OrganisationID)
	if err != nil {
		s.logger.Errorf("failed to list domains for organisation %s: %v", stack.OrganisationID, err)
		return nil, err
	}
	if len(domains) == 0 {
		return nil, errors.NotFound("no domains found for organisation %s", stack.OrganisationID)
	}

	// In the future we can pin a domain to a project or group.
	// For now we just return the first one
	return domains[0], nil
}

func (s *stackDomainService) PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	if !stackResourceHasExposedPorts(resource.Ports) {
		return s.deleteStaleDomainsForResourceWithTx(ctx, resource)
	}

	domainToUse, err := s.DomainToUseForStack(ctx, stack)
	if err != nil {
		return err
	}

	AssignExposedPortFQDNs(stack.ID, resource.Name, domainToUse.Domain, resource.Ports)

	for j := range resource.Ports {
		if !resource.Ports[j].ExposedToPublic {
			continue
		}

		domain := &models.StackDomain{
			Fqdn:              resource.Ports[j].ExposedFqdn,
			OrganisationID:    stack.OrganisationID,
			StackID:           stack.ID,
			StackResourceID:   resource.ID,
			StackResourceName: resource.Name,
			TargetPort:        resource.Ports[j].Number,
		}

		present, err := s.domainPresentForStackResourceAndPort(ctx, resource.ID, resource.Ports[j].Number)
		if err != nil {
			return err
		}
		if present {
			continue
		}
		if _, cerr := s.InternalCreateWithTx(ctx, domain); cerr != nil {
			if cerr.Code == errors.ErrorConflict {
				existing, _ := s.GetByFqdn(ctx, resource.Ports[j].ExposedFqdn)
				if existing != nil {
					return errors.Conflict(
						"domain '%s' is already in use by resource '%s' in another stack (id: %s)",
						resource.Ports[j].ExposedFqdn, existing.StackResourceName, existing.StackID)
				}
				return errors.Conflict("domain '%s' is already in use", resource.Ports[j].ExposedFqdn)
			}
			return errors.GeneralError(
				"failed to create domain for stack resource '%s': %s", resource.Name, cerr.Error())
		}
	}

	return s.deleteStaleDomainsForResourceWithTx(ctx, resource)
}

func (s *stackDomainService) InternalDeleteDomainsForResourceWithTx(ctx context.Context, resourceID string) *errors.ServiceError {
	existingDomains, err := s.domainsStore.ListByStackResourceID(ctx, resourceID)
	if err != nil {
		return err
	}
	for _, existingDomain := range existingDomains {
		if err := s.domainsStore.DeleteWithTx(ctx, existingDomain.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *stackDomainService) deleteStaleDomainsForResourceWithTx(ctx context.Context, resource *models.StackResource) *errors.ServiceError {
	existingDomains, err := s.domainsStore.ListByStackResourceID(ctx, resource.ID)
	if err != nil {
		return err
	}
	if len(existingDomains) == 0 {
		return nil
	}

	currentPortDomainMap := exposedPortNumberSet(resource.Ports)
	for _, existingDomain := range existingDomains {
		if _, ok := currentPortDomainMap[existingDomain.TargetPort]; !ok {
			if err := s.domainsStore.DeleteWithTx(ctx, existingDomain.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *stackDomainService) domainPresentForStackResourceAndPort(ctx context.Context, stackResourceID string, port int) (bool, *errors.ServiceError) {
	domain, err := s.domainsStore.GetByStackResourceAndPort(ctx, stackResourceID, port)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			return false, nil
		}
		return false, err
	}
	return domain != nil, nil
}

func (s *stackDomainService) Create(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	if len(spec.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}
	existing, err := s.domainsStore.GetByFqdn(ctx, spec.Fqdn)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to check if domain exists: %v", err)
		return nil, err
	}
	if !isValidDomain(spec.Fqdn) {
		return nil, errors.BadRequest("domain fqdn is invalid")
	}
	if existing != nil {
		return nil, errors.Conflict("domain with fqdn '%s' already exists", spec.Fqdn)
	}
	if spec.OrganisationID == "" {
		return nil, errors.BadRequest("organisation id is required")
	}
	if spec.StackID == "" {
		return nil, errors.BadRequest("stack id is required")
	}
	if spec.StackResourceID == "" {
		return nil, errors.BadRequest("stack resource id is required")
	}
	if spec.StackResourceName == "" {
		return nil, errors.BadRequest("stack resource name is required")
	}
	if spec.TargetPort == 0 {
		return nil, errors.BadRequest("target port is required")
	}

	if existing != nil {
		return nil, errors.Conflict("domain with fqdn '%s' already exists", spec.Fqdn)
	}

	return s.domainsStore.Create(ctx, spec)
}

func (s *stackDomainService) Get(ctx context.Context, id string) (*models.StackDomain, *errors.ServiceError) {
	return s.domainsStore.Get(ctx, id)
}

func (s *stackDomainService) Delete(ctx context.Context, id string) *errors.ServiceError {
	return s.domainsStore.Delete(ctx, id)
}

func (s *stackDomainService) GetByFqdn(ctx context.Context, fqdn string) (*models.StackDomain, *errors.ServiceError) {
	return s.domainsStore.GetByFqdn(ctx, fqdn)
}

func (s *stackDomainService) ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.StackDomain, *errors.ServiceError) {
	return s.domainsStore.ListByFqdnPrefix(ctx, prefix)
}

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}
