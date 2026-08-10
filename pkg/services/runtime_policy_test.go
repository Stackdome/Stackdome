package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"

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
			IsolationPolicyVersion: "policy-v1",
		})
		Expect(policy.OrganisationProvisioningMode()).To(Equal(ProvisioningModeDatabaseOnly))
		Expect(policy.DraftProvisioningMode()).To(Equal(ProvisioningModeDatabaseOnly))
		Expect(policy.IsolationPolicyVersion()).To(Equal("policy-v1"))
	})
})

// fakeCloudTrialService keeps this test focused on runtime policy selection.
// Delegation details are covered by cloud_trial_service_test.go.
type fakeCloudTrialService struct {
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
