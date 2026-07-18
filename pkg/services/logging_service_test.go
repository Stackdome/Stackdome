package services

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/services/clusterresource"
)

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

	It("errors when the build has not started (no source revision)", func() {
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(&models.ImageBuild{
			StackResourceName: "api",
			Namespace:         "stack-ns",
			Status:            &models.ImageBuildStatus{BuildSourceRevision: ""},
		}, nil)
		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).To(MatchError(ContainSubstring("has not started")))
	})

	It("resolves the job name and delegates to the cluster logging service", func() {
		build := &models.ImageBuild{
			StackResourceName: "api",
			Namespace:         "stack-ns",
			Status:            &models.ImageBuildStatus{BuildSourceRevision: "abc123"},
		}
		mockBuilds.EXPECT().GetByID(ctx, "build-1").Return(build, nil)

		expectedJob := buildsv1alpha1.BuildJobName("api", "abc123")
		mockCLS.EXPECT().
			GetLogsForBuildPod(ctx, "org-1", "stack-ns", expectedJob, gomock.Any()).
			Return(nil, nil)

		_, err := svc.StreamLogsForBuild(ctx, "org-1", "build-1", &LoggingParams{})
		Expect(err).NotTo(HaveOccurred())
	})
})
