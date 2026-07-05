package services

import (
	"context"
	"testing"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	stackvalidator "github.com/Stackdome/stackdome/pkg/validator/stack"
	"go.uber.org/mock/gomock"
)

// TestGitIntegrationAttachesThroughStackValidation exercises the real
// resolver and real git integration service (mocked store, real encryption)
// end-to-end through stack validation: a git build source with no explicit
// credentials must clone-probe with the org integration's credentials.
func TestGitIntegrationAttachesThroughStackValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	encryption := newTestEncryptionService(t)

	// Sealed org integration for gitlab.example.com.
	auth := models.GitIntegrationAuth{Token: "tok-123"}
	blob, err := auth.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	encrypted, encErr := encryption.EncryptData(blob)
	if encErr != nil {
		t.Fatalf("encrypt failed: %v", encErr)
	}
	integration := &models.GitIntegration{
		ID:             "gi-1",
		OrganisationID: "org-1",
		Type:           models.GitIntegrationTypeGitCredentials,
		Host:           "gitlab.example.com",
		Status:         models.GitIntegrationStatusActive,
		EncryptedAuth:  encrypted,
		DataHash:       gitIntegrationDataHash(blob),
	}

	store := mocks.NewMockGitIntegrationStore(ctrl)
	store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "gitlab.example.com").Return(integration, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	gitIntegrationService := NewGitIntegrationService(GitIntegrationServiceSpec{
		Store:             store,
		EncryptionService: encryption,
		Permissions:       permissions,
	})

	resolver := NewCredentialResolver(CredentialResolverSpec{
		GitIntegrationService: gitIntegrationService,
	})

	// The clone probe must receive the integration's token.
	gitClient := mocks.NewMockGitClient(ctrl)
	gitClient.EXPECT().
		CheckAccess(gomock.Any(), "https://gitlab.example.com/acme/api").
		Return(true, nil)
	gitClients := mocks.NewMockgitClientProvider(ctrl)
	gitClients.EXPECT().
		ClientFor("https://gitlab.example.com/acme/api", gitclient.GitCredentials{Token: "tok-123"}).
		Return(gitClient, nil)

	validator := stackvalidator.NewStackValidator(stackvalidator.StackValidatorSpec{
		CredentialResolver: resolver,
		GitClients:         gitClients,
	})

	spec := &models.Stack{
		Name:           "demo",
		OrganisationID: "org-1",
		UserID:         "user-1",
		StackResources: []*models.StackResource{
			{
				Name: "api",
				BuildConfig: &models.BuildConfigSpec{
					SourceContext: models.BuildContextSource{
						Git: &models.GitBuildSource{RepoURL: "https://gitlab.example.com/acme/api"},
					},
					SourceRevision: models.BuildSourceRevision{
						Git: &models.GitRevision{Branch: "main"},
					},
					BuildImageRepository: models.BuildImageRepository{UseInClusterRegistry: true, InsecureRegistry: true},
				},
			},
		},
	}

	if err := validator.ValidateForCreate(context.Background(), spec); err != nil {
		t.Fatalf("expected integration-backed validation to pass, got %v", err)
	}
}
