package stackfile

import (
	"context"
	"fmt"
	"testing"

	openapi "github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"k8s.io/utils/ptr"
)

type mockResolver struct {
	secrets map[string]string
	addons  map[string]string
}

func (m *mockResolver) ResolveSecretByName(_ context.Context, name string) (string, error) {
	if id, ok := m.secrets[name]; ok {
		return id, nil
	}
	return "", fmt.Errorf("secret %q not found", name)
}

func (m *mockResolver) ResolveAddonByName(_ context.Context, _, name string) (string, error) {
	if id, ok := m.addons[name]; ok {
		return id, nil
	}
	return "", fmt.Errorf("addon %q not found", name)
}

func TestResolveStack_SecretConnectionResolved(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			Connections: []openapi.StackConnection{
				{
					Kind: "env",
					From: openapi.TopologyNodeRef{Type: "secret", Name: ptr.To("app-config")},
					To:   openapi.TopologyNodeRef{Type: "stack_resource", Name: ptr.To("api")},
				},
			},
		},
	}

	resolver := &mockResolver{
		secrets: map[string]string{"app-config": "secret-uuid-123"},
	}

	if err := ResolveStack(context.Background(), stack, resolver); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	from := stack.Spec.Connections[0].From
	if from.Id == nil || *from.Id != "secret-uuid-123" {
		t.Errorf("expected resolved ID 'secret-uuid-123', got %v", from.Id)
	}
	if from.Name != nil {
		t.Errorf("expected name to be cleared after resolution, got %v", from.Name)
	}
}

func TestResolveStack_SecretConnectionNotFound(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			Connections: []openapi.StackConnection{
				{
					Kind: "env",
					From: openapi.TopologyNodeRef{Type: "secret", Name: ptr.To("missing")},
					To:   openapi.TopologyNodeRef{Type: "stack_resource", Name: ptr.To("api")},
				},
			},
		},
	}

	resolver := &mockResolver{secrets: map[string]string{}}

	if err := ResolveStack(context.Background(), stack, resolver); err == nil {
		t.Fatal("expected error for missing secret")
	}
}

func TestResolveStack_ImageResourceSkipped(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			StackResources: []openapi.StackResource{
				{
					Name:   "web",
					Source: &openapi.SourceSpec{Image: openapi.NewImageSource("nginx:latest")},
				},
			},
		},
	}

	resolver := &mockResolver{secrets: map[string]string{}}

	if err := ResolveStack(context.Background(), stack, resolver); err != nil {
		t.Fatalf("unexpected error for image resource: %v", err)
	}
}
