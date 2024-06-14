package db

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/db/migrations"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/golang/glog"
	"gorm.io/gorm"
)

func Migrate(g2 *gorm.DB) {
	m := gormigrate.New(g2, gormigrate.DefaultOptions, migrations.MigrationList)
	if err := m.Migrate(); err != nil {
		glog.Fatalf("Could not migrate: %v", err)
	}
}
