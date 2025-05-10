package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type VolumeMountStore interface {
	ListByStackResourceID(ctx context.Context, stackResourceID string) ([]*models.VolumeMount, *errors.ServiceError)
	ListByStackID(ctx context.Context, stackID string) ([]*models.VolumeMount, *errors.ServiceError)
	ListBySourceVolumeID(ctx context.Context, sourceVolumeID string) ([]*models.VolumeMount, *errors.ServiceError)
	Create(ctx context.Context, volumeMount *models.VolumeMount) (*models.VolumeMount, *errors.ServiceError)
	CreateWithTx(ctx context.Context, volumeMount *models.VolumeMount) (*models.VolumeMount, *errors.ServiceError)
	BulkCreateWithTx(ctx context.Context, volumeMounts []*models.VolumeMount) ([]*models.VolumeMount, *errors.ServiceError)
	DeleteForStackResource(ctx context.Context, stackResourceID string) *errors.ServiceError
	DeleteForStackResourceWithTx(ctx context.Context, stackResourceID string) *errors.ServiceError
	DeleteForStack(ctx context.Context, stackID string) *errors.ServiceError
	DeleteForStackWithTx(ctx context.Context, stackID string) *errors.ServiceError
}
