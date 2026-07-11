package db

import (
	"github.com/Stackdome/stackdome/pkg/db/migrations"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var log = logger.NewLogger()

func Migrate(g2 *gorm.DB) {
	options := gormigrate.DefaultOptions
	options.UseTransaction = true
	m := gormigrate.New(g2, options, migrations.MigrationList)
	if err := m.Migrate(); err != nil {
		log.Fatalf("Could not migrate: %v", err)
	}
}
