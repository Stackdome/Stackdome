package environment

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/clusterimageregistry"
	imagebuildcontroller "github.com/ashishmax31/stackdome-api-server/pkg/controllers/imagebuild"
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
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/openshift-online/ocm-sdk-go/leadership"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"sigs.k8s.io/controller-runtime/pkg/client"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
	usersv1alpha1 "stackdome.io/cluster-agent/api/users/v1alpha1"
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
		d.loadServices,
		d.initializeClusterManager,
		d.injectClusterResourceServices,
		// d.initializeDefaultOrgAndCluster,
		d.initializeBaseResourceAccessPolicies,
		d.ensureDefaultPlatformAdminUser,
		d.startClusterManager,
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
	d.Logger = applogger.NewLoggerWithPrefix(ctx, "applicationLogger").SetLevel(logLevel)
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
		},
	})
	d.Services.ClusterService.InjectClusterManager(d.ClusterManager)
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

func (d *developmentEnvironment) loadServices(ctx context.Context) error {
	d.Logger.Debugf("Initializing services")

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
		Logger:                    d.Logger,
	})

	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              d.DBSession,
		Logger:                      d.Logger,
		JwtSecretKey:                d.Config.JwtSecret,
		ResourceAccessPolicyManager: d.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
		OrganisationService:         organisationService,
	})

	imageRegistryService := services.NewClusterImageRegistryService(services.ImageRegistryServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		ClusterManager:       d.ClusterManager,
		ImageRegistryService: imageRegistryService,
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		ClusterService: clusterService,
		UserService:    userService,
	})

	volumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	stackResourceService := services.NewStackResourceService(services.StackResourceServiceSpec{
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
		WorkspaceUserService: workspaceUserService,
	})

	imageBuildService := services.NewImageBuildService(services.ImageBuildServiceSpec{
		StackResourceService: stackResourceService,
		SessionFactory:       d.DBSession,
		Logger:               d.Logger,
	})

	namespaceService := services.NewNamespaceService(services.NamespaceServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
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
	})

	loggingService := services.NewLoggingService(services.LoggingServiceSpec{
		ClusterService:       clusterService,
		StackResourceService: stackResourceService,
		Logger:               d.Logger,
	})

	d.Services = Services{
		UserService:                 userService,
		WorkspaceUserService:        workspaceUserService,
		OrganisationService:         organisationService,
		ClusterService:              clusterService,
		VolumeService:               volumeService,
		StackService:                stackService,
		StackResourceService:        stackResourceService,
		ImageBuildService:           imageBuildService,
		ClusterImageRegistryService: imageRegistryService,
		StackDomainService:          stackDomainService,
		OrganisationDomainService:   organisationDomainService,
		NamespaceService:            namespaceService,
		LoggingService:              loggingService,
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

	deps := services.ClusterResourceServiceDeps{
		ClusterStackService:     clusterStackService,
		ClusterNamespaceService: clusterNamespaceService,
		ClusterVolumeService:    volumeClusterResourceService,
		ClusterLoggingService:   clusterLoggingService,
	}

	d.Services.WorkspaceUserService.InjectClusterResourceService(workspaceUserClusterResourceService)
	d.Services.VolumeService.InjectClusterResourceService(volumeClusterResourceService)
	d.Services.StackService.InjectClusterResourceServiceDeps(deps)
	d.Services.NamespaceService.InjectClusterResourceServiceDeps(deps)
	d.Services.LoggingService.InjectClusterResourceServiceDeps(deps)
	d.Services.ClusterImageRegistryService.InjectClusterResourceService(clusterImageRegistryService)
	return nil
}

func (d *developmentEnvironment) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	d.Logger.Debugf("Initializing base resource access policies")

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
		if err := d.ResourceAccessPolicyManager.AddPolicy(
			policy.subject,
			policy.domain,
			policy.resource,
			policy.action,
			policy.resourceOwnerID,
		); err != nil {
			return fmt.Errorf("failed to add %s policy: %w", policy.subject, err)
		}
	}
	d.Logger.Debugf("Base resource access policies initialized")
	return nil
}

func (d *developmentEnvironment) ensureDefaultPlatformAdminUser(ctx context.Context) error {
	d.Logger.Debugf("Ensuring default platform admin user exists")

	defaultOrg, err := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
		SessionFactory: d.DBSession,
	}).GetDefaultOrg(ctx)
	if err != nil {
		return err
	}

	_, err = d.Services.UserService.GetDefaultUser(ctx)
	if err != nil {
		if err.Code == errors.ErrorNotFound {
			defaultUserInfo := d.BootstrapConfig.DefaultUser
			if validateErr := defaultUserInfo.Validate(); validateErr != nil {
				return fmt.Errorf("invalid default user config: %w", validateErr)
			}

			defaultUser := &models.User{
				Email:          defaultUserInfo.Email,
				Role:           models.PlatformAdminRole,
				Name:           defaultUserInfo.Name,
				Password:       defaultUserInfo.Password, // Assumes service handles hashing
				OrganisationID: defaultOrg.ID,
				DefaultUser:    true,
			}

			if _, createErr := d.Services.UserService.Create(ctx, defaultUser); createErr != nil {
				return fmt.Errorf("failed to create default user: %v", createErr)
			}
			d.Logger.Infof("Created default platform admin user")
			return nil
		}
		return fmt.Errorf("error checking for default user: %v", err)
	}
	return nil
}

func (d *developmentEnvironment) startClusterManager(ctx context.Context) error {
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
	return nil
}

func initializeClusterClient(cfg *config.ClusterConfig) (client.Client, error) {
	cadata, err := base64.StdEncoding.DecodeString(cfg.ClusterCAData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster CA data: %w", err)
	}
	token, err := base64.StdEncoding.DecodeString(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cluster token: %w", err)
	}

	_, err = certutil.NewPoolFromBytes(cadata)
	if err != nil {
		return nil, fmt.Errorf("failed to create cert pool: %w", err)
	}

	restConfig := &rest.Config{
		Host:        cfg.ClusterURL,
		BearerToken: string(token),
		TLSClientConfig: rest.TLSClientConfig{
			CAData: cadata,
		},
	}
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := corev1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := buildsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if err := usersv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	clientset, err := client.New(restConfig, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}
	return clientset, nil
}
