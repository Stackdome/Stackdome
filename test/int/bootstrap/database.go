package bootstrap

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	_ "github.com/lib/pq"
)

type DatabaseManager struct {
	config         *DatabaseConfig
	adminDB        *sql.DB
	sessionFactory db.SessionFactory
}

type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

func NewDatabaseManager() *DatabaseManager {
	host := "localhost"
	if val, ok := config.EnvDBHost.Lookup(); ok {
		host = val
	}

	port := 5432
	if val, ok := config.EnvDBPort.Lookup(); ok {
		port = val
	}

	user := "postgres"
	if val, ok := config.EnvDBUsername.Lookup(); ok {
		user = val
	}

	password := "foobar-bizz-buzz"
	if val, ok := config.EnvDBPassword.Lookup(); ok {
		password = val
	}

	dbConfig := &DatabaseConfig{
		Host:     host,
		Port:     port,
		Name:     fmt.Sprintf("stackdome_test_%d", os.Getpid()),
		User:     user,
		Password: password,
	}

	return &DatabaseManager{
		config: dbConfig,
	}
}

func (dm *DatabaseManager) Bootstrap(ctx context.Context) error {
	// Connect to postgres database to create test database
	adminConnStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		dm.config.Host, dm.config.Port, dm.config.User, dm.config.Password)

	adminDB, err := sql.Open("postgres", adminConnStr)
	if err != nil {
		return fmt.Errorf("failed to connect to admin database: %w", err)
	}
	dm.adminDB = adminDB

	// Create test database
	_, err = adminDB.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", dm.config.Name))
	if err != nil {
		return fmt.Errorf("failed to create test database: %w", err)
	}

	// Create database config for test database
	testDBConfig := &config.DatabaseConfig{
		Dialect:            "postgres",
		SSLMode:            "disable",
		Debug:              false,
		MaxOpenConnections: 50,
		DBConnectionConfig: config.DBConnectionConfig{
			Host:     dm.config.Host,
			Port:     dm.config.Port,
			Name:     dm.config.Name,
			Username: dm.config.User,
			Password: dm.config.Password,
		},
	}

	// Initialize session factory
	sessionFactory := db.NewSessionFactory(testDBConfig)
	dm.sessionFactory = sessionFactory

	// Run migrations
	if err := dm.runMigrations(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (dm *DatabaseManager) runMigrations(ctx context.Context) error {
	session := dm.sessionFactory.New(ctx)
	db.Migrate(session)
	return nil
}

func (dm *DatabaseManager) GetSessionFactory() db.SessionFactory {
	return dm.sessionFactory
}

func (dm *DatabaseManager) GetConfig() *config.DatabaseConfig {
	return &config.DatabaseConfig{
		Dialect:            "postgres",
		SSLMode:            "disable",
		Debug:              false,
		MaxOpenConnections: 50,
		DBConnectionConfig: config.DBConnectionConfig{
			Host:     dm.config.Host,
			Port:     dm.config.Port,
			Name:     dm.config.Name,
			Username: dm.config.User,
			Password: dm.config.Password,
		},
	}
}

func (dm *DatabaseManager) ClearData(ctx context.Context) error {
	session := dm.sessionFactory.New(ctx)

	// List of tables to truncate (much faster than DELETE)
	// NOTE: users, organisations, clusters, and cluster_image_registries are NOT cleared
	// because they are created during bootstrap and should persist across all test specs
	tables := []string{
		"org_invites",
		"postgres_addon_databases",
		"postgres_backups",
		"postgres_addons",
		"object_stores",
		"stack_connections",
		"stack_domains",
		"stack_volumes",
		"stack_resources",
		"stack_storages",
		"stacks",
		"volumes",
		"resource_references",
		"secrets",
	}

	directDB, err := session.DB()
	if err != nil {
		return fmt.Errorf("failed to get direct DB connection: %w", err)
	}

	// TRUNCATE is much faster than DELETE and automatically resets sequences
	// CASCADE ensures dependent rows in other tables are also removed
	for _, table := range tables {
		if _, err := directDB.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			// Ignore "relation does not exist" errors as some tables might not exist in test
			continue
		}
	}

	return nil
}

func (dm *DatabaseManager) Cleanup(ctx context.Context) error {
	// Close connections to test database
	if dm.sessionFactory != nil {
		dm.sessionFactory.Close()
	}

	// Drop test database
	if dm.adminDB != nil {
		// Terminate all connections to the test database
		_, err := dm.adminDB.ExecContext(ctx, fmt.Sprintf(`
			SELECT pg_terminate_backend(pg_stat_activity.pid)
			FROM pg_stat_activity
			WHERE pg_stat_activity.datname = '%s'
			AND pid <> pg_backend_pid()`, dm.config.Name))
		if err != nil {
			return fmt.Errorf("failed to terminate connections: %w", err)
		}

		// Wait a bit for connections to close
		time.Sleep(100 * time.Millisecond)

		// Drop the database
		_, err = dm.adminDB.ExecContext(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dm.config.Name))
		if err != nil {
			return fmt.Errorf("failed to drop test database: %w", err)
		}

		dm.adminDB.Close()
	}

	return nil
}
