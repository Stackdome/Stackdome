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
	userHandler := handlers.NewUserServiceHandler(handlers.UserServiceHandlerSpec{
		UserService: services.UserService,
	})

	workspaceUserHandler := handlers.NewWorkspaceUserHandler(handlers.WorkspaceUserHandlerSpec{
		WorkspaceUserService: services.WorkspaceUserService,
	})

	storageHandler := handlers.NewWorkspaceStorageHandler(handlers.WorkspaceStorageHandlerSpec{
		WorkspaceStorageService: services.WorkspaceStorageService,
	})

	workspaceHandler := handlers.NewWorkspaceHandler(handlers.WorkspaceHandlerSpec{
		WorkspaceService:              services.WorkspaceService,
		WorkspaceResourceService:      services.WorkspaceResourceService,
		WorkspaceResourceBuildService: services.WorkspaceResourceBuildService,
	})

	workspaceResourceHandler := handlers.NewWorkspaceResourceHandler(handlers.WorkspaceResourceHandlerSpec{
		WorkspaceResourceService: services.WorkspaceResourceService,
		WorkspaceService:         services.WorkspaceService,
	})

	workspaceResourceBuildHandler := handlers.NewWorkspaceResourceBuildHandler(handlers.WorkspaceResourceBuildHandlerSpec{
		WorkspaceResourceService:     services.WorkspaceResourceService,
		WorkspaceService:             services.WorkspaceService,
		WorkspaceResouceBuildService: services.WorkspaceResourceBuildService,
	})

	authenticationMiddleware := auth.NewAuthMiddleware(services.UserService)

	apiV1Router := mainRouter.PathPrefix("/api/v1").Subrouter()
	userRouter := apiV1Router.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("", userHandler.Create).Methods(http.MethodPost)
	organizationsRouter := apiV1Router.PathPrefix("/organizations").Subrouter()
	authenticatedUserRouter := userRouter.NewRoute().Subrouter()
	authenticatedUserRouter.Use(authenticationMiddleware.AuthenticateUser)
	authenticatedUserRouter.HandleFunc("/me", userHandler.GetCurrentUser).Methods(http.MethodGet)
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

	workspaceStorageRouter := organizationsRouter.PathPrefix("/{org_id}/workspace-storages").Subrouter()
	workspaceStorageRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceStorageRouter.HandleFunc("", storageHandler.Create).Methods(http.MethodPost)
	workspaceStorageRouter.HandleFunc("/current", storageHandler.ListByUser).Methods(http.MethodGet)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.GetByID).Methods(http.MethodGet)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Update).Methods(http.MethodPut)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Delete).Methods(http.MethodDelete)
	workspaceStorageRouter.HandleFunc("/{id}/volumes", storageHandler.ListVolumes).Methods(http.MethodGet)

	workspaceRouter := organizationsRouter.PathPrefix("/{org_id}/workspaces").Subrouter()
	workspaceRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceRouter.HandleFunc("", workspaceHandler.Create).Methods(http.MethodPost)
	workspaceRouter.HandleFunc("", workspaceHandler.ListByOrganisationID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/current", workspaceHandler.ListByUser).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}", workspaceHandler.GetByID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}", workspaceHandler.Update).Methods(http.MethodPut)
	workspaceRouter.HandleFunc("/{id}", workspaceHandler.Delete).Methods(http.MethodDelete)
	// Resources
	workspaceRouter.HandleFunc("/{id}/resources", workspaceResourceHandler.List).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/resources/{resource_name}", workspaceResourceHandler.GetByResourceName).Methods(http.MethodGet)

	// Builds
	workspaceRouter.HandleFunc("/{id}/resources/{resource_name}/builds", workspaceResourceBuildHandler.ListByResourceName).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/builds", workspaceResourceBuildHandler.ListByWorkspaceID).Methods(http.MethodGet)
	workspaceRouter.HandleFunc("/{id}/builds/{build_id}", workspaceResourceBuildHandler.GetByID).Methods(http.MethodGet)

	return mainRouter
}
