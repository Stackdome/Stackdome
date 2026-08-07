package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func deleteSeededPlatformDomains() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608070001",
		Migrate: func(tx *gorm.DB) error {
			if err := tx.Exec(`
WITH platform_base AS MATERIALIZED (
	SELECT d.domain
	FROM organisation_domains d
	JOIN organisations o ON o.id = d.organisation_id
	WHERE o.platform = true
	ORDER BY d.created_at
	LIMIT 1
)
DELETE FROM organisation_domains AS d
WHERE EXISTS (
	SELECT 1
	FROM platform_base AS b
	WHERE d.domain = b.domain
		OR (
			d.domain LIKE '%.' || b.domain
			AND length(d.domain) - length(replace(d.domain, '.', '')) =
				length(b.domain) - length(replace(b.domain, '.', '')) + 1
		)
)`).Error; err != nil {
				return fmt.Errorf("delete seeded platform organisation domains: %w", err)
			}
			return nil
		},
	}
}
