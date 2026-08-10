package release

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"go.uber.org/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const testCloudPolicyVersion = "policy-v1"

var _ = Describe("cloud provisioning prerequisite", func() {
	var (
		ctx        context.Context
		ctrl       *gomock.Controller
		policy     *MockruntimePolicy
		stacks     *MockstackService
		volumes    *MockvolumeService
		namespaces *MocknamespaceService
		addons     *MockpostgresAddonService
		enqueuer   *mocks.MockBackgroundJobEnqueuer
		clusters   *mocks.MockClusterManager
		release    *models.StackRelease
	)

	BeforeEach(func() {
		ctx = context.Background()
		ctrl = gomock.NewController(GinkgoT())
		policy = NewMockruntimePolicy(ctrl)
		stacks = NewMockstackService(ctrl)
		volumes = NewMockvolumeService(ctrl)
		namespaces = NewMocknamespaceService(ctrl)
		addons = NewMockpostgresAddonService(ctrl)
		enqueuer = mocks.NewMockBackgroundJobEnqueuer(ctrl)
		clusters = mocks.NewMockClusterManager(ctrl)
		release = &models.StackRelease{ID: "release-1", StackID: "stack-1", Snapshot: models.StackSnapshot{
			Stack:       models.StackShellSnapshot{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo-ns"},
			Volumes:     []*models.Volume{{ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", NamespaceID: "namespace-1", Namespace: "demo-ns"}},
			Connections: models.StackConnections{{From: models.TopologyNodeRef{Type: models.TopologyNodeTypePostgresAddon, Id: "addon-1"}}},
		}}
	})

	newReconciler := func() *provisioningPrerequisiteReconciler {
		return newProvisioningPrerequisiteReconciler(ReleaseWorkerSpec{
			RuntimePolicy: policy, StackService: stacks, VolumeService: volumes,
			NamespaceService: namespaces, PostgresAddonService: addons,
			ReleaseWorkerEnqueuer: enqueuer, ClusterManager: clusters,
		})
	}

	It("fails closed before enqueueing when the allocation is inactive", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(errors.TrialInactive())
		result, err := newReconciler().Reconcile(ctx, release)
		Expect(result).To(Equal(resultNil))
		Expect(err).To(MatchError(ContainSubstring(errors.ErrorCodeTrialInactive)))
	})

	It("fails closed before enqueueing when the cloud isolation version is absent", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(nil)
		policy.EXPECT().IsolationPolicyVersion().Return("")

		result, err := newReconciler().Reconcile(ctx, release)
		Expect(result).To(Equal(resultNil))
		Expect(err).To(MatchError("cloud isolation policy version is not configured"))
	})

	It("enqueues prerequisites and requeues until all are observed ready", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(nil)
		ns := &models.Namespace{ID: "namespace-1", Name: "demo-ns", OrganisationID: "org-1", Labels: models.Labels{{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue}}}
		namespaces.EXPECT().Get(ctx, "namespace-1").Return(ns, nil)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo-ns"}, nil)
		volumes.EXPECT().InternalGet(ctx, "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", NamespaceID: "namespace-1", Namespace: "demo-ns", Status: &models.VolumeStatus{Phase: models.VolumePhasePending}}, nil)
		addons.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(&models.PostgresAddon{ID: "addon-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStatePending}}, nil)
		enqueuer.EXPECT().Enqueue(models.StackOperand{ID: "stack-1"}).Return(nil)
		enqueuer.EXPECT().Enqueue(models.VolumeOperand{ID: "volume-1"}).Return(nil)
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "addon-1"}).Return(nil)
		clusters.EXPECT().GetClient("cluster-1").Return(fake.NewClientBuilder().WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo-ns", Labels: map[string]string{models.CloudTenantLabelKey: models.CloudTenantLabelValue}}}).Build(), nil)
		policy.EXPECT().IsolationPolicyVersion().Return(testCloudPolicyVersion)

		result, err := newReconciler().Reconcile(ctx, release)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(resultRequeue))
	})

	It("uses only release snapshot volumes after the draft gains another volume", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(nil)
		ns := &models.Namespace{ID: "namespace-1", Name: "demo-ns", OrganisationID: "org-1", Labels: models.Labels{{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue}}}
		namespaces.EXPECT().Get(ctx, "namespace-1").Return(ns, nil)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo-ns", Volumes: []*models.Volume{{ID: "volume-added-after-release"}}}, nil)
		volumes.EXPECT().InternalGet(ctx, "volume-1").Return(&models.Volume{ID: "volume-1", OrganisationID: "org-1", ProjectID: "project-1", NamespaceID: "namespace-1", Namespace: "demo-ns", Status: &models.VolumeStatus{Phase: models.VolumePhaseReady}}, nil)
		addons.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(&models.PostgresAddon{ID: "addon-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStateReady}}, nil)
		enqueuer.EXPECT().Enqueue(models.StackOperand{ID: "stack-1"}).Return(nil)
		enqueuer.EXPECT().Enqueue(models.VolumeOperand{ID: "volume-1"}).Return(nil)
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "addon-1"}).Return(nil)
		clusters.EXPECT().GetClient("cluster-1").Return(fake.NewClientBuilder().WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo-ns", Labels: map[string]string{models.CloudTenantLabelKey: models.CloudTenantLabelValue, models.CloudPolicyReadyLabelKey: testCloudPolicyVersion}}}).Build(), nil)
		policy.EXPECT().IsolationPolicyVersion().Return(testCloudPolicyVersion)

		result, err := newReconciler().Reconcile(ctx, release)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(resultNil))
	})

	It("requeues when tenant identity exists but guard policy readiness is absent", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(nil)
		release.Snapshot.Volumes = nil
		ns := &models.Namespace{ID: "namespace-1", Name: "demo-ns", OrganisationID: "org-1", Labels: models.Labels{{Key: models.CloudTenantLabelKey, Value: models.CloudTenantLabelValue}}}
		namespaces.EXPECT().Get(ctx, "namespace-1").Return(ns, nil)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo-ns"}, nil)
		addons.EXPECT().InternalGetPostgresAddon(ctx, "addon-1").Return(&models.PostgresAddon{ID: "addon-1", Status: models.PostgresAddonStatus{State: models.PostgresAddonStateReady}}, nil)
		enqueuer.EXPECT().Enqueue(models.StackOperand{ID: "stack-1"}).Return(nil)
		enqueuer.EXPECT().Enqueue(models.PostgresAddonOperand{ID: "addon-1"}).Return(nil)
		clusters.EXPECT().GetClient("cluster-1").Return(fake.NewClientBuilder().WithObjects(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo-ns", Labels: map[string]string{models.CloudTenantLabelKey: models.CloudTenantLabelValue}}}).Build(), nil)
		policy.EXPECT().IsolationPolicyVersion().Return(testCloudPolicyVersion)

		result, err := newReconciler().Reconcile(ctx, release)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(resultRequeue))
	})

	It("rejects a current stack whose persisted identity differs from the release snapshot", func() {
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		policy.EXPECT().RequireActiveAllocation(ctx, "org-1").Return(nil)
		policy.EXPECT().IsolationPolicyVersion().Return(testCloudPolicyVersion)
		namespaces.EXPECT().Get(ctx, "namespace-1").Return(&models.Namespace{ID: "namespace-1", Name: "demo-ns", OrganisationID: "org-1"}, nil)
		stacks.EXPECT().InternalGetStack(ctx, "stack-1").Return(&models.Stack{ID: "stack-1", OrganisationID: "other-org", ClusterID: "cluster-1", NamespaceID: "namespace-1", Namespace: "demo-ns"}, nil)

		result, err := newReconciler().Reconcile(ctx, release)
		Expect(result).To(Equal(resultNil))
		Expect(err).To(MatchError("release release-1 stack identity does not match its snapshot"))
	})
})
