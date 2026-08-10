package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const (
	dropWorkspaceUsersMigrationID = "202608100003_drop_workspace_users"
	workspaceNamespacesTableName  = "workspace_namespaces"
	workspaceUsersTableName       = "workspace_users"
	casbinRuleTableName           = "casbin_rule"
	casbinResourceColumnName      = "v2"
	workspaceUserResourceName     = "workspace-users"
	workspaceUserResourcePattern  = workspaceUserResourceName + "/%"
)

func dropWorkspaceUserTables() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: dropWorkspaceUsersMigrationID,
		Migrate: func(tx *gorm.DB) error {
			if tx.Migrator().HasTable(workspaceUsersTableName) {
				var workspaceUserCount int64
				if err := tx.Table(workspaceUsersTableName).Count(&workspaceUserCount).Error; err != nil {
					return fmt.Errorf("count %s: %w", workspaceUsersTableName, err)
				}
				if workspaceUserCount > 0 {
					return fmt.Errorf(
						"%s contains %d records; delete workspace users before upgrading",
						workspaceUsersTableName,
						workspaceUserCount,
					)
				}
			}
			if tx.Migrator().HasTable(casbinRuleTableName) {
				deleteWorkspaceUserPolicies := fmt.Sprintf(
					"DELETE FROM %s WHERE %s = ? OR %s LIKE ?",
					casbinRuleTableName,
					casbinResourceColumnName,
					casbinResourceColumnName,
				)
				if err := tx.Exec(
					deleteWorkspaceUserPolicies,
					workspaceUserResourceName,
					workspaceUserResourcePattern,
				).Error; err != nil {
					return fmt.Errorf("delete workspace user policies: %w", err)
				}
			}
			if err := tx.Migrator().DropTable(workspaceNamespacesTableName); err != nil {
				return fmt.Errorf("drop %s: %w", workspaceNamespacesTableName, err)
			}
			if err := tx.Migrator().DropTable(workspaceUsersTableName); err != nil {
				return fmt.Errorf("drop %s: %w", workspaceUsersTableName, err)
			}
			return nil
		},
		// Dropping these obsolete tables deletes their data, so the migration
		// cannot safely recreate them during rollback.
	}
}
