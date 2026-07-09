package pgstore_test

import (
	"context"
	"database/sql"
	stderrors "errors"

	"github.com/Stackdome/stackdome/config"
	sqlitedriver "github.com/glebarez/go-sqlite"
	"github.com/glebarez/sqlite"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

// SQLite extended result codes for constraint violations (SQLITE_CONSTRAINT_*).
const (
	sqliteConstraintPrimaryKey = 1555
	sqliteConstraintUnique     = 2067
)

// translatingSQLiteDialector adds gorm's ErrorTranslator to the sqlite
// dialector (glebarez/sqlite v1.7.0 does not implement it) so that unique
// constraint violations surface as gorm.ErrDuplicatedKey, matching the
// production session which opens gorm with TranslateError enabled.
type translatingSQLiteDialector struct {
	gorm.Dialector
}

func (d translatingSQLiteDialector) Translate(err error) error {
	var driverErr *sqlitedriver.Error
	if stderrors.As(err, &driverErr) {
		switch driverErr.Code() {
		case sqliteConstraintUnique, sqliteConstraintPrimaryKey:
			return gorm.ErrDuplicatedKey
		}
	}
	return err
}

type sqliteSessionFactory struct {
	db *gorm.DB
}

func newSQLiteSessionFactory(ddlStatements ...string) *sqliteSessionFactory {
	gdb, err := gorm.Open(
		translatingSQLiteDialector{Dialector: sqlite.Open(":memory:")},
		&gorm.Config{TranslateError: true},
	)
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
