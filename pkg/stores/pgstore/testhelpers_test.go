package pgstore_test

import (
	"context"
	"database/sql"

	"github.com/ashishmax31/stackdome-api-server/config"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

type sqliteSessionFactory struct {
	db *gorm.DB
}

func newSQLiteSessionFactory(ddlStatements ...string) *sqliteSessionFactory {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	Expect(err).NotTo(HaveOccurred(), "failed to open sqlite db")

	for _, ddl := range ddlStatements {
		Expect(gdb.Exec(ddl).Error).NotTo(HaveOccurred(), "failed to execute DDL")
	}

	return &sqliteSessionFactory{db: gdb}
}

func (f *sqliteSessionFactory) Init(*config.DatabaseConfig) {}
func (f *sqliteSessionFactory) DirectDB() *sql.DB           { d, _ := f.db.DB(); return d }
func (f *sqliteSessionFactory) New(context.Context) *gorm.DB {
	return f.db.Session(&gorm.Session{})
}
func (f *sqliteSessionFactory) CheckConnection() error { return nil }
func (f *sqliteSessionFactory) Close() error           { d, _ := f.db.DB(); return d.Close() }
