package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("ClusterImageRegistry cluster ownership guard", func() {
	const (
		guardOrgID       = "org-guard"
		ownedClusterID   = "cluster-owned"
		sharedCluster    = "cluster-shared"
		foreignClusterID = "cluster-foreign"
	)

	var (
		ctrl         *gomock.Controller
		clusterStore *mocks.MockClusterStore
		perms        *mocks.MockPermissionService
		svc          *clusterImageRegistryService
		ctx          context.Context
	)

	newSpec := func(clusterID string) *models.ClusterImageRegistry {
		return &models.ClusterImageRegistry{
			ClusterID:      clusterID,
			OrganisationID: guardOrgID,
			Name:           "reg",
		}
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		clusterStore = mocks.NewMockClusterStore(ctrl)
		perms = mocks.NewMockPermissionService(ctrl)
		perms.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		svc = &clusterImageRegistryService{
			clusterStore: clusterStore,
			permissions:  perms,
			logger:       logger.NewLogger(),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Create", func() {
		It("rejects a cluster the org does not own when it owns another", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), guardOrgID).
				Return([]*models.Cluster{{ID: ownedClusterID}}, nil)

			_, serr := svc.Create(ctx, newSpec(foreignClusterID))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorNotFound))
		})

		It("rejects the shared-compute cluster when the org owns none — seed registries are not an API path", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), guardOrgID).Return(nil, nil)

			_, serr := svc.Create(ctx, newSpec(sharedCluster))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorNotFound))
		})

		It("propagates non-NotFound owned-cluster lookup errors", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), guardOrgID).
				Return(nil, errors.GeneralError("db down"))

			_, serr := svc.Create(ctx, newSpec(ownedClusterID))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorGeneral))
		})

		It("accepts the org's own cluster", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), guardOrgID).
				Return([]*models.Cluster{{ID: ownedClusterID}}, nil)

			Expect(svc.validateOwnedCluster(ctx, guardOrgID, ownedClusterID)).To(BeNil())
		})
	})

	Describe("PopulateInClusterRegistryNameForResource", func() {
		newBuildResource := func() *models.StackResource {
			return &models.StackResource{
				Name: "web",
				BuildConfig: &models.BuildConfigSpec{
					BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true},
				},
			}
		}

		It("resolves the registry on the stack's cluster", func() {
			registryStore := mocks.NewMockClusterImageRegistryStore(ctrl)
			svc.clusterImageRegistryStore = registryStore
			registryStore.EXPECT().GetForOrgAndCluster(gomock.Any(), guardOrgID, ownedClusterID).
				Return(&models.ClusterImageRegistry{Name: "owned-reg"}, nil)

			resource := newBuildResource()
			serr := svc.PopulateInClusterRegistryNameForResource(ctx, guardOrgID, ownedClusterID, "stack", resource)
			Expect(serr).To(BeNil())
			Expect(resource.BuildConfig.BuildImageRepository.ClusterRegistryName).To(Equal("owned-reg"))
		})

		It("fails fast when the stack's cluster has no registry, even if another cluster does", func() {
			registryStore := mocks.NewMockClusterImageRegistryStore(ctrl)
			svc.clusterImageRegistryStore = registryStore
			registryStore.EXPECT().GetForOrgAndCluster(gomock.Any(), guardOrgID, ownedClusterID).
				Return(nil, errors.NotFound("cluster image registry not found"))

			serr := svc.PopulateInClusterRegistryNameForResource(ctx, guardOrgID, ownedClusterID, "stack", newBuildResource())
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorBadRequest))
		})

		It("is a no-op for resources not using the in-cluster registry", func() {
			serr := svc.PopulateInClusterRegistryNameForResource(ctx, guardOrgID, ownedClusterID, "stack", &models.StackResource{Name: "web"})
			Expect(serr).To(BeNil())
		})
	})

	Describe("InternalCreateSeedRegistry", func() {
		It("rejects a target that is not a shared-compute cluster", func() {
			clusterStore.EXPECT().Get(gomock.Any(), ownedClusterID).
				Return(&models.Cluster{ID: ownedClusterID}, nil)

			_, serr := svc.InternalCreateSeedRegistry(ctx, newSpec(ownedClusterID))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorNotFound))
		})

		It("propagates a missing shared-compute cluster", func() {
			clusterStore.EXPECT().Get(gomock.Any(), sharedCluster).
				Return(nil, errors.NotFound("shared-compute cluster not found"))

			_, serr := svc.InternalCreateSeedRegistry(ctx, newSpec(sharedCluster))
			Expect(serr).ToNot(BeNil())
			Expect(serr.Code).To(Equal(errors.ErrorNotFound))
		})

		It("creates the registry on a shared-compute cluster", func() {
			clusterStore.EXPECT().Get(gomock.Any(), sharedCluster).
				Return(&models.Cluster{ID: sharedCluster, SharedCompute: true}, nil)

			registryStore := mocks.NewMockClusterImageRegistryStore(ctrl)
			crs := mocks.NewMockClusterResourceImageRegistryService(ctrl)
			svc.clusterImageRegistryStore = registryStore
			svc.clusterResourceService = crs

			spec := newSpec(sharedCluster)
			registryStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				})
			registryStore.EXPECT().CreateWithTx(gomock.Any(), spec).Return(spec, nil)
			crs.EXPECT().CreateImageRegistryInCluster(gomock.Any(), spec).Return(nil)

			created, serr := svc.InternalCreateSeedRegistry(ctx, spec)
			Expect(serr).To(BeNil())
			Expect(created).To(Equal(spec))
			Expect(created.BackendStorageSize).To(Equal(models.DefaultRegistryStorageSize))
		})
	})
})
