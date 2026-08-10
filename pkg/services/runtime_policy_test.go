package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RuntimePolicy", func() {
	It("keeps self-hosted provisioning eager without compute access admission", func() {
		policy := NewSelfHostedRuntimePolicy()
		Expect(policy.OrganisationProvisioningMode()).To(Equal(ProvisioningModeEager))
		Expect(policy.DraftProvisioningMode()).To(Equal(ProvisioningModeEager))
		Expect(policy.ActivateComputeAccessWithTx(context.Background(), "org-1")).To(BeNil())
		Expect(policy.RequireComputeAccessWithTx(context.Background(), "org-1")).To(BeNil())
		admission, serr := policy.AdmitComputeMutationWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(admission.ReconcileCluster).To(BeTrue())
		Expect(policy.AdmitOrganisationDeletion(context.Background(), "org-1")).To(BeNil())
		Expect(policy.IsolationPolicyVersion()).To(BeEmpty())
	})

	It("delegates cloud mutation and organisation deletion admission", func() {
		ctrl := gomock.NewController(GinkgoT())
		access := NewMockComputeAccessService(ctrl)
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess: access, StackLimits: &fakeStackLimitStore{}, IsolationPolicyVersion: "policy-v1",
			MaxStacks: 2, MaxResources: 6, Replicas: 1,
		})
		access.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(nil, errors.ComputeAccessInactive())
		_, serr := policy.AdmitComputeMutationWithTx(context.Background(), "org-1")
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		access.EXPECT().EnsureNoLease(gomock.Any(), "org-1").Return(errors.BadRequest("allocation exists"))
		Expect(policy.AdmitOrganisationDeletion(context.Background(), "org-1").Reason).To(Equal("allocation exists"))
	})

	It("distinguishes database-only draft mutations from active compute access mutations", func() {
		access := &fakeComputeAccessService{}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess: access, StackLimits: &fakeStackLimitStore{}, IsolationPolicyVersion: "policy-v1",
			MaxStacks: 2, MaxResources: 6, Replicas: 1,
		})

		admission, serr := policy.AdmitComputeMutationWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(admission.ReconcileCluster).To(BeFalse())

		access.existingAccess = &models.ComputeAccess{Lease: &models.SharedComputeLease{OrganisationID: "org-1"}}
		admission, serr = policy.AdmitComputeMutationWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(admission.ReconcileCluster).To(BeTrue())
	})

	It("rejects counted stack mutations before reading usage when compute access is inactive", func() {
		ctrl := gomock.NewController(GinkgoT())
		access := NewMockComputeAccessService(ctrl)
		limits := &fakeStackLimitStore{}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess: access, StackLimits: limits, IsolationPolicyVersion: "policy-v1",
			MaxStacks: 2, MaxResources: 6, Replicas: 1,
		})
		access.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(nil, errors.ComputeAccessInactive())

		serr := policy.AdmitStackMutationWithTx(context.Background(), StackMutation{Kind: StackMutationCreate, OrganisationID: "org-1"})

		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		Expect(limits.called).To(BeFalse())
	})

	It("makes cloud organisations and drafts database-only and delegates compute access admission", func() {
		computeAccess := &fakeComputeAccessService{}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess:          computeAccess,
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
			ComputeAccess:          &fakeComputeAccessService{},
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
			ComputeAccess:          &fakeComputeAccessService{},
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
			ComputeAccess:          &fakeComputeAccessService{},
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

// fakeComputeAccessService keeps this test focused on runtime policy selection.
// Delegation details are covered by compute_access_service_test.go.
type fakeComputeAccessService struct {
	existingAccess *models.ComputeAccess
}

type fakeStackLimitStore struct {
	usage          stores.StackUsage
	excludeStackID string
	called         bool
}

func newCloudRuntimePolicyForTest() RuntimePolicy {
	return newCloudRuntimePolicyWithStoreForTest(&fakeStackLimitStore{})
}

func newCloudRuntimePolicyWithStoreForTest(stackLimits stores.StackLimitStore) RuntimePolicy {
	return NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
		ComputeAccess:          &fakeComputeAccessService{},
		StackLimits:            stackLimits,
		IsolationPolicyVersion: "policy-v1",
		MaxStacks:              2,
		MaxResources:           6,
		Replicas:               1,
	})
}

func (f *fakeStackLimitStore) LockOrganisationAndGetUsageWithTx(_ context.Context, _, excludeStackID string) (stores.StackUsage, *errors.ServiceError) {
	f.called = true
	f.excludeStackID = excludeStackID
	return f.usage, nil
}

func (f *fakeComputeAccessService) ActivateWithTx(context.Context, string) (*models.ComputeAccess, *errors.ServiceError) {
	return &models.ComputeAccess{}, nil
}

func (f *fakeComputeAccessService) RequireWithTx(context.Context, string) (*models.ComputeAccess, *errors.ServiceError) {
	return &models.ComputeAccess{}, nil
}

func (f *fakeComputeAccessService) AdmitComputeMutationWithTx(context.Context, string) (*models.ComputeAccess, *errors.ServiceError) {
	return f.existingAccess, nil
}

func (f *fakeComputeAccessService) RequireAccess(context.Context, string) (*models.ComputeAccess, *errors.ServiceError) {
	return &models.ComputeAccess{}, nil
}

func (f *fakeComputeAccessService) EnsureNoLease(context.Context, string) *errors.ServiceError {
	return nil
}
