package environment

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspacestorage"
	"github.com/ashishmax31/stackdome-api-server/pkg/controllers/workspacevolume"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
	wprworker "github.com/ashishmax31/stackdome-api-server/pkg/workers/workspaceprovisionrequests"
	"github.com/google/uuid"

	"github.com/openshift-online/ocm-sdk-go/leadership"
	"github.com/spf13/pflag"
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
			Name:   "development",
			Config: config.NewApplicationConfig(),
		},
	}
}

func (d *developmentEnvironment) Environment() *Env {
	return d.Env
}

func (d *developmentEnvironment) AddFlags(flags *pflag.FlagSet) error {
	d.Config.AddFlags(flags)
	return setConfigDefaults(flags, d.defaultFlags())
}

func (d *developmentEnvironment) Init(ctx context.Context) error {
	d.Config.ReadEnvironmentVariables()
	if err := d.Config.ReadConfigFiles(); err != nil {
		return err
	}
	logger := logger.NewLogger(ctx)
	d.DBSession = db.NewSessionFactory(d.Config.Database)

	if err := d.initializeClients(ctx); err != nil {
		return err
	}

	uuid := uuid.New().String()

	leadershipFlag, err := leadership.NewFlag().
		Process(uuid).
		Name("stackdome-api-server").
		Handle(d.DBSession.DirectDB()).
		Logger(logger).Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to create leadership flag: %w", err)
	}

	d.WorkerManager = workermanager.NewWorkerManager(workermanager.WorkerManagerSpec{
		Environment: d.Env.Name,
	})

	d.Services = d.loadSevices(logger)

	d.ClusterManager = clustermanager.NewClusterManager(clustermanager.ClusterManagerConfig{
		LeadershipFlag: leadershipFlag,
		ControllersToRegister: []clustermanager.Controller{
			workspacestorage.NewWorskspaceStorageReconciler(workspacestorage.WorskspaceStorageReconcilerSpec{
				Log:                     logger,
				WorkspaceStorageService: d.Services.WorkspaceStorageService,
				VolumeService:           d.Services.WorkspaceVolumeService,
				Env:                     d.Env.Name,
			}),
			workspacevolume.NewWorkspaceVolumeReconciler(workspacevolume.WorkspaceVolumeReconcilerSpec{
				Log:                     logger,
				WorkspaceStorageService: d.Services.WorkspaceStorageService,
				VolumeService:           d.Services.WorkspaceVolumeService,
				Env:                     d.Env.Name,
			}),
		},
	})

	d.createAndRegisterWorkers()
	if err := d.WorkerManager.Start(ctx); err != nil {
		return err
	}
	d.ClusterManager.Start(ctx)

	if err := d.initializeDefaultOrgAndCluster(ctx); err != nil {
		return err
	}
	return nil
}

func (d *developmentEnvironment) createAndRegisterWorkers() {
	wprWorker := wprworker.NewWorkspaceProvisionRequestWorker(wprworker.WorkspaceProvisionRequestReconcileWorkerSpec{
		ClusterClient:       d.Clients.DefaultClusterClient,
		Env:                 d.Env.Name,
		WPRService:          d.Services.WorkspaceProvisionRequestService,
		UserService:         d.Services.UserService,
		OrganisationService: d.Services.OrganisationService,
		ClusterService:      d.Services.ClusterService,
	})
	d.WorkerManager.RegisterWorker(wprWorker, &models.WorkspaceProvisionRequest{})
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
		return nil, err
	}
	token, err := base64.StdEncoding.DecodeString(cfg.Token)
	if err != nil {
		return nil, err
	}

	_, err = certutil.NewPoolFromBytes(cadata)
	if err != nil {
		return nil, err
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

func (d *developmentEnvironment) initializeDefaultOrgAndCluster(ctx context.Context) error {
	clusterStore := pgstore.NewClusterStore(pgstore.ClusterStoreSpec{
		SessionFactory: d.DBSession,
	})

	orgStore := pgstore.NewOrganisationStore(pgstore.OrganisationStoreSpec{
		SessionFactory: d.DBSession,
	})

	if _, err := orgStore.GetDefaultOrg(ctx); err != nil {
		if err.Code == errors.ErrorNotFound {
			desiredOrg := &models.Organisation{
				Name:       d.Config.DefaultOrganisation.Name,
				DomainName: d.Config.DefaultOrganisation.DomainName,
				Default:    true,
			}
			if _, err := orgStore.Create(ctx, desiredOrg); err != nil {
				return fmt.Errorf("failed to create default org: %w", err)
			}

		} else {
			return err
		}
	}

	defaultOrg, err := orgStore.GetDefaultOrg(ctx)
	if err != nil {
		return err
	}

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

func (d *developmentEnvironment) InitDatabase(ctx context.Context) error {
	d.Config.ReadEnvironmentVariables()
	if err := d.Config.ReadDBConfigFiles(); err != nil {
		return fmt.Errorf("failed to read DB config file: %w", err)
	}
	d.DBSession = db.NewSessionFactory(d.Config.Database)
	return nil
}

func (d *developmentEnvironment) loadSevices(logger logger.Logger) Services {
	return Services{
		UserService: services.NewUserService(services.UserServiceSpec{
			SessionFactory: d.DBSession,
			Logger:         logger,
			JwtSecretKey:   d.Config.Server.JwtSecret,
		}),
		WorkspaceProvisionRequestService: services.NewWorkspaceProvisionRequestService(services.WorkspaceProvisionRequestServiceSpec{
			SessionFactory:        d.DBSession,
			Logger:                logger,
			ClusterClient:         d.Clients.DefaultClusterClient,
			BackgroundJobEnqueuer: d.WorkerManager,
		}),
		OrganisationService: services.NewOrganisationService(services.OrganisationServiceSpec{
			SessionFactory: d.DBSession,
			Logger:         logger,
		}),
		ClusterService: services.NewClusterService(services.ClusterServiceSpec{
			SessionFactory: d.DBSession,
			Logger:         logger,
		}),
		WorkspaceStorageService: services.NewWorkspaceStorageService(services.WorkspaceStorageServiceSpec{
			SessionFactory: d.DBSession,
			Logger:         logger,
		}),
	}
}

func (d *developmentEnvironment) defaultFlags() map[string]string {
	return map[string]string{
		"v":                      "10",
		"api-server-bindaddress": "0.0.0.0:8000",
	}
}
