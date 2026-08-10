package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/computequota"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"gorm.io/gorm/clause"
)

type computeUsageStore struct{}

var _ computequota.UsageStore = (*computeUsageStore)(nil)

func NewComputeUsageStore() computequota.UsageStore {
	return &computeUsageStore{}
}

func (*computeUsageStore) LockOrganisationAndGetUsage(ctx context.Context, organisationID, excludeStackID string) (computequota.ComputeUsage, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return computequota.ComputeUsage{}, errors.GeneralError("transaction not found in context")
	}

	organisationQuery := tx.Model(&models.Organisation{}).
		Select("id").
		Where("id = ?", organisationID).
		Clauses(clause.Locking{Strength: rowLockStrengthUpdate})
	var lockedOrganisationID string
	result := organisationQuery.Scan(&lockedOrganisationID)
	if result.Error != nil {
		return computequota.ComputeUsage{}, errors.GeneralError("failed to lock organisation for compute usage: %v", result.Error)
	}
	if result.RowsAffected != 1 {
		return computequota.ComputeUsage{}, errors.NotFound("organisation '%s' not found", organisationID)
	}

	usage := computequota.ComputeUsage{}
	if err := tx.Model(&models.Stack{}).
		Where("organisation_id = ?", organisationID).
		Count(&usage.StackCount).Error; err != nil {
		return computequota.ComputeUsage{}, errors.GeneralError("failed to count organisation stacks: %v", err)
	}

	stackResourceQuery := tx.Table("stack_resources AS sr").
		Joins("JOIN stacks AS s ON s.id = sr.stack_id").
		Where("s.organisation_id = ?", organisationID)
	if excludeStackID != "" {
		stackResourceQuery = stackResourceQuery.Where("s.id <> ?", excludeStackID)
	}
	if err := stackResourceQuery.Count(&usage.StackResourceCount).Error; err != nil {
		return computequota.ComputeUsage{}, errors.GeneralError("failed to count organisation stack resources: %v", err)
	}

	if err := tx.Model(&models.Volume{}).
		Where("organisation_id = ?", organisationID).
		Count(&usage.VolumeCount).Error; err != nil {
		return computequota.ComputeUsage{}, errors.GeneralError("failed to count organisation volumes: %v", err)
	}

	if err := tx.Model(&models.PostgresAddon{}).
		Where("organisation_id = ?", organisationID).
		Count(&usage.PostgresAddonCount).Error; err != nil {
		return computequota.ComputeUsage{}, errors.GeneralError("failed to count organisation PostgreSQL addons: %v", err)
	}

	return usage, nil
}
