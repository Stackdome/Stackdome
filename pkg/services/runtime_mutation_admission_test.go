package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cloud runtime mutation admission", func() {
	It("rejects a stack shell update before persistence", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockStackStore(ctrl)
		svc := &stackService{stackStore: store, runtimePolicy: policy}
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive())

		updated, serr := svc.InternalUpdateShellWithTx(context.Background(), &models.Stack{}, &models.Stack{ID: "stack-1", OrganisationID: "org-1"})

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
	})

	It("rejects connection creation inside its write transaction", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockStackStore(ctrl)
		svc := &stackService{stackStore: store, runtimePolicy: policy}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive())

		created, serr := svc.createStackConnection(context.Background(), &models.Stack{ID: "stack-1", OrganisationID: "org-1"}, &models.StackConnection{})

		Expect(created).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
	})

	It("rejects resource restart before its direct Kubernetes path", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		resources := mocks.NewMockStackResourceStore(ctrl)
		permissions := mocks.NewMockPermissionService(ctrl)
		svc := &stackResourceService{stackStore: stacks, stackResourceStore: resources, permissions: permissions, runtimePolicy: policy}
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1"}
		stacks.EXPECT().GetByID(gomock.Any(), "stack-1").Return(stack, nil)
		permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceStacks, "stack-1", auth.ActionWrite).Return(nil)
		resources.EXPECT().GetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(&models.StackResource{ID: "resource-1"}, nil)
		stacks.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive())

		updated, serr := svc.Restart(context.Background(), "stack-1", "web")

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
	})

	It("keeps an admitted pre-release resource restart out of Kubernetes", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		resources := mocks.NewMockStackResourceStore(ctrl)
		permissions := mocks.NewMockPermissionService(ctrl)
		clusterManager := mocks.NewMockClusterManager(ctrl)
		svc := &stackResourceService{
			stackStore: stacks, stackResourceStore: resources, permissions: permissions,
			runtimePolicy: policy, clusterManager: clusterManager,
		}
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1",
			ClusterID: "cluster-1", Namespace: "namespace-1",
		}
		resource := &models.StackResource{ID: "resource-1", StackID: stack.ID, Name: "web", Namespace: stack.Namespace}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil)
		permissions.EXPECT().Check(gomock.Any(), stack.ProjectID, auth.ResourceStacks, stack.ID, auth.ActionWrite).Return(nil)
		resources.EXPECT().GetByStackIDAndResourceName(gomock.Any(), stack.ID, resource.Name).Return(resource, nil)
		stacks.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), stack.OrganisationID).
			Return(ComputeMutationAdmission{ReconcileCluster: false}, nil)
		resources.EXPECT().UpdateWithTx(gomock.Any(), resource.ID, gomock.Any(), stack).Return(resource, nil)

		updated, serr := svc.Restart(context.Background(), stack.ID, resource.Name)

		Expect(serr).To(BeNil())
		Expect(updated).To(BeIdenticalTo(resource))
	})

	It("fences a resource restart and rechecks stack deletion before Kubernetes", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		resources := mocks.NewMockStackResourceStore(ctrl)
		permissions := mocks.NewMockPermissionService(ctrl)
		clusterManager := mocks.NewMockClusterManager(ctrl)
		clusterWrites := &recordingClusterMutationCoordinator{}
		svc := &stackResourceService{
			stackStore: stacks, stackResourceStore: resources, permissions: permissions,
			runtimePolicy: policy, clusterManager: clusterManager, clusterWrites: clusterWrites,
		}
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1",
			ClusterID: "cluster-1", Namespace: "namespace-1",
		}
		deleting := *stack
		deletingAt := time.Now().UTC()
		deleting.DeletionTimestamp = &deletingAt
		resource := &models.StackResource{ID: "resource-1", StackID: stack.ID, Name: "web", Namespace: stack.Namespace}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil)
		permissions.EXPECT().Check(gomock.Any(), stack.ProjectID, auth.ResourceStacks, stack.ID, auth.ActionWrite).Return(nil)
		resources.EXPECT().GetByStackIDAndResourceName(gomock.Any(), stack.ID, resource.Name).Return(resource, nil)
		stacks.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), stack.OrganisationID).
			Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		resources.EXPECT().UpdateWithTx(gomock.Any(), resource.ID, gomock.Any(), stack).Return(resource, nil)
		stacks.EXPECT().LockByID(gomock.Any(), stack.ID).Return(nil)
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(&deleting, nil)

		updated, serr := svc.Restart(context.Background(), stack.ID, resource.Name)

		Expect(serr).To(BeNil())
		Expect(updated).To(BeIdenticalTo(resource))
		Expect(clusterWrites.lockedClusterID).To(Equal(stack.ClusterID))
		Expect(clusterWrites.lockedNamespace).To(Equal(stack.Namespace))
		Expect(clusterWrites.unlocked).To(BeTrue())
	})

	It("rejects a volume source revision before database or cluster mutation", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		svc := &volumeService{volumeStore: store, runtimePolicy: policy}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive())

		updated, serr := svc.UpdateGitRepoSourceRevision(context.Background(), "volume-1", models.GitRepoRevision{Branch: "main"})

		Expect(updated).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
	})

	It("updates a pre-release volume revision in the database without touching Kubernetes", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		svc := &volumeService{volumeStore: store, runtimePolicy: policy}
		revision := models.GitRepoRevision{Branch: "main"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: false}, nil)
		store.EXPECT().UpdateGitRepoSourceRevisionWithTx(gomock.Any(), "volume-1", revision).Return(updated, nil)

		result, serr := svc.UpdateGitRepoSourceRevision(context.Background(), "volume-1", revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
	})

	It("updates an active released volume revision in the database and Kubernetes", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		clusterWrites := &recordingClusterMutationCoordinator{}
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
			clusterWrites: clusterWrites,
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1", Namespace: "namespace-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: true}, nil).Times(2)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), "volume-1", "new-hash").Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseID := "release-1"
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: updated.OrganisationID, ClusterID: "cluster-1", Namespace: updated.Namespace,
			Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseID}},
		}
		release := &models.StackRelease{
			ID: releaseID, StackID: stack.ID, State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{
				ID: stack.ID, OrganisationID: stack.OrganisationID, ClusterID: stack.ClusterID, Namespace: stack.Namespace,
			}},
		}
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, "volume-1").Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil).Times(3)
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(release, nil).Times(3)
		releases.EXPECT().GetByID(gomock.Any(), releaseID).Return(release, nil).Times(3)
		stacks.EXPECT().LockByID(gomock.Any(), stack.ID).Return(nil)
		releases.EXPECT().LockByID(gomock.Any(), releaseID).Return(nil)
		clusterResources.EXPECT().UpdateVolumeRemoteDirRevisionInCluster(gomock.Any(), updated).Return(nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), "volume-1", revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
		Expect(clusterWrites.lockedClusterID).To(Equal(stack.ClusterID))
		Expect(clusterWrites.lockedNamespace).To(Equal(updated.Namespace))
		Expect(clusterWrites.unlocked).To(BeTrue())
	})

	It("does not mutate Kubernetes for a retained historical volume reference", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), "volume-1", "new-hash").Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseID := "release-a"
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, "volume-1").Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		stack := &models.Stack{ID: "stack-1", Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-b"}}}
		latest := &models.StackRelease{ID: releaseID, StackID: stack.ID, State: models.ReleaseStateReleased}
		converged := &models.StackRelease{
			ID: "release-b", StackID: stack.ID, State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: stack.ID}},
		}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil)
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(latest, nil)
		releases.EXPECT().GetByID(gomock.Any(), converged.ID).Return(converged, nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), "volume-1", revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
	})

	It("does not let a converged volume reference override a newer active release", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		store.EXPECT().GetByID(gomock.Any(), updated.ID).Return(updated, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), updated.OrganisationID).Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), updated.ID, revision.CurrentDirectoryHash).Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseAID := "release-a"
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, updated.ID).
			Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseAID}}, nil)
		stack := &models.Stack{ID: "stack-1", OrganisationID: updated.OrganisationID, Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseAID}}}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil)
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(&models.StackRelease{
			ID: "release-b", StackID: stack.ID, State: models.ReleaseStateInProgress,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: stack.ID, OrganisationID: stack.OrganisationID}},
		}, nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), updated.ID, revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
	})

	It("uses a superseded LastConverged release after a newer attempt fails", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
			clusterWrites: &recordingClusterMutationCoordinator{},
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		store.EXPECT().GetByID(gomock.Any(), updated.ID).Return(updated, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), updated.OrganisationID).Return(ComputeMutationAdmission{ReconcileCluster: true}, nil).Times(2)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), updated.ID, revision.CurrentDirectoryHash).Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseAID := "release-a"
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, updated.ID).
			Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseAID}}, nil)
		stack := &models.Stack{ID: "stack-1", OrganisationID: updated.OrganisationID, Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseAID}}}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil).Times(3)
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(&models.StackRelease{
			ID: "release-b", StackID: stack.ID, State: models.ReleaseStateFailed,
		}, nil).Times(3)
		releaseA := &models.StackRelease{
			ID: releaseAID, StackID: stack.ID, State: models.ReleaseStateSuperseded,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: stack.ID, OrganisationID: stack.OrganisationID}},
		}
		releases.EXPECT().GetByID(gomock.Any(), releaseAID).Return(releaseA, nil).Times(3)
		stacks.EXPECT().LockByID(gomock.Any(), stack.ID).Return(nil)
		releases.EXPECT().LockByID(gomock.Any(), releaseAID).Return(nil)
		clusterResources.EXPECT().UpdateVolumeRemoteDirRevisionInCluster(gomock.Any(), updated).Return(nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), updated.ID, revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
	})

	It("rechecks allocation immediately before a volume Kubernetes write", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
			clusterWrites: &recordingClusterMutationCoordinator{},
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		store.EXPECT().GetByID(gomock.Any(), updated.ID).Return(updated, nil)
		gomock.InOrder(
			policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), updated.OrganisationID).Return(ComputeMutationAdmission{ReconcileCluster: true}, nil),
			policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), updated.OrganisationID).Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive()),
		)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), updated.ID, revision.CurrentDirectoryHash).Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseID := "release-a"
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, updated.ID).
			Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		stack := &models.Stack{ID: "stack-1", OrganisationID: updated.OrganisationID, Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseID}}}
		stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil).Times(3)
		release := &models.StackRelease{
			ID: releaseID, StackID: stack.ID, State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{ID: stack.ID, OrganisationID: stack.OrganisationID}},
		}
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(release, nil).Times(3)
		releases.EXPECT().GetByID(gomock.Any(), releaseID).Return(release, nil).Times(3)
		stacks.EXPECT().LockByID(gomock.Any(), stack.ID).Return(nil)
		releases.EXPECT().LockByID(gomock.Any(), releaseID).Return(nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), updated.ID, revision)

		Expect(result).To(BeNil())
		Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
	})

	It("rechecks stack deletion under the fence before a volume Kubernetes write", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		clusterResources := mocks.NewMockVolumeClusterResourceService(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		releases := mocks.NewMockStackReleaseStore(ctrl)
		stacks := mocks.NewMockStackStore(ctrl)
		clusterWrites := &recordingClusterMutationCoordinator{}
		svc := &volumeService{
			volumeStore: store, runtimePolicy: policy, clusterResourceService: clusterResources,
			referenceService: references, releaseStore: releases, stackStore: stacks,
			clusterWrites: clusterWrites,
		}
		revision := models.RemoteDirSource{CurrentDirectoryHash: "new-hash"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1", Namespace: "namespace-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			})
		store.EXPECT().GetByID(gomock.Any(), updated.ID).Return(updated, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), updated.OrganisationID).Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		store.EXPECT().UpdateRemoteDirSourceHashWithTx(gomock.Any(), updated.ID, revision.CurrentDirectoryHash).Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		releaseID := "release-a"
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, updated.ID).
			Return(true, []models.ResourceReference{{StackID: "stack-1", ReleaseID: &releaseID}}, nil)
		stack := &models.Stack{
			ID: "stack-1", OrganisationID: updated.OrganisationID, ClusterID: "cluster-1", Namespace: updated.Namespace,
			Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: releaseID}},
		}
		deleting := *stack
		deletingAt := time.Now().UTC()
		deleting.DeletionTimestamp = &deletingAt
		gomock.InOrder(
			stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(stack, nil),
			stacks.EXPECT().LockByID(gomock.Any(), stack.ID).Return(nil),
			stacks.EXPECT().GetByID(gomock.Any(), stack.ID).Return(&deleting, nil),
		)
		release := &models.StackRelease{
			ID: releaseID, StackID: stack.ID, State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{Stack: models.StackShellSnapshot{
				ID: stack.ID, OrganisationID: stack.OrganisationID, ClusterID: stack.ClusterID, Namespace: stack.Namespace,
			}},
		}
		releases.EXPECT().GetLatestByStackID(gomock.Any(), stack.ID).Return(release, nil)
		releases.EXPECT().GetByID(gomock.Any(), releaseID).Return(release, nil)

		result, serr := svc.UpdateRemoteSourceRevision(context.Background(), updated.ID, revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
		Expect(clusterWrites.lockedClusterID).To(Equal(stack.ClusterID))
		Expect(clusterWrites.unlocked).To(BeTrue())
	})

	It("keeps an unrelated cloud draft volume database-only while another allocation is active", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockVolumeStore(ctrl)
		references := mocks.NewMockReferenceService(ctrl)
		svc := &volumeService{volumeStore: store, runtimePolicy: policy, referenceService: references}
		revision := models.GitRepoRevision{Branch: "main"}
		updated := &models.Volume{ID: "volume-1", OrganisationID: "org-1"}
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		store.EXPECT().GetByID(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1"}, nil)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		store.EXPECT().UpdateGitRepoSourceRevisionWithTx(gomock.Any(), "volume-1", revision).Return(updated, nil)
		policy.EXPECT().DraftProvisioningMode().Return(ProvisioningModeDatabaseOnly)
		references.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, "volume-1").Return(true, []models.ResourceReference{{ReferentID: "volume-1"}}, nil)

		result, serr := svc.UpdateGitRepoSourceRevision(context.Background(), "volume-1", revision)

		Expect(serr).To(BeNil())
		Expect(result).To(BeIdenticalTo(updated))
	})

	DescribeTable("rejects PostgreSQL lifecycle mutation before persistence and enqueue",
		func(invoke func(*postgresAddonService) *errors.ServiceError) {
			ctrl := gomock.NewController(GinkgoT())
			policy := NewMockRuntimePolicy(ctrl)
			store := mocks.NewMockPostgresAddonStore(ctrl)
			permissions := mocks.NewMockPermissionService(ctrl)
			svc := &postgresAddonService{postgresAddonStore: store, permissions: permissions, runtimePolicy: policy}
			store.EXPECT().GetByID(gomock.Any(), "addon-1").Return(&models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1"}, nil)
			permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceAddonsPostgres, "addon-1", auth.ActionWrite).Return(nil)
			store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				},
			)
			policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{}, errors.ComputeAccessInactive())

			serr := invoke(svc)

			Expect(serr.Reason).To(Equal(errors.ErrorCodeComputeAccessInactive))
		},
		Entry("backup", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerBackup(context.Background(), "addon-1")
		}),
		Entry("hibernate", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerHibernate(context.Background(), "addon-1", true)
		}),
		Entry("fence", func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerFence(context.Background(), "addon-1", true)
		}),
	)

	DescribeTable("persists an admitted PostgreSQL lifecycle mutation in the admission transaction",
		func(expectLifecycleUpdate func(*mocks.MockPostgresAddonStore), invoke func(*postgresAddonService) *errors.ServiceError) {
			ctrl := gomock.NewController(GinkgoT())
			policy := NewMockRuntimePolicy(ctrl)
			store := mocks.NewMockPostgresAddonStore(ctrl)
			permissions := mocks.NewMockPermissionService(ctrl)
			enqueuer := mocks.NewMockBackgroundJobEnqueuer(ctrl)
			addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1"}
			svc := &postgresAddonService{
				postgresAddonStore: store, permissions: permissions, runtimePolicy: policy,
				BackgroundJobEnqueuerDep: BackgroundJobEnqueuerDep{BackgroundJobEnqueuer: enqueuer},
			}
			store.EXPECT().GetByID(gomock.Any(), "addon-1").Return(addon, nil)
			permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceAddonsPostgres, "addon-1", auth.ActionWrite).Return(nil)
			store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
				func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
					return fn(ctx)
				},
			)
			policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
			expectLifecycleUpdate(store)
			enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "addon-1"}).Return(nil)

			Expect(invoke(svc)).To(BeNil())
		},
		Entry("hibernate", func(store *mocks.MockPostgresAddonStore) {
			store.EXPECT().SetHibernationWithTx(gomock.Any(), "addon-1", true).Return(&models.PostgresAddon{ID: "addon-1"}, nil)
		}, func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerHibernate(context.Background(), "addon-1", true)
		}),
		Entry("fence", func(store *mocks.MockPostgresAddonStore) {
			store.EXPECT().SetFencingWithTx(gomock.Any(), "addon-1", true).Return(&models.PostgresAddon{ID: "addon-1"}, nil)
		}, func(svc *postgresAddonService) *errors.ServiceError {
			return svc.TriggerFence(context.Background(), "addon-1", true)
		}),
	)

	It("does not enqueue a PostgreSQL lifecycle mutation when the transactional update fails", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockRuntimePolicy(ctrl)
		store := mocks.NewMockPostgresAddonStore(ctrl)
		permissions := mocks.NewMockPermissionService(ctrl)
		svc := &postgresAddonService{postgresAddonStore: store, permissions: permissions, runtimePolicy: policy}
		addon := &models.PostgresAddon{ID: "addon-1", OrganisationID: "org-1", ProjectID: "project-1"}
		store.EXPECT().GetByID(gomock.Any(), "addon-1").Return(addon, nil)
		permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourceAddonsPostgres, "addon-1", auth.ActionWrite).Return(nil)
		store.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			},
		)
		policy.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1").Return(ComputeMutationAdmission{ReconcileCluster: true}, nil)
		store.EXPECT().SetHibernationWithTx(gomock.Any(), "addon-1", true).Return(nil, errors.GeneralError("update failed"))

		serr := svc.TriggerHibernate(context.Background(), "addon-1", true)

		Expect(serr.Reason).To(ContainSubstring("update failed"))
	})
})

type recordingClusterMutationCoordinator struct {
	lockedClusterID string
	lockedNamespace string
	unlocked        bool
}

func (c *recordingClusterMutationCoordinator) LockClusterNamespace(clusterID, namespace string) func() {
	c.lockedClusterID = clusterID
	c.lockedNamespace = namespace
	return func() { c.unlocked = true }
}
