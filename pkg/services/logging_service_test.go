package services

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
)

func startedBuild() *models.ImageBuild {
	return &models.ImageBuild{
		ID:                "build-1",
		StackResourceName: "api",
		Namespace:         "stack-ns",
		Status: &models.ImageBuildStatus{
			BuildSourceRevision: "abc123",
			Conditions: []models.Condition{{
				Type:   string(buildsv1alpha1.BuildJobCreated),
				Status: string(metav1.ConditionTrue),
			}},
		},
	}
}

var _ = Describe("LoggingService.StreamLogsForStackResource", func() {
	var (
		ctrl      *gomock.Controller
		mockRes   *mocks.MockStackResourceService
		mockCLS   *MockClusterLoggingService
		svc       LoggingService
		ctx       context.Context
		crashing  *models.StackResource
		noWorkloa *models.StackResource
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRes = mocks.NewMockStackResourceService(ctrl)
		mockCLS = NewMockClusterLoggingService(ctrl)
		svc = NewLoggingService(LoggingServiceSpec{StackResourceService: mockRes})
		svc.InjectClusterResourceServiceDeps(clusterresource.ClusterResourceServiceDeps{
			ClusterLoggingService: mockCLS,
		})
		ctx = context.Background()

		crashing = &models.StackResource{
			Name: "web",
			Status: &models.StackResourceStatus{
				State:               models.StackResourcePhaseFailed,
				InternalServiceName: ptr.To("web-svc"),
			},
		}
		noWorkloa = &models.StackResource{
			Name:   "web",
			Status: &models.StackResourceStatus{State: models.StackResourcePhasePending},
		}
	})

	AfterEach(func() { ctrl.Finish() })

	It("serves logs for a crash-looping resource", func() {
		mockRes.EXPECT().GetByStackIDAndResourceName(ctx, "stack-1", "web").Return(crashing, nil)
		mockCLS.EXPECT().
			GetLogsForResources(ctx, "org-1", []*models.StackResource{crashing}, gomock.Any()).
			Return(nil, nil)

		_, err := svc.StreamLogsForStackResource(ctx, "org-1", "stack-1", "web", &LoggingParams{})
		Expect(err).To(BeNil())
	})

	It("returns not-found when there is no workload to read", func() {
		mockRes.EXPECT().GetByStackIDAndResourceName(ctx, "stack-1", "web").Return(noWorkloa, nil)
		mockCLS.EXPECT().
			GetLogsForResources(ctx, "org-1", []*models.StackResource{noWorkloa}, gomock.Any()).
			Return(nil, fmt.Errorf("wrap: %w", clusterresource.ErrNoWorkload))

		_, err := svc.StreamLogsForStackResource(ctx, "org-1", "stack-1", "web", &LoggingParams{})
		Expect(err.Code).To(Equal(errors.ErrorNotFound))
	})

	It("keeps a genuine cluster failure a 500", func() {
		mockRes.EXPECT().GetByStackIDAndResourceName(ctx, "stack-1", "web").Return(crashing, nil)
		mockCLS.EXPECT().
			GetLogsForResources(ctx, "org-1", []*models.StackResource{crashing}, gomock.Any()).
			Return(nil, fmt.Errorf("cluster unreachable"))

		_, err := svc.StreamLogsForStackResource(ctx, "org-1", "stack-1", "web", &LoggingParams{})
		Expect(err.Code).To(Equal(errors.ErrorGeneral))
	})
})

var _ = Describe("LoggingService.StreamLogsForBuild", func() {
	var (
		ctrl       *gomock.Controller
		mockBuilds *MockImageBuildService
		mockCLS    *MockClusterLoggingService
		svc        LoggingService
		ctx        context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockBuilds = NewMockImageBuildService(ctrl)
		mockCLS = NewMockClusterLoggingService(ctrl)
		svc = NewLoggingService(LoggingServiceSpec{ImageBuildService: mockBuilds})
		svc.InjectClusterResourceServiceDeps(clusterresource.ClusterResourceServiceDeps{
			ClusterLoggingService: mockCLS,
		})
		ctx = context.Background()
	})

	AfterEach(func() { ctrl.Finish() })

	It("propagates the RBAC/lookup error from GetByID", func() {
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(nil, errors.Forbidden("nope"))
		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).To(HaveOccurred())
	})

	It("errors 409 when the BuildJobCreated condition is absent", func() {
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(&models.ImageBuild{
			StackResourceName: "api",
			Namespace:         "stack-ns",
			Status:            &models.ImageBuildStatus{BuildSourceRevision: "abc123"},
		}, nil)
		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).To(MatchError(ContainSubstring("has not been created")))
	})

	It("streams when the build has no source revision but the job was created", func() {
		build := startedBuild()
		build.Status.BuildSourceRevision = ""
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(build, nil)
		mockCLS.EXPECT().
			GetLogsForBuildPod(ctx, "org-1", "stack-ns", "build-1", gomock.Any()).
			Return(nil, nil)
		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("maps ErrBuildPodNotReady to a conflict", func() {
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(startedBuild(), nil)
		mockCLS.EXPECT().
			GetLogsForBuildPod(ctx, "org-1", "stack-ns", gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("wrap: %w", clusterresource.ErrBuildPodNotReady))
		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).To(MatchError(ContainSubstring("logs are not available yet")))
	})

	It("delegates to the cluster logging service with the build CR name", func() {
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(startedBuild(), nil)

		// The CR name is the build ID, not StackResourceName and not the never-populated Name.
		mockCLS.EXPECT().
			GetLogsForBuildPod(ctx, "org-1", "stack-ns", "build-1", gomock.Any()).
			Return(nil, nil)

		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).NotTo(HaveOccurred())
	})
})
