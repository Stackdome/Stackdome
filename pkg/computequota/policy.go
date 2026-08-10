package computequota

import (
	"context"
	"fmt"

	"github.com/Stackdome/stackdome/pkg/computeaccess"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Policy keeps request-layer access checks, limits, and runtime-owned defaults
// out of request services.
type Policy interface {
	EnsureAccess(ctx context.Context, organisationID string) *errors.ServiceError
	ValidateStackLimits(ctx context.Context, change StackLimitChange) *errors.ServiceError
	ValidateVolumeLimits(ctx context.Context, organisationID, size string) *errors.ServiceError
	ValidatePostgresAddonLimits(ctx context.Context, organisationID, existingAddonID string, addon *models.PostgresAddon) *errors.ServiceError
	ApplyStackResourceDefaults(resource *models.StackResource)
	ApplyVolumeDefaults(volume *models.Volume)
	ApplyPostgresAddonDefaults(addon *models.PostgresAddon)
}

type selfHostedPolicy struct{}

func NewSelfHostedPolicy() Policy {
	return selfHostedPolicy{}
}

func (selfHostedPolicy) EnsureAccess(context.Context, string) *errors.ServiceError {
	return nil
}

func (selfHostedPolicy) ValidateStackLimits(context.Context, StackLimitChange) *errors.ServiceError {
	return nil
}

func (selfHostedPolicy) ValidateVolumeLimits(context.Context, string, string) *errors.ServiceError {
	return nil
}

func (selfHostedPolicy) ValidatePostgresAddonLimits(context.Context, string, string, *models.PostgresAddon) *errors.ServiceError {
	return nil
}

func (selfHostedPolicy) ApplyStackResourceDefaults(*models.StackResource) {}

func (selfHostedPolicy) ApplyVolumeDefaults(*models.Volume) {}

func (selfHostedPolicy) ApplyPostgresAddonDefaults(*models.PostgresAddon) {}

type stackdomeCloudPolicy struct {
	computeAccess          computeaccess.Service
	computeUsage           UsageStore
	limits                 ComputeLimits
	maxVolumeSize          resource.Quantity
	maxPostgresStorageSize resource.Quantity
}

type StackdomeCloudPolicySpec struct {
	ComputeAccess computeaccess.Service
	ComputeUsage  UsageStore
	Limits        ComputeLimits
}

func NewStackdomeCloudPolicy(spec StackdomeCloudPolicySpec) Policy {
	return &stackdomeCloudPolicy{
		computeAccess:          spec.ComputeAccess,
		computeUsage:           spec.ComputeUsage,
		limits:                 spec.Limits,
		maxVolumeSize:          mustParsePolicyQuantity("MaxVolumeSize", spec.Limits.MaxVolumeSize),
		maxPostgresStorageSize: mustParsePolicyQuantity("MaxPostgresStorageSize", spec.Limits.MaxPostgresStorageSize),
	}
}

func mustParsePolicyQuantity(name, value string) resource.Quantity {
	quantity, err := resource.ParseQuantity(value)
	if err != nil || quantity.Sign() <= 0 {
		panic(fmt.Sprintf("computequota.NewStackdomeCloudPolicy: %s must be a positive Kubernetes quantity", name))
	}
	return quantity
}

func (p *stackdomeCloudPolicy) EnsureAccess(ctx context.Context, organisationID string) *errors.ServiceError {
	_, serr := p.computeAccess.Activate(ctx, organisationID)
	return serr
}

func (p *stackdomeCloudPolicy) ValidateStackLimits(ctx context.Context, change StackLimitChange) *errors.ServiceError {
	excludeStackID, serr := stackResourceExclusionFor(change)
	if serr != nil {
		return serr
	}

	currentUsage, serr := p.computeUsage.LockOrganisationAndGetUsage(ctx, change.OrganisationID, excludeStackID)
	if serr != nil {
		return serr
	}

	proposedUsage, serr := stackUsageAfterChange(currentUsage, change)
	if serr != nil {
		return serr
	}
	if proposedUsage.StackCount > p.limits.MaxStacksPerOrganization {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum of %d stacks per organisation", p.limits.MaxStacksPerOrganization)
	}
	if proposedUsage.StackResourceCount > p.limits.MaxStackResourcesPerOrganization {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum of %d stack resources per organisation", p.limits.MaxStackResourcesPerOrganization)
	}
	return nil
}

func (p *stackdomeCloudPolicy) ValidateVolumeLimits(ctx context.Context, organisationID, size string) *errors.ServiceError {
	usage, serr := p.computeUsage.LockOrganisationAndGetUsage(ctx, organisationID, "")
	if serr != nil {
		return serr
	}
	if usage.VolumeCount+1 > p.limits.MaxVolumesPerOrganization {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum of %d volumes per organisation", p.limits.MaxVolumesPerOrganization)
	}

	requested, err := resource.ParseQuantity(size)
	if err != nil || requested.Sign() <= 0 {
		return errors.BadRequest("volume size must be a positive Kubernetes quantity")
	}
	if requested.Cmp(p.maxVolumeSize) > 0 {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum volume size of %s", p.limits.MaxVolumeSize)
	}
	return nil
}

func (p *stackdomeCloudPolicy) ValidatePostgresAddonLimits(ctx context.Context, organisationID, existingAddonID string, addon *models.PostgresAddon) *errors.ServiceError {
	usage, serr := p.computeUsage.LockOrganisationAndGetUsage(ctx, organisationID, "")
	if serr != nil {
		return serr
	}

	addonCount := usage.PostgresAddonCount
	if existingAddonID == "" {
		addonCount++
	}
	if addonCount > p.limits.MaxPostgresAddonsPerOrganization {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum of %d PostgreSQL addon per organisation", p.limits.MaxPostgresAddonsPerOrganization)
	}
	if addon.Instances.Count != p.limits.PostgresInstances {
		return errors.ComputeQuotaExceeded("Stackdome Cloud PostgreSQL addons must use exactly %d instance", p.limits.PostgresInstances)
	}

	requested, err := resource.ParseQuantity(addon.Storage.Size)
	if err != nil || requested.Sign() <= 0 {
		return errors.BadRequest("PostgreSQL addon storage size must be a positive Kubernetes quantity")
	}
	if requested.Cmp(p.maxPostgresStorageSize) > 0 {
		return errors.ComputeQuotaExceeded("Stackdome Cloud allows a maximum PostgreSQL addon storage size of %s", p.limits.MaxPostgresStorageSize)
	}
	return nil
}

func (p *stackdomeCloudPolicy) ApplyStackResourceDefaults(stackResource *models.StackResource) {
	replicas := p.limits.ReplicasPerStackResource
	stackResource.Replicas = &replicas
}

func (p *stackdomeCloudPolicy) ApplyVolumeDefaults(volume *models.Volume) {
	if volume.Size == "" {
		volume.Size = p.limits.MaxVolumeSize
	}
	if volume.StorageClass == "" {
		volume.StorageClass = p.limits.VolumeStorageClass
	}
}

func (p *stackdomeCloudPolicy) ApplyPostgresAddonDefaults(addon *models.PostgresAddon) {
	if addon.Instances.Count == 0 {
		addon.Instances.Count = p.limits.PostgresInstances
	}
	if addon.Storage.Size == "" {
		addon.Storage.Size = p.limits.MaxPostgresStorageSize
	}
}
