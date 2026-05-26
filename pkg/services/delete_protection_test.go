package services

import (
	"context"
	"testing"

	"github.com/ashishmax31/stackdome-api-server/pkg/mocks"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"go.uber.org/mock/gomock"
)

func TestSecretDeleteIsBlockedByConnectionUsage(t *testing.T) {
	ctrl := gomock.NewController(t)

	secretStore := mocks.NewMockSecretStore(ctrl)
	secretStore.EXPECT().GetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", TeamID: "team-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	checker := NewMockconnectionUsageChecker(ctrl)
	checker.EXPECT().IsNodeReferencedAsSource(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil)

	svc := &secretService{
		secretStore:            secretStore,
		connectionUsageChecker: checker,
		resourceUsageService:   mocks.NewMockResourceUsageService(ctrl),
		permissions:            permissions,
	}

	err := svc.Delete(context.Background(), "sec-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: secret with ID 'sec-1' is in use by stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestSecretDeleteIsAllowedAfterConnectionRemoval(t *testing.T) {
	ctrl := gomock.NewController(t)

	secretStore := mocks.NewMockSecretStore(ctrl)
	secretStore.EXPECT().GetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", TeamID: "team-1"}, nil)
	secretStore.EXPECT().Delete(gomock.Any(), "sec-1").Return(nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	checker := NewMockconnectionUsageChecker(ctrl)
	checker.EXPECT().IsNodeReferencedAsSource(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, nil)

	resourceUsage := mocks.NewMockResourceUsageService(ctrl)
	resourceUsage.EXPECT().IsResourceInUse(gomock.Any(), "secret", "sec-1").
		Return(false, nil)

	svc := &secretService{
		secretStore:            secretStore,
		connectionUsageChecker: checker,
		resourceUsageService:   resourceUsage,
		permissions:            permissions,
	}

	if err := svc.Delete(context.Background(), "sec-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSecretDeleteIsBlockedByDirectConfigUsage(t *testing.T) {
	ctrl := gomock.NewController(t)

	secretStore := mocks.NewMockSecretStore(ctrl)
	secretStore.EXPECT().GetByID(gomock.Any(), "sec-1").
		Return(&models.Secret{ID: "sec-1", TeamID: "team-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	checker := NewMockconnectionUsageChecker(ctrl)
	checker.EXPECT().IsNodeReferencedAsSource(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, nil)

	resourceUsage := mocks.NewMockResourceUsageService(ctrl)
	resourceUsage.EXPECT().IsResourceInUse(gomock.Any(), "secret", "sec-1").
		Return(true, nil)

	svc := &secretService{
		secretStore:            secretStore,
		connectionUsageChecker: checker,
		resourceUsageService:   resourceUsage,
		permissions:            permissions,
	}

	err := svc.Delete(context.Background(), "sec-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: secret with ID 'sec-1' is in use by stacks"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestPostgresAddonDeleteIsBlockedByConnectionUsage(t *testing.T) {
	ctrl := gomock.NewController(t)

	addonStore := mocks.NewMockPostgresAddonStore(ctrl)
	addonStore.EXPECT().GetByID(gomock.Any(), "pg-1").
		Return(&models.PostgresAddon{ID: "pg-1", TeamID: "team-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	checker := NewMockconnectionUsageChecker(ctrl)
	checker.EXPECT().IsNodeReferencedAsSource(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(true, nil)

	svc := &postgresAddonService{
		postgresAddonStore:     addonStore,
		connectionUsageChecker: checker,
		permissions:            permissions,
	}

	_, err := svc.DeletePostgresAddon(context.Background(), "pg-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: addon is in use by one or more stacks and cannot be deleted"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestVolumeDeleteIsBlockedByVolumeMountConnectionSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := newVolumeServiceForDeleteProtectionTest(ctrl, true)

	err := svc.Delete(context.Background(), "vol-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: volume is in use by one or more stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func TestVolumeDeleteIsBlockedByBuildArtifactConnectionTarget(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := newVolumeServiceForDeleteProtectionTest(ctrl, true)

	err := svc.Delete(context.Background(), "vol-1")
	if err == nil {
		t.Fatalf("expected delete to be blocked")
	}
	if got, want := err.Error(), "error: volume is in use by one or more stack connections"; got != want {
		t.Fatalf("unexpected error: got %q want %q", got, want)
	}
}

func newVolumeServiceForDeleteProtectionTest(ctrl *gomock.Controller, referenced bool) *volumeService {
	volumeStore := mocks.NewMockVolumeStore(ctrl)
	volumeStore.EXPECT().GetByID(gomock.Any(), "vol-1").
		Return(&models.Volume{ID: "vol-1", Name: "uploads", TeamID: "team-1"}, nil)

	permissions := mocks.NewMockPermissionService(ctrl)
	permissions.EXPECT().Check(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)

	stackVolumeStore := mocks.NewMockStackVolumeStore(ctrl)
	stackVolumeStore.EXPECT().GetByVolumeID(gomock.Any(), "vol-1").
		Return(&models.StackVolume{StackID: "stack-1"}, nil)

	checker := NewMockconnectionUsageChecker(ctrl)
	checker.EXPECT().IsNodeReferenced(gomock.Any(), "stack-1", gomock.Any()).
		Return(referenced, nil)

	return &volumeService{
		volumeStore:            volumeStore,
		stackVolumeStore:       stackVolumeStore,
		connectionUsageChecker: checker,
		clusterResourceService: mocks.NewMockVolumeClusterResourceService(ctrl),
		logger:                 mocks.NewMockLogger(ctrl),
		permissions:            permissions,
	}
}
