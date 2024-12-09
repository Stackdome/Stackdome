package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type volumeMountStore struct {
	sessionFactory db.SessionFactory
}

type VolumeMountStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewVolumeMountStore(spec VolumeMountStoreSpec) stores.VolumeMountStore {
	return &volumeMountStore{
		sessionFactory: spec.SessionFactory,
	}
}

// list volumeMounts by workspaceStorageID

func (v *volumeMountStore) ListByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.VolumeMount, *errors.ServiceError) {
	var volumeMounts []*models.VolumeMount
	if err := v.sessionFactory.New(ctx).Where("workspace_storage_id = ?", workspaceStorageID).Find(&volumeMounts).Error; err != nil {
		return nil, errors.GeneralError("failed to get volume mounts: %s", err.Error())
	}
	return volumeMounts, nil
}
