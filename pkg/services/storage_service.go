package services

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services/clusterresource"
)

type StackStorageService interface {
	Get(ctx context.Context, ID string, userID string) (*models.StackStorage, *errors.ServiceError)
	GetByID(ctx context.Context, ID string) (*models.StackStorage, *errors.ServiceError)
	GetbyWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.StackStorage, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.StackStorage, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.StackStorage, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.StackStorage, *errors.ServiceError)
	Create(ctx context.Context, spec *models.StackStorage, userID string) (*models.StackStorage, *errors.ServiceError)
	Update(ctx context.Context, ID string, userID string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.StackStorageStatus) *errors.ServiceError
	InjectClusterResourceService(clusterResourceService clusterresource.StackStorageClusterResourceService)
	ListVolumes(ctx context.Context, workspaceStorageID string, userID string) ([]*models.Volume, *errors.ServiceError)
	Delete(ctx context.Context, ID string, userID string) *errors.ServiceError
	MarkAsSynced(ctx context.Context, userID string, storageID string, volumeID string) *errors.ServiceError
}

// type StorageServiceSpec struct {
// 	SessionFactory db.SessionFactory
// 	Logger         logger.Logger
// }

// func NewStorageService(spec StorageServiceSpec) StackStorageService {
// 	return &storageService{
// 		storageStore: pgstore.NewStackStorageStore(pgstore.StackStorageStoreSpec{
// 			SessionFactory: spec.SessionFactory,
// 		}),
// 		workspaceNamespaceStore: pgstore.NewWorkspaceNamespaceStore(pgstore.WorkspaceNamespaceStoreSpec{
// 			SessionFactory: spec.SessionFactory,
// 		}),
// 		volumeService: NewVolumeService(VolumeServiceSpec{
// 			SessionFactory: spec.SessionFactory,
// 			Logger:         spec.Logger,
// 		}),
// 		volumeMountStore: pgstore.NewVolumeMountStore(pgstore.VolumeMountStoreSpec{
// 			SessionFactory: spec.SessionFactory,
// 		}),
// 		logger: spec.Logger,
// 	}
// }

// type storageService struct {
// 	storageStore            stores.StorageStore
// 	wsUserService           WorkspaceUserService
// 	workspaceNamespaceStore stores.WorkspaceNamespaceStore
// 	clusterResourceService  clusterresource.StackStorageClusterResourceService
// 	volumeService           VolumeService
// 	volumeMountStore        stores.VolumeMountStore
// 	logger                  logger.Logger
// }

// func (s *storageService) InjectClusterResourceService(clusterResourceService clusterresource.StackStorageClusterResourceService) {
// 	s.clusterResourceService = clusterResourceService
// }

// func (s *storageService) Get(ctx context.Context, ID string, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	storage, err := s.storageStore.GetByIDorName(ctx, ID, userID)
// 	if err != nil {
// 		s.logger.Errorf("failed to get storage: %v", err)
// 		return nil, err
// 	}

// 	return storage, nil
// }

// func (s *storageService) GetByID(ctx context.Context, ID string) (*models.StackStorage, *errors.ServiceError) {
// 	storage, err := s.storageStore.GetByID(ctx, ID)
// 	if err != nil {
// 		s.logger.Errorf("failed to get storage: %v", err)
// 		return nil, err
// 	}
// 	return storage, nil
// }

// func (s *storageService) GetbyWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	storage, err := s.storageStore.GetByWorkspaceName(ctx, workspaceName, userID)
// 	if err != nil {
// 		s.logger.Errorf("failed to get  storage: %v", err)
// 		return nil, err
// 	}
// 	return storage, nil
// }

// func (s *storageService) InternalGet(ctx context.Context, ID string) (*models.StackStorage, *errors.ServiceError) {
// 	storage, err := s.storageStore.GetByID(ctx, ID)
// 	if err != nil {
// 		s.logger.Errorf("failed to get storage: %v", err)
// 		return nil, err
// 	}

// 	return storage, nil
// }

// func (s *storageService) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.StackStorage, *errors.ServiceError) {
// 	storages, err := s.storageStore.ListByOrganisationID(ctx, organisationID)
// 	if err != nil {
// 		s.logger.Errorf("failed to list storages: %v", err)
// 		return nil, err
// 	}
// 	return storages, nil
// }

