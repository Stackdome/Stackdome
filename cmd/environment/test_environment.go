package environment

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/clusterimageregistry"
	imagebuildcontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/imagebuild"
	postgresaddoncontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/postgres_addon"
	postgresbackupcontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/postgres_backup"
	stackcontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/stack"
	stackresourcecontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/stackresource"
	volumecontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/volume"
	workspaceusercontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspaceuser"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	applogger "github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	postgresaddonworker "github.com/ashishmax31/stackdome-api-server/pkg/worker/postgresaddon"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/stack"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	"github.com/sirupsen/logrus"
)

type testEnvironment struct {
	*Env
}

type EnvConfigOption interface {
	ApplyToEnv(env *Env)
}

type ApplicationConfigOption func(*config.ApplicationConfig)

type BootstrapConfigOption func(*config.BootstrapConfig)

func (o ApplicationConfigOption) ApplyToEnv(env *Env) {
	o(env.Config)
}

func (o BootstrapConfigOption) ApplyToEnv(env *Env) {
	o(env.BootstrapConfig)
}

func WithApplicationConfig(cfg *config.ApplicationConfig) EnvConfigOption {
	return ApplicationConfigOption(func(env *config.ApplicationConfig) {
		*env = *cfg
	})
}

func WithBootstrapConfig(cfg *config.BootstrapConfig) EnvConfigOption {
	return BootstrapConfigOption(func(env *config.BootstrapConfig) {
		*env = *cfg
	})
}

func NewTestEnvironment(sessionFactory db.SessionFactory, dbConfig *config.DatabaseConfig, opts ...EnvConfigOption) EnvImpl {
	res := &testEnvironment{
		Env: &Env{
			Name:            "test",
			DBSession:       sessionFactory,
			Config:          config.NewApplicationConfig(),
			BootstrapConfig: config.NewBootstrapConfig(),
		},
	}

	for _, opt := range opts {
		opt.ApplyToEnv(res.Env)
	}
	return res
}

func (te *testEnvironment) Environment() *Env {
	return te.Env
}

func (te *testEnvironment) Init(ctx context.Context) error {
	initializerSteps := []func(context.Context) error{
		te.loadEnvAndConfigs,
		te.setupLogger,
		te.initializeResourceAccessPolicyManager,
		te.loadServices,
		te.initializeClusterManager,
		te.initializeWorkerManager,
		te.injectClusterResourceServices,
		te.initializeBaseResourceAccessPolicies,
		te.ensureDefaultPlatformAdminUser,
		te.startManagers,
	}

	for _, step := range initializerSteps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("failed to initialize test environment: %w", err)
		}
	}
	te.Logger.Infof("Test environment initialized successfully")
	return nil
}

func (te *testEnvironment) InitDatabase(ctx context.Context) error {
	// Database is already initialized in test setup
	return nil
}

func (te *testEnvironment) loadEnvAndConfigs(ctx context.Context) error {
	// Load .env file (optional for tests)
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist in test environment - we'll log this later
	}

	// Load environment variables with test-specific fallbacks
	te.loadTestEnvVariables()

	if err := te.Config.Validate(); err != nil {
		return fmt.Errorf("invalid application config: %w", err)
	}

	if err := te.BootstrapConfig.Validate(); err != nil {
		return fmt.Errorf("invalid bootstrap config: %w", err)
	}
	return nil
}

func (te *testEnvironment) loadTestEnvVariables() {
	// Load standard environment variables
	te.Config.LoadEnvVariables()
	te.BootstrapConfig.LoadEnvVariables()

	// Override with test-specific defaults if not set
	if te.Config.JwtSecret == "" {
		if testSecret := os.Getenv("TEST_JWT_SECRET"); testSecret != "" {
			te.Config.JwtSecret = testSecret
		} else {
			te.Config.JwtSecret = "test-jwt-secret-key-must-be-longer-than-this-for-security-requirements"
		}
	}

	if te.Config.EncryptionKey == "" {
		if testKey := os.Getenv("TEST_ENCRYPTION_KEY"); testKey != "" {
			te.Config.EncryptionKey = testKey
		} else {
			te.Config.EncryptionKey = "test-encryption-key-must-be-at-least-64-characters-long-for-security-requirements"
		}
	}

	if te.Config.LogLevel == "" {
		if testLevel := os.Getenv("TEST_LOG_LEVEL"); testLevel != "" {
			te.Config.LogLevel = testLevel
		} else {
			te.Config.LogLevel = "info"
		}
	}

	// Set test default user if not configured
	if te.BootstrapConfig.DefaultUser.Email == "" {
		te.BootstrapConfig.DefaultUser.Email = "test-admin@stackdome.io"
	}
	if te.BootstrapConfig.DefaultUser.Name == "" {
		te.BootstrapConfig.DefaultUser.Name = "Test Platform Admin"
	}
	if te.BootstrapConfig.DefaultUser.Password == "" {
		te.BootstrapConfig.DefaultUser.Password = "test-welcome@123"
	}
}

