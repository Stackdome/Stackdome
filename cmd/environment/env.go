package environment

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/bootstrap"
	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/controllers/clusterimageregistry"
	clusterinfocontroller "github.com/Stackdome/stackdome/pkg/controllers/clusterinfo"
	imagebuildcontroller "github.com/Stackdome/stackdome/pkg/controllers/imagebuild"
	postgresaddoncontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_addon"
	postgresbackupcontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_backup"
	stackcontroller "github.com/Stackdome/stackdome/pkg/controllers/stack"
	stackresourcecontroller "github.com/Stackdome/stackdome/pkg/controllers/stackresource"
	volumecontroller "github.com/Stackdome/stackdome/pkg/controllers/volume"
	"github.com/Stackdome/stackdome/pkg/db"
	emailpkg "github.com/Stackdome/stackdome/pkg/email"
	applogger "github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/observability"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	stackresourcevalidator "github.com/Stackdome/stackdome/pkg/validator/stackresource"
	clusterimageregistryworker "github.com/Stackdome/stackdome/pkg/worker/clusterimageregistry"
	inviteworker "github.com/Stackdome/stackdome/pkg/worker/invite"
	postgresaddonworker "github.com/Stackdome/stackdome/pkg/worker/postgresaddon"
	previewworker "github.com/Stackdome/stackdome/pkg/worker/preview"
	releaseworker "github.com/Stackdome/stackdome/pkg/worker/release"
	releasegcworker "github.com/Stackdome/stackdome/pkg/worker/releasegc"
	"github.com/Stackdome/stackdome/pkg/worker/stack"
	volumeworker "github.com/Stackdome/stackdome/pkg/worker/volume"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	"github.com/sirupsen/logrus"
)

type environmentImpl struct {
	*Env
	spec envSpec
}

type configLoader func() error

func newEnvironment(spec envSpec) *environmentImpl {
	return &environmentImpl{
		spec: spec,
		Env: &Env{
			Name:           spec.name,
			Config:         config.NewApplicationConfig(),
			PlatformConfig: config.NewPlatformConfig(),
		},
	}
}

type EnvConfigOption interface {
	ApplyToEnv(env *Env)
}

type ApplicationConfigOption func(*config.ApplicationConfig)

type PlatformConfigOption func(*config.PlatformConfig)

func (o ApplicationConfigOption) ApplyToEnv(env *Env) {
	o(env.Config)
}

func (o PlatformConfigOption) ApplyToEnv(env *Env) {
	o(env.PlatformConfig)
}

func WithApplicationConfig(cfg *config.ApplicationConfig) EnvConfigOption {
	return ApplicationConfigOption(func(env *config.ApplicationConfig) {
		*env = *cfg
	})
}

func WithPlatformConfig(cfg *config.PlatformConfig) EnvConfigOption {
	return PlatformConfigOption(func(env *config.PlatformConfig) {
		*env = *cfg
	})
}

func NewTestEnvironment(sessionFactory db.SessionFactory, opts ...EnvConfigOption) EnvImpl {
	res := newEnvironment(testSpec)
	res.DBSession = sessionFactory

	for _, opt := range opts {
		opt.ApplyToEnv(res.Env)
	}
	return res
}

func (e *environmentImpl) Environment() *Env {
	return e.Env
}

// loggerName prefixes component logger names so test runs are distinguishable.
func (e *environmentImpl) loggerName(component string) string {
	return e.spec.logPrefix + component
}

func (e *environmentImpl) Init(ctx context.Context) error {
	initializerSteps := []func(context.Context) error{
		e.loadEnvAndConfigs,
		e.setupLogger,
		e.setupDatabase,
		e.auditPersistedComputeTopology,
		e.setupObservability,
		e.initializeResourceAccessPolicyManager,
		e.initializePermissionService,
		e.loadServices,
		e.initializeClusterManager,
		e.initializeWorkerManager,
		e.injectClusterResourceServices,
		e.initializeBaseResourceAccessPolicies,
		e.startManagers,
		e.bootstrapSharedComputeInfrastructure,
	}

	for _, step := range initializerSteps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("failed to initialize %s environment: %w", e.Name, err)
		}
	}
	e.Logger.Infof("%s environment initialized successfully", e.Name)
	return nil
}

func (e *environmentImpl) InitDatabase(ctx context.Context) error {
	if err := e.setupDatabase(ctx); err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	return nil
}

func (e *environmentImpl) loadEnvAndConfigs(ctx context.Context) error {
	if e.spec.dependencySource.createsDependencies() {
		_ = godotenv.Load()
		loaders := []configLoader{
			e.Config.LoadEnvVariables,
			e.PlatformConfig.LoadEnvVariables,
			e.Config.LoadStackdomeCloudConfig,
		}
		if err := runConfigLoaders(loaders); err != nil {
			return err
		}
	} else {
		if err := e.loadTestDefaults(); err != nil {
			return fmt.Errorf("load test defaults: %w", err)
		}
	}

	if err := e.Config.Validate(); err != nil {
		return fmt.Errorf("invalid application config: %w", err)
	}

	if err := e.validateSharedComputeProvisioning(); err != nil {
		return fmt.Errorf("invalid shared compute provisioning config: %w", err)
	}
	if err := e.validatePlatformRouting(); err != nil {
		return fmt.Errorf("invalid platform routing config: %w", err)
	}
	return nil
}

