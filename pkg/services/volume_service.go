package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type VolumeService interface {
	Get(ctx context.Context, ID string, workspaceStorageID string) (*models.Volume, *errors.ServiceError)
	GetByWorkspaceStorageID(ctx context.Context, workspaceStorageID string) ([]*models.Volume, *errors.ServiceError)
	Update(ctx context.Context, ID string, workspaceStorageID string, spec *models.Volume) (*models.Volume, *errors.ServiceError)
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

func (s *volumeService) Update(ctx context.Context, ID string, workspaceStorageID string, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.Update(ctx, ID, workspaceStorageID, spec)
	if err != nil {
		s.logger.Errorf("failed to update volume: %v", err)
		return nil, err
	}
	return volume, nil
}

func (s *volumeService) Delete(ctx context.Context, ID string, workspaceStorageID string) *errors.ServiceError {
	err := s.volumeStore.Delete(ctx, ID, workspaceStorageID)
	if err != nil {
		s.logger.Errorf("failed to delete volume: %v", err)
		return err
	}
	return nil
}
