package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

type githubAppServiceMocks struct {
	store         *mocks.MockGitIntegrationStore
	installations *mocks.MockGitInstallationStore
	oauthStates   *mocks.MockOAuthStateStore
	organisations *mocks.MockOrganisationStore
	appClient     *mocks.MockGitHubAppClient
	encryption    EncryptionService
}

func newGitHubAppServiceForTest(t *testing.T) (GitIntegrationService, *githubAppServiceMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	deps := &githubAppServiceMocks{
		store:         mocks.NewMockGitIntegrationStore(ctrl),
		installations: mocks.NewMockGitInstallationStore(ctrl),
		oauthStates:   mocks.NewMockOAuthStateStore(ctrl),
		organisations: mocks.NewMockOrganisationStore(ctrl),
		appClient:     mocks.NewMockGitHubAppClient(ctrl),
		encryption:    newTestEncryptionService(t),
	}
	deps.organisations.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&models.Organisation{ID: "org-1", Name: "Acme Corp"}, nil).AnyTimes()
	svc := NewGitIntegrationService(GitIntegrationServiceSpec{
		Store:             deps.store,
		InstallationStore: deps.installations,
		OAuthStateStore:   deps.oauthStates,
		OrganisationStore: deps.organisations,
		GitHubAppClient:   deps.appClient,
		EncryptionService: deps.encryption,
		Permissions:       permissions,
		ExternalURL:       "https://hub.example.com",
	})
	return svc, deps
}

func sealedGitHubApp(t *testing.T, encryption EncryptionService, app models.GitHubAppCredentials) *models.GitIntegration {
	t.Helper()
	auth := models.GitIntegrationAuth{GitHubApp: &app}
	blob, err := auth.Marshal()
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	encrypted, encErr := encryption.EncryptData(blob)
	if encErr != nil {
		t.Fatalf("encrypt failed: %v", encErr)
	}
	return &models.GitIntegration{
		ID:             "gi-app",
		OrganisationID: "org-1",
		Type:           models.GitIntegrationTypeGitHubApp,
		Host:           "github.com",
		Status:         models.GitIntegrationStatusInstalled,
		EncryptedAuth:  encrypted,
		DataHash:       gitIntegrationDataHash(blob),
	}
}

func TestCreateGitHubAppManifestShape(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)

	deps.store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(nil, errors.NotFound("none"))
	deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, integration *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
			integration.ID = "gi-new"
			if integration.Type != models.GitIntegrationTypeGitHubApp || integration.Status != models.GitIntegrationStatusPendingInstall {
				t.Fatalf("unexpected integration row %+v", integration)
			}
			return integration, nil
		})
	deps.oauthStates.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, state *models.OAuthState) *errors.ServiceError {
			if state.Provider != models.OAuthProviderGitHubAppManifest {
				t.Fatalf("unexpected provider %q", state.Provider)
			}
			if !strings.HasSuffix(state.State, ":gi-new") {
				t.Fatalf("state must carry the integration id, got %q", state.State)
			}
			return nil
		})

	flow, serr := svc.CreateGitHubAppManifest(context.Background(), "org-1")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if flow.Manifest["redirect_url"] != "https://hub.example.com/api/v1/git-integrations/github/manifest/callback" {
		t.Fatalf("unexpected redirect_url %v", flow.Manifest["redirect_url"])
	}
	hook, ok := flow.Manifest["hook_attributes"].(map[string]any)
	if !ok || hook["url"] != "https://hub.example.com/api/v1/webhooks/github" {
		t.Fatalf("unexpected hook attributes %v", flow.Manifest["hook_attributes"])
	}
	if flow.Manifest["public"] != true {
		t.Fatal("expected a public app manifest")
	}
	if flow.Manifest["name"] != "stackdome-acme-corp" {
		t.Fatalf("expected name slugified from org name, got %v", flow.Manifest["name"])
	}
	if !strings.Contains(flow.GitHubURL, "state=") {
		t.Fatalf("expected state in github url, got %q", flow.GitHubURL)
	}
}

func TestHandleGitHubManifestCallbackSealsCredentials(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)

	integration := &models.GitIntegration{
		ID:             "gi-app",
		OrganisationID: "org-1",
		Type:           models.GitIntegrationTypeGitHubApp,
		Host:           "github.com",
		Status:         models.GitIntegrationStatusPendingInstall,
	}
	deps.oauthStates.EXPECT().Consume(gomock.Any(), "state-1:gi-app", models.OAuthProviderGitHubAppManifest).
		Return(&models.OAuthState{State: "state-1:gi-app", CreatedAt: time.Now().UTC()}, nil)
	deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
	deps.appClient.EXPECT().ConvertManifestCode(gomock.Any(), "tmp-code").Return(&githubapp.AppCredentials{
		AppID:         4242,
		Slug:          "stackdome-test",
		PEM:           "PEM-DATA",
		WebhookSecret: "hook-secret",
	}, nil)

	var stored *models.GitIntegration
	deps.store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
			stored = updated
			return updated, nil
		})

	redirect, serr := svc.HandleGitHubManifestCallback(context.Background(), "tmp-code", "state-1:gi-app")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if redirect != "https://github.com/apps/stackdome-test/installations/new" {
		t.Fatalf("unexpected redirect %q", redirect)
	}
	if stored.EncryptedAuth == "" || strings.Contains(stored.EncryptedAuth, "PEM-DATA") {
		t.Fatal("expected app credentials to be encrypted at rest")
	}
	if stored.Auth != nil {
		t.Fatal("transient auth must be cleared before storage")
	}
	if stored.Status != models.GitIntegrationStatusPendingInstall {
		t.Fatalf("expected pending_install until the webhook, got %q", stored.Status)
	}
}

