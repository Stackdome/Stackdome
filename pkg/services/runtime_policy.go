package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
)

type ProvisioningMode string

const (
	ProvisioningModeEager        ProvisioningMode = "eager"
	ProvisioningModeDatabaseOnly ProvisioningMode = "database_only"
)

//go:generate mockgen -source=runtime_policy.go -destination=runtime_policy_mock_test.go -package=services -self_package=github.com/Stackdome/stackdome/pkg/services
type RuntimePolicy interface {
	OrganisationProvisioningMode() ProvisioningMode
	DraftProvisioningMode() ProvisioningMode
	IsolationPolicyVersion() string
	AdmitFirstReleaseWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	AdmitRollbackWithTx(ctx context.Context, organisationID string) *errors.ServiceError
	RequireActiveAllocation(ctx context.Context, organisationID string) *errors.ServiceError
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

type stackdomeCloudRuntimePolicy struct {
	trials                 CloudTrialService
	isolationPolicyVersion string
}

type StackdomeCloudRuntimePolicySpec struct {
	Trials                 CloudTrialService
	IsolationPolicyVersion string
}

func NewStackdomeCloudRuntimePolicy(spec StackdomeCloudRuntimePolicySpec) RuntimePolicy {
	if spec.Trials == nil {
		panic("services.NewStackdomeCloudRuntimePolicy: CloudTrialService is required")
	}
	if spec.IsolationPolicyVersion == "" {
		panic("services.NewStackdomeCloudRuntimePolicy: IsolationPolicyVersion is required")
	}
	return &stackdomeCloudRuntimePolicy{
		trials:                 spec.Trials,
		isolationPolicyVersion: spec.IsolationPolicyVersion,
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
