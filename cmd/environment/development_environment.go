package environment

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/resourcebuild"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspace"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspaceresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspacestorage"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspaceuser"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspacevolume"
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
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
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
		d.initializeClients,
		d.loadServices,
		d.initializeClusterManager,
		d.injectClusterResourceServices,
		d.initializeDefaultOrgAndCluster,
		d.initializeResourceAccessPolicyManager,
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
	d.Logger.Infof("Database initialized successfully")
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
	d.Logger.Debugf("Initializing database session")
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
			workspacestorage.NewWorskspaceStorageReconciler(workspacestorage.WorskspaceStorageReconcilerSpec{
				Log:                     applogger.NewLoggerWithPrefix(ctx, "workspace-storage-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceStorageService: d.Services.WorkspaceStorageService,
				VolumeService:           d.Services.WorkspaceVolumeService,
				Env:                     d.Env.Name,
			}),
			workspacevolume.NewWorkspaceVolumeReconciler(workspacevolume.WorkspaceVolumeReconcilerSpec{
				Log:                     applogger.NewLoggerWithPrefix(ctx, "workspace-volume-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceStorageService: d.Services.WorkspaceStorageService,
				VolumeService:           d.Services.WorkspaceVolumeService,
				Env:                     d.Env.Name,
			}),
			workspaceuser.NewWorkspaceUserReconciler(workspaceuser.WorkspaceUserReconcilerSpec{
				Log:                  applogger.NewLoggerWithPrefix(ctx, "workspace-user-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceUserService: d.Services.WorkspaceUserService,
				ClusterService:       d.Services.ClusterService,
				Env:                  d.Env.Name,
			}),
			workspace.NewWorkspaceReconciler(workspace.WorkspaceReconcilerSpec{
				Log:              applogger.NewLoggerWithPrefix(ctx, "workspace-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceService: d.Services.WorkspaceService,
				Env:              d.Env.Name,
			}),
			workspaceresource.NewWorkspaceResourceReconciler(workspaceresource.WorkspaceResourceReconcilerSpec{
				Log:                      applogger.NewLoggerWithPrefix(ctx, "workspace-resource-controller").SetLevel(d.Logger.GetLevel()),
				WorkspaceService:         d.Services.WorkspaceService,
				WorkspaceResourceService: d.Services.WorkspaceResourceService,
				Env:                      d.Env.Name,
			}),
			resourcebuild.NewResourceBuildReconciler(resourcebuild.ResourceBuildReconcilerSpec{
				Log:                    applogger.NewLoggerWithPrefix(ctx, "resource-build-controller").SetLevel(d.Logger.GetLevel()),
				DBResourceBuildService: d.Services.WorkspaceResourceBuildService,
				DBResourceService:      d.Services.WorkspaceResourceService,
			}),
		},
	})
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
	userService := services.NewUserService(services.UserServiceSpec{
		SessionFactory:              d.DBSession,
		Logger:                      d.Logger,
		JwtSecretKey:                d.Config.JwtSecret,
		ResourceAccessPolicyManager: d.ResourceAccessPolicyManager,
		JWTClaimsBuilder:            auth.NewJWTClaimsBuilder(),
	})

	organisationService := services.NewOrganisationService(services.OrganisationServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	clusterService := services.NewClusterService(services.ClusterServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	workspaceUserService := services.NewWorkspaceUserService(services.WorkspaceUserServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
		ClusterService: clusterService,
		UserService:    userService,
	})

	workspaceStorageService := services.NewWorkspaceStorageService(services.WorkspaceStorageServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	workspaceVolumeService := services.NewVolumeService(services.VolumeServiceSpec{
		SessionFactory: d.DBSession,
		Logger:         d.Logger,
	})

	workspaceService := services.NewWorkspaceService(services.WorkspaceServiceSpec{
		SessionFactory:          d.DBSession,
		Logger:                  d.Logger,
		WorkspaceUserService:    workspaceUserService,
		WorkspaceStorageService: workspaceStorageService,
		OrganisationService:     organisationService,
	})

	workspaceResourceService := services.NewWorkspaceResourceService(services.WorkspaceResourceServiceSpec{
		SessionFactory:          d.DBSession,
		Logger:                  d.Logger,
		WorkspaceUserService:    workspaceUserService,
		WorkspaceStorageService: workspaceStorageService,
		WorkspaceService:        workspaceService,
	})

	workspaceResourceBuildService := services.NewResourceBuildService(services.ResourceBuildServiceSpec{
		WorkspaceResourceService: workspaceResourceService,
		SessionFactory:           d.DBSession,
		Logger:                   d.Logger,
	})

	d.Services = Services{
		UserService:                   userService,
		WorkspaceUserService:          workspaceUserService,
		OrganisationService:           organisationService,
		ClusterService:                clusterService,
		WorkspaceStorageService:       workspaceStorageService,
		WorkspaceVolumeService:        workspaceVolumeService,
		WorkspaceService:              workspaceService,
		WorkspaceResourceService:      workspaceResourceService,
		WorkspaceResourceBuildService: workspaceResourceBuildService,
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

	workspacestorageClusterResourceService := clusterresource.NewWorkspaceStorageClusterResourceService(clusterresource.WorkspaceStorageClusterResourceServiceSpec{
		ClusterManager:       d.ClusterManager,
		Logger:               d.Logger,
		ClusterService:       d.Services.ClusterService,
		WorkspaceUserService: d.Services.WorkspaceUserService,
	})

	clusterWorkspaceService := clusterresource.NewClusterWorkspaceService(clusterresource.ClusterWorkspaceServiceSpec{
		ClusterManager:      d.ClusterManager,
		OrganisationService: d.Services.OrganisationService,
		Logger:              d.Logger,
		ClusterService:      d.Services.ClusterService,
	})

	d.Services.WorkspaceUserService.InjectClusterResourceService(workspaceUserClusterResourceService)
	d.Services.WorkspaceStorageService.InjectClusterResourceService(workspacestorageClusterResourceService)
	d.Services.WorkspaceService.InjectClusterResourceService(clusterWorkspaceService)
	return nil
}

func (d *developmentEnvironment) initializeDefaultOrgAndCluster(ctx context.Context) error {
	d.Logger.Debugf("Initializing default organization and cluster")

	orgStore := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
		SessionFactory: d.DBSession,
	})

	defaultOrg, serr := orgStore.GetDefaultOrg(ctx)
	if serr != nil {
		return fmt.Errorf("failed to get default organization: %w", serr)
	}

	if err := d.initializeDefaultCluster(ctx, defaultOrg); err != nil {
		return fmt.Errorf("failed to initialize default cluster: %w", err)
	}
	return nil
}

func (d *developmentEnvironment) initializeDefaultCluster(ctx context.Context, defaultOrg *models.Organisation) error {
	clusterStore := pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
		SessionFactory: d.DBSession,
	})

	// TODO: Add cluster as a separate resource using a cluster CREATE API
	// TODO: Add org as a separate resource using a org CREATE API or seed a default org during migration.

	if _, err := clusterStore.GetDefaultCluster(ctx); err != nil {
		if err.Code == errors.ErrorNotFound {
			desiredCluster := &models.Cluster{
				OrganisationID: defaultOrg.ID,
				Name:           d.Config.ClusterConfig.Name,
				ClusterURL:     d.Config.ClusterConfig.ClusterURL,
				ClusterCAData:  string(d.Config.ClusterConfig.ClusterCAData),
				Token:          string(d.Config.ClusterConfig.Token),
				Default:        true,
			}
			if _, err := clusterStore.Create(ctx, desiredCluster); err != nil {
				return fmt.Errorf("failed to create default cluster: %w", err)
			}
		} else {
			return err
		}
	}

	// Temporary:
	cluster, err := clusterStore.GetDefaultCluster(ctx)
	if err != nil {
		return err
	}

	return d.ClusterManager.RegisterCluster(cluster)
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
				Organisation:   defaultOrg.Name,
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
	d.ClusterManager.Start(ctx)
	return nil
}

func (d *developmentEnvironment) initializeClients(ctx context.Context) error {
	clusterClient, err := initializeClusterClient(d.Config.ClusterConfig)
	if err != nil {
		return fmt.Errorf("failed to intialize cluster client: %w", err)
	}
	d.Clients = Clients{
		DefaultClusterClient: clusterClient,
	}
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

	if err := workspacev1alpha1.AddToScheme(scheme); err != nil {
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
