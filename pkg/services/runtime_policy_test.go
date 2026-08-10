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
	It("keeps self-hosted compute policy unrestricted", func() {
		policy := NewSelfHostedRuntimePolicy()
		Expect(policy.EnsureComputeAccess(context.Background(), "org-1")).To(BeNil())
		Expect(policy.ValidateOrganisationDeletion(context.Background(), "org-1")).To(BeNil())
		Expect(policy.IsolationPolicyVersion()).To(BeEmpty())
	})

	It("delegates cloud compute access and organisation deletion checks", func() {
		ctrl := gomock.NewController(GinkgoT())
		access := NewMockComputeAccessService(ctrl)
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess: access, ComputeUsage: &fakeComputeUsageStore{}, IsolationPolicyVersion: "policy-v1", Limits: testComputeLimits(),
		})
		access.EXPECT().Activate(gomock.Any(), "org-1").Return(nil, errors.ComputeAccessInactive())
		serr := policy.EnsureComputeAccess(context.Background(), "org-1")
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		access.EXPECT().EnsureNoLease(gomock.Any(), "org-1").Return(errors.BadRequest("allocation exists"))
		Expect(policy.ValidateOrganisationDeletion(context.Background(), "org-1").Reason).To(Equal("allocation exists"))
	})

	It("publishes the cloud isolation policy", func() {
		computeAccess := &fakeComputeAccessService{}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess:          computeAccess,
			ComputeUsage:           &fakeComputeUsageStore{},
			IsolationPolicyVersion: "policy-v1",
			Limits:                 testComputeLimits(),
		})
		Expect(policy.IsolationPolicyVersion()).To(Equal("policy-v1"))
	})

	It("enforces cloud stack and resource limits from locked usage", func() {
		stackLimits := &fakeComputeUsageStore{usage: stores.ComputeUsage{StackCount: 2, StackResourceCount: 5}}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess:          &fakeComputeAccessService{},
			ComputeUsage:           stackLimits,
			IsolationPolicyVersion: "policy-v1",
			Limits:                 testComputeLimits(),
		})
		serr := policy.ValidateStackLimits(context.Background(), StackLimitChange{Kind: StackLimitCreate, OrganisationID: "org-1", DesiredResourceCount: 1})
		Expect(serr).ToNot(BeNil())
		Expect(serr.Reason).To(ContainSubstring("maximum of 2 stacks"))

		serr = policy.ValidateStackLimits(context.Background(), StackLimitChange{Kind: StackLimitUpdate, OrganisationID: "org-1", StackID: "stack-1", DesiredResourceCount: 2})
		Expect(serr).ToNot(BeNil())
		Expect(serr.Reason).To(ContainSubstring("maximum of 6 stack resources"))
		Expect(stackLimits.excludeStackID).To(Equal("stack-1"))
	})

	It("checks a whole-stack update using persisted resources outside that stack plus the desired replacement", func() {
		stackLimits := &fakeComputeUsageStore{usage: stores.ComputeUsage{StackCount: 2, StackResourceCount: 2}}
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess:          &fakeComputeAccessService{},
			ComputeUsage:           stackLimits,
			IsolationPolicyVersion: "policy-v1",
			Limits:                 testComputeLimits(),
		})

		serr := policy.ValidateStackLimits(context.Background(), StackLimitChange{
			Kind: StackLimitUpdate, OrganisationID: "org-1", StackID: "stack-1", DesiredResourceCount: 4,
		})

		Expect(serr).To(BeNil())
		Expect(stackLimits.excludeStackID).To(Equal("stack-1"))
	})

	It("forces every cloud stack resource to the configured replica count", func() {
		policy := NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
			ComputeAccess:          &fakeComputeAccessService{},
			ComputeUsage:           &fakeComputeUsageStore{},
			IsolationPolicyVersion: "policy-v1",
			Limits:                 testComputeLimits(),
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
type fakeComputeAccessService struct{}

type fakeComputeUsageStore struct {
	usage          stores.ComputeUsage
	excludeStackID string
	called         bool
}

func newCloudRuntimePolicyForTest() RuntimePolicy {
	return newCloudRuntimePolicyWithStoreForTest(&fakeComputeUsageStore{})
}

func newCloudRuntimePolicyWithStoreForTest(computeUsage stores.ComputeUsageStore) RuntimePolicy {
	return NewStackdomeCloudRuntimePolicy(StackdomeCloudRuntimePolicySpec{
		ComputeAccess:          &fakeComputeAccessService{},
		ComputeUsage:           computeUsage,
		IsolationPolicyVersion: "policy-v1",
		Limits:                 testComputeLimits(),
	})
}

func (f *fakeComputeUsageStore) LockOrganisationAndGetUsageWithTx(_ context.Context, _, excludeStackID string) (stores.ComputeUsage, *errors.ServiceError) {
	f.called = true
	f.excludeStackID = excludeStackID
	return f.usage, nil
}

func testComputeLimits() models.ComputeLimits {
	return models.ComputeLimits{
		MaxStacksPerOrganization: 2, MaxStackResourcesPerOrganization: 6, ReplicasPerStackResource: 1,
		MaxVolumesPerOrganization: 2, MaxVolumeSize: "2Gi", MaxPostgresAddonsPerOrganization: 1,
		PostgresInstances: 1, MaxPostgresStorageSize: "2Gi", ConcurrentBuilds: 1,
	}
}

func (f *fakeComputeAccessService) Activate(context.Context, string) (*models.ComputeAccess, *errors.ServiceError) {
	return &models.ComputeAccess{}, nil
}

func (f *fakeComputeAccessService) EnsureNoLease(context.Context, string) *errors.ServiceError {
	return nil
}
