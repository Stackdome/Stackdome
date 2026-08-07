package shared

import (
	"context"
	"testing"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

func TestGetIngressesForStackResourceUsesControllerOwnership(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("registering networking types: %v", err)
	}

	controlled := true
	resource := &corev1alpha1.StackResource{ObjectMeta: metav1.ObjectMeta{
		Name:      "todo-app",
		Namespace: "test-ns",
		UID:       types.UID("resource-uid"),
	}}
	wildcardIngress := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{
		Name:      "todo-app-wildcard",
		Namespace: "test-ns",
		OwnerReferences: []metav1.OwnerReference{{
			APIVersion: "core.stackdome.io/v1alpha1",
			Kind:       "StackResource",
			Name:       "todo-app",
			UID:        resource.UID,
			Controller: &controlled,
		}},
	}}
	httpIngress := wildcardIngress.DeepCopy()
	httpIngress.Name = "todo-app-http-proxy"
	unrelatedIngress := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{
		Name:      "unrelated-http-proxy",
		Namespace: "test-ns",
	}}
	clusterClient := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(wildcardIngress, httpIngress, unrelatedIngress).
		Build()

	ingresses, err := GetIngressesForStackResource(
		context.Background(), clusterClient, resource,
	)
	if err != nil {
		t.Fatalf("listing StackResource ingresses: %v", err)
	}
	if len(ingresses) != 2 {
		t.Fatalf("expected both owned ingresses, got %#v", ingresses)
	}
}
