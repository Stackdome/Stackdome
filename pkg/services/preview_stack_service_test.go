package services

import (
	"context"
	"testing"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestPreviewGitClientForConfigResolvesExplicitIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	resolver := mocks.NewMockCredentialResolver(ctrl)
	svc := &previewStackService{credentialResolver: resolver}

	config := &models.StackPreviewConfig{
		OrganisationID: "org-1",
		GitRepository: models.PreviewGitRepository{
			RepoURL:       "https://github.com/acme/api",
			IntegrationID: "int-1",
		},
	}

	resolver.EXPECT().
		GitCredentials(gomock.Any(), "org-1", "https://github.com/acme/api", credentials.GitAuthSelector{
			IntegrationID: "int-1",
		}).
		Return(&credentials.ResolvedGitCredential{
			Source:      credentials.SourceIntegration,
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
