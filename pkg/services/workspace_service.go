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

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	GetWorkspace(ctx context.Context, ID string) (*models.Workspace, *errors.ServiceError)
	GetWorkspaceByName(ctx context.Context, name string, userID string) (*models.Workspace, *errors.ServiceError)
	GetWorkspacesByUserID(ctx context.Context, userID string) ([]*models.Workspace, *errors.ServiceError)
	GetWorkspacesByOrganisationID(ctx context.Context, organisationID string) ([]*models.Workspace, *errors.ServiceError)
	UpdateWorkspace(ctx context.Context, ID string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStatus) *errors.ServiceError
	DeleteWorkspace(ctx context.Context, ID string) *errors.ServiceError
	InjectClusterResourceService(workspaceClusterService clusterresource.ClusterWorkspaceService)
}

type WorkspaceServiceSpec struct {
	SessionFactory          db.SessionFactory
	WorkspaceUserService    WorkspaceUserService
	WorkspaceStorageService WorkspaceStorageService
	ClusterService          ClusterService
	Logger                  logger.Logger
}

type workspaceService struct {
	workspaceStore          stores.WorkspaceStore
	logger                  logger.Logger
	sessionFactory          db.SessionFactory
	workspaceUserService    WorkspaceUserService
	clusterResourceService  clusterresource.ClusterWorkspaceService
	workspaceStorageService WorkspaceStorageService
}

