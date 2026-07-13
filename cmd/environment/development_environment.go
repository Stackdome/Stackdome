package environment

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/clustermanager"
	"github.com/Stackdome/stackdome/pkg/controllers/clusterimageregistry"
	imagebuildcontroller "github.com/Stackdome/stackdome/pkg/controllers/imagebuild"
	postgresaddoncontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_addon"
	postgresbackupcontroller "github.com/Stackdome/stackdome/pkg/controllers/postgres_backup"
	stackcontroller "github.com/Stackdome/stackdome/pkg/controllers/stack"
	stackresourcecontroller "github.com/Stackdome/stackdome/pkg/controllers/stackresource"
	emailpkg "github.com/Stackdome/stackdome/pkg/email"
	"github.com/Stackdome/stackdome/pkg/stackdeploy"
	inviteworker "github.com/Stackdome/stackdome/pkg/worker/invite"
	postgresaddonworker "github.com/Stackdome/stackdome/pkg/worker/postgresaddon"
	previewworker "github.com/Stackdome/stackdome/pkg/worker/preview"
	releaseworker "github.com/Stackdome/stackdome/pkg/worker/release"
	releasegcworker "github.com/Stackdome/stackdome/pkg/worker/releasegc"
	"github.com/Stackdome/stackdome/pkg/worker/stack"
	volumeworker "github.com/Stackdome/stackdome/pkg/worker/volume"
	"github.com/Stackdome/stackdome/pkg/worker/workermanager"

	volumecontroller "github.com/Stackdome/stackdome/pkg/controllers/volume"
	workspaceusercontroller "github.com/Stackdome/stackdome/pkg/controllers/workspaceuser"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/db"
	applogger "github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/resourceaccess"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
	stackresourcevalidator "github.com/Stackdome/stackdome/pkg/validator/stackresource"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	"github.com/sirupsen/logrus"
)

type developmentEnvironment struct {
	*Env
}

func NewDevelopmentEnvironment() EnvImpl {
	return &developmentEnvironment{
		Env: &Env{
			Name:            config.EnvironmentDevelopment,
			Config:          config.NewApplicationConfig(),
			BootstrapConfig: config.NewBootstrapConfig(),
		},
	}
}

func (d *developmentEnvironment) Environment() *Env {
	return d.Env
}

func (d *developmentEnvironment) Init(ctx context.Context) error {
	initializerSteps := []func(context.Context) error{
		d.loadEnvAndConfigs,
		d.setupLogger,
		d.setupDatabase,
		d.initializeResourceAccessPolicyManager,
		d.initializePermissionService,
		d.loadServices,
		d.initializeClusterManager,
		d.initializeWorkerManager,
		d.injectClusterResourceServices,
		d.initializeBaseResourceAccessPolicies,
		d.startManagers,
	}

	for _, step := range initializerSteps {
		if err := step(ctx); err != nil {
			return fmt.Errorf("failed to initialize environment: %w", err)
		}
	}
	d.Logger.Infof("Development environment initialized successfully")
	return nil
}

func (d *developmentEnvironment) InitDatabase(ctx context.Context) error {
	if err := d.setupDatabase(ctx); err != nil {
		return fmt.Errorf("failed to setup database: %w", err)
	}
	return nil
}

