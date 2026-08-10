package services

//go:generate mockgen -source=volume_service.go -destination=../mocks/mock_volume_service.go -package=mocks

import (
	"context"
	"github.com/Stackdome/stackdome/pkg/auth"
	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
	"github.com/Stackdome/stackdome/pkg/stores"
	"github.com/Stackdome/stackdome/pkg/stores/pgstore"
)

type VolumeService interface {
	Get(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	InternalGet(ctx context.Context, ID string) (*models.Volume, *errors.ServiceError)
	GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError)
	InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError)
	InternalListNotReady(ctx context.Context) ([]*models.Volume, *errors.ServiceError)
	ListByUserID(ctx context.Context, projectID, userID string) ([]*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError
	InjectClusterResourceService(volumeClusterService clusterresource.VolumeClusterResourceService)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, stack *models.Stack, volume *models.Volume) (*models.Volume, *errors.ServiceError)
	PrepareForCreate(ctx context.Context, volume *models.Volume) *errors.ServiceError
	InternalSyncVolumesWithTx(ctx context.Context, stack *models.Stack, existingStack *models.Stack, desired []*models.Volume) *errors.ServiceError
}

// PrepareForCreate resolves mutable Git references before the volume enters a
// transaction. The persisted volume definition is then stable across retries.
func (s *volumeService) PrepareForCreate(ctx context.Context, volume *models.Volume) *errors.ServiceError {
	if volume.VolumeSource == nil || volume.VolumeSource.GitRepoSource == nil || volume.VolumeSource.GitRepoSource.Revision.Commit != "" {
		return nil
	}
	source := volume.VolumeSource.GitRepoSource
	client, err := gitclient.NewGitClientForRepo(source.RepoUrl, gitclient.GitCredentials{})
	if err != nil {
		return errors.BadRequest("failed to create Git client for volume '%s': %s", volume.Name, err.Error())
	}
	resolved, err := gitclient.ResolveGitRepoRevision(ctx, client, source.RepoUrl, source.Revision)
	if err != nil {
		return errors.BadRequest("failed to resolve Git revision for volume '%s': %s", volume.Name, err.Error())
	}
	if resolved.Commit == "" {
		return errors.GeneralError("resolved Git revision for volume '%s' has no commit", volume.Name)
	}
	volume.VolumeSource.GitRepoSource.Revision = resolved
	return nil
}

type VolumeServiceSpec struct {
	SessionFactory   db.SessionFactory
	ReferenceService ReferenceService
	RuntimePolicy    RuntimePolicy
	Logger           logger.Logger
	Permissions      auth.PermissionService
}

