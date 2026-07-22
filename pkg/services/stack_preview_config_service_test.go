package services

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("StackPreviewConfigService repo uniqueness", func() {
	var (
		ctrl        *gomock.Controller
		svc         *stackPreviewConfigService
		store       *mocks.MockStackPreviewConfigStore
		permissions *mocks.MockPermissionService
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		store = mocks.NewMockStackPreviewConfigStore(ctrl)
		permissions = mocks.NewMockPermissionService(ctrl)
		svc = &stackPreviewConfigService{
			store:             store,
			previewStackStore: mocks.NewMockPreviewStackStore(ctrl),
			permissions:       permissions,
		}
	})

	It("rejects Create when a config already owns the repo", func() {
		permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourcePreviewConfigs, "", auth.ActionCreate).Return(nil)
		cfg := &models.StackPreviewConfig{
			ProjectID:      "project-1",
			OrganisationID: "org-1",
			Name:           "web",
			GitRepository:  models.PreviewGitRepository{RepoURL: "https://github.com/acme/app.git", BaseBranch: "main"},
		}
		store.EXPECT().GetByOrgAndRepo(gomock.Any(), "org-1", cfg.NormalizedRepoURL()).
			Return(&models.StackPreviewConfig{ID: "other"}, nil)

		_, serr := svc.Create(context.Background(), cfg)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorConflict))
	})

	It("rejects Update when the new repo is owned by a different config", func() {
		existing := &models.StackPreviewConfig{
			ID: "cfg-1", ProjectID: "project-1", OrganisationID: "org-1", Name: "web",
			GitRepository: models.PreviewGitRepository{RepoURL: "https://github.com/acme/old.git", BaseBranch: "main"},
		}
		store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(existing, nil)
		permissions.EXPECT().Check(gomock.Any(), "project-1", auth.ResourcePreviewConfigs, "cfg-1", auth.ActionWrite).Return(nil)
		updated := &models.StackPreviewConfig{
			GitRepository: models.PreviewGitRepository{RepoURL: "https://github.com/acme/new.git", BaseBranch: "main"},
		}
		store.EXPECT().GetByOrgAndRepo(gomock.Any(), "org-1", models.NormalizeRepoURL("https://github.com/acme/new.git")).
			Return(&models.StackPreviewConfig{ID: "cfg-2"}, nil)

		_, serr := svc.Update(context.Background(), "cfg-1", updated)
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorConflict))
	})
})

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
	config := &models.StackPreviewConfig{ID: "cfg-1", ProjectID: "project-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
	permissions.EXPECT().Check(gomock.Any(), config.ProjectID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil)
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
	config := &models.StackPreviewConfig{ID: "cfg-1", ProjectID: "project-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil)
	permissions.EXPECT().Check(gomock.Any(), config.ProjectID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil)
	previewStore.EXPECT().CountActiveByConfigID(gomock.Any(), "cfg-1").Return(int64(2), nil)

	if err := svc.Delete(context.Background(), "cfg-1"); err == nil || err.Code != errors.ErrorConflict {
		t.Fatalf("expected conflict error, got %v", err)
	}
}

func TestPreviewConfigDeleteRowDeleteFailureRetrySucceeds(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	svc, store, previewStore, permissions := newPreviewConfigDeleteService(ctrl)
	config := &models.StackPreviewConfig{ID: "cfg-1", ProjectID: "project-1"}

	store.EXPECT().GetByID(gomock.Any(), "cfg-1").Return(config, nil).Times(2)
	permissions.EXPECT().Check(gomock.Any(), config.ProjectID, auth.ResourcePreviewConfigs, "cfg-1", auth.ActionDelete).Return(nil).Times(2)
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
