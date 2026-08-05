package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Stackdome/stackdome/pkg/clients/githubapp"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

type githubAppServiceMocks struct {
	store         *mocks.MockGitIntegrationStore
	installations *mocks.MockGitInstallationStore
	oauthStates   *mocks.MockOAuthStateStore
	organisations *mocks.MockOrganisationStore
	atomic        *mocks.MockAtomicExecutor
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
		atomic:        mocks.NewMockAtomicExecutor(ctrl),
		appClient:     mocks.NewMockGitHubAppClient(ctrl),
		encryption:    newTestEncryptionService(t),
	}
	deps.organisations.EXPECT().Get(gomock.Any(), gomock.Any()).
		Return(&models.Organisation{ID: "org-1", Name: "Acme Corp"}, nil).AnyTimes()
	// Pass-through: run the transaction body against the mocked stores.
	deps.atomic.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
			return fn(ctx)
		}).AnyTimes()
	svc := NewGitIntegrationService(GitIntegrationServiceSpec{
		Store:             deps.store,
		InstallationStore: deps.installations,
		OAuthStateStore:   deps.oauthStates,
		OrganisationStore: deps.organisations,
		AtomicExecutor:    deps.atomic,
		GitHubAppClient:   deps.appClient,
		EncryptionService: deps.encryption,
		Permissions:       permissions,
		Logger:            logger.NewLogger(),
		ExternalURL:       "https://hub.example.com",
	})
	return svc, deps
}

func newGitHubWebhookServiceForTest(t *testing.T) (GitHubWebhookService, *githubAppServiceMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	deps := &githubAppServiceMocks{
		store:         mocks.NewMockGitIntegrationStore(ctrl),
		installations: mocks.NewMockGitInstallationStore(ctrl),
		encryption:    newTestEncryptionService(t),
	}
	svc := NewGitHubWebhookService(GitHubWebhookServiceSpec{
		Store:             deps.store,
		InstallationStore: deps.installations,
		EncryptionService: deps.encryption,
		PreviewWebhook:    NewMockPreviewWebhookService(ctrl),
		Logger:            logger.NewLogger(),
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
		Return(&models.OAuthState{State: "state-1:org-1:gi-app", CreatedAt: time.Now().UTC()}, nil)
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

func TestListInstallationsRefreshPrunesSuspendedAndStale(t *testing.T) {
	svc, deps := newGitHubAppServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, Slug: "s", PEM: "PEM"})

	deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
	deps.appClient.EXPECT().ListInstallations(gomock.Any(), gomock.Any()).Return([]githubapp.Installation{
		{ID: 77, AccountLogin: "acme", AccountType: string(models.GitAccountTypeOrganization), RepositorySelection: "all"},
		{ID: 88, AccountLogin: "suspended-acct", Suspended: true},
	}, nil)

	// Only the live installation is upserted; the suspended one is skipped.
	deps.installations.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, in *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
			if in.InstallationID != 77 {
				t.Fatalf("only the live installation should be upserted, got %d", in.InstallationID)
			}
			return in, nil
		})
	// Local table holds a stale row (99) and the now-suspended one (88); both pruned.
	deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return([]*models.GitInstallation{
		{InstallationID: 77}, {InstallationID: 99}, {InstallationID: 88},
	}, nil).AnyTimes()
	deps.installations.EXPECT().DeleteByInstallationID(gomock.Any(), "gi-app", int64(99)).Return(nil)
	deps.installations.EXPECT().DeleteByInstallationID(gomock.Any(), "gi-app", int64(88)).Return(nil)

	if _, serr := svc.ListInstallations(context.Background(), "gi-app", true); serr != nil {
		t.Fatalf("unexpected error: %v", serr)
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
	svc, deps := newGitHubWebhookServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, Slug: "stackdome-test", WebhookSecret: "hook-secret"})
	integration.Status = models.GitIntegrationStatusPendingInstall
	payload := installationWebhookPayload(t, "created")

	deps.installations.EXPECT().GetByInstallationID(gomock.Any(), int64(77)).Return(nil, errors.NotFound("no installation row yet"))
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
	svc, deps := newGitHubWebhookServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, WebhookSecret: "hook-secret"})
	payload := installationWebhookPayload(t, "created")

	deps.installations.EXPECT().GetByInstallationID(gomock.Any(), int64(77)).Return(nil, errors.NotFound("no installation row yet"))
	deps.store.EXPECT().ListGitHubApps(gomock.Any()).Return([]*models.GitIntegration{integration}, nil)

	serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, signWebhook(payload, "wrong-secret"))
	if serr == nil || !serr.IsForbidden() {
		t.Fatalf("expected forbidden on bad signature, got %v", serr)
	}
}