func runConfigLoaders(loaders []configLoader) error {
	for _, loader := range loaders {
		if err := loader(); err != nil {
			return fmt.Errorf("load configuration: %w", err)
		}
	}
	return nil
}

func (e *environmentImpl) validateSharedComputeProvisioning() error {
	return config.ValidateSharedComputeProvisioning(e.Config.ComputeMode, e.Config.SharedComputeCluster)
}

func (e *environmentImpl) validatePlatformRouting() error {
	return config.ValidatePlatformRouting(e.Config.RuntimeMode, e.Config.ComputeMode, e.PlatformConfig)
}

// loadTestDefaults keeps tests runnable without a .env file while still letting
// CI configure everything through environment variables. Shared-compute and
// platform-routing configuration stays opt-in: unset in unit runs, set by the
// integration bootstrap.
func (e *environmentImpl) loadTestDefaults() error {
	if err := e.Config.SharedComputeCluster.LoadEnvVariables(); err != nil {
		return fmt.Errorf("load shared compute config: %w", err)
	}
	if err := e.PlatformConfig.LoadEnvVariables(); err != nil {
		return fmt.Errorf("load platform config: %w", err)
	}

	if e.Config.JwtSecret == "" {
		if val, ok := config.EnvTestJWTSecret.Lookup(); ok {
			e.Config.JwtSecret = val
		} else {
			e.Config.JwtSecret = "ScmCX4vNcS5nj9HFSQbq7PYnRaxM29Lz9E5Z5r1A5RAWZz9li6CMqi2YSxJK5uEU"
		}
	}

	if e.Config.EncryptionKey == "" {
		if val, ok := config.EnvTestEncryptionKey.Lookup(); ok {
			e.Config.EncryptionKey = val
		} else {
			e.Config.EncryptionKey = "6193d7a7dec2e569548f0eaa46a87fb6a2d9288649dd35c827208d5e2b751d3c"
		}
	}

	if e.Config.LogLevel == "" {
		if val, ok := config.EnvTestLogLevel.Lookup(); ok {
			e.Config.LogLevel = val
		} else {
			e.Config.LogLevel = "info"
		}
	}
	return nil
}

func (e *environmentImpl) setupLogger(ctx context.Context) error {
	logLevel, err := logrus.ParseLevel(e.Config.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", e.Config.LogLevel, err)
	}
	e.Logger = applogger.NewLoggerWithPrefix(ctx, e.loggerName("api-server")).SetLevel(logLevel)

	if !e.spec.dependencySource.createsDependencies() {
		logrus.SetOutput(os.Stdout)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
	}

	e.Logger.Debugf("Logger initialized with level: %s", logLevel.String())
	return nil
}

func (e *environmentImpl) setupDatabase(ctx context.Context) error {
	if !e.spec.dependencySource.createsDependencies() {
		// The session factory is supplied by the test bootstrap.
		return nil
	}
	_ = godotenv.Load()
	e.Config.Database.LoadEnvVariables()
	if err := e.Config.Database.Validate(); err != nil {
		return fmt.Errorf("invalid database config: %w", err)
	}
	e.DBSession = db.NewSessionFactory(e.Config.Database)
	return nil
}

func (e *environmentImpl) auditPersistedComputeTopology(ctx context.Context) error {
	clusterStore := pgstore.NewClusterStore(pgstore.ClusterStoreSpec{SessionFactory: e.DBSession})
	return checkPersistedComputeTopology(ctx, e.Config.ComputeMode, clusterStore)
}

func checkPersistedComputeTopology(ctx context.Context, mode config.ComputeMode, clusterStore stores.ClusterStore) error {
	var incompatibleSharedCompute bool
	switch mode {
	case config.ComputeModeBYOC:
		incompatibleSharedCompute = true
	case config.ComputeModeShared:
		incompatibleSharedCompute = false
	default:
		return fmt.Errorf("check persisted compute topology: unsupported compute mode %q", mode)
	}

	clusterID, err := clusterStore.FindAnyClusterIDBySharedCompute(ctx, incompatibleSharedCompute)
	if err != nil {
		return fmt.Errorf("find incompatible persisted cluster: %w", err)
	}
	if clusterID == "" {
		return nil
	}
	if mode == config.ComputeModeBYOC {
		return fmt.Errorf(
			"bring-your-own compute cannot start while shared-compute cluster %q exists; "+
				"set COMPUTE_MODE=shared or remove the shared-compute cluster and dependent resources",
			clusterID,
		)
	}
	return fmt.Errorf(
		"shared compute cannot start while tenant-owned cluster %q exists; "+
			"set COMPUTE_MODE=bring_your_own or remove the tenant-owned cluster and dependent resources",
		clusterID,
	)
}

func (e *environmentImpl) setupObservability(context.Context) error {
	e.Observability = observability.NewMetrics()
	e.Observability.RegisterStackCollector(observability.NewDatabaseStackSnapshotSource(e.DBSession))
	return nil
}

