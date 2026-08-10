package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
)

type ProvisioningMode string

const (
	ProvisioningModeEager        ProvisioningMode = "eager"
	ProvisioningModeDatabaseOnly ProvisioningMode = "database_only"
)

// RuntimePolicy keeps mode-specific access and limits out of core services.
// Cloud enforces them here while self-hosted remains eager and unrestricted.
//
//go:generate mockgen -source=runtime_policy.go -destination=runtime_policy_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type RuntimePolicy interface {
	OrganisationProvisioningMode() ProvisioningMode
	DraftProvisioningMode() ProvisioningMode
	IsolationPolicyVersion() string
	ActivateComputeAccessWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	RequireComputeAccessWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	RequireComputeAccess(ctx context.Context, organisationID string) *errors.ServiceError
	AdmitComputeMutationWithTx(ctx context.Context, organisationID string) (ComputeMutationAdmission, *errors.ServiceError)
	AdmitOrganisationDeletion(ctx context.Context, organisationID string) *errors.ServiceError
	AdmitStackMutationWithTx(ctx context.Context, mutation StackMutation) *errors.ServiceError
	ApplyStackResourceDefaults(resource *models.StackResource)
}

type StackMutationKind string

const (
	StackMutationCreate         StackMutationKind = "create_stack"
	StackMutationUpdate         StackMutationKind = "update_stack"
	StackMutationCreateResource StackMutationKind = "create_resource"
	StackMutationUpdateResource StackMutationKind = "update_resource"
)

type StackMutation struct {
	Kind                 StackMutationKind
	OrganisationID       string
	StackID              string
	DesiredResourceCount int64
}

// ComputeMutationAdmission carries the runtime side effects authorized by the
// entitlement and lease rows locked in the mutation transaction.
type ComputeMutationAdmission struct {
	ReconcileCluster bool
}

type selfHostedRuntimePolicy struct{}

func NewSelfHostedRuntimePolicy() RuntimePolicy {
	return selfHostedRuntimePolicy{}
}

func (selfHostedRuntimePolicy) OrganisationProvisioningMode() ProvisioningMode {
	return ProvisioningModeEager
}

func (selfHostedRuntimePolicy) DraftProvisioningMode() ProvisioningMode {
	return ProvisioningModeEager
}

func (selfHostedRuntimePolicy) IsolationPolicyVersion() string {
	return ""
}

func (selfHostedRuntimePolicy) ActivateComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) RequireComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) RequireComputeAccess(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) AdmitComputeMutationWithTx(context.Context, string) (ComputeMutationAdmission, *errors.ServiceError) {
	return ComputeMutationAdmission{ReconcileCluster: true}, nil
}

func (selfHostedRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) AdmitStackMutationWithTx(context.Context, StackMutation) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}

type stackdomeCloudRuntimePolicy struct {
	computeAccess          ComputeAccessService
	stackLimits            stores.StackLimitStore
	isolationPolicyVersion string
	maxStacks              int64
	maxResources           int64
	replicas               int32
}

type StackdomeCloudRuntimePolicySpec struct {
	ComputeAccess          ComputeAccessService
	StackLimits            stores.StackLimitStore
	IsolationPolicyVersion string
	MaxStacks              int64
	MaxResources           int64
	Replicas               int32
}

func NewStackdomeCloudRuntimePolicy(spec StackdomeCloudRuntimePolicySpec) RuntimePolicy {
	if spec.IsolationPolicyVersion == "" {
		panic("services.NewStackdomeCloudRuntimePolicy: IsolationPolicyVersion is required")
	}
	if spec.MaxStacks <= 0 {
		panic("services.NewStackdomeCloudRuntimePolicy: MaxStacks must be greater than zero")
	}
	if spec.MaxResources <= 0 {
		panic("services.NewStackdomeCloudRuntimePolicy: MaxResources must be greater than zero")
	}
	if spec.Replicas <= 0 {
		panic("services.NewStackdomeCloudRuntimePolicy: Replicas must be greater than zero")
	}
	return &stackdomeCloudRuntimePolicy{
		computeAccess:          spec.ComputeAccess,
		stackLimits:            spec.StackLimits,
		isolationPolicyVersion: spec.IsolationPolicyVersion,
		maxStacks:              spec.MaxStacks,
		maxResources:           spec.MaxResources,
		replicas:               spec.Replicas,
	}
}

func (*stackdomeCloudRuntimePolicy) OrganisationProvisioningMode() ProvisioningMode {
	return ProvisioningModeDatabaseOnly
}

func (*stackdomeCloudRuntimePolicy) DraftProvisioningMode() ProvisioningMode {
	return ProvisioningModeDatabaseOnly
}

func (p *stackdomeCloudRuntimePolicy) IsolationPolicyVersion() string {
	return p.isolationPolicyVersion
}

func (p *stackdomeCloudRuntimePolicy) ActivateComputeAccessWithTx(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.computeAccess.ActivateWithTx(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) RequireComputeAccessWithTx(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.computeAccess.RequireWithTx(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) RequireComputeAccess(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.computeAccess.RequireAccess(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) AdmitComputeMutationWithTx(ctx context.Context, organisationID string) (ComputeMutationAdmission, *errors.ServiceError) {
	access, serr := p.computeAccess.AdmitComputeMutationWithTx(ctx, organisationID)
	return ComputeMutationAdmission{ReconcileCluster: access != nil}, serr
}

func (p *stackdomeCloudRuntimePolicy) AdmitOrganisationDeletion(ctx context.Context, organisationID string) *errors.ServiceError {
	return p.computeAccess.EnsureNoLease(ctx, organisationID)
}

func (p *stackdomeCloudRuntimePolicy) AdmitStackMutationWithTx(ctx context.Context, mutation StackMutation) *errors.ServiceError {
	if _, serr := p.AdmitComputeMutationWithTx(ctx, mutation.OrganisationID); serr != nil {
		return serr
	}
	excludedStackID := ""
	switch mutation.Kind {
	case StackMutationCreate, StackMutationCreateResource, StackMutationUpdateResource:
	case StackMutationUpdate:
		if mutation.StackID == "" {
			return errors.GeneralError("stack ID is required for a whole-stack update")
		}
		excludedStackID = mutation.StackID
	default:
		return errors.GeneralError("unsupported stack mutation kind %q", mutation.Kind)
	}
	usage, serr := p.stackLimits.LockOrganisationAndGetUsageWithTx(ctx, mutation.OrganisationID, excludedStackID)
	if serr != nil {
		return serr
	}

	stackCount := usage.StackCount
	resourceCount := usage.StackResourceCount
	switch mutation.Kind {
	case StackMutationCreate:
		stackCount++
		resourceCount += mutation.DesiredResourceCount
	case StackMutationUpdate:
		resourceCount += mutation.DesiredResourceCount
	case StackMutationCreateResource:
		resourceCount++
	case StackMutationUpdateResource:
	}
	if stackCount > p.maxStacks {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d stacks per organisation", p.maxStacks)
	}
	if resourceCount > p.maxResources {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d stack resources per organisation", p.maxResources)
	}
	return nil
}

func (p *stackdomeCloudRuntimePolicy) ApplyStackResourceDefaults(resource *models.StackResource) {
	replicas := p.replicas
	resource.Replicas = &replicas
}