func TestProcessGitHubWebhookInstallationDeleted(t *testing.T) {
	svc, deps := newGitHubWebhookServiceForTest(t)
	integration := sealedGitHubApp(t, deps.encryption, models.GitHubAppCredentials{AppID: 4242, WebhookSecret: "hook-secret"})
	payload := installationWebhookPayload(t, "deleted")

	deps.installations.EXPECT().GetByInstallationID(gomock.Any(), int64(77)).Return(nil, errors.NotFound("no installation row yet"))
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
	svc, _ := newGitHubWebhookServiceForTest(t)
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

var _ = Describe("platform GitHub App", func() {
	const platformSlug = "stackdome-cloud"

	var (
		ctrl     *gomock.Controller
		svc      GitIntegrationService
		deps     *githubAppServiceMocks
		platform *githubapp.AppCredentials
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).NotTo(HaveOccurred())

		permissions := mocks.NewMockPermissionService(ctrl)
		permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		deps = &githubAppServiceMocks{
			store:         mocks.NewMockGitIntegrationStore(ctrl),
			installations: mocks.NewMockGitInstallationStore(ctrl),
			oauthStates:   mocks.NewMockOAuthStateStore(ctrl),
			organisations: mocks.NewMockOrganisationStore(ctrl),
			atomic:        mocks.NewMockAtomicExecutor(ctrl),
			appClient:     mocks.NewMockGitHubAppClient(ctrl),
			encryption:    encryption,
		}
		deps.atomic.EXPECT().WithTransaction(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) *errors.ServiceError) *errors.ServiceError {
				return fn(ctx)
			}).AnyTimes()

		platform = &githubapp.AppCredentials{
			AppID:         4242,
			Slug:          platformSlug,
			PEM:           "-----BEGIN RSA PRIVATE KEY-----",
			WebhookSecret: "hook-secret",
		}
		svc = NewGitIntegrationService(GitIntegrationServiceSpec{
			Store:             deps.store,
			InstallationStore: deps.installations,
			OAuthStateStore:   deps.oauthStates,
			OrganisationStore: deps.organisations,
			AtomicExecutor:    deps.atomic,
			GitHubAppClient:   deps.appClient,
			EncryptionService: encryption,
			Permissions:       permissions,
			Logger:            logger.NewLogger(),
			ExternalURL:       "https://hub.example.com",
			PlatformApp:       platform,
		})
	})

	Describe("CreateGitHubAppManifest", func() {
		It("skips app creation and sends the user to the install page", func() {
			deps.store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(nil, errors.NotFound("none"))
			deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&models.GitIntegration{
				ID:             "gi-app",
				OrganisationID: "org-1",
				Type:           models.GitIntegrationTypeGitHubApp,
				Status:         models.GitIntegrationStatusPendingInstall,
			}, nil)

			var stored *models.OAuthState
			deps.oauthStates.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, state *models.OAuthState) *errors.ServiceError {
					stored = state
					return nil
				})

			flow, serr := svc.CreateGitHubAppManifest(context.Background(), "org-1")
			Expect(serr).To(BeNil())
			Expect(flow.Manifest).To(BeNil())
			Expect(flow.GitHubURL).To(Equal("https://github.com/apps/" + platformSlug + "/installations/new?state=" + flow.State))
			Expect(stored.Provider).To(Equal(models.OAuthProviderGitHubAppInstall))
			Expect(stored.State).To(HaveSuffix(":gi-app"))
		})
	})

	Describe("HandleGitHubAppSetup", func() {
		var integration *models.GitIntegration

		BeforeEach(func() {
			// Platform-backed rows carry no credentials; they resolve from config.
			integration = &models.GitIntegration{
				ID:             "gi-app",
				OrganisationID: "org-1",
				Type:           models.GitIntegrationTypeGitHubApp,
				Status:         models.GitIntegrationStatusPendingInstall,
				DataHash:       gitIntegrationDataHash(nil),
			}
		})

		It("binds the new installation to the org that started the flow", func() {
			deps.oauthStates.EXPECT().Consume(gomock.Any(), "state-1", models.OAuthProviderGitHubAppInstall).
				Return(&models.OAuthState{State: "uuid:org-1:gi-app", CreatedAt: time.Now().UTC()}, nil)
			deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
			deps.appClient.EXPECT().GetInstallation(gomock.Any(), gomock.Any(), int64(77)).Return(&githubapp.Installation{
				ID: 77, AccountLogin: "acme", AccountType: string(models.GitAccountTypeOrganization), RepositorySelection: "all",
			}, nil)

			var upserted *models.GitInstallation
			deps.installations.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
					upserted = in
					return in, nil
				})
			deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").
				Return([]*models.GitInstallation{{InstallationID: 77}}, nil)
			deps.store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
					return in, nil
				})

			redirect, serr := svc.HandleGitHubAppSetup(context.Background(), 77, "state-1")
			Expect(serr).To(BeNil())
			Expect(redirect).To(Equal("https://hub.example.com/git-integrations?setup_action=install"))
			Expect(upserted.GitIntegrationID).To(Equal("gi-app"))
			Expect(upserted.AccountLogin).To(Equal("acme"))
			Expect(integration.Status).To(Equal(models.GitIntegrationStatusInstalled))
		})

		It("falls back to the org's current row when the state's row was deleted mid-flow", func() {
			recreated := &models.GitIntegration{
				ID:             "gi-new",
				OrganisationID: "org-1",
				Type:           models.GitIntegrationTypeGitHubApp,
				Status:         models.GitIntegrationStatusPendingInstall,
				DataHash:       gitIntegrationDataHash(nil),
			}
			deps.oauthStates.EXPECT().Consume(gomock.Any(), "state-1", models.OAuthProviderGitHubAppInstall).
				Return(&models.OAuthState{State: "uuid:org-1:gi-app", CreatedAt: time.Now().UTC()}, nil)
			deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(nil, errors.NotFound("gone"))
			deps.store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(recreated, nil)
			deps.appClient.EXPECT().GetInstallation(gomock.Any(), gomock.Any(), int64(77)).Return(&githubapp.Installation{
				ID: 77, AccountLogin: "acme", AccountType: string(models.GitAccountTypeOrganization), RepositorySelection: "all",
			}, nil)

			var upserted *models.GitInstallation
			deps.installations.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
					upserted = in
					return in, nil
				})
			deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-new").
				Return([]*models.GitInstallation{{InstallationID: 77}}, nil)
			deps.store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
					return in, nil
				})

			_, serr := svc.HandleGitHubAppSetup(context.Background(), 77, "state-1")
			Expect(serr).To(BeNil())
			Expect(upserted.GitIntegrationID).To(Equal("gi-new"))
		})

		It("rejects an installation the platform app does not have", func() {
			deps.oauthStates.EXPECT().Consume(gomock.Any(), "state-1", models.OAuthProviderGitHubAppInstall).
				Return(&models.OAuthState{State: "uuid:org-1:gi-app", CreatedAt: time.Now().UTC()}, nil)
			deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
			deps.appClient.EXPECT().GetInstallation(gomock.Any(), gomock.Any(), int64(77)).Return(nil, fmt.Errorf("404 not found"))

			_, serr := svc.HandleGitHubAppSetup(context.Background(), 77, "state-1")
			Expect(serr).NotTo(BeNil())
			Expect(serr.Is404()).To(BeTrue())
		})

		It("leaves another org's installation of the same app alone on refresh", func() {
			integration.Status = models.GitIntegrationStatusInstalled
			deps.store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
			deps.appClient.EXPECT().ListInstallations(gomock.Any(), gomock.Any()).Return([]githubapp.Installation{
				{ID: 77, AccountLogin: "acme", AccountType: string(models.GitAccountTypeOrganization), RepositorySelection: "all"},
				{ID: 99, AccountLogin: "other-org", AccountType: string(models.GitAccountTypeOrganization), RepositorySelection: "all"},
			}, nil)
			deps.installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").
				Return([]*models.GitInstallation{{InstallationID: 77}}, nil).AnyTimes()
			deps.installations.EXPECT().Upsert(gomock.Any(), gomock.Any()).DoAndReturn(
				func(_ context.Context, in *models.GitInstallation) (*models.GitInstallation, *errors.ServiceError) {
					Expect(in.InstallationID).To(Equal(int64(77)))
					return in, nil
				})

			_, serr := svc.ListInstallations(context.Background(), "gi-app", true)
			Expect(serr).To(BeNil())
		})
	})
})

