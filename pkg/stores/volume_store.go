package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type VolumeStore interface {
	Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	GetByID(ctx context.Context, id string, workspaceStorageID string) (*models.Volume, *errors.ServiceError)
	Update(ctx context.Context, id string, workspaceStorageID string, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	Delete(ctx context.Context, id string, workspaceStorageID string) *errors.ServiceError
	GetByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.Volume, *errors.ServiceError)
}
