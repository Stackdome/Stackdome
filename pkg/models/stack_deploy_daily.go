package models

import "time"

// DeployHistoryDays is the window the stacks list draws — a fortnight is long
// enough to show a rhythm and short enough that a gap still means something.
const DeployHistoryDays = 14

// StackDeployDaily is one stack's deploy count for one day.
//
// It exists because counting stack_releases under-reports: the release GC
// prunes each stack to its retention limit, so a busy stack would draw a quiet
// chart. This tally is written in the same transaction as the release and is
// never pruned.
type StackDeployDaily struct {
	StackID     string    `gorm:"primaryKey;column:stack_id"`
	Day         time.Time `gorm:"primaryKey;column:day;type:date"`
	DeployCount int       `gorm:"column:deploy_count;not null"`
}

func (StackDeployDaily) TableName() string {
	return "stack_deploy_daily"
}
