package release

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("deadlineReconciler", func() {
	var (
		ctrl       *gomock.Controller
		releaseSvc *MockreleaseService
		r          *deadlineReconciler
		ctx        context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		releaseSvc = NewMockreleaseService(ctrl)
		r = newDeadlineReconciler(ReleaseWorkerSpec{ReleaseService: releaseSvc})
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("fails an InProgress release older than the convergence timeout", func() {
		release := &models.StackRelease{
			ID:        "rel-1",
			State:     models.ReleaseStateInProgress,
			CreatedAt: time.Now().Add(-convergenceTimeout - time.Minute),
		}
		releaseSvc.EXPECT().
			MarkFailed(ctx, "rel-1", "release did not converge within 45 minutes", nil).
			Return(true, nil)

		result, err := r.Reconcile(ctx, release)
		Expect(err).To(BeNil())
		Expect(result.resultStop).To(BeTrue())
	})

	It("leaves an InProgress release inside the timeout alone", func() {
		release := &models.StackRelease{
			ID:        "rel-1",
			State:     models.ReleaseStateInProgress,
			CreatedAt: time.Now().Add(-time.Minute),
		}

		result, err := r.Reconcile(ctx, release)
		Expect(err).To(BeNil())
		Expect(result.resultNil).To(BeTrue())
	})

	It("ignores releases not yet InProgress even when old", func() {
		release := &models.StackRelease{
			ID:        "rel-1",
			State:     models.ReleaseStatePending,
			CreatedAt: time.Now().Add(-convergenceTimeout - time.Hour),
		}

		result, err := r.Reconcile(ctx, release)
		Expect(err).To(BeNil())
		Expect(result.resultNil).To(BeTrue())
	})
})