func (e *environmentImpl) initializeClusterManager(ctx context.Context) error {
	e.Logger.Debugf("Setting up leadership flag")
	uuid := uuid.New().String()
	leadershipFlag, err := leadership.NewFlag().
		Process(uuid).
		Name("stackdome-" + e.loggerName("api-server")).
		Handle(e.DBSession.DirectDB()).
		Logger(e.Logger).
		Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to create leadership flag: %w", err)
	}

	controllerLogger := func(component, clusterID string) applogger.Logger {
		return applogger.NewLoggerWithPrefix(ctx, e.loggerName(component)).SetLevel(e.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID)
	}

	e.LeadershipFlag = leadershipFlag
	e.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag:      leadershipFlag,
		CredentialDecryptor: e.EncryptionService,
		Logger:              applogger.NewLoggerWithPrefix(ctx, "cluster-manager").SetLevel(e.Logger.GetLevel()),
		ControllersToRegister: []clustermanager.ControllerFn{
			func(clusterID string) clustermanager.Controller {
				return volumecontroller.NewVolumeReconciler(volumecontroller.VolumeReconcilerSpec{
					Log:            controllerLogger("volume-controller", clusterID),
					StorageService: e.Services.StackStorageService,
					VolumeService:  e.Services.VolumeService,
					Env:            e.Name,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return stackcontroller.NewStackReconciler(stackcontroller.StackReconcilerSpec{
					Log:            controllerLogger("stack-controller", clusterID),
					StackService:   e.Services.StackService,
					Env:            e.Name,
					ReleaseChecker: e.Services.StackReleaseService,
					Enqueuer:       e.WorkerManager,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return stackresourcecontroller.NewStackResourceReconciler(stackresourcecontroller.StackResourceReconcilerSpec{
					Log:                  controllerLogger("stack-resource-controller", clusterID),
					StackService:         e.Services.StackService,
					StackResourceService: e.Services.StackResourceService,
					Env:                  e.Name,
					ReleaseResolver:      e.Services.StackReleaseService,
					EventRecorder:        e.Services.ReleaseEventRecorder,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return imagebuildcontroller.NewImageBuildReconciler(imagebuildcontroller.ImageBuildReconcilerSpec{
					Log:                   controllerLogger("image-build-controller", clusterID),
					ClusterID:             clusterID,
					DBImageBuildService:   e.Services.ImageBuildService,
					DBResourceService:     e.Services.StackResourceService,
					GitIntegrationService: e.Services.GitIntegrationService,
					ReleaseResolver:       e.Services.StackReleaseService,
					EventRecorder:         e.Services.ReleaseEventRecorder,
					StackService:          e.Services.StackService,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return clusterimageregistry.NewClusterImageRegistryReconciler(clusterimageregistry.ClusterImageRegistryReconcilerSpec{
					Logger:                 controllerLogger("cluster-image-registry-controller", clusterID),
					DBImageRegistryService: e.Services.ClusterImageRegistryService,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return postgresaddoncontroller.NewPostgresAddonReconciler(postgresaddoncontroller.PostgresAddonReconcilerSpec{
					Log:                  controllerLogger("postgres-addon-controller", clusterID),
					PostgresAddonService: e.Services.PostgresAddonService,
					Env:                  e.Name,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return postgresbackupcontroller.NewPostgresBackupReconciler(postgresbackupcontroller.PostgresBackupReconcilerSpec{
					Log:                   controllerLogger("postgres-backup-controller", clusterID),
					PostgresBackupService: e.Services.PostgresBackupService,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return clusterinfocontroller.NewClusterInfoReconciler(clusterinfocontroller.ClusterInfoReconcilerSpec{
					Log:            controllerLogger("cluster-info-controller", clusterID),
					ClusterID:      clusterID,
					ClusterService: e.Services.ClusterService,
				})
			},
		},
	})
	e.Services.ClusterService.InjectClusterManager(e.ClusterManager)
	e.Services.PostgresAddonService.InjectClusterManager(e.ClusterManager)
	return nil
}

func (e *environmentImpl) initializeResourceAccessPolicyManager(ctx context.Context) error {
	e.Logger.Debugf("Initializing resource access policy manager")
	debugModeEnabled := e.Logger.GetLevel() == logrus.DebugLevel
	resourceAccessPolicyMgr, err := resourceaccess.NewResourceAccessPolicyManager(
		resourceaccess.CasbinResourceAccessPolicyManagerConfig{
			DBConnectionString:     e.Config.Database.ConnectionString(),
			EnableDebugLog:         debugModeEnabled,
			PolicyAutoLoadInterval: time.Minute,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policy manager: %w", err)
	}
	e.ResourceAccessPolicyManager = resourceAccessPolicyMgr
	return nil
}

func (e *environmentImpl) initializePermissionService(ctx context.Context) error {
	e.PermissionService = auth.NewPermissionService(auth.PermissionServiceSpec{
		PolicyManager: e.ResourceAccessPolicyManager,
		ProjectStore: pgstore.NewProjectStore(pgstore.ProjectStoreSpec{
			SessionFactory: e.DBSession,
		}),
		Logger: e.Logger,
	})
	return nil
}

// newEmailService returns an SMTP client when the environment manages its own
// wiring and SMTP_HOST is set; otherwise a no-op client.
func (e *environmentImpl) newEmailService(ctx context.Context) emailpkg.EmailService {
	noop := emailpkg.NewNoopEmailService(applogger.NewLoggerWithPrefix(ctx, e.loggerName("email-service")).SetLevel(e.Logger.GetLevel()))
	if !e.spec.dependencySource.createsDependencies() {
		return noop
	}

	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		e.Logger.Infof("SMTP not configured, using no-op email service")
		return noop
	}

	e.Logger.Infof("SMTP email service configured")
	return emailpkg.NewSMTPEmailService(emailpkg.SMTPConfig{
		Host:        smtpHost,
		Port:        os.Getenv("SMTP_PORT"),
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		FromAddress: os.Getenv("SMTP_FROM_ADDRESS"),
		AppBaseURL:  os.Getenv("APP_BASE_URL"),
	})
}

func (e *environmentImpl) loadServices(ctx context.Context) error {
	e.Logger.Debugf("Initializing services")
	stackdomeCloudRuntime := e.Config.IsStackdomeCloud()
	customDomainsDisabled := stackdomeCloudRuntime && !e.Config.CustomDomainsEnabled()
	externalPostgresImportDisabled := stackdomeCloudRuntime && !e.Config.ExternalPostgresImportEnabled()

	e.RuntimePolicy = services.NewSelfHostedRuntimePolicy()
	if stackdomeCloudRuntime {
		cloudTrials := services.NewCloudTrialService(services.CloudTrialServiceSpec{
			Store: pgstore.NewTrialAllocationStore(pgstore.TrialAllocationStoreSpec{
				SessionFactory: e.DBSession,
			}),
			Capacity: e.Config.StackdomeCloud.Capacity.MaxActiveTrialAllocations,
			TTL:      e.Config.StackdomeCloud.Capacity.AllocationTTL.Duration(),
		})
		e.RuntimePolicy = services.NewStackdomeCloudRuntimePolicy(cloudTrials)
	}

	encryptionService, err := services.NewAESEncryptionService(services.EncryptionServiceSpec{
		Masterkey: e.Config.EncryptionKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create encryption service: %w", err)
	}
	e.EncryptionService = encryptionService
	stackDomainService := services.NewStackDomainsService(services.StackDomainsServiceSpec{
		SessionFactory:        e.DBSession,
		Logger:                e.Logger,
		PlatformBaseDomain:    e.PlatformConfig.BaseDomain,
		CustomDomainsDisabled: customDomainsDisabled,
	})

	organisationDomainService := services.NewOrganisationDomainsService(services.OrganisationDomainsServiceSpec{
		SessionFactory:        e.DBSession,
		Logger:                e.Logger,
		CustomDomainsDisabled: customDomainsDisabled,
	})

	projectService := services.NewProjectService(services.ProjectServiceSpec{
		SessionFactory: e.DBSession,
		PolicyManager:  e.ResourceAccessPolicyManager,
		Permissions:    e.PermissionService,
		Logger:         e.Logger,
	})

	imageRegistryService := services.NewClusterImageRegistryService(services.ImageRegistryServiceSpec{
		SessionFactory: e.DBSession,
		Logger:         e.Logger,
		Permissions:    e.PermissionService,
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		OrganisationDomainService: organisationDomainService,
		ImageRegistryService:      imageRegistryService,
		OrgRegistryDefaults:       e.PlatformConfig.OrgRegistry,
		StackQueryService:         e.Services.StackService,
		SessionFactory:            e.DBSession,
		ProjectService:            projectService,
		PolicyManager:             e.ResourceAccessPolicyManager,
		Permissions:               e.PermissionService,
		Logger:                    e.Logger,
		CustomDomainsDisabled:     customDomainsDisabled,
		RuntimePolicy:             e.RuntimePolicy,
	})

	stackStore := pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: e.DBSession})

	referenceService := services.NewReferenceService(services.ReferenceServiceSpec{
		SessionFactory: e.DBSession,
		StackStore:     stackStore,
	})

	secretService := services.NewSecretService(services.SecretServiceSpec{
		SessionFactory:    e.DBSession,
		Logger:            e.Logger,
		EncryptionService: encryptionService,
		ProjectService:    projectService,
		Permissions:       e.PermissionService,
		ReferenceService:  referenceService,
	})

	registryCredentialService := services.NewRegistryCredentialService(services.RegistryCredentialServiceSpec{
		Store: pgstore.NewRegistryCredentialStore(pgstore.RegistryCredentialStoreSpec{
			SessionFactory: e.DBSession,
		}),
		StackStore:        stackStore,
		ReferenceService:  referenceService,
		EncryptionService: encryptionService,
		Permissions:       e.PermissionService,
		Logger:            e.Logger,
	})

	gitIntegrationStore := pgstore.NewGitIntegrationStore(pgstore.GitIntegrationStoreSpec{
		SessionFactory: e.DBSession,
	})
	gitInstallationStore := pgstore.NewGitInstallationStore(pgstore.GitInstallationStoreSpec{
		SessionFactory: e.DBSession,
	})
	gitIntegrationService := services.NewGitIntegrationService(services.GitIntegrationServiceSpec{
		Store:             gitIntegrationStore,
		InstallationStore: gitInstallationStore,
		OAuthStateStore: pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
			SessionFactory: e.DBSession,
		}),
		OrganisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: e.DBSession,
		}),
		AtomicExecutor: pgstore.NewAtomicExecutor(e.DBSession),
		GitHubAppClient: githubapp.NewClient(githubapp.ClientSpec{
			BaseURL: e.Config.GitHubAPIBaseURL,
		}),
		ExternalURL:       e.Config.ServerExternalURL,
		PlatformApp:       platformGitHubApp(e.Config.GitHubApp),
		EncryptionService: encryptionService,
		Permissions:       e.PermissionService,
		Logger:            e.Logger,
	})

	credentialResolver := services.NewCredentialResolver(services.CredentialResolverSpec{
		RegistryCredentialService: registryCredentialService,
		GitIntegrationService:     gitIntegrationService,
	})

	e.RefreshTokenStore = pgstore.NewRefreshTokenStore(pgstore.RefreshTokenStoreSpec{
		SessionFactory: e.DBSession,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              e.DBSession,
		Logger:                      e.Logger,
		JwtSecretKey:                e.Config.JwtSecret,
		ResourceAccessPolicyManager: e.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
		Permissions:                 e.PermissionService,
		ProjectService:              projectService,
		RefreshTokenStore:           e.RefreshTokenStore,
		AtomicExecutor:              pgstore.NewAtomicExecutor(e.DBSession),
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		ClusterManager:       e.ClusterManager,
		ImageRegistryService: imageRegistryService,
		SessionFactory:       e.DBSession,
		ComputeMode:          e.Config.ComputeMode,
		PlatformTLSEnabled:   e.PlatformConfig.PlatformTLSEnabled,
		Logger:               e.Logger,
		Permissions:          e.PermissionService,
		EncryptionService:    encryptionService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory:   e.DBSession,
		Logger:           e.Logger,
		Permissions:      e.PermissionService,
		ReferenceService: referenceService,
	})

	resourceValidator := stackresourcevalidator.NewValidator(stackresourcevalidator.ValidatorSpec{
		Volumes: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: e.DBSession,
		}),
		Secrets: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: e.DBSession,
		}),
		Domains:            organisationDomainService,
		Credentials:        credentialResolver,
		GitIntegrations:    gitIntegrationService,
		PlatformBaseDomain: e.PlatformConfig.BaseDomain,
	})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:         e.DBSession,
		Logger:                 e.Logger,
		Permissions:            e.PermissionService,
		StackStore:             stackStore,
		ClusterRegistryService: imageRegistryService,
		StackDomainService:     stackDomainService,
		ReferenceService:       referenceService,
		ResourceValidator:      resourceValidator,
	})

	imageBuildService := services.NewImageBuildService(services.ImageBuildServiceSpec{
		StackResourceService: stackResourceService,
		SessionFactory:       e.DBSession,
		Logger:               e.Logger,
		Permissions:          e.PermissionService,
		StackStore:           stackStore,
	})

	namespaceService := services.NewNamespaceService(services.NamespaceServiceSpec{
		SessionFactory: e.DBSession,
		Logger:         e.Logger,
		SharedCompute:  e.Config.UsesSharedCompute(),
	})

	loggingService := services.NewLoggingService(services.LoggingServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		ImageBuildService:    imageBuildService,
		Logger:               e.Logger,
	})

	objectStoreService := services.NewObjectStoreService(services.ObjectStoreServiceSpec{
		SessionFactory: e.DBSession,
		SecretService:  secretService,
		ProjectService: projectService,
		ClusterManager: e.ClusterManager,
		Logger:         e.Logger,
		Permissions:    e.PermissionService,
	})

	postgresBackupService := services.NewPostgresBackupService(services.PostgresBackupServiceSpec{
		SessionFactory: e.DBSession,
		Logger:         e.Logger,
	})

	postgresAddonService := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
		SessionFactory:         e.DBSession,
		NamespaceService:       namespaceService,
		ClusterService:         clusterService,
		SecretService:          secretService,
		PostgresBackupService:  postgresBackupService,
		ObjectStoreService:     objectStoreService,
		ProjectService:         projectService,
		ClusterManager:         e.ClusterManager,
		Logger:                 e.Logger,
		Permissions:            e.PermissionService,
		ReferenceService:       referenceService,
		ExternalImportDisabled: externalPostgresImportDisabled,
	})

	stackService := services.NewStackService(services.StackServiceSpec{
		SessionFactory:        e.DBSession,
		Logger:                e.Logger,
		VolumeService:         volumeService,
		OrganisationService:   organisationService,
		StackResourceService:  stackResourceService,
		ClusterService:        clusterService,
		NamespaceService:      namespaceService,
		SecretService:         secretService,
		PostgresAddonService:  postgresAddonService,
		ProjectService:        projectService,
		Permissions:           e.PermissionService,
		ReferenceService:      referenceService,
		CredentialResolver:    credentialResolver,
		GitIntegrationService: gitIntegrationService,
		PlatformBaseDomain:    e.PlatformConfig.BaseDomain,
		RuntimePolicy:         e.RuntimePolicy,
	})

	metricsService := services.NewMetricsService(services.MetricsServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		StackService:         stackService,
		Logger:               e.Logger,
	})

	apiTokenService := services.NewAPITokenService(services.APITokenServiceSpec{
		SessionFactory: e.DBSession,
		Logger:         e.Logger,
	})

	e.EmailService = e.newEmailService(ctx)

	orgInviteStore := pgstore.NewOrgInviteStore(pgstore.OrgInviteStoreSpec{
		SessionFactory: e.DBSession,
	})
	orgInviteService := services.NewOrgInviteService(services.OrgInviteServiceSpec{
		InviteStore:       orgInviteStore,
		ProjectService:    projectService,
		UserService:       userService,
		EncryptionService: encryptionService,
		Permissions:       e.PermissionService,
		Logger:            e.Logger,
	})

	signupService := services.NewSignupService(services.SignupServiceSpec{
		UserService:         userService,
		OrgInviteService:    orgInviteService,
		OrganisationService: organisationService,
		ProjectService:      projectService,
		PolicyManager:       e.ResourceAccessPolicyManager,
		RefreshTokenStore:   e.RefreshTokenStore,
		AtomicExecutor:      pgstore.NewAtomicExecutor(e.DBSession),
		JWTSecretKey:        e.Config.JwtSecret,
		JWTClaimsBuilder:    auth.NewJWTClaimsBuilder(),
		Logger:              e.Logger,
	})

	e.OAuthStateStore = pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
		SessionFactory: e.DBSession,
	})

	stackReleaseStore := pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{
		SessionFactory: e.DBSession,
	})

	releaseEventStore := pgstore.NewReleaseEventStore(pgstore.ReleaseEventStoreSpec{
		SessionFactory: e.DBSession,
	})
	releaseEventRecorder := services.NewReleaseEventRecorder(services.ReleaseEventRecorderSpec{
		Store: releaseEventStore,
	})

	stackReleaseService := services.NewStackReleaseService(services.StackReleaseServiceSpec{
		Store:              stackReleaseStore,
		StackService:       stackService,
		CredentialResolver: credentialResolver,
		Permissions:        e.PermissionService,
		ReferenceService:   referenceService,
		EventStore:         releaseEventStore,
		EventRecorder:      releaseEventRecorder,
		RuntimePolicy:      e.RuntimePolicy,
	})

	stackService.SetReleaseService(stackReleaseService)

	stackPreviewConfigStore := pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
		SessionFactory: e.DBSession,
	})
	previewStackStore := pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
		SessionFactory: e.DBSession,
	})

	stackPreviewConfigService := services.NewStackPreviewConfigService(services.StackPreviewConfigServiceSpec{
		Store:              stackPreviewConfigStore,
		PreviewStackStore:  previewStackStore,
		CredentialResolver: credentialResolver,
		Permissions:        e.PermissionService,
	})

	previewStackService := services.NewPreviewStackService(services.PreviewStackServiceSpec{
		Store:              previewStackStore,
		ConfigStore:        stackPreviewConfigStore,
		StackService:       stackService,
		ReleaseService:     stackReleaseService,
		SecretService:      secretService,
		CredentialResolver: credentialResolver,
		Permissions:        e.PermissionService,
		Logger:             e.Logger,
	})

	previewWebhookService := services.NewPreviewWebhookService(services.PreviewWebhookServiceSpec{
		ConfigStore:       stackPreviewConfigStore,
		PreviewStackStore: previewStackStore,
		PreviewService:    previewStackService,
		Logger:            e.Logger,
	})

	githubWebhookService := services.NewGitHubWebhookService(services.GitHubWebhookServiceSpec{
		Store:             gitIntegrationStore,
		InstallationStore: gitInstallationStore,
		EncryptionService: encryptionService,
		PreviewWebhook:    previewWebhookService,
		Logger:            e.Logger,
		PlatformApp:       platformGitHubApp(e.Config.GitHubApp),
	})

	e.Services = Services{
		UserService:                 userService,
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
		GitHubWebhookService:        githubWebhookService,
		ObjectStoreService:          objectStoreService,
		PostgresAddonService:        postgresAddonService,
		PostgresBackupService:       postgresBackupService,
		APITokenService:             apiTokenService,
		ProjectService:              projectService,
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

func (e *environmentImpl) initializeWorkerManager(ctx context.Context) error {
	e.Logger.Debugf("Initializing worker manager")
	e.WorkerManager = workermanager.NewWorkerManager(workermanager.WorkerManagerSpec{
		Environment: e.Name,
		Metrics:     e.Observability,
	})

	stackWorker := stack.NewStackWorker(stack.StackWorkerSpec{
		StackService:     e.Services.StackService,
		SecretService:    e.Services.SecretService,
		ClusterManager:   e.ClusterManager,
		VolumeService:    e.Services.VolumeService,
		NamespaceService: e.Services.NamespaceService,
		Env:              e.Name,
		RuntimePolicy:    e.RuntimePolicy,
	})

	e.WorkerManager.RegisterWorker(stackWorker, models.StackOperand{})

	releaseWorker := releaseworker.NewReleaseWorker(releaseworker.ReleaseWorkerSpec{
		ReleaseService:       e.Services.StackReleaseService,
		EventRecorder:        e.Services.ReleaseEventRecorder,
		StackService:         e.Services.StackService,
		ImageBuildService:    e.Services.ImageBuildService,
		ClusterManager:       e.ClusterManager,
		SecretService:        e.Services.SecretService,
		CredentialResolver:   e.Services.CredentialResolver,
		PostgresAddonService: e.Services.PostgresAddonService,
		VolumeService:        e.Services.VolumeService,
		NamespaceService:     e.Services.NamespaceService,
		RuntimePolicy:        e.RuntimePolicy,
		CRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			CredentialResolver: e.Services.CredentialResolver,
			ComputeMode:        e.Config.ComputeMode,
			PlatformTLSEnabled: e.PlatformConfig.PlatformTLSEnabled,
			PlatformBaseDomain: e.PlatformConfig.BaseDomain,
		}),
		SecretBuilder: builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		Resolver: stackdeploy.NewResolver(stackdeploy.ResolverSpec{
			VolumeService:        e.Services.VolumeService,
			PostgresAddonService: e.Services.PostgresAddonService,
			SecretService:        e.Services.SecretService,
		}),
		ValidationRecords: pgstore.NewResourceValidationRecordStore(pgstore.ResourceValidationRecordStoreSpec{
			SessionFactory: e.DBSession,
		}),
		ReleaseWorkerEnqueuer: e.WorkerManager,
		Env:                   e.Name,
	})
	e.WorkerManager.RegisterWorker(releaseWorker, models.StackReleaseOperand{})

	releaseGCWorker := releasegcworker.NewReleaseGCWorker(releasegcworker.ReleaseGCWorkerSpec{
		ReleaseStore: pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{SessionFactory: e.DBSession}),
		StackStore:   pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: e.DBSession}),
		Env:          e.Name,
	})
	e.WorkerManager.RegisterWorker(releaseGCWorker, releasegcworker.ReleaseGCRequest{})

	volumeWorker := volumeworker.NewVolumeWorker(volumeworker.VolumeWorkerSpec{
		VolumeService:  e.Services.VolumeService,
		StackService:   e.Services.StackService,
		ClusterManager: e.ClusterManager,
		StackVolumeStore: pgstore.NewStackVolumeStore(pgstore.StackVolumeStoreSpec{
			SessionFactory: e.DBSession,
		}),
		VolumeCrBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
		Env:             e.Name,
		RuntimePolicy:   e.RuntimePolicy,
	})
	e.WorkerManager.RegisterWorker(volumeWorker, models.VolumeOperand{})

	clusterImageRegistryResource := clusterresource.NewClusterImageRegistryService(clusterresource.ClusterImageRegistryServiceSpec{
		ClusterManager: e.ClusterManager,
		Logger:         e.Logger,
	})
	clusterImageRegistryWorker := clusterimageregistryworker.NewClusterImageRegistryWorker(clusterimageregistryworker.ClusterImageRegistryWorkerSpec{
		ClusterStore: pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
			SessionFactory: e.DBSession,
		}),
		ImageRegistryStore: pgstore.NewClusterImageRegistryStore(pgstore.ClusterImageRegistryStoreSpec{
			SessionFactory: e.DBSession,
		}),
		ClusterManager:  e.ClusterManager,
		ClusterResource: clusterImageRegistryResource,
		Env:             e.Name,
	})
	e.WorkerManager.RegisterWorker(clusterImageRegistryWorker, models.ClusterImageRegistryOperand{})

	pgAddonWorker := postgresaddonworker.NewPostgresAddonWorker(postgresaddonworker.PostgresAddonWorkerSpec{
		PostgresAddonService: e.Services.PostgresAddonService,
		ObjectStoreService:   e.Services.ObjectStoreService,
		NamespaceService:     e.Services.NamespaceService,
		SecretService:        e.Services.SecretService,
		ReferenceService:     e.Services.ReferenceService,
		ClusterManager:       e.ClusterManager,
		CRBuilder:            builders.NewPostgresClusterBuilder(),
		Env:                  e.Name,
		RuntimePolicy:        e.RuntimePolicy,
	})
	e.WorkerManager.RegisterWorker(pgAddonWorker, models.PostgresAddonOperand{})

	inviteEmailWorker := inviteworker.NewInviteWorker(inviteworker.InviteWorkerSpec{
		InviteService:  e.Services.OrgInviteService,
		EmailService:   e.EmailService,
		LeadershipFlag: e.LeadershipFlag,
		Env:            e.Name,
	})
	e.WorkerManager.RegisterWorker(inviteEmailWorker, models.OrgInviteOperand{})

	inviteCleanupWorker := inviteworker.NewInviteCleanupWorker(inviteworker.InviteCleanupWorkerSpec{
		InviteService:  e.Services.OrgInviteService,
		LeadershipFlag: e.LeadershipFlag,
		Env:            e.Name,
	})
	e.WorkerManager.RegisterWorker(inviteCleanupWorker, inviteworker.InviteCleanupRequest{})

	previewStackStore := pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
		SessionFactory: e.DBSession,
	})
	previewConfigStore := pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
		SessionFactory: e.DBSession,
	})
	previewCommentService := services.NewPreviewCommentService(services.PreviewCommentServiceSpec{
		ConfigStore:     previewConfigStore,
		GitIntegrations: e.Services.GitIntegrationService,
		Commenter:       githubapp.NewPullRequestCommenter(githubapp.PullRequestCommenterSpec{}),
		Logger:          e.Logger,
	})
	previewWorker := previewworker.NewPreviewWorker(previewworker.PreviewWorkerSpec{
		PreviewStackService: e.Services.PreviewStackService,
		PreviewStackStore:   previewStackStore,
		ConfigStore:         previewConfigStore,
		ReleaseService:      e.Services.StackReleaseService,
		StackService:        e.Services.StackService,
		CommentService:      previewCommentService,
		Env:                 e.Name,
	})
	e.WorkerManager.RegisterWorker(previewWorker, models.PreviewStackOperand{})

	return nil
}