func (te *testEnvironment) setupLogger(ctx context.Context) error {
	logLevel, err := logrus.ParseLevel(te.Config.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", te.Config.LogLevel, err)
	}
	te.Logger = applogger.NewLoggerWithPrefix(ctx, "test-api-server").SetLevel(logLevel)

	// Configure logrus to output to stdout for test visibility
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	te.Logger.Debugf("Test logger initialized with level: %s", logLevel.String())
	return nil
}

func (te *testEnvironment) initializeResourceAccessPolicyManager(ctx context.Context) error {
	te.Logger.Debugf("Initializing resource access policy manager for tests")
	debugModeEnabled := te.Logger.GetLevel() == logrus.DebugLevel

	rootdir, err := findGoModDir()
	if err != nil {
		return fmt.Errorf("failed to find root directory for policy file: %w", err)
	}
	policyFilePath := filepath.Join(rootdir, "pkg/resourceaccess/casbin_model.conf")
	resourceAccessPolicyMgr, err := resourceaccess.NewResourceAccessPolicyManager(
		resourceaccess.CasbinResourceAccessPolicyManagerConfig{
			DBConnectionString:     te.Config.Database.ConnectionString(),
			EnableDebugLog:         debugModeEnabled,
			PolicyAutoLoadInterval: time.Minute,
			PolicyFilePath:         policyFilePath,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policy manager: %w", err)
	}
	te.ResourceAccessPolicyManager = resourceAccessPolicyMgr
	return nil
}

func (te *testEnvironment) loadServices(ctx context.Context) error {
	te.Logger.Debugf("Initializing all services for test environment")

	encryptionService, err := services.NewAESEncryptionService(services.EncryptionServiceSpec{
		Masterkey: te.Config.EncryptionKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create encryption service: %w", err)
	}
	secretService := services.NewSecretService(services.SecretServiceSpec{
		SessionFactory:    te.DBSession,
		Logger:            te.Logger,
		EncryptionService: encryptionService,
	})

	stackDomainService := services.NewStackDomainsService(services.StackDomainsServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	organisationDomainService := services.NewOrganisationDomainsService(services.OrganisationDomainsServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		OrganisationDomainService: organisationDomainService,
		StackQueryService:         te.Services.StackService,
		SessionFactory:            te.DBSession,
		Logger:                    te.Logger,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              te.DBSession,
		Logger:                      te.Logger,
		JwtSecretKey:                te.Config.JwtSecret,
		ResourceAccessPolicyManager: te.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
	})

	imageRegistryService := services.NewClusterImageRegistryService(services.ImageRegistryServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		ClusterManager:       te.ClusterManager,
		ImageRegistryService: imageRegistryService,
		SessionFactory:       te.DBSession,
		Logger:               te.Logger,
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
		ClusterService: clusterService,
		UserService:    userService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:       te.DBSession,
		Logger:               te.Logger,
		WorkspaceUserService: workspaceUserService,
	})

	imageBuildService := services.NewImageBuildService(services.ImageBuildServiceSpec{
		StackResourceService: stackResourceService,
		SessionFactory:       te.DBSession,
		Logger:               te.Logger,
	})

	namespaceService := services.NewNamespaceService(services.NamespaceServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	loggingService := services.NewLoggingService(services.LoggingServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		Logger:               te.Logger,
	})

	objectStoreService := services.NewObjectStoreService(services.ObjectStoreServiceSpec{
		SessionFactory: te.DBSession,
		SecretService:  secretService,
		ClusterManager: te.ClusterManager,
		Logger:         te.Logger,
	})

	postgresBackupService := services.NewPostgresBackupService(services.PostgresBackupServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	addonUsageService := services.NewAddonUsageService(services.AddonUsageServiceSpec{
		SessionFactory: te.DBSession,
	})

	postgresAddonService := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
		SessionFactory:        te.DBSession,
		NamespaceService:      namespaceService,
		ClusterService:        clusterService,
		SecretService:         secretService,
		PostgresBackupService: postgresBackupService,
		ObjectStoreService:    objectStoreService,
		ClusterManager:        te.ClusterManager,
		Logger:                te.Logger,
	})

	stackService := services.NewStackService(services.StackServiceSpec{
		SessionFactory:         te.DBSession,
		Logger:                 te.Logger,
		VolumeService:          volumeService,
		OrganisationService:    organisationService,
		StackResourceService:   stackResourceService,
		ClusterService:         clusterService,
		ClusterRegistryService: imageRegistryService,
		NamespaceService:       namespaceService,
		SecretService:          secretService,
		PostgresAddonService:   postgresAddonService,
	})

	metricsService := services.NewMetricsService(services.MetricsServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		StackService:         stackService,
		Logger:               te.Logger,
	})

	te.Services = Services{
		UserService:                 userService,
		WorkspaceUserService:        workspaceUserService,
		OrganisationService:         organisationService,
		ClusterService:              clusterService,
		StackStorageService:         nil, // Not implemented yet
		VolumeService:               volumeService,
		StackService:                stackService,
		StackResourceService:        stackResourceService,
		ImageBuildService:           imageBuildService,
		ClusterImageRegistryService: imageRegistryService,
		StackDomainService:          stackDomainService,
		OrganisationDomainService:   organisationDomainService,
		NamespaceService:            namespaceService,
		LoggingService:              loggingService,
		MetricsService:              metricsService,
		EncryptionService:           encryptionService,
		SecretService:               secretService,
		ObjectStoreService:          objectStoreService,
		PostgresAddonService:        postgresAddonService,
		PostgresBackupService:       postgresBackupService,
		AddonUsageService:           addonUsageService,
	}

	return nil
}