// func (s *storageService) ListByUserID(ctx context.Context, userID string) ([]*models.StackStorage, *errors.ServiceError) {
// 	storages, err := s.storageStore.ListByUserID(ctx, userID)
// 	if err != nil {
// 		s.logger.Errorf("failed to list storages by user ID: %v", err)
// 		return nil, err
// 	}
// 	return storages, nil
// }

// func (s *storageService) ListVolumes(ctx context.Context, stackStorageID string, userID string) ([]*models.Volume, *errors.ServiceError) {
// 	storage, err := s.storageStore.GetByIDorName(ctx, stackStorageID, userID)
// 	if err != nil {
// 		s.logger.Errorf("failed to get storage: %v", err)
// 		return nil, err
// 	}

// 	return storage.Volumes, nil
// }

// func (s *storageService) Create(ctx context.Context, spec *models.StackStorage, userID string) (*models.StackStorage, *errors.ServiceError) {
// 	if spec.OrganisationID == "" {
// 		return nil, errors.BadRequest("organisation ID is required")
// 	}
// 	s.logger.Infof("creating storage: %+v", *spec)

// 	WSNamespace, serr := s.workspaceNamespaceStore.GetByWorkspaceName(ctx, spec.WorkspaceName, userID)
// 	if serr != nil {
// 		s.logger.Errorf("failed to get workspace namespace: %v", serr)
// 		if serr.Code == errors.ErrorNotFound {
// 			return nil, errors.NotFound("workspace '%s' not found", spec.WorkspaceName)
// 		}
// 		return nil, serr
// 	}
// 	spec.Namespace = WSNamespace.Namespace

// 	currentuserWorkspaceStorages, err := s.storageStore.ListByUserID(ctx, spec.UserID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	workspaceUser, err := s.wsUserService.GetWorkspaceUser(ctx, userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// If no ssh config is provided, use the currect workspace user's ssh public key.
// 	if len(spec.SSHConfig.PublicKey) == 0 {
// 		spec.SSHConfig = models.SSHConfig{
// 			PublicKey: workspaceUser.SshPublicKey,
// 		}
// 	}

// 	for _, storage := range currentuserWorkspaceStorages {
// 		if storage.Name == spec.Name {
// 			return nil, errors.BadRequest("storage with name '%s' already exists for user", spec.Name)
// 		}
// 	}

// 	var createdStorage *models.StackStorage
// 	createErr := s.storageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
// 		createdStorage, serr = s.storageStore.CreateWithTx(ctx, spec)
// 		if serr != nil {
// 			s.logger.Errorf("failed to create storage: %v", serr)
// 			return serr
// 		}
// 		if err := s.clusterResourceService.UpsertStorageInCluster(ctx, createdStorage); err != nil {
// 			s.logger.Errorf("failed to create storage in cluster: %v", err)
// 			return errors.GeneralError("failed to create storage in cluster: %s", err.Error())
// 		}
// 		return nil
// 	})
// 	if createErr != nil {
// 		return nil, createErr
// 	}

// 	return createdStorage, nil
// }

// // Validate incoming spec for update
// func (s *storageService) validateUpdateSpec(spec *models.StackStorage, existingStorage *models.StackStorage) *errors.ServiceError {
// 	// TODO: Move this to a validation layer
// 	if existingStorage.WorkspaceName != spec.WorkspaceName {
// 		return errors.BadRequest("workspace cannot be updated")
// 	}

// 	if existingStorage.Name != spec.Name {
// 		return errors.BadRequest("storage name cannot be updated")
// 	}

// 	for _, volume := range spec.Volumes {
// 		if existingVolume, ok := existingStorage.VolumeMap()[volume.Name]; ok {
// 			if !allowedVolumeUpdate(existingVolume, volume) {
// 				return errors.BadRequest("volume fields like size, storage class and name cannot be updated")
// 			}
// 		}
// 	}
// 	return nil
// }

