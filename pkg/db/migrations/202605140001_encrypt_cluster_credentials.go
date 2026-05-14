package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func encryptClusterCredentials() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605140001_encrypt_cluster_credentials",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec("ALTER TABLE clusters ADD COLUMN IF NOT EXISTS encrypted_token TEXT NOT NULL DEFAULT ''").Error; err != nil {
				return fmt.Errorf("error adding encrypted_token column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE clusters ADD COLUMN IF NOT EXISTS encrypted_cluster_ca_data TEXT NOT NULL DEFAULT ''").Error; err != nil {
				return fmt.Errorf("error adding encrypted_cluster_ca_data column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_token_check").Error; err != nil {
				return fmt.Errorf("error dropping token check constraint: %w", err)
			}
			if err := tx.Exec("ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_cluster_ca_data_check").Error; err != nil {
				return fmt.Errorf("error dropping cluster_ca_data check constraint: %w", err)
			}
			if err := tx.Exec("ALTER TABLE clusters DROP COLUMN IF EXISTS token").Error; err != nil {
				return fmt.Errorf("error dropping token column: %w", err)
			}
			if err := tx.Exec("ALTER TABLE clusters DROP COLUMN IF EXISTS cluster_ca_data").Error; err != nil {
				return fmt.Errorf("error dropping cluster_ca_data column: %w", err)
			}
			return nil
		},
	}
}
