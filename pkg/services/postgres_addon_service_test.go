package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator/postgresaddon"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

// Suite bootstrapped by TestAESEncryptionService in encryption_service_test.go.

var _ = Describe("PostgresAddonService cloud validation", func() {
	It("rejects an external import before provisioning dependencies are used", func() {
		ctrl := gomock.NewController(GinkgoT())
		permissions := mocks.NewMockPermissionService(ctrl)
		permissions.EXPECT().Check(
			gomock.Any(), "project-1", auth.ResourceAddonsPostgres, "", auth.ActionCreate,
		).Return(nil)
		service := NewPostgresAddonService(PostgresAddonServiceSpec{
			Permissions:            permissions,
			ExternalImportDisabled: true,
			RuntimePolicy:          NewSelfHostedRuntimePolicy(),
		})
		spec := &models.PostgresAddon{
			Name:            addonType,
			OrganisationID:  "organisation-1",
			ProjectID:       "project-1",
			UserID:          "user-1",
			PostgresVersion: models.PostgresVersion{Major: 16},
			Instances:       models.PostgresInstances{Count: 1},
			Storage:         models.PostgresStorage{Size: "10Gi"},
			Initialization: models.PostgresInitialization{
				ImportFromExternal: &models.PostgresImportFromExternal{
					Host: "database.example.com", Port: 5432, Database: "app", Username: "app", PasswordSecretID: "secret-1",
				},
			},
		}

		_, err := service.CreatePostgresAddon(context.Background(), spec)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(apperrors.ErrorValidation))
		details, ok := err.Details.(apperrors.ValidationErrorDetails)
		Expect(ok).To(BeTrue())
		Expect(details.Errors).To(ConsistOf(apperrors.FieldError{
			Field:   postgresaddon.ExternalImportDisabledField,
			Code:    apperrors.VErrPostgresExternalImportDisabled,
			Message: postgresaddon.ExternalImportDisabledMessage,
		}))
	})
})

var _ = Describe("PostgresAddonService DeletePostgresAddon", func() {
	var (
		ctrl        *gomock.Controller
		addonStore  *mocks.MockPostgresAddonStore
		permissions *mocks.MockPermissionService
		refs        *mocks.MockReferenceService
		enqueuer    *mocks.MockBackgroundJobEnqueuer
		policy      *MockRuntimePolicy
		svc         *postgresAddonService
		ctx         context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonStore = mocks.NewMockPostgresAddonStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		refs = mocks.NewMockReferenceService(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		policy = NewMockRuntimePolicy(ctrl)
		ctx = context.Background()

		svc = &postgresAddonService{
			logger:             logger.NewLogger(),
			postgresAddonStore: addonStore,
			referenceService:   refs,
			permissions:        permissions,
			runtimePolicy:      policy,
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
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "pg-1"}).Return(nil)

		deleted, err := svc.DeletePostgresAddon(ctx, "pg-1")
		Expect(err).To(BeNil())
		Expect(deleted.DeletionTimestamp).NotTo(BeNil())
	})
})

