package services

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CloudTrialService", func() {
	It("passes configured capacity and TTL to the ambient transaction store", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockTrialAllocationStore(ctrl)
		now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		expected := &models.TrialAllocation{ID: "allocation-1", OrganisationID: "org-1"}
		store.EXPECT().AcquireWithTx(gomock.Any(), "org-1", now, now.Add(6*time.Hour), 10).Return(expected, nil)
		svc := NewCloudTrialService(CloudTrialServiceSpec{
			Store: store, Capacity: 10, TTL: 6 * time.Hour, Now: func() time.Time { return now },
		})

		allocation, serr := svc.AcquireWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(allocation).To(BeIdenticalTo(expected))
	})

	It("uses the same clock for rollback and worker revalidation", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockTrialAllocationStore(ctrl)
		now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
		expected := &models.TrialAllocation{ID: "allocation-1", OrganisationID: "org-1"}
		store.EXPECT().RevalidateWithTx(gomock.Any(), "org-1", now).Return(expected, nil)
		store.EXPECT().GetActiveByOrganisationID(gomock.Any(), "org-1", now).Return(expected, nil)
		svc := NewCloudTrialService(CloudTrialServiceSpec{
			Store: store, Capacity: 10, TTL: 6 * time.Hour, Now: func() time.Time { return now },
		})

		allocation, serr := svc.RevalidateWithTx(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(allocation).To(BeIdenticalTo(expected))
		allocation, serr = svc.RequireActive(context.Background(), "org-1")
		Expect(serr).To(BeNil())
		Expect(allocation).To(BeIdenticalTo(expected))
	})

	It("allows organisation deletion only when no allocation row exists", func() {
		ctrl := gomock.NewController(GinkgoT())
		store := mocks.NewMockTrialAllocationStore(ctrl)
		svc := NewCloudTrialService(CloudTrialServiceSpec{Store: store})

		store.EXPECT().GetByOrganisationID(gomock.Any(), "draft-org").Return(nil, errors.NotFound("trial allocation not found"))
		Expect(svc.EnsureNoAllocation(context.Background(), "draft-org")).To(BeNil())

		store.EXPECT().GetByOrganisationID(gomock.Any(), "allocated-org").Return(&models.TrialAllocation{ID: "allocation-1"}, nil)
		serr := svc.EnsureNoAllocation(context.Background(), "allocated-org")
		Expect(serr).NotTo(BeNil())
		Expect(serr.Reason).To(Equal("cannot delete organisation while its cloud trial allocation exists"))
	})
})
