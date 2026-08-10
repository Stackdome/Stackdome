package services

import (
	"context"

	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StackResource registry preparation", func() {
	newResource := func() *models.StackResource {
		return &models.StackResource{BuildConfig: &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{
				UseInClusterRegistry: true,
				ClusterRegistryName:  "stale-registry",
			},
		}}
	}

	It("keeps cloud drafts database-only without resolving or persisting a registry", func() {
		ctrl := gomock.NewController(GinkgoT())
		registry := mocks.NewMockImageRegistryService(ctrl)
		svc := &stackResourceService{clusterRegistryService: registry, runtimePolicy: newCloudRuntimePolicyForTest()}
		resource := newResource()

		serr := svc.prepareResource(context.Background(), &models.Stack{ID: "stack-1", OrganisationID: "org-1", ClusterID: "cluster-1"}, resource)

		Expect(serr).To(BeNil())
		Expect(resource.BuildConfig.BuildImageRepository.ClusterRegistryName).To(BeEmpty())
	})

	It("keeps eager self-hosted registry resolution unchanged", func() {
		ctrl := gomock.NewController(GinkgoT())
		registry := mocks.NewMockImageRegistryService(ctrl)
		svc := &stackResourceService{clusterRegistryService: registry, runtimePolicy: NewSelfHostedRuntimePolicy()}
		resource := newResource()
		registry.EXPECT().PopulateInClusterRegistryNameForResource(gomock.Any(), "org-1", "cluster-1", "demo", resource).Return(nil)

		serr := svc.prepareResource(context.Background(), &models.Stack{ID: "stack-1", Name: "demo", OrganisationID: "org-1", ClusterID: "cluster-1"}, resource)

		Expect(serr).To(BeNil())
	})
})
