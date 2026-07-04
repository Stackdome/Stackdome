package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/validator/secret"
	"go.uber.org/mock/gomock"
)

func newSecretServiceForManagedTest(t *testing.T) (*secretService, *mocks.MockSecretStore) {
	t.Helper()
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	store := mocks.NewMockSecretStore(ctrl)
	svc := &secretService{
		secretStore:       store,
		validator:         secret.NewSecretValidator(),
		encryptionService: newTestEncryptionService(t),
	}
	return svc, store
}

func managedGitSpec() ManagedSecretSpec {
	return ManagedSecretSpec{
		OrganisationID: "org-1",
		TeamID:         "team-1",
		UserID:         "user-1",
		OwnerKind:      models.ManagedByKindStackResource,
		OwnerID:        models.StackResourceManagedOwnerID("stack-1", "api"),
		Slot:           models.ManagedSecretSlotGit,
		Name:           models.ManagedSecretName("demo", "api", models.ManagedSecretSlotGit),
		Type:           models.SecretTypeGitCredentials,
		Data: map[string]string{
			models.UsernameSecretKey: "alice",
			models.PasswordSecretKey: "s3cret",
		},
	}
}

func TestCreateManagedCreatesWhenNoneExists(t *testing.T) {
	svc, store := newSecretServiceForManagedTest(t)
	spec := managedGitSpec()

	store.EXPECT().GetManagedByOwner(gomock.Any(), spec.OwnerKind, spec.OwnerID, spec.Slot).
		Return(nil, errors.NotFound("none"))

	var stored *models.Secret
	store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s *models.Secret) (*models.Secret, *errors.ServiceError) {
			stored = s
			return s, nil
		})

	created, serr := svc.CreateManaged(context.Background(), spec)
	if serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
	if !stored.Managed || stored.ManagedByKind != spec.OwnerKind || stored.ManagedByID != spec.OwnerID || stored.ManagedSlot != spec.Slot {
		t.Fatalf("expected managed ownership to be stamped, got %+v", stored)
	}
	if stored.EncryptedData == "" || stored.Data != nil {
		t.Fatal("expected data to be encrypted and plaintext cleared")
	}
	if created.Name != spec.Name {
		t.Fatalf("unexpected name %q", created.Name)
	}
}

func TestCreateManagedUpsertsExistingSecret(t *testing.T) {
	svc, store := newSecretServiceForManagedTest(t)
	spec := managedGitSpec()

	store.EXPECT().GetManagedByOwner(gomock.Any(), spec.OwnerKind, spec.OwnerID, spec.Slot).
		Return(&models.Secret{ID: "existing-1"}, nil)
	store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, s *models.Secret) (*models.Secret, *errors.ServiceError) {
			if s.ID != "existing-1" {
				t.Fatalf("expected update of existing secret, got id %q", s.ID)
			}
			return s, nil
		})

	if _, serr := svc.CreateManaged(context.Background(), spec); serr != nil {
		t.Fatalf("unexpected error: %v", serr)
	}
}
