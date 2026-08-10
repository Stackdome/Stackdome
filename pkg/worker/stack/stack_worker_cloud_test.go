package stack

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackWorker cloud admission", func() {
	It("does not reconcile a pending draft without an active allocation", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		policy := &inactiveWorkerRuntimePolicy{}
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1", Status: &models.StackStatus{State: models.StackPending}}, nil)
		w := &stackWorker{stackService: stacks, runtimePolicy: policy, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test")}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1"})
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

type inactiveWorkerRuntimePolicy struct{}

func (*inactiveWorkerRuntimePolicy) OrganisationProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveWorkerRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveWorkerRuntimePolicy) IsolationPolicyVersion() string { return "v1" }
func (*inactiveWorkerRuntimePolicy) AdmitFirstReleaseWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) AdmitRollbackWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return errors.TrialInactive()
}
func (*inactiveWorkerRuntimePolicy) AdmitMutationWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
