package server

import (
	"net/http"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/api"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers"
	"github.com/ashishmax31/stackdome-api-server/pkg/web"
	"github.com/gorilla/mux"
)

func (s apiServer) routes() *mux.Router {
	mainRouter := mux.NewRouter()

	mainRouter.NotFoundHandler = http.HandlerFunc(api.SendNotFound)

	// Health check endpoint
	mainRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods(http.MethodGet)
	services := s.environment.Environment().Services
	logger := s.environment.Environment().Logger

	userHandler := handlers.NewUserServiceHandler(handlers.UserServiceHandlerSpec{
		UserService: services.UserService,
	})

	organizationHandler := handlers.NewOrganisationHandler(handlers.OrganisationHandlerSpec{
		OrganisationService: services.OrganisationService,
	})

	teamHandler := handlers.NewTeamHandler(handlers.TeamHandlerSpec{
		TeamService: services.TeamService,
	})

	workspaceUserHandler := handlers.NewWorkspaceUserHandler(handlers.WorkspaceUserHandlerSpec{
		WorkspaceUserService: services.WorkspaceUserService,
		TeamService:          services.TeamService,
	})

	volumeHandler := handlers.NewVolumeHandler(handlers.VolumeHandlerSpec{
		VolumeService: services.VolumeService,
		TeamService:   services.TeamService,
	})

	stackHandler := handlers.NewStackHandler(handlers.StackHandlerSpec{
		StackService:         services.StackService,
		StackResourceService: services.StackResourceService,
		ImageBuildService:    services.ImageBuildService,
		LoggingService:       services.LoggingService,
		MetricsService:       services.MetricsService,
		TeamService:          services.TeamService,
		Logger:               logger,
	})

	stackResourceHandler := handlers.NewStackResourceHandler(handlers.StackResourceHandlerSpec{
		StackResourceService: services.StackResourceService,
		LoggingService:       services.LoggingService,
		MetricsService:       services.MetricsService,
		Logger:               logger,
	})

	imageBuildHandler := handlers.NewImageBuildHandler(handlers.ImageBuildHandlerSpec{
		ImageBuildService: services.ImageBuildService,
		Logger:            logger,
	})

	clusterHandler := handlers.NewClusterHandler(handlers.ClusterHandlerSpec{
		ClusterService: services.ClusterService,
	})

	clusterImageRegistryHandler := handlers.NewClusterImageRegistryHandler(handlers.ClusterImageRegistryHandlerSpec{
		ClusterImageRegistryService: services.ClusterImageRegistryService,
	})

	secretHandler := handlers.NewSecretHandler(handlers.SecretHandlerSpec{
		SecretService: services.SecretService,
		TeamService:   services.TeamService,
		Logger:        logger,
	})

	postgresAddonHandler := handlers.NewPostgresAddonHandler(handlers.PostgresAddonHandlerSpec{
		PostgresAddonService: services.PostgresAddonService,
		TeamService:          services.TeamService,
		Logger:               logger,
	})

	objectStoreHandler := handlers.NewObjectStoreHandler(handlers.ObjectStoreHandlerSpec{
		ObjectStoreService: services.ObjectStoreService,
		TeamService:        services.TeamService,
	})

	authenticationMiddleware := auth.NewAuthMiddleware(services.UserService)

	apiV1Router := mainRouter.PathPrefix("/api/v1").Subrouter()

	userSignupRouter := apiV1Router.PathPrefix("/user-signup").Subrouter()
	userSignupRouter.HandleFunc("", userHandler.Create).Methods(http.MethodPost)

	userRouter := apiV1Router.PathPrefix("/users").Subrouter()
	organizationsRouter := apiV1Router.PathPrefix("/organizations").Subrouter()
	organizationsRouter.Use(authenticationMiddleware.AuthenticateUser)
	organizationsRouter.HandleFunc("", organizationHandler.Create).Methods(http.MethodPost)
	organizationsRouter.HandleFunc("/default", organizationHandler.GetDefault).Methods(http.MethodGet)
	organizationsRouter.HandleFunc("/{id}", organizationHandler.GetByID).Methods(http.MethodGet)
	organizationsRouter.HandleFunc("/{id}", organizationHandler.Update).Methods(http.MethodPut)

	// Cluster routes (org-scoped)
	clusterRouter := apiV1Router.PathPrefix("/organizations/{org_id}/clusters").Subrouter()
	clusterRouter.Use(authenticationMiddleware.AuthenticateUser)
	clusterRouter.HandleFunc("", clusterHandler.ListClustersForOrg).Methods(http.MethodGet)
	clusterRouter.HandleFunc("", clusterHandler.AddClusterForOrg).Methods(http.MethodPost)
	clusterRouter.HandleFunc("/{id}", clusterHandler.GetClusterForOrg).Methods(http.MethodGet)
	clusterRouter.HandleFunc("/{id}", clusterHandler.DeleteClusterForOrg).Methods(http.MethodDelete)

	// Cluster image registry routes (org-scoped, nested under clusters)
	clusterRouter.HandleFunc("/{cluster_id}/image_registries", clusterImageRegistryHandler.ListRegistriesForCluster).Methods(http.MethodGet)
	clusterRouter.HandleFunc("/{cluster_id}/image_registries", clusterImageRegistryHandler.CreateRegistry).Methods(http.MethodPost)
	clusterRouter.HandleFunc("/{cluster_id}/image_registries/{id}", clusterImageRegistryHandler.GetRegistry).Methods(http.MethodGet)
	clusterRouter.HandleFunc("/{cluster_id}/image_registries/{id}", clusterImageRegistryHandler.DeleteRegistry).Methods(http.MethodDelete)

	authenticatedUserRouter := userRouter.NewRoute().Subrouter()
	authenticatedUserRouter.Use(authenticationMiddleware.AuthenticateUser)
	authenticatedUserRouter.HandleFunc("/current", userHandler.GetCurrentUser).Methods(http.MethodGet)
	authenticatedUserRouter.HandleFunc("/{id}", userHandler.Get).Methods(http.MethodGet)

	authenticationRouter := apiV1Router.PathPrefix("/auth").Subrouter()
	authenticationRouter.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	refreshHandler := auth.NewRefreshHandler(auth.RefreshHandlerSpec{
		RefreshTokenStore: s.environment.Environment().RefreshTokenStore,
		UserGetter:        services.UserService,
		JWTSecret:         []byte(s.environment.Environment().Config.JwtSecret),
		JWTClaimsBuilder:  auth.NewJWTClaimsBuilder(),
	})
	authenticationRouter.HandleFunc("/refresh", refreshHandler.HandleRefresh).Methods(http.MethodPost)

	if s.environment.Environment().Config.GitHubOAuth.Enabled() {
		oauthHandler := auth.NewGitHubOAuthHandler(auth.GitHubOAuthHandlerSpec{
			ClientID:          s.environment.Environment().Config.GitHubOAuth.ClientID,
			ClientSecret:      s.environment.Environment().Config.GitHubOAuth.ClientSecret,
			RedirectURI:       s.environment.Environment().Config.GitHubOAuth.RedirectURI,
			UserGetter:        services.UserService,
			UserCreator:       services.UserService,
			RefreshTokenStore: s.environment.Environment().RefreshTokenStore,
			JWTSecret:         []byte(s.environment.Environment().Config.JwtSecret),
			JWTClaimsBuilder:  auth.NewJWTClaimsBuilder(),
		})
		authenticationRouter.HandleFunc("/github", oauthHandler.HandleInitiate).Methods(http.MethodGet)
		authenticationRouter.HandleFunc("/github/callback", oauthHandler.HandleCallback).Methods(http.MethodGet)
	}

	apiTokenHandler := handlers.NewAPITokenHandler(handlers.APITokenHandlerSpec{
		APITokenService: services.APITokenService,
	})
	apiTokenRouter := apiV1Router.PathPrefix("/api-tokens").Subrouter()
	apiTokenRouter.Use(authenticationMiddleware.AuthenticateUser)
	apiTokenRouter.HandleFunc("", apiTokenHandler.Create).Methods(http.MethodPost)
	apiTokenRouter.HandleFunc("", apiTokenHandler.List).Methods(http.MethodGet)
	apiTokenRouter.HandleFunc("/{id}", apiTokenHandler.GetByID).Methods(http.MethodGet)
	apiTokenRouter.HandleFunc("/{id}", apiTokenHandler.Revoke).Methods(http.MethodDelete)

	// Team CRUD routes
	teamRouter := organizationsRouter.PathPrefix("/{org_id}/teams").Subrouter()
	teamRouter.HandleFunc("", teamHandler.Create).Methods(http.MethodPost)
	teamRouter.HandleFunc("", teamHandler.List).Methods(http.MethodGet)
	teamRouter.HandleFunc("/{team_name}", teamHandler.GetByName).Methods(http.MethodGet)
	teamRouter.HandleFunc("/{team_name}", teamHandler.Update).Methods(http.MethodPut)
	teamRouter.HandleFunc("/{team_name}", teamHandler.Delete).Methods(http.MethodDelete)

	// Team membership routes
	teamRouter.HandleFunc("/{team_name}/members", teamHandler.AddMember).Methods(http.MethodPost)
	teamRouter.HandleFunc("/{team_name}/members", teamHandler.ListMembers).Methods(http.MethodGet)
	teamRouter.HandleFunc("/{team_name}/members/{id}", teamHandler.UpdateMemberRole).Methods(http.MethodPut)
	teamRouter.HandleFunc("/{team_name}/members/{id}", teamHandler.RemoveMember).Methods(http.MethodDelete)

	// OrgAdmin management routes
	organizationsRouter.HandleFunc("/{org_id}/admins", teamHandler.PromoteToAdmin).Methods(http.MethodPost)
	organizationsRouter.HandleFunc("/{org_id}/admins", teamHandler.ListAdmins).Methods(http.MethodGet)
	organizationsRouter.HandleFunc("/{org_id}/admins/{user_id}", teamHandler.DemoteAdmin).Methods(http.MethodDelete)

	// Team-scoped resource routes
	teamResourceRouter := teamRouter.PathPrefix("/{team_name}").Subrouter()

	// Stacks (team-scoped)
	teamResourceRouter.HandleFunc("/stacks", stackHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/stacks", stackHandler.ListByTeamID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/current", stackHandler.ListByUser).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.Update).Methods(http.MethodPut)
	teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.Delete).Methods(http.MethodDelete)
	teamResourceRouter.HandleFunc("/stacks/{id}/logs", stackHandler.StreamLogs).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/metrics", stackHandler.GetMetrics).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/resources", stackResourceHandler.List).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}", stackResourceHandler.GetByResourceName).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/logs", stackResourceHandler.StreamLogs).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/metrics", stackResourceHandler.GetMetrics).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/builds", imageBuildHandler.ListByResourceName).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/builds", imageBuildHandler.ListByStackID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/stacks/{id}/builds/{build_id}", imageBuildHandler.GetByID).Methods(http.MethodGet)

	// Secrets (team-scoped)
	teamResourceRouter.HandleFunc("/secrets", secretHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/secrets", secretHandler.ListByTeamID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.Update).Methods(http.MethodPut)
	teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.Delete).Methods(http.MethodDelete)

	// Volumes (team-scoped)
	teamResourceRouter.HandleFunc("/volumes", volumeHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/volumes/current", volumeHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/volumes/{id}", volumeHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/volumes/{id}", volumeHandler.Delete).Methods(http.MethodDelete)

	// Postgres addons (team-scoped)
	teamResourceRouter.HandleFunc("/addons/postgres", postgresAddonHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/addons/postgres", postgresAddonHandler.List).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.Update).Methods(http.MethodPut)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.Delete).Methods(http.MethodDelete)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/backup", postgresAddonHandler.Backup).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/fence", postgresAddonHandler.Fence).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/hibernate", postgresAddonHandler.Hibernate).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}/backups", postgresAddonHandler.ListBackups).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/addons/postgres/{id}/credentials/{database}", postgresAddonHandler.GetCredentials).Methods(http.MethodGet)

	// Object stores (team-scoped)
	teamResourceRouter.HandleFunc("/object-stores", objectStoreHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/object-stores", objectStoreHandler.List).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.GetByID).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.Update).Methods(http.MethodPut)
	teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.Delete).Methods(http.MethodDelete)

	// Workspace users (team-scoped)
	teamResourceRouter.HandleFunc("/workspace-users", workspaceUserHandler.Create).Methods(http.MethodPost)
	teamResourceRouter.HandleFunc("/workspace-users/current", workspaceUserHandler.Current).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Get).Methods(http.MethodGet)
	teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Update).Methods(http.MethodPut)
	teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Delete).Methods(http.MethodDelete)

	// Exclude /api/ and /health so unknown paths return JSON 404 instead of index.html.
	mainRouter.PathPrefix("/").
		MatcherFunc(func(r *http.Request, _ *mux.RouteMatch) bool {
			p := r.URL.Path
			return !strings.HasPrefix(p, "/api/") && p != "/health"
		}).
		Handler(web.Handler())

	return mainRouter
}
