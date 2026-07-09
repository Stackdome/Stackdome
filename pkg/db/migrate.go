package db

import (
	"github.com/Stackdome/stackdome/pkg/db/migrations"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

func Migrate(g2 *gorm.DB) {
	options := gormigrate.DefaultOptions
	options.UseTransaction = true
	m := gormigrate.New(g2, options, migrations.MigrationList)
	if err := m.Migrate(); err != nil {
		glog.Fatalf("Could not migrate: %v", err)
	}
}