func NewWorkspaceService(spec WorkspaceServiceSpec) WorkspaceService {
	return &workspaceService{
		workspaceStore: pgstore.NewWorkspaceStore(&pgstore.WorkspaceStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		workspaceUserService:    spec.WorkspaceUserService,
		workspaceStorageService: spec.WorkspaceStorageService,
		logger:                  spec.Logger,
		sessionFactory:          spec.SessionFactory,
	}
}

func (s *workspaceService) InjectClusterResourceService(workspaceClusterService clusterresource.ClusterWorkspaceService) {
	s.clusterResourceService = workspaceClusterService
}

func (s *workspaceService) CreateWorkspace(ctx context.Context, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	workspaceUser, err := s.workspaceUserService.GetWorkspaceUser(ctx, spec.UserID)
	if err != nil {
		return nil, errors.GeneralError("failed to get workspaceuser for user '%s': %s", spec.UserID, err.Error())
	}
	userWorkspaceNamespaceMap := workspaceUser.WorkspaceNamespaceMap()
	if workpaceNamepace, ok := userWorkspaceNamespaceMap[spec.Name]; ok {
		spec.Namespace = workpaceNamepace.Namespace
	} else {
		return nil, errors.BadRequest("workspace with name '%s' not defined for the user", spec.Name)
	}

	for i := range spec.WorkspaceResources {
		spec.WorkspaceResources[i].UserID = spec.UserID
	}

	if spec.HasVolumeMounts() {
		currentUserWorkpaceStorage, err := s.workspaceStorageService.GetbyWorkspaceName(ctx, spec.Name, spec.UserID)
		if err != nil {
			return nil, errors.GeneralError("failed to get workspace storage for workspace '%s': %s", spec.Name, err.Error())
		}
		if err := s.validateWorkspaceVolumeMounts(currentUserWorkpaceStorage, spec); err != nil {
			return nil, err
		}

		setWorkspaceStorageAssociation(spec, currentUserWorkpaceStorage)
	}

	var createdWorkspace *models.Workspace
	err = s.workspaceStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var creatErr *errors.ServiceError
		createdWorkspace, creatErr = s.workspaceStore.CreateWithTx(ctx, spec)
		if creatErr != nil {
			return creatErr
		}
		if err := s.clusterResourceService.CreateWorkspaceInCluster(ctx, createdWorkspace); err != nil {
			return errors.GeneralError("failed to create workspace in cluster: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return createdWorkspace, nil
}

func (s *workspaceService) GetWorkspace(ctx context.Context, ID string) (*models.Workspace, *errors.ServiceError) {
	workspace, err := s.workspaceStore.GetByID(ctx, ID)
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

func (s *workspaceService) GetWorkspaceByName(ctx context.Context, name string, userID string) (*models.Workspace, *errors.ServiceError) {
	workspace, err := s.workspaceStore.GetByName(ctx, name, userID)
	if err != nil {
		return nil, err
	}
	return workspace, nil
}

func (s *workspaceService) GetWorkspacesByUserID(ctx context.Context, userID string) ([]*models.Workspace, *errors.ServiceError) {
	workspaces, err := s.workspaceStore.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (s *workspaceService) GetWorkspacesByOrganisationID(ctx context.Context, organisationID string) ([]*models.Workspace, *errors.ServiceError) {
	workspaces, err := s.workspaceStore.ListByOrganisationID(ctx, organisationID)
	if err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (s *workspaceService) DeleteWorkspace(ctx context.Context, ID string) *errors.ServiceError {
	workspace, err := s.GetWorkspace(ctx, ID)
	if err != nil {
		return err
	}
	err = s.workspaceStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		if err := s.clusterResourceService.DeleteWorkspaceInCluster(ctx, workspace); err != nil {
			return errors.GeneralError("failed to delete workspace in cluster: %s", err.Error())
		}
		if err := s.workspaceStore.DeleteWithTx(ctx, ID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *workspaceService) UpdateWorkspace(ctx context.Context, ID string, spec *models.Workspace) (*models.Workspace, *errors.ServiceError) {
	workspace, err := s.GetWorkspace(ctx, ID)
	if err != nil {
		return nil, err
	}
	if spec.Name != workspace.Name {
		return nil, errors.BadRequest("workspace name cannot be updated")
	}
	if spec.UserID != workspace.UserID {
		return nil, errors.BadRequest("workspace user cannot be updated")
	}
	if spec.OrganisationID != workspace.OrganisationID {
		return nil, errors.BadRequest("workspace organisation cannot be updated")
	}
	spec.Namespace = workspace.Namespace

	if spec.HasVolumeMounts() {
		currentUserWorkpaceStorage, err := s.workspaceStorageService.GetbyWorkspaceName(ctx, spec.Name, spec.UserID)
		if err != nil {
			return nil, errors.GeneralError("failed to get workspace storage for workspace '%s': %s", spec.Name, err.Error())
		}
		if err := s.validateWorkspaceVolumeMounts(currentUserWorkpaceStorage, spec); err != nil {
			return nil, err
		}

		setWorkspaceStorageAssociation(spec, currentUserWorkpaceStorage)
	}

	var updatedWorkspace *models.Workspace
	err = s.workspaceStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		var updateErr *errors.ServiceError
		updatedWorkspace, updateErr = s.workspaceStore.UpdateWithTx(ctx, ID, spec)
		if updateErr != nil {
			return updateErr
		}
		if err := s.clusterResourceService.UpdateWorkspaceInCluster(ctx, updatedWorkspace); err != nil {
			return errors.GeneralError("failed to update workspace in cluster: %s", err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updatedWorkspace, nil
}

func (s *workspaceService) UpdateStatus(ctx context.Context, ID string, status *models.WorkspaceStatus) *errors.ServiceError {
	err := s.workspaceStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		return err
	}
	return nil
}

func (s *workspaceService) validateWorkspaceVolumeMounts(currentUserWorkpaceStorage *models.WorkspaceStorage, spec *models.Workspace) *errors.ServiceError {
	for i := range spec.WorkspaceResources {
		currentResource := spec.WorkspaceResources[i]
		for j := range spec.WorkspaceResources[i].VolumeMounts {
			currentVolumeMount := currentResource.VolumeMounts[j]
			if !currentUserWorkpaceStorage.VolumeExists(currentVolumeMount.SourceVolumeID) {
				return errors.BadRequest("volume '%s' not defined in the users workspace storage", currentVolumeMount.SourceVolumeID)
			}
		}
	}
	return nil
}

func setWorkspaceStorageAssociation(workspace *models.Workspace, currentUserWorkpaceStorage *models.WorkspaceStorage) {
	volumeMap := currentUserWorkpaceStorage.VolumeMap()
	for i := range workspace.WorkspaceResources {
		for j := range workspace.WorkspaceResources[i].VolumeMounts {
			sourceVolumeID := workspace.WorkspaceResources[i].VolumeMounts[j].SourceVolumeID
			workspace.WorkspaceResources[i].VolumeMounts[j].WorkspaceStorageID = currentUserWorkpaceStorage.ID
			workspace.WorkspaceResources[i].VolumeMounts[j].SourceVolumeType = volumeMap[sourceVolumeID].VolumeSourceType()
		}
	}
}
