package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Stackdome/stackdome/cmd/environment"
	"github.com/Stackdome/stackdome/cmd/server"
	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/go-logr/logr"
)

const (
	ServerPort = 8987
)

type ServerManager struct {
	config         *config.ApplicationConfig
	env            environment.EnvImpl
	apiServer      server.Server
	baseURL        string
	port           int
	sessionFactory db.SessionFactory
	logger         logr.Logger
}

func NewServerManager(sessionFactory db.SessionFactory, dbConfig *config.DatabaseConfig, logger logr.Logger) *ServerManager {
	return &ServerManager{
		config: &config.ApplicationConfig{
			Server: &config.ServerConfig{
				Hostname:    "localhost",
				BindAddress: fmt.Sprintf("0.0.0.0:%d", ServerPort),
			},
			Database:    dbConfig,
			GitHubOAuth: &config.GitHubOAuthConfig{},
		},
		port:           ServerPort,
		baseURL:        fmt.Sprintf("http://localhost:%d", ServerPort),
		sessionFactory: sessionFactory,
		logger:         logger,
	}
}

func (sm *ServerManager) Bootstrap(ctx context.Context) error {
	sm.logger.Info("Starting server bootstrap")

	// Create test environment
	env := environment.NewTestEnvironment(sm.sessionFactory, sm.config.Database, environment.WithApplicationConfig(sm.config))

	// Use a background context for environment init. The environment starts
	// long-lived goroutines (worker manager, cluster manager) that must outlive
	// the bootstrap phase. Using a timeout context here would cancel those
	// goroutines when bootstrap returns.
	if err := env.Init(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize environment: %w", err)
	}

	sm.env = env

	// Create and start server
	sm.apiServer = server.NewAPIServer(env)

	// Start server in background using Listen + Serve pattern
	listener, err := sm.apiServer.Listen()
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}

	go func() {
		sm.logger.Info("Starting API server", "address", sm.config.Server.BindAddress)
		sm.apiServer.Serve(listener)
	}()
	sm.logger.Info("API server started", "address", sm.config.Server.BindAddress)
	// Wait for server to be ready

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	if err := sm.waitForReady(ctxWithTimeout); err != nil {
		return fmt.Errorf("server failed to start: %w", err)
	}

	sm.logger.Info("Server bootstrap completed successfully", "baseURL", sm.baseURL)
	return nil
}

func (sm *ServerManager) GetBaseURL() string {
	return sm.baseURL
}

func (sm *ServerManager) GetEnvironment() environment.EnvImpl {
	return sm.env
}

func (sm *ServerManager) waitForReady(ctx context.Context) error {
	healthURL := fmt.Sprintf("%s/health", sm.baseURL)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func (sm *ServerManager) Cleanup(ctx context.Context) error {
	if sm.apiServer != nil {
		sm.logger.Info("Shutting down API server")
		if err := sm.apiServer.Stop(); err != nil {
			sm.logger.Error(err, "Failed to stop API server")
			return fmt.Errorf("failed to stop server: %w", err)
		}
		sm.logger.Info("Server shutdown completed")
	}

	// Shutdown the environment to cleanup resources
	if sm.env != nil {
		sm.logger.Info("Shutting down test environment")
		if err := sm.env.Shutdown(ctx); err != nil {
			sm.logger.Error(err, "Failed to shutdown environment")
			return fmt.Errorf("failed to shutdown environment: %w", err)
		}
	}

	return nil
}
