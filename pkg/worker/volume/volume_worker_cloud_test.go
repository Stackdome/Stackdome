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
	It("does not reconcile a draft volume selected by the periodic scanner", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		stackVolumes := NewMockstackVolumeStore(ctrl)
		stacks := NewMockstackService(ctrl)
		volumes.EXPECT().InternalGet(gomock.Any(), "volume-1").Return(&models.Volume{ID: "volume-1"}, nil)
		stackVolumes.EXPECT().GetByVolumeID(gomock.Any(), "volume-1").Return(&models.StackVolume{StackID: "stack-1"}, nil)
		stacks.EXPECT().InternalGetStack(gomock.Any(), "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1"}, nil)
		volumes.EXPECT().InternalListNotReady(gomock.Any()).Return([]*models.Volume{{ID: "volume-1"}}, nil)
		w := &volumeWorker{volumeService: volumes, stackVolumeStore: stackVolumes, stackService: stacks, runtimePolicy: &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test")}

		operands, serr := w.GetInput(context.Background())
		Expect(serr).To(BeNil())
		Expect(operands).To(ConsistOf(models.VolumeOperand{ID: "volume-1"}))
		result, serr := w.Execute(context.Background(), operands[0])
		Expect(serr).To(BeNil())
		Expect(result).To(Equal(worker.Result{}))
	})

	It("provisions the immutable release volume even when the draft has drifted", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumes := NewMockvolumeService(ctrl)
		releases := NewMockreleaseService(ctrl)
		clusters := mocks.NewMockClusterManager(ctrl)
		scheme := runtime.NewScheme()
		Expect(storagev1alpha1.AddToScheme(scheme)).To(Succeed())
		clusterClient := fake.NewClientBuilder().WithScheme(scheme).Build()
		snapshotVolume := &models.Volume{
			ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", UserID: "user-1",
			Name: "data", NamespaceID: "namespace-1", Namespace: "stack-1", Size: "1Gi",
			VolumeSource: &models.VolumeSource{RemoteDirSource: &models.RemoteDirSource{Path: "/data", CurrentDirectoryHash: "release-hash"}},
		}
		releases.EXPECT().InternalGet(gomock.Any(), "release-1").Return(&models.StackRelease{
			ID: "release-1", StackID: "stack-1", State: models.ReleaseStatePending,
			Snapshot: models.StackSnapshot{
				Stack:   models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ProjectID: "project-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "stack-1"},
				Volumes: []*models.Volume{snapshotVolume},
			},
		}, nil)
		clusters.EXPECT().GetClient("cluster-1").Return(clusterClient, nil)
		w := &volumeWorker{
			volumeService: volumes, releaseService: releases, clusterManager: clusters,
			volumeCRbuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
			runtimePolicy:   &activeVolumeRuntimePolicy{}, BaseWorker: worker.NewBaseWorker(VolumeWorkerName, "test"),
		}

		_, serr := w.Execute(context.Background(), models.VolumeOperand{ID: "volume-1", ReleaseID: "release-1"})

		Expect(serr).To(BeNil())
		created := &storagev1alpha1.Volume{}
		Expect(clusterClient.Get(context.Background(), client.ObjectKey{Name: "data", Namespace: "stack-1"}, created)).To(Succeed())
		Expect(created.Spec.Source.RemoteDir.CurrentDirectoryHash).To(Equal("release-hash"))
	})
})

type activeVolumeRuntimePolicy struct{ inactiveVolumeRuntimePolicy }

func (*activeVolumeRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return nil
}

type inactiveVolumeRuntimePolicy struct{}

func (*inactiveVolumeRuntimePolicy) OrganisationProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) DraftProvisioningMode() services.ProvisioningMode {
	return services.ProvisioningModeDatabaseOnly
}
func (*inactiveVolumeRuntimePolicy) IsolationPolicyVersion() string { return "v1" }
func (*inactiveVolumeRuntimePolicy) AdmitFirstReleaseWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) AdmitRollbackWithTx(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) RequireActiveAllocation(context.Context, string) *errors.ServiceError {
	return errors.TrialInactive()
}
func (*inactiveVolumeRuntimePolicy) AdmitMutationWithTx(context.Context, string) (services.MutationAdmission, *errors.ServiceError) {
	return services.MutationAdmission{}, nil
}
func (*inactiveVolumeRuntimePolicy) AdmitOrganisationDeletion(context.Context, string) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) AdmitStackMutationWithTx(context.Context, services.StackMutation) *errors.ServiceError {
	return nil
}
func (*inactiveVolumeRuntimePolicy) ApplyStackResourceDefaults(*models.StackResource) {}
