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
		AuthzClient:          authzClient,
		Logger:               logger,
	})

	stackResourceHandler := handlers.NewStackResourceHandler(handlers.StackResourceHandlerSpec{
		StackResourceService: services.StackResourceService,
		StackService:         services.StackService,
		AuthzClient:          authzClient,
	})

	imageBuildHandler := handlers.NewImageBuildHandler(handlers.ImageBuildHandlerSpec{
		StackResourceService: services.StackResourceService,
		StackService:         services.StackService,
		ImageBuildService:    services.ImageBuildService,
		AuthzClient:          authzClient,
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

	workspaceStorageRouter := organizationsRouter.PathPrefix("/{org_id}/volumes").Subrouter()
	workspaceStorageRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceStorageRouter.HandleFunc("", volumeHandler.Create).Methods(http.MethodPost)
	workspaceStorageRouter.HandleFunc("/current", volumeHandler.GetByID).Methods(http.MethodGet)
	workspaceStorageRouter.HandleFunc("/{id}", volumeHandler.GetByID).Methods(http.MethodGet)
	// workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Update).Methods(http.MethodPut)
	workspaceStorageRouter.HandleFunc("/{id}", volumeHandler.Delete).Methods(http.MethodDelete)

	workspaceRouter := organizationsRouter.PathPrefix("/{org_id}/stacks").Subrouter()
	workspaceRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceRouter.HandleFunc("", stackHandler.Create).Methods(http.MethodPost)
	workspaceRouter.HandleFunc("", stackHandler.ListByOrganisationID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/current", stackHandler.ListByUser).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}", stackHandler.GetByID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}", stackHandler.Update).Methods(http.MethodPut)
	workspaceRouter.HandleFunc("/{id}", stackHandler.Delete).Methods(http.MethodDelete)
	// Resources
	workspaceRouter.HandleFunc("/{id}/resources", stackResourceHandler.List).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/resources/{resource_name}", stackResourceHandler.GetByResourceName).Methods(http.MethodGet)

	// Builds
	workspaceRouter.HandleFunc("/{id}/resources/{resource_name}/builds", imageBuildHandler.ListByResourceName).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/builds", imageBuildHandler.ListByStackID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/builds/{build_id}", imageBuildHandler.GetByID).Methods(http.MethodGet)

	return mainRouter
}
