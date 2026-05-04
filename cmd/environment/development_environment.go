package environment

import (
	"context"
	"fmt"
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
	postgresaddonworker "github.com/ashishmax31/stackdome-api-server/pkg/worker/postgresaddon"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/stack"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"

	volumecontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/volume"
	workspaceusercontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspaceuser"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	applogger "github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
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
			Name:            "development",
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
		Environment: d.Env.Name,
	})

	stackWorker := stack.NewStackWorker(stack.StackWorkerSpec{
		StackService:         d.Services.StackService,
		SecretService:        d.Services.SecretService,
		ClusterManager:       d.ClusterManager,
		VolumeService:        d.Services.VolumeService,
		NamespaceService:     d.Services.NamespaceService,
		PostgresAddonService: d.Services.PostgresAddonService,
		AddonUsageService:    d.Services.AddonUsageService,
		Env:                  d.Env.Name,
		CRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{
			SecretService: d.Services.SecretService,
		}),
		SecretBuilder: builders.NewSecretBuilder(builders.SecretBuilderSpec{
			SecretFetcher: d.Services.SecretService,
		}),
	})

	d.WorkerManager.RegisterWorker(stackWorker, &models.Stack{})

	pgAddonWorker := postgresaddonworker.NewPostgresAddonWorker(postgresaddonworker.PostgresAddonWorkerSpec{
		PostgresAddonService: d.Services.PostgresAddonService,
		ObjectStoreService:   d.Services.ObjectStoreService,
		NamespaceService:     d.Services.NamespaceService,
		SecretService:        d.Services.SecretService,
		AddonUsageStore:      d.Services.AddonUsageService,
		ClusterManager:       d.ClusterManager,
		CRBuilder:            builders.NewPostgresClusterBuilder(),
		Env:                  d.Env.Name,
	})
	d.WorkerManager.RegisterWorker(pgAddonWorker, &models.PostgresAddon{})

	return nil
}

func (d *developmentEnvironment) loadEnvAndConfigs(ctx context.Context) error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	d.Config.LoadEnvVariables()
	d.BootstrapConfig.LoadEnvVariables()

	if err := d.Config.Validate(); err != nil {
		return fmt.Errorf("invalid application config: %w", err)
	}

	if err := d.BootstrapConfig.Validate(); err != nil {
		return fmt.Errorf("invalid bootstrap config: %w", err)
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
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}
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

	d.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag: leadershipFlag,
		ControllersToRegister: []clustermanager.Controller{
			volumecontroller.NewVolumeReconciler(volumecontroller.VolumeReconcilerSpec{
				Log:            applogger.NewLoggerWithPrefix(ctx, "volume-controller").SetLevel(d.Logger.GetLevel()),
				StorageService: d.Services.StackStorageService,
				VolumeService:  d.Services.VolumeService,
				Env:            d.Env.Name,
			}),
			workspaceusercontroller.NewWorkspaceUserReconciler(workspaceusercontroller.WorkspaceUserReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "workspace-user-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceUserService: d.Services.WorkspaceUserService,
				ClusterService:       d.Services.ClusterService,
				Env:                  d.Env.Name,
			}),
			stackcontroller.NewStackReconciler(stackcontroller.StackReconcilerSpec{
				Log:          applogger.NewLoggerWithPrefix(ctx, "stack-controller").SetLevel(d.Logger.GetLevel()),
				StackService: d.Services.StackService,
				Env:          d.Env.Name,
			}),
			stackresourcecontroller.NewStackResourceReconciler(stackresourcecontroller.StackResourceReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "stack-resource-controller").SetLevel(d.Logger.GetLevel()),
				StackService:         d.Services.StackService,
				StackResourceService: d.Services.StackResourceService,
				Env:                  d.Env.Name,
			}),
			imagebuildcontroller.NewImageBuildReconciler(imagebuildcontroller.ImageBuildReconcilerSpec{
				Log:                 applogger.NewLoggerWithPrefix(ctx, "image-build-controller").SetLevel(d.Logger.GetLevel()),
				DBImageBuildService: d.Services.ImageBuildService,
				DBResourceService:   d.Services.StackResourceService,
			}),
			clusterimageregistry.NewClusterImageRegistryReconciler(clusterimageregistry.ClusterImageRegistryReconcilerSpec{
				Logger:                 applogger.NewLoggerWithPrefix(ctx, "cluster-image-registry-controller").SetLevel(d.Logger.GetLevel()),
				DBImageRegistryService: d.Services.ClusterImageRegistryService,
			}),
			postgresaddoncontroller.NewPostgresAddonReconciler(postgresaddoncontroller.PostgresAddonReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "postgres-addon-controller").SetLevel(d.Logger.GetLevel()),
				PostgresAddonService: d.Services.PostgresAddonService,
				Env:                  d.Env.Name,
			}),
			postgresbackupcontroller.NewPostgresBackupReconciler(postgresbackupcontroller.PostgresBackupReconcilerSpec{
				Log:                   applogger.NewLoggerWithPrefix(ctx, "postgres-backup-controller").SetLevel(d.Logger.GetLevel()),
				PostgresBackupService: d.Services.PostgresBackupService,
			}),
		},
	})
	d.Services.ClusterService.InjectClusterManager(d.ClusterManager)
	d.Services.PostgresAddonService.InjectClusterManager(d.ClusterManager)
	return nil
}

