package services

//go:generate mockgen -source=volume_service.go -destination=../mocks/mock_volume_service.go -package=mocks

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/auth"
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
	CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateInDbWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError)
	CreateInCluster(ctx context.Context, spec *models.Volume) *errors.ServiceError
	ListByUserID(ctx context.Context, projectID, userID string) ([]*models.Volume, *errors.ServiceError)
	UpdateStatus(ctx context.Context, ID string, status *models.VolumeStatus) *errors.ServiceError
	UpdateGitRepoSourceRevision(ctx context.Context, ID string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError)
	UpdateRemoteSourceRevision(ctx context.Context, ID string, revision models.RemoteDirSource) (*models.Volume, *errors.ServiceError)
	InjectClusterResourceService(volumeClusterService clusterresource.VolumeClusterResourceService)
	Delete(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteFromDB(ctx context.Context, ID string) *errors.ServiceError
	InternalDeleteVolumesUsedByStackFromDB(ctx context.Context, stackID string) *errors.ServiceError
	DeleteWithTx(ctx context.Context, ID string) *errors.ServiceError
	ListVolumesUsedByStack(ctx context.Context, stackID string) ([]*models.Volume, *errors.ServiceError)
	InternalCreateWithTx(ctx context.Context, stack *models.Stack, volume *models.Volume) (*models.Volume, *errors.ServiceError)
	InternalSyncVolumesWithTx(ctx context.Context, stack *models.Stack, existingStack *models.Stack, desired []*models.Volume) *errors.ServiceError
}

type VolumeServiceSpec struct {
	SessionFactory   db.SessionFactory
	ReferenceService ReferenceService
	Logger           logger.Logger
	Permissions      auth.PermissionService
	RuntimePolicy    RuntimePolicy
	ClusterService   ClusterService
	ClusterWrites    ClusterMutationCoordinator
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
		releaseStore: pgstore.NewStackReleaseStore(pgstore.StackReleaseStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		stackStore:     pgstore.NewStackStore(&pgstore.StackStoreSpec{SessionFactory: spec.SessionFactory}),
		logger:         spec.Logger,
		permissions:    spec.Permissions,
		runtimePolicy:  spec.RuntimePolicy,
		clusterService: spec.ClusterService,
		clusterWrites:  spec.ClusterWrites,
	}
}

