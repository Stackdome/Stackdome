package migrations

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

type Namespace struct {
	ID             string `gorm:"primaryKey;default:gen_random_uuid()"`
	Name           string `gorm:"not null;unique"`
	OrganisationID string `gorm:"not null"`
	Labels         []byte `gorm:"type:jsonb"`
	Annotations    []byte `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func createNamespaceTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202505161800_create_namespace_table",
		Migrate: func(tx *gorm.DB) error {
			return tx.Migrator().CreateTable(&Namespace{})
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable("namespaces")
		},
	}
}
