package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type VolumeMountStore interface {
	ListByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.VolumeMount, *errors.ServiceError)
}