func TestHandleGitHubManifestCallbackRejectsExpiredState(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)

	deps.oauthStates.EXPECT().Consume(gomock.Any(), "state-old:gi-app", models.OAuthProviderGitHubAppManifest).
		Return(&models.OAuthState{
			State:     "state-old:gi-app",
			CreatedAt: time.Now().UTC().Add(-time.Hour),
		}, nil)

	if _, serr := svc.HandleGitHubManifestCallback(context.Background(), "tmp-code", "state-old:gi-app"); serr == nil {
		t.Fatal("expected expired state to be rejected")
	}
}

func signWebhook(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func installationWebhookPayload(t *testing.T, action string) []byte {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"action": action,
		"installation": map[string]any{
			"id":     77,
			"app_id": 4242,
			"account": map[string]any{
				"login": "acme",
				"type":  models.GitAccountTypeOrganization,
			},
			"repository_selection": "all",
		},
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	return payload
}

func TestProcessGitHubWebhookInstallationCreated(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, Slug: "stackdome-test", WebhookSecret: "hook-secret"})
	integration.Status = models.GitIntegrationStatusPendingInstall
	payload := installationWebhookPayload(t, "created")

	deps.store.EXPECT().ListGitHubApps(gomock.Any()).Return([]*models.GitIntegration{integration}, nil)
	deps.installations.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, installation *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
			if installation.GitIntegrationID != "gi-app" || installation.InstallationID != 77 || installation.AccountLogin != "acme" {
				t.Fatalf("unexpected installation %+v", installation)
			}
			return installation, nil
		})
	deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").
		Return([]*models.GitInstallation{{InstallationID: 77}}, nil)
	deps.store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
			if updated.Status != models.GitIntegrationStatusInstalled {
				t.Fatalf("expected installed status, got %q", updated.Status)
			}
			return updated, nil
		})

	if serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, signWebhook(payload, "hook-secret")); serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
}

func TestProcessGitHubWebhookRejectsBadSignature(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, WebhookSecret: "hook-secret"})
	payload := installationWebhookPayload(t, "created")

	deps.store.EXPECT().ListGitHubApps(gomock.Any()).Return([]*models.GitIntegration{integration}, nil)

	serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, signWebhook(payload, "wrong-secret"))
	if serr == nil || !serr.IsForbidden() {
		t.Fatalf("expected forbidden on bad signature, got %v", serr)
	}
}

func TestProcessGitHubWebhookInstallationDeleted(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, WebhookSecret: "hook-secret"})
	payload := installationWebhookPayload(t, "deleted")

	deps.store.EXPECT().ListGitHubApps(gomock.Any()).Return([]*models.GitIntegration{integration}, nil)
	deps.installations.EXPECT().DeleteByInstallationID(gomock.Any(), "gi-app", int64(77)).Return(nil)
	deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return(nil, nil)
	deps.store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, updated *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
			if updated.Status != models.GitIntegrationStatusPendingInstall {
				t.Fatalf("expected pending_install after last uninstall, got %q", updated.Status)
			}
			return updated, nil
		})

	if serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, signWebhook(payload, "hook-secret")); serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
}

func TestProcessGitHubWebhookDropsPushEvents(t *testing.T) {
	svc, _ := newGitHubAppServiceForTest(t)
	if serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventPush, []byte(`{}`), ""); serr != nil {
		t.Fatalf("expected push events to be dropped, got %v", serr)
	}
}

func TestInternalMintForRepoMatchesOwner(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, Slug: "stackdome-test", PEM: "PEM"})
	expiresAt := time.Now().Add(time.Hour).UTC()

	deps.store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(integration, nil)
	deps.installations.EXPECT().GetByIntegrationAndAccount(gomock.Any(), "gi-app", "acme").
		Return(&models.GitInstallation{InstallationID: 77}, nil)
	deps.appClient.EXPECT().MintInstallationToken(gomock.Any(), gomock.Any(), int64(77)).
		Return(&githubapp.Token{Value: "ghs_minted", ExpiresAt: expiresAt}, nil)

	mint, serr := svc.InternalMintForRepo(context.Background(), "org-1", "https://github.com/acme/api")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if mint.Token != "ghs_minted" || !mint.ExpiresAt.Equal(expiresAt) || mint.IntegrationID != "gi-app" {
		t.Fatalf("unexpected mint %+v", mint)
	}
}

func TestInternalMintForRepoFallsThroughWhenNoInstallationCoversOwner(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242})

	deps.store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(integration, nil)
	deps.installations.EXPECT().GetByIntegrationAndAccount(gomock.Any(), "gi-app", "other").
		Return(nil, errors.NotFound("no installation for account '%s'", "other"))

	_, serr := svc.InternalMintForRepo(context.Background(), "org-1", "https://github.com/other/api")
	if serr == nil || !serr.Is404() {
		t.Fatalf("expected 404 to fall through, got %v", serr)
	}
}
