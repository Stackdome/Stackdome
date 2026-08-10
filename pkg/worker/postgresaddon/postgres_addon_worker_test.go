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
		w.runtimePolicy = &inactiveAddonRuntimePolicy{}
		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: "addon-1"})
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

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
func (*inactiveAddonRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveAddonRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
