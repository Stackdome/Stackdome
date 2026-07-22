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

var _ = Describe("Organisation.Platform", func() {
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

	It("persists the Platform flag on InternalCreate and returns it from InternalGetPlatformOrg", func() {
		created := &models.Organisation{ID: "org-1", Name: "platform", Platform: true}

		store.EXPECT().Create(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, spec *models.Organisation) (*models.Organisation, *errors.ServiceError) {
				Expect(spec.Platform).To(BeTrue())
				return created, nil
			})
		store.EXPECT().Get(gomock.Any(), "org-1").Return(created, nil)

		org, serr := svc.InternalCreate(ctx, &models.Organisation{Name: "platform", Platform: true})
		Expect(serr).To(BeNil())
		Expect(org.Platform).To(BeTrue())

		store.EXPECT().GetPlatformOrg(gomock.Any()).Return(created, nil)

		got, serr := svc.InternalGetPlatformOrg(ctx)
		Expect(serr).To(BeNil())
		Expect(got.ID).To(Equal("org-1"))
		Expect(got.Platform).To(BeTrue())
	})

	It("returns ErrorNotFound from InternalGetPlatformOrg when no org is flagged platform", func() {
		store.EXPECT().GetPlatformOrg(gomock.Any()).Return(nil, errors.NotFound("platform organisation not found"))

		org, serr := svc.InternalGetPlatformOrg(ctx)
		Expect(org).To(BeNil())
		Expect(serr).ToNot(BeNil())
		Expect(serr.Code).To(Equal(errors.ErrorNotFound))
	})
})
