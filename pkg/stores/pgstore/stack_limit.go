package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm/clause"
)

type computeUsageStore struct{}

var _ stores.ComputeUsageStore = (*computeUsageStore)(nil)

func NewComputeUsageStore() stores.ComputeUsageStore {
	return &computeUsageStore{}
}

func (*computeUsageStore) LockOrganisationAndGetUsageWithTx(ctx context.Context, organisationID, excludeStackID string) (stores.ComputeUsage, *errors.ServiceError) {
	tx := db.TxFromContext(ctx)
	if tx == nil {
		return stores.ComputeUsage{}, errors.GeneralError("transaction not found in context")
	}

	query := tx.Model(&models.Organisation{}).Select("id").Where("id = ?", organisationID)
	if tx.Name() == postgresDialectName {
		query = query.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	var lockedOrganisationID string
	result := query.Scan(&lockedOrganisationID)
	if result.Error != nil {
		return stores.ComputeUsage{}, errors.GeneralError("failed to lock organisation for compute usage: %s", result.Error.Error())
	}
	if result.RowsAffected != 1 {
		return stores.ComputeUsage{}, errors.NotFound("organisation '%s' not found", organisationID)
	}

	usage := stores.ComputeUsage{}
	if err := tx.Model(&models.Stack{}).Where("organisation_id = ?", organisationID).Count(&usage.StackCount).Error; err != nil {
		return stores.ComputeUsage{}, errors.GeneralError("failed to count organisation stacks: %s", err.Error())
	}
	resourceQuery := tx.Table("stack_resources AS sr").
		Joins("JOIN stacks AS s ON s.id = sr.stack_id").
		Where("s.organisation_id = ?", organisationID)
	if excludeStackID != "" {
		resourceQuery = resourceQuery.Where("s.id <> ?", excludeStackID)
	}
	if err := resourceQuery.Count(&usage.StackResourceCount).Error; err != nil {
		return stores.ComputeUsage{}, errors.GeneralError("failed to count organisation stack resources: %s", err.Error())
	}
	if err := tx.Model(&models.Volume{}).Where("organisation_id = ?", organisationID).Count(&usage.VolumeCount).Error; err != nil {
		return stores.ComputeUsage{}, errors.GeneralError("failed to count organisation volumes: %s", err.Error())
	}
	if err := tx.Model(&models.PostgresAddon{}).Where("organisation_id = ?", organisationID).Count(&usage.PostgresAddonCount).Error; err != nil {
		return stores.ComputeUsage{}, errors.GeneralError("failed to count organisation PostgreSQL addons: %s", err.Error())
	}
	return usage, nil
}
