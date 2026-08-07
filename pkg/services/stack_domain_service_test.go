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

var _ = Describe("Stack domain allocation", func() {
	var (
		ctrl                *gomock.Controller
		domains             *mocks.MockStackDomainsStore
		organisationDomains *mocks.MockOrganisationDomainStore
		serviceLogger       *mocks.MockLogger
		txCtx               context.Context
		stack               *models.Stack
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		domains = mocks.NewMockStackDomainsStore(ctrl)
		organisationDomains = mocks.NewMockOrganisationDomainStore(ctrl)
		serviceLogger = mocks.NewMockLogger(ctrl)
		txCtx = db.CtxWithTransaction(context.Background(), &gorm.DB{})
		stack = &models.Stack{ID: "stack-1", OrganisationID: "organisation-1"}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	newService := func(platformBaseDomain string) *stackDomainService {
		return &stackDomainService{
			domainsStore:            domains,
			organisationDomainStore: organisationDomains,
			logger:                  serviceLogger,
			platformBaseDomain:      platformBaseDomain,
		}
	}

	expectCreate := func(wantFQDN string, wantPort int) {
		domains.EXPECT().
			GetByFqdn(txCtx, wantFQDN).
			Return(nil, errors.NotFound("stack domain not found"))
		domains.EXPECT().
			CreateWithTx(txCtx, gomock.Any()).
			DoAndReturn(func(_ context.Context, domain *models.StackDomain) (*models.StackDomain, *errors.ServiceError) {
				Expect(domain.Fqdn).To(Equal(wantFQDN))
				Expect(domain.TargetPort).To(Equal(wantPort))
				Expect(domain.OrganisationID).To(Equal(stack.OrganisationID))
				Expect(domain.StackID).To(Equal(stack.ID))
				return domain, nil
			})
	}

	It("reuses the persisted FQDN without looking up an organisation domain", func() {
		existing := models.StackDomainList{{
			ID: "domain-1", Fqdn: "original-format.customer.example", StackResourceID: "resource-1", TargetPort: 8080,
		}}
		domains.EXPECT().
			ListByStackResourceID(txCtx, "resource-1").
			Return(existing, nil)
		resource := &models.StackResource{
			ID: "resource-1", Name: "renamed-resource",
			Ports: models.Ports{{Number: 8080, ExposedToPublic: true, SubdomainPrefix: "changed-prefix"}},
		}

		err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).To(BeNil())
		Expect(resource.Ports[0].ExposedFqdn).To(Equal("original-format.customer.example"))
	})

	It("uses the first custom domain for a missing assignment", func() {
		domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(nil, nil)
		organisationDomains.EXPECT().
			ListByOrganisationID(txCtx, stack.OrganisationID).
			Return([]*models.OrganisationDomain{
				{Domain: "first.customer.example"},
				{Domain: "second.customer.example"},
			}, nil)
		expectCreate("api.first.customer.example", 8080)
		resource := &models.StackResource{
			ID: "resource-1", Name: "web",
			Ports: models.Ports{{Number: 8080, ExposedToPublic: true, SubdomainPrefix: "api"}},
		}

		err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).To(BeNil())
		Expect(resource.Ports[0].ExposedFqdn).To(Equal("api.first.customer.example"))
	})

	It("keeps an existing FQDN immutable while assigning only the missing port", func() {
		existing := models.StackDomainList{{
			ID: "domain-1", Fqdn: "legacy-shape.customer.example", StackResourceID: "resource-1", TargetPort: 8080,
		}}
		domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(existing, nil)
		organisationDomains.EXPECT().
			ListByOrganisationID(txCtx, stack.OrganisationID).
			Return([]*models.OrganisationDomain{{Domain: "selected.customer.example"}}, nil)
		expectCreate("new-api.selected.customer.example", 9090)
		resource := &models.StackResource{
			ID: "resource-1", Name: "renamed-resource",
			Ports: models.Ports{
				{Number: 8080, ExposedToPublic: true, SubdomainPrefix: "changed-existing-prefix"},
				{Number: 9090, ExposedToPublic: true, SubdomainPrefix: "new-api"},
			},
		}

		err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).To(BeNil())
		Expect(resource.Ports[0].ExposedFqdn).To(Equal("legacy-shape.customer.example"))
		Expect(resource.Ports[1].ExposedFqdn).To(Equal("new-api.selected.customer.example"))
	})

	It("falls back to the platform domain and stores the eight-character ID", func() {
		domains.EXPECT().ListByStackResourceID(txCtx, "resource-a").Return(nil, nil)
		organisationDomains.EXPECT().ListByOrganisationID(txCtx, stack.OrganisationID).Return(nil, nil)
		expectCreate("my-app-aa982eec.platform.example", 8080)
		resource := &models.StackResource{
			ID: "resource-a", Name: "My App",
			Ports: models.Ports{{Number: 8080, ExposedToPublic: true}},
		}

		err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).To(BeNil())
		Expect(resource.Ports[0].ExposedFqdn).To(Equal("my-app-aa982eec.platform.example"))
		Expect(resource.Ports[0].GeneratedSubdomainPrefix).To(Equal("aa982eec"))
	})

	It("rejects a missing custom and platform domain", func() {
		domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(nil, nil)
		organisationDomains.EXPECT().ListByOrganisationID(txCtx, stack.OrganisationID).Return(nil, nil)
		resource := &models.StackResource{
			ID: "resource-1", Name: "web",
			Ports: models.Ports{{Number: 8080, ExposedToPublic: true}},
		}

		err := newService("").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(errors.ErrorBadRequest))
		Expect(err.Reason).To(And(ContainSubstring("domain"), ContainSubstring("configured")))
	})

	DescribeTable("preserves custom-domain formatting",
		func(port models.Port, wantFQDN, wantGeneratedID string) {
			domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(nil, nil)
			organisationDomains.EXPECT().
				ListByOrganisationID(txCtx, stack.OrganisationID).
				Return([]*models.OrganisationDomain{{Domain: "customer.example"}}, nil)
			expectCreate(wantFQDN, 8080)
			resource := &models.StackResource{ID: "resource-1", Name: "web", Ports: models.Ports{port}}

			err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

			Expect(err).To(BeNil())
			Expect(resource.Ports[0].ExposedFqdn).To(Equal(wantFQDN))
			Expect(resource.Ports[0].GeneratedSubdomainPrefix).To(Equal(wantGeneratedID))
		},
		Entry("explicit prefix",
			models.Port{Number: 8080, ExposedToPublic: true, SubdomainPrefix: "api"},
			"api.customer.example", "",
		),
		Entry("legacy generated prefix",
			models.Port{Number: 8080, ExposedToPublic: true},
			"ek2dzkf3b56u664x.web.customer.example", "ek2dzkf3b56u664x",
		),
	)

	DescribeTable("deletes private and removed assignments",
		func(ports models.Ports, assertCleared bool) {
			existing := models.StackDomainList{{
				ID: "domain-1", Fqdn: "stale.customer.example", StackResourceID: "resource-1", TargetPort: 8080,
			}}
			domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(existing, nil)
			domains.EXPECT().DeleteWithTx(txCtx, "domain-1").Return(nil)
			resource := &models.StackResource{ID: "resource-1", Name: "web", Ports: ports}

			err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

			Expect(err).To(BeNil())
			if assertCleared {
				Expect(resource.Ports[0].ExposedFqdn).To(BeEmpty())
				Expect(resource.Ports[0].GeneratedSubdomainPrefix).To(BeEmpty())
			}
		},
		Entry("private port", models.Ports{{
			Number: 8080, ExposedFqdn: "stale.customer.example", GeneratedSubdomainPrefix: "stale-prefix",
		}}, true),
		Entry("removed port", models.Ports(nil), false),
	)

	It("propagates a collision without retrying", func() {
		collision := errors.Conflict("domain 'api.customer.example' is already in use")
		domains.EXPECT().ListByStackResourceID(txCtx, "resource-1").Return(nil, nil)
		organisationDomains.EXPECT().
			ListByOrganisationID(txCtx, stack.OrganisationID).
			Return([]*models.OrganisationDomain{{Domain: "customer.example"}}, nil)
		domains.EXPECT().GetByFqdn(txCtx, "api.customer.example").Return(nil, errors.NotFound("stack domain not found"))
		domains.EXPECT().CreateWithTx(txCtx, gomock.Any()).Return(nil, collision)
		serviceLogger.EXPECT().Error(txCtx, "failed to create domain: %v", collision)
		resource := &models.StackResource{
			ID: "resource-1", Name: "web",
			Ports: models.Ports{{Number: 8080, ExposedToPublic: true, SubdomainPrefix: "api"}},
		}

		err := newService("platform.example").PopulateAndSaveExposedPortDomainsForResourceWithTx(txCtx, stack, resource)

		Expect(err).To(BeIdenticalTo(collision))
	})
})
