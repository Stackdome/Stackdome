package stackfile

import (
	"context"
	"fmt"
	"testing"

	openapi "github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
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

func TestResolveStack_GitSecretResolved(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			StackResources: []openapi.StackResource{
				{
					Name: "api",
					BuildSpec: &openapi.StackResourceBuildSpec{
						SourceContext: openapi.BuildSourceContext{
							GitRepo: &openapi.BuildSourceContextGitRepo{
								RepoUrl:   "https://github.com/myorg/repo.git",
								GitSecret: &openapi.SecretRef{SecretId: "my-git-token"},
							},
						},
					},
				},
			},
		},
	}

	resolver := &mockResolver{
		secrets: map[string]string{"my-git-token": "secret-uuid-123"},
	}

	err := ResolveStack(context.Background(), stack, resolver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gitSecret := stack.Spec.StackResources[0].BuildSpec.SourceContext.GitRepo.GitSecret
	if gitSecret.SecretId != "secret-uuid-123" {
		t.Errorf("expected resolved ID 'secret-uuid-123', got %q", gitSecret.SecretId)
	}
}

func TestResolveStack_GitSecretNotFound(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			StackResources: []openapi.StackResource{
				{
					Name: "api",
					BuildSpec: &openapi.StackResourceBuildSpec{
						SourceContext: openapi.BuildSourceContext{
							GitRepo: &openapi.BuildSourceContextGitRepo{
								RepoUrl:   "https://github.com/myorg/repo.git",
								GitSecret: &openapi.SecretRef{SecretId: "nonexistent"},
							},
						},
					},
				},
			},
		},
	}

	resolver := &mockResolver{secrets: map[string]string{}}

	err := ResolveStack(context.Background(), stack, resolver)
	if err == nil {
		t.Fatal("expected error for missing git secret")
	}
}

func TestResolveStack_NoGitSecret(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			StackResources: []openapi.StackResource{
				{
					Name: "api",
					BuildSpec: &openapi.StackResourceBuildSpec{
						SourceContext: openapi.BuildSourceContext{
							GitRepo: &openapi.BuildSourceContextGitRepo{
								RepoUrl: "https://github.com/myorg/public.git",
							},
						},
					},
				},
			},
		},
	}

	resolver := &mockResolver{secrets: map[string]string{}}

	err := ResolveStack(context.Background(), stack, resolver)
	if err != nil {
		t.Fatalf("unexpected error for public repo: %v", err)
	}
}

func TestResolveStack_ImageResourceSkipped(t *testing.T) {
	stack := &openapi.Stack{
		Spec: openapi.StackSpec{
			StackResources: []openapi.StackResource{
				{
					Name:      "web",
					ImageSpec: &openapi.ImageSpec{Image: "nginx:latest"},
				},
			},
		},
	}

	resolver := &mockResolver{secrets: map[string]string{}}

	err := ResolveStack(context.Background(), stack, resolver)
	if err != nil {
		t.Fatalf("unexpected error for image resource: %v", err)
	}
}
