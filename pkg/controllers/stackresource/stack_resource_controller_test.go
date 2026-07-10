package stackresource

import (
	"context"
	"time"

	apperrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	corev1alpha1 "stackdome.io/cluster-agent/api/core/v1alpha1"
)

var _ = Describe("mapClusterStatusToServerStatus", func() {
	It("maps a runtime crash to a container failure", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseFailed,
			StatusHash: "abc123",
			LastFailureDetails: []corev1alpha1.LastFailureDetail{
				{
					ContainerName:           "web",
					RestartCount:            5,
					LastTerminationReason:   "CrashLoopBackOff",
					LastTerminationMessage:  "back-off restarting",
					LastTerminationExitCode: ptr.To(int32(1)),
				},
			},
		}

		got := mapClusterStatusToServerStatus(cr)

		Expect(got.LastFailure).NotTo(BeNil())
		Expect(got.LastFailure.Type).To(Equal(models.FailureTypeRuntimeCrash))
		Expect(got.LastFailure.Container).NotTo(BeNil())
		Expect(got.LastFailure.Container.FailureType).To(Equal("crash_loop"))
		Expect(got.LastFailure.Container.RestartCount).To(Equal(int32(5)))
		Expect(got.LastFailure.InitContainer).To(BeNil())
	})

	It("maps an init container crash to an init-container failure", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Name = "api"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseFailed,
			StatusHash: "def456",
			LastFailureDetails: []corev1alpha1.LastFailureDetail{
				{
					ContainerName:           "api-init",
					RestartCount:            0,
					LastTerminationReason:   "Error",
					LastTerminationExitCode: ptr.To(int32(2)),
				},
			},
		}

		got := mapClusterStatusToServerStatus(cr)

		Expect(got.LastFailure).NotTo(BeNil())
		Expect(got.LastFailure.InitContainer).NotTo(BeNil())
		Expect(got.LastFailure.InitContainer.FailureType).To(Equal("exit_error"))
		Expect(got.LastFailure.Container).To(BeNil())
	})

	It("leaves LastFailure nil for a ready resource", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseReady,
			StatusHash: "ghi789",
		}

		got := mapClusterStatusToServerStatus(cr)

		Expect(got.LastFailure).To(BeNil())
	})

	It("maps LastRestartRequestProcessedAt", func() {
		now := metav1.NewTime(time.Now())
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			LastRestartRequestProcessedAt: &now,
		}

		got := mapClusterStatusToServerStatus(cr)

		Expect(got.LastRestartRequestProcessedAt).NotTo(BeNil())
	})
})

var _ = Describe("computeStatusRewrite", func() {
	It("preserves a build failure when the CR has no failure details", func() {
		buildFailure := &models.StackResourceFailure{Type: models.FailureTypeBuildFailure}
		current := &models.StackResourceStatus{
			LastObservedStatusHash: "old-hash",
			LastFailure:            buildFailure,
		}
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseFailed,
			StatusHash: "new-hash",
		}

		got := computeStatusRewrite(current, cr)

		Expect(got.LastFailure).To(BeIdenticalTo(buildFailure))
		Expect(got.LastObservedStatusHash).To(Equal("new-hash"))
	})

	It("lets CR failure details overwrite a build failure", func() {
		current := &models.StackResourceStatus{
			LastFailure: &models.StackResourceFailure{Type: models.FailureTypeBuildFailure},
		}
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase: corev1alpha1.StackResourcePhaseFailed,
			LastFailureDetails: []corev1alpha1.LastFailureDetail{
				{
					ContainerName:           "web",
					RestartCount:            3,
					LastTerminationReason:   "CrashLoopBackOff",
					LastTerminationExitCode: ptr.To(int32(1)),
				},
			},
		}

		got := computeStatusRewrite(current, cr)

		Expect(got.LastFailure).NotTo(BeNil())
		Expect(got.LastFailure.Type).To(Equal(models.FailureTypeRuntimeCrash))
	})

	It("does not carry a non-build failure over the rewrite", func() {
		current := &models.StackResourceStatus{
			LastFailure: &models.StackResourceFailure{Type: models.FailureTypeRuntimeCrash},
		}
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase: corev1alpha1.StackResourcePhaseReady,
		}

		got := computeStatusRewrite(current, cr)

		Expect(got.LastFailure).To(BeNil())
	})

	It("keeps a cleared build failure cleared", func() {
		current := &models.StackResourceStatus{}
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase: corev1alpha1.StackResourcePhaseReady,
		}

		got := computeStatusRewrite(current, cr)

		Expect(got.LastFailure).To(BeNil())
	})

	It("handles a nil current status", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Name = "web"
		cr.Status = corev1alpha1.StackResourceStatus{
			Phase: corev1alpha1.StackResourcePhaseReady,
		}

		got := computeStatusRewrite(nil, cr)

		Expect(got).NotTo(BeNil())
		Expect(got.LastFailure).To(BeNil())
	})
})