func NewVolumeService(spec VolumeServiceSpec) VolumeService {
	return &volumeService{
		volumeStore: pgstore.NewVolumeStore(pgstore.VolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackVolumeStore: pgstore.NewStackVolumeStore(pgstore.StackVolumeStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		referenceService: spec.ReferenceService,
		runtimePolicy:    spec.RuntimePolicy,
		logger:           spec.Logger,
		permissions:      spec.Permissions,
	}
}

type volumeService struct {
	volumeStore            stores.VolumeStore
	stackVolumeStore       stores.StackVolumeStore
	referenceService       ReferenceService
	runtimePolicy          RuntimePolicy
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
		s.logger.Error(ctx, "failed to get volume: %v", err)
		return nil, err
	}
	if permErr := s.permissions.Check(ctx, volume.ProjectID, auth.ResourceVolumes, ID, auth.ActionRead); permErr != nil {
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
		s.logger.Error(ctx, "failed to list volumes used by stack: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) InternalCreateWithTx(ctx context.Context, stack *models.Stack, volume *models.Volume) (*models.Volume, *errors.ServiceError) {
	volume.NamespaceID = stack.NamespaceID
	volume.OrganisationID = stack.OrganisationID
	volume.ProjectID = stack.ProjectID
	volume.UserID = stack.UserID
	volume.Namespace = stack.Namespace
	if policyErr := s.runtimePolicy.EnsureComputeAccess(ctx, stack.OrganisationID); policyErr != nil {
		return nil, policyErr
	}
	if policyErr := s.runtimePolicy.ValidateVolumeLimits(ctx, stack.OrganisationID, volume.Size); policyErr != nil {
		return nil, policyErr
	}

	createdVolume, err := s.volumeStore.CreateWithTx(ctx, volume)
	if err != nil {
		return nil, errors.GeneralError("failed to create volume '%s': %s", volume.Name, err.Error())
	}
	if err := s.associateVolumeWithStackWithTx(ctx, createdVolume.ID, stack.ID); err != nil {
		return nil, err
	}
	return createdVolume, nil
}

func (s *volumeService) InternalSyncVolumesWithTx(ctx context.Context, stack *models.Stack, existingStack *models.Stack, desired []*models.Volume) *errors.ServiceError {
	existingVolumesMap := existingStack.VolumesMap()
	for _, volume := range desired {
		if _, found := existingVolumesMap[volume.Name]; found {
			continue
		}
		if _, err := s.InternalCreateWithTx(ctx, stack, volume); err != nil {
			return err
		}
	}
	return nil
}

func (s *volumeService) associateVolumeWithStackWithTx(ctx context.Context, volumeID, stackID string) *errors.ServiceError {
	return s.stackVolumeStore.CreateWithTx(ctx, &models.StackVolume{
		VolumeID: volumeID,
		StackID:  stackID,
	})
}

func (s *volumeService) GetByVolumeNameAndNamespace(ctx context.Context, volumeName, namespace string) (*models.Volume, *errors.ServiceError) {
	volume, err := s.volumeStore.GetByVolumeNameAndNamespace(ctx, volumeName, namespace)
	if err != nil {
		s.logger.Error(ctx, "failed to get volume by name and namespace: %v", err)
		return nil, err
	}
	if volume == nil {
		return nil, errors.NotFound("volume not found")
	}
	return volume, nil
}

func (s *volumeService) ListByUserID(ctx context.Context, projectID, userID string) ([]*models.Volume, *errors.ServiceError) {
	if permErr := s.permissions.Check(ctx, projectID, auth.ResourceVolumes, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	volumes, err := s.volumeStore.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error(ctx, "failed to list volumes by user ID: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) InternalList(ctx context.Context, ids []string) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.volumeStore.InternalList(ctx, ids)
	if err != nil {
		s.logger.Error(ctx, "failed to list volumes: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) InternalListNotReady(ctx context.Context) ([]*models.Volume, *errors.ServiceError) {
	volumes, err := s.volumeStore.InternalListNotReady(ctx)
	if err != nil {
		s.logger.Error(ctx, "failed to list not-ready volumes: %v", err)
		return nil, err
	}
	return volumes, nil
}

func (s *volumeService) UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError {
	err := s.volumeStore.UpdateStatus(ctx, ID, status)
	if err != nil {
		s.logger.Error(ctx, "failed to update volume status: %v", err)
		return err
	}
	return nil
}

func (s *volumeService) Delete(ctx context.Context, ID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get volume for deletion: %v", err)
		return err
	}

	if permErr := s.permissions.Check(ctx, volume.ProjectID, auth.ResourceVolumes, ID, auth.ActionDelete); permErr != nil {
		return permErr
	}
	inUse, refs, refErr := s.referenceService.IsReferentInUse(ctx, models.ReferentVolume, volume.ID)
	if refErr != nil {
		return refErr
	}
	if inUse {
		return errors.Conflict("volume '%s' is in use by %s and cannot be deleted", volume.ID, describeReferences(refs))
	}

	deleteErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		cErr := s.clusterResourceService.DeleteVolumeInCluster(ctx, volume)
		if cErr != nil {
			s.logger.Error(ctx, "failed to delete volume in cluster: %v", cErr)
			return errors.GeneralError("failed to delete volume in cluster: %s", cErr.Error())
		}
		err = s.volumeStore.DeleteWithTx(ctx, ID)
		if err != nil {
			s.logger.Error(ctx, "failed to delete volume: %v", err)
			return err
		}
		return nil
	})
	if deleteErr != nil {
		s.logger.Error(ctx, "failed to delete volume: %v", deleteErr)
		return deleteErr
	}

	return nil
}

func (s *volumeService) InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, ID)
	if err != nil {
		return err
	}
	inUse, refs, refErr := s.referenceService.IsReferentInUse(ctx, models.ReferentVolume, volume.ID)
	if refErr != nil {
		return refErr
	}
	if inUse {
		return errors.Conflict("volume '%s' is in use by %s and cannot be deleted", volume.ID, describeReferences(refs))
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
		s.logger.Error(ctx, "failed to get volume for deletion: %v", err)
		return err
	}
	inUse, refs, refErr := s.referenceService.IsReferentInUse(ctx, models.ReferentVolume, volume.ID)
	if refErr != nil {
		return refErr
	}
	if inUse {
		return errors.Conflict("volume '%s' is in use by %s and cannot be deleted", volume.ID, describeReferences(refs))
	}

	cErr := s.clusterResourceService.DeleteVolumeInCluster(ctx, volume)
	if cErr != nil {
		s.logger.Error(ctx, "failed to delete volume in cluster: %v", cErr)
		return errors.GeneralError("failed to delete volume in cluster: %s", cErr.Error())
	}
	err = s.volumeStore.DeleteWithTx(ctx, ID)
	if err != nil {
		s.logger.Error(ctx, "failed to delete volume: %v", err)
		return err
	}
	return nil
}
