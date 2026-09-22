package stack

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("deprovisionReconciler.deleteResourcesFromDB", func() {
	var (
		ctrl              *gomock.Controller
		stackSvc          *MockstackService
		volumeSvc         *MockvolumeService
		namespaceSvc      *MocknamespaceService
		validationRecords *mocks.MockResourceValidationRecordStore
		reconciler        *deprovisionReconciler
		ctx               context.Context
		stack             *models.Stack
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		stackSvc = NewMockstackService(ctrl)
		volumeSvc = NewMockvolumeService(ctrl)
		namespaceSvc = NewMocknamespaceService(ctrl)
		validationRecords = mocks.NewMockResourceValidationRecordStore(ctrl)
		ctx = context.Background()
		reconciler = &deprovisionReconciler{
			stackService:      stackSvc,
			volumeService:     volumeSvc,
			namespaceService:  namespaceSvc,
			validationRecords: validationRecords,
			logger:            logger.NewLoggerWithPrefix(ctx, "test"),
		}
		stack = &models.Stack{ID: "stack-1", NamespaceID: "ns-1"}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("deletes validation records before the stack, volumes, and namespace", func() {
		gomock.InOrder(
			validationRecords.EXPECT().DeleteByStack(ctx, stack.ID).Return(nil),
			stackSvc.EXPECT().InternalDeleteFromDB(ctx, stack.ID).Return(nil),
			volumeSvc.EXPECT().InternalDeleteVolumesUsedByStackFromDB(ctx, stack.ID).Return(nil),
			namespaceSvc.EXPECT().InternalDeleteFromDB(ctx, stack.NamespaceID).Return(nil),
		)

		Expect(reconciler.deleteResourcesFromDB(ctx, stack)).To(Succeed())
	})

	It("does not delete the stack if validation-record cleanup fails", func() {
		validationRecords.EXPECT().DeleteByStack(ctx, stack.ID).Return(errors.GeneralError("boom"))

		err := reconciler.deleteResourcesFromDB(ctx, stack)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to delete resource validation records"))
	})
})