func (d *developmentEnvironment) initializeResourceAccessPolicyManager(ctx context.Context) error {
	d.Logger.Debugf("Initializing resource access policy manager")
	debugModeEnabled := d.Logger.GetLevel() == logrus.DebugLevel
	rootdir, err := findGoModDir()
	if err != nil {
		return fmt.Errorf("failed to find root directory for policy file: %w", err)
	}
	policyFilePath := filepath.Join(rootdir, "pkg/resourceaccess/casbin_model.conf")
	resourceAccessPolicyMgr, err := resourceaccess.NewResourceAccessPolicyManager(
		resourceaccess.CasbinResourceAccessPolicyManagerConfig{
			DBConnectionString:     d.Config.Database.ConnectionString(),
			EnableDebugLog:         debugModeEnabled,
			PolicyAutoLoadInterval: time.Minute,
			PolicyFilePath:         policyFilePath,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create policy manager: %w", err)
	}
	d.ResourceAccessPolicyManager = resourceAccessPolicyMgr
	return nil
}

func (d *developmentEnvironment) initializePermissionService(ctx context.Context) error {
	teamStore := pgstore.NewTeamStore(pgstore.TeamStoreSpec{
		SessionFactory: d.DBSession,
	})
	d.PermissionService = auth.NewPermissionService(auth.PermissionServiceConfig{
		PolicyManager: d.ResourceAccessPolicyManager,
		TeamStore:     teamStore,
		Logger:        d.Logger,
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
	secretService := services.NewSecretService(services.SecretServiceSpec{
		SessionFactory:    d.DBSession,
		Logger:            d.Logger,
		EncryptionService: encryptionService,
		Permissions:       d.PermissionService,
	})

	stackDomainService := services.NewStackDomainsService(services.StackDomainsServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	organisationDomainService := services.NewOrganisationDomainsService(services.OrganisationDomainsServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		OrganisationDomainService: organisationDomainService,
		StackQueryService:         d.Services.StackService,
		SessionFactory:            d.DBSession,
		Permissions:               d.PermissionService,
		Logger:                    d.Logger,
	})

	teamService := services.NewTeamService(services.TeamServiceSpec{
		SessionFactory: d.DBSession,
		PolicyManager:  d.ResourceAccessPolicyManager,
		Permissions:    d.PermissionService,
		Logger:         d.Logger,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              d.DBSession,
		Logger:                      d.Logger,
		JwtSecretKey:                d.Config.JwtSecret,
		ResourceAccessPolicyManager: d.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
		Permissions:                 d.PermissionService,
		TeamService:                 teamService,
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
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		ClusterService: clusterService,
		UserService:    userService,
		Permissions:    d.PermissionService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		Permissions:    d.PermissionService,
	})

	stackStore := pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: d.DBSession})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
		WorkspaceUserService: workspaceUserService,
		Permissions:          d.PermissionService,
		StackStore:           stackStore,
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
		ClusterManager: d.ClusterManager,
		Logger:         d.Logger,
		Permissions:    d.PermissionService,
	})

	postgresBackupService := services.NewPostgresBackupService(services.PostgresBackupServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	addonUsageService := services.NewAddonUsageService(services.AddonUsageServiceSpec{
		SessionFactory: d.DBSession,
	})

	postgresAddonService := services.NewPostgresAddonService(services.PostgresAddonServiceSpec{
		SessionFactory:        d.DBSession,
		NamespaceService:      namespaceService,
		ClusterService:        clusterService,
		SecretService:         secretService,
		PostgresBackupService: postgresBackupService,
		ObjectStoreService:    objectStoreService,
		ClusterManager:        d.ClusterManager,
		Logger:                d.Logger,
		Permissions:           d.PermissionService,
	})

	stackService := services.NewStackService(services.StackServiceSpec{
		SessionFactory:         d.DBSession,
		Logger:                 d.Logger,
		VolumeService:          volumeService,
		OrganisationService:    organisationService,
		StackResourceService:   stackResourceService,
		ClusterService:         clusterService,
		ClusterRegistryService: imageRegistryService,
		NamespaceService:       namespaceService,
		SecretService:          secretService,
		PostgresAddonService:   postgresAddonService,
		Permissions:            d.PermissionService,
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

	d.RefreshTokenStore = pgstore.NewRefreshTokenStore(pgstore.RefreshTokenStoreSpec{
		SessionFactory: d.DBSession,
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
		ObjectStoreService:          objectStoreService,
		PostgresAddonService:        postgresAddonService,
		PostgresBackupService:       postgresBackupService,
		AddonUsageService:           addonUsageService,
		APITokenService:             apiTokenService,
		TeamService:                 teamService,
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

	clusterStackService := clusterresource.NewClusterStackService(clusterresource.ClusterStackServiceSpec{
		ClusterManager:      d.ClusterManager,
		OrganisationService: d.Services.OrganisationService,
		Logger:              d.Logger,
		ClusterService:      d.Services.ClusterService,
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
		ClusterStackService:     clusterStackService,
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
	d.Services.PostgresAddonService.InjectBackgroundJobEnqueuer(dep)
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
