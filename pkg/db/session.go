package db

import (
	"context"
	"database/sql"

	"github.com/ashishmax31/stackdome-api-server/config"
	"gorm.io/gorm"
)

type SessionFactory interface {
	Init(*config.DatabaseConfig)
	DirectDB() *sql.DB
	New(ctx context.Context) *gorm.DB
	CheckConnection() error
	Close() error
}
