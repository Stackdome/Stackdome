package services

import (
	"context"
	"testing"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

func newPreviewConfigDeleteService(ctrl *gomock.Controller) (*stackPreviewConfigService, *mocks.MockStackPreviewConfigStore, *mocks.MockPreviewStackStore, *mocks.MockPermissionService) {
	store := mocks.NewMockStackPreviewConfigStore(ctrl)
	previewStore := mocks.NewMockPreviewStackStore(ctrl)
	permissions := mocks.NewMockPermissionService(ctrl)
	svc := &stackPreviewConfigService{
		store:             store,
		previewStackStore: previewStore,
		permissions:       permissions,
	}
	return svc, store, previewStore, permissions
}

func TestPreviewConfigDeleteRemovesRow(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(0), nil)
	store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(nil)

	if err := svc.Delete(context.Background(), "cfg-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPreviewConfigDeleteBlockedByActivePreviews(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(2), nil)

	if err := svc.Delete(context.Background(), "cfg-1"); err == nil || err.Code != errors.ErrorConflict {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestPreviewConfigDeleteRowDeleteFailureRetrySucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", TeamID: "team-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil).Times(2)
	permissions.EXPECT().Check(gomock.Any(), config.TeamID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil).Times(2)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(0), nil).Times(2)
	gomock.InOrder(
		store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(errors.GeneralError("delete boom")),
		store.EXPECT().Delete(gomock.Any(), "cfg-1").Return(nil),
	)

	if err := svc.Delete(context.Background(), "cfg-1"); err == nil {
		t.Fatal("expected row-delete failure to be returned")
	}
	if err := svc.Delete(context.Background(), "cfg-1"); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
}
