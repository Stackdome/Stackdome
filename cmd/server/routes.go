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

	authenticationMiddleware := auth.NewAuthMiddleware(services.UserService)

	apiV1Router := mainRouter.PathPrefix("/api/v1").Subrouter()
	userRouter := apiV1Router.PathPrefix("/users").Subrouter()
	userRouter.HandleFunc("", userHandler.Create).Methods(http.MethodPost)
	userRouter.HandleFunc("/{id}", userHandler.Get).Methods(http.MethodGet)

	authenticationRouter := apiV1Router.PathPrefix("/auth").Subrouter()
	authenticationRouter.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	workspaceProvisionRequestRouter := apiV1Router.PathPrefix("/workspace-provision-requests").Subrouter()
	workspaceProvisionRequestRouter.Use(authenticationMiddleware.AuthenticateUser)
	workspaceProvisionRequestRouter.HandleFunc("", wprHandler.Create).Methods(http.MethodPost)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Get).Methods(http.MethodGet)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Update).Methods(http.MethodPut)
	workspaceProvisionRequestRouter.HandleFunc("/{id}", wprHandler.Delete).Methods(http.MethodDelete)
	return mainRouter
}
