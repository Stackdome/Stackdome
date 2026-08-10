package postgresaddon

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	addonsv1alpha1 "stackdome.io/cluster-agent/api/addons/v1alpha1"
)

var _ = Describe("DeprovisionReconciler", func() {
	var (
		ctrl          *gomock.Controller
		addonSvc      *MockpostgresAddonService
		refSvc        *MockreferenceService
		clusterMgr    *MockclusterClientGetter
		clusterClient client.Client
		reconciler    *deprovisionReconciler
		ctx           context.Context
		addon         *models.PostgresAddon
	)

	newFakeClient := func(objs ...client.Object) client.Client {
		scheme := runtime.NewScheme()
		Expect(addonsv1alpha1.AddToScheme(scheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		addonSvc = NewMockpostgresAddonService(ctrl)
		refSvc = NewMockreferenceService(ctrl)
		clusterMgr = NewMockclusterClientGetter(ctrl)
		ctx = context.Background()

		reconciler = &deprovisionReconciler{
			postgresAddonService: addonSvc,
			referenceService:     refSvc,
			clusterManager:       clusterMgr,
			logger:               logger.NewLoggerWithPrefix(ctx, "test"),
		}

		addon = &models.PostgresAddon{
			ID:        "addon-1",
			ClusterID: "cluster-1",
			Name:      "test-delete",
			Namespace: "stackdome-addons-postgres-test-delete",
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("when DeletionTimestamp is nil", func() {
		It("no-ops even if the status state is Deleting", func() {
			addon.Status.State = models.PostgresAddonStateDeleting

			result, err := reconciler.Reconcile(ctx, addon)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultNil))
		})
	})

	Context("when DeletionTimestamp is set", func() {
		BeforeEach(func() {
			addon.DeletionTimestamp = ptr.To(time.Now().UTC())
		})

		It("deprovisions even when the status state was clobbered by a status-path write", func() {
			// Regression: the cluster watcher overwrote Status.State after the
			// DELETE request marked the addon Deleting; deletion intent must
			// survive via the DeletionTimestamp column.
			addon.Status.State = models.PostgresAddonStatePending

			cr := &addonsv1alpha1.PostgresCluster{
				ObjectMeta: metav1.ObjectMeta{Name: addon.Name, Namespace: addon.Namespace},
			}
			clusterClient = newFakeClient(cr)

			refSvc.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentPostgresAddon, addon.ID).Return(false, nil, nil)
			clusterMgr.EXPECT().GetClient(addon.ClusterID).Return(clusterClient, nil)
			addonSvc.EXPECT().InternalDeleteFromDB(gomock.Any(), addon.ID).Return(nil)

			result, err := reconciler.Reconcile(ctx, addon)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))

			var deleted addonsv1alpha1.PostgresCluster
			getErr := clusterClient.Get(ctx, client.ObjectKey{Name: addon.Name, Namespace: addon.Namespace}, &deleted)
			Expect(errors.IsNotFound(getErr)).To(BeTrue(), "PostgresCluster CR should be deleted")
		})

		It("deletes the DB row when the PostgresCluster CR is already gone", func() {
			addon.Status.State = models.PostgresAddonStateDeleting
			clusterClient = newFakeClient()

			refSvc.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentPostgresAddon, addon.ID).Return(false, nil, nil)
			clusterMgr.EXPECT().GetClient(addon.ClusterID).Return(clusterClient, nil)
			addonSvc.EXPECT().InternalDeleteFromDB(gomock.Any(), addon.ID).Return(nil)

			result, err := reconciler.Reconcile(ctx, addon)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultStop))
		})

		It("requeues while the addon is still in use by stacks", func() {
			addon.Status.State = models.PostgresAddonStateDeleting

			refSvc.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentPostgresAddon, addon.ID).Return(true, nil, nil)

			result, err := reconciler.Reconcile(ctx, addon)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(resultRequeueAfter(30 * time.Second)))
		})
	})
})