var _ = Describe("platform GitHub App webhooks", func() {
	It("resolves deliveries via the installation row and verifies with the platform secret", func() {
		ctrl := gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).NotTo(HaveOccurred())

		platform := &githubapp.AppCredentials{
			AppID:         4242,
			Slug:          "stackdome-cloud",
			PEM:           "-----BEGIN RSA PRIVATE KEY-----",
			WebhookSecret: "platform-hook-secret",
		}
		store := mocks.NewMockGitIntegrationStore(ctrl)
		installations := mocks.NewMockGitInstallationStore(ctrl)
		svc := NewGitHubWebhookService(GitHubWebhookServiceSpec{
			Store:             store,
			InstallationStore: installations,
			EncryptionService: encryption,
			PreviewWebhook:    NewMockPreviewWebhookService(ctrl),
			Logger:            logger.NewLogger(),
			PlatformApp:       platform,
		})

		// Platform-backed row: no sealed auth of its own.
		integration := &models.GitIntegration{
			ID:             "gi-app",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitHubApp,
			Status:         models.GitIntegrationStatusInstalled,
		}
		payload, err := json.Marshal(map[string]any{
			"action": "deleted",
			"installation": map[string]any{
				"id":     77,
				"app_id": 4242,
				"account": map[string]any{
					"login": "acme",
					"type":  models.GitAccountTypeOrganization,
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())

		installations.EXPECT().GetByInstallationID(gomock.Any(), int64(77)).
			Return(&models.GitInstallation{GitIntegrationID: "gi-app", InstallationID: 77}, nil)
		store.EXPECT().GetByID(gomock.Any(), "gi-app").Return(integration, nil)
		installations.EXPECT().DeleteByInstallationID(gomock.Any(), "gi-app", int64(77)).Return(nil)
		installations.EXPECT().ListByIntegrationID(gomock.Any(), "gi-app").Return(nil, nil)
		store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, in *models.GitIntegration) (*models.GitIntegration, *errors.ServiceError) {
				Expect(in.Status).To(Equal(models.GitIntegrationStatusPendingInstall))
				return in, nil
			})

		serr := svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, signWebhookGinkgo(payload, platform.WebhookSecret))
		Expect(serr).To(BeNil())
	})

	It("acks an installation event with no binding yet instead of erroring", func() {
		ctrl := gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).NotTo(HaveOccurred())

		store := mocks.NewMockGitIntegrationStore(ctrl)
		installations := mocks.NewMockGitInstallationStore(ctrl)
		svc := NewGitHubWebhookService(GitHubWebhookServiceSpec{
			Store:             store,
			InstallationStore: installations,
			EncryptionService: encryption,
			PreviewWebhook:    NewMockPreviewWebhookService(ctrl),
			Logger:            logger.NewLogger(),
			PlatformApp:       &githubapp.AppCredentials{AppID: 4242, Slug: "stackdome-cloud", PEM: "k", WebhookSecret: "s"},
		})

		payload, err := json.Marshal(map[string]any{
			"action":       "created",
			"installation": map[string]any{"id": 88, "app_id": 4242},
		})
		Expect(err).NotTo(HaveOccurred())

		// Setup callback hasn't bound installation 88 yet; nothing sealed to
		// match by app id either. The delivery must be acked, not retried.
		installations.EXPECT().GetByInstallationID(gomock.Any(), int64(88)).
			Return(nil, errors.NotFound("not bound yet"))
		store.EXPECT().ListGitHubApps(gomock.Any()).Return(nil, nil)

		Expect(svc.ProcessGitHubWebhook(context.Background(), GitHubEventInstallation, payload, "sha256=irrelevant")).To(BeNil())
	})
})