func cond(condType string, status string, reason, message string) models.Condition {
	return models.Condition{Type: condType, Status: status, Reason: reason, Message: message}
}

type resourceEventCase struct {
	conditions  []models.Condition
	failure     *models.StackResourceFailure
	converged   bool
	wantType    models.ReleaseEventType
	wantReason  string
	wantMessage string
	wantEmit    bool
}

var _ = Describe("resourceEvent", func() {
	runtimeCrash := &models.StackResourceFailure{
		Type: models.FailureTypeRuntimeCrash,
		Container: &models.ContainerFailureDetail{
			Reason:  "CrashLoopBackOff",
			Message: "back-off restarting failed container",
		},
	}

	DescribeTable("mapping conditions and failures to release events",
		func(tc resourceEventCase) {
			gotType, gotReason, gotMessage, gotEmit := resourceEvent(tc.conditions, tc.failure, tc.converged)
			Expect(gotEmit).To(Equal(tc.wantEmit))
			if !tc.wantEmit {
				return
			}
			Expect(gotType).To(Equal(tc.wantType))
			Expect(gotReason).To(Equal(tc.wantReason))
			Expect(gotMessage).To(Equal(tc.wantMessage))
		},
		Entry("dependencies not ready records waiting with the condition reason and message", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceDependenciesReady), string(models.ConditionFalse), "DependenciesNotReady", "waiting for mysql"),
			},
			wantType:    models.ReleaseEventTypeResourceWaiting,
			wantReason:  "DependenciesNotReady",
			wantMessage: "waiting for mysql",
			wantEmit:    true,
		}),
		Entry("build in progress suppresses the resource event", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceBuildReady), string(models.ConditionFalse), "BuildNotReady", "application build is not yet ready"),
			},
			wantEmit: false,
		}),
		Entry("dependencies-not-ready wins over build-not-ready", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceBuildReady), string(models.ConditionFalse), "BuildNotReady", "building"),
				cond(string(corev1alpha1.StackResourceDependenciesReady), string(models.ConditionFalse), "DependenciesNotReady", "waiting for redis"),
			},
			wantType:    models.ReleaseEventTypeResourceWaiting,
			wantReason:  "DependenciesNotReady",
			wantMessage: "waiting for redis",
			wantEmit:    true,
		}),
		Entry("stalled records failed with the condition reason and message", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceStalled), string(models.ConditionTrue), "JobFailed", "job has reached the specified backoff limit"),
			},
			wantType:    models.ReleaseEventTypeResourceFailed,
			wantReason:  "JobFailed",
			wantMessage: "job has reached the specified backoff limit",
			wantEmit:    true,
		}),
		Entry("stalled by a terminal build failure is the imagebuild controller's event", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceStalled), string(models.ConditionTrue), stalledReasonBuildFailed, "application build failed terminally"),
			},
			wantEmit: false,
		}),
		Entry("runtime crash records failed with both reason and message", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionFalse), "NotConverged", "rollout not converged"),
			},
			failure:     runtimeCrash,
			wantType:    models.ReleaseEventTypeResourceFailed,
			wantReason:  "CrashLoopBackOff",
			wantMessage: "back-off restarting failed container",
			wantEmit:    true,
		}),
		Entry("carried-over build failure does not masquerade as a runtime failure", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionFalse), "NotConverged", "rollout not converged"),
			},
			failure: &models.StackResourceFailure{
				Type:  models.FailureTypeBuildFailure,
				Build: &models.BuildFailureDetail{Reason: "Error", Message: "kaniko exited"},
			},
			wantType:    models.ReleaseEventTypeResourceDeploying,
			wantReason:  "NotConverged",
			wantMessage: "rollout not converged",
			wantEmit:    true,
		}),
		Entry("available and converged records ready", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceStatusAvailable), string(models.ConditionTrue), "StackResourceAvailable", "StackResource is available"),
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionTrue), "FullyConverged", "all replicas updated"),
			},
			converged: true,
			wantType:  models.ReleaseEventTypeResourceReady,
			wantEmit:  true,
		}),
		Entry("not converged records deploying with the convergence detail", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionFalse), "NotConverged", "rollout not converged: 0/1 updated"),
			},
			wantType:    models.ReleaseEventTypeResourceDeploying,
			wantReason:  "NotConverged",
			wantMessage: "rollout not converged: 0/1 updated",
			wantEmit:    true,
		}),
		Entry("degraded rollout (previous revision serving) records deploying, not ready", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceStatusAvailable), string(models.ConditionTrue), "StackResourceAvailable", "available on previous revision; current rollout not converged"),
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionFalse), "NotConverged", "rollout not converged: 1/2 updated, 1/2 ready, 1 unavailable"),
			},
			wantType:    models.ReleaseEventTypeResourceDeploying,
			wantReason:  "NotConverged",
			wantMessage: "previous revision still serving; rollout not converged: 1/2 updated, 1/2 ready, 1 unavailable",
			wantEmit:    true,
		}),
		Entry("stale Converged=True from a previous generation is not ready", resourceEventCase{
			conditions: []models.Condition{
				cond(string(corev1alpha1.StackResourceStatusAvailable), string(models.ConditionTrue), "StackResourceAvailable", "StackResource is available"),
				cond(string(corev1alpha1.StackResourceConverged), string(models.ConditionTrue), "FullyConverged", "all replicas updated"),
			},
			converged:   false,
			wantType:    models.ReleaseEventTypeResourceDeploying,
			wantReason:  "",
			wantMessage: "previous revision still serving",
			wantEmit:    true,
		}),
		Entry("no conditions emits nothing", resourceEventCase{
			wantEmit: false,
		}),
	)
})

