package stack

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackWorker cloud behavior", func() {
	It("does not select draft-only stacks for periodic cloud reconciliation", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		releases.EXPECT().InternalListAuthoritativeWorkload(gomock.Any()).Return(&models.WorkloadAuthorityScan{}, nil)
		w := &stackWorker{stackService: stacks, releaseService: releases, BaseWorker: worker.NewBaseWorker(StackWorkerName, "test")}

		operands, serr := w.GetInput(context.Background())
		Expect(serr).To(BeNil())
		Expect(operands).To(BeEmpty())
	})

	It("periodically reconciles an existing released workload", func() {
		ctrl := gomock.NewController(GinkgoT())
		stacks := NewMockstackService(ctrl)
		releases := NewMockreleaseService(ctrl)
		reconciler := NewMocksubReconciler(ctrl)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1",
			Status: &models.StackStatus{State: models.StackPending, LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-1"}},
		}
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil)
		release := &models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", Namespace: "stack-1"}},
		}
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil)
		reconciler.EXPECT().Name().Return("test-reconciler")
		reconciler.EXPECT().Reconcile(gomock.Any(), stack).Return(resultNil, nil)
		w := &stackWorker{
			stackService: stacks, releaseService: releases, subReconcilers: []subReconciler{reconciler},
			BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
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
			BaseWorker: worker.NewBaseWorker(StackWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.StackOperand{ID: "stack-1", ReleaseID: "release-a"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

})
