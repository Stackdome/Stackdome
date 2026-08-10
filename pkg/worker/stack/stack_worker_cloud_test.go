package stack

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackWorker cloud admission", func() {
	It("does not select draft-only stacks for periodic cloud reconciliation", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		policy := &activeWorkerRuntimePolicy{}
		releases.EXPECT().InternalListAuthoritativeWorkload(gomock.Any()).Return(&models.WorkloadAuthorityScan{}, nil)
		w := &stackWorker{stackService: stacks, releaseService: releases, runtimePolicy: policy, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test")}

		operands, serr := w.GetInput(context.Background())
		Expect(serr).To(BeNil())
		Expect(operands).To(BeEmpty())
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
		release := &models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}},
		}
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil).Times(2)
		reconciler.EXPECT().Name().Return("test-reconciler")
		reconciler.EXPECT().Reconcile(gomock.Any(), stack, gomock.Any()).Return(resultNil, nil)
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
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(nil, nil)
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
		gomock.InOrder(
			stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil),
			releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(pending, nil),
			stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil),
			releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(nil, nil),
		)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: &activeWorkerRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1", ReleaseID: "release-1"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("stops a cluster write when deletion begins after initial authorization", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		policy := services.NewMockRuntimePolicy(ctrl)
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}
		deleting := *stack
		deletedAt := time.Now().UTC()
		deleting.DeletionTimestamp = &deletedAt
		release := &models.StackRelease{
			ID: "release-1", StackID: stack.ID, State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{
				ID: stack.ID, OrganisationID: stack.OrganisationID, ClusterID: stack.ClusterID, Namespace: stack.Namespace,
			}},
		}
		gomock.InOrder(
			stacks.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil),
			releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil),
			stacks.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil),
			releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil),
			stacks.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(&deleting, nil),
		)
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireComputeAccess(gomock.Any(), stack.OrganisationID).Return(nil)
		reconciler.EXPECT().Name().Return("write-boundary")
		reconciler.EXPECT().Reconcile(gomock.Any(), stack, gomock.Any()).DoAndReturn(
			func(ctx context.Context, _ *models.Stack, authorize worker.MutationAuthorizer) (subReconcilerResult, error) {
				return resultNil, authorize(ctx)
			},
		)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: policy, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: stack.ID, ReleaseID: release.ID})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("stops a cluster write when the allocation expires after initial authorization", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		policy := services.NewMockRuntimePolicy(ctrl)
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}
		release := &models.StackRelease{
			ID: "release-1", StackID: stack.ID, State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{
				ID: stack.ID, OrganisationID: stack.OrganisationID, ClusterID: stack.ClusterID, Namespace: stack.Namespace,
			}},
		}
		stacks.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil).Times(3)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil).Times(3)
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		gomock.InOrder(
			policy.EXPECT().RequireComputeAccess(gomock.Any(), stack.OrganisationID).Return(nil),
			policy.EXPECT().RequireComputeAccess(gomock.Any(), stack.OrganisationID).Return(errors.ComputeAccessInactive()),
		)
		reconciler.EXPECT().Name().Return("write-boundary")
		reconciler.EXPECT().Reconcile(gomock.Any(), stack, gomock.Any()).DoAndReturn(
			func(ctx context.Context, _ *models.Stack, authorize worker.MutationAuthorizer) (subReconcilerResult, error) {
				return resultNil, authorize(ctx)
			},
		)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			runtimePolicy: policy, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: stack.ID, ReleaseID: release.ID})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

type activeWorkerRuntimePolicy struct{ inactiveWorkerRuntimePolicy }

func (*activeWorkerRuntimePolicy) RequireComputeAccess(context.Context, string) *errors.ServiceError {
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
func (*inactiveWorkerRuntimePolicy) ActivateComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) RequireComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) RequireComputeAccess(context.Context, string) *errors.ServiceError {
	return errors.ComputeAccessInactive()
}
func (*inactiveWorkerRuntimePolicy) AdmitComputeMutationWithTx(context.Context, string) (services.ComputeMutationAdmission, *errors.ServiceError) {
	return services.ComputeMutationAdmission{}, nil
}
func (*inactiveWorkerRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveWorkerRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