var _ = Describe("convergedForGeneration", func() {
	crWithConverged := func(generation, observedGeneration int64, status metav1.ConditionStatus) *corev1alpha1.StackResource {
		cr := &corev1alpha1.StackResource{}
		cr.Generation = generation
		cr.Status.Conditions = []metav1.Condition{{
			Type:               string(corev1alpha1.StackResourceConverged),
			Status:             status,
			ObservedGeneration: observedGeneration,
			Reason:             "FullyConverged",
		}}
		return cr
	}

	It("is true when Converged=True was written for the current generation", func() {
		Expect(convergedForGeneration(crWithConverged(3, 3, metav1.ConditionTrue))).To(BeTrue())
	})

	It("is false when Converged=True is from a previous generation", func() {
		Expect(convergedForGeneration(crWithConverged(3, 2, metav1.ConditionTrue))).To(BeFalse())
	})

	It("is false when Converged=False", func() {
		Expect(convergedForGeneration(crWithConverged(3, 3, metav1.ConditionFalse))).To(BeFalse())
	})

	It("is false without a Converged condition", func() {
		Expect(convergedForGeneration(&corev1alpha1.StackResource{})).To(BeFalse())
	})
})

func stackResourceTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
	Expect(corev1alpha1.AddToScheme(scheme)).To(Succeed())
	return scheme
}

func readyConditions() []metav1.Condition {
	return []metav1.Condition{
		{Type: string(corev1alpha1.StackResourceStatusAvailable), Status: metav1.ConditionTrue, Reason: "StackResourceAvailable", Message: "StackResource is available"},
		{Type: string(corev1alpha1.StackResourceConverged), Status: metav1.ConditionTrue, Reason: "FullyConverged", Message: "all replicas updated"},
	}
}