func (d *developmentEnvironment) initializeWorkerManager(ctx context.Context) error {
	d.Logger.Debugf("Initializing worker manager")
	d.WorkerManager = workermanager.NewWorkerManager(workermanager.WorkerManagerSpec{
		Environment: d.Name,
	})

	stackWorker := stack.NewStackWorker(stack.StackWorkerSpec{
		StackService:     d.Services.StackService,
		SecretService:    d.Services.SecretService,
		ClusterManager:   d.ClusterManager,
		VolumeService:    d.Services.VolumeService,
		NamespaceService: d.Services.NamespaceService,
		Env:              d.Name,
	})

	d.WorkerManager.RegisterWorker(stackWorker, &models.Stack{})

	releaseWorker := releaseworker.NewReleaseWorker(releaseworker.ReleaseWorkerSpec{
		ReleaseService:       d.Services.StackReleaseService,
		EventRecorder:        d.Services.ReleaseEventRecorder,
		StackService:         d.Services.StackService,
		ClusterManager:       d.ClusterManager,
		SecretService:        d.Services.SecretService,
		CredentialResolver:   d.Services.CredentialResolver,
		PostgresAddonService: d.Services.PostgresAddonService,
		VolumeService:        d.Services.VolumeService,
		CRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			CredentialResolver: d.Services.CredentialResolver,
		}),
		SecretBuilder: builders.NewSecretBuilder(builders.SecretBuilderSpec{}),
		Resolver: stackdeploy.NewResolver(stackdeploy.ResolverSpec{
			VolumeService:        d.Services.VolumeService,
			PostgresAddonService: d.Services.PostgresAddonService,
			SecretService:        d.Services.SecretService,
		}),
		ValidationRecords: pgstore.NewResourceValidationRecordStore(pgstore.ResourceValidationRecordStoreSpec{
			SessionFactory: d.DBSession,
		}),
		ReleaseWorkerEnqueuer: d.WorkerManager,
		Env:                   d.Name,
	})
	d.WorkerManager.RegisterWorker(releaseWorker, &models.StackRelease{})

	releaseGCWorker := releasegcworker.NewReleaseGCWorker(releasegcworker.ReleaseGCWorkerSpec{
		ReleaseStore: pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{SessionFactory: d.DBSession}),
		StackStore:   pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: d.DBSession}),
		Env:          d.Name,
	})
	d.WorkerManager.RegisterWorker(releaseGCWorker, &releasegcworker.ReleaseGCRequest{})

	volumeWorker := volumeworker.NewVolumeWorker(volumeworker.VolumeWorkerSpec{
		VolumeService:  d.Services.VolumeService,
		StackService:   d.Services.StackService,
		ClusterManager: d.ClusterManager,
		StackVolumeStore: pgstore.NewStackVolumeStore(pgstore.StackVolumeStoreSpec{
			SessionFactory: d.DBSession,
		}),
		VolumeCrBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
		Env:             d.Name,
	})
	d.WorkerManager.RegisterWorker(volumeWorker, &models.Volume{})

	pgAddonWorker := postgresaddonworker.NewPostgresAddonWorker(postgresaddonworker.PostgresAddonWorkerSpec{
		PostgresAddonService: d.Services.PostgresAddonService,
		ObjectStoreService:   d.Services.ObjectStoreService,
		NamespaceService:     d.Services.NamespaceService,
		SecretService:        d.Services.SecretService,
		ReferenceService:     d.Services.ReferenceService,
		ClusterManager:       d.ClusterManager,
		CRBuilder:            builders.NewPostgresClusterBuilder(),
		Env:                  d.Name,
	})
	d.WorkerManager.RegisterWorker(pgAddonWorker, &models.PostgresAddon{})

	inviteEmailWorker := inviteworker.NewInviteWorker(inviteworker.InviteWorkerSpec{
		InviteService:  d.Services.OrgInviteService,
		EmailService:   d.EmailService,
		LeadershipFlag: d.LeadershipFlag,
		Env:            d.Name,
	})
	d.WorkerManager.RegisterWorker(inviteEmailWorker, &models.OrgInvite{})

	inviteCleanupWorker := inviteworker.NewInviteCleanupWorker(inviteworker.InviteCleanupWorkerSpec{
		InviteService:  d.Services.OrgInviteService,
		LeadershipFlag: d.LeadershipFlag,
		Env:            d.Name,
	})
	d.WorkerManager.RegisterWorker(inviteCleanupWorker, &inviteworker.InviteCleanupBatch{})

	previewWorker := previewworker.NewPreviewWorker(previewworker.PreviewWorkerSpec{
		PreviewStackService: d.Services.PreviewStackService,
		PreviewStackStore: pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
			SessionFactory: d.DBSession,
		}),
		ConfigStore: pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
			SessionFactory: d.DBSession,
		}),
		ReleaseService: d.Services.StackReleaseService,
		StackService:   d.Services.StackService,
		Env:            d.Name,
	})
	d.WorkerManager.RegisterWorker(previewWorker, &models.PreviewStack{})

	return nil
}

