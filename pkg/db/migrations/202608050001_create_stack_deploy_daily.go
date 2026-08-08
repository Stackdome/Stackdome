package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// A per-stack, per-day tally of deploys.
//
// The obvious way to draw a deploy chart is to count rows in stack_releases,
// and it is wrong: the release GC prunes each stack down to its retention limit
// (10 by default), so a stack that deployed forty times in a fortnight would
// draw ten bars and read as quiet. A chart that under-reports is worse than no
// chart, so the count is kept separately and never pruned.
//
// One small row per stack per day it deployed — a stack deploying every day for
// a year costs 365 rows. The counter is incremented inside the same transaction
// that inserts the release, so the tally cannot drift from reality.
//
// The backfill seeds what the GC has not taken yet. It is knowingly partial for
// stacks that have already been trimmed; from this migration onward the numbers
// are exact.
func createStackDeployDaily() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202608050001",
		Migrate: func(tx *gorm.DB) error {
			sql := `
CREATE TABLE IF NOT EXISTS stack_deploy_daily (
    stack_id     TEXT NOT NULL REFERENCES stacks(id) ON DELETE CASCADE,
    day          DATE NOT NULL,
    deploy_count INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (stack_id, day)
);
CREATE INDEX IF NOT EXISTS idx_stack_deploy_daily_stack_day
    ON stack_deploy_daily (stack_id, day DESC);
`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create stack_deploy_daily table: %w", err)
			}

			// Seed from the releases that survive today. ON CONFLICT keeps the
			// migration re-runnable against a partially seeded table.
			backfill := `
INSERT INTO stack_deploy_daily (stack_id, day, deploy_count)
SELECT stack_id, (created_at AT TIME ZONE 'UTC')::date AS day, COUNT(*)
FROM stack_releases
GROUP BY stack_id, day
ON CONFLICT (stack_id, day) DO UPDATE SET deploy_count = EXCLUDED.deploy_count;
`
			if err := tx.Exec(backfill).Error; err != nil {
				return fmt.Errorf("failed to backfill stack_deploy_daily: %w", err)
			}
			return nil
		},
		Rollback: func(tx *gorm.DB) error {
			if err := tx.Exec(`DROP TABLE IF EXISTS stack_deploy_daily`).Error; err != nil {
				return fmt.Errorf("failed to drop stack_deploy_daily: %w", err)
			}
			return nil
		},
	}
}