func stackResourceCR(hash string, conditions []metav1.Condition) *corev1alpha1.StackResource {
	return &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "ns-1",
			Labels:    map[string]string{corev1alpha1.LabelStackID: "stack-1"},
		},
		Status: corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseReady,
			StatusHash: hash,
			Conditions: conditions,
		},
	}
}

// failedStackResourceCR builds a CR whose LastFailureDetails map to a
// runtime-crash LastFailure, optionally carrying extra status conditions.
func failedStackResourceCR(hash string, conditions []metav1.Condition) *corev1alpha1.StackResource {
	return &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "web",
			Namespace: "ns-1",
			Labels:    map[string]string{corev1alpha1.LabelStackID: "stack-1"},
		},
		Status: corev1alpha1.StackResourceStatus{
			Phase:      corev1alpha1.StackResourcePhaseFailed,
			StatusHash: hash,
			Conditions: conditions,
			LastFailureDetails: []corev1alpha1.LastFailureDetail{
				{
					ContainerName:           "web",
					RestartCount:            3,
					LastTerminationReason:   "CrashLoopBackOff",
					LastTerminationMessage:  "back-off restarting failed container",
					LastTerminationExitCode: ptr.To(int32(1)),
				},
			},
		},
	}
}

func newReconcilerForTest(cr *corev1alpha1.StackResource) (
	*stackResourceReconciler,
	*mocks.MockStackResourceService,
	*MockreleaseActiveChecker,
	*MockresourceEventRecorder,
) {
	mockCtrl := gomock.NewController(GinkgoT())

	svc := mocks.NewMockStackResourceService(mockCtrl)
	checker := NewMockreleaseActiveChecker(mockCtrl)
	recorder := NewMockresourceEventRecorder(mockCtrl)
	r := &stackResourceReconciler{
		client:               fake.NewClientBuilder().WithScheme(stackResourceTestScheme()).WithObjects(cr).Build(),
		stackResourceService: svc,
		releaseChecker:       checker,
		eventRecorder:        recorder,
		logger:               logger.NewLoggerWithPrefix(context.Background(), "sr-test"),
	}
	return r, svc, checker, recorder
}

func reconcileRequest() ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Name: "web", Namespace: "ns-1"}}
}

var _ = Describe("Reconcile", func() {
	// Unchanged StatusHash keeps the controller out of the status-rewrite gate, so
	// nothing downstream — status update, active-release lookup, event record — runs.
	It("records nothing when the StatusHash is unchanged", func() {
		cr := stackResourceCR("h1", readyConditions())
		r, svc, _, _ := newReconcilerForTest(cr)

		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "h1"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		// No UpdateStatus / checker / recorder expectations: gomock fails on any call.

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// No active release means the resource event is silently skipped, and the
	// reconcile still succeeds because the status update already persisted.
	It("skips the event and succeeds when there is no active release", func() {
		cr := stackResourceCR("h2", readyConditions())
		r, svc, checker, _ := newReconcilerForTest(cr)

		db := &models.StackResource{ID: "sr-1", Name: "web", Status: nil}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(nil, nil)
		// recorder must not be called.

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// A recorder failure is non-fatal: the reconcile still returns no error.
	It("does not fail the reconcile when the recorder errors", func() {
		cr := stackResourceCR("h3", readyConditions())
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "", "").
			Return(apperrors.GeneralError("boom"))

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// A crashing resource records resource_failed carrying both the reason code
	// and the long termination message from its LastFailure.
	It("records the LastFailure reason and message for a failed resource", func() {
		cr := failedStackResourceCR("h-fail", nil)
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceFailed, "CrashLoopBackOff", "back-off restarting failed container").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// A Stalled=True condition wins over the LastFailure-derived detail.
	It("lets a stalled condition override the LastFailure detail", func() {
		cr := failedStackResourceCR("h-stalled", []metav1.Condition{
			{
				Type:    string(corev1alpha1.StackResourceStalled),
				Status:  metav1.ConditionTrue,
				Reason:  "InvalidSpec",
				Message: "invalid workload spec",
			},
		})
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceFailed, "InvalidSpec", "invalid workload spec").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// Happy path: a changed StatusHash on an active release records the mapped event.
	It("records the resource event for an active release", func() {
		cr := stackResourceCR("h4", readyConditions())
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "", "").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})
})
