package services

import (
	"context"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
	workspacev1alpha1 "soradev.io/cluster-agent/api/v1alpha1"
)

type VolumeService interface {
	Get(ctx context.Context, ID string, workspaceStorageID string) (*models.Volume, *errors.ServiceError)
	GetByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, workspaceStorageID string, status *models.VolumeStatus) *errors.ServiceError
	AddLastSyncedAtAnnotation(ctx context.Context, ID string, workspaceStorageID string) *errors.ServiceError
	Delete(ctx context.Context, ID string, workspaceStorageID string) *errors.ServiceError
}

type VolumeServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewVolumeService(spec VolumeServiceSpec) VolumeService {
	return &volumeService{
		volumeStore: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type volumeService struct {
	volumeStore stores.VolumeStore
	logger      logger.Logger
}

func (s *volumeService) Get(ctx context.Context, ID string, wsID string) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.GetByID(ctx, ID, wsID)
	if err != nil {
		s.logger.Errorf("failed to get volume: %v", err)
		return nil, err
	}
	return volume, nil
}

func (s *volumeService) GetByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.volumeStore.GetByWorkspaceStorageID(ctx, workspaceStorageID)
	if err != nil {
		s.logger.Errorf("failed to get volumes by workspace storage ID: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.Create(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create volume: %v", err)
		return nil, err
	}
	return volume, nil
}

func (s *volumeService) UpdateStatus(ctx context.Context, ID string, workspaceStorageID string, status *models.VolumeStatus) *errors.ServiceError {
	err := s.volumeStore.UpdateStatus(ctx, ID, workspaceStorageID, status)
	if err != nil {
		s.logger.Errorf("failed to update volume status: %v", err)
		return err
	}
	return nil
}

func (s *volumeService) AddLastSyncedAtAnnotation(ctx context.Context, ID string, workspaceStorageID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID, workspaceStorageID)
	if err != nil {
		return err
	}
	if volume.Annotations == nil {
		volume.Annotations = models.Annotations{}
	}
	volume.Annotations = append(volume.Annotations, models.Annotation{
		Key:   workspacev1alpha1.LastSyncedAtAnnotation,
		Value: time.Now().UTC().Format(time.RFC3339),
	})
	_, updateErr := s.volumeStore.Update(ctx, ID, workspaceStorageID, volume)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func (s *volumeService) Delete(ctx context.Context, ID string, workspaceStorageID string) *errors.ServiceError {
	err := s.volumeStore.Delete(ctx, ID, workspaceStorageID)
	if err != nil {
		s.logger.Errorf("failed to delete volume: %v", err)
		return err
	}
	return nil
}
