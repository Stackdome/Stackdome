package volume

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/builders"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/Stackdome/stackdome/pkg/worker"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	storagev1alpha1 "stackdome.io/cluster-agent/api/storage/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("VolumeWorker cloud admission", func() {
	It("does not select draft-only volumes for periodic cloud reconciliation", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		releases := NewMockreleaseService(ctrl)
		refs := NewMockreferenceService(ctrl)
		releases.EXPECT().InternalListAuthoritativeWorkload(gomock.Any()).Return(&models.WorkloadAuthorityScan{}, nil)
		refs.EXPECT().InternalListReleaseReferents(gomock.Any(), []string{}, models.ReferentVolume).Return([]models.ResourceReference{}, nil)
		w := &volumeWorker{volumeService: volumes, releaseService: releases, referenceService: refs, runtimePolicy: &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test")}

		operands, serr := w.GetInput(context.Background())
		Expect(serr).To(BeNil())
		Expect(operands).To(BeEmpty())
	})

	It("selects declared release volumes without loading release snapshots", func() {
		ctrl := gomock.NewController(GinkgoT())
		releases := NewMockreleaseService(ctrl)
		refs := NewMockreferenceService(ctrl)
		releases.EXPECT().InternalListAuthoritativeWorkload(gomock.Any()).Return(&models.WorkloadAuthorityScan{
			Releases: []models.WorkloadReleaseRef{{StackID: "stack-1", ReleaseID: "release-1"}},
		}, nil)
		releaseID := "release-1"
		refs.EXPECT().InternalListReleaseReferents(gomock.Any(), []string{releaseID}, models.ReferentVolume).Return([]models.ResourceReference{
			{ReleaseID: &releaseID, ReferentType: models.ReferentVolume, ReferentID: "volume-1", RelationKind: models.RelationVolumeDeclaration},
			{ReleaseID: &releaseID, ReferentType: models.ReferentVolume, ReferentID: "volume-1", RelationKind: models.RelationVolumeMount},
		}, nil)
		w := &volumeWorker{
			releaseService: releases, referenceService: refs, runtimePolicy: &activeVolumeRuntimePolicy{},
			BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		operands, serr := w.GetInput(context.Background())

		Expect(serr).To(BeNil())
		Expect(operands).To(ConsistOf(models.VolumeOperand{ID: "volume-1", ReleaseID: releaseID}))
	})

	It("provisions the immutable release volume even when the draft has drifted", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		clusters := mocks.NewMockClusterManager(ctrl)
		scheme := runtime.NewScheme()
		Expect(storagev1alpha1.AddToScheme(scheme)).To(Succeed())
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		snapshotVolume := &models.Volume{
			ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			Name: "data", NamespaceID: "namespace-1", Namespace: "stack-1", Size: "1Gi",
			VolumeSource: &models.VolumeSource{RemoteDirSource: &models.RemoteDirSource{Path: "/data", CurrentDirectoryHash: "release-hash"}},
		}
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"}
		release := &models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{
				Stack:   models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"},
				Volumes: []*models.Volume{snapshotVolume},
			},
		}
		releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(release, nil)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil).Times(3)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil).Times(3)
		clusters.EXPECT().GetClient("cluster-1").Return(clusterClient, nil)
		w := &volumeWorker{
			volumeService: volumes, stackService: stacks, releaseService: releases, clusterManager: clusters,
			volumeCRbuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
			runtimePolicy:   &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		_, serr := w.Execute(context.Background(), models.VolumeOperand{ID: "volume-1", ReleaseID: "release-1"})

		Expect(serr).To(BeNil())
		created := &storagev1alpha1.Volume{}
		Expect(clusterClient.Get(context.Background(), client.ObjectKey{Name: "data", Namespace: "stack-1"}, created)).To(Succeed())
		Expect(created.Spec.Source.RemoteDir.CurrentDirectoryHash).To(Equal("release-hash"))
	})

	It("creates a volume after its release was admitted", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		policy := services.NewMockRuntimePolicy(ctrl)
		clusters := mocks.NewMockClusterManager(ctrl)
		scheme := runtime.NewScheme()
		Expect(storagev1alpha1.AddToScheme(scheme)).To(Succeed())
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		volume := &models.Volume{
			ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			Name: "data", NamespaceID: "namespace-1", Namespace: "stack-1", Size: "1Gi",
			VolumeSource: &models.VolumeSource{RemoteDirSource: &models.RemoteDirSource{Path: "/data", CurrentDirectoryHash: "release-hash"}},
		}
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"}
		release := &models.StackRelease{
			ID: "release-a", StackID: stack.ID, State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{
				Stack:   models.StackShellSnapshot{ID: stack.ID, OrganisationID: stack.OrganisationID, ProjectID: stack.ProjectID, ClusterID: stack.ClusterID, NamespaceID: stack.NamespaceID, Namespace: stack.Namespace},
				Volumes: []*models.Volume{volume},
			},
		}
		releases.EXPECT().InternalGet(gomock.Any(), release.ID).Return(release, nil)
		stacks.EXPECT().InternalGetStack(gomock.Any(), stack.ID).Return(stack, nil).Times(3)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil).Times(3)
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		clusters.EXPECT().GetClient(stack.ClusterID).Return(clusterClient, nil)
		w := &volumeWorker{
			volumeService: volumes, stackService: stacks, releaseService: releases, clusterManager: clusters,
			volumeCRbuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
			runtimePolicy:   policy, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.VolumeOperand{ID: volume.ID, ReleaseID: release.ID})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
		observed := &storagev1alpha1.Volume{}
		Expect(clusterClient.Get(context.Background(), client.ObjectKey{Name: volume.Name, Namespace: volume.Namespace}, observed)).
			To(Succeed())
	})

	It("does not replay a historical volume after a newer release converges", func() {
		ctrl := gomock.NewController(GinkgoT())
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		clusters := mocks.NewMockClusterManager(ctrl)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(&models.Stack{
			ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1",
			Status: &models.StackStatus{LastConverged: &models.StackConvergenceRecord{ReleaseID: "release-b"}},
		}, nil)
		historicalRelease := &models.StackRelease{
			ID: "release-a", StackID: "stack-1", State: models.ReleaseStateReleased,
			Snapshot: models.StackSnapshot{
				Stack:   models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"},
				Volumes: []*models.Volume{{ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", NamespaceID: "namespace-1", Namespace: "stack-1"}},
			},
		}
		releases.EXPECT().InternalGet(gomock.Any(), "release-a").Return(historicalRelease, nil)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), gomock.Any()).Return(nil, nil)
		w := &volumeWorker{
			stackService: stacks, releaseService: releases, clusterManager: clusters,
			runtimePolicy: &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		result, serr := w.Execute(context.Background(), models.VolumeOperand{ID: "volume-1", ReleaseID: "release-a"})

		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("uses a persisted git commit without resolving the repository again", func() {
		ctrl := gomock.NewController(GinkgoT())
		releases := NewMockreleaseService(ctrl)
		stacks := NewMockstackService(ctrl)
		clusters := mocks.NewMockClusterManager(ctrl)
		scheme := runtime.NewScheme()
		Expect(storagev1alpha1.AddToScheme(scheme)).To(Succeed())
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		stack := &models.Stack{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"}
		snapshotVolume := &models.Volume{
			ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", Name: "data", NamespaceID: "namespace-1", Namespace: "stack-1", Size: "1Gi",
			VolumeSource: &models.VolumeSource{GitRepoSource: &models.GitRepoSource{
				RepoUrl: "not-a-network-url", Revision: models.GitRepoRevision{Branch: "main", Commit: "commit-x"},
			}},
		}
		release := &models.StackRelease{ID: "release-1", StackID: "stack-1", State: models.ReleaseStatePending, Snapshot: models.StackSnapshot{
			Stack: models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"}, Volumes: []*models.Volume{snapshotVolume},
		}}
		releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(release, nil)
		releases.EXPECT().InternalResolveAuthoritativeWorkloadRelease(gomock.Any(), stack).Return(release, nil).Times(3)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(stack, nil).Times(3)
		clusters.EXPECT().GetClient("cluster-1").Return(clusterClient, nil)
		w := &volumeWorker{
			stackService: stacks, releaseService: releases, clusterManager: clusters,
			volumeCRbuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
			runtimePolicy:   &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		_, serr := w.Execute(context.Background(), models.VolumeOperand{ID: "volume-1", ReleaseID: "release-1"})

		Expect(serr).To(BeNil())
		created := &storagev1alpha1.Volume{}
		Expect(clusterClient.Get(context.Background(), client.ObjectKey{Name: "data", Namespace: "stack-1"}, created)).To(Succeed())
		Expect(created.Spec.Source.GitRepo.Revision.Commit).To(Equal("commit-x"))
	})
})

type activeVolumeRuntimePolicy struct{ inactiveVolumeRuntimePolicy }

type inactiveVolumeRuntimePolicy struct{}

func (*inactiveVolumeRuntimePolicy) OrganisationProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) IsolationPolicyVersion() string { return "v1" }
func (*inactiveVolumeRuntimePolicy) ActivateComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) RequireComputeAccessWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) AdmitComputeMutationWithTx(context.Context, string) (services.ComputeMutationAdmission, *errors.ServiceError) {
	return services.ComputeMutationAdmission{}, nil
}
func (*inactiveVolumeRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
