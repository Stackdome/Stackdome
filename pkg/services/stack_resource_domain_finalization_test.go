package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/db"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

var _ = Describe("Stack resource domain finalization", func() {
	It("persists domain metadata cleared by the allocator for a private-only update", func() {
		ctrl := gomock.NewController(GinkgoT())
		DeferCleanup(ctrl.Finish)
		resourceStore := mocks.NewMockStackResourceStore(ctrl)
		domains := mocks.NewMockStackDomainsService(ctrl)
		service := &stackResourceService{
			stackResourceStore: resourceStore,
			domainNameService:  domains,
		}
		txCtx := db.CtxWithTransaction(context.Background(), &gorm.DB{})
		stack := &models.Stack{ID: "stack-1", UserID: "user-1", Namespace: "namespace-1"}
		input := &models.StackResource{
			Name:  "worker",
			Ports: models.Ports{{Number: 8080}},
		}
		updated := &models.StackResource{
			ID: "resource-1", Name: "worker",
			Ports: models.Ports{{
				Number: 8080, ExposedFqdn: "stale.customer.example", GeneratedSubdomainPrefix: "stale-prefix",
			}},
		}

		resourceStore.EXPECT().
			UpdateWithTx(txCtx, updated.ID, gomock.Any(), stack).
			Return(updated, nil)
		domains.EXPECT().
			PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, updated).
			DoAndReturn(func(_ context.Context, _ *models.Stack, resource *models.StackResource) *errors.ServiceError {
				resource.Ports[0].ExposedFqdn = ""
				resource.Ports[0].GeneratedSubdomainPrefix = ""
				return nil
			})
		resourceStore.EXPECT().
			UpdatePortsWithTx(txCtx, updated.ID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, resource *models.StackResource) *errors.ServiceError {
				Expect(resource.Ports[0].ExposedFqdn).To(BeEmpty())
				Expect(resource.Ports[0].GeneratedSubdomainPrefix).To(BeEmpty())
				return nil
			})

		result, err := service.InternalUpdateWithTx(txCtx, stack, updated.ID, input)

		Expect(err).To(BeNil())
		Expect(result.Ports[0].ExposedFqdn).To(BeEmpty())
		Expect(result.Ports[0].GeneratedSubdomainPrefix).To(BeEmpty())
	})
})