type volumeService struct {
	volumeStore            stores.VolumeStore
	stackVolumeStore       stores.StackVolumeStore
	referenceService       ReferenceService
	releaseStore           stores.StackReleaseStore
	stackStore             stores.StackStore
	clusterResourceService clusterresource.VolumeClusterResourceService
	logger                 logger.Logger
	permissions            auth.PermissionService
	runtimePolicy          RuntimePolicy
	clusterService         ClusterService
	clusterWrites          ClusterMutationCoordinator
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

	createdVolume, err := s.CreateInDbWithTx(ctx, volume)
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

func (s *volumeService) UpdateGitRepoSourceRevision(ctx context.Context, ID string, revision models.GitRepoRevision) (*models.Volume, *errors.ServiceError) {
	var updatedVolume *models.Volume
	var err *errors.ServiceError

	if !revision.IsValid() {
		return nil, errors.BadRequest("invalid git repo revision")
	}
	updateErr := s.volumeStore.WithTransaction(ctx, func(ctx context.Context) *errors.ServiceError {
		volume, getErr := s.volumeStore.GetByID(ctx, ID)
		if getErr != nil {
			return getErr
		}
		admission, policyErr := s.runtimePolicy.AdmitMutationWithTx(ctx, volume.OrganisationID)
		if policyErr != nil {
			return policyErr
		}
		updatedVolume, err = s.volumeStore.UpdateGitRepoSourceRevisionWithTx(ctx, ID, revision)
		if err != nil {
			s.logger.Error(ctx, "failed to update volume git repo source revision: %v", err)
			return err
		}
		reconcileCluster, stack, release, admissionErr := s.resolveVolumeRevisionAuthority(ctx, updatedVolume.ID, admission)
		if admissionErr != nil {
			return admissionErr
		}
		if !reconcileCluster {
			return nil
		}
		unlockCluster, authorized, authorityErr := s.lockAndRevalidateVolumeRevisionAuthority(ctx, updatedVolume, stack, release)
		if authorityErr != nil {
			return authorityErr
		}
		defer unlockCluster()
		if !authorized {
			return nil
		}
		cErr := s.clusterResourceService.UpdateVolumeGitRevisionInCluster(ctx, updatedVolume)
		if cErr != nil {
			s.logger.Error(ctx, "failed to update volume git revision in cluster: %v", cErr)
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
		volume, getErr := s.volumeStore.GetByID(ctx, ID)
		if getErr != nil {
			return getErr
		}
		admission, policyErr := s.runtimePolicy.AdmitMutationWithTx(ctx, volume.OrganisationID)
		if policyErr != nil {
			return policyErr
		}
		updatedVolume, err = s.volumeStore.UpdateRemoteDirSourceHashWithTx(ctx, ID, revision.CurrentDirectoryHash)
		if err != nil {
			s.logger.Error(ctx, "failed to update volume remote source revision: %v", err)
			return err
		}

		reconcileCluster, stack, release, admissionErr := s.resolveVolumeRevisionAuthority(ctx, updatedVolume.ID, admission)
		if admissionErr != nil {
			return admissionErr
		}
		if !reconcileCluster {
			return nil
		}
		unlockCluster, authorized, authorityErr := s.lockAndRevalidateVolumeRevisionAuthority(ctx, updatedVolume, stack, release)
		if authorityErr != nil {
			return authorityErr
		}
		defer unlockCluster()
		if !authorized {
			return nil
		}
		cErr := s.clusterResourceService.UpdateVolumeRemoteDirRevisionInCluster(ctx, updatedVolume)
		if cErr != nil {
			s.logger.Error(ctx, "failed to update volume git revision in cluster: %v", cErr)
			return errors.GeneralError("failed to update volume git revision in cluster: %s", cErr.Error())
		}
		return nil
	})
	if updateErr != nil {
		return nil, updateErr
	}
	return updatedVolume, nil
}

// shouldReconcileVolumeRevisionWithCluster keeps cloud draft revisions in the
// database until the volume is part of a released workload. Self-hosted mode
// remains eager.
func (s *volumeService) resolveVolumeRevisionAuthority(
	ctx context.Context,
	volumeID string,
	admission MutationAdmission,
) (bool, *models.Stack, *models.StackRelease, *errors.ServiceError) {
	if !admission.ReconcileCluster {
		return false, nil, nil, nil
	}
	if s.runtimePolicy.DraftProvisioningMode() == ProvisioningModeEager {
		return true, nil, nil, nil
	}

	inUse, refs, serr := s.referenceService.IsReferentInUse(ctx, models.ReferentVolume, volumeID)
	if serr != nil {
		return false, nil, nil, serr
	}
	if !inUse {
		return false, nil, nil, nil
	}
	for _, ref := range refs {
		if ref.ReleaseID == nil || *ref.ReleaseID == "" {
			continue
		}
		stack, stackErr := s.stackStore.GetByID(ctx, ref.StackID)
		if stackErr != nil {
			return false, nil, nil, stackErr
		}
		if stack.DeletionTimestamp != nil {
			continue
		}
		release, releaseErr := resolveAuthoritativeWorkloadRelease(ctx, s.releaseStore, stack)
		if releaseErr != nil {
			return false, nil, nil, releaseErr
		}
		if release != nil && release.ID == *ref.ReleaseID {
			return true, stack, release, nil
		}
	}
	return false, nil, nil, nil
}

func (s *volumeService) lockAndRevalidateVolumeRevisionAuthority(
	ctx context.Context,
	volume *models.Volume,
	stack *models.Stack,
	release *models.StackRelease,
) (func(), bool, *errors.ServiceError) {
	clusterID := ""
	if stack == nil {
		cluster, serr := s.clusterService.GetClusterForOrg(ctx, volume.OrganisationID)
		if serr != nil {
			return nil, false, serr
		}
		clusterID = cluster.ID
	} else {
		clusterID = stack.ClusterID
	}
	unlockCluster := s.clusterWrites.LockClusterNamespace(clusterID, volume.Namespace)

	if stack == nil {
		return unlockCluster, true, nil
	}
	if lockErr := s.stackStore.LockByID(ctx, stack.ID); lockErr != nil {
		unlockCluster()
		return nil, false, lockErr
	}
	currentStack, serr := s.stackStore.GetByID(ctx, stack.ID)
	if serr != nil {
		unlockCluster()
		return nil, false, serr
	}
	if currentStack.DeletionTimestamp != nil {
		return unlockCluster, false, nil
	}
	currentRelease, serr := resolveAuthoritativeWorkloadRelease(ctx, s.releaseStore, currentStack)
	if serr != nil {
		unlockCluster()
		return nil, false, serr
	}
	if currentRelease == nil || currentRelease.ID != release.ID {
		return unlockCluster, false, nil
	}
	if lockErr := s.releaseStore.LockByID(ctx, currentRelease.ID); lockErr != nil {
		unlockCluster()
		return nil, false, lockErr
	}
	currentStack, serr = s.stackStore.GetByID(ctx, stack.ID)
	if serr != nil {
		unlockCluster()
		return nil, false, serr
	}
	if currentStack.DeletionTimestamp != nil {
		return unlockCluster, false, nil
	}
	currentRelease, serr = resolveAuthoritativeWorkloadRelease(ctx, s.releaseStore, currentStack)
	if serr != nil {
		unlockCluster()
		return nil, false, serr
	}
	if currentRelease == nil || currentRelease.ID != release.ID {
		return unlockCluster, false, nil
	}
	admission, serr := s.runtimePolicy.AdmitMutationWithTx(ctx, currentStack.OrganisationID)
	if serr != nil {
		unlockCluster()
		return nil, false, serr
	}
	if !admission.ReconcileCluster {
		unlockCluster()
		return nil, false, errors.TrialInactive()
	}
	return unlockCluster, true, nil
}

// Assume ctx already has a transaction
func (s *volumeService) CreateWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	var createdVolume *models.Volume
	var err *errors.ServiceError
	createdVolume, err = s.volumeStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to create volume: %v", err)
		return nil, err
	}
	cerr := s.clusterResourceService.CreateVolumeInCluster(ctx, createdVolume)
	if cerr != nil {
		s.logger.Error(ctx, "failed to create volume in cluster: %v", cerr)
		return nil, errors.GeneralError("failed to create volume in cluster: %s", cerr.Error())
	}
	return createdVolume, nil
}

func (s *volumeService) CreateInCluster(ctx context.Context, spec *models.Volume) *errors.ServiceError {
	volume, err := s.volumeStore.GetByID(ctx, spec.ID)
	if err != nil {
		s.logger.Error(ctx, "failed to get volume for creation in cluster: %v", err)
		return err
	}
	if volume == nil {
		s.logger.Error(ctx, "volume not found for creation in cluster: %v", spec.ID)
		return errors.NotFound("volume not found")
	}
	cErr := s.clusterResourceService.CreateVolumeInCluster(ctx, volume)
	if cErr != nil {
		s.logger.Error(ctx, "failed to create volume in cluster: %v", cErr)
		return errors.GeneralError("failed to create volume in cluster: %s", cErr.Error())
	}
	return nil
}

func (s *volumeService) CreateInDbWithTx(ctx context.Context, spec *models.Volume) (*models.Volume, *errors.ServiceError) {
	var createdVolume *models.Volume
	var err *errors.ServiceError
	createdVolume, err = s.volumeStore.CreateWithTx(ctx, spec)
	if err != nil {
		s.logger.Error(ctx, "failed to create volume: %v", err)
		return nil, err
	}
	return createdVolume, nil
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