func (e *environmentImpl) injectClusterResourceServices(ctx context.Context) error {
	volumeClusterResourceService := clusterresource.NewVolumeClusterResourceService(clusterresource.VolumeClusterResourceServiceSpec{
		ClusterService: e.Services.ClusterService,
		ClusterManager: e.ClusterManager,
		Logger:         e.Logger,
	})

	clusterNamespaceService := clusterresource.NewNamespaceClusterResourceService(clusterresource.NamespaceClusterResourceServiceSpec{
		ClusterManager: e.ClusterManager,
		Logger:         e.Logger,
		ClusterService: e.Services.ClusterService,
	})

	clusterLoggingService := clusterresource.NewLoggingService(clusterresource.LoggingServiceSpec{
		ClusterManager: e.ClusterManager,
		ClusterService: e.Services.ClusterService,
		Logger:         e.Logger,
	})

	clusterMetricsService := clusterresource.NewClusterMetricsService(clusterresource.ClusterMetricsServiceSpec{
		ClusterManager: e.ClusterManager,
		ClusterService: e.Services.ClusterService,
		Logger:         e.Logger,
	})

	deps := services.ClusterResourceServiceDeps{
		ClusterNamespaceService: clusterNamespaceService,
		ClusterVolumeService:    volumeClusterResourceService,
		ClusterLoggingService:   clusterLoggingService,
		ClusterMetricsService:   clusterMetricsService,
	}

	dep := services.BackgroundJobEnqueuerDep{
		BackgroundJobEnqueuer: e.WorkerManager,
	}

	e.Services.VolumeService.InjectClusterResourceService(volumeClusterResourceService)
	e.Services.StackService.InjectClusterResourceServiceDeps(deps)
	e.Services.NamespaceService.InjectClusterResourceServiceDeps(deps)
	e.Services.LoggingService.InjectClusterResourceServiceDeps(deps)
	e.Services.MetricsService.InjectClusterResourceServiceDeps(deps)
	e.Services.ClusterImageRegistryService.InjectBackgroundJobEnqueuer(dep)
	e.Services.ClusterService.InjectBackgroundJobEnqueuer(dep)
	e.Services.StackService.InjectBackgroundJobEnqueuer(dep)
	e.Services.StackResourceService.InjectClusterManager(e.ClusterManager)
	e.Services.PostgresAddonService.InjectBackgroundJobEnqueuer(dep)
	e.Services.OrgInviteService.InjectBackgroundJobEnqueuer(dep)
	e.Services.StackReleaseService.InjectBackgroundJobEnqueuer(dep)
	e.Services.PreviewStackService.InjectBackgroundJobEnqueuer(dep)
	return nil
}