var _ = Describe("CreatePostgresAddon storage class defaulting", func() {
	const (
		orgID         = "org-1"
		projectID     = "project-1"
		userID        = "user-1"
		clusterID     = "cluster-1"
		addonID       = "addon-1"
		addonName     = "test-addon"
		namespaceID   = "namespace-1"
		namespaceName = "ns-test-addon"
	)

	var (
		ctrl            *gomock.Controller
		addonStore      *mocks.MockPostgresAddonStore
		permissions     *mocks.MockPermissionService
		namespaceSvc    *mocks.MockNamespaceService
		clusterService  *mocks.MockClusterService
		databaseService *mocks.MockPostgresAddonDatabaseService
		enqueuer        *mocks.MockBackgroundJobEnqueuer
		runtimePolicy   *MockRuntimePolicy
		svc             *postgresAddonService
		ctx             context.Context
		capturedAddon   *models.PostgresAddon
	)

	newAddonFixture := func() *models.PostgresAddon {
		return &models.PostgresAddon{
			Name:            addonName,
			OrganisationID:  orgID,
			ProjectID:       projectID,
			UserID:          userID,
			PostgresVersion: models.PostgresVersion{Major: 16},
			Instances:       models.PostgresInstances{Count: 1},
			Storage:         models.PostgresStorage{Size: "10Gi"},
		}
	}

	// post-resolution persistence mocks
	expectPersistence := func(wantStorageClass string) {
		addonStore.EXPECT().GetByName(gomock.Any(), orgID, addonName).Return(nil, nil)
		addonStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
				return fn(ctx)
			})
		runtimePolicy.EXPECT().ActivateComputeAccessWithTx(gomock.Any(), orgID).Return(nil)
		namespaceSvc.EXPECT().CreateInDBWithTx(gomock.Any(), gomock.Any()).
			Return(&models.Namespace{ID: namespaceID, Name: namespaceName}, nil)
		addonStore.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *apperrors.ServiceError) {
				Expect(addon.Storage.StorageClass).To(Equal(wantStorageClass))
				addon.ID = addonID
				capturedAddon = addon
				return addon, nil
			})
		databaseService.EXPECT().CreateWithTx(gomock.Any(), gomock.Any()).
			Return(&models.PostgresAddonDatabase{}, nil)
		addonStore.EXPECT().GetByID(gomock.Any(), addonID).
			DoAndReturn(func(_ context.Context, _ string) (*models.PostgresAddon, *apperrors.ServiceError) {
				return capturedAddon, nil
			})
		permissions.EXPECT().Check(gomock.Any(), projectID, auth.ResourceAddonsPostgres, addonID, auth.ActionRead).Return(nil)
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: addonID}).Return(nil)
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonStore = mocks.NewMockPostgresAddonStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		namespaceSvc = mocks.NewMockNamespaceService(ctrl)
		clusterService = mocks.NewMockClusterService(ctrl)
		databaseService = mocks.NewMockPostgresAddonDatabaseService(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		runtimePolicy = NewMockRuntimePolicy(ctrl)
		ctx = context.Background()
		capturedAddon = nil

		svc = &postgresAddonService{
			logger:             logger.NewLogger(),
			postgresAddonStore: addonStore,
			permissions:        permissions,
			runtimePolicy:      runtimePolicy,
			namespaceService:   namespaceSvc,
			clusterService:     clusterService,
			databaseService:    databaseService,
			validator:          postgresaddon.NewPostgresAddonValidator(postgresaddon.PostgresAddonValidatorSpec{}),
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}

		permissions.EXPECT().Check(gomock.Any(), projectID, auth.ResourceAddonsPostgres, "", auth.ActionCreate).Return(nil)
		namespaceSvc.EXPECT().PrepareNamespaceForAddon(gomock.Any(), gomock.Any(), orgID).
			Return(&models.Namespace{ID: namespaceID, Name: namespaceName}, nil)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("stamps the cluster default when the request omits storage_class", func() {
		clusterService.EXPECT().GetClusterForOrg(gomock.Any(), orgID).Return(&models.Cluster{ID: clusterID}, nil)
		clusterService.EXPECT().DefaultStorageClass(gomock.Any(), clusterID).Return("local-path", nil)
		expectPersistence("local-path")

		addon := newAddonFixture()
		addon.Storage.StorageClass = ""

		created, err := svc.CreatePostgresAddon(ctx, addon)
		Expect(err).To(BeNil())
		Expect(created.Storage.StorageClass).To(Equal("local-path"))
	})

	It("rejects the create when the cluster reports no default class", func() {
		clusterService.EXPECT().GetClusterForOrg(gomock.Any(), orgID).Return(&models.Cluster{ID: clusterID}, nil)
		clusterService.EXPECT().DefaultStorageClass(gomock.Any(), clusterID).Return("", nil)

		addon := newAddonFixture()
		addon.Storage.StorageClass = ""

		_, err := svc.CreatePostgresAddon(ctx, addon)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("no default storage class"))
	})

	It("keeps an explicit storage_class and does not consult the cluster", func() {
		clusterService.EXPECT().GetClusterForOrg(gomock.Any(), orgID).Return(&models.Cluster{ID: clusterID}, nil)
		clusterService.EXPECT().DefaultStorageClass(gomock.Any(), gomock.Any()).Times(0)
		expectPersistence("io2")

		addon := newAddonFixture()
		addon.Storage.StorageClass = "io2"

		created, err := svc.CreatePostgresAddon(ctx, addon)
		Expect(err).To(BeNil())
		Expect(created.Storage.StorageClass).To(Equal("io2"))
	})
})

