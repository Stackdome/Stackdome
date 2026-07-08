package pgstore

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ResourceValidationRecordStoreSpec struct {
	SessionFactory db.SessionFactory
}

type resourceValidationRecordStore struct {
	sessionFactory db.SessionFactory
}

func NewResourceValidationRecordStore(spec ResourceValidationRecordStoreSpec) stores.ResourceValidationRecordStore {
	return &resourceValidationRecordStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *resourceValidationRecordStore) Get(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind) (*models.ResourceValidationRecord, *errors.ServiceError) {
	var record models.ResourceValidationRecord
	if err := s.sessionFactory.New(ctx).
		Where("stack_id = ? AND resource_name = ? AND check_kind = ?", stackID, resourceName, kind).
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("no validation record for stack '%s' resource '%s' check '%s'", stackID, resourceName, kind)
		}
		return nil, errors.GeneralError("failed to get resource validation record: %s", err.Error())
	}
	return &record, nil
}

func (s *resourceValidationRecordStore) Upsert(ctx context.Context, record *models.ResourceValidationRecord) *errors.ServiceError {
	if err := s.sessionFactory.New(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "stack_id"}, {Name: "resource_name"}, {Name: "check_kind"}},
			UpdateAll: true,
		}).
		Create(record).Error; err != nil {
		return errors.GeneralError("failed to upsert resource validation record: %s", err.Error())
	}
	return nil
}
