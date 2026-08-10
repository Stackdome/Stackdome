package migrations

import (
	"errors"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func renameClusterPlatformToSharedCompute() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608100001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`ALTER TABLE clusters RENAME COLUMN platform TO shared_compute`).Error; err != nil {
				return fmt.Errorf("rename clusters.platform to shared_compute: %w", err)
			}

			type duplicateClusterURL struct {
				ClusterURL string `gorm:"column:cluster_url"`
				Count      int64  `gorm:"column:cluster_count"`
			}
			var duplicate duplicateClusterURL
			err := tx.Table("clusters").
				Select("cluster_url, COUNT(*) AS cluster_count").
				Group("cluster_url").
				Having("COUNT(*) > 1").
				Order("cluster_url").
				Take(&duplicate).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("check duplicate cluster URLs: %w", err)
			}
			if duplicate.Count > 0 {
				return fmt.Errorf(
					"cannot enforce unique clusters.cluster_url: URL %q is used by %d clusters",
					duplicate.ClusterURL,
					duplicate.Count,
				)
			}
			if err := tx.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_clusters_cluster_url_unique ON clusters (cluster_url)`).Error; err != nil {
				return fmt.Errorf("create unique index on clusters.cluster_url: %w", err)
			}
			return nil
		},
	}
}
