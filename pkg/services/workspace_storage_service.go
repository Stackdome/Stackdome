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

type WorkspaceStorageService interface {
	Get(ctx context.Context, ID string, userID string) (*models.WorkspaceStorage, *errors.ServiceError)
	GetbyWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.WorkspaceStorage, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.WorkspaceStorage, *errors.ServiceError)
	ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError)
	Create(ctx context.Context, spec *models.WorkspaceStorage, userID string) (*models.WorkspaceStorage, *errors.ServiceError)
	Update(ctx context.Context, ID string, userID string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStorageStatus) *errors.ServiceError
	InjectClusterResourceService(clusterResourceService clusterresource.WorkspaceStorageClusterResourceService)
	ListVolumes(ctx context.Context, workspaceStorageID string, userID string) ([]*models.Volume, *errors.ServiceError)
	Delete(ctx context.Context, ID string, userID string) *errors.ServiceError
}

type WorkspaceStorageServiceSpec struct {
	SessionFactory db.SessionFactory
	Logger         logger.Logger
}

func NewWorkspaceStorageService(spec WorkspaceStorageServiceSpec) WorkspaceStorageService {
	return &workspaceStorageService{
		wsStorageStore: pgstore.NewWorkspaceStorageStore(pgstore.WorkspaceStorageStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceNamespaceStore: pgstore.NewWorkspaceNamespaceStore(pgstore.WorkspaceNamespaceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		logger: spec.Logger,
	}
}

type workspaceStorageService struct {
	wsStorageStore          stores.WorkspaceStorageStore
	workspaceNamespaceStore stores.WorkspaceNamespaceStore
	clusterResourceService  clusterresource.WorkspaceStorageClusterResourceService
	logger                  logger.Logger
}

func (s *workspaceStorageService) InjectClusterResourceService(clusterResourceService clusterresource.WorkspaceStorageClusterResourceService) {
	s.clusterResourceService = clusterResourceService
}

func (s *workspaceStorageService) Get(ctx context.Context, ID string, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.GetByIDorName(ctx, ID, userID)
	if err != nil {
		s.logger.Errorf("failed to get workspace storage: %v", err)
		return nil, err
	}

	return storage, nil
}

func (s *workspaceStorageService) GetbyWorkspaceName(ctx context.Context, workspaceName string, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.GetByWorkspaceName(ctx, workspaceName, userID)
	if err != nil {
		s.logger.Errorf("failed to get workspace storage: %v", err)
		return nil, err
	}
	return storage, nil
}

func (s *workspaceStorageService) InternalGet(ctx context.Context, ID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	storage, err := s.wsStorageStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Errorf("failed to get workspace storage: %v", err)
		return nil, err
	}

	return storage, nil
}

func (s *workspaceStorageService) ListByOrganisationID(ctx context.Context, organisationID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	storages, err := s.wsStorageStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		s.logger.Errorf("failed to list workspace storages: %v", err)
		return nil, err
	}
	return storages, nil
}

func (s *workspaceStorageService) ListByUserID(ctx context.Context, userID string) ([]*models.WorkspaceStorage, *errors.ServiceError) {
	storages, err := s.wsStorageStore.ListByUserID(ctx, userID)
	if err != nil {
		s.logger.Errorf("failed to list workspace storages by user ID: %v", err)
		return nil, err
	}
	return storages, nil
}

func (s *workspaceStorageService) ListVolumes(ctx context.Context, workspaceStorageID string, userID string) ([]*models.Volume, *errors.ServiceError) {
	storage, err := s.wsStorageStore.GetByIDorName(ctx, workspaceStorageID, userID)
	if err != nil {
		s.logger.Errorf("failed to get workspace storage: %v", err)
		return nil, err
	}

	return storage.Volumes, nil
}

