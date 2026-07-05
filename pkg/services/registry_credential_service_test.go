package services

import (
	"context"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/go-containerregistry/pkg/name"
	"go.uber.org/mock/gomock"
)

func newTestEncryptionService(t *testing.T) EncryptionService {
	t.Helper()
	encryption, err := NewAESEncryptionService(EncryptionServiceSpec{
		Masterkey: strings.Repeat("k", 64),
	})
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}
	return encryption
}

type registryCredentialServiceMocks struct {
	store            *mocks.MockRegistryCredentialStore
	stackStore       *mocks.MockStackStore
	referenceService *mocks.MockReferenceService
	registryClients  *MockverifyRegistryClientProvider
	encryption       EncryptionService
}

func newRegistryCredentialServiceForTest(t *testing.T) (RegistryCredentialService, *registryCredentialServiceMocks) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	deps := &registryCredentialServiceMocks{
		store:            mocks.NewMockRegistryCredentialStore(ctrl),
		stackStore:       mocks.NewMockStackStore(ctrl),
		referenceService: mocks.NewMockReferenceService(ctrl),
		registryClients:  NewMockverifyRegistryClientProvider(ctrl),
		encryption:       newTestEncryptionService(t),
	}

	svc := NewRegistryCredentialService(RegistryCredentialServiceSpec{
		Store:             deps.store,
		StackStore:        deps.stackStore,
		ReferenceService:  deps.referenceService,
		EncryptionService: deps.encryption,
		Permissions:       permissions,
		RegistryClients:   deps.registryClients,
	})
	return svc, deps
}

func TestRegistryCredentialCreateEncryptsPasswordAtRest(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "ghcr.io").Return(nil, nil)

	var stored *models.RegistryCredential
	deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
			stored = credential
			return credential, nil
		})

	created, serr := svc.Create(context.Background(), &models.RegistryCredential{
		OrganisationID: "org-1",
		Host:           "ghcr.io",
		Username:       "acme-bot",
		Password:       "hunter2",
	})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}

	if stored.EncryptedPassword == "" || stored.EncryptedPassword == "hunter2" {
		t.Fatalf("expected password to be encrypted at rest, got %q", stored.EncryptedPassword)
	}
	if strings.Contains(stored.EncryptedPassword, "hunter2") {
		t.Fatal("stored row contains the plaintext password")
	}
	if stored.Password != "" || created.Password != "" {
		t.Fatal("transient password must be cleared before storage")
	}
	if stored.DataHash == "" {
		t.Fatal("expected data hash to be stamped")
	}

	decrypted, err := deps.encryption.DecryptData(stored.EncryptedPassword)
	if err != nil {
		t.Fatalf("failed to decrypt stored password: %v", err)
	}
	if string(decrypted) != "hunter2" {
		t.Fatalf("decrypted password mismatch: got %q", decrypted)
	}
	if stored.Purpose != models.RegistryCredentialPurposeBoth {
		t.Fatalf("expected purpose to default to both, got %q", stored.Purpose)
	}
}

func TestRegistryCredentialCreateNormalizesHost(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", name.DefaultRegistry).Return(nil, nil)
	deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
			return credential, nil
		})

	created, serr := svc.Create(context.Background(), &models.RegistryCredential{
		OrganisationID: "org-1",
		Host:           "docker.io",
		Username:       "acme-bot",
		Password:       "hunter2",
	})
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if created.Host != name.DefaultRegistry {
		t.Fatalf("expected docker.io to normalize to %s, got %q", name.DefaultRegistry, created.Host)
	}
}

func TestRegistryCredentialCreatePurposeConflicts(t *testing.T) {
	cases := []struct {
		name     string
		existing models.RegistryCredentialPurpose
		new      models.RegistryCredentialPurpose
		conflict bool
	}{
		{name: "both conflicts with pull", existing: models.RegistryCredentialPurposeBoth, new: models.RegistryCredentialPurposePull, conflict: true},
		{name: "pull conflicts with both", existing: models.RegistryCredentialPurposePull, new: models.RegistryCredentialPurposeBoth, conflict: true},
		{name: "duplicate pull conflicts", existing: models.RegistryCredentialPurposePull, new: models.RegistryCredentialPurposePull, conflict: true},
		{name: "pull and push coexist", existing: models.RegistryCredentialPurposePull, new: models.RegistryCredentialPurposePush, conflict: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, deps := newRegistryCredentialServiceForTest(t)

			deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "ghcr.io").
				Return([]*models.RegistryCredential{{Host: "ghcr.io", Purpose: tc.existing}}, nil)
			if !tc.conflict {
				deps.store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
					func(_ context.Context, credential *models.RegistryCredential) (*models.RegistryCredential, *errors.ServiceError) {
						return credential, nil
					})
			}

			_, serr := svc.Create(context.Background(), &models.RegistryCredential{
				OrganisationID: "org-1",
				Host:           "ghcr.io",
				Username:       "acme-bot",
				Password:       "hunter2",
				Purpose:        tc.new,
			})
			if tc.conflict {
				if serr == nil || !serr.IsConflict() {
					t.Fatalf("expected 409 conflict, got %v", serr)
				}
			} else if serr != nil {
				t.Fatalf("unexpected error: %v", serr)
			}
		})
	}
}

