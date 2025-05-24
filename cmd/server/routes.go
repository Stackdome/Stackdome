package server

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers"
	"github.com/gorilla/mux"
)

func (s apiServer) routes() *mux.Router {
	mainRouter := mux.NewRouter()

	mainRouter.NotFoundHandler = http.HandlerFunc(api.SendNotFound)
	services := s.environment.Environment().Services
	logger := s.environment.Environment().Logger

	authzClient := auth.NewAuthorizationHandler(auth.AuthorizationHanlderSpec{
		AuthorizerBackend: s.environment.Environment().ResourceAccessPolicyManager,
	})

	userHandler := handlers.NewUserServiceHandler(handlers.UserServiceHandlerSpec{
		UserService: services.UserService,
		AuthzClient: authzClient,
	})

	organizationHandler := handlers.NewOrganisationHandler(handlers.OrganisationHandlerSpec{
		OrganisationService: services.OrganisationService,
		AuthzClient:         authzClient,
	})

	workspaceUserHandler := handlers.NewWorkspaceUserHandler(handlers.WorkspaceUserHandlerSpec{
		WorkspaceUserService: services.WorkspaceUserService,
		AuthzClient:          authzClient,
	})

	volumeHandler := handlers.NewVolumeHandler(handlers.VolumeHandlerSpec{
		VolumeService: services.VolumeService,
		AuthzClient:   authzClient,
	})

	stackHandler := handlers.NewStackHandler(handlers.StackHandlerSpec{
		StackService:         services.StackService,
		StackResourceService: services.StackResourceService,
		ImageBuildService:    services.ImageBuildService,
		LoggingService:       services.LoggingService,
		AuthzClient:          authzClient,
		Logger:               logger,
	})

	stackResourceHandler := handlers.NewStackResourceHandler(handlers.StackResourceHandlerSpec{
		StackResourceService: services.StackResourceService,
		StackService:         services.StackService,
		LoggingService:       services.LoggingService,
		AuthzClient:          authzClient,
	})

	imageBuildHandler := handlers.NewImageBuildHandler(handlers.ImageBuildHandlerSpec{
		StackResourceService: services.StackResourceService,
		StackService:         services.StackService,
		ImageBuildService:    services.ImageBuildService,
		AuthzClient:          authzClient,
	})

	clusterHandler := handlers.NewClusterHandler(handlers.ClusterHandlerSpec{
		ClusterService: services.ClusterService,
		AuthzClient:    authzClient,
	})

	clusterImageRegistryHandler := handlers.NewClusterImageRegistryHandler(handlers.ClusterImageRegistryHandlerSpec{
		ClusterImageRegistryService: services.ClusterImageRegistryService,
		AuthzClient:                 authzClient,
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

	// Cluster routes
	clusterRouter := apiV1Router.PathPrefix("/organizations/{org_id}/clusters").Subrouter()
	clusterRouter.Use(authenticationMiddleware.AuthenticateUser)
	clusterRouter.HandleFunc("", clusterHandler.ListClustersForOrg).Methods(http.MethodGet)
	clusterRouter.HandleFunc("", clusterHandler.AddClusterForOrg).Methods(http.MethodPost)
	clusterRouter.HandleFunc("/{id}", clusterHandler.GetClusterForOrg).Methods(http.MethodGet)
	clusterRouter.HandleFunc("/{id}", clusterHandler.DeleteClusterForOrg).Methods(http.MethodDelete)

	// Cluster image registry routes
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

	workspaceUsersRouter := apiV1Router.PathPrefix("/workspace-users").Subrouter()
	workspaceUsersRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceUsersRouter.HandleFunc("", workspaceUserHandler.Create).Methods(http.MethodPost)
	workspaceUsersRouter.HandleFunc("/current", workspaceUserHandler.Current).Methods(http.MethodGet)
	workspaceUsersRouter.HandleFunc("/{id}", workspaceUserHandler.Get).Methods(http.MethodGet)
	workspaceUsersRouter.HandleFunc("/{id}", workspaceUserHandler.Update).Methods(http.MethodPut)
	workspaceUsersRouter.HandleFunc("/{id}", workspaceUserHandler.Delete).Methods(http.MethodDelete)

	stackStorageRouter := organizationsRouter.PathPrefix("/{org_id}/volumes").Subrouter()
	stackStorageRouter.Use(authenticationMiddleware.AuthenticateUser)
	stackStorageRouter.HandleFunc("", volumeHandler.Create).Methods(http.MethodPost)
	stackStorageRouter.HandleFunc("/current", volumeHandler.GetByID).Methods(http.MethodGet)
	stackStorageRouter.HandleFunc("/{id}", volumeHandler.GetByID).Methods(http.MethodGet)
	// workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Update).Methods(http.MethodPut)
	stackStorageRouter.HandleFunc("/{id}", volumeHandler.Delete).Methods(http.MethodDelete)

	stackRouter := organizationsRouter.PathPrefix("/{org_id}/stacks").Subrouter()
	stackRouter.Use(authenticationMiddleware.AuthenticateUser)
	stackRouter.HandleFunc("", stackHandler.Create).Methods(http.MethodPost)
	stackRouter.HandleFunc("", stackHandler.ListByOrganisationID).Methods(http.MethodGet)
	stackRouter.HandleFunc("/current", stackHandler.ListByUser).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}", stackHandler.GetByID).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}/logs", stackHandler.StreamLogs).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}", stackHandler.Update).Methods(http.MethodPut)
	stackRouter.HandleFunc("/{id}", stackHandler.Delete).Methods(http.MethodDelete)
	// Resources
	stackRouter.HandleFunc("/{id}/resources", stackResourceHandler.List).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}/resources/{resource_name}", stackResourceHandler.GetByResourceName).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}/resources/{resource_name}/logs", stackResourceHandler.StreamLogs).Methods(http.MethodGet)

	// Builds
	stackRouter.HandleFunc("/{id}/resources/{resource_name}/builds", imageBuildHandler.ListByResourceName).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}/builds", imageBuildHandler.ListByStackID).Methods(http.MethodGet)
	stackRouter.HandleFunc("/{id}/builds/{build_id}", imageBuildHandler.GetByID).Methods(http.MethodGet)

	return mainRouter
}
