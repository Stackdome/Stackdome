package volume

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/worker"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("VolumeWorker", func() {
	var (
		ctrl      *gomock.Controller
		volumeSvc *MockvolumeService
		w         *volumeWorker
		ctx       context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		volumeSvc = NewMockvolumeService(ctrl)
		ctx = context.Background()

		w = &volumeWorker{
			volumeService: volumeSvc,
			BaseWorker:    worker.NewBaseWorker(VolumeWorkerName, "test"),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetInput", func() {
		It("enlists not-ready volumes so lost enqueues are retried", func() {
			volumeSvc.EXPECT().InternalListNotReady(gomock.Any()).
				Return([]*models.Volume{{ID: "vol-1"}, {ID: "vol-2"}}, nil)

			operands, err := w.GetInput(ctx)
			Expect(err).To(BeNil())
			Expect(operands).To(HaveLen(2))
			Expect(operands[0]).To(Equal(worker.Operand(models.VolumeOperand{ID: "vol-1"})))
			Expect(operands[1]).To(Equal(worker.Operand(models.VolumeOperand{ID: "vol-2"})))
		})
	})

	Describe("Interval", func() {
		It("runs periodically", func() {
			var periodic worker.PeriodicReconcilable = w
			Expect(periodic.Interval()).To(Equal(30 * time.Second))
		})
	})
})
