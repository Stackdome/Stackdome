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

	wprHandler := handlers.NewWorkspaceProvisionRequestServiceHandler(handlers.WorkspaceProvisionRequestServiceHandlerSpec{
		WorkspaceProvisionRequestService: services.WorkspaceProvisionRequestService,
	})

	storageHandler := handlers.NewWorkspaceStorageHandler(handlers.WorkspaceStorageHandlerSpec{
		WorkspaceStorageService: services.WorkspaceStorageService,
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

	workspaceProvisionRequestRouter := apiV1Router.PathPrefix("/workspace-provision-requests").Subrouter()
	workspaceProvisionRequestRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceProvisionRequestRouter.HandleFunc("", wprHandler.Create).Methods(http.MethodPost)
	workspaceProvisionRequestRouter.HandleFunc("/current", wprHandler.Current).Methods(http.MethodGet)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Get).Methods(http.MethodGet)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Update).Methods(http.MethodPut)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Delete).Methods(http.MethodDelete)

	workspaceStorageRouter := organizationsRouter.PathPrefix("/{org_id}/workspace-storages").Subrouter()
	workspaceStorageRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceStorageRouter.HandleFunc("", storageHandler.Create).Methods(http.MethodPost)
	workspaceStorageRouter.HandleFunc("/current", storageHandler.ListByUser).Methods(http.MethodGet)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.GetByID).Methods(http.MethodGet)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Update).Methods(http.MethodPut)
	workspaceStorageRouter.HandleFunc("/{id}", storageHandler.Delete).Methods(http.MethodDelete)
	return mainRouter
}
