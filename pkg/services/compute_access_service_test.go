package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeAccessService", func() {
	It("builds the default entitlement activation", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockComputeAccessStore(ctrl)
		now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		expiresAt := now.Add(6 * time.Hour)
		expected := &models.ComputeAccess{Entitlement: &models.ComputeEntitlement{ID: "entitlement-1"}}
		store.EXPECT().ActivateWithTx(gomock.Any(), stores.ComputeAccessActivation{
			OrganisationID:    "org-1",
			EntitlementSource: models.ComputeEntitlementSourceTrial,
			StartsAt:          now,
			ExpiresAt:         &expiresAt,
		}).Return(expected, nil)
		svc := NewComputeAccessService(ComputeAccessServiceSpec{
			Store: store, DefaultEntitlementSource: models.ComputeEntitlementSourceTrial,
			DefaultEntitlementDuration: 6 * time.Hour, Now: func() time.Time { return now },
		})

		access, serr := svc.ActivateWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(access).To(BeIdenticalTo(expected))
	})

	It("uses the same clock for transactional and worker admission", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockComputeAccessStore(ctrl)
		now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		expected := &models.ComputeAccess{Lease: &models.SharedComputeLease{ID: "lease-1"}}
		store.EXPECT().RequireWithTx(gomock.Any(), "org-1", now).Return(expected, nil)
		store.EXPECT().AdmitComputeMutationWithTx(gomock.Any(), "org-1", now).Return(expected, nil)
		store.EXPECT().GetActiveByOrganisationID(gomock.Any(), "org-1", now).Return(expected, nil)
		svc := NewComputeAccessService(ComputeAccessServiceSpec{Store: store, Now: func() time.Time { return now }})

		access, serr := svc.RequireWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(access).To(BeIdenticalTo(expected))
		access, serr = svc.AdmitComputeMutationWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(access).To(BeIdenticalTo(expected))
		access, serr = svc.RequireAccess(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(access).To(BeIdenticalTo(expected))
	})

	It("allows organisation deletion only when no shared compute lease exists", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockComputeAccessStore(ctrl)
		svc := NewComputeAccessService(ComputeAccessServiceSpec{Store: store})

		store.EXPECT().HasSharedComputeLease(gomock.Any(), "draft-org").Return(false, nil)
		Expect(svc.EnsureNoLease(context.Background(), "draft-org")).To(BeNil())

		store.EXPECT().HasSharedComputeLease(gomock.Any(), "allocated-org").Return(true, nil)
		serr := svc.EnsureNoLease(context.Background(), "allocated-org")
		Expect(serr).NotTo(BeNil())
		Expect(serr.Reason).To(Equal("cannot delete organisation while its shared compute lease exists"))
	})
})
