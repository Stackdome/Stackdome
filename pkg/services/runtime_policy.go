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

//go:generate mockgen -source=runtime_policy.go -destination=runtime_policy_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type RuntimePolicy interface {
	OrganisationProvisioningMode() ProvisioningMode
	DraftProvisioningMode() ProvisioningMode
	IsolationPolicyVersion() string
	AdmitFirstReleaseWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	AdmitRollbackWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	RequireActiveAllocation(ctx context.Context, organisationID string) *errors.ServiceError
	AdmitMutationWithTx(ctx context.Context, organisationID string) (MutationAdmission, *errors.ServiceError)
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

// MutationAdmission describes the side effects allowed by the allocation row
// observed and locked in the mutation transaction.
type MutationAdmission struct {
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

func (selfHostedRuntimePolicy) AdmitFirstReleaseWithTx(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) AdmitRollbackWithTx(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) AdmitMutationWithTx(context.Context, string) (MutationAdmission, *errors.ServiceError) {
	return MutationAdmission{ReconcileCluster: true}, nil
}

func (selfHostedRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) AdmitStackMutationWithTx(context.Context, StackMutation) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}

type stackdomeCloudRuntimePolicy struct {
	trials                 CloudTrialService
	stackLimits            stores.StackLimitStore
	isolationPolicyVersion string
	maxStacks              int64
	maxResources           int64
	replicas               int32
}

type StackdomeCloudRuntimePolicySpec struct {
	Trials                 CloudTrialService
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
		trials:                 spec.Trials,
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

func (p *stackdomeCloudRuntimePolicy) AdmitFirstReleaseWithTx(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.trials.AcquireWithTx(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) AdmitRollbackWithTx(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.trials.RevalidateWithTx(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) RequireActiveAllocation(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.trials.RequireActive(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) AdmitMutationWithTx(ctx context.Context, organisationID string) (MutationAdmission, *errors.ServiceError) {
	allocation, serr := p.trials.RevalidateIfExistsWithTx(ctx, organisationID)
	return MutationAdmission{ReconcileCluster: allocation != nil}, serr
}

func (p *stackdomeCloudRuntimePolicy) AdmitOrganisationDeletion(ctx context.Context, organisationID string) *errors.ServiceError {
	return p.trials.EnsureNoAllocation(ctx, organisationID)
}

func (p *stackdomeCloudRuntimePolicy) AdmitStackMutationWithTx(ctx context.Context, mutation StackMutation) *errors.ServiceError {
	if _, serr := p.AdmitMutationWithTx(ctx, mutation.OrganisationID); serr != nil {
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
