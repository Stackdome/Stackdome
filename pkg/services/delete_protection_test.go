package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestSecretDeleteBlockedReturns409(t *testing.T) {
	ctrl := gomock.NewController(t)

	secretStore := mocks.NewMockSecretStore(ctrl)
	secretStore.EXPECT().GetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", ProjectID: "project-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	refs := mocks.NewMockReferenceService(ctrl)
	refs.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentSecret, "sec-1").
		Return(true, []models.ResourceReference{{StackID: "stack-1", RelationKind: models.RelationImagePull}}, nil)

	svc := &secretService{
		secretStore:      secretStore,
		referenceService: refs,
		permissions:      permissions,
		logger:           logger.NewLogger(),
	}

	err := svc.Delete(context.Background(), "sec-1")
	if err == nil || !err.IsConflict() {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
}

func TestSecretDeleteAllowedWhenNotReferenced(t *testing.T) {
	ctrl := gomock.NewController(t)

	secretStore := mocks.NewMockSecretStore(ctrl)
	secretStore.EXPECT().GetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", ProjectID: "project-1"}, nil)
	secretStore.EXPECT().Delete(gomock.Any(), "sec-1").Return(nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	refs := mocks.NewMockReferenceService(ctrl)
	refs.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentSecret, "sec-1").
		Return(false, nil, nil)

	svc := &secretService{
		secretStore:      secretStore,
		referenceService: refs,
		permissions:      permissions,
		logger:           logger.NewLogger(),
	}

	if err := svc.Delete(context.Background(), "sec-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresAddonDeleteBlockedReturns409(t *testing.T) {
	ctrl := gomock.NewController(t)

	addonStore := mocks.NewMockPostgresAddonStore(ctrl)
	addonStore.EXPECT().GetByID(gomock.Any(), "pg-1").
		Return(&models.PostgresAddon{ID: "pg-1", ProjectID: "project-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	refs := mocks.NewMockReferenceService(ctrl)
	refs.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentPostgresAddon, "pg-1").
		Return(true, []models.ResourceReference{{StackID: "stack-1", RelationKind: models.RelationEnv}}, nil)

	svc := &postgresAddonService{
		postgresAddonStore: addonStore,
		referenceService:   refs,
		permissions:        permissions,
		logger:             logger.NewLogger(),
	}

	_, err := svc.DeletePostgresAddon(context.Background(), "pg-1")
	if err == nil || !err.IsConflict() {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
}

func TestVolumeDeleteBlockedReturns409(t *testing.T) {
	ctrl := gomock.NewController(t)

	volumeStore := mocks.NewMockVolumeStore(ctrl)
	volumeStore.EXPECT().GetByID(gomock.Any(), "vol-1").
		Return(&models.Volume{ID: "vol-1", Name: "uploads", ProjectID: "project-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	refs := mocks.NewMockReferenceService(ctrl)
	refs.EXPECT().IsReferentInUse(gomock.Any(), models.ReferentVolume, "vol-1").
		Return(true, []models.ResourceReference{{StackID: "stack-1", RelationKind: models.RelationVolumeMount}}, nil)

	svc := &volumeService{
		volumeStore:      volumeStore,
		referenceService: refs,
		permissions:      permissions,
		logger:           logger.NewLogger(),
	}

	err := svc.Delete(context.Background(), "vol-1")
	if err == nil || !err.IsConflict() {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
}
