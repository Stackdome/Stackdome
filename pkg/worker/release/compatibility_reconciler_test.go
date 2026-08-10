package release

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CompatibilityReconciler", func() {
	It("terminally fails a cloud release whose Git volume is not pinned", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockruntimePolicy(ctrl)
		releases := NewMockreleaseService(ctrl)
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeDatabaseOnly)
		release := &models.StackRelease{
			ID: "release-1",
			Snapshot: models.StackSnapshot{Volumes: []*models.Volume{{
				ID: "volume-1", Name: "source",
				VolumeSource: &models.VolumeSource{GitRepoSource: &models.GitRepoSource{
					RepoUrl:  "https://example.com/repo.git",
					Revision: models.GitRepoRevision{Branch: "main"},
				}},
			}}},
		}
		releases.EXPECT().MarkFailed(gomock.Any(), release.ID, gomock.Any(), nil).Return(true, nil)
		reconciler := newCompatibilityReconciler(ReleaseWorkerSpec{RuntimePolicy: policy, ReleaseService: releases})

		result, err := reconciler.Reconcile(context.Background(), release)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(resultStop))
		Expect(release.State).To(Equal(models.ReleaseStateFailed))
	})

	It("leaves eager self-hosted releases compatible", func() {
		ctrl := gomock.NewController(GinkgoT())
		policy := NewMockruntimePolicy(ctrl)
		policy.EXPECT().DraftProvisioningMode().Return(services.ProvisioningModeEager)
		reconciler := newCompatibilityReconciler(ReleaseWorkerSpec{RuntimePolicy: policy})

		result, err := reconciler.Reconcile(context.Background(), &models.StackRelease{})

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(resultNil))
	})
})
