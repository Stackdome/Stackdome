package environment

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ashishmax31/soradev-api-server/config"
	"github.com/ashishmax31/soradev-api-server/pkg/db"
	"github.com/ashishmax31/soradev-api-server/pkg/errors"
	"github.com/ashishmax31/soradev-api-server/pkg/logger"
	"github.com/ashishmax31/soradev-api-server/pkg/models"
	"github.com/ashishmax31/soradev-api-server/pkg/services"
	"github.com/ashishmax31/soradev-api-server/pkg/stores/pgstore"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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

	if err := d.initializeDefaultOrgAndCluster(ctx); err != nil {
		return err
	}
	if err := d.initializeClients(ctx); err != nil {
		return err
	}
	d.Services = d.loadSevices(logger)
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
		return nil, err
	}
	clientCert, err := base64.StdEncoding.DecodeString(cfg.ClientCertData)
	if err != nil {
		return nil, err
	}
	clientKey, err := base64.StdEncoding.DecodeString(cfg.ClientKeyData)
	if err != nil {
		return nil, err
	}
	restConfig := &rest.Config{
		Host: cfg.ClusterURL,
		TLSClientConfig: rest.TLSClientConfig{
			CAData:   cadata,
			CertData: clientCert,
			KeyData:  clientKey,
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
				ClientCertData: string(d.Config.ClusterConfig.ClientCertData),
				ClientKeyData:  string(d.Config.ClusterConfig.ClientKeyData),
				Default:        true,
			}
			if _, err := clusterStore.Create(ctx, desiredCluster); err != nil {
				return fmt.Errorf("failed to create default cluster: %w", err)
			}
		} else {
			return err
		}
	}

	return nil
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
			SessionFactory: d.DBSession,
			Logger:         logger,
			ClusterClient:  d.Clients.DefaultClusterClient,
		}),
	}
}

func (d *developmentEnvironment) defaultFlags() map[string]string {
	return map[string]string{
		"v":                      "10",
		"api-server-hostname":    "localhost",
		"api-server-bindaddress": "localhost:8000",
	}
}