func (te *testEnvironment) initializeClusterManager(ctx context.Context) error {
	te.Logger.Debugf("Setting up leadership flag for test environment")
	uuid := uuid.New().String()
	leadershipFlag, err := leadership.NewFlag().
		Process(uuid).
		Name("stackdome-test-api-server").
		Handle(te.DBSession.DirectDB()).
		Logger(te.Logger).
		Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to create leadership flag: %w", err)
	}

	te.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag: leadershipFlag,
		ControllersToRegister: []clustermanager.Controller{
			volumecontroller.NewVolumeReconciler(volumecontroller.VolumeReconcilerSpec{
				Log:            applogger.NewLoggerWithPrefix(ctx, "test-volume-controller").SetLevel(te.Logger.GetLevel()),
				StorageService: te.Services.StackStorageService,
				VolumeService:  te.Services.VolumeService,
				Env:            te.Env.Name,
			}),
			workspaceusercontroller.NewWorkspaceUserReconciler(workspaceusercontroller.WorkspaceUserReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "test-workspace-user-controller").SetLevel(te.Logger.GetLevel()),
				WorkspaceUserService: te.Services.WorkspaceUserService,
				ClusterService:       te.Services.ClusterService,
				Env:                  te.Env.Name,
			}),
			stackcontroller.NewStackReconciler(stackcontroller.StackReconcilerSpec{
				Log:          applogger.NewLoggerWithPrefix(ctx, "test-stack-controller").SetLevel(te.Logger.GetLevel()),
				StackService: te.Services.StackService,
				Env:          te.Env.Name,
			}),
			stackresourcecontroller.NewStackResourceReconciler(stackresourcecontroller.StackResourceReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "test-stack-resource-controller").SetLevel(te.Logger.GetLevel()),
				StackService:         te.Services.StackService,
				StackResourceService: te.Services.StackResourceService,
				Env:                  te.Env.Name,
			}),
			imagebuildcontroller.NewImageBuildReconciler(imagebuildcontroller.ImageBuildReconcilerSpec{
				Log:                 applogger.NewLoggerWithPrefix(ctx, "test-image-build-controller").SetLevel(te.Logger.GetLevel()),
				DBImageBuildService: te.Services.ImageBuildService,
				DBResourceService:   te.Services.StackResourceService,
			}),
			clusterimageregistry.NewClusterImageRegistryReconciler(clusterimageregistry.ClusterImageRegistryReconcilerSpec{
				Logger:                 applogger.NewLoggerWithPrefix(ctx, "test-cluster-image-registry-controller").SetLevel(te.Logger.GetLevel()),
				DBImageRegistryService: te.Services.ClusterImageRegistryService,
			}),
			postgresaddoncontroller.NewPostgresAddonReconciler(postgresaddoncontroller.PostgresAddonReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "test-postgres-addon-controller").SetLevel(te.Logger.GetLevel()),
				PostgresAddonService: te.Services.PostgresAddonService,
				Env:                  te.Env.Name,
			}),
			postgresbackupcontroller.NewPostgresBackupReconciler(postgresbackupcontroller.PostgresBackupReconcilerSpec{
				Log:                   applogger.NewLoggerWithPrefix(ctx, "test-postgres-backup-controller").SetLevel(te.Logger.GetLevel()),
				PostgresBackupService: te.Services.PostgresBackupService,
			}),
		},
	})
	te.Services.ClusterService.InjectClusterManager(te.ClusterManager)
	te.Services.PostgresAddonService.InjectClusterManager(te.ClusterManager)
	return nil
}