// func (s *storageService) Update(ctx context.Context, ID string, userID string, spec *models.StackStorage) (*models.StackStorage, *errors.ServiceError) {
// 	existingStorage, serr := s.storageStore.GetByIDorName(ctx, ID, userID)
// 	if serr != nil {
// 		s.logger.Errorf("failed to get storage: %v", serr)
// 		return nil, serr
// 	}

// 	if err := s.validateUpdateSpec(spec, existingStorage); err != nil {
// 		return nil, err
// 	}

// 	var updatedStorage *models.StackStorage
// 	updateErr := s.storageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
// 		updatedStorage, serr = s.storageStore.UpdateWithTx(ctx, existingStorage.ID, spec)
// 		if serr != nil {
// 			s.logger.Errorf("failed to update storage: %v", serr)
// 			return serr
// 		}
// 		if err := s.clusterResourceService.UpsertStorageInCluster(ctx, updatedStorage); err != nil {
// 			s.logger.Errorf("failed to update storage in cluster: %v", err)
// 			return errors.GeneralError("failed to update storage in cluster: %s", err.Error())
// 		}
// 		return nil
// 	})

// 	if updateErr != nil {
// 		s.logger.Errorf("failed to update storage: %v", updateErr)
// 		return nil, updateErr
// 	}
// 	return updatedStorage, nil
// }

// func (s *storageService) MarkAsSynced(ctx context.Context, userID string, storageID string, volumeID string) *errors.ServiceError {
// 	// After adding the annotation we update the storage object in the cluster.
// 	err := s.volumeService.AddLastSyncedAtAnnotation(ctx, volumeID, storageID)
// 	if err != nil {
// 		return errors.GeneralError("failed to mark volume as synced: %s", err.Error())
// 	}
// 	existingStorage, serr := s.storageStore.GetByIDorName(ctx, storageID, userID)
// 	if serr != nil {
// 		s.logger.Errorf("failed to get storage: %v", serr)
// 		return serr
// 	}

// 	if err := s.clusterResourceService.UpsertStorageInCluster(ctx, existingStorage); err != nil {
// 		s.logger.Errorf("failed to update storage in cluster: %v", err)
// 		return errors.GeneralError("failed to update storage in cluster: %s", err.Error())
// 	}
// 	return nil
// }

// func allowedVolumeUpdate(existingVolume *models.Volume, volume *models.Volume) bool {
// 	return existingVolume.Size == volume.Size && existingVolume.StorageClass == volume.StorageClass && volume.Name == existingVolume.Name
// }

// func (s *storageService) UpdateStatus(ctx context.Context, ID string, status *models.StackStorageStatus) *errors.ServiceError {
// 	_, err := s.storageStore.UpdateStatus(ctx, ID, status)
// 	if err != nil {
// 		s.logger.Errorf("failed to update storage status: %v", err)
// 		return err
// 	}
// 	return nil
// }

// func (s *storageService) Delete(ctx context.Context, ID string, userID string) *errors.ServiceError {
// 	deleteErr := s.storageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
// 		existingStorage, serr := s.storageStore.GetByIDorName(ctx, ID, userID)
// 		if serr != nil {
// 			s.logger.Errorf("failed to get storage: %v", serr)
// 			return serr
// 		}

// 		volumeMounts, serr := s.volumeMountStore.ListByStorageID(ctx, existingStorage.ID)
// 		if serr != nil {
// 			s.logger.Errorf("failed to list volume mounts: %v", serr)
// 			return serr
// 		}

// 		if len(volumeMounts) > 0 {
// 			return errors.BadRequest("storage is currently mounted in a stack")
// 		}

// 		if err := s.clusterResourceService.DeleteStorageInCluster(ctx, existingStorage); err != nil {
// 			s.logger.Errorf("failed to delete storage in cluster: %v", err)
// 			return errors.GeneralError("failed to delete storage in cluster: %s", err.Error())
// 		}

// 		if err := s.storageStore.DeleteWithTx(ctx, existingStorage.ID); err != nil {
// 			s.logger.Errorf("failed to delete storage: %v", err)
// 			return err
// 		}
// 		return nil
// 	})
// 	if deleteErr != nil {
// 		s.logger.Errorf("failed to delete storage: %v", deleteErr)
// 		return deleteErr
// 	}
// 	return nil
// }
