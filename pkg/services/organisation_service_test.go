package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Organisation.Default", func() {
	var (
		ctrl  *gomock.Controller
		store *mocks.MockOrganisationStore
		svc   *organisationService
		ctx   context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
		store = mocks.NewMockOrganisationStore(ctrl)
		svc = &organisationService{
			organisationStore: store,
			logger:            logger.NewLogger(),
		}
	})

	It("persists the Default flag on InternalCreate and returns it from InternalGetDefaultOrg", func() {
		created := &models.Organisation{ID: "org-1", Name: "platform", Default: true}

		store.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
				Expect(spec.Default).To(BeTrue())
				return created, nil
			})
		store.EXPECT().Get(gomock.Any(), "org-1").Return(created, nil)

		org, serr := svc.InternalCreate(ctx, &models.Organisation{Name: "platform", Default: true})
		Expect(serr).To(BeNil())
		Expect(org.Default).To(BeTrue())

		store.EXPECT().GetDefaultOrg(gomock.Any()).Return(created, nil)

		got, serr := svc.InternalGetDefaultOrg(ctx)
		Expect(serr).To(BeNil())
		Expect(got.ID).To(Equal("org-1"))
		Expect(got.Default).To(BeTrue())
	})

	It("returns ErrorNotFound from InternalGetDefaultOrg when no org is flagged default", func() {
		store.EXPECT().GetDefaultOrg(gomock.Any()).Return(nil, errors.NotFound("default organisation not found"))

		org, serr := svc.InternalGetDefaultOrg(ctx)
		Expect(org).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})
})
