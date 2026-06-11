package stack

import (
	"context"
	"testing"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

func TestRevisionReconcilerComputesAndPersistsHash(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	stackSvc := NewMockstackService(ctrl)

	stack := &models.Stack{
		ID:        "stack-1",
		Name:      "app",
		Namespace: "ns-1",
		StackResources: []*models.StackResource{
			{
				Name:        "api",
				Namespace:   "ns-1",
				ImageConfig: &models.ImageConfigSpec{Image: "example/api:latest"},
			},
		},
	}

	builder := builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{})
	expectedHash, err := builder.GetStackCRHash(stack)
	g.Expect(err).NotTo(HaveOccurred())

	stackSvc.EXPECT().UpdateStackCrRevision(gomock.Any(), "stack-1", expectedHash).Return(nil)

	reconciler := NewRevisionReconciler(RevisionReconcilerSpec{
		StackService:   stackSvc,
		StackCRBuilder: builder,
	})
	res, rerr := reconciler.Reconcile(context.Background(), stack)

	g.Expect(rerr).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(resultNil))
	g.Expect(stack.CrRevision).To(Equal(expectedHash))
}

func TestRevisionReconcilerSkipsPersistWhenUnchanged(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	stackSvc := NewMockstackService(ctrl)

	stack := &models.Stack{
		ID:        "stack-1",
		Name:      "app",
		Namespace: "ns-1",
		StackResources: []*models.StackResource{
			{
				Name:        "api",
				Namespace:   "ns-1",
				ImageConfig: &models.ImageConfigSpec{Image: "example/api:latest"},
			},
		},
	}
	builder := builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{})
	currentHash, err := builder.GetStackCRHash(stack)
	g.Expect(err).NotTo(HaveOccurred())
	stack.CrRevision = currentHash

	reconciler := NewRevisionReconciler(RevisionReconcilerSpec{
		StackService:   stackSvc,
		StackCRBuilder: builder,
	})
	res, rerr := reconciler.Reconcile(context.Background(), stack)

	g.Expect(rerr).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(resultNil))
}

func TestRevisionReconcilerSkipsDeletedStacks(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	stackSvc := NewMockstackService(ctrl)
	now := time.Now()
	stack := &models.Stack{ID: "stack-1", DeletionTimestamp: &now}

	reconciler := NewRevisionReconciler(RevisionReconcilerSpec{
		StackService:   stackSvc,
		StackCRBuilder: builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{}),
	})
	res, err := reconciler.Reconcile(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(resultNil))
}
