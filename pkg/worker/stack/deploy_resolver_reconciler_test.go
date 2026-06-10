package stack

import (
	"context"
	"testing"
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/builders"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	. "github.com/onsi/gomega"
)

func TestDeployResolverReconcilerChangesEffectiveHash(t *testing.T) {
	g := NewWithT(t)
	stack := &models.Stack{
		ID:        "stack-1",
		Name:      "app",
		Namespace: "ns-1",
		StackResources: []*models.StackResource{
			{
				ID:          "api-id",
				StackID:     "stack-1",
				Name:        "api",
				Namespace:   "ns-1",
				ImageConfig: &models.ImageConfigSpec{Image: "example/api:latest"},
				ExecutionConfig: &models.ExecutionConfig{
					Env: []models.EnvVar{{Name: "PUBLIC_URL", SelfOutput: "public.http.url"}},
				},
				Ports: models.Ports{
					{Name: "http", Number: 8080, Protocol: "http", ExposedToPublic: true, ExposedFqdn: "api.example.com"},
				},
			},
		},
	}

	builder := builders.NewClusterResourceBuilder(builders.ClusterResourceBuilderSpec{})
	rawHash, err := builder.GetStackCRHash(stack)
	g.Expect(err).NotTo(HaveOccurred())

	reconciler := NewDeployResolverReconciler(DeployResolverReconcilerSpec{})
	res, rerr := reconciler.Reconcile(context.Background(), stack)
	g.Expect(rerr).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(resultNil))

	effectiveHash, err := builder.GetStackCRHash(stack)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(effectiveHash).NotTo(Equal(rawHash))
}

func TestDeployResolverReconcilerSkipsDeletedStacks(t *testing.T) {
	g := NewWithT(t)
	now := time.Now()
	stack := &models.Stack{ID: "stack-1", DeletionTimestamp: &now}

	reconciler := NewDeployResolverReconciler(DeployResolverReconcilerSpec{})
	res, err := reconciler.Reconcile(context.Background(), stack)

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(resultNil))
}
