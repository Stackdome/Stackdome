package release

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	buildsv1alpha1 "stackdome.io/cluster-agent/api/builds/v1alpha1"

	"github.com/Stackdome/stackdome/pkg/models"
)

var _ = Describe("deadlineReconciler", func() {
	var (
		ctrl       *gomock.Controller
		releaseSvc *MockreleaseService
		buildSvc   *MockimageBuildService
		r          *deadlineReconciler
		ctx        context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		releaseSvc = NewMockreleaseService(ctrl)
		buildSvc = NewMockimageBuildService(ctrl)
		r = newDeadlineReconciler(ReleaseWorkerSpec{ReleaseService: releaseSvc, ImageBuildService: buildSvc})
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	// inProgressRelease returns a release created `age` ago. With
	// withBuild=true the snapshot contains one build-from-source resource,
	// so the release expects an ImageBuild before its converge clock starts.
	inProgressRelease := func(age time.Duration, withBuild bool) *models.StackRelease {
		release := &models.StackRelease{
			ID:        "rel-1",
			StackID:   "stack-1",
			State:     models.ReleaseStateInProgress,
			CreatedAt: time.Now().Add(-age),
		}
		if withBuild {
			release.Snapshot.Resources = []*models.StackResource{
				{BuildConfig: &models.BuildConfigSpec{}},
			}
		}
		return release
	}

	withConvergeClock := func(release *models.StackRelease, startedAgo time.Duration) *models.StackRelease {
		startedAt := time.Now().Add(-startedAgo)
		release.WorkerStatus = &models.ReleaseWorkerStatus{ConvergeClockStartedAt: &startedAt}
		return release
	}

	buildWithState := func(releaseID string, state buildsv1alpha1.BuildPhase) *models.ImageBuild {
		return &models.ImageBuild{
			Status: &models.ImageBuildStatus{ReleaseID: releaseID, State: string(state)},
		}
	}

	It("ignores releases that are not InProgress", func() {
		release := inProgressRelease(10*time.Hour, false)
		release.State = models.ReleaseStatePending

		result, err := r.Reconcile(ctx, release)
		Expect(err).To(BeNil())
		Expect(result.resultNil).To(BeTrue())
	})

	It("fails a release past the absolute lifetime cap regardless of builds", func() {
		release := inProgressRelease(releaseLifetimeCap+time.Minute, true)
		releaseSvc.EXPECT().
			MarkFailed(ctx, "rel-1", "release exceeded maximum lifetime of 6 hours", nil).
			Return(true, nil)

		result, err := r.Reconcile(ctx, release)
		Expect(err).To(BeNil())
		Expect(result.resultStop).To(BeTrue())
	})

	Context("stack without builds", func() {
		It("starts the converge clock on the first tick", func() {
			release := inProgressRelease(time.Minute, false)
			releaseSvc.EXPECT().
				SetConvergeClockStartedAt(ctx, "rel-1", gomock.Not(gomock.Nil())).
				Return(nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("leaves a release alone while the converge clock is inside the window", func() {
			release := withConvergeClock(inProgressRelease(time.Hour, false), time.Minute)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("fails a release whose converge clock has expired", func() {
			release := withConvergeClock(inProgressRelease(time.Hour, false), convergenceTimeout+time.Minute)
			releaseSvc.EXPECT().
				MarkFailed(ctx, "rel-1", "release did not converge within 45 minutes", nil).
				Return(true, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultStop).To(BeTrue())
		})
	})

	Context("stack with builds", func() {
		It("does nothing while the build has not appeared yet", func() {
			release := inProgressRelease(time.Hour, true)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").Return(nil, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("treats another release's build as not appeared for this release", func() {
			release := inProgressRelease(time.Hour, true)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-other", buildsv1alpha1.BuildPhasePending)}, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("does nothing while its build is running and the clock never started", func() {
			release := inProgressRelease(time.Hour, true)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-1", buildsv1alpha1.BuildPhasePending)}, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("clears a stale converge clock when a build shows up running", func() {
			release := withConvergeClock(inProgressRelease(time.Hour, true), convergenceTimeout+time.Minute)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-1", buildsv1alpha1.BuildPhasePending)}, nil)
			releaseSvc.EXPECT().SetConvergeClockStartedAt(ctx, "rel-1", gomock.Nil()).Return(nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("treats a build in an unknown state as still running", func() {
			release := inProgressRelease(time.Hour, true)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-1", buildsv1alpha1.BuildPhase("Building"))}, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("starts the converge clock once the build succeeds", func() {
			release := inProgressRelease(time.Hour, true)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-1", buildsv1alpha1.BuildPhaseSuccess)}, nil)
			releaseSvc.EXPECT().
				SetConvergeClockStartedAt(ctx, "rel-1", gomock.Not(gomock.Nil())).
				Return(nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultNil).To(BeTrue())
		})

		It("fails a release whose post-build converge clock has expired", func() {
			release := withConvergeClock(inProgressRelease(2*time.Hour, true), convergenceTimeout+time.Minute)
			buildSvc.EXPECT().ListByStackID(ctx, "stack-1").
				Return([]*models.ImageBuild{buildWithState("rel-1", buildsv1alpha1.BuildPhaseSuccess)}, nil)
			releaseSvc.EXPECT().
				MarkFailed(ctx, "rel-1", "release did not converge within 45 minutes of build completion", nil).
				Return(true, nil)

			result, err := r.Reconcile(ctx, release)
			Expect(err).To(BeNil())
			Expect(result.resultStop).To(BeTrue())
		})
	})
})
