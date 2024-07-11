package environment

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
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
	Name          string
	Services      Services
	DBSession     db.SessionFactory
	Config        *config.ApplicationConfig
	Clients       Clients
	WorkerManager workermanager.WorkerManager
}

type Clients struct {
	DefaultClusterClient client.Client
}

type Database struct {
	SessionFactory db.SessionFactory
}

type Services struct {
	UserService                      services.UserService
	WorkspaceProvisionRequestService services.WorkspaceProvisionRequestService
	OrganisationService              services.OrganisationService
	ClusterService                   services.ClusterService
	WorkspaceStorageService          services.WorkspaceStorageService
}
