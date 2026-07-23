package services

import (
	"context"
	"encoding/base64"
	stderrors "errors"

	"github.com/Stackdome/stackdome/pkg/auth"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	certutil "k8s.io/client-go/util/cert"
)

// Suite bootstrapped by TestServices in services_suite_test.go.

var _ = Describe("ClusterService", func() {
	const masterKey = "this-is-a-very-secure-master-key-that-is-at-least-64-characters-long-for-security-validation"

	var (
		ctrl           *gomock.Controller
		clusterStore   *mocks.MockClusterStore
		clusterManager *mocks.MockClusterManager
		orgStore       *mocks.MockOrganisationStore
		registrySvc    *mocks.MockImageRegistryService
		encryption     EncryptionService
		svc            *clusterService
		ctx            context.Context
		caData         string
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		clusterStore = mocks.NewMockClusterStore(ctrl)
		clusterManager = mocks.NewMockClusterManager(ctrl)

		var err error
		encryption, err = NewAESEncryptionService(EncryptionServiceSpec{Masterkey: masterKey})
		Expect(err).ToNot(HaveOccurred())

		certPEM, _, certErr := certutil.GenerateSelfSignedCertKey("localhost", nil, nil)
		Expect(certErr).ToNot(HaveOccurred())
		caData = base64.StdEncoding.EncodeToString(certPEM)

		orgStore = mocks.NewMockOrganisationStore(ctrl)
		registrySvc = mocks.NewMockImageRegistryService(ctrl)
		svc = &clusterService{
			clusterStore:         clusterStore,
			organisationStore:    orgStore,
			imageRegistryService: registrySvc,
			clusterManager:       clusterManager,
			logger:               logger.NewLogger(),
			permissions:          auth.NewPermissionService(auth.PermissionServiceSpec{}),
			encryptionService:    encryption,
		}

		ctx = auth.SetIdentityInContext(context.Background(), &auth.Identity{IsSystem: true})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	encryptCluster := func(cluster *models.Cluster) {
		Expect(svc.encryptClusterCredentials(cluster)).To(BeNil())
	}

	Describe("GetClusterForOrg read-time fallback", func() {
		It("returns the org's own cluster without consulting the platform cluster", func() {
			owned := &models.Cluster{ID: "cluster-owned", OrganisationID: "org-1", Token: "tok", ClusterCAData: caData}
			encryptCluster(owned)

			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-1").Return(owned, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-owned"))
		})

		It("falls back to the platform cluster when the org owns none", func() {
			def := &models.Cluster{ID: "cluster-platform", OrganisationID: "org-platform", Platform: true, Token: "tok", ClusterCAData: caData}
			encryptCluster(def)

			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-1").
				Return(nil, apperrors.NotFound("cluster for organisation 'org-1' not found"))
			clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).Return(def, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-platform"))
		})

		It("returns NotFound when neither an owned nor a platform cluster exists", func() {
			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-1").
				Return(nil, apperrors.NotFound("cluster for organisation 'org-1' not found"))
			clusterStore.EXPECT().GetPlatformCluster(gomock.Any()).
				Return(nil, apperrors.NotFound("platform cluster not found"))

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(apperrors.ErrorNotFound))
		})
	})

	Describe("GetOwnedClusterForOrg", func() {
		It("returns the org's own cluster when it owns one", func() {
			owned := &models.Cluster{ID: "cluster-owned", OrganisationID: "org-1", Token: "tok", ClusterCAData: caData}
			encryptCluster(owned)

			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-1").Return(owned, nil)

			result, err := svc.GetOwnedClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-owned"))
		})

		It("returns NotFound without falling back to the platform cluster when the org owns none", func() {
			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-1").
				Return(nil, apperrors.NotFound("cluster for organisation 'org-1' not found"))

			result, err := svc.GetOwnedClusterForOrg(ctx, "org-1")
			Expect(result).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err.Code).To(Equal(apperrors.ErrorNotFound))
		})
	})

	Describe("AddCluster registry provisioning", func() {
		const tenantOrgID = "org-tenant"

		newClusterSpec := func() *models.Cluster {
			return &models.Cluster{
				Name:           "owned",
				OrganisationID: tenantOrgID,
				ClusterURL:     "https://example.com:6443",
				Token:          "tok",
				ClusterCAData:  caData,
			}
		}

		expectClusterCreated := func(created *models.Cluster) {
			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), tenantOrgID).
				Return(nil, apperrors.NotFound("no cluster"))
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NotFound("not found")).AnyTimes()
			clusterStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
					return fn(ctx)
				})
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			clusterManager.EXPECT().GetClient(created.ID).Return(nil, stderrors.New("no client in test"))
		}

		It("auto-creates a default registry named <slug>-<shortOrgID> when none is supplied", func() {
			orgID := "11112222-3333-4444-5555-666677778888"
			spec := newClusterSpec()
			spec.OrganisationID = orgID
			created := &models.Cluster{ID: "cluster-owned", OrganisationID: orgID}

			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), orgID).
				Return(nil, apperrors.NotFound("no cluster"))
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NotFound("not found")).AnyTimes()
			orgStore.EXPECT().Get(gomock.Any(), orgID).
				Return(&models.Organisation{ID: orgID, Name: "Acme Inc"}, nil)
			clusterStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
					return fn(ctx)
				})
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			registrySvc.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, r *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *apperrors.ServiceError) {
					Expect(r.Name).To(Equal("acme-inc-11112222"))
					Expect(r.ClusterID).To(Equal("cluster-owned"))
					Expect(r.OrganisationID).To(Equal(orgID))
					return r, nil
				})
			clusterManager.EXPECT().GetClient(created.ID).Return(nil, stderrors.New("no client in test"))

			result, err := svc.AddCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ImageRegistries).To(HaveLen(1))
		})

		It("keeps the explicitly supplied registry", func() {
			spec := newClusterSpec()
			spec.ImageRegistries = []*models.ClusterImageRegistry{{Name: "custom-registry"}}
			created := &models.Cluster{ID: "cluster-owned", OrganisationID: tenantOrgID}

			expectClusterCreated(created)
			registrySvc.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, r *models.ClusterImageRegistry) (*models.ClusterImageRegistry, *apperrors.ServiceError) {
					Expect(r.Name).To(Equal("custom-registry"))
					return r, nil
				})

			result, err := svc.AddCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ImageRegistries).To(HaveLen(1))
		})

		It("creates no registry for the platform org", func() {
			spec := newClusterSpec()
			created := &models.Cluster{ID: "cluster-platform", OrganisationID: tenantOrgID}

			expectClusterCreated(created)
			orgStore.EXPECT().Get(gomock.Any(), tenantOrgID).
				Return(&models.Organisation{ID: tenantOrgID, Name: "platform", Platform: true}, nil)

			result, err := svc.AddCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ImageRegistries).To(BeEmpty())
		})
	})

	Describe("InternalUpsertPlatformCluster", func() {
		It("delegates to AddCluster when no cluster exists for the URL", func() {
			spec := &models.Cluster{
				Name:           "default",
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          "tok",
				ClusterCAData:  caData,
			}
			created := &models.Cluster{ID: "cluster-new", OrganisationID: "org-platform"}

			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), spec.ClusterURL).
				Return(nil, apperrors.NotFound("cluster with this api URL not found")).AnyTimes()
			clusterStore.EXPECT().GetClusterForOrg(gomock.Any(), "org-platform").
				Return(nil, apperrors.NotFound("cluster for organisation 'org-platform' not found"))
			orgStore.EXPECT().Get(gomock.Any(), "org-platform").
				Return(&models.Organisation{ID: "org-platform", Name: "platform", Platform: true}, nil)
			clusterStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
					return fn(ctx)
				})
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			clusterManager.EXPECT().GetClient(created.ID).Return(nil, stderrors.New("no client in test"))

			result, err := svc.InternalUpsertPlatformCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-new"))
		})

		It("is a no-op when the stored credentials already match", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Token: token, ClusterCAData: caData}
			encryptCluster(existing)

			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)

			result, err := svc.InternalUpsertPlatformCluster(ctx, &models.Cluster{
				ClusterURL:    "https://example.com:6443",
				Token:         token,
				ClusterCAData: caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
		})

		It("rotates credentials and re-registers the cluster when they change", func() {
			oldToken := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			newToken := base64.StdEncoding.EncodeToString([]byte("token-v2"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Token: oldToken, ClusterCAData: caData}
			encryptCluster(existing)
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform"}

			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)
			clusterStore.EXPECT().UpdateCredentials(gomock.Any(), "cluster-1", gomock.Not(""), gomock.Not("")).Return(nil)
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(fresh, nil)
			clusterManager.EXPECT().ReRegisterCluster(fresh).Return(nil)

			result, err := svc.InternalUpsertPlatformCluster(ctx, &models.Cluster{
				ClusterURL:    "https://example.com:6443",
				Token:         newToken,
				ClusterCAData: caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
		})
	})
})