var _ = Describe("UpdatePostgresAddon storage class carry-forward", func() {
	const (
		orgID                = "org-1"
		projectID            = "project-1"
		userID               = "user-1"
		addonID              = "addon-1"
		addonName            = "test-addon"
		existingStorageClass = "existing-class"
	)

	var (
		ctrl        *gomock.Controller
		addonStore  *mocks.MockPostgresAddonStore
		permissions *mocks.MockPermissionService
		enqueuer    *mocks.MockBackgroundJobEnqueuer
		svc         *postgresAddonService
		ctx         context.Context
		existing    *models.PostgresAddon
	)

	newUpdateFixture := func() *models.PostgresAddon {
		return &models.PostgresAddon{
			ID:              addonID,
			Name:            addonName,
			OrganisationID:  orgID,
			ProjectID:       projectID,
			UserID:          userID,
			PostgresVersion: models.PostgresVersion{Major: 16},
			Instances:       models.PostgresInstances{Count: 1},
			Storage:         models.PostgresStorage{Size: "10Gi"},
		}
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonStore = mocks.NewMockPostgresAddonStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		ctx = context.Background()

		svc = &postgresAddonService{
			logger:             logger.NewLogger(),
			postgresAddonStore: addonStore,
			permissions:        permissions,
			runtimePolicy:      NewSelfHostedRuntimePolicy(),
			validator:          postgresaddon.NewPostgresAddonValidator(postgresaddon.PostgresAddonValidatorSpec{}),
			BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{
				BackgroundJobEnqueuer: enqueuer,
			},
		}

		existing = &models.PostgresAddon{
			ID:              addonID,
			Name:            addonName,
			OrganisationID:  orgID,
			ProjectID:       projectID,
			PostgresVersion: models.PostgresVersion{Major: 16},
			Storage:         models.PostgresStorage{Size: "10Gi", StorageClass: existingStorageClass},
		}

		addonStore.EXPECT().GetByID(gomock.Any(), addonID).Return(existing, nil)
		permissions.EXPECT().Check(gomock.Any(), projectID, auth.ResourceAddonsPostgres, addonID, auth.ActionWrite).Return(nil)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("preserves the existing storage class when the update omits it", func() {
		addonStore.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *apperrors.ServiceError) *apperrors.ServiceError {
				return fn(ctx)
			})
		addonStore.EXPECT().UpdateWithTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, addon *models.PostgresAddon) (*models.PostgresAddon, *apperrors.ServiceError) {
				Expect(addon.Storage.StorageClass).To(Equal(existingStorageClass))
				return addon, nil
			})
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: addonID}).Return(nil)

		update := newUpdateFixture()
		update.Storage.StorageClass = ""

		updated, err := svc.UpdatePostgresAddon(ctx, addonID, update)
		Expect(err).To(BeNil())
		Expect(updated.Storage.StorageClass).To(Equal(existingStorageClass))
	})

	It("rejects an explicit different storage class", func() {
		update := newUpdateFixture()
		update.Storage.StorageClass = "different-class"

		_, err := svc.UpdatePostgresAddon(ctx, addonID, update)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring("Storage class cannot be changed"))
	})
})
