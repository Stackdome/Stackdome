package services

import (
	"context"
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

var domainRegex = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

type StackDomainsService interface {
	Create(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, spec *models.StackDomain) (*models.StackDomain, *errors.ServiceError)
	Get(ctx context.Context, id string) (*models.StackDomain, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	InternalDeleteWithTx(ctx context.Context, id string) *errors.ServiceError
	GetByFqdn(ctx context.Context, fqdn string) (*models.StackDomain, *errors.ServiceError)
	ListByFqdnPrefix(ctx context.Context, prefix string) ([]*models.StackDomain, *errors.ServiceError)
	DomainToUseForStack(ctx context.Context, stack *models.Stack) (*models.OrganisationDomain, *errors.ServiceError)
	PopulateAndSaveExposedPortDomainsForStackWithTx(ctx context.Context, stack *models.Stack) *errors.ServiceError
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

func (s *stackDomainService) PopulateAndSaveExposedPortDomainsForStackWithTx(ctx context.Context, stack *models.Stack) *errors.ServiceError {
	if !stack.HasExposedPorts() {
		return nil
	}

	// Get the domain to use for the stack
	domainToUse, err := s.DomainToUseForStack(ctx, stack)
	if err != nil {
		return err
	}

	for i := range stack.StackResources {
		curr := stack.StackResources[i]
		if len(curr.Ports) == 0 {
			continue
		}
		for j := range curr.Ports {
			if curr.Ports[j].ExposedToPublic {
				// If the exposed port does not have a subdomain prefix, set it to the
				// encoded stack resource ID and port number.
				if curr.Ports[j].SubdomainPrefix == "" {
					curr.Ports[j].GeneratedSubdomainPrefix = encodeStackResourceIDAndPort(curr.ID, curr.Ports[j].Number)
					// Set the exposed port's FQDN using the subdomain prefix and the domain to use.
					curr.Ports[j].ExposedFqdn = fmt.Sprintf(
						"%s.%s.%s", curr.Ports[j].GeneratedSubdomainPrefix, curr.Name, domainToUse.Domain)
				} else {
					curr.Ports[j].ExposedFqdn = fmt.Sprintf(
						"%s.%s", curr.Ports[j].SubdomainPrefix, domainToUse.Domain)
				}
				// Create the domain object for the exposed port.
				domain := &models.StackDomain{
					Fqdn:              curr.Ports[j].ExposedFqdn,
					OrganisationID:    stack.OrganisationID,
					StackID:           stack.ID,
					StackResourceID:   curr.ID,
					StackResourceName: curr.Name,
					TargetPort:        curr.Ports[j].Number,
				}

				present, err := s.domainPresentForStackResourceAndPort(ctx, curr.ID, curr.Ports[j].Number)
				if err != nil {
					return err
				}
				if present {
					// If the domain already exists, skip creating it.
					// TODO: Check if the existing domain matches the new one.
					// If it doesn't, we might need to update it.
					continue
				}
				if _, cerr := s.InternalCreateWithTx(ctx, domain); cerr != nil {
					if cerr.Code == errors.ErrorConflict {
						existing, _ := s.GetByFqdn(ctx, curr.Ports[j].ExposedFqdn)
						if existing != nil {
							return errors.Conflict(
								"domain '%s' is already in use by resource '%s' in another stack (id: %s)",
								curr.Ports[j].ExposedFqdn, existing.StackResourceName, existing.StackID)
						}
						return errors.Conflict(
							"domain '%s' is already in use", curr.Ports[j].ExposedFqdn)
					} else {
						return errors.GeneralError(
							"failed to create domain for stack resource '%s': %s", curr.Name, cerr.Error())
					}
				}
			}
		}
	}

	// Delete any existing domains that are not in the current stack resource's ports.
	for i := range stack.StackResources {
		curr := stack.StackResources[i]
		existingDomains, err := s.domainsStore.ListByStackResourceID(ctx, curr.ID)
		if err != nil {
			return err
		}
		if len(existingDomains) == 0 {
			continue
		}

		currentPortDomainMap := make(map[int]struct{})
		for _, port := range curr.Ports {
			if port.ExposedToPublic {
				currentPortDomainMap[port.Number] = struct{}{}
			}
		}
		for _, existingDomain := range existingDomains {
			if _, ok := currentPortDomainMap[existingDomain.TargetPort]; !ok {
				// If the domain is not in the current stack resource's ports, delete it.
				if err := s.domainsStore.DeleteWithTx(ctx, existingDomain.ID); err != nil {
					return err
				}
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

func encodeStackResourceIDAndPort(uuid string, port int) string {
	// Combine UUID and port into a single string
	input := uuid + ":" + strconv.Itoa(port)

	hasher := md5.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)

	// Encode the hash using Base32 (URL-safe) and trim padding
	base32Encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash)

	// Truncate the result to 16 characters for a shorter subdomain
	if len(base32Encoded) > 16 {
		base32Encoded = base32Encoded[:16]
	}

	return strings.ToLower(base32Encoded)
}
