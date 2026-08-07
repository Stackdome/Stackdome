package services

import (
	"context"
	"regexp"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
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
	PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError
	InternalDeleteDomainsForResourceWithTx(ctx context.Context, resourceID string) *errors.ServiceError
}

type stackDomainService struct {
	domainsStore            stores.StackDomainsStore
	organisationDomainStore stores.OrganisationDomainStore
	logger                  logger.Logger
	platformBaseDomain      string
}

type domainSelection struct {
	domain   string
	platform bool
}

type StackDomainsServiceSpec struct {
	SessionFactory     db.SessionFactory
	Logger             logger.Logger
	PlatformBaseDomain string
}

func NewStackDomainsService(spec StackDomainsServiceSpec) StackDomainsService {
	return &stackDomainService{
		domainsStore: pgstore.NewStackDomainsStore(pgstore.StackDomainsStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		organisationDomainStore: pgstore.NewOrganisationDomainStore(pgstore.OrganisationDomainStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger:             spec.Logger,
		platformBaseDomain: spec.PlatformBaseDomain,
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
		s.logger.Error(ctx, "failed to check if domain exists: %v", err)
		return nil, err
	}
	if existing != nil {
		return nil, errors.Conflict("domain '%s' is already in use", spec.Fqdn)
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
		s.logger.Error(ctx, "failed to create domain: %v", err)
		return nil, err
	}
	return createdDomain, nil
}

func (s *stackDomainService) InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	return s.domainsStore.DeleteWithTx(ctx, id)
}

func (s *stackDomainService) PopulateAndSaveExposedPortDomainsForResourceWithTx(ctx context.Context, stack *models.Stack, resource *models.StackResource) *errors.ServiceError {
	existingDomains, err := s.domainsStore.ListByStackResourceID(ctx, resource.ID)
	if err != nil {
		return err
	}

	existingByPort := stackDomainsByPort(existingDomains)
	var selection *domainSelection
	if hasUnassignedPublicPort(resource.Ports, existingByPort) {
		selection, err = s.domainForNewAssignments(ctx, stack.OrganisationID)
		if err != nil {
			return err
		}
	}

	publicPorts := make(map[int]struct{})
	for i := range resource.Ports {
		port := &resource.Ports[i]
		if !port.ExposedToPublic {
			clearPortDomain(port)
			continue
		}
		publicPorts[port.Number] = struct{}{}

		if existing := existingByPort[port.Number]; existing != nil {
			port.ExposedFqdn = existing.Fqdn
			continue
		}

		if createErr := s.assignDomainToPortWithTx(ctx, stack, resource, port, selection); createErr != nil {
			return createErr
		}
	}

	return s.deleteStaleDomainsWithTx(ctx, existingDomains, publicPorts)
}

func stackDomainsByPort(domains models.StackDomainList) map[int]*models.StackDomain {
	byPort := make(map[int]*models.StackDomain, len(domains))
	for _, domain := range domains {
		byPort[domain.TargetPort] = domain
	}
	return byPort
}

func hasUnassignedPublicPort(ports models.Ports, existingByPort map[int]*models.StackDomain) bool {
	for _, port := range ports {
		if port.ExposedToPublic && existingByPort[port.Number] == nil {
			return true
		}
	}
	return false
}

func (s *stackDomainService) domainForNewAssignments(ctx context.Context, organisationID string) (*domainSelection, *errors.ServiceError) {
	organisationDomains, err := s.organisationDomainStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		s.logger.Error(ctx, "failed to list domains for organisation %s: %v", organisationID, err)
		return nil, err
	}
	if len(organisationDomains) != 0 {
		return &domainSelection{domain: organisationDomains[0].Domain}, nil
	}
	if s.platformBaseDomain != "" {
		return &domainSelection{domain: s.platformBaseDomain, platform: true}, nil
	}
	return nil, errors.BadRequest("no domain is configured for public port allocation")
}

func clearPortDomain(port *models.Port) {
	port.ExposedFqdn = ""
	port.GeneratedSubdomainPrefix = ""
}

func (s *stackDomainService) assignDomainToPortWithTx(
	ctx context.Context,
	stack *models.Stack,
	resource *models.StackResource,
	port *models.Port,
	selection *domainSelection,
) *errors.ServiceError {
	if selection.platform {
		fqdn, generatedPrefix, err := FQDNForPortWithPlatformDomain(resource.ID, resource.Name, selection.domain, *port)
		if err != nil {
			return errors.BadRequest("failed to allocate platform domain: %s", err.Error())
		}
		port.ExposedFqdn = fqdn
		port.GeneratedSubdomainPrefix = generatedPrefix
	} else {
		port.ExposedFqdn, port.GeneratedSubdomainPrefix = FQDNForPortWithCustomDomain(stack.ID, resource.Name, selection.domain, *port)
	}

	domain := &models.StackDomain{
		Fqdn:              port.ExposedFqdn,
		OrganisationID:    stack.OrganisationID,
		StackID:           stack.ID,
		StackResourceID:   resource.ID,
		StackResourceName: resource.Name,
		TargetPort:        port.Number,
	}
	_, err := s.InternalCreateWithTx(ctx, domain)
	return err
}

func (s *stackDomainService) deleteStaleDomainsWithTx(
	ctx context.Context,
	existingDomains models.StackDomainList,
	publicPorts map[int]struct{},
) *errors.ServiceError {
	for _, existingDomain := range existingDomains {
		if _, ok := publicPorts[existingDomain.TargetPort]; ok {
			continue
		}
		if deleteErr := s.domainsStore.DeleteWithTx(ctx, existingDomain.ID); deleteErr != nil {
			return deleteErr
		}
	}

	return nil
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

func (s *stackDomainService) Create(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
	if len(spec.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}
	existing, err := s.domainsStore.GetByFqdn(ctx, spec.Fqdn)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Error(ctx, "failed to check if domain exists: %v", err)
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
