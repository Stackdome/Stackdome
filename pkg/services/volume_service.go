package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
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
	InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError)
	InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError)
	Create(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateInDbWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateInCluster(ctx context.Context, spec *models.Volume) *errors.ServiceError
	ListByUserID(ctx context.Context, teamID, userID string) ([]*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError
	UpdateGitRepoSourceRevision(ctx context.Context, ID string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError)
	UpdateRemoteSourceRevision(ctx context.Context, ID string, revision models.RemoteDirSource) (*models.Volume, *errors.ServiceError)
	InjectClusterResourceService(volumeClusterService clusterresource.VolumeClusterResourceService)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	UpdateVolumeInUseByStackWithTx(ctx context.Context, volumeID string, stackID string) *errors.ServiceError
	CreateVolumesInDBForStackWithTx(ctx context.Context, stack *models.Stack) ([]*models.Volume, *errors.ServiceError)
	UpdateVolumesInDBForStackWithTx(ctx context.Context, patch *models.Stack, existingStack *models.Stack) ([]*models.Volume, *errors.ServiceError)
}

type VolumeServiceSpec struct {
	SessionFactory         db.SessionFactory
	ConnectionUsageChecker connectionUsageChecker
	Logger                 logger.Logger
	Permissions            auth.PermissionService
}

func NewVolumeService(spec VolumeServiceSpec) VolumeService {
	return &volumeService{
		volumeStore: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		volumeMountStore: pgstore.NewVolumeMountStore(pgstore.VolumeMountStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackVolumeStore: pgstore.NewStackVolumeStore(pgstore.StackVolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		connectionUsageChecker: spec.ConnectionUsageChecker,
		logger:                 spec.Logger,
		permissions:            spec.Permissions,
	}
}

type volumeService struct {
	volumeStore            stores.VolumeStore
	volumeMountStore       stores.VolumeMountStore
	stackVolumeStore       stores.StackVolumeStore
	connectionUsageChecker connectionUsageChecker
	clusterResourceService clusterresource.VolumeClusterResourceService
	logger                 logger.Logger
	permissions            auth.PermissionService
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
	if permErr := s.permissions.Check(ctx, volume.TeamID, auth.ResourceVolumes, ID, auth.ActionRead); permErr != nil {
		return nil, permErr
	}
	return volume, nil
}

func (s *volumeService) InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError) {
	return s.volumeStore.GetByID(ctx, ID)
}

func (s *volumeService) ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.stackVolumeStore.ListVolumesByStackID(ctx, stackID)
	if err != nil {
		s.logger.Errorf("failed to list volumes used by stack: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) CreateVolumesInDBForStackWithTx(ctx context.Context, stack *models.Stack) ([]*models.Volume, *errors.ServiceError) {
	createdVolumes := make([]*models.Volume, 0)
	for _, volume := range stack.Volumes {
		volume.NamespaceID = stack.NamespaceID
		volume.OrganisationID = stack.OrganisationID
		volume.UserID = stack.UserID
		volume.Namespace = stack.Namespace

		createdVolume, err := s.CreateInDbWithTx(ctx, volume)
		if err != nil {
			return nil, errors.GeneralError("failed to create volume '%s': %s", volume.Name, err.Error())
		}
		createdVolumes = append(createdVolumes, createdVolume)
	}
	return createdVolumes, nil
}

func (s *volumeService) UpdateVolumesInDBForStackWithTx(ctx context.Context, patch *models.Stack, existingStack *models.Stack) ([]*models.Volume, *errors.ServiceError) {
	existingVolumesMap := existingStack.VolumesMap()
	newlyCreatedVolumes := make([]*models.Volume, 0)
	for _, volume := range patch.Volumes {
		volume.NamespaceID = existingStack.NamespaceID
		volume.OrganisationID = existingStack.OrganisationID
		volume.UserID = existingStack.UserID
		volume.Namespace = existingStack.Namespace
		if _, found := existingVolumesMap[volume.Name]; found {
			// Updating existing volumes currently not implemented.
			// NOOP
		} else {
			createdVolume, err := s.CreateInDbWithTx(ctx, volume)
			if err != nil {
				return nil, errors.GeneralError("failed to create volume '%s': %s", volume.Name, err.Error())
			}
			newlyCreatedVolumes = append(newlyCreatedVolumes, createdVolume)
		}
	}
	return newlyCreatedVolumes, nil
}

func (s *volumeService) UpdateVolumeInUseByStackWithTx(ctx context.Context, volumeID string, stackID string) *errors.ServiceError {
	return s.stackVolumeStore.CreateWithTx(ctx, &models.StackVolume{
		VolumeID: volumeID,
		StackID:  stackID,
	})
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

func (s *volumeService) ListByUserID(ctx context.Context, teamID, userID string) ([]*models.Volume, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, teamID, auth.ResourceVolumes, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
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
	if permErr := s.permissions.Check(ctx, spec.TeamID, auth.ResourceVolumes, "", auth.ActionCreate); permErr != nil {
		return nil, permErr
	}

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

// Assume ctx already has a transaction
func (s *volumeService) CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	var createdVolume *models.Volume
	var err *errors.ServiceError
	createdVolume, err = s.volumeStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create volume: %v", err)
		return nil, err
	}
	cerr := s.clusterResourceService.CreateVolumeInCluster(ctx, createdVolume)
	if cerr != nil {
		s.logger.Errorf("failed to create volume in cluster: %v", cerr)
		return nil, errors.GeneralError("failed to create volume in cluster: %s", cerr.Error())
	}
	return createdVolume, nil
}

func (s *volumeService) CreateInCluster(ctx context.Context, spec *models.Volume) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, spec.ID)
	if err != nil {
		s.logger.Errorf("failed to get volume for creation in cluster: %v", err)
		return err
	}
	if volume == nil {
		s.logger.Errorf("volume not found for creation in cluster: %v", spec.ID)
		return errors.NotFound("volume not found")
	}
	cErr := s.clusterResourceService.CreateVolumeInCluster(ctx, volume)
	if cErr != nil {
		s.logger.Errorf("failed to create volume in cluster: %v", cErr)
		return errors.GeneralError("failed to create volume in cluster: %s", cErr.Error())
	}
	return nil
}

func (s *volumeService) CreateInDbWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	var createdVolume *models.Volume
	var err *errors.ServiceError
	createdVolume, err = s.volumeStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Errorf("failed to create volume: %v", err)
		return nil, err
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

	if permErr := s.permissions.Check(ctx, volume.TeamID, auth.ResourceVolumes, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	if err := s.validateVolumeNotReferencedByConnections(ctx, volume); err != nil {
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

func (s *volumeService) InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	if err := s.validateVolumeNotReferencedByConnections(ctx, volume); err != nil {
		return err
	}
	volumeMounts, err := s.volumeMountStore.ListBySourceVolumeID(ctx, ID)
	if err != nil {
		return err
	}
	if len(volumeMounts) > 0 {
		return errors.BadRequest("cannot delete volume with mounts")
	}
	if err := s.volumeStore.Delete(ctx, ID); err != nil {
		return err
	}
	return nil
}

func (s *volumeService) InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError {
	volumesUsedByStack, err := s.ListVolumesUsedByStack(ctx, stackID)
	if err != nil {
		return err
	}
	for _, volume := range volumesUsedByStack {
		if err := s.InternalDeleteFromDB(ctx, volume.ID); err != nil {
			if err.Is404() {
				continue
			}
			return err
		}
	}
	return nil
}

func (s *volumeService) DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get volume for deletion: %v", err)
		return err
	}
	if err := s.validateVolumeNotReferencedByConnections(ctx, volume); err != nil {
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
}

func (s *volumeService) validateVolumeNotReferencedByConnections(ctx context.Context, volume *models.Volume) *errors.ServiceError {
	stackVolume, err := s.stackVolumeStore.GetByVolumeID(ctx, volume.ID)
	if err != nil {
		if err.Is404() {
			return nil
		}
		return err
	}

	inUse, usageErr := s.connectionUsageChecker.IsNodeReferenced(ctx, stackVolume.StackID, models.TopologyNodeRef{
		Type: models.TopologyNodeTypeVolume,
		Name: volume.Name,
	})
	if usageErr != nil {
		return errors.GeneralError("failed to check connection usages for volume ID %s: %s", volume.ID, usageErr.Error())
	}
	if inUse {
		return errors.BadRequest("volume is in use by one or more stack connections")
	}
	return nil
}
