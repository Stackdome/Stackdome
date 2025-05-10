package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

type VolumeService interface {
	Get(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError)
	InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError)
	Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError
	UpdateGitRepoSourceRevision(ctx context.Context, ID string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError)
	UpdateRemoteSourceRevision(ctx context.Context, ID string, revision models.RemoteDirSource) (*models.Volume, *errors.ServiceError)
	InjectClusterResourceService(volumeClusterService clusterresource.VolumeClusterResourceService)
	Delete(ctx context.Context, ID string) *errors.ServiceError
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
		volumeMountStore: pgstore.NewVolumeMountStore(pgstore.VolumeMountStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type volumeService struct {
	volumeStore            stores.VolumeStore
	volumeMountStore       stores.VolumeMountStore
	clusterResourceService clusterresource.VolumeClusterResourceService
	logger                 logger.Logger
}

func (s *volumeService) InjectClusterResourceService(volumeClusterService clusterresource.VolumeClusterResourceService) {
	s.clusterResourceService = volumeClusterService
}

func (s *volumeService) Get(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get volume: %v", err)
		return nil, err
	}
	return volume, nil
}

func (s *volumeService) GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.GetByVolumeNameAndNamespace(ctx, volumeName, namespace)
	if err != nil {
		s.logger.Errorf("failed to get volume by name and namespace: %v", err)
		return nil, err
	}
	if volume == nil {
		return nil, errors.NotFound("volume not found")
	}
	return volume, nil
}

func (s *volumeService) ListByUserID(ctx context.Context, userID string) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.volumeStore.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to list volumes by user ID: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.volumeStore.InternalList(ctx, ids)
	if err != nil {
		s.logger.Errorf("failed to list volumes: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) UpdateGitRepoSourceRevision(ctx context.Context, ID string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	var updatedVolume *models.Volume
	var err *errors.ServiceError

	if !revision.IsValid() {
		return nil, errors.BadRequest("invalid git repo revision")
	}
	updateErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedVolume, err = s.volumeStore.UpdateGitRepoSourceRevisionWithTx(ctx, ID, revision)
		if err != nil {
			s.logger.Errorf("failed to update volume git repo source revision: %v", err)
			return err
		}
		cErr := s.clusterResourceService.UpdateVolumeGitRevisionInCluster(ctx, updatedVolume)
		if cErr != nil {
			s.logger.Errorf("failed to update volume git revision in cluster: %v", cErr)
			return errors.GeneralError("failed to update volume git revision in cluster: %s", cErr.Error())
		}
		return nil
	})
	if updateErr != nil {
		return nil, updateErr
	}
	return updatedVolume, nil
}

func (s *volumeService) UpdateRemoteSourceRevision(ctx context.Context, ID string, revision models.RemoteDirSource) (*models.Volume, *errors.ServiceError) {
	var updatedVolume *models.Volume
	var err *errors.ServiceError
	if revision.CurrentDirectoryHash == "" {
		return nil, errors.BadRequest("current directory hash is required")
	}

	updateErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedVolume, err = s.volumeStore.UpdateRemoteDirSourceHashWithTx(ctx, ID, revision.CurrentDirectoryHash)
		if err != nil {
			s.logger.Errorf("failed to update volume remote source revision: %v", err)
			return err
		}

		cErr := s.clusterResourceService.UpdateVolumeGitRevisionInCluster(ctx, updatedVolume)
		if cErr != nil {
			s.logger.Errorf("failed to update volume git revision in cluster: %v", cErr)
			return errors.GeneralError("failed to update volume git revision in cluster: %s", cErr.Error())
		}
		return nil
	})
	if updateErr != nil {
		return nil, updateErr
	}
	return updatedVolume, nil
}

func (s *volumeService) Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	var createdVolume *models.Volume
	var err *errors.ServiceError
	createErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		createdVolume, err = s.volumeStore.CreateWithTx(ctx, spec)
		if err != nil {
			s.logger.Errorf("failed to create volume: %v", err)
			return err
		}
		cerr := s.clusterResourceService.CreateVolumeInCluster(ctx, createdVolume)
		if cerr != nil {
			s.logger.Errorf("failed to create volume in cluster: %v", cerr)
			return errors.GeneralError("failed to create volume in cluster: %s", cerr.Error())
		}
		return nil
	})
	if createErr != nil {
		return nil, createErr
	}
	return createdVolume, nil
}

func (s *volumeService) UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError {
	err := s.volumeStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		s.logger.Errorf("failed to update volume status: %v", err)
		return err
	}
	return nil
}

func (s *volumeService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get volume for deletion: %v", err)
		return err
	}

	volumeMounts, err := s.volumeMountStore.ListBySourceVolumeID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to list volume mounts for deletion: %v", err)
	}
	if len(volumeMounts) > 0 {
		s.logger.Errorf("cannot delete volume with mounts: %v", volumeMounts)
		return errors.BadRequest("cannot delete volume with mounts")
	}

	deleteErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		cErr := s.clusterResourceService.DeleteVolumeInCluster(ctx, volume)
		if cErr != nil {
			s.logger.Errorf("failed to delete volume in cluster: %v", cErr)
			return errors.GeneralError("failed to delete volume in cluster: %s", cErr.Error())
		}
		err = s.volumeStore.DeleteWithTx(ctx, ID)
		if err != nil {
			s.logger.Errorf("failed to delete volume: %v", err)
			return err
		}
		return nil
	})
	if deleteErr != nil {
		s.logger.Errorf("failed to delete volume: %v", deleteErr)
		return deleteErr
	}
	return nil
}
