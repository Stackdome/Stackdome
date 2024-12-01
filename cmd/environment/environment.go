package environment

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvImpl interface {
	Init(context.Context) error
	InitDatabase(context.Context) error
	AddFlags(flags *pflag.FlagSet) error
	Environment() *Env
}

type Env struct {
	Name           string
	Services       Services
	DBSession      db.SessionFactory
	Config         *config.ApplicationConfig
	Clients        Clients
	ClusterManager clustermanager.ClusterManager
}

type Clients struct {
	DefaultClusterClient client.Client
}

type Database struct {
	SessionFactory db.SessionFactory
}

type Services struct {
	UserService                   services.UserService
	WorkspaceUserService          services.WorkspaceUserService
	OrganisationService           services.OrganisationService
	ClusterService                services.ClusterService
	WorkspaceStorageService       services.WorkspaceStorageService
	WorkspaceVolumeService        services.VolumeService
	WorkspaceService              services.WorkspaceService
	WorkspaceResourceService      services.WorkspaceResourceService
	WorkspaceResourceBuildService services.ResourceBuildService
}
