package services

import (
	"context"
	"testing"

	gitclient "github.com/ashishmax31/stackdome-api-server/pkg/clients/git"
	"github.com/ashishmax31/stackdome-api-server/pkg/credentials"
	"github.com/ashishmax31/stackdome-api-server/pkg/mocks"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"
)

func TestPreviewGitClientForConfigResolvesExplicitSecretRef(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	svc := &previewStackService{credentialResolver: resolver}

	config := &models.StackPreviewConfig{
		OrganisationID: "org-1",
		GitRepository: models.PreviewGitRepository{
			RepoURL:     "https://github.com/acme/api",
			GitSecretID: ptr.To("sec-1"),
		},
	}

	resolver.EXPECT().
		GitCredentials(gomock.Any(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{
			SecretRef: &models.SecretReference{SecretID: "sec-1"},
		}).
		Return(&credentials.ResolvedGitCredential{
			Source:      credentials.SourceSecretRef,
			Credentials: gitclient.GitCredentials{Token: "tok"},
		}, nil)

	client, err := svc.gitClientForConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected a git client")
	}
}

func TestPreviewGitClientForConfigFallsThroughAnonymously(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	svc := &previewStackService{credentialResolver: resolver}

	config := &models.StackPreviewConfig{
		OrganisationID: "org-1",
		GitRepository: models.PreviewGitRepository{
			RepoURL: "https://github.com/acme/api",
		},
	}

	// No secret and no integration -> resolver is still consulted (org-level
	// integrations can resolve here) and falls through to anonymous.
	resolver.EXPECT().
		GitCredentials(gomock.Any(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{}).
		Return(&credentials.ResolvedGitCredential{Source: credentials.SourceAnonymous}, nil)

	client, err := svc.gitClientForConfig(context.Background(), config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected a git client")
	}
}
