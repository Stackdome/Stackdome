package environment

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/controllers/clusterimageregistry"
	imagebuildcontroller "github.com/Stackdome/stackdome/pkg/controllers/imagebuild"
	postgresaddoncontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_addon"
	postgresbackupcontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_backup"
	stackcontroller "github.com/Stackdome/stackdome/pkg/controllers/stack"
	stackresourcecontroller "github.com/Stackdome/stackdome/pkg/controllers/stackresource"
	volumecontroller "github.com/Stackdome/stackdome/pkg/controllers/volume"
	workspaceusercontroller "github.com/Stackdome/stackdome/pkg/controllers/workspaceuser"
	"github.com/Stackdome/stackdome/pkg/db"
	emailpkg "github.com/Stackdome/stackdome/pkg/email"
	applogger "github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	stackresourcevalidator "github.com/Stackdome/stackdome/pkg/validator/stackresource"
	inviteworker "github.com/Stackdome/stackdome/pkg/worker/invite"
	postgresaddonworker "github.com/Stackdome/stackdome/pkg/worker/postgresaddon"
	previewworker "github.com/Stackdome/stackdome/pkg/worker/preview"
	releaseworker "github.com/Stackdome/stackdome/pkg/worker/release"
	releasegcworker "github.com/Stackdome/stackdome/pkg/worker/releasegc"
	"github.com/Stackdome/stackdome/pkg/worker/stack"
	volumeworker "github.com/Stackdome/stackdome/pkg/worker/volume"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
	"github.com/google/uuid"
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
			Name:            config.EnvironmentTest,
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
		te.initializePermissionService,
		te.loadServices,
		te.initializeClusterManager,
		te.initializeWorkerManager,
		te.injectClusterResourceServices,
		te.initializeBaseResourceAccessPolicies,
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
	// Load sane defaults for test environment, which can be overridden by environment variables.
	// This allows tests to run successfully without requiring a .env file, while still allowing configuration via env vars in CI.
	te.loadSaneDefaults()

	if err := te.Config.Validate(); err != nil {
		return fmt.Errorf("invalid application config: %w", err)
	}

	if err := te.BootstrapConfig.Validate(); err != nil {
		return fmt.Errorf("invalid bootstrap config: %w", err)
	}
	return nil
}

