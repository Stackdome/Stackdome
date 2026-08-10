package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Organisation domain deletion", func() {
	var (
		ctrl         *gomock.Controller
		domainStore  *mocks.MockOrganisationDomainStore
		stackDomains *mocks.MockStackDomainsStore
		service      *organisationDomainService
		ctx          context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		domainStore = mocks.NewMockOrganisationDomainStore(ctrl)
		stackDomains = mocks.NewMockStackDomainsStore(ctrl)
		service = &organisationDomainService{
			organisationDomainStore: domainStore,
			stackDomainStore:        stackDomains,
		}
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	type deletePath struct {
		name         string
		delete       func(context.Context, string) *errors.ServiceError
		expectDelete func(context.Context, string) *gomock.Call
	}

	paths := []deletePath{
		{
			name: "Delete",
			delete: func(ctx context.Context, id string) *errors.ServiceError {
				return service.Delete(ctx, id)
			},
			expectDelete: func(ctx context.Context, id string) *gomock.Call {
				return domainStore.EXPECT().Delete(ctx, id)
			},
		},
		{
			name: "InternalDeleteWithTx",
			delete: func(ctx context.Context, id string) *errors.ServiceError {
				return service.InternalDeleteWithTx(ctx, id)
			},
			expectDelete: func(ctx context.Context, id string) *gomock.Call {
				return domainStore.EXPECT().DeleteWithTx(ctx, id)
			},
		},
	}

	for _, path := range paths {
		path := path
		Describe(path.name, func() {
			expectDomainAndStackDomains := func(fqdns ...string) {
				domainStore.EXPECT().Get(ctx, "domain-1").Return(&models.OrganisationDomain{
					ID: "domain-1", OrganisationID: "organisation-1", Domain: "example.com",
				}, nil)
				stackDomainList := make(models.StackDomainList, 0, len(fqdns))
				for _, fqdn := range fqdns {
					stackDomainList = append(stackDomainList, &models.StackDomain{Fqdn: fqdn})
				}
				stackDomains.EXPECT().ListByOrganisationID(ctx, "organisation-1").Return(stackDomainList, nil)
			}

			It("allows deletion when only an unrelated platform hostname exists", func() {
				expectDomainAndStackDomains("app.platform.example")
				path.expectDelete(ctx, "domain-1").Return(nil)

				Expect(path.delete(ctx, "domain-1")).To(BeNil())
			})

			It("allows deletion when only a lookalike hostname exists", func() {
				expectDomainAndStackDomains("api.notexample.com")
				path.expectDelete(ctx, "domain-1").Return(nil)

				Expect(path.delete(ctx, "domain-1")).To(BeNil())
			})

			DescribeTable("blocks deletion when a stack hostname uses the custom domain",
				func(fqdn string) {
					expectDomainAndStackDomains(fqdn)

					err := path.delete(ctx, "domain-1")

					Expect(err).NotTo(BeNil())
					Expect(err.Code).To(Equal(errors.ErrorConflict))
				},
				Entry("exact custom domain", "example.com"),
				Entry("custom domain subdomain", "api.example.com"),
			)
		})
	}
})

var _ = Describe("Organisation custom domains disabled by runtime configuration", func() {
	DescribeTable("rejects user-selected custom domains before persistence",
		func(write func(*organisationDomainService, *models.OrganisationDomain) *errors.ServiceError) {
			service := &organisationDomainService{customDomainsDisabled: true}
			domain := &models.OrganisationDomain{
				OrganisationID: "organisation-1",
				Domain:         "customer.example.com",
			}

			err := write(service, domain)

			Expect(err).NotTo(BeNil())
			Expect(err.Code).To(Equal(errors.ErrorBadRequest))
			Expect(err.Reason).To(Equal(customDomainsDisabledInRuntime))
		},
		Entry("create", func(service *organisationDomainService, domain *models.OrganisationDomain) *errors.ServiceError {
			_, err := service.Create(context.Background(), domain)
			return err
		}),
		Entry("update", func(service *organisationDomainService, domain *models.OrganisationDomain) *errors.ServiceError {
			_, err := service.Update(context.Background(), "domain-1", domain)
			return err
		}),
	)
})