func (e *environmentImpl) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	e.Logger.Debugf("Initializing base resource access policies")
	if err := auth.LoadDefaultPolicies(e.ResourceAccessPolicyManager.AddPolicy); err != nil {
		return fmt.Errorf("failed to load default policies: %w", err)
	}
	e.Logger.Debugf("Base resource access policies initialized")
	return nil
}

func (e *environmentImpl) startManagers(ctx context.Context) error {
	e.Logger.Debugf("Starting cluster manager")
	// Add clusters to the manager when booting up.
	clusters, err := e.Services.ClusterService.InternalListAllClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}
	for _, cluster := range clusters {
		e.Logger.Debugf("Adding cluster %s to cluster manager", cluster.ID)
		if err := e.ClusterManager.RegisterCluster(cluster); err != nil {
			return fmt.Errorf("failed to register cluster %s: %w", cluster.ID, err)
		}
	}

	e.ClusterManager.Start(ctx)

	e.Logger.Debugf("Starting worker manager")
	return e.WorkerManager.Start(ctx)
}

func (e *environmentImpl) bootstrapSharedComputeInfrastructure(ctx context.Context) error {
	svc := bootstrap.NewService(bootstrap.Spec{
		OrganisationService: e.Services.OrganisationService,
		ClusterService:      e.Services.ClusterService,
		PlatformConfig:      e.PlatformConfig,
		ClusterConfig:       e.Config.SharedComputeCluster,
		Logger:              e.Logger,
	})
	if err := svc.Run(ctx); err != nil {
		return fmt.Errorf("platform bootstrap failed: %w", err)
	}
	return nil
}

