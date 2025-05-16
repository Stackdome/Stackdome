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

type DomainsService interface {
	Create(ctx context.Context, spec *models.Domain) (*models.Domain, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, spec *models.Domain) (*models.Domain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.Domain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	ListByOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) (models.DomainList, *errors.ServiceError)
	DeleteForOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError
	DeleteForOwnerWithTx(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError
	GetByFqdn(ctx context.Context, fqdn string) (*models.Domain, *errors.ServiceError)
	ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.Domain, *errors.ServiceError)
	DomainToUseForStack(ctx context.Context, stack *models.Stack) (*models.Domain, *errors.ServiceError)
}

type domainsService struct {
	domainsStore stores.DomainsStore
	logger       logger.Logger
}

type DomainsServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewDomainsService(spec DomainsServiceSpec) DomainsService {
	return &domainsService{
		domainsStore: pgstore.NewDomainsStore(pgstore.DomainsStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

func (s *domainsService) InternalCreateWithTx(ctx context.Context, spec *models.Domain) (*models.Domain, *errors.ServiceError) {
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
	if spec.OwnerType == "" {
		return nil, errors.BadRequest("domain owner type is required")
	}

	createdDomain, err := s.domainsStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create domain: %v", err)
		return nil, err
	}
	return createdDomain, nil
}

func (s *domainsService) InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError {
	return s.domainsStore.DeleteWithTx(ctx, id)
}

func (s *domainsService) DeleteForOwnerWithTx(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError {
	return s.domainsStore.DeleteForOwnerWithTx(ctx, ownerID, ownerType)
}

func (s *domainsService) DomainToUseForStack(ctx context.Context, stack *models.Stack) (*models.Domain, *errors.ServiceError) {
	domains, err := s.domainsStore.ListByOwner(ctx, stack.OrganisationID, models.OwnerTypeOrganisation)
	if err != nil {
		s.logger.Errorf("failed to list domains for organisation %s: %v", stack.OrganisationID, err)
		return nil, err
	}
	if len(domains) == 0 {
		return nil, errors.NotFound("no domains found for organisation %s", stack.OrganisationID)
	}

	// In the future we can pin a domain to a workspace or group
	// for now we just return the first one
	return domains[0], nil
}

func (s *domainsService) Create(ctx context.Context, spec *models.Domain) (*models.Domain, *errors.ServiceError) {
	if len(spec.Fqdn) == 0 {
		return nil, errors.BadRequest("domain fqdn is required")
	}

	existing, err := s.domainsStore.GetByFqdn(ctx, spec.Fqdn)
	if err != nil && err.Code != errors.ErrorNotFound {
		s.logger.Errorf("failed to check if domain exists: %v", err)
		return nil, err
	}

	if existing != nil {
		return nil, errors.Conflict("domain with fqdn '%s' already exists", spec.Fqdn)
	}

	return s.domainsStore.Create(ctx, spec)
}

func (s *domainsService) Get(ctx context.Context, id string) (*models.Domain, *errors.ServiceError) {
	return s.domainsStore.Get(ctx, id)
}

func (s *domainsService) Delete(ctx context.Context, id string) *errors.ServiceError {
	return s.domainsStore.Delete(ctx, id)
}

func (s *domainsService) ListByOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) (models.DomainList, *errors.ServiceError) {
	return s.domainsStore.ListByOwner(ctx, ownerID, ownerType)
}

func (s *domainsService) GetByFqdn(ctx context.Context, fqdn string) (*models.Domain, *errors.ServiceError) {
	return s.domainsStore.GetByFqdn(ctx, fqdn)
}

func (s *domainsService) ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.Domain, *errors.ServiceError) {
	return s.domainsStore.ListByFqdnPrefix(ctx, prefix)
}

func (s *domainsService) DeleteForOwner(ctx context.Context, ownerID string, ownerType models.OwnerType) *errors.ServiceError {
	return s.domainsStore.DeleteForOwner(ctx, ownerID, ownerType)
}

func isValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}
