package server

import (
	"net/http"

	"github.com/ashishmax31/soradev-api-server/pkg/api"
	"github.com/ashishmax31/soradev-api-server/pkg/handlers"
	"github.com/gorilla/mux"
)

func (s apiServer) routes() *mux.Router {
	mainRouter := mux.NewRouter()
	mainRouter.NotFoundHandler = http.HandlerFunc(api.SendNotFound)
	services := s.environment.Environment().Services
	userHandler := handlers.NewUserServiceHandler(handlers.UserServiceHandlerSpec{
		UserService: services.UserService,
	})

	apiV1Router := mainRouter.PathPrefix("/api/v1").Subrouter()
	userRouter := apiV1Router.PathPrefix("/user").Subrouter()
	userRouter.HandleFunc("", userHandler.Create).Methods(http.MethodPost)
	userRouter.HandleFunc("/{id}", userHandler.Get).Methods(http.MethodGet)

	authenticationRouter := apiV1Router.PathPrefix("/auth").Subrouter()
	authenticationRouter.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)
	return mainRouter
}
