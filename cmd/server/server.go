package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ashishmax31/soradev-api-server/cmd/environment"
	"github.com/ashishmax31/soradev-api-server/pkg/auth"
	"github.com/golang/glog"
)

type Server interface {
	Start()
	Stop() error
	Listen() (net.Listener, error)
	Serve(net.Listener)
}

type apiServer struct {
	httpServer  *http.Server
	environment environment.EnvImpl
}

var _ Server = &apiServer{}

func (a *apiServer) env() environment.EnvImpl {
	return a.environment
}

func NewAPIServer(env environment.EnvImpl) Server {
	s := &apiServer{environment: env}

	mainRouter := s.routes()

	// referring to the router as type http.Handler allows us to add middleware via more handlers
	var mainHandler http.Handler = mainRouter

	// Setup authentication handlers.
	mainHandler = setupAuthenticationMiddleWare(
		mainHandler,
		env,
	)

	mainHandler = removeTrailingSlash(mainHandler)

	s.httpServer = &http.Server{
		Addr:        env.Environment().Config.Server.BindAddress,
		Handler:     WithRequestTimeoutMiddleware(mainHandler, 180*time.Second),
		ReadTimeout: 120 * time.Second,
	}

	return s
}

func setupAuthenticationMiddleWare(mainHandler http.Handler, env environment.EnvImpl) http.Handler {
	authenticationHandler := NewAuthSelectHandler(AuthSelectorHandlerSpec{
		MainHandler: mainHandler,
		PublicPaths: []string{
			"^/api/v1/users",
			"^/api/v1/auth",
		},
		DefaultAuthHandler: auth.NewJwtAuthHandler(mainHandler, []byte(env.Environment().Config.Server.JwtSecret)),
	})

	return authenticationHandler
}

// Serve start the blocking call to Serve.
// Useful for breaking up ListenAndServer (Start) when you require the server to be listening before continuing
func (s apiServer) Serve(listener net.Listener) {
	var err error

	glog.Infof("Serving without TLS at %s", s.env().Environment().Config.Server.BindAddress)
	err = s.httpServer.Serve(listener)

	// Web server terminated.
	check(err, "Web server terminated with errors")
	glog.Info("Web server terminated")
}

// Listen only start the listener, not the server.
// Useful for breaking up ListenAndServer (Start) when you require the server to be listening before continuing
func (s apiServer) Listen() (listener net.Listener, err error) {
	return net.Listen("tcp", s.env().Environment().Config.Server.BindAddress)
}

// Start listening on the configured port and start the server. This is a convenience wrapper for Listen() and Serve(listener Listener)
func (s apiServer) Start() {
	listener, err := s.Listen()
	if err != nil {
		glog.Fatalf("Unable to start API server: %s", err)
	}
	s.Serve(listener)

	err = s.env().Environment().DBSession.Close()
	if err != nil {
		glog.Errorf("Cannot close all sql connections: %v", err)
	}
}

func (s apiServer) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

func WithRequestTimeoutMiddleware(next http.Handler, timeoutDuration time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancelFn := context.WithTimeout(r.Context(), timeoutDuration)
		defer cancelFn()
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func check(err error, msg string) {
	if err != nil && err != http.ErrServerClosed {
		glog.Errorf("%s: %s", msg, err)
		os.Exit(1)
	}
}
