package environment

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/clustermanager"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	applogger "github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/worker/workermanager"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvImpl interface {
	Init(context.Context) error
	InitDatabase(context.Context) error
	Environment() *Env
	Shutdown(context.Context) error
}

type Env struct {
	Name                        string
	Services                    Services
	DBSession                   db.SessionFactory
	Config                      *config.ApplicationConfig
	BootstrapConfig             *config.BootstrapConfig
	Clients                     Clients
	ClusterManager              clustermanager.ClusterManager
	WorkerManager               workermanager.WorkerManager
	ResourceAccessPolicyManager resourceaccess.ResourceAccessPolicyManager
	PermissionService           auth.PermissionService
	RefreshTokenStore           stores.RefreshTokenStore
	OAuthStateStore             stores.OAuthStateStore
	Logger                      applogger.Logger
}

type Clients struct {
	DefaultClusterClient client.Client
}

type Database struct {
	SessionFactory db.SessionFactory
}

type Services struct {
	UserService                 services.UserService
	WorkspaceUserService        services.WorkspaceUserService
	OrganisationService         services.OrganisationService
	ClusterService              services.ClusterService
	StackStorageService         services.StackStorageService
	VolumeService               services.VolumeService
	StackService                services.StackService
	StackResourceService        services.StackResourceService
	ImageBuildService           services.ImageBuildService
	ClusterImageRegistryService services.ImageRegistryService
	StackDomainService          services.StackDomainsService
	OrganisationDomainService   services.OrganisationDomainsService
	NamespaceService            services.NamespaceService
	LoggingService              services.LoggingService
	MetricsService              services.MetricsService
	SecretService               services.SecretService
	EncryptionService           services.EncryptionService
	ObjectStoreService          services.ObjectStoreService
	PostgresAddonService        services.PostgresAddonService
	PostgresBackupService       services.PostgresBackupService
	AddonUsageService           services.AddonUsageService
	APITokenService             services.APITokenService
	TeamService                 services.TeamService
}