func (e *environmentImpl) Shutdown(ctx context.Context) error {
	e.Logger.Infof("Shutting down %s environment", e.Name)

	// Stop worker manager first
	if e.WorkerManager != nil {
		e.Logger.Debugf("Stopping worker manager")
		e.WorkerManager.Stop(false) // Don't drain the queue
	}

	// Stop cluster manager
	if e.ClusterManager != nil {
		e.Logger.Debugf("Stopping cluster manager")
		if err := e.ClusterManager.Stop(ctx); err != nil {
			e.Logger.Errorf("Failed to stop cluster manager: %v", err)
		}
	}

	// Close database connections
	if e.DBSession != nil {
		e.Logger.Debugf("Closing database connections")
		if err := e.DBSession.Close(); err != nil {
			e.Logger.Errorf("Failed to close database connections: %v", err)
			return err
		}
	}

	e.Logger.Infof("%s environment shutdown completed", e.Name)
	return nil
}

// platformGitHubApp converts the GITHUB_APP_* config into app credentials, or
// nil when the hub runs without a platform-wide app and each org creates its
// own through the manifest flow.
func platformGitHubApp(cfg *config.GitHubAppConfig) *githubapp.AppCredentials {
	if !cfg.Configured() {
		return nil
	}
	return &githubapp.AppCredentials{
		AppID:         cfg.AppID,
		Slug:          cfg.Slug,
		PEM:           cfg.PrivateKey,
		WebhookSecret: cfg.WebhookSecret,
	}
}
