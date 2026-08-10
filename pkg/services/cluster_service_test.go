package services

import (
	"context"
	"encoding/base64"
	stderrors "errors"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	certutil "k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
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

		It("rejects tenant cluster creation in shared compute mode", func() {
			permissions := mocks.NewMockPermissionService(ctrl)
			permissions.EXPECT().Check(ctx, tenantOrgID, auth.ResourceClusters, "", auth.ActionCreate).Return(nil)
			sharedService := NewClusterService(ClusterServiceSpec{
				ClusterStore:  clusterStore,
				SharedCompute: true,
				Permissions:   permissions,
			}).(*clusterService)

			result, err := sharedService.AddCluster(ctx, newClusterSpec())

			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: tenant clusters cannot be added when shared compute is enabled"))
			Expect(err.Code).To(Equal(apperrors.ErrorBadRequest))
		})

		It("returns an authorization failure before the shared compute policy", func() {
			permissions := mocks.NewMockPermissionService(ctrl)
			permissionError := apperrors.Forbidden("cluster creation is not allowed")
			permissions.EXPECT().Check(ctx, tenantOrgID, auth.ResourceClusters, "", auth.ActionCreate).
				Return(permissionError)
			sharedService := NewClusterService(ClusterServiceSpec{
				ClusterStore:  clusterStore,
				SharedCompute: true,
				Permissions:   permissions,
			}).(*clusterService)

			result, err := sharedService.AddCluster(ctx, newClusterSpec())

			Expect(result).To(BeNil())
			Expect(err).To(BeIdenticalTo(permissionError))
		})

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

		It("auto-creates a default registry named <slug>-<shortOrgID>-<shortClusterID> when none is supplied", func() {
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
					Expect(r.Name).To(Equal("acme-inc-11112222-clustero"))
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

			orgStore.EXPECT().Get(gomock.Any(), tenantOrgID).
				Return(&models.Organisation{ID: tenantOrgID, Name: "Tenant Org"}, nil)
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
		BeforeEach(func() {
			svc.sharedCompute = true
		})

		It("creates the platform cluster through the trusted internal path", func() {
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

		It("rejects platform topology in bring-your-own mode before accessing the store", func() {
			svc.sharedCompute = false

			result, err := svc.InternalUpsertPlatformCluster(ctx, &models.Cluster{})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: platform clusters require shared compute mode"))
			Expect(err.Code).To(Equal(apperrors.ErrorBadRequest))
		})

		It("is a no-op when the stored credentials already match", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Platform: true, Token: token, ClusterCAData: caData}
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

		It("normalizes raw credentials before comparing, so a raw env token is not a rotation", func() {
			raw := "token-v1"
			stored := base64.StdEncoding.EncodeToString([]byte(raw))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Platform: true, Token: stored, ClusterCAData: caData}
			encryptCluster(existing)

			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)

			result, err := svc.InternalUpsertPlatformCluster(ctx, &models.Cluster{
				ClusterURL:    "https://example.com:6443",
				Token:         raw,
				ClusterCAData: caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
		})

		It("reconciles drifted name and platform flag without rotating credentials", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Name: "old-name", Platform: false, Token: token, ClusterCAData: caData}
			encryptCluster(existing)

			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)
			clusterStore.EXPECT().UpdateNameAndPlatform(gomock.Any(), "cluster-1", "platform-cluster").Return(nil)

			result, err := svc.InternalUpsertPlatformCluster(ctx, &models.Cluster{
				Name:          "platform-cluster",
				ClusterURL:    "https://example.com:6443",
				Token:         token,
				ClusterCAData: caData,
			})
			Expect(err).To(BeNil())
			Expect(result.Name).To(Equal("platform-cluster"))
			Expect(result.Platform).To(BeTrue())
		})

		It("rotates credentials, re-registers the cluster, and returns decrypted credentials", func() {
			oldToken := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			newToken := base64.StdEncoding.EncodeToString([]byte("token-v2"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Platform: true, Token: oldToken, ClusterCAData: caData}
			encryptCluster(existing)
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Platform: true, Token: newToken, ClusterCAData: caData}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""

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
			Expect(result.Token).To(Equal(newToken))
			Expect(result.ClusterCAData).To(Equal(caData))
		})
	})

	Describe("DefaultStorageClass", func() {
		It("returns the default class from the persisted snapshot", func() {
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(&models.Cluster{
				ID: "cluster-1",
				ClusterInfo: &models.ClusterInfo{StorageClasses: []models.ClusterStorageClass{
					{Name: "local-path", IsDefault: true},
				}},
			}, nil)

			name, err := svc.DefaultStorageClass(ctx, "cluster-1")
			Expect(err).To(BeNil())
			Expect(name).To(Equal("local-path"))
		})

		It("returns empty without an error when the cluster has no info yet", func() {
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(&models.Cluster{ID: "cluster-1"}, nil)

			name, err := svc.DefaultStorageClass(ctx, "cluster-1")
			Expect(err).To(BeNil())
			Expect(name).To(BeEmpty())
		})
	})

	Describe("Platform wildcard TLS", func() {
		const (
			cloudflareToken = "cf-token"
			contactEmail    = "ops@example.com"
			baseDomain      = "apps.example.com"
			tlsNamespace    = "stackdome-control-plane"
		)

		var (
			k8sClient       client.Client
			platformCluster *models.Cluster
			tlsCtx          context.Context
		)

		fullConfig := func() *config.BootstrapConfig {
			return &config.BootstrapConfig{
				BaseDomain:            baseDomain,
				DNSCloudflareAPIToken: cloudflareToken,
				ACMEEnvironment:       config.ACMEEnvironmentStaging,
				TLSNamespace:          tlsNamespace,
			}
		}

		key := func(name, namespace string) client.ObjectKey {
			return client.ObjectKey{Name: name, Namespace: namespace}
		}

		BeforeEach(func() {
			scheme := runtime.NewScheme()
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			Expect(cmv1.AddToScheme(scheme)).To(Succeed())
			k8sClient = fake.NewClientBuilder().WithScheme(scheme).Build()
			platformCluster = &models.Cluster{ID: "cluster-platform", Platform: true}
			tlsCtx = auth.SetIdentityInContext(context.Background(), &auth.Identity{
				IsSystem:     true,
				ContactEmail: contactEmail,
			})
		})

		It("creates the Cloudflare token, DNS issuer, and wildcard certificate", func() {
			clusterManager.EXPECT().GetClient(platformCluster.ID).Return(k8sClient, nil)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, fullConfig())).To(BeNil())

			secret := &corev1.Secret{}
			Expect(k8sClient.Get(tlsCtx, key(models.CloudflareAPITokenSecretName, tlsNamespace), secret)).To(Succeed())
			Expect(secret.Data).To(HaveKeyWithValue(models.CloudflareAPITokenSecretKey, []byte(cloudflareToken)))

			issuer := &cmv1.Issuer{}
			Expect(k8sClient.Get(tlsCtx, key(models.DNSIssuerName, tlsNamespace), issuer)).To(Succeed())
			Expect(issuer.Spec.ACME.Server).To(Equal(config.ACMEStagingDirectoryURL))
			Expect(issuer.Spec.ACME.Email).To(Equal(contactEmail))
			Expect(issuer.Spec.ACME.PrivateKey.Name).To(Equal(models.DNSIssuerPrivateKeySecretName))
			Expect(issuer.Spec.ACME.Solvers).To(HaveLen(1))
			cloudflare := issuer.Spec.ACME.Solvers[0].DNS01.Cloudflare
			Expect(cloudflare.APIToken.Name).To(Equal(models.CloudflareAPITokenSecretName))
			Expect(cloudflare.APIToken.Key).To(Equal(models.CloudflareAPITokenSecretKey))

			certificate := &cmv1.Certificate{}
			Expect(k8sClient.Get(tlsCtx, key(models.PlatformWildcardTLSName, tlsNamespace), certificate)).To(Succeed())
			Expect(certificate.Spec.DNSNames).To(Equal([]string{"*.apps.example.com"}))
			Expect(certificate.Spec.SecretName).To(Equal("platform-wildcard-tls"))
			Expect(certificate.Spec.IssuerRef.Name).To(Equal(models.DNSIssuerName))
			Expect(certificate.Spec.IssuerRef.Kind).To(Equal(cmv1.IssuerKind))
			Expect(certificate.Spec.SecretTemplate).NotTo(BeNil())
			Expect(certificate.Spec.SecretTemplate.Labels).To(HaveKeyWithValue(corev1alpha1.LabelPlatformWildcardTLSSecret, "true"))
		})

		It("creates a missing TLS namespace", func() {
			clusterManager.EXPECT().GetClient(platformCluster.ID).Return(k8sClient, nil)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, fullConfig())).To(BeNil())

			namespace := &corev1.Namespace{}
			Expect(k8sClient.Get(tlsCtx, key(tlsNamespace, ""), namespace)).To(Succeed())
		})

		It("updates the Cloudflare token on a subsequent invocation", func() {
			clusterManager.EXPECT().GetClient(platformCluster.ID).Return(k8sClient, nil).Times(2)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, fullConfig())).To(BeNil())
			rotated := fullConfig()
			rotated.DNSCloudflareAPIToken = "cf-token-v2"
			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, rotated)).To(BeNil())

			secret := &corev1.Secret{}
			Expect(k8sClient.Get(tlsCtx, key(models.CloudflareAPITokenSecretName, tlsNamespace), secret)).To(Succeed())
			Expect(secret.Data).To(HaveKeyWithValue(models.CloudflareAPITokenSecretKey, []byte("cf-token-v2")))
		})

		It("does not look up a cluster client without a base domain", func() {
			cfg := fullConfig()
			cfg.BaseDomain = ""

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, cfg)).To(BeNil())
		})

		It("returns a service error when the cluster client is unavailable", func() {
			clusterManager.EXPECT().GetClient(platformCluster.ID).Return(nil, stderrors.New("unavailable"))

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, platformCluster, fullConfig())).ToNot(BeNil())
		})
	})

	Describe("InternalUpdateClusterInfo", func() {
		It("delegates to the cluster store", func() {
			info := &models.ClusterInfo{StorageClasses: []models.ClusterStorageClass{
				{Name: "local-path", IsDefault: true},
			}}
			clusterStore.EXPECT().UpdateClusterInfo(gomock.Any(), "cluster-1", info).Return(nil)

			err := svc.InternalUpdateClusterInfo(ctx, "cluster-1", info)
			Expect(err).To(BeNil())
		})
	})
})
