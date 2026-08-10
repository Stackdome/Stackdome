package services

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"k8s.io/apimachinery/pkg/api/resource"
)

// RuntimePolicy keeps access, limits, and runtime-owned defaults out of core
// services. Stack deployment itself is release-driven in every runtime.
//
//go:generate mockgen -source=runtime_policy.go -destination=runtime_policy_mock.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type RuntimePolicy interface {
	IsolationPolicyVersion() string
	EnsureComputeAccess(ctx context.Context, organisationID string) *errors.ServiceError
	ValidateOrganisationDeletion(ctx context.Context, organisationID string) *errors.ServiceError
	ValidateStackLimits(ctx context.Context, change StackLimitChange) *errors.ServiceError
	ValidateVolumeLimits(ctx context.Context, organisationID, size string) *errors.ServiceError
	ValidatePostgresAddonLimits(ctx context.Context, organisationID, existingAddonID string, addon *models.PostgresAddon) *errors.ServiceError
	ApplyStackResourceDefaults(resource *models.StackResource)
	ApplyPostgresAddonDefaults(addon *models.PostgresAddon)
}

type StackLimitChangeKind string

const (
	StackLimitCreate         StackLimitChangeKind = "create_stack"
	StackLimitUpdate         StackLimitChangeKind = "update_stack"
	StackLimitCreateResource StackLimitChangeKind = "create_resource"
	StackLimitUpdateResource StackLimitChangeKind = "update_resource"
)

// StackLimitChange describes how a proposed draft edit changes counted stack
// usage. It does not decide compute access or runtime reconciliation.
type StackLimitChange struct {
	Kind                 StackLimitChangeKind
	OrganisationID       string
	StackID              string
	DesiredResourceCount int64
}

type selfHostedRuntimePolicy struct{}

func NewSelfHostedRuntimePolicy() RuntimePolicy {
	return selfHostedRuntimePolicy{}
}

func (selfHostedRuntimePolicy) IsolationPolicyVersion() string {
	return ""
}

