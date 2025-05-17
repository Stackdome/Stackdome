package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type StackVolumeStore interface {
	Create(ctx context.Context, sv *models.StackVolume) *errors.ServiceError
	Delete(ctx context.Context, stackID, volumeID string) *errors.ServiceError
	GetByVolumeID(ctx context.Context, volumeID string) (*models.StackVolume, *errors.ServiceError)
	ListVolumesByStackID(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	CreateWithTx(ctx context.Context, sv *models.StackVolume) *errors.ServiceError
	DeleteWithTx(ctx context.Context, stackID, volumeID string) *errors.ServiceError
}