func (s *workspaceStorageService) Create(ctx context.Context, spec *models.WorkspaceStorage, userID string) (*models.WorkspaceStorage, *errors.ServiceError) {
	if spec.OrganisationID == "" {
		return nil, errors.BadRequest("organisation ID is required")
	}
	s.logger.Infof("creating workspace storage: %+v", *spec)

	WSNamespace, serr := s.workspaceNamespaceStore.GetByWorkspaceName(ctx, spec.WorkspaceName, userID)
	if serr != nil {
		s.logger.Errorf("failed to get workspace namespace: %v", serr)
		if serr.Code == errors.ErrorNotFound {
			return nil, errors.NotFound("workspace '%s' not found", spec.WorkspaceName)
		}
		return nil, serr
	}
	spec.Namespace = WSNamespace.Namespace

	currentuserWorkspaceStorages, err := s.wsStorageStore.ListByUserID(ctx, spec.UserID)
	if err != nil {
		return nil, err
	}

	for _, storage := range currentuserWorkspaceStorages {
		if storage.Name == spec.Name {
			return nil, errors.BadRequest("workspace storage with name '%s' already exists for user", spec.Name)
		}
	}

	var createdStorage *models.WorkspaceStorage
	createErr := s.wsStorageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		createdStorage, serr = s.wsStorageStore.CreateWithTx(ctx, spec)
		if serr != nil {
			s.logger.Errorf("failed to create workspace storage: %v", serr)
			return serr
		}
		if err := s.clusterResourceService.UpsertWorkspaceStorageInCluster(ctx, createdStorage); err != nil {
			s.logger.Errorf("failed to create workspace storage in cluster: %v", err)
			return errors.GeneralError("failed to create workspace storage in cluster: %s", err.Error())
		}
		return nil
	})
	if createErr != nil {
		return nil, createErr
	}

	return createdStorage, nil
}

// Validate incoming spec for update
func (s *workspaceStorageService) validateUpdateSpec(spec *models.WorkspaceStorage, existingStorage *models.WorkspaceStorage) *errors.ServiceError {
	// TODO: Move this to a validation layer
	if existingStorage.WorkspaceName != spec.WorkspaceName {
		return errors.BadRequest("workspace name cannot be updated")
	}

	if existingStorage.Name != spec.Name {
		return errors.BadRequest("workspace storage name cannot be updated")
	}

	for _, volume := range spec.Volumes {
		if existingVolume, ok := existingStorage.VolumeMap()[volume.Name]; ok {
			if !allowedVolumeUpdate(existingVolume, volume) {
				return errors.BadRequest("volume fields like size, storage class and name cannot be updated")
			}
		}
	}
	return nil
}

func (s *workspaceStorageService) Update(ctx context.Context, ID string, userID string, spec *models.WorkspaceStorage) (*models.WorkspaceStorage, *errors.ServiceError) {
	existingStorage, serr := s.wsStorageStore.GetByIDorName(ctx, ID, userID)
	if serr != nil {
		s.logger.Errorf("failed to get workspace storage: %v", serr)
		return nil, serr
	}

	if err := s.validateUpdateSpec(spec, existingStorage); err != nil {
		return nil, err
	}

	var updatedStorage *models.WorkspaceStorage
	updateErr := s.wsStorageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		updatedStorage, serr = s.wsStorageStore.UpdateWithTx(ctx, existingStorage.ID, spec)
		if serr != nil {
			s.logger.Errorf("failed to update workspace storage: %v", serr)
			return serr
		}
		if err := s.clusterResourceService.UpsertWorkspaceStorageInCluster(ctx, updatedStorage); err != nil {
			s.logger.Errorf("failed to update workspace storage in cluster: %v", err)
			return errors.GeneralError("failed to update workspace storage in cluster: %s", err.Error())
		}
		return nil
	})

	if updateErr != nil {
		s.logger.Errorf("failed to update workspace storage: %v", updateErr)
		return nil, updateErr
	}
	return updatedStorage, nil
}

func allowedVolumeUpdate(existingVolume *models.Volume, volume *models.Volume) bool {
	return existingVolume.Size == volume.Size && existingVolume.StorageClass == volume.StorageClass && volume.Name == existingVolume.Name
}

func (s *workspaceStorageService) UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStorageStatus) *errors.ServiceError {
	_, err := s.wsStorageStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		s.logger.Errorf("failed to update workspace storage status: %v", err)
		return err
	}
	return nil
}

func (s *workspaceStorageService) Delete(ctx context.Context, ID string, userID string) *errors.ServiceError {
	deleteErr := s.wsStorageStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		existingStorage, serr := s.wsStorageStore.GetByIDorName(ctx, ID, userID)
		if serr != nil {
			s.logger.Errorf("failed to get workspace storage: %v", serr)
			return serr
		}
		if err := s.clusterResourceService.DeleteWorkspaceStorageInCluster(ctx, existingStorage); err != nil {
			s.logger.Errorf("failed to delete workspace storage in cluster: %v", err)
			return errors.GeneralError("failed to delete workspace storage in cluster: %s", err.Error())
		}

		if err := s.wsStorageStore.DeleteWithTx(ctx, existingStorage.ID); err != nil {
			s.logger.Errorf("failed to delete workspace storage: %v", err)
			return err
		}
		return nil
	})
	if deleteErr != nil {
		s.logger.Errorf("failed to delete workspace storage: %v", deleteErr)
		return deleteErr
	}
	return nil
}
