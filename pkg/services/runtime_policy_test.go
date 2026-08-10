package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuntimePolicy", func() {
	It("keeps self-hosted provisioning eager without trial admission", func() {
		policy := NewSelfHostedRuntimePolicy()
		Expect(policy.OrganisationProvisioningMode()).To(Equal(ProvisioningModeEager))
		Expect(policy.DraftProvisioningMode()).To(Equal(ProvisioningModeEager))
		Expect(policy.AdmitFirstReleaseWithTx(context.Background(), "org-1")).To(BeNil())
		Expect(policy.AdmitRollbackWithTx(context.Background(), "org-1")).To(BeNil())
		Expect(policy.RequireActiveAllocation(context.Background(), "org-1")).To(BeNil())
		Expect(policy.IsolationPolicyVersion()).To(BeEmpty())
	})

	It("makes cloud organisations and drafts database-only and delegates trial admission", func() {
		cloudTrials := &fakeCloudTrialService{}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			Trials:                 cloudTrials,
			StackLimits:            &fakeStackLimitStore{},
			IsolationPolicyVersion: "policy-v1",
			MaxStacks:              2,
			MaxResources:           6,
			Replicas:               1,
		})
		Expect(policy.OrganisationProvisioningMode()).To(Equal(ProvisioningModeDatabaseOnly))
		Expect(policy.DraftProvisioningMode()).To(Equal(ProvisioningModeDatabaseOnly))
		Expect(policy.IsolationPolicyVersion()).To(Equal("policy-v1"))
	})

	It("enforces cloud stack and resource limits from locked usage", func() {
		stackLimits := &fakeStackLimitStore{usage: stores.StackUsage{StackCount: 2, StackResourceCount: 5}}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			Trials:                 &fakeCloudTrialService{},
			StackLimits:            stackLimits,
			IsolationPolicyVersion: "policy-v1",
			MaxStacks:              2,
			MaxResources:           6,
			Replicas:               1,
		})
		serr := policy.AdmitStackMutationWithTx(context.Background(), StackMutation{Kind: StackMutationCreate, OrganisationID: "org-1", DesiredResourceCount: 1})
		Expect(serr).ToNot(BeNil())
		Expect(serr.Reason).To(ContainSubstring("maximum of 2 stacks"))

		serr = policy.AdmitStackMutationWithTx(context.Background(), StackMutation{Kind: StackMutationUpdate, OrganisationID: "org-1", StackID: "stack-1", DesiredResourceCount: 2})
		Expect(serr).ToNot(BeNil())
		Expect(serr.Reason).To(ContainSubstring("maximum of 6 stack resources"))
		Expect(stackLimits.excludeStackID).To(Equal("stack-1"))
	})

	It("checks a whole-stack update using persisted resources outside that stack plus the desired replacement", func() {
		stackLimits := &fakeStackLimitStore{usage: stores.StackUsage{StackCount: 2, StackResourceCount: 2}}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			Trials:                 &fakeCloudTrialService{},
			StackLimits:            stackLimits,
			IsolationPolicyVersion: "policy-v1",
			MaxStacks:              2,
			MaxResources:           6,
			Replicas:               1,
		})

		serr := policy.AdmitStackMutationWithTx(context.Background(), StackMutation{
			Kind: StackMutationUpdate, OrganisationID: "org-1", StackID: "stack-1", DesiredResourceCount: 4,
		})

		Expect(serr).To(BeNil())
		Expect(stackLimits.excludeStackID).To(Equal("stack-1"))
	})

	It("forces every cloud stack resource to the configured replica count", func() {
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			Trials:                 &fakeCloudTrialService{},
			StackLimits:            &fakeStackLimitStore{},
			IsolationPolicyVersion: "policy-v1",
			MaxStacks:              2,
			MaxResources:           6,
			Replicas:               1,
		})
		replicas := int32(12)
		resource := &models.StackResource{Replicas: &replicas}
		policy.ApplyStackResourceDefaults(resource)
		Expect(resource.Replicas).ToNot(BeNil())
		Expect(*resource.Replicas).To(Equal(int32(1)))
	})
})

// fakeCloudTrialService keeps this test focused on runtime policy selection.
// Delegation details are covered by cloud_trial_service_test.go.
type fakeCloudTrialService struct {
}

type fakeStackLimitStore struct {
	usage          stores.StackUsage
	excludeStackID string
}

func newCloudRuntimePolicyForTest() RuntimePolicy {
	return newCloudRuntimePolicyWithStoreForTest(&fakeStackLimitStore{})
}

func newCloudRuntimePolicyWithStoreForTest(stackLimits stores.StackLimitStore) RuntimePolicy {
	return NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
		Trials:                 &fakeCloudTrialService{},
		StackLimits:            stackLimits,
		IsolationPolicyVersion: "policy-v1",
		MaxStacks:              2,
		MaxResources:           6,
		Replicas:               1,
	})
}

func (f *fakeStackLimitStore) LockOrganisationAndGetUsageWithTx(_ context.Context, _, excludeStackID string) (stores.StackUsage, *errors.ServiceError) {
	f.excludeStackID = excludeStackID
	return f.usage, nil
}

func (f *fakeCloudTrialService) AcquireWithTx(context.Context, string) (*models.TrialAllocation, *errors.ServiceError) {
	return &models.TrialAllocation{}, nil
}

func (f *fakeCloudTrialService) RevalidateWithTx(context.Context, string) (*models.TrialAllocation, *errors.ServiceError) {
	return &models.TrialAllocation{}, nil
}

func (f *fakeCloudTrialService) RequireActive(context.Context, string) (*models.TrialAllocation, *errors.ServiceError) {
	return &models.TrialAllocation{}, nil
}
