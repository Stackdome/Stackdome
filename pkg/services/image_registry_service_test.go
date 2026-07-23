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
		platformCluster  = "cluster-platform"
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

	It("rejects a cluster the org does not own when it owns another", func() {
		clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), guardOrgID).
			Return(&models.Cluster{ID: ownedClusterID}, nil)

		_, serr := svc.Create(ctx, newSpec(foreignClusterID))
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("rejects a non-platform cluster when the org owns none", func() {
		clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), guardOrgID).
			Return(nil, errors.NotFound("no cluster"))
		clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).
			Return(&models.Cluster{ID: platformCluster}, nil)

		_, serr := svc.Create(ctx, newSpec(foreignClusterID))
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("rejects any cluster when the org owns none and no platform cluster exists", func() {
		clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), guardOrgID).
			Return(nil, errors.NotFound("no cluster"))
		clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).
			Return(nil, errors.NotFound("platform cluster not found"))

		_, serr := svc.Create(ctx, newSpec(platformCluster))
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})

	It("accepts the org's own cluster", func() {
		clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), guardOrgID).
			Return(&models.Cluster{ID: ownedClusterID}, nil)

		Expect(svc.validateClusterForOrg(ctx, guardOrgID, ownedClusterID)).To(BeNil())
	})

	It("accepts the platform cluster when the org owns none", func() {
		clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), guardOrgID).
			Return(nil, errors.NotFound("no cluster"))
		clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).
			Return(&models.Cluster{ID: platformCluster}, nil)

		Expect(svc.validateClusterForOrg(ctx, guardOrgID, platformCluster)).To(BeNil())
	})
})
