package postgresaddon

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("PostgresAddonWorker", func() {
	var (
		ctrl     *gomock.Controller
		addonSvc *MockpostgresAddonService
		w        *postgresAddonWorker
		ctx      context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonSvc = NewMockpostgresAddonService(ctrl)
		ctx = context.Background()

		w = &postgresAddonWorker{
			postgresAddonService: addonSvc,
			runtimePolicy:        &eagerAddonRuntimePolicy{},
			BaseWorker:           worker.NewBaseWorker(WorkerName, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetInput", func() {
		It("picks up addons marked for deletion regardless of their status state", func() {
			// Regression: an addon whose Deleting state was clobbered by a
			// status-path write must still be selected via deletion_timestamp.
			addonSvc.EXPECT().InternalList(gomock.Any(),
				"status->>'state' IN ? OR deletion_timestamp IS NOT NULL",
				[]string{
					string(models.PostgresAddonStatePending),
					string(models.PostgresAddonStateError),
					string(models.PostgresAddonStateDeleting),
				},
			).Return([]*models.PostgresAddon{{ID: "addon-1"}}, nil)

			operands, err := w.GetInput(ctx)
			Expect(err).To(BeNil())
			Expect(operands).To(HaveLen(1))
			Expect(operands[0]).To(Equal(worker.Operand(models.PostgresAddonOperand{ID: "addon-1"})))
		})
	})

	It("does not reconcile a pending cloud addon without an active allocation", func() {
		addonSvc.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(&models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}, nil)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		release := &models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStatePending, Snapshot: models.StackSnapshot{
				Stack:       models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1"},
				Connections: models.StackConnections{{From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "addon-1"}}},
			},
		}
		releases.EXPECT().InternalGet(ctx, "release-1").Return(release, nil)
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1"}
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(ctx, stack).Return(release, nil)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(stack, nil)
		w.releaseService = releases
		w.stackService = stacks
		w.runtimePolicy = &inactiveAddonRuntimePolicy{}
		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: "addon-1", ReleaseID: "release-1"})
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("does not select draft-only addons for periodic cloud reconciliation", func() {
		releases := NewMockreleaseService(ctrl)
		releases.EXPECT().InternalListAuthoritativeWorkloadReleases(ctx).Return([]*models.StackRelease{}, nil)
		addonSvc.EXPECT().InternalList(ctx, "deletion_timestamp IS NOT NULL").Return([]*models.PostgresAddon{}, nil)
		w.releaseService = releases
		w.runtimePolicy = &activeAddonRuntimePolicy{}

		operands, serr := w.GetInput(ctx)
		Expect(serr).To(BeNil())
		Expect(operands).To(BeEmpty())
	})

	It("periodically reconciles an addon referenced by a release while its allocation is active", func() {
		refs := NewMockreferenceService(ctrl)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		reconciler := &recordingAddonReconciler{}
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}
		releaseID := "release-1"
		addonSvc.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(addon, nil)
		refs.EXPECT().IsReferentInUse(ctx, models.ReferentPostgresAddon, "addon-1").Return(true, []models.ResourceReference{{ReleaseID: &releaseID, ReferentID: "addon-1"}}, nil)
		release := &models.StackRelease{
			ID: releaseID, StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{
				Stack:       models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1"},
				Connections: models.StackConnections{{From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "addon-1"}}},
			},
		}
		releases.EXPECT().InternalGet(ctx, releaseID).Return(release, nil).Times(2)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1",
			Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseID}},
		}
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(stack, nil).Times(2)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(ctx, stack).Return(release, nil).Times(2)
		w.referenceService = refs
		w.releaseService = releases
		w.stackService = stacks
		w.subReconcilers = []subReconciler{reconciler}
		w.runtimePolicy = &activeAddonRuntimePolicy{}

		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: "addon-1"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
		Expect(reconciler.called).To(BeTrue())
	})

	It("does not reconcile an addon from a historical release reference", func() {
		refs := NewMockreferenceService(ctrl)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		reconciler := &recordingAddonReconciler{}
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}
		releaseID := "release-a"
		addonSvc.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(addon, nil)
		refs.EXPECT().IsReferentInUse(ctx, models.ReferentPostgresAddon, "addon-1").Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		releases.EXPECT().InternalGet(ctx, releaseID).Return(&models.StackRelease{
			ID: releaseID, StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1"}, Connections: models.StackConnections{{From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "addon-1"}}}},
		}, nil)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-b"}},
		}
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(stack, nil)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(ctx, stack).Return(nil, nil)
		w.referenceService = refs
		w.releaseService = releases
		w.stackService = stacks
		w.subReconcilers = []subReconciler{reconciler}
		w.runtimePolicy = &activeAddonRuntimePolicy{}

		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: "addon-1"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
		Expect(reconciler.called).To(BeFalse())
	})
})

type recordingAddonReconciler struct{ called bool }

func (*recordingAddonReconciler) Name() string { return "test-reconciler" }
func (r *recordingAddonReconciler) Reconcile(context.Context, *models.PostgresAddon, worker.MutationAuthorizer) (subReconcilerResult, error) {
	r.called = true
	return resultNil, nil
}

type activeAddonRuntimePolicy struct{ inactiveAddonRuntimePolicy }

func (*activeAddonRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return nil
}

type eagerAddonRuntimePolicy struct{ inactiveAddonRuntimePolicy }

func (*eagerAddonRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeEager
}

type inactiveAddonRuntimePolicy struct{}

func (*inactiveAddonRuntimePolicy) OrganisationProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveAddonRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveAddonRuntimePolicy) IsolationPolicyVersion() string { return "v1" }
func (*inactiveAddonRuntimePolicy) AdmitFirstReleaseWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveAddonRuntimePolicy) AdmitRollbackWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveAddonRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return errors.TrialInactive()
}
func (*inactiveAddonRuntimePolicy) AdmitMutationWithTx(context.Context, string) (services.MutationAdmission, *errors.ServiceError) {
	return services.MutationAdmission{}, nil
}
func (*inactiveAddonRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveAddonRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveAddonRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