func (te *testEnvironment) initializeWorkerManager(ctx context.Context) error {
	te.Logger.Debugf("Initializing worker manager for test environment")
	te.WorkerManager = workermanager.NewWorkerManager(workermanager.WorkerManagerSpec{
		Environment: te.Env.Name,
	})

	stackWorker := stack.NewStackWorker(stack.StackWorkerSpec{
		StackService:         te.Services.StackService,
		SecretService:        te.Services.SecretService,
		ClusterManager:       te.ClusterManager,
		VolumeService:        te.Services.VolumeService,
		NamespaceService:     te.Services.NamespaceService,
		PostgresAddonService: te.Services.PostgresAddonService,
		AddonUsageService:    te.Services.AddonUsageService,
		Env:                  te.Env.Name,
		CRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			SecretService: te.Services.SecretService,
		}),
		SecretBuilder: builders.NewSecretBuilder(builders.SecretBuilderSpec{
			SecretFetcher: te.Services.SecretService,
		}),
	})

	te.WorkerManager.RegisterWorker(stackWorker, &models.Stack{})

	pgAddonWorker := postgresaddonworker.NewPostgresAddonWorker(postgresaddonworker.PostgresAddonWorkerSpec{
		PostgresAddonService: te.Services.PostgresAddonService,
		ObjectStoreService:   te.Services.ObjectStoreService,
		NamespaceService:     te.Services.NamespaceService,
		SecretService:        te.Services.SecretService,
		AddonUsageStore:      te.Services.AddonUsageService,
		ClusterManager:       te.ClusterManager,
		CRBuilder:            builders.NewPostgresClusterBuilder(),
		Env:                  te.Env.Name,
	})
	te.WorkerManager.RegisterWorker(pgAddonWorker, &models.PostgresAddon{})

	return nil
}

func (te *testEnvironment) injectClusterResourceServices(ctx context.Context) error {
	workspaceUserClusterResourceService := clusterresource.NewWorkspaceUserClusterResourceService(clusterresource.WorkspaceUserClusterResourceServiceSpec{
		ClusterManager: te.ClusterManager,
		Logger:         te.Logger,
		ClusterService: te.Services.ClusterService,
		UserService:    te.Services.UserService,
	})

	volumeClusterResourceService := clusterresource.NewVolumeClusterResourceService(clusterresource.VolumeClusterResourceServiceSpec{
		ClusterService:       te.Services.ClusterService,
		ClusterManager:       te.ClusterManager,
		Logger:               te.Logger,
		WorkspaceUserService: te.Services.WorkspaceUserService,
	})

	clusterStackService := clusterresource.NewClusterStackService(clusterresource.ClusterStackServiceSpec{
		ClusterManager:      te.ClusterManager,
		OrganisationService: te.Services.OrganisationService,
		Logger:              te.Logger,
		ClusterService:      te.Services.ClusterService,
	})

	clusterImageRegistryService := clusterresource.NewClusterImageRegistryService(clusterresource.ClusterImageRegistryServiceSpec{
		ClusterManager: te.ClusterManager,
		Logger:         te.Logger,
		ClusterService: te.Services.ClusterService,
	})

	clusterNamespaceService := clusterresource.NewNamespaceClusterResourceService(clusterresource.NamespaceClusterResourceServiceSpec{
		ClusterManager: te.ClusterManager,
		Logger:         te.Logger,
		ClusterService: te.Services.ClusterService,
	})

	clusterLoggingService := clusterresource.NewLoggingService(clusterresource.LoggingServiceSpec{
		ClusterManager: te.ClusterManager,
		ClusterService: te.Services.ClusterService,
		Logger:         te.Logger,
	})

	clusterMetricsService := clusterresource.NewClusterMetricsService(clusterresource.ClusterMetricsServiceSpec{
		ClusterManager: te.ClusterManager,
		ClusterService: te.Services.ClusterService,
		Logger:         te.Logger,
	})

	deps := services.ClusterResourceServiceDeps{
		ClusterStackService:     clusterStackService,
		ClusterNamespaceService: clusterNamespaceService,
		ClusterVolumeService:    volumeClusterResourceService,
		ClusterLoggingService:   clusterLoggingService,
		ClusterMetricsService:   clusterMetricsService,
	}

	dep := services.BackgroundJobEnqueuerDep{
		BackgroundJobEnqueuer: te.WorkerManager,
	}

	te.Services.WorkspaceUserService.InjectClusterResourceService(workspaceUserClusterResourceService)
	te.Services.VolumeService.InjectClusterResourceService(volumeClusterResourceService)
	te.Services.StackService.InjectClusterResourceServiceDeps(deps)
	te.Services.NamespaceService.InjectClusterResourceServiceDeps(deps)
	te.Services.LoggingService.InjectClusterResourceServiceDeps(deps)
	te.Services.MetricsService.InjectClusterResourceServiceDeps(deps)
	te.Services.ClusterImageRegistryService.InjectClusterResourceService(clusterImageRegistryService)
	te.Services.StackService.InjectBackgroundJobEnqueuer(dep)
	te.Services.PostgresAddonService.InjectBackgroundJobEnqueuer(dep)
	return nil
}

