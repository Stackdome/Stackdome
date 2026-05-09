package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ashishmax31/stackdome-api-server/cmd/environment"
	"github.com/ashishmax31/stackdome-api-server/config/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	applogger "github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/golang/glog"
	gorillahandlers "github.com/gorilla/handlers"
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

	openapi, err := NewOpenAPIMiddleWare(openapi.OpenAPISpec)
	check(err, "Unable to create openapi spec middleware")
	mainRouter.Use(openapi.Middleware)

	// referring to the router as type http.Handler allows us to add middleware via more handlers
	var mainHandler http.Handler = mainRouter

	// Setup authentication handlers.
	// Currently this uses jwt as the default auth mechanism - Reads the jwt token from the authorization header and sets
	// the jwt payload in the request context.
	authProtected := setupAuthenticationMiddleWare(mainHandler, env)

	// removeTrailingSlash turns "/" into ""; restore it so http.ServeFileFS
	// sees a valid absolute path for the SPA root.
	mainHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			authProtected.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		mainRouter.ServeHTTP(w, r)
	})

	// Setup CORS
	// TODO: Update this to restrict origins in production
	// From a config file or environment variables.
	mainHandler = gorillahandlers.CORS(
		gorillahandlers.AllowedOrigins([]string{"*"}),
	)(mainHandler)

	mainHandler = removeTrailingSlash(mainHandler)
	mainHandler = injectLoggerMiddleware(mainHandler, env.Environment().Logger)

	s.httpServer = &http.Server{
		Addr:        env.Environment().Config.Server.BindAddress,
		Handler:     WithRequestTimeoutMiddleware(mainHandler, 180*time.Second),
		ReadTimeout: 120 * time.Second,
	}

	return s
}

func injectLoggerMiddleware(next http.Handler, logger applogger.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := applogger.AddLoggerToContext(r.Context(), logger)
		r = r.WithContext(ctx)
		logger.Infof("Received request: %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		logger.Infof("Response sent for: %s %s", r.Method, r.URL.Path)
	})
}

func setupAuthenticationMiddleWare(mainHandler http.Handler, env environment.EnvImpl) http.Handler {
	authenticationHandler := NewAuthSelectHandler(AuthSelectorHandlerSpec{
		MainHandler: mainHandler,
		PublicPaths: []string{
			"^/api/v1/user-signup",
			"^/api/v1/auth",
			"^/health",
		},
		// Set JWT authentication as the default handler.
		DefaultAuthHandler: auth.NewJwtAuthnHandler(mainHandler, auth.JWTAuthnHandlerSpec{
			JWTSecret:  []byte(env.Environment().Config.JwtSecret),
			UserGetter: env.Environment().Services.UserService,
		}),
	})

	// Add JWT cookie authentication
	authenticationHandler.Add(
		auth.NewJwtCookieAuthnHandler(mainHandler, auth.JWTCookieAuthnHandlerSpec{
			JWTSecret:  []byte(env.Environment().Config.JwtSecret),
			UserGetter: env.Environment().Services.UserService,
		}),
		auth.CanAuthenticateWithCookie,
	)

	// Add API token authentication
	apiTokenAuthnHandler := auth.NewAPITokenHandler(mainHandler, auth.ApiTokenAuthnHandlerSpec{
		TokenLookup: env.Environment().Services.APITokenService,
		UserGetter:  env.Environment().Services.UserService,
	})
	authenticationHandler.Add(apiTokenAuthnHandler, auth.CanAuthenticateWithAPIToken)

	return authenticationHandler
}

// tokenLookupAdapter adapts APITokenService to the auth.TokenLookup interface.
type tokenLookupAdapter struct {
	svc services.APITokenService
}

func (a *tokenLookupAdapter) ValidateToken(ctx context.Context, rawToken string) (*models.APIToken, error) {
	token, serr := a.svc.ValidateToken(ctx, rawToken)
	if serr != nil {
		return nil, serr
	}
	return token, nil
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
