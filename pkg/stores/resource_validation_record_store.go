package stores

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
)

//go:generate mockgen -source=resource_validation_record_store.go -destination=../mocks/mock_resource_validation_record_store.go -package=mocks
type ResourceValidationRecordStore interface {
	Get(ctx context.Context, stackID, resourceName string, kind models.ResourceValidationCheckKind) (*models.ResourceValidationRecord, *errors.ServiceError)
	Upsert(ctx context.Context, record *models.ResourceValidationRecord) *errors.ServiceError
}
