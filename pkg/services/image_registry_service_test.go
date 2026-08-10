package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ClusterImageRegistry service", func() {
	const (
		orgID     = "org-1"
		clusterID = "cluster-1"
	)

	var (
		ctx           context.Context
		ctrl          *gomock.Controller
		clusterStore  *mocks.MockClusterStore
		registryStore *mocks.MockClusterImageRegistryStore
		enqueuer      *mocks.MockBackgroundJobEnqueuer
		permissions   *mocks.MockPermissionService
		service       *clusterImageRegistryService
	)

	newRegistry := func() *models.ClusterImageRegistry {
		return &models.ClusterImageRegistry{
			ID:             "registry-1",
			ClusterID:      clusterID,
			OrganisationID: orgID,
			Name:           "registry-1",
		}
	}

	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		clusterStore = mocks.NewMockClusterStore(ctrl)
		registryStore = mocks.NewMockClusterImageRegistryStore(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		service = &clusterImageRegistryService{
			clusterImageRegistryStore: registryStore,
			clusterStore:              clusterStore,
			permissions:               permissions,
			logger:                    logger.NewLogger(),
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}
	})

	It("keeps the public create path scoped to an organisation-owned cluster", func() {
		permissions.EXPECT().Check(gomock.Any(), orgID, auth.ResourceImageRegistries, "", auth.ActionCreate).Return(nil)
		clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), orgID).
			Return([]*models.Cluster{{ID: "another-cluster"}}, nil)

		created, err := service.Create(ctx, newRegistry())
		Expect(created).To(BeNil())
		Expect(err).ToNot(BeNil())
		Expect(err.Code).To(Equal(errors.ErrorNotFound))
	})

	It("persists pending create intent and enqueues the cluster after commit", func() {
		registry := newRegistry()
		registryStore.EXPECT().CreateWithTx(ctx, registry).Return(registry, nil)
		enqueuer.EXPECT().EnqueueAfterCommit(ctx, models.ClusterImageRegistryOperand{ClusterID: clusterID}).Return(nil)

		created, err := service.CreateWithTx(ctx, registry)
		Expect(err).To(BeNil())
		Expect(created).To(Equal(registry))
		Expect(created.Status.State).To(Equal(models.RegistryStatePending))
		Expect(created.BackendStorageSize).To(Equal(models.DefaultRegistryStorageSize))
	})

	It("persists deleting intent and enqueues the cluster after commit", func() {
		registry := newRegistry()
		permissions.EXPECT().Check(gomock.Any(), orgID, auth.ResourceImageRegistries, registry.ID, auth.ActionDelete).Return(nil)
		registryStore.EXPECT().MarkDeletingWithTx(ctx, registry.ID).Return(nil)
		enqueuer.EXPECT().EnqueueAfterCommit(ctx, models.ClusterImageRegistryOperand{ClusterID: clusterID}).Return(nil)

		Expect(service.DeleteWithTx(ctx, orgID, registry)).To(BeNil())
	})
})
