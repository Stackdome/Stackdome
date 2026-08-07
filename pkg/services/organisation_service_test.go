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

var _ = Describe("Organisation.Platform", func() {
	var (
		ctrl  *gomock.Controller
		store *mocks.MockOrganisationStore
		svc   *organisationService
		ctx   context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		store = mocks.NewMockOrganisationStore(ctrl)
		svc = &organisationService{
			organisationStore: store,
			logger:            logger.NewLogger(),
		}
	})

	It("persists the Platform flag on InternalCreate and returns it from InternalGetPlatformOrg", func() {
		created := &models.Organisation{ID: "org-1", Name: "platform", Platform: true}

		store.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
				Expect(spec.Platform).To(BeTrue())
				return created, nil
			})
		store.EXPECT().Get(gomock.Any(), "org-1").Return(created, nil)

		org, serr := svc.InternalCreate(ctx, &models.Organisation{Name: "platform", Platform: true})
		Expect(serr).To(BeNil())
		Expect(org.Platform).To(BeTrue())

		store.EXPECT().GetPlatformOrg(gomock.Any()).Return(created, nil)

		got, serr := svc.InternalGetPlatformOrg(ctx)
		Expect(serr).To(BeNil())
		Expect(got.ID).To(Equal("org-1"))
		Expect(got.Platform).To(BeTrue())
	})

	It("returns ErrorNotFound from InternalGetPlatformOrg when no org is flagged platform", func() {
		store.EXPECT().GetPlatformOrg(gomock.Any()).Return(nil, errors.NotFound("platform organisation not found"))

		org, serr := svc.InternalGetPlatformOrg(ctx)
		Expect(org).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})
})

var _ = Describe("OrganisationService platform-infra seeding", func() {
	const (
		orgName      = "Acme Inc"
		orgID        = "11112222-3333-4444-5555-666677778888"
		expectedSlug = "acme-inc"
		registryName = expectedSlug + "-11112222-cluster1"
		storageSize  = "50Gi"
		storageClass = "standard"
	)

	var (
		ctrl         *gomock.Controller
		orgStore     *mocks.MockOrganisationStore
		clusterStore *mocks.MockClusterStore
		registrySvc  *mocks.MockImageRegistryService
		svc          *organisationService
		ctx          context.Context
	)

	tenantOrg := func() *models.Organisation {
		return &models.Organisation{Name: orgName}
	}

	expectOrgCreated := func() {
		store := &models.Organisation{ID: orgID, Name: orgName}
		orgStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(store, nil)
	}

	expectOrgFetched := func() {
		orgStore.EXPECT().Get(gomock.Any(), orgID).Return(&models.Organisation{ID: orgID, Name: orgName}, nil)
	}

	expectPlatformCluster := func() {
		clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).Return(&models.Cluster{ID: "cluster-1"}, nil)
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		orgStore = mocks.NewMockOrganisationStore(ctrl)
		clusterStore = mocks.NewMockClusterStore(ctrl)
		registrySvc = mocks.NewMockImageRegistryService(ctrl)
		svc = &organisationService{
			organisationStore:    orgStore,
			clusterStore:         clusterStore,
			imageRegistryService: registrySvc,
			orgRegistryDefaults:  models.OrgRegistryDefaults{StorageSize: storageSize, StorageClass: storageClass},
			logger:               logger.NewLogger(),
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("skips seeding when no platform cluster exists", func() {
		expectOrgCreated()
		clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).
			Return(nil, errors.NotFound("platform cluster not found"))
		expectOrgFetched()

		org, serr := svc.InternalCreate(ctx, tenantOrg())
		Expect(serr).To(BeNil())
		Expect(org).ToNot(BeNil())
	})

	It("does not seed the platform org itself", func() {
		created := &models.Organisation{ID: orgID, Name: orgName, Platform: true}
		orgStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(created, nil)
		orgStore.EXPECT().Get(gomock.Any(), orgID).Return(created, nil)

		org, serr := svc.InternalCreate(ctx, &models.Organisation{Name: orgName, Platform: true})
		Expect(serr).To(BeNil())
		Expect(org.Platform).To(BeTrue())
	})

	It("seeds a registry on the platform cluster", func() {
		expectOrgCreated()
		expectPlatformCluster()
		registrySvc.EXPECT().InternalCreateSeedRegistry(gomock.Any(), &models.ClusterImageRegistry{
			ClusterID:           "cluster-1",
			OrganisationID:      orgID,
			Name:                registryName,
			BackendStorageSize:  storageSize,
			BackendStorageClass: storageClass,
		}).Return(&models.ClusterImageRegistry{}, nil)
		expectOrgFetched()

		org, serr := svc.InternalCreate(ctx, tenantOrg())
		Expect(serr).To(BeNil())
		Expect(org).ToNot(BeNil())
	})

	It("fails org creation when seed-registry creation fails", func() {
		expectOrgCreated()
		expectPlatformCluster()
		boom := errors.GeneralError("cluster unreachable")
		registrySvc.EXPECT().InternalCreateSeedRegistry(gomock.Any(), gomock.Any()).Return(nil, boom)

		org, serr := svc.InternalCreate(ctx, tenantOrg())
		Expect(org).To(BeNil())
		Expect(serr).To(Equal(boom))
	})
})

var _ = Describe("orgRegistryName", func() {
	const (
		nameOrgID     = "11112222-3333-4444-5555-666677778888"
		nameClusterID = "aaaabbbb-cccc-dddd-eeee-ffff00001111"
	)

	It("joins the slug and short IDs when within budget", func() {
		Expect(orgRegistryName("Acme Inc", nameOrgID, nameClusterID)).
			To(Equal("acme-inc-11112222-aaaabbbb"))
	})

	It("truncates the slug so the name fits Kubernetes label budgets", func() {
		name := orgRegistryName("Owned Cluster Org 1785184642064461000", nameOrgID, nameClusterID)
		Expect(len(name)).To(BeNumerically("<=", maxRegistryNameLength))
		Expect(name).To(HavePrefix("owned-cluster-org-"))
		Expect(name).To(HaveSuffix("-11112222-aaaabbbb"))
	})
})