func TestRegistryCredentialDeleteReportsAffectedStacks(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)

	credential := &models.RegistryCredential{
		ID:             "rc-1",
		OrganisationID: "org-1",
		Host:           "ghcr.io",
		Purpose:        models.RegistryCredentialPurposeBoth,
	}
	deps.store.EXPECT().GetByID(gomock.Any(), "rc-1").Return(credential, nil)
	deps.stackStore.EXPECT().ListByOrganisationID(gomock.Any(), "org-1").Return([]*models.Stack{
		{
			ID: "stack-implicit-pull", Name: "implicit-pull",
			StackResources: []*models.StackResource{
				{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "ghcr.io/acme/app:1"}},
			},
		},
		{
			ID: "stack-explicit-ref", Name: "explicit-ref",
			StackResources: []*models.StackResource{
				{Name: "web", ImageConfig: &models.ImageConfigSpec{
					Image:                "ghcr.io/acme/app:1",
					RegistryCredentialID: "rc-other",
				}},
			},
		},
		{
			ID: "stack-implicit-push", Name: "implicit-push",
			StackResources: []*models.StackResource{
				{Name: "builder", BuildConfig: &models.BuildConfigSpec{
					BuildImageRepository: models.BuildImageRepository{ExternalImageRef: "ghcr.io/acme/api"},
				}},
			},
		},
		{
			ID: "stack-other-host", Name: "other-host",
			StackResources: []*models.StackResource{
				{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
			},
		},
		// Referenced only via an explicit ref recorded in a retained release;
		// its live spec no longer resolves against the credential (host match
		// only lives in the retained snapshot), so the implicit scan misses it.
		{
			ID: "stack-retained", Name: "retained",
			StackResources: []*models.StackResource{
				{Name: "web", ImageConfig: &models.ImageConfigSpec{Image: "nginx:latest"}},
			},
		},
	}, nil)

	retainedReleaseID := "rel-1"
	deps.referenceService.EXPECT().
		IsReferentInUse(gomock.Any(), models.ReferentRegistryCredential, "rc-1").
		Return(true, []models.ResourceReference{
			// Explicit ref from a RETAINED RELEASE (ReleaseID set) — proves the
			// release dimension is covered, not just the live spec.
			{StackID: "stack-retained", ReleaseID: &retainedReleaseID, ReferentType: models.ReferentRegistryCredential, ReferentID: "rc-1", RelationKind: models.RelationImagePull},
		}, nil)

	deps.store.EXPECT().Delete(gomock.Any(), "rc-1").Return(nil)

	result, serr := svc.Delete(context.Background(), "rc-1")
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}

	got := make(map[string]string, len(result.AffectedStacks))
	for _, stack := range result.AffectedStacks {
		got[stack.ID] = stack.Name
	}
	want := map[string]string{
		"stack-implicit-pull": "implicit-pull",
		"stack-implicit-push": "implicit-push",
		"stack-retained":      "retained",
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d affected stacks, got %v", len(want), result.AffectedStacks)
	}
	for id, name := range want {
		if got[id] != name {
			t.Fatalf("expected affected stack %q with name %q, got %q (all: %v)", id, name, got[id], result.AffectedStacks)
		}
	}
}

func TestRegistryCredentialVerifyRunsProbes(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	encrypted, err := deps.encryption.EncryptData([]byte("hunter2"))
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}
	credential := &models.RegistryCredential{
		ID:                "rc-1",
		OrganisationID:    "org-1",
		Host:              "ghcr.io",
		Purpose:           models.RegistryCredentialPurposeBoth,
		Username:          "acme-bot",
		EncryptedPassword: encrypted,
	}
	deps.store.EXPECT().GetByID(gomock.Any(), "rc-1").Return(credential, nil)

	registryClient := mocks.NewMockRegistryClient(ctrl)
	registryClient.EXPECT().CheckImage(gomock.Any(), "ghcr.io/acme/app").Return(true, nil)
	registryClient.EXPECT().CheckPushAccess(gomock.Any(), "ghcr.io/acme/app").Return(nil)
	deps.registryClients.EXPECT().ClientFor("acme-bot", "hunter2").Return(registryClient, nil)

	if serr := svc.Verify(context.Background(), "rc-1", "ghcr.io/acme/app", ""); serr != nil {
		t.Fatalf("expected verify to pass, got %v", serr)
	}
}

func TestRegistryCredentialVerifyRejectsHostMismatch(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)

	deps.store.EXPECT().GetByID(gomock.Any(), "rc-1").Return(&models.RegistryCredential{
		ID:             "rc-1",
		OrganisationID: "org-1",
		Host:           "ghcr.io",
		Purpose:        models.RegistryCredentialPurposeBoth,
	}, nil)

	serr := svc.Verify(context.Background(), "rc-1", "quay.io/acme/app", "")
	if serr == nil {
		t.Fatal("expected host mismatch to be rejected")
	}
}

func TestInternalGetForHostMatchesPurpose(t *testing.T) {
	svc, deps := newRegistryCredentialServiceForTest(t)

	encrypted, err := deps.encryption.EncryptData([]byte("hunter2"))
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}
	pullCred := &models.RegistryCredential{
		ID:                "rc-pull",
		Host:              "ghcr.io",
		Purpose:           models.RegistryCredentialPurposePull,
		Username:          "acme-bot",
		EncryptedPassword: encrypted,
		DataHash:          registryCredentialDataHash("acme-bot", "hunter2"),
	}

	deps.store.EXPECT().GetByOrgAndHost(gomock.Any(), "org-1", "ghcr.io").
		Return([]*models.RegistryCredential{pullCred}, nil).Times(2)

	got, serr := svc.InternalGetForHost(context.Background(), "org-1", "ghcr.io", models.RegistryCredentialPurposePull)
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if got.Password != "hunter2" {
		t.Fatalf("expected decrypted password, got %q", got.Password)
	}

	if _, serr := svc.InternalGetForHost(context.Background(), "org-1", "ghcr.io", models.RegistryCredentialPurposePush); serr == nil || !serr.Is404() {
		t.Fatalf("expected 404 for uncovered purpose, got %v", serr)
	}
}
