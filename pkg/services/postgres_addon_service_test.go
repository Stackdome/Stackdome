package services

import (
	"context"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

// Suite bootstrapped by TestAESEncryptionService in encryption_service_test.go.

var _ = Describe("PostgresAddonService DeletePostgresAddon", func() {
	var (
		ctrl        *gomock.Controller
		addonStore  *mocks.MockPostgresAddonStore
		permissions *mocks.MockPermissionService
		refs        *mocks.MockReferenceService
		enqueuer    *mocks.MockBackgroundJobEnqueuer
		svc         *postgresAddonService
		ctx         context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonStore = mocks.NewMockPostgresAddonStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		refs = mocks.NewMockReferenceService(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		ctx = context.Background()

		svc = &postgresAddonService{
			logger:             logger.NewLogger(),
			postgresAddonStore: addonStore,
			referenceService:   refs,
			permissions:        permissions,
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("records deletion intent in the deletion timestamp column, not only the status state", func() {
		addonStore.EXPECT().GetByID(gomock.Any(), "pg-1").
			Return(&models.PostgresAddon{ID: "pg-1", ProjectID: "project-1", Name: "test-delete"}, nil)
		permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
		refs.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentPostgresAddon, "pg-1").
			Return(false, nil, nil)

		addonStore.EXPECT().UpdateDeletionTimestamp(gomock.Any(), "pg-1", gomock.Not(gomock.Nil())).Return(nil)
		addonStore.EXPECT().UpdateStatus(gomock.Any(), "pg-1", gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, status *models.PostgresAddonStatus) *apperrors.ServiceError {
				Expect(status.State).To(Equal(models.PostgresAddonStateDeleting))
				return nil
			})
		enqueuer.EXPECT().Enqueue(&models.PostgresAddon{ID: "pg-1"}).Return(nil)

		deleted, err := svc.DeletePostgresAddon(ctx, "pg-1")
		Expect(err).To(BeNil())
		Expect(deleted.DeletionTimestamp).NotTo(BeNil())
	})
})
