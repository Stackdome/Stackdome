package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ashishmax31/stackdome-api-server/config"
)

type Session struct {
	config *config.DatabaseConfig

	g2 *gorm.DB
	// Direct database connection.
	// It is used:
	// - to setup/close connection because GORM V2 removed gorm.Close()
	// - to work with pq.CopyIn because connection returned by GORM V2 gorm.DB() in "not the same"
	db *sql.DB
}

var _ SessionFactory = &Session{}

func NewSessionFactory(config *config.DatabaseConfig) *Session {
	conn := &Session{}
	conn.Init(config)
	return conn
}

// Init will initialize a singleton connection as needed and return the same instance.
// Go includes database connection pooling in the platform. Gorm uses the same and provides a method to
// clone a connection via New(), which is safe for use by concurrent Goroutines.
func (f *Session) Init(config *config.DatabaseConfig) {
	// Only the first time
	var (
		dbx *sql.DB
		g2  *gorm.DB
		err error
	)

	// Open connection to DB via standard library
	dbx, err = sql.Open(config.Dialect, config.ConnectionString(config.SSLMode != "disable"))
	if err != nil {
		dbx, err = sql.Open(config.Dialect, config.ConnectionString(false))
		if err != nil {
			panic(fmt.Sprintf(
				"SQL failed to connect to %s database %s with connection string: %s\nError: %s",
				config.Dialect,
				config.Name,
				config.LogSafeConnectionString(config.SSLMode != "disable"),
				err.Error(),
			))
		}
	}
	dbx.SetMaxOpenConns(config.MaxOpenConnections)

	// Connect GORM to use the same connection
	conf := &gorm.Config{
		PrepareStmt:          false,
		FullSaveAssociations: false,
	}
	g2, err = gorm.Open(postgres.New(postgres.Config{
		Conn: dbx,
		// Disable implicit prepared statement usage (GORM V2 uses pgx as database/sql driver and it enables prepared
		/// statement cache by default)
		// In migrations we both change tables' structure and running SQLs to modify data.
		// This way all prepared statements becomes invalid.
		PreferSimpleProtocol: true,
	}), conf)
	if err != nil {
		panic(fmt.Sprintf(
			"GORM failed to connect to %s database %s with connection string: %s\nError: %s",
			config.Dialect,
			config.Name,
			config.LogSafeConnectionString(config.SSLMode != "disable"),
			err.Error(),
		))
	}

	f.config = config
	f.g2 = g2
	f.db = dbx
}

func (f *Session) DirectDB() *sql.DB {
	return f.db
}

func (f *Session) New(ctx context.Context) *gorm.DB {
	conn := f.g2.Session(&gorm.Session{
		Context:              ctx,
		Logger:               f.g2.Logger.LogMode(logger.Silent),
		FullSaveAssociations: true,
	})
	if f.config.Debug {
		conn = conn.Debug()
	}
	return conn
}

func (f *Session) CheckConnection() error {
	return f.g2.Exec("SELECT 1").Error
}

func (f *Session) Close() error {
	return f.db.Close()
}