func (te *testEnvironment) loadSaneDefaults() {
	// We dont load from .env file in test environment since we want to rely on environment variables for configuration in CI.
	// The test bootstrap will use sensible defaults for any config values not set in environment variables, so that tests can run successfully without requiring a .env file.
	// This also ensures that CI can configure the environment via env vars without needing to manage a .env file.
	// Load standard environment variables
	// te.Config.LoadEnvVariables()
	// te.BootstrapConfig.LoadEnvVariables()
	if te.Config.JwtSecret == "" {
		if val, ok := config.EnvTestJWTSecret.Lookup(); ok {
			te.Config.JwtSecret = val
		} else {
			te.Config.JwtSecret = "ScmCX4vNcS5nj9HFSQbq7PYnRaxM29Lz9E5Z5r1A5RAWZz9li6CMqi2YSxJK5uEU"
		}
	}

	if te.Config.EncryptionKey == "" {
		if val, ok := config.EnvTestEncryptionKey.Lookup(); ok {
			te.Config.EncryptionKey = val
		} else {
			te.Config.EncryptionKey = "6193d7a7dec2e569548f0eaa46a87fb6a2d9288649dd35c827208d5e2b751d3c"
		}
	}

	if te.Config.LogLevel == "" {
		if val, ok := config.EnvTestLogLevel.Lookup(); ok {
			te.Config.LogLevel = val
		} else {
			te.Config.LogLevel = "info"
		}
	}

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

	resourceAccessPolicyMgr, err := resourceaccess.NewResourceAccessPolicyManager(
		resourceaccess.CasbinResourceAccessPolicyManagerConfig{
			DBConnectionString:     te.Config.Database.ConnectionString(),
			EnableDebugLog:         debugModeEnabled,
			PolicyAutoLoadInterval: time.Minute,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policy manager: %w", err)
	}
	te.ResourceAccessPolicyManager = resourceAccessPolicyMgr
	return nil
}

func (te *testEnvironment) initializePermissionService(ctx context.Context) error {
	te.PermissionService = auth.NewPermissionService(auth.PermissionServiceSpec{
		PolicyManager: te.ResourceAccessPolicyManager,
		ProjectStore: pgstore.NewProjectStore(pgstore.ProjectStoreSpec{
			SessionFactory: te.DBSession,
		}),
		Logger: te.Logger,
	})
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
	te.EncryptionService = encryptionService
	stackDomainService := services.NewStackDomainsService(services.StackDomainsServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	organisationDomainService := services.NewOrganisationDomainsService(services.OrganisationDomainsServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	projectService := services.NewProjectService(services.ProjectServiceSpec{
		SessionFactory: te.DBSession,
		PolicyManager:  te.ResourceAccessPolicyManager,
		Permissions:    te.PermissionService,
		Logger:         te.Logger,
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		OrganisationDomainService: organisationDomainService,
		StackQueryService:         te.Services.StackService,
		SessionFactory:            te.DBSession,
		ProjectService:               projectService,
		PolicyManager:             te.ResourceAccessPolicyManager,
		Logger:                    te.Logger,
		Permissions:               te.PermissionService,
	})

	stackStore := pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: te.DBSession})

	referenceService := services.NewReferenceService(services.ReferenceServiceSpec{
		SessionFactory: te.DBSession,
		StackStore:     stackStore,
	})

	secretService := services.NewSecretService(services.SecretServiceSpec{
		SessionFactory:    te.DBSession,
		Logger:            te.Logger,
		EncryptionService: encryptionService,
		ProjectService:       projectService,
		Permissions:       te.PermissionService,
		ReferenceService:  referenceService,
	})

	registryCredentialService := services.NewRegistryCredentialService(services.RegistryCredentialServiceSpec{
		Store: pgstore.NewRegistryCredentialStore(pgstore.RegistryCredentialStoreSpec{
			SessionFactory: te.DBSession,
		}),
		StackStore:        stackStore,
		ReferenceService:  referenceService,
		EncryptionService: encryptionService,
		Permissions:       te.PermissionService,
		Logger:            te.Logger,
	})

	gitIntegrationService := services.NewGitIntegrationService(services.GitIntegrationServiceSpec{
		Store: pgstore.NewGitIntegrationStore(pgstore.GitIntegrationStoreSpec{
			SessionFactory: te.DBSession,
		}),
		InstallationStore: pgstore.NewGitInstallationStore(pgstore.GitInstallationStoreSpec{
			SessionFactory: te.DBSession,
		}),
		OAuthStateStore: pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
			SessionFactory: te.DBSession,
		}),
		OrganisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: te.DBSession,
		}),
		AtomicExecutor: pgstore.NewAtomicExecutor(te.DBSession),
		GitHubAppClient: githubapp.NewClient(githubapp.ClientSpec{
			BaseURL: te.Config.GitHubAPIBaseURL,
		}),
		ExternalURL:       te.Config.ServerExternalURL,
		EncryptionService: encryptionService,
		Permissions:       te.PermissionService,
		Logger:            te.Logger,
	})

	credentialResolver := services.NewCredentialResolver(services.CredentialResolverSpec{
		RegistryCredentialService: registryCredentialService,
		GitIntegrationService:     gitIntegrationService,
	})

	te.RefreshTokenStore = pgstore.NewRefreshTokenStore(pgstore.RefreshTokenStoreSpec{
		SessionFactory: te.DBSession,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              te.DBSession,
		Logger:                      te.Logger,
		JwtSecretKey:                te.Config.JwtSecret,
		ResourceAccessPolicyManager: te.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
		Permissions:                 te.PermissionService,
		ProjectService:                 projectService,
		RefreshTokenStore:           te.RefreshTokenStore,
	})

	imageRegistryService := services.NewClusterImageRegistryService(services.ImageRegistryServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
		Permissions:    te.PermissionService,
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		ClusterManager:       te.ClusterManager,
		ImageRegistryService: imageRegistryService,
		SessionFactory:       te.DBSession,
		Logger:               te.Logger,
		Permissions:          te.PermissionService,
		EncryptionService:    encryptionService,
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
		ClusterService: clusterService,
		UserService:    userService,
		Permissions:    te.PermissionService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory:   te.DBSession,
		Logger:           te.Logger,
		Permissions:      te.PermissionService,
		ReferenceService: referenceService,
	})

	resourceValidator := stackresourcevalidator.NewValidator(stackresourcevalidator.ValidatorSpec{
		Volumes: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: te.DBSession,
		}),
		Secrets: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: te.DBSession,
		}),
		Domains:         organisationDomainService,
		Credentials:     credentialResolver,
		GitIntegrations: gitIntegrationService,
	})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:         te.DBSession,
		Logger:                 te.Logger,
		WorkspaceUserService:   workspaceUserService,
		Permissions:            te.PermissionService,
		StackStore:             stackStore,
		ClusterRegistryService: imageRegistryService,
		StackDomainService:     stackDomainService,
		ReferenceService:       referenceService,
		ResourceValidator:      resourceValidator,
	})

	imageBuildService := services.NewImageBuildService(services.ImageBuildServiceSpec{
		StackResourceService: stackResourceService,
		SessionFactory:       te.DBSession,
		Logger:               te.Logger,
		Permissions:          te.PermissionService,
		StackStore:           stackStore,
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
		ProjectService:    projectService,
		ClusterManager: te.ClusterManager,
		Logger:         te.Logger,
		Permissions:    te.PermissionService,
	})

	postgresBackupService := services.NewPostgresBackupService(services.PostgresBackupServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	postgresAddonService := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
		SessionFactory:        te.DBSession,
		NamespaceService:      namespaceService,
		ClusterService:        clusterService,
		SecretService:         secretService,
		PostgresBackupService: postgresBackupService,
		ObjectStoreService:    objectStoreService,
		ProjectService:           projectService,
		ClusterManager:        te.ClusterManager,
		Logger:                te.Logger,
		Permissions:           te.PermissionService,
		ReferenceService:      referenceService,
	})

	stackService := services.NewStackService(services.StackServiceSpec{
		SessionFactory:        te.DBSession,
		Logger:                te.Logger,
		VolumeService:         volumeService,
		OrganisationService:   organisationService,
		StackResourceService:  stackResourceService,
		ClusterService:        clusterService,
		NamespaceService:      namespaceService,
		SecretService:         secretService,
		PostgresAddonService:  postgresAddonService,
		ProjectService:           projectService,
		Permissions:           te.PermissionService,
		ReferenceService:      referenceService,
		CredentialResolver:    credentialResolver,
		GitIntegrationService: gitIntegrationService,
	})

	metricsService := services.NewMetricsService(services.MetricsServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		StackService:         stackService,
		Logger:               te.Logger,
	})

	apiTokenService := services.NewAPITokenService(services.APITokenServiceSpec{
		SessionFactory: te.DBSession,
		Logger:         te.Logger,
	})

	te.EmailService = emailpkg.NewNoopEmailService(applogger.NewLoggerWithPrefix(ctx, "test-email-service").SetLevel(te.Logger.GetLevel()))

	orgInviteStore := pgstore.NewOrgInviteStore(pgstore.OrgInviteStoreSpec{
		SessionFactory: te.DBSession,
	})
	orgInviteService := services.NewOrgInviteService(services.OrgInviteServiceSpec{
		InviteStore:       orgInviteStore,
		ProjectService:       projectService,
		UserService:       userService,
		EncryptionService: encryptionService,
		Permissions:       te.PermissionService,
		Logger:            te.Logger,
	})

	signupService := services.NewSignupService(services.SignupServiceSpec{
		UserService:         userService,
		OrgInviteService:    orgInviteService,
		OrganisationService: organisationService,
		ProjectService:         projectService,
		PolicyManager:       te.ResourceAccessPolicyManager,
		RefreshTokenStore:   te.RefreshTokenStore,
		JWTSecretKey:        te.Config.JwtSecret,
		JWTClaimsBuilder:    auth.NewJWTClaimsBuilder(),
		Logger:              te.Logger,
	})

	te.OAuthStateStore = pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
		SessionFactory: te.DBSession,
	})

	stackReleaseStore := pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{
		SessionFactory: te.DBSession,
	})

	releaseEventStore := pgstore.NewReleaseEventStore(pgstore.ReleaseEventStoreSpec{
		SessionFactory: te.DBSession,
	})
	releaseEventRecorder := services.NewReleaseEventRecorder(services.ReleaseEventRecorderSpec{
		Store: releaseEventStore,
	})

	stackReleaseService := services.NewStackReleaseService(services.StackReleaseServiceSpec{
		Store:              stackReleaseStore,
		StackService:       stackService,
		CredentialResolver: credentialResolver,
		Permissions:        te.PermissionService,
		ReferenceService:   referenceService,
		EventStore:         releaseEventStore,
		EventRecorder:      releaseEventRecorder,
	})

	stackService.SetReleaseService(stackReleaseService)

	stackPreviewConfigStore := pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
		SessionFactory: te.DBSession,
	})
	previewStackStore := pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
		SessionFactory: te.DBSession,
	})

	stackPreviewConfigService := services.NewStackPreviewConfigService(services.StackPreviewConfigServiceSpec{
		Store:              stackPreviewConfigStore,
		PreviewStackStore:  previewStackStore,
		CredentialResolver: credentialResolver,
		Permissions:        te.PermissionService,
	})

	previewStackService := services.NewPreviewStackService(services.PreviewStackServiceSpec{
		Store:              previewStackStore,
		ConfigStore:        stackPreviewConfigStore,
		StackService:       stackService,
		ReleaseService:     stackReleaseService,
		SecretService:      secretService,
		CredentialResolver: credentialResolver,
		Permissions:        te.PermissionService,
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
		CredentialResolver:          credentialResolver,
		RegistryCredentialService:   registryCredentialService,
		GitIntegrationService:       gitIntegrationService,
		ObjectStoreService:          objectStoreService,
		PostgresAddonService:        postgresAddonService,
		PostgresBackupService:       postgresBackupService,
		APITokenService:             apiTokenService,
		ProjectService:                 projectService,
		OrgInviteService:            orgInviteService,
		SignupService:               signupService,
		StackReleaseService:         stackReleaseService,
		ReleaseEventRecorder:        releaseEventRecorder,
		ReferenceService:            referenceService,
		StackPreviewConfigService:   stackPreviewConfigService,
		PreviewStackService:         previewStackService,
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

	te.LeadershipFlag = leadershipFlag
	te.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag:      leadershipFlag,
		CredentialDecryptor: te.EncryptionService,
		ControllersToRegister: []clustermanager.ControllerFn{
			func() clustermanager.Controller {
				return volumecontroller.NewVolumeReconciler(volumecontroller.VolumeReconcilerSpec{
					Log:            applogger.NewLoggerWithPrefix(ctx, "test-volume-controller").SetLevel(te.Logger.GetLevel()),
					StorageService: te.Services.StackStorageService,
					VolumeService:  te.Services.VolumeService,
					Env:            te.Name,
				})
			},
			func() clustermanager.Controller {
				return workspaceusercontroller.NewWorkspaceUserReconciler(workspaceusercontroller.WorkspaceUserReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "test-workspace-user-controller").SetLevel(te.Logger.GetLevel()),
					WorkspaceUserService: te.Services.WorkspaceUserService,
					ClusterService:       te.Services.ClusterService,
					Env:                  te.Name,
				})
			},
			func() clustermanager.Controller {
				return stackcontroller.NewStackReconciler(stackcontroller.StackReconcilerSpec{
					Log:            applogger.NewLoggerWithPrefix(ctx, "test-stack-controller").SetLevel(te.Logger.GetLevel()),
					StackService:   te.Services.StackService,
					Env:            te.Name,
					ReleaseChecker: te.Services.StackReleaseService,
					Enqueuer:       te.WorkerManager,
				})
			},
			func() clustermanager.Controller {
				return stackresourcecontroller.NewStackResourceReconciler(stackresourcecontroller.StackResourceReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "test-stack-resource-controller").SetLevel(te.Logger.GetLevel()),
					StackService:         te.Services.StackService,
					StackResourceService: te.Services.StackResourceService,
					Env:                  te.Name,
					ReleaseChecker:       te.Services.StackReleaseService,
					EventRecorder:        te.Services.ReleaseEventRecorder,
				})
			},
			func() clustermanager.Controller {
				return imagebuildcontroller.NewImageBuildReconciler(imagebuildcontroller.ImageBuildReconcilerSpec{
					Log:                   applogger.NewLoggerWithPrefix(ctx, "test-image-build-controller").SetLevel(te.Logger.GetLevel()),
					DBImageBuildService:   te.Services.ImageBuildService,
					DBResourceService:     te.Services.StackResourceService,
					GitIntegrationService: te.Services.GitIntegrationService,
					ReleaseChecker:        te.Services.StackReleaseService,
					EventRecorder:         te.Services.ReleaseEventRecorder,
				})
			},
			func() clustermanager.Controller {
				return clusterimageregistry.NewClusterImageRegistryReconciler(clusterimageregistry.ClusterImageRegistryReconcilerSpec{
					Logger:                 applogger.NewLoggerWithPrefix(ctx, "test-cluster-image-registry-controller").SetLevel(te.Logger.GetLevel()),
					DBImageRegistryService: te.Services.ClusterImageRegistryService,
				})
			},
			func() clustermanager.Controller {
				return postgresaddoncontroller.NewPostgresAddonReconciler(postgresaddoncontroller.PostgresAddonReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "test-postgres-addon-controller").SetLevel(te.Logger.GetLevel()),
					PostgresAddonService: te.Services.PostgresAddonService,
					Env:                  te.Name,
				})
			},
			func() clustermanager.Controller {
				return postgresbackupcontroller.NewPostgresBackupReconciler(postgresbackupcontroller.PostgresBackupReconcilerSpec{
					Log:                   applogger.NewLoggerWithPrefix(ctx, "test-postgres-backup-controller").SetLevel(te.Logger.GetLevel()),
					PostgresBackupService: te.Services.PostgresBackupService,
				})
			},
		},
	})
	te.Services.ClusterService.InjectClusterManager(te.ClusterManager)
	te.Services.PostgresAddonService.InjectClusterManager(te.ClusterManager)
	return nil
}