func (d *developmentEnvironment) loadEnvAndConfigs(ctx context.Context) error {
	_ = godotenv.Load()

	d.Config.LoadEnvVariables()
	d.BootstrapConfig.LoadEnvVariables()

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid application config: %w", err)
	}
	return nil
}

func (d *developmentEnvironment) setupLogger(ctx context.Context) error {
	logLevel, err := logrus.ParseLevel(d.Config.LogLevel)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", d.Config.LogLevel, err)
	}
	d.Logger = applogger.NewLoggerWithPrefix(ctx, "api-server").SetLevel(logLevel)
	d.Logger.Debugf("Logger initialized with level: %s", logLevel.String())
	return nil
}

func (d *developmentEnvironment) setupDatabase(ctx context.Context) error {
	_ = godotenv.Load()
	d.Config.Database.LoadEnvVariables()
	if err := d.Config.Database.Validate(); err != nil {
		return fmt.Errorf("invalid database config: %w", err)
	}
	d.DBSession = db.NewSessionFactory(d.Config.Database)
	return nil
}

func (d *developmentEnvironment) initializeClusterManager(ctx context.Context) error {
	d.Logger.Debugf("Setting up leadership flag")
	uuid := uuid.New().String()
	leadershipFlag, err := leadership.NewFlag().
		Process(uuid).
		Name("stackdome-api-server").
		Handle(d.DBSession.DirectDB()).
		Logger(d.Logger).
		Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to create leadership flag: %w", err)
	}

	d.LeadershipFlag = leadershipFlag
	d.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag:      leadershipFlag,
		CredentialDecryptor: d.EncryptionService,
		Logger:              applogger.NewLoggerWithPrefix(ctx, "cluster-manager").SetLevel(d.Logger.GetLevel()),
		ControllersToRegister: []clustermanager.ControllerFn{
			func(clusterID string) clustermanager.Controller {
				return volumecontroller.NewVolumeReconciler(volumecontroller.VolumeReconcilerSpec{
					Log:            applogger.NewLoggerWithPrefix(ctx, "volume-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					StorageService: d.Services.StackStorageService,
					VolumeService:  d.Services.VolumeService,
					Env:            d.Name,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return workspaceusercontroller.NewWorkspaceUserReconciler(workspaceusercontroller.WorkspaceUserReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "workspace-user-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					WorkspaceUserService: d.Services.WorkspaceUserService,
					ClusterService:       d.Services.ClusterService,
					Env:                  d.Name,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return stackcontroller.NewStackReconciler(stackcontroller.StackReconcilerSpec{
					Log:            applogger.NewLoggerWithPrefix(ctx, "stack-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					StackService:   d.Services.StackService,
					Env:            d.Name,
					ReleaseChecker: d.Services.StackReleaseService,
					Enqueuer:       d.WorkerManager,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return stackresourcecontroller.NewStackResourceReconciler(stackresourcecontroller.StackResourceReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "stack-resource-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					StackService:         d.Services.StackService,
					StackResourceService: d.Services.StackResourceService,
					Env:                  d.Name,
					ReleaseChecker:       d.Services.StackReleaseService,
					EventRecorder:        d.Services.ReleaseEventRecorder,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return imagebuildcontroller.NewImageBuildReconciler(imagebuildcontroller.ImageBuildReconcilerSpec{
					Log:                   applogger.NewLoggerWithPrefix(ctx, "image-build-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					DBImageBuildService:   d.Services.ImageBuildService,
					DBResourceService:     d.Services.StackResourceService,
					GitIntegrationService: d.Services.GitIntegrationService,
					ReleaseChecker:        d.Services.StackReleaseService,
					EventRecorder:         d.Services.ReleaseEventRecorder,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return clusterimageregistry.NewClusterImageRegistryReconciler(clusterimageregistry.ClusterImageRegistryReconcilerSpec{
					Logger:                 applogger.NewLoggerWithPrefix(ctx, "cluster-image-registry-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					DBImageRegistryService: d.Services.ClusterImageRegistryService,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return postgresaddoncontroller.NewPostgresAddonReconciler(postgresaddoncontroller.PostgresAddonReconcilerSpec{
					Log:                  applogger.NewLoggerWithPrefix(ctx, "postgres-addon-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					PostgresAddonService: d.Services.PostgresAddonService,
					Env:                  d.Name,
				})
			},
			func(clusterID string) clustermanager.Controller {
				return postgresbackupcontroller.NewPostgresBackupReconciler(postgresbackupcontroller.PostgresBackupReconcilerSpec{
					Log:                   applogger.NewLoggerWithPrefix(ctx, "postgres-backup-controller").SetLevel(d.Logger.GetLevel()).WithField(applogger.FieldClusterID, clusterID),
					PostgresBackupService: d.Services.PostgresBackupService,
				})
			},
		},
	})
	d.Services.ClusterService.InjectClusterManager(d.ClusterManager)
	d.Services.PostgresAddonService.InjectClusterManager(d.ClusterManager)
	return nil
}

func (d *developmentEnvironment) initializeResourceAccessPolicyManager(ctx context.Context) error {
	d.Logger.Debugf("Initializing resource access policy manager")
	debugModeEnabled := d.Logger.GetLevel() == logrus.DebugLevel
	resourceAccessPolicyMgr, err := resourceaccess.NewResourceAccessPolicyManager(
		resourceaccess.CasbinResourceAccessPolicyManagerConfig{
			DBConnectionString:     d.Config.Database.ConnectionString(),
			EnableDebugLog:         debugModeEnabled,
			PolicyAutoLoadInterval: time.Minute,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policy manager: %w", err)
	}
	d.ResourceAccessPolicyManager = resourceAccessPolicyMgr
	return nil
}

func (d *developmentEnvironment) initializePermissionService(ctx context.Context) error {
	d.PermissionService = auth.NewPermissionService(auth.PermissionServiceSpec{
		PolicyManager: d.ResourceAccessPolicyManager,
		ProjectStore: pgstore.NewProjectStore(pgstore.ProjectStoreSpec{
			SessionFactory: d.DBSession,
		}),
		Logger: d.Logger,
	})
	return nil
}

func (d *developmentEnvironment) loadServices(ctx context.Context) error {
	d.Logger.Debugf("Initializing services")

	encryptionService, err := services.NewAESEncryptionService(services.EncryptionServiceSpec{
		Masterkey: d.Config.EncryptionKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create encryption service: %w", err)
	}
	d.EncryptionService = encryptionService
	stackDomainService := services.NewStackDomainsService(services.StackDomainsServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	organisationDomainService := services.NewOrganisationDomainsService(services.OrganisationDomainsServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	projectService := services.NewProjectService(services.ProjectServiceSpec{
		SessionFactory: d.DBSession,
		PolicyManager:  d.ResourceAccessPolicyManager,
		Permissions:    d.PermissionService,
		Logger:         d.Logger,
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		OrganisationDomainService: organisationDomainService,
		StackQueryService:         d.Services.StackService,
		SessionFactory:            d.DBSession,
		ProjectService:            projectService,
		PolicyManager:             d.ResourceAccessPolicyManager,
		Permissions:               d.PermissionService,
		Logger:                    d.Logger,
	})

	stackStore := pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: d.DBSession})

	referenceService := services.NewReferenceService(services.ReferenceServiceSpec{
		SessionFactory: d.DBSession,
		StackStore:     stackStore,
	})

	secretService := services.NewSecretService(services.SecretServiceSpec{
		SessionFactory:    d.DBSession,
		Logger:            d.Logger,
		EncryptionService: encryptionService,
		ProjectService:    projectService,
		Permissions:       d.PermissionService,
		ReferenceService:  referenceService,
	})

	registryCredentialService := services.NewRegistryCredentialService(services.RegistryCredentialServiceSpec{
		Store: pgstore.NewRegistryCredentialStore(pgstore.RegistryCredentialStoreSpec{
			SessionFactory: d.DBSession,
		}),
		StackStore:        stackStore,
		ReferenceService:  referenceService,
		EncryptionService: encryptionService,
		Permissions:       d.PermissionService,
		Logger:            d.Logger,
	})

	gitIntegrationService := services.NewGitIntegrationService(services.GitIntegrationServiceSpec{
		Store: pgstore.NewGitIntegrationStore(pgstore.GitIntegrationStoreSpec{
			SessionFactory: d.DBSession,
		}),
		InstallationStore: pgstore.NewGitInstallationStore(pgstore.GitInstallationStoreSpec{
			SessionFactory: d.DBSession,
		}),
		OAuthStateStore: pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
			SessionFactory: d.DBSession,
		}),
		OrganisationStore: pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
			SessionFactory: d.DBSession,
		}),
		AtomicExecutor: pgstore.NewAtomicExecutor(d.DBSession),
		GitHubAppClient: githubapp.NewClient(githubapp.ClientSpec{
			BaseURL: d.Config.GitHubAPIBaseURL,
		}),
		ExternalURL:       d.Config.ServerExternalURL,
		EncryptionService: encryptionService,
		Permissions:       d.PermissionService,
		Logger:            d.Logger,
	})

	credentialResolver := services.NewCredentialResolver(services.CredentialResolverSpec{
		RegistryCredentialService: registryCredentialService,
		GitIntegrationService:     gitIntegrationService,
	})

	d.RefreshTokenStore = pgstore.NewRefreshTokenStore(pgstore.RefreshTokenStoreSpec{
		SessionFactory: d.DBSession,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              d.DBSession,
		Logger:                      d.Logger,
		JwtSecretKey:                d.Config.JwtSecret,
		ResourceAccessPolicyManager: d.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
		Permissions:                 d.PermissionService,
		ProjectService:              projectService,
		RefreshTokenStore:           d.RefreshTokenStore,
	})

	imageRegistryService := services.NewClusterImageRegistryService(services.ImageRegistryServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		Permissions:    d.PermissionService,
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		ClusterManager:       d.ClusterManager,
		ImageRegistryService: imageRegistryService,
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
		Permissions:          d.PermissionService,
		EncryptionService:    encryptionService,
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		ClusterService: clusterService,
		UserService:    userService,
		Permissions:    d.PermissionService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory:   d.DBSession,
		Logger:           d.Logger,
		Permissions:      d.PermissionService,
		ReferenceService: referenceService,
	})

	resourceValidator := stackresourcevalidator.NewValidator(stackresourcevalidator.ValidatorSpec{
		Volumes: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: d.DBSession,
		}),
		Secrets: pgstore.NewSecretStore(pgstore.SecretStoreSpec{
			SessionFactory: d.DBSession,
		}),
		Domains:         organisationDomainService,
		Credentials:     credentialResolver,
		GitIntegrations: gitIntegrationService,
	})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:         d.DBSession,
		Logger:                 d.Logger,
		WorkspaceUserService:   workspaceUserService,
		Permissions:            d.PermissionService,
		StackStore:             stackStore,
		ClusterRegistryService: imageRegistryService,
		StackDomainService:     stackDomainService,
		ReferenceService:       referenceService,
		ResourceValidator:      resourceValidator,
	})

	imageBuildService := services.NewImageBuildService(services.ImageBuildServiceSpec{
		StackResourceService: stackResourceService,
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
		Permissions:          d.PermissionService,
		StackStore:           stackStore,
	})

	namespaceService := services.NewNamespaceService(services.NamespaceServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})
	loggingService := services.NewLoggingService(services.LoggingServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		Logger:               d.Logger,
	})

	objectStoreService := services.NewObjectStoreService(services.ObjectStoreServiceSpec{
		SessionFactory: d.DBSession,
		SecretService:  secretService,
		ProjectService: projectService,
		ClusterManager: d.ClusterManager,
		Logger:         d.Logger,
		Permissions:    d.PermissionService,
	})

	postgresBackupService := services.NewPostgresBackupService(services.PostgresBackupServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	postgresAddonService := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
		SessionFactory:        d.DBSession,
		NamespaceService:      namespaceService,
		ClusterService:        clusterService,
		SecretService:         secretService,
		PostgresBackupService: postgresBackupService,
		ObjectStoreService:    objectStoreService,
		ProjectService:        projectService,
		ClusterManager:        d.ClusterManager,
		Logger:                d.Logger,
		Permissions:           d.PermissionService,
		ReferenceService:      referenceService,
	})

	stackService := services.NewStackService(services.StackServiceSpec{
		SessionFactory:        d.DBSession,
		Logger:                d.Logger,
		VolumeService:         volumeService,
		OrganisationService:   organisationService,
		StackResourceService:  stackResourceService,
		ClusterService:        clusterService,
		NamespaceService:      namespaceService,
		SecretService:         secretService,
		PostgresAddonService:  postgresAddonService,
		ProjectService:        projectService,
		Permissions:           d.PermissionService,
		ReferenceService:      referenceService,
		CredentialResolver:    credentialResolver,
		GitIntegrationService: gitIntegrationService,
	})

	metricsService := services.NewMetricsService(services.MetricsServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		StackService:         stackService,
		Logger:               d.Logger,
	})

	apiTokenService := services.NewAPITokenService(services.APITokenServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	var emailSvc emailpkg.EmailService
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost != "" {
		emailSvc = emailpkg.NewSMTPEmailService(emailpkg.SMTPConfig{
			Host:        smtpHost,
			Port:        os.Getenv("SMTP_PORT"),
			Username:    os.Getenv("SMTP_USERNAME"),
			Password:    os.Getenv("SMTP_PASSWORD"),
			FromAddress: os.Getenv("SMTP_FROM_ADDRESS"),
			AppBaseURL:  os.Getenv("APP_BASE_URL"),
		})
		d.Logger.Infof("SMTP email service configured")
	} else {
		emailSvc = emailpkg.NewNoopEmailService(applogger.NewLoggerWithPrefix(ctx, "email-service").SetLevel(d.Logger.GetLevel()))
		d.Logger.Infof("SMTP not configured, using no-op email service")
	}
	d.EmailService = emailSvc

	orgInviteStore := pgstore.NewOrgInviteStore(pgstore.OrgInviteStoreSpec{
		SessionFactory: d.DBSession,
	})
	orgInviteService := services.NewOrgInviteService(services.OrgInviteServiceSpec{
		InviteStore:       orgInviteStore,
		ProjectService:    projectService,
		UserService:       userService,
		EncryptionService: encryptionService,
		Permissions:       d.PermissionService,
		Logger:            d.Logger,
	})

	signupService := services.NewSignupService(services.SignupServiceSpec{
		UserService:         userService,
		OrgInviteService:    orgInviteService,
		OrganisationService: organisationService,
		ProjectService:      projectService,
		PolicyManager:       d.ResourceAccessPolicyManager,
		RefreshTokenStore:   d.RefreshTokenStore,
		JWTSecretKey:        d.Config.JwtSecret,
		JWTClaimsBuilder:    auth.NewJWTClaimsBuilder(),
		Logger:              d.Logger,
	})

	d.OAuthStateStore = pgstore.NewOAuthStateStore(pgstore.OAuthStateStoreSpec{
		SessionFactory: d.DBSession,
	})

	stackReleaseStore := pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{
		SessionFactory: d.DBSession,
	})

	releaseEventStore := pgstore.NewReleaseEventStore(pgstore.ReleaseEventStoreSpec{
		SessionFactory: d.DBSession,
	})
	releaseEventRecorder := services.NewReleaseEventRecorder(services.ReleaseEventRecorderSpec{
		Store: releaseEventStore,
	})

	stackReleaseService := services.NewStackReleaseService(services.StackReleaseServiceSpec{
		Store:              stackReleaseStore,
		StackService:       stackService,
		CredentialResolver: credentialResolver,
		Permissions:        d.PermissionService,
		ReferenceService:   referenceService,
		EventStore:         releaseEventStore,
		EventRecorder:      releaseEventRecorder,
	})

	stackService.SetReleaseService(stackReleaseService)

	stackPreviewConfigStore := pgstore.NewStackPreviewConfigStore(pgstore.StackPreviewConfigStoreSpec{
		SessionFactory: d.DBSession,
	})
	previewStackStore := pgstore.NewPreviewStackStore(pgstore.PreviewStackStoreSpec{
		SessionFactory: d.DBSession,
	})

	stackPreviewConfigService := services.NewStackPreviewConfigService(services.StackPreviewConfigServiceSpec{
		Store:              stackPreviewConfigStore,
		PreviewStackStore:  previewStackStore,
		CredentialResolver: credentialResolver,
		Permissions:        d.PermissionService,
	})

	previewStackService := services.NewPreviewStackService(services.PreviewStackServiceSpec{
		Store:              previewStackStore,
		ConfigStore:        stackPreviewConfigStore,
		StackService:       stackService,
		ReleaseService:     stackReleaseService,
		SecretService:      secretService,
		CredentialResolver: credentialResolver,
		Permissions:        d.PermissionService,
	})

	d.Services = Services{
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

func (d *developmentEnvironment) injectClusterResourceServices(ctx context.Context) error {
	workspaceUserClusterResourceService := clusterresource.NewWorkspaceUserClusterResourceService(clusterresource.WorkspaceUserClusterResourceServiceSpec{
		ClusterManager: d.ClusterManager,
		Logger:         d.Logger,
		ClusterService: d.Services.ClusterService,
		UserService:    d.Services.UserService,
	})

	volumeClusterResourceService := clusterresource.NewVolumeClusterResourceService(clusterresource.VolumeClusterResourceServiceSpec{
		ClusterService:       d.Services.ClusterService,
		ClusterManager:       d.ClusterManager,
		Logger:               d.Logger,
		WorkspaceUserService: d.Services.WorkspaceUserService,
	})

	clusterImageRegistryService := clusterresource.NewClusterImageRegistryService(clusterresource.ClusterImageRegistryServiceSpec{
		ClusterManager: d.ClusterManager,
		Logger:         d.Logger,
		ClusterService: d.Services.ClusterService,
	})

	clusterNamespaceService := clusterresource.NewNamespaceClusterResourceService(clusterresource.NamespaceClusterResourceServiceSpec{
		ClusterManager: d.ClusterManager,
		Logger:         d.Logger,
		ClusterService: d.Services.ClusterService,
	})

	clusterLoggingService := clusterresource.NewLoggingService(clusterresource.LoggingServiceSpec{
		ClusterManager: d.ClusterManager,
		ClusterService: d.Services.ClusterService,
		Logger:         d.Logger,
	})

	clusterMetricsService := clusterresource.NewClusterMetricsService(clusterresource.ClusterMetricsServiceSpec{
		ClusterManager: d.ClusterManager,
		ClusterService: d.Services.ClusterService,
		Logger:         d.Logger,
	})

	deps := services.ClusterResourceServiceDeps{
		ClusterNamespaceService: clusterNamespaceService,
		ClusterVolumeService:    volumeClusterResourceService,
		ClusterLoggingService:   clusterLoggingService,
		ClusterMetricsService:   clusterMetricsService,
	}

	dep := services.BackgroundJobEnqueuerDep{
		BackgroundJobEnqueuer: d.WorkerManager,
	}

	d.Services.WorkspaceUserService.InjectClusterResourceService(workspaceUserClusterResourceService)
	d.Services.VolumeService.InjectClusterResourceService(volumeClusterResourceService)
	d.Services.StackService.InjectClusterResourceServiceDeps(deps)
	d.Services.NamespaceService.InjectClusterResourceServiceDeps(deps)
	d.Services.LoggingService.InjectClusterResourceServiceDeps(deps)
	d.Services.MetricsService.InjectClusterResourceServiceDeps(deps)
	d.Services.ClusterImageRegistryService.InjectClusterResourceService(clusterImageRegistryService)
	d.Services.StackService.InjectBackgroundJobEnqueuer(dep)
	d.Services.StackResourceService.InjectClusterManager(d.ClusterManager)
	d.Services.PostgresAddonService.InjectBackgroundJobEnqueuer(dep)
	d.Services.OrgInviteService.InjectBackgroundJobEnqueuer(dep)
	d.Services.StackReleaseService.InjectBackgroundJobEnqueuer(dep)
	d.Services.PreviewStackService.InjectBackgroundJobEnqueuer(dep)
	return nil
}

func (d *developmentEnvironment) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	d.Logger.Debugf("Initializing base resource access policies")
	if err := auth.LoadDefaultPolicies(d.ResourceAccessPolicyManager.AddPolicy); err != nil {
		return fmt.Errorf("failed to load default policies: %w", err)
	}
	d.Logger.Debugf("Base resource access policies initialized")
	return nil
}

