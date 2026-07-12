package migrations

import (
	"fmt"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createProjectMembershipsTable() *gormigrate.Migration {
	type ProjectMembership struct {
		ID        string    `gorm:"primary_key;default:gen_random_uuid()"`
		ProjectID string    `gorm:"not null;uniqueIndex:idx_project_memberships_project_user"`
		UserID    string    `gorm:"not null;uniqueIndex:idx_project_memberships_project_user"`
		Role      string    `gorm:"not null"`
		CreatedAt time.Time `gorm:"not null;default:now()"`
	}
	return &gormigrate.Migration{
		ID: "202605040002_create_project_memberships_table",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Migrator().AutoMigrate(&ProjectMembership{}); err != nil {
				return fmt.Errorf("error creating project_memberships table: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_project_memberships_project_id ON project_memberships(project_id)").Error; err != nil {
				return fmt.Errorf("error creating project_memberships project_id index: %w", err)
			}
			if err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_project_memberships_user_id ON project_memberships(user_id)").Error; err != nil {
				return fmt.Errorf("error creating project_memberships user_id index: %w", err)
			}
			if err := tx.Exec("ALTER TABLE project_memberships ADD CONSTRAINT fk_project_memberships_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding project_memberships project foreign key: %w", err)
			}
			if err := tx.Exec("ALTER TABLE project_memberships ADD CONSTRAINT fk_project_memberships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE").Error; err != nil {
				return fmt.Errorf("error adding project_memberships user foreign key: %w", err)
			}
			return nil
		},
	}
}