func (te *testEnvironment) initializeWorkerManager(ctx context.Context) error {
	te.Logger.Debugf("Initializing worker manager for test environment")
	te.WorkerManager = workermanager.NewWorkerManager(workermanager.WorkerManagerSpec{
		Environment: te.Name,
	})

	stackWorker := stack.NewStackWorker(stack.StackWorkerSpec{
		StackService:     te.Services.StackService,
		SecretService:    te.Services.SecretService,
		ClusterManager:   te.ClusterManager,
		VolumeService:    te.Services.VolumeService,
		NamespaceService: te.Services.NamespaceService,
		Env:              te.Name,
	})

	te.WorkerManager.RegisterWorker(stackWorker, &models.Stack{})

	releaseWorker := releaseworker.NewReleaseWorker(releaseworker.ReleaseWorkerSpec{
		ReleaseService:       te.Services.StackReleaseService,
		EventRecorder:        te.Services.ReleaseEventRecorder,
		StackService:         te.Services.StackService,
		ClusterManager:       te.ClusterManager,
		SecretService:        te.Services.SecretService,
		CredentialResolver:   te.Services.CredentialResolver,
		PostgresAddonService: te.Services.PostgresAddonService,
		VolumeService:        te.Services.VolumeService,
		CRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			CredentialResolver: te.Services.CredentialResolver,
		}),
		SecretBuilder: builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		Resolver: stackdeploy.NewResolver(stackdeploy.ResolverSpec{
			VolumeService:        te.Services.VolumeService,
			PostgresAddonService: te.Services.PostgresAddonService,
			SecretService:        te.Services.SecretService,
		}),
		ValidationRecords: pgstore.NewResourceValidationRecordStore(pgstore.ResourceValidationRecordStoreSpec{
			SessionFactory: te.DBSession,
		}),
		ReleaseWorkerEnqueuer: te.WorkerManager,
		Env:                   te.Name,
	})
	te.WorkerManager.RegisterWorker(releaseWorker, &models.StackRelease{})

	releaseGCWorker := releasegcworker.NewReleaseGCWorker(releasegcworker.ReleaseGCWorkerSpec{
		ReleaseStore: pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{SessionFactory: te.DBSession}),
		StackStore:   pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: te.DBSession}),
		Env:          te.Name,
	})
	te.WorkerManager.RegisterWorker(releaseGCWorker, &releasegcworker.ReleaseGCRequest{})

	volumeWorker := volumeworker.NewVolumeWorker(volumeworker.VolumeWorkerSpec{
		VolumeService:  te.Services.VolumeService,
		StackService:   te.Services.StackService,
		ClusterManager: te.ClusterManager,
		StackVolumeStore: pgstore.NewStackVolumeStore(pgstore.StackVolumeStoreSpec{
			SessionFactory: te.DBSession,
		}),
		VolumeCrBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
		Env:             te.Name,
	})
	te.WorkerManager.RegisterWorker(volumeWorker, &models.Volume{})

	pgAddonWorker := postgresaddonworker.NewPostgresAddonWorker(postgresaddonworker.PostgresAddonWorkerSpec{
		PostgresAddonService: te.Services.PostgresAddonService,
		ObjectStoreService:   te.Services.ObjectStoreService,
		NamespaceService:     te.Services.NamespaceService,
		SecretService:        te.Services.SecretService,
		ReferenceService:     te.Services.ReferenceService,
		ClusterManager:       te.ClusterManager,
		CRBuilder:            builders.NewPostgresClusterBuilder(),
		Env:                  te.Name,
	})
	te.WorkerManager.RegisterWorker(pgAddonWorker, &models.PostgresAddon{})

	inviteEmailWorker := inviteworker.NewInviteWorker(inviteworker.InviteWorkerSpec{
		InviteService:  te.Services.OrgInviteService,
		EmailService:   te.EmailService,
		LeadershipFlag: te.LeadershipFlag,
		Env:            te.Name,
	})
	te.WorkerManager.RegisterWorker(inviteEmailWorker, &models.OrgInvite{})

	inviteCleanupWorker := inviteworker.NewInviteCleanupWorker(inviteworker.InviteCleanupWorkerSpec{
		InviteService:  te.Services.OrgInviteService,
		LeadershipFlag: te.LeadershipFlag,
		Env:            te.Name,
	})
	te.WorkerManager.RegisterWorker(inviteCleanupWorker, &inviteworker.InviteCleanupBatch{})

	previewWorker := previewworker.NewPreviewWorker(previewworker.PreviewWorkerSpec{
		PreviewStackService: te.Services.PreviewStackService,
		PreviewStackStore: pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
			SessionFactory: te.DBSession,
		}),
		ConfigStore: pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
			SessionFactory: te.DBSession,
		}),
		ReleaseService: te.Services.StackReleaseService,
		StackService:   te.Services.StackService,
		Env:            te.Name,
	})
	te.WorkerManager.RegisterWorker(previewWorker, &models.PreviewStack{})

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
	te.Services.StackResourceService.InjectClusterManager(te.ClusterManager)
	te.Services.PostgresAddonService.InjectBackgroundJobEnqueuer(dep)
	te.Services.OrgInviteService.InjectBackgroundJobEnqueuer(dep)
	te.Services.StackReleaseService.InjectBackgroundJobEnqueuer(dep)
	te.Services.PreviewStackService.InjectBackgroundJobEnqueuer(dep)
	return nil
}

func (te *testEnvironment) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	te.Logger.Debugf("Initializing base resource access policies for test environment")
	if err := auth.LoadDefaultPolicies(te.ResourceAccessPolicyManager.AddPolicy); err != nil {
		return fmt.Errorf("failed to load default policies: %w", err)
	}
	te.Logger.Debugf("Base resource access policies initialized for test environment")
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
