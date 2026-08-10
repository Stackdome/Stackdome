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
	It("does not reconcile an unrelated pending draft selected by the periodic scanner", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		policy := &activeWorkerRuntimePolicy{}
		stacks.EXPECT().InternalList(gomock.Any(), gomock.Any(), gomock.Any()).Return([]*models.Stack{{ID: "stack-1"}}, nil)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1", Status: &models.StackStatus{State: models.StackPending}}, nil)
		w := &stackWorker{stackService: stacks, runtimePolicy: policy, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test")}

		operands, serr := w.GetInput(context.Background())
		Expect(serr).To(BeNil())
		Expect(operands).To(ConsistOf(models.StackOperand{ID: "stack-1"}))
		result, serr := w.Execute(context.Background(), operands[0])
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("periodically reconciles an existing released workload while its allocation is active", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1",
			Status: &models.StackStatus{State: models.StackPending, LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-1"}},
		}
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil).Times(2)
		releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(&models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}},
		}, nil).Times(2)
		reconciler.EXPECT().Name().Return("test-reconciler")
		reconciler.EXPECT().Reconcile(gomock.Any(), stack).Return(resultNil, nil)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: &activeWorkerRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("does not replay a historical released stack after a newer release converges", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1",
			Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-b"}},
		}
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil)
		releases.EXPECT().InternalGet(gomock.Any(), "release-a").Return(&models.StackRelease{
			ID: "release-a", StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}},
		}, nil)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: &activeWorkerRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1", ReleaseID: "release-a"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("fails closed when an active release is cancelled before reconciliation", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}
		pending := &models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}},
		}
		cancelled := *pending
		cancelled.State = models.ReleaseStateCancelled
		gomock.InOrder(
			stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil),
			releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(pending, nil),
			releases.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(pending, nil),
			stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil),
			releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(&cancelled, nil),
		)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: &activeWorkerRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1", ReleaseID: "release-1"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

type activeWorkerRuntimePolicy struct{ inactiveWorkerRuntimePolicy }

func (*activeWorkerRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return nil
}

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
func (*inactiveWorkerRuntimePolicy) AdmitMutationWithTx(context.Context, string) (services.MutationAdmission, *errors.ServiceError) {
	return services.MutationAdmission{}, nil
}
func (*inactiveWorkerRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
