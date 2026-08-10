package postgresaddon

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

	Describe("GetInput", func() {
		It("selects pending, failed, and deleting addons", func() {
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
			Expect(operands).To(ConsistOf(worker.Operand(models.PostgresAddonOperand{ID: "addon-1"})))
		})
	})

	It("reconciles a persisted addon without release authorization", func() {
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}
		addonSvc.EXPECT().InternalGetPostgresAddon(ctx, addon.ID).Return(addon, nil).Times(2)
		w.subReconcilers = []subReconciler{&authorizingAddonReconciler{}}

		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: addon.ID})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("stops a cluster mutation when addon deletion begins", func() {
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}
		deleting := *addon
		deletedAt := time.Now().UTC()
		deleting.DeletionTimestamp = &deletedAt
		gomock.InOrder(
			addonSvc.EXPECT().InternalGetPostgresAddon(ctx, addon.ID).Return(addon, nil),
			addonSvc.EXPECT().InternalGetPostgresAddon(ctx, addon.ID).Return(&deleting, nil),
		)
		w.subReconcilers = []subReconciler{&authorizingAddonReconciler{}}

		result, serr := w.Execute(ctx, models.PostgresAddonOperand{ID: addon.ID})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})
})

type authorizingAddonReconciler struct{}

func (*authorizingAddonReconciler) Name() string { return "authorize-mutation" }
func (*authorizingAddonReconciler) Reconcile(ctx context.Context, _ *models.PostgresAddon, authorize worker.MutationAuthorizer) (subReconcilerResult, error) {
	return resultNil, authorize(ctx)
}