func (te *testEnvironment) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	te.Logger.Debugf("Initializing base resource access policies for test environment")

	policies := []struct {
		subject         string
		domain          string
		resource        string
		action          string
		resourceOwnerID string
	}{
		{models.UserRole.String(), "*", "/*", "*", "self"},
		{models.OrganisationAdminRole.String(), "*", "/*", "*", "*"},
		{models.PlatformAdminRole.String(), "*", "/*", "*", "*"},
	}

	for _, policy := range policies {
		if err := te.ResourceAccessPolicyManager.AddPolicy(
			policy.subject,
			policy.domain,
			policy.resource,
			policy.action,
			policy.resourceOwnerID,
		); err != nil {
			return fmt.Errorf("failed to add %s policy: %w", policy.subject, err)
		}
	}
	te.Logger.Debugf("Base resource access policies initialized for test environment")
	return nil
}

func (te *testEnvironment) ensureDefaultPlatformAdminUser(ctx context.Context) error {
	te.Logger.Debugf("Ensuring default platform admin user exists in test environment")

	defaultOrg, err := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
		SessionFactory: te.DBSession,
	}).GetDefaultOrg(ctx)
	if err != nil {
		return err
	}

	_, err = te.Services.UserService.GetDefaultUser(ctx)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			defaultUserInfo := te.BootstrapConfig.DefaultUser
			if validateErr := defaultUserInfo.Validate(); validateErr != nil {
				return fmt.Errorf("invalid default user config: %w", validateErr)
			}

			defaultUser := &models.User{
				Email:          defaultUserInfo.Email,
				Role:           models.PlatformAdminRole,
				Name:           defaultUserInfo.Name,
				Password:       defaultUserInfo.Password,
				OrganisationID: defaultOrg.ID,
				DefaultUser:    true,
			}

			if _, createErr := te.Services.UserService.Create(ctx, defaultUser); createErr != nil {
				return fmt.Errorf("failed to create default user: %v", createErr)
			}
			te.Logger.Infof("Created default platform admin user for test environment")
			return nil
		}
		return fmt.Errorf("error checking for default user: %v", err)
	}
	return nil
}

func (te *testEnvironment) startManagers(ctx context.Context) error {
	te.Logger.Debugf("Starting cluster manager for test environment")
	// Add clusters to the manager when booting up.
	clusters, err := te.Services.ClusterService.InternalListAllClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}
	for _, cluster := range clusters {
		te.Logger.Debugf("Adding cluster %s to cluster manager", cluster.ID)
		if err := te.ClusterManager.RegisterCluster(cluster); err != nil {
			return fmt.Errorf("failed to register cluster %s: %w", cluster.ID, err)
		}
	}

	te.ClusterManager.Start(ctx)

	te.Logger.Debugf("Starting worker manager for test environment")
	return te.WorkerManager.Start(ctx)
}

func (te *testEnvironment) Shutdown(ctx context.Context) error {
	te.Logger.Infof("Shutting down test environment")

	// Stop worker manager first
	if te.WorkerManager != nil {
		te.Logger.Debugf("Stopping worker manager")
		te.WorkerManager.Stop(false) // Don't drain the queue
	}

	// Stop cluster manager
	if te.ClusterManager != nil {
		te.Logger.Debugf("Stopping cluster manager")
		if err := te.ClusterManager.Stop(ctx); err != nil {
			te.Logger.Errorf("Failed to stop cluster manager: %v", err)
		}
	}

	// Close database connections
	if te.DBSession != nil {
		te.Logger.Debugf("Closing database connections")
		if err := te.DBSession.Close(); err != nil {
			te.Logger.Errorf("Failed to close database connections: %v", err)
			return err
		}
	}

	te.Logger.Infof("Test environment shutdown completed")
	return nil
}
