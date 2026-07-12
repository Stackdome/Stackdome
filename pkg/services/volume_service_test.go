package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

// Suite bootstrapped by TestAESEncryptionService in encryption_service_test.go.

var _ = Describe("VolumeService InternalListNotReady", func() {
	It("returns the not-ready volumes from the store", func() {
		ctrl := gomock.NewController(GinkgoT())
		volumeStore := mocks.NewMockVolumeStore(ctrl)
		volumeStore.EXPECT().InternalListNotReady(gomock.Any()).
			Return([]*models.Volume{{ID: "vol-1"}}, nil)

		svc := &volumeService{volumeStore: volumeStore}

		res, err := svc.InternalListNotReady(context.Background())
		Expect(err).To(BeNil())
		Expect(res).To(HaveLen(1))
		Expect(res[0].ID).To(Equal("vol-1"))
	})
})
