package services

import (
	"context"
	"fmt"
	"strings"
	"testing"

	gitclient "github.com/Stackdome/stackdome/pkg/clients/git"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

type gitIntegrationServiceMocks struct {
	store      *mocks.MockGitIntegrationStore
	gitClients *MockverifyGitClientProvider
	gitClient  *mocks.MockGitClient
	encryption EncryptionService
}

func newGitIntegrationServiceForTest(t *testing.T) (GitIntegrationService, *gitIntegrationServiceMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	deps := &gitIntegrationServiceMocks{
		store:      mocks.NewMockGitIntegrationStore(ctrl),
		gitClients: NewMockverifyGitClientProvider(ctrl),
		gitClient:  mocks.NewMockGitClient(ctrl),
		encryption: newTestEncryptionService(t),
	}
	svc := NewGitIntegrationService(GitIntegrationServiceSpec{
		Store:             deps.store,
		EncryptionService: deps.encryption,
		Permissions:       permissions,
		Logger:            logger.NewLogger(),
		GitClients:        deps.gitClients,
	})
	return svc, deps
}

func sealedIntegration(t *testing.T, encryption EncryptionService, auth models.GitIntegrationAuth) *models.GitIntegration {
	t.Helper()
	blob, err := auth.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	encrypted, err := encryption.EncryptData(blob)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	return &models.GitIntegration{
		ID:             "gi-1",
		OrganisationID: "org-1",
		Type:           models.GitIntegrationTypeGitCredentials,
		Host:           "gitlab.example.com",
		Status:         models.GitIntegrationStatusActive,
		EncryptedAuth:  encrypted,
		DataHash:       gitIntegrationDataHash(blob),
	}
}

func TestGitIntegrationCreateEncryptsAuthAtRest(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "gitlab.example.com").
		Return(nil, errors.NotFound("none"))

	var stored *models.GitIntegration
	deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
			stored = integration
			return integration, nil
		})

	created, serr := svc.Create(context.Background(), &models.GitIntegration{
		OrganisationID: "org-1",
		Host:           "https://gitlab.example.com:8443/acme/api.git",
		Auth:           &models.GitIntegrationAuth{Token: "tok-123"},
	})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}

	if stored.Host != "gitlab.example.com" {
		t.Fatalf("expected host to be normalized, got %q", stored.Host)
	}
	if stored.Type != models.GitIntegrationTypeGitCredentials || stored.Status != models.GitIntegrationStatusActive {
		t.Fatalf("unexpected defaults %+v", stored)
	}
	if stored.EncryptedAuth == "" || strings.Contains(stored.EncryptedAuth, "tok-123") {
		t.Fatal("expected auth to be encrypted at rest")
	}
	if stored.Auth != nil || created.Auth != nil {
		t.Fatal("transient auth must be cleared before storage")
	}
	if stored.DataHash == "" {
		t.Fatal("expected data hash to be stamped")
	}

	decrypted, err := deps.encryption.DecryptData(stored.EncryptedAuth)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if !strings.Contains(string(decrypted), "tok-123") {
		t.Fatalf("decrypted blob missing token: %s", decrypted)
	}
}

func TestGitIntegrationCreateRejectsDuplicateHost(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "gitlab.example.com").
		Return(&models.GitIntegration{ID: "existing"}, nil)

	_, serr := svc.Create(context.Background(), &models.GitIntegration{
		OrganisationID: "org-1",
		Host:           "gitlab.example.com",
		Auth:           &models.GitIntegrationAuth{Token: "tok-123"},
	})
	if serr == nil || !serr.IsConflict() {
		t.Fatalf("expected 409 conflict, got %v", serr)
	}
}

func TestGitIntegrationVerifySucceeds(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)
	integration := sealedIntegration(t, deps.encryption, models.GitIntegrationAuth{Token: "tok-123"})

	deps.store.EXPECT().GetByID(gomock.Any(), "gi-1").Return(integration, nil)
	deps.gitClients.EXPECT().
		ClientFor("https://gitlab.example.com/acme/api", gitclient.GitCredentials{Token: "tok-123"}).
		Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().CheckAccess(gomock.Any(), "https://gitlab.example.com/acme/api").Return(true, nil)

	if serr := svc.Verify(context.Background(), "gi-1", "https://gitlab.example.com/acme/api"); serr != nil {
		t.Fatalf("expected verify to pass, got %v", serr)
	}
}

func TestGitIntegrationVerifyAuthFailure(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)
	integration := sealedIntegration(t, deps.encryption, models.GitIntegrationAuth{
		Basic: &models.GitIntegrationBasicAuth{Username: "alice", Password: "bad"},
	})

	deps.store.EXPECT().GetByID(gomock.Any(), "gi-1").Return(integration, nil)
	deps.gitClients.EXPECT().
		ClientFor(gomock.Any(), gitclient.GitCredentials{Username: "alice", Password: "bad"}).
		Return(deps.gitClient, nil)
	deps.gitClient.EXPECT().CheckAccess(gomock.Any(), gomock.Any()).
		Return(false, fmt.Errorf("authentication failed: %w", gitclient.ErrAuthFailed))

	serr := svc.Verify(context.Background(), "gi-1", "https://gitlab.example.com/acme/api")
	if serr == nil {
		t.Fatal("expected verify to fail")
	}
}

func TestGitIntegrationVerifyRejectsHostMismatch(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)
	integration := sealedIntegration(t, deps.encryption, models.GitIntegrationAuth{Token: "tok-123"})

	deps.store.EXPECT().GetByID(gomock.Any(), "gi-1").Return(integration, nil)

	if serr := svc.Verify(context.Background(), "gi-1", "https://github.com/acme/api"); serr == nil {
		t.Fatal("expected host mismatch to be rejected")
	}
}

func TestGitIntegrationInternalGetForHostDecrypts(t *testing.T) {
	svc, deps := newGitIntegrationServiceForTest(t)
	integration := sealedIntegration(t, deps.encryption, models.GitIntegrationAuth{Token: "tok-123"})

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "gitlab.example.com").Return(integration, nil)

	got, serr := svc.InternalGetForHost(context.Background(), "org-1", "gitlab.example.com")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if got.Auth == nil || got.Auth.Token != "tok-123" {
		t.Fatalf("expected decrypted auth, got %+v", got.Auth)
	}
}
