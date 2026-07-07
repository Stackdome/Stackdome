package stackresource

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=validator.go -destination=validator_mock_test.go -package=stackresource

// Validator runs every cheap (input-only, DB-only, sibling) validation for a
// single stack resource and returns all failures. It never touches the network.
type Validator interface {
	Validate(ctx context.Context, stack *models.Stack, resource *models.StackResource, siblings []*models.StackResource) []errors.FieldError
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

type ValidatorSpec struct {
	Volumes     volumeGetter
	Secrets     secretGetter
	Domains     domainLister
	Credentials credentials.Resolver
}

type validator struct {
	volumes     volumeGetter
	secrets     secretGetter
	domains     domainLister
	credentials credentials.Resolver
}

func NewValidator(spec ValidatorSpec) Validator {
	return &validator{
		volumes:     spec.Volumes,
		secrets:     spec.Secrets,
		domains:     spec.Domains,
		credentials: spec.Credentials,
	}
}

func (v *validator) Validate(ctx context.Context, stack *models.Stack, resource *models.StackResource, siblings []*models.StackResource) []errors.FieldError {
	var errs []errors.FieldError
	errs = append(errs, validateInputRules(resource)...)
	errs = append(errs, v.validateReferences(ctx, stack, resource)...)
	errs = append(errs, validateSiblingRules(resource, siblings)...)
	return errs
}

// validateReferences is a placeholder until Task 4 lands DB-only reference
// checks (volumes, secrets, domains, credentials).
func (v *validator) validateReferences(ctx context.Context, stack *models.Stack, resource *models.StackResource) []errors.FieldError {
	return nil
}

// validateSiblingRules is a placeholder until Task 5 lands cross-resource
// checks within the same stack (duplicate names, dependency cycles, etc).
func validateSiblingRules(resource *models.StackResource, siblings []*models.StackResource) []errors.FieldError {
	return nil
}
