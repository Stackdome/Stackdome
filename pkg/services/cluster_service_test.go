package services

import (
	"context"
	"encoding/base64"
	stderrors "errors"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/db"
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
			computeMode:          config.ComputeModeBYOC,
		}

		ctx = auth.SetIdentityInContext(context.Background(), &auth.Identity{IsSystem: true})
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	encryptCluster := func(cluster *models.Cluster) {
		Expect(svc.encryptClusterCredentials(cluster)).To(BeNil())
	}
	runTransactionWithOutcome := func(outcome *apperrors.ServiceError) {
		clusterStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(parent context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
				txCtx, hooks := db.CtxWithPostCommitHooks(parent)
				err := fn(txCtx)
				if err != nil {
					Expect(hooks.RunRollback(db.ContextAfterTransaction(parent))).To(BeEmpty())
					return err
				}
				if outcome != nil {
					Expect(hooks.RunRollback(db.ContextAfterTransaction(parent))).To(BeEmpty())
					return outcome
				}
				hooks.Run()
				return nil
			})
	}
	runTransaction := func() {
		runTransactionWithOutcome(nil)
	}
	expectTopologyLock := func(orgID string) {
		orgStore.EXPECT().LockByID(gomock.Any(), orgID).Return(nil)
	}

	Describe("GetClusterForOrg mode-directed resolution", func() {
		It("returns exactly one BYOC cluster without consulting shared compute", func() {
			owned := &models.Cluster{ID: "cluster-owned", OrganisationID: "org-1", Token: "tok", ClusterCAData: caData}
			encryptCluster(owned)

			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), "org-1").Return([]*models.Cluster{owned}, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-owned"))
		})

		It("returns exactly one shared-compute cluster without consulting BYOC", func() {
			svc.computeMode = config.ComputeModeShared
			def := &models.Cluster{ID: "cluster-shared", OrganisationID: "org-platform", SharedCompute: true, Token: "tok", ClusterCAData: caData}
			encryptCluster(def)

			clusterStore.EXPECT().ListSharedComputeClusters(gomock.Any()).Return([]*models.Cluster{def}, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-shared"))
		})

		It("fails closed when multiple BYOC clusters exist", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), "org-1").Return([]*models.Cluster{{}, {}}, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: multiple BYOC clusters found for organisation 'org-1'"))
		})

		It("fails closed when multiple shared-compute clusters exist", func() {
			svc.computeMode = config.ComputeModeShared
			clusterStore.EXPECT().ListSharedComputeClusters(gomock.Any()).Return([]*models.Cluster{{}, {}}, nil)

			result, err := svc.GetClusterForOrg(ctx, "org-1")
			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: multiple shared-compute clusters configured"))
		})
	})

	Describe("GetOwnedClusterForOrg", func() {
		It("returns the org's own cluster when it owns one", func() {
			owned := &models.Cluster{ID: "cluster-owned", OrganisationID: "org-1", Token: "tok", ClusterCAData: caData}
			encryptCluster(owned)

			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), "org-1").Return([]*models.Cluster{owned}, nil)

			result, err := svc.GetOwnedClusterForOrg(ctx, "org-1")
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-owned"))
		})

		It("returns NotFound without falling back to shared compute when the org owns none", func() {
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), "org-1").Return(nil, nil)

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
				ClusterStore: clusterStore,
				ComputeMode:  config.ComputeModeShared,
				Permissions:  permissions,
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
				ClusterStore: clusterStore,
				ComputeMode:  config.ComputeModeShared,
				Permissions:  permissions,
			}).(*clusterService)

			result, err := sharedService.AddCluster(ctx, newClusterSpec())

			Expect(result).To(BeNil())
			Expect(err).To(BeIdenticalTo(permissionError))
		})

		It("rejects a second BYOC cluster as a service policy", func() {
			runTransaction()
			expectTopologyLock(tenantOrgID)
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), tenantOrgID).
				Return([]*models.Cluster{{ID: "existing-cluster"}}, nil)

			result, err := svc.AddCluster(ctx, newClusterSpec())

			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: cluster already exists for org"))
			Expect(err.Code).To(Equal(apperrors.ErrorConflict))
		})

		It("propagates an organisation-lock failure before reading topology", func() {
			lockErr := apperrors.GeneralError("failed to lock organisation")
			runTransaction()
			orgStore.EXPECT().LockByID(gomock.Any(), tenantOrgID).Return(lockErr)

			result, err := svc.AddCluster(ctx, newClusterSpec())

			Expect(result).To(BeNil())
			Expect(err).To(BeIdenticalTo(lockErr))
		})

		expectClusterCreated := func(created *models.Cluster) {
			runTransaction()
			expectTopologyLock(tenantOrgID)
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), tenantOrgID).Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NotFound("not found")).AnyTimes()
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			clusterManager.EXPECT().GetClient(created.ID).Return(nil, stderrors.New("no client in test"))
		}

		It("auto-creates a default registry named <slug>-<shortOrgID>-<shortClusterID> when none is supplied", func() {
			orgID := "11112222-3333-4444-5555-666677778888"
			spec := newClusterSpec()
			spec.OrganisationID = orgID
			created := &models.Cluster{ID: "cluster-owned", OrganisationID: orgID}

			runTransaction()
			expectTopologyLock(orgID)
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), orgID).Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NotFound("not found")).AnyTimes()
			orgStore.EXPECT().Get(gomock.Any(), orgID).
				Return(&models.Organisation{ID: orgID, Name: "Acme Inc"}, nil)
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

		It("continues issuer creation for BYOC when platform TLS is disabled", func() {
			svc.platformTLSEnabled = false
			spec := newClusterSpec()
			created := &models.Cluster{ID: "cluster-for-platform-org", OrganisationID: tenantOrgID}

			expectClusterCreated(created)
			orgStore.EXPECT().Get(gomock.Any(), tenantOrgID).
				Return(&models.Organisation{ID: tenantOrgID, Name: "platform", Platform: true}, nil)

			result, err := svc.AddCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ImageRegistries).To(BeEmpty())
		})

		It("unregisters a cluster when registry persistence rolls back", func() {
			spec := newClusterSpec()
			created := &models.Cluster{ID: "cluster-rolled-back", OrganisationID: tenantOrgID}
			registryErr := apperrors.GeneralError("registry persistence failed")

			runTransaction()
			expectTopologyLock(tenantOrgID)
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), tenantOrgID).Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), spec.ClusterURL).
				Return(nil, apperrors.NotFound("not found"))
			orgStore.EXPECT().Get(gomock.Any(), tenantOrgID).
				Return(&models.Organisation{ID: tenantOrgID, Name: "Tenant Org"}, nil)
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			registrySvc.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(nil, registryErr)
			clusterManager.EXPECT().UnregisterCluster(created.ID).Return(nil)

			result, err := svc.AddCluster(ctx, spec)

			Expect(result).To(BeNil())
			Expect(err).To(BeIdenticalTo(registryErr))
		})

		It("fails before registry creation when manager registration fails", func() {
			spec := newClusterSpec()
			created := &models.Cluster{ID: "cluster-registration-failed", OrganisationID: tenantOrgID}
			registrationErr := stderrors.New("manager registration failed")

			runTransaction()
			expectTopologyLock(tenantOrgID)
			clusterStore.EXPECT().ListBYOCClustersForOrg(gomock.Any(), tenantOrgID).Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), spec.ClusterURL).
				Return(nil, apperrors.NotFound("not found"))
			orgStore.EXPECT().Get(gomock.Any(), tenantOrgID).
				Return(&models.Organisation{ID: tenantOrgID, Name: "Tenant Org"}, nil)
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(registrationErr)
			clusterManager.EXPECT().UnregisterCluster(created.ID).Return(nil)

			result, serr := svc.AddCluster(ctx, spec)

			Expect(result).To(BeNil())
			Expect(serr).To(MatchError("error: failed to register cluster with manager: manager registration failed"))
		})
	})

	Describe("Delete", func() {
		It("persists cluster and registry deletion intent before enqueueing after commit", func() {
			cluster := &models.Cluster{ID: "cluster-delete", OrganisationID: "org-delete"}
			enqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)
			svc.BackgroundJobEnqueuer = enqueuer

			clusterStore.EXPECT().Get(ctx, cluster.ID).Return(cluster, nil)
			runTransaction()
			orgStore.EXPECT().LockByID(gomock.Any(), cluster.OrganisationID).Return(nil)
			clusterStore.EXPECT().Get(gomock.Any(), cluster.ID).Return(cluster, nil)
			clusterStore.EXPECT().MarkDeletingWithTx(gomock.Any(), cluster.ID, gomock.Any()).Return(nil)
			registrySvc.EXPECT().InternalMarkAllDeletingByClusterIDWithTx(gomock.Any(), cluster.ID).Return(nil)
			enqueuer.EXPECT().EnqueueAfterCommit(gomock.Any(), models.ClusterImageRegistryOperand{ClusterID: cluster.ID}).Return(nil)

			Expect(svc.Delete(ctx, cluster.ID)).To(BeNil())
		})
	})

	Describe("InternalUpsertSharedComputeCluster", func() {
		BeforeEach(func() {
			svc.computeMode = config.ComputeModeShared
		})

		It("creates the shared-compute cluster through the trusted internal path", func() {
			svc.platformTLSEnabled = true
			spec := &models.Cluster{
				Name:           "default",
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          "tok",
				ClusterCAData:  caData,
			}
			created := &models.Cluster{ID: "cluster-new", OrganisationID: "org-platform", SharedCompute: true}

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), spec.ClusterURL).
				Return(nil, apperrors.NotFound("cluster with this api URL not found")).AnyTimes()
			orgStore.EXPECT().Get(gomock.Any(), "org-platform").
				Return(&models.Organisation{ID: "org-platform", Name: "platform", Platform: true}, nil)
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)
			clusterManager.EXPECT().GetClient(created.ID).Return(nil, stderrors.New("no client in test"))

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-new"))
			Expect(spec.SharedCompute).To(BeTrue())
		})

		It("skips issuer creation for shared compute when platform TLS is disabled", func() {
			spec := &models.Cluster{
				Name:           "default",
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          "tok",
				ClusterCAData:  caData,
			}
			created := &models.Cluster{ID: "cluster-new", OrganisationID: "org-platform", SharedCompute: true}

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), spec.ClusterURL).
				Return(nil, apperrors.NotFound("cluster with this api URL not found")).AnyTimes()
			orgStore.EXPECT().Get(gomock.Any(), "org-platform").
				Return(&models.Organisation{ID: "org-platform", Name: "platform", Platform: true}, nil)
			clusterStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).Return(created, nil)
			clusterManager.EXPECT().RegisterCluster(created).Return(nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, spec)
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-new"))
		})

		It("rejects shared-compute topology in bring-your-own mode before accessing the store", func() {
			svc.computeMode = config.ComputeModeBYOC

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: shared-compute clusters require shared compute mode"))
			Expect(err.Code).To(Equal(apperrors.ErrorBadRequest))
		})

		It("updates the single organisation-owned shared cluster when its URL changes", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{
				ID:             "cluster-1",
				OrganisationID: "org-platform",
				Name:           models.SharedComputeClusterName,
				ClusterURL:     "https://old.example.com:6443",
				SharedCompute:  true,
				Token:          token,
				ClusterCAData:  caData,
			}
			encryptCluster(existing)
			fresh := &models.Cluster{
				ID:             existing.ID,
				OrganisationID: existing.OrganisationID,
				Name:           existing.Name,
				ClusterURL:     "https://new.example.com:6443",
				SharedCompute:  true,
				Token:          token,
				ClusterCAData:  caData,
			}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""

			runTransaction()
			expectTopologyLock(existing.OrganisationID)
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), existing.OrganisationID).
				Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), fresh.ClusterURL).
				Return(nil, apperrors.NotFound("not found"))
			clusterStore.EXPECT().UpdateSharedComputeCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Cluster) *apperrors.ServiceError {
					Expect(updated.ID).To(Equal(existing.ID))
					Expect(updated.ClusterURL).To(Equal(fresh.ClusterURL))
					return nil
				})
			clusterStore.EXPECT().Get(gomock.Any(), existing.ID).Return(fresh, nil)
			clusterManager.EXPECT().ReRegisterCluster(fresh).Return(nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: existing.OrganisationID,
				ClusterURL:     fresh.ClusterURL,
				Token:          token,
				ClusterCAData:  caData,
			})

			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal(existing.ID))
			Expect(result.ClusterURL).To(Equal(fresh.ClusterURL))
		})

		It("never adopts a URL owned by another organisation", func() {
			foreign := &models.Cluster{ID: "cluster-other", OrganisationID: "org-other", ClusterURL: "https://example.com:6443"}
			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), foreign.ClusterURL).Return(foreign, nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     foreign.ClusterURL,
				Token:          "token",
				ClusterCAData:  caData,
			})

			Expect(result).To(BeNil())
			Expect(err.IsConflict()).To(BeTrue())
		})

		It("fails closed when the organisation already owns multiple shared clusters", func() {
			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").
				Return([]*models.Cluster{{ID: "cluster-1"}, {ID: "cluster-2"}}, nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				OrganisationID: "org-platform",
			})

			Expect(result).To(BeNil())
			Expect(err).To(MatchError("error: multiple shared-compute clusters found for organisation 'org-platform'"))
		})

		It("is a no-op when the stored credentials already match", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: token, ClusterCAData: caData}
			encryptCluster(existing)

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          token,
				ClusterCAData:  caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
		})

		It("normalizes raw credentials before comparing, so a raw env token is not a rotation", func() {
			raw := "token-v1"
			stored := base64.StdEncoding.EncodeToString([]byte(raw))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: stored, ClusterCAData: caData}
			encryptCluster(existing)

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          raw,
				ClusterCAData:  caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
		})

		It("reconciles drifted name and shared-compute flag without rotating credentials", func() {
			token := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", ClusterURL: "https://example.com:6443", Name: "old-name", SharedCompute: false, Token: token, ClusterCAData: caData}
			encryptCluster(existing)

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return(nil, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil).Times(2)
			clusterStore.EXPECT().UpdateSharedComputeCluster(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, updated *models.Cluster) *apperrors.ServiceError {
					Expect(updated.ID).To(Equal("cluster-1"))
					Expect(updated.Name).To(Equal(models.SharedComputeClusterName))
					Expect(updated.SharedCompute).To(BeTrue())
					return nil
				})
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: token, ClusterCAData: caData}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(fresh, nil)

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          token,
				ClusterCAData:  caData,
			})
			Expect(err).To(BeNil())
			Expect(result.Name).To(Equal(models.SharedComputeClusterName))
			Expect(result.SharedCompute).To(BeTrue())
		})

		It("rotates credentials, re-registers the cluster, and returns decrypted credentials", func() {
			oldToken := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			newToken := base64.StdEncoding.EncodeToString([]byte("token-v2"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: oldToken, ClusterCAData: caData}
			encryptCluster(existing)
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: newToken, ClusterCAData: caData}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""

			transactionOpen := false
			clusterStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
				DoAndReturn(func(parent context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
					transactionOpen = true
					txCtx, hooks := db.CtxWithPostCommitHooks(parent)
					err := fn(txCtx)
					transactionOpen = false
					if err == nil {
						hooks.Run()
					}
					return err
				})
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), "https://example.com:6443").Return(existing, nil)
			clusterStore.EXPECT().UpdateSharedComputeCluster(gomock.Any(), gomock.Any()).Return(nil)
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(fresh, nil)
			clusterManager.EXPECT().ReRegisterCluster(fresh).
				DoAndReturn(func(_ *models.Cluster) error {
					Expect(transactionOpen).To(BeTrue())
					return nil
				})

			result, err := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     "https://example.com:6443",
				Token:          newToken,
				ClusterCAData:  caData,
			})
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal("cluster-1"))
			Expect(result.Token).To(Equal(newToken))
			Expect(result.ClusterCAData).To(Equal(caData))
		})

		It("restores the prior manager configuration after commit failure", func() {
			oldToken := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			newToken := base64.StdEncoding.EncodeToString([]byte("token-v2"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://old.example.com:6443", SharedCompute: true, Token: oldToken, ClusterCAData: caData}
			encryptCluster(existing)
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://new.example.com:6443", SharedCompute: true, Token: newToken, ClusterCAData: caData}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""
			commitErr := apperrors.GeneralError("commit failed")

			runTransactionWithOutcome(commitErr)
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), fresh.ClusterURL).Return(nil, apperrors.NotFound("not found"))
			clusterStore.EXPECT().UpdateSharedComputeCluster(gomock.Any(), gomock.Any()).Return(nil)
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(fresh, nil)
			clusterManager.EXPECT().ReRegisterCluster(fresh).Return(nil)
			clusterManager.EXPECT().ReRegisterCluster(gomock.Any()).
				DoAndReturn(func(previous *models.Cluster) error {
					Expect(previous.ID).To(Equal(existing.ID))
					Expect(previous.ClusterURL).To(Equal("https://old.example.com:6443"))
					Expect(previous.Token).To(Equal(oldToken))
					Expect(previous.ClusterCAData).To(Equal(caData))
					return nil
				})

			result, serr := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     fresh.ClusterURL,
				Token:          newToken,
				ClusterCAData:  caData,
			})

			Expect(result).To(BeNil())
			Expect(serr).To(BeIdenticalTo(commitErr))
		})

		It("fails the update and restores the prior manager configuration when re-registration fails", func() {
			oldToken := base64.StdEncoding.EncodeToString([]byte("token-v1"))
			newToken := base64.StdEncoding.EncodeToString([]byte("token-v2"))
			existing := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: oldToken, ClusterCAData: caData}
			encryptCluster(existing)
			fresh := &models.Cluster{ID: "cluster-1", OrganisationID: "org-platform", Name: models.SharedComputeClusterName, ClusterURL: "https://example.com:6443", SharedCompute: true, Token: newToken, ClusterCAData: caData}
			encryptCluster(fresh)
			fresh.Token, fresh.ClusterCAData = "", ""

			runTransaction()
			expectTopologyLock("org-platform")
			clusterStore.EXPECT().ListSharedComputeClustersForOrg(gomock.Any(), "org-platform").Return([]*models.Cluster{existing}, nil)
			clusterStore.EXPECT().GetByClusterUrl(gomock.Any(), existing.ClusterURL).Return(existing, nil)
			clusterStore.EXPECT().UpdateSharedComputeCluster(gomock.Any(), gomock.Any()).Return(nil)
			clusterStore.EXPECT().Get(gomock.Any(), "cluster-1").Return(fresh, nil)
			clusterManager.EXPECT().ReRegisterCluster(fresh).Return(stderrors.New("new manager configuration rejected"))
			clusterManager.EXPECT().ReRegisterCluster(gomock.Any()).
				DoAndReturn(func(previous *models.Cluster) error {
					Expect(previous.Token).To(Equal(oldToken))
					Expect(previous.ClusterCAData).To(Equal(caData))
					return nil
				})

			result, serr := svc.InternalUpsertSharedComputeCluster(ctx, &models.Cluster{
				Name:           models.SharedComputeClusterName,
				OrganisationID: "org-platform",
				ClusterURL:     existing.ClusterURL,
				Token:          newToken,
				ClusterCAData:  caData,
			})

			Expect(result).To(BeNil())
			Expect(serr).To(MatchError("error: failed to re-register cluster with manager: new manager configuration rejected"))
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
			k8sClient            client.Client
			sharedComputeCluster *models.Cluster
			tlsCtx               context.Context
		)

		fullConfig := func() *config.PlatformConfig {
			return &config.PlatformConfig{
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
			sharedComputeCluster = &models.Cluster{ID: "cluster-shared", SharedCompute: true}
			tlsCtx = auth.SetIdentityInContext(context.Background(), &auth.Identity{
				IsSystem:     true,
				ContactEmail: contactEmail,
			})
		})

		It("creates the Cloudflare token, DNS issuer, and wildcard certificate", func() {
			clusterManager.EXPECT().GetClient(sharedComputeCluster.ID).Return(k8sClient, nil)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, fullConfig())).To(BeNil())

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
			clusterManager.EXPECT().GetClient(sharedComputeCluster.ID).Return(k8sClient, nil)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, fullConfig())).To(BeNil())

			namespace := &corev1.Namespace{}
			Expect(k8sClient.Get(tlsCtx, key(tlsNamespace, ""), namespace)).To(Succeed())
		})

		It("updates the Cloudflare token on a subsequent invocation", func() {
			clusterManager.EXPECT().GetClient(sharedComputeCluster.ID).Return(k8sClient, nil).Times(2)

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, fullConfig())).To(BeNil())
			rotated := fullConfig()
			rotated.DNSCloudflareAPIToken = "cf-token-v2"
			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, rotated)).To(BeNil())

			secret := &corev1.Secret{}
			Expect(k8sClient.Get(tlsCtx, key(models.CloudflareAPITokenSecretName, tlsNamespace), secret)).To(Succeed())
			Expect(secret.Data).To(HaveKeyWithValue(models.CloudflareAPITokenSecretKey, []byte("cf-token-v2")))
		})

		It("does not look up a cluster client without a base domain", func() {
			cfg := fullConfig()
			cfg.BaseDomain = ""

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, cfg)).To(BeNil())
		})

		It("returns a service error when the cluster client is unavailable", func() {
			clusterManager.EXPECT().GetClient(sharedComputeCluster.ID).Return(nil, stderrors.New("unavailable"))

			Expect(svc.InternalEnsurePlatformWildcardTLS(tlsCtx, sharedComputeCluster, fullConfig())).ToNot(BeNil())
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