func (d *developmentEnvironment) startManagers(ctx context.Context) error {
	d.Logger.Debugf("Starting cluster manager")
	// Add clusters to the manager when booting up.
	clusters, err := d.Services.ClusterService.InternalListAllClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to list clusters: %w", err)
	}
	for _, cluster := range clusters {
		d.Logger.Debugf("Adding cluster %s to cluster manager", cluster.ID)
		if err := d.ClusterManager.RegisterCluster(cluster); err != nil {
			return fmt.Errorf("failed to register cluster %s: %w", cluster.ID, err)
		}
	}

	d.ClusterManager.Start(ctx)

	d.Logger.Debugf("Starting worker manager")
	return d.WorkerManager.Start(ctx)
}

func (d *developmentEnvironment) Shutdown(ctx context.Context) error {
	d.Logger.Infof("Shutting down development environment")

	// Stop worker manager first
	if d.WorkerManager != nil {
		d.Logger.Debugf("Stopping worker manager")
		d.WorkerManager.Stop(false) // Don't drain the queue
	}

	// Stop cluster manager
	if d.ClusterManager != nil {
		d.Logger.Debugf("Stopping cluster manager")
		if err := d.ClusterManager.Stop(ctx); err != nil {
			d.Logger.Errorf("Failed to stop cluster manager: %v", err)
		}
	}

	// Close database connections
	if d.DBSession != nil {
		d.Logger.Debugf("Closing database connections")
		if err := d.DBSession.Close(); err != nil {
			d.Logger.Errorf("Failed to close database connections: %v", err)
			return err
		}
	}

	d.Logger.Infof("Development environment shutdown completed")
	return nil
}
