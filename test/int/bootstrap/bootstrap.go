package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/testutil"
	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

type Environment struct {
	Client    *openapi.APIClient
	ClusterID string
	OrgID     string
	UserToken string
	Database  *DatabaseManager
	Cluster   *testutil.TestCluster

	// Internal managers
	clusterManager *ClusterManager
	serverManager  *ServerManager
	clientManager  *ClientManager
	logger         logr.Logger
}

func Setup(env *Environment, ctx context.Context) error {
	// Configure logger to output to stdout for visibility during test execution
	logger := stdr.New(log.New(os.Stderr, "", log.LstdFlags))
	logger.Info("Starting integration test bootstrap")

	env.logger = logger
	// Set up recovery to ensure cleanup on panic
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Errorf("panic during bootstrap: %v", r), "Bootstrap failed with panic, attempting cleanup")
			env.Cleanup()
			panic(r) // Re-panic after cleanup
		}
	}()

	// Initialize database
	logger.Info("Bootstrapping database")
	dbManager := NewDatabaseManager()
	env.Database = dbManager
	if err := dbManager.Bootstrap(ctx); err != nil {
		// env.Cleanup()
		return fmt.Errorf("database bootstrap failed: %v", err)
	}

	// Initialize cluster
	clusterManager := NewClusterManager()
	env.clusterManager = clusterManager
	if err := clusterManager.Bootstrap(ctx); err != nil {
		// env.Cleanup()
		return fmt.Errorf("cluster bootstrap failed: %v", err)
	}
	env.Cluster = clusterManager.GetCluster()

	// Initialize server
	logger.Info("Bootstrapping server")
	serverManager := NewServerManager(dbManager.GetSessionFactory(), dbManager.GetConfig(), logger)
	env.serverManager = serverManager
	if err := serverManager.Bootstrap(ctx, dbManager.GetConfig()); err != nil {
		// env.Cleanup()
		return fmt.Errorf("server bootstrap failed: %v", err)
	}

	// Initialize client and register cluster
	logger.Info("Bootstrapping client and registering cluster")
	clientManager := NewClientManager(serverManager.GetBaseURL(), clusterManager.GetCluster(), logger)
	env.clientManager = clientManager
	if err := clientManager.Bootstrap(ctx); err != nil {
		// env.Cleanup()
		return fmt.Errorf("client bootstrap failed: %v", err)
	}

	// Set final client details
	env.Client = clientManager.GetClient()
	env.ClusterID = clientManager.GetClusterID()
	env.OrgID = clientManager.GetOrgID()
	env.UserToken = clientManager.GetUserToken()

	logger.Info("Integration test bootstrap completed successfully",
		"orgID", env.OrgID,
		"clusterID", env.ClusterID,
		"serverURL", serverManager.GetBaseURL())

	return nil
}

func (env *Environment) Cleanup() {
	if env.logger.GetSink() == nil {
		// Logger not initialized yet, create a basic one with stdout output
		env.logger = stdr.New(log.New(os.Stdout, "", log.LstdFlags))
	}

	env.logger.Info("Starting integration test cleanup")

	// Create context with timeout for cleanup operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Only cleanup if KEEP_CLUSTER is not set
	if os.Getenv("KEEP_CLUSTER") != "true" {
		// Cleanup in reverse order of setup with error recovery
		if env.serverManager != nil {
			env.logger.Info("Cleaning up server")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during server cleanup: %v", r), "Server cleanup panicked")
					}
				}()
				if err := env.serverManager.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup server")
				}
			}()
		}

		if env.clusterManager != nil {
			env.logger.Info("Cleaning up cluster")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during cluster cleanup: %v", r), "Cluster cleanup panicked")
					}
				}()
				if err := env.clusterManager.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup cluster")
				}
			}()
		}

		if env.Database != nil {
			env.logger.Info("Cleaning up database")
			func() {
				defer func() {
					if r := recover(); r != nil {
						env.logger.Error(fmt.Errorf("panic during database cleanup: %v", r), "Database cleanup panicked")
					}
				}()
				if err := env.Database.Cleanup(ctx); err != nil {
					env.logger.Error(err, "Failed to cleanup database")
				}
			}()
		}
	} else {
		env.logger.Info("Skipping cleanup due to KEEP_CLUSTER=true")
		env.logger.Info("Manual cleanup required:",
			"cluster", env.clusterManager != nil && env.clusterManager.GetCluster() != nil,
			"database", env.Database != nil)
	}

	env.logger.Info("Integration test cleanup completed")
}

func (env *Environment) CreateAdditionalCluster() *testutil.TestCluster {
	config := testutil.DefaultClusterConfig("additional-test-cluster", env.logger)

	cluster := testutil.NewTestCluster(config)
	ctx := context.Background()

	if err := cluster.Setup(ctx); err != nil {
		panic(fmt.Sprintf("Failed to setup additional cluster: %v", err))
	}

	return cluster
}

func (env *Environment) RegisterAdditionalCluster(cluster *testutil.TestCluster) string {
	// Create a new client manager for the additional cluster
	clientManager := NewClientManager(env.serverManager.GetBaseURL(), cluster, env.logger)

	// Set existing authentication
	clientManager.userToken = env.UserToken
	clientManager.orgID = env.OrgID
	clientManager.configureAuthentication()

	ctx := context.Background()

	// Register the cluster
	clusterID, err := clientManager.registerCluster(ctx)
	if err != nil {
		panic(fmt.Sprintf("Failed to register additional cluster: %v", err))
	}

	return clusterID
}
