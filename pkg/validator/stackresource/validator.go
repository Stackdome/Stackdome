package stackresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=validator.go -destination=validator_mock_test.go -package=stackresource

// Validator runs every cheap (input-only, DB-only, sibling) validation for a
// single stack resource and returns all failures. It never touches the
// network. A non-nil ServiceError means a lookup failed for reasons other
// than not-found (e.g. a DB outage) and validation was aborted; the caller
// should turn it into a 500 rather than trusting the (nil) FieldError slice.
//
//go:generate mockgen -destination=../../mocks/mock_stack_resource_validator.go -package=mocks github.com/Stackdome/stackdome/pkg/validator/stackresource Validator
type Validator interface {
	Validate(ctx context.Context, stack *models.Stack, resource *models.StackResource, siblings []*models.StackResource) ([]errors.FieldError, *errors.ServiceError)
}

// Narrow read-only seams so the validator does not depend on pkg/services.
type volumeGetter interface {
	GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError)
}

type secretGetter interface {
	GetByName(ctx context.Context, organisationID, name string) (*models.Secret, *errors.ServiceError)
}

// domainLister mirrors the organisation-domain seam already used by
// pkg/validator/stack.stackValidator (organisationDomainService).
type domainLister interface {
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.OrganisationDomain, *errors.ServiceError)
}

// gitIntegrationGetter mirrors the DB-only lookup tier of the production
// credential resolver (gitIntegrationResolverSource.InternalGetByID in
// pkg/services/credential_resolver.go). It is used instead of
// credentials.Resolver.GitCredentials for existence checks because that
// resolver mints a GitHub App installation token over the network when the
// integration type is github_app - this seam never does that.
type gitIntegrationGetter interface {
	InternalGetByID(ctx context.Context, ID string) (*models.GitIntegration, *errors.ServiceError)
}

type ValidatorSpec struct {
	Volumes         volumeGetter
	Secrets         secretGetter
	Domains         domainLister
	Credentials     credentials.Resolver
	GitIntegrations gitIntegrationGetter
}

type validator struct {
	volumes         volumeGetter
	secrets         secretGetter
	domains         domainLister
	credentials     credentials.Resolver
	gitIntegrations gitIntegrationGetter
}

func NewValidator(spec ValidatorSpec) Validator {
	return &validator{
		volumes:         spec.Volumes,
		secrets:         spec.Secrets,
		domains:         spec.Domains,
		credentials:     spec.Credentials,
		gitIntegrations: spec.GitIntegrations,
	}
}

func (v *validator) Validate(ctx context.Context, stack *models.Stack, resource *models.StackResource, siblings []*models.StackResource) ([]errors.FieldError, *errors.ServiceError) {
	var errs []errors.FieldError
	errs = append(errs, validateInputRules(resource)...)

	refErrs, serr := v.validateReferences(ctx, stack, resource)
	if serr != nil {
		return nil, serr
	}
	errs = append(errs, refErrs...)

	errs = append(errs, validateSiblingRules(resource, siblings)...)
	return errs, nil
}