func (selfHostedRuntimePolicy) EnsureComputeAccess(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ValidateOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ValidateStackLimits(context.Context, StackLimitChange) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ValidateVolumeLimits(context.Context, string, string) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ValidatePostgresAddonLimits(context.Context, string, string, *models.PostgresAddon) *errors.ServiceError {
	return nil
}

func (selfHostedRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}

func (selfHostedRuntimePolicy) ApplyPostgresAddonDefaults(*models.PostgresAddon) {}

type stackdomeCloudRuntimePolicy struct {
	computeAccess          ComputeAccessService
	computeUsage           stores.ComputeUsageStore
	isolationPolicyVersion string
	limits                 models.ComputeLimits
	maxVolumeSize          resource.Quantity
	maxPostgresStorageSize resource.Quantity
}

type StackdomeCloudRuntimePolicySpec struct {
	ComputeAccess          ComputeAccessService
	ComputeUsage           stores.ComputeUsageStore
	IsolationPolicyVersion string
	Limits                 models.ComputeLimits
}

func NewStackdomeCloudRuntimePolicy(spec StackdomeCloudRuntimePolicySpec) RuntimePolicy {
	if spec.IsolationPolicyVersion == "" {
		panic("services.NewStackdomeCloudRuntimePolicy: IsolationPolicyVersion is required")
	}
	maxVolumeSize := mustParsePolicyQuantity("MaxVolumeSize", spec.Limits.MaxVolumeSize)
	maxPostgresStorageSize := mustParsePolicyQuantity("MaxPostgresStorageSize", spec.Limits.MaxPostgresStorageSize)
	return &stackdomeCloudRuntimePolicy{
		computeAccess:          spec.ComputeAccess,
		computeUsage:           spec.ComputeUsage,
		isolationPolicyVersion: spec.IsolationPolicyVersion,
		limits:                 spec.Limits,
		maxVolumeSize:          maxVolumeSize,
		maxPostgresStorageSize: maxPostgresStorageSize,
	}
}

func mustParsePolicyQuantity(name, value string) resource.Quantity {
	quantity, err := resource.ParseQuantity(value)
	if err != nil || quantity.Sign() <= 0 {
		panic(fmt.Sprintf("services.NewStackdomeCloudRuntimePolicy: %s must be a positive Kubernetes quantity", name))
	}
	return quantity
}

func (p *stackdomeCloudRuntimePolicy) IsolationPolicyVersion() string {
	return p.isolationPolicyVersion
}

func (p *stackdomeCloudRuntimePolicy) EnsureComputeAccess(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.computeAccess.Activate(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudRuntimePolicy) ValidateOrganisationDeletion(ctx context.Context, organisationID string) *errors.ServiceError {
	return p.computeAccess.EnsureNoLease(ctx, organisationID)
}

func (p *stackdomeCloudRuntimePolicy) ValidateStackLimits(ctx context.Context, change StackLimitChange) *errors.ServiceError {
	excludedStackID := ""
	switch change.Kind {
	case StackLimitCreate, StackLimitCreateResource, StackLimitUpdateResource:
	case StackLimitUpdate:
		if change.StackID == "" {
			return errors.GeneralError("stack ID is required for a whole-stack update")
		}
		excludedStackID = change.StackID
	default:
		return errors.GeneralError("unsupported stack limit change kind %q", change.Kind)
	}
	usage, serr := p.computeUsage.LockOrganisationAndGetUsageWithTx(ctx, change.OrganisationID, excludedStackID)
	if serr != nil {
		return serr
	}

	stackCount := usage.StackCount
	resourceCount := usage.StackResourceCount
	switch change.Kind {
	case StackLimitCreate:
		stackCount++
		resourceCount += change.DesiredResourceCount
	case StackLimitUpdate:
		resourceCount += change.DesiredResourceCount
	case StackLimitCreateResource:
		resourceCount++
	case StackLimitUpdateResource:
	}
	if stackCount > p.limits.MaxStacksPerOrganization {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d stacks per organisation", p.limits.MaxStacksPerOrganization)
	}
	if resourceCount > p.limits.MaxStackResourcesPerOrganization {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d stack resources per organisation", p.limits.MaxStackResourcesPerOrganization)
	}
	return nil
}

func (p *stackdomeCloudRuntimePolicy) ValidateVolumeLimits(ctx context.Context, organisationID, size string) *errors.ServiceError {
	usage, serr := p.computeUsage.LockOrganisationAndGetUsageWithTx(ctx, organisationID, "")
	if serr != nil {
		return serr
	}
	if usage.VolumeCount+1 > p.limits.MaxVolumesPerOrganization {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d volumes per organisation", p.limits.MaxVolumesPerOrganization)
	}
	requested, err := resource.ParseQuantity(size)
	if err != nil || requested.Sign() <= 0 {
		return errors.BadRequest("volume size must be a positive Kubernetes quantity")
	}
	if requested.Cmp(p.maxVolumeSize) > 0 {
		return errors.BadRequest("Stackdome Cloud allows a maximum volume size of %s", p.limits.MaxVolumeSize)
	}
	return nil
}

func (p *stackdomeCloudRuntimePolicy) ValidatePostgresAddonLimits(ctx context.Context, organisationID, existingAddonID string, addon *models.PostgresAddon) *errors.ServiceError {
	usage, serr := p.computeUsage.LockOrganisationAndGetUsageWithTx(ctx, organisationID, "")
	if serr != nil {
		return serr
	}
	addonCount := usage.PostgresAddonCount
	if existingAddonID == "" {
		addonCount++
	}
	if addonCount > p.limits.MaxPostgresAddonsPerOrganization {
		return errors.BadRequest("Stackdome Cloud allows a maximum of %d PostgreSQL addon per organisation", p.limits.MaxPostgresAddonsPerOrganization)
	}
	if addon.Instances.Count != p.limits.PostgresInstances {
		return errors.BadRequest("Stackdome Cloud PostgreSQL addons must use exactly %d instance", p.limits.PostgresInstances)
	}
	requested, err := resource.ParseQuantity(addon.Storage.Size)
	if err != nil || requested.Sign() <= 0 {
		return errors.BadRequest("PostgreSQL addon storage size must be a positive Kubernetes quantity")
	}
	if requested.Cmp(p.maxPostgresStorageSize) > 0 {
		return errors.BadRequest("Stackdome Cloud allows a maximum PostgreSQL addon storage size of %s", p.limits.MaxPostgresStorageSize)
	}
	return nil
}

func (p *stackdomeCloudRuntimePolicy) ApplyStackResourceDefaults(resource *models.StackResource) {
	replicas := p.limits.ReplicasPerStackResource
	resource.Replicas = &replicas
}

func (p *stackdomeCloudRuntimePolicy) ApplyPostgresAddonDefaults(addon *models.PostgresAddon) {
	if addon.Instances.Count == 0 {
		addon.Instances.Count = p.limits.PostgresInstances
	}
	if addon.Storage.Size == "" {
		addon.Storage.Size = p.limits.MaxPostgresStorageSize
	}
}