func signWebhookGinkgo(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

var _ = Describe("platform GitHub App re-run", func() {
	It("allows starting the flow again when already installed, to add another account", func() {
		ctrl := gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)

		encryption, err := NewAESEncryptionService(EncryptionServiceSpec{Masterkey: strings.Repeat("k", 64)})
		Expect(err).NotTo(HaveOccurred())
		permissions := mocks.NewMockPermissionService(ctrl)
		permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

		store := mocks.NewMockGitIntegrationStore(ctrl)
		oauthStates := mocks.NewMockOAuthStateStore(ctrl)
		svc := NewGitIntegrationService(GitIntegrationServiceSpec{
			Store:             store,
			InstallationStore: mocks.NewMockGitInstallationStore(ctrl),
			OAuthStateStore:   oauthStates,
			OrganisationStore: mocks.NewMockOrganisationStore(ctrl),
			AtomicExecutor:    mocks.NewMockAtomicExecutor(ctrl),
			GitHubAppClient:   mocks.NewMockGitHubAppClient(ctrl),
			EncryptionService: encryption,
			Permissions:       permissions,
			Logger:            logger.NewLogger(),
			ExternalURL:       "https://hub.example.com",
			PlatformApp:       &githubapp.AppCredentials{AppID: 4242, Slug: "stackdome-cloud", PEM: "k", WebhookSecret: "s"},
		})

		store.EXPECT().GetGitHubAppForOrg(gomock.Any(), "org-1").Return(&models.GitIntegration{
			ID:             "gi-app",
			OrganisationID: "org-1",
			Type:           models.GitIntegrationTypeGitHubApp,
			Status:         models.GitIntegrationStatusInstalled,
		}, nil)
		oauthStates.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		flow, serr := svc.CreateGitHubAppManifest(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(flow.GitHubURL).To(ContainSubstring("/apps/stackdome-cloud/installations/new?state="))
	})
})
