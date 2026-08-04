package stackresource

import (
	"context"
	"time"

	"github.com/Stackdome/stackdome/pkg/controllers"
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
					LastTerminationReason:   controllers.ReasonCrashLoopBackOff,
					LastTerminationMessage:  "back-off restarting",
					LastTerminationExitCode: ptr.To(int32(1)),
				},
			},
		}

		got := mapClusterStatusToServerStatus(cr)

		Expect(got.LastFailure).NotTo(BeNil())
		Expect(got.LastFailure.Type).To(Equal(models.FailureTypeRuntimeCrash))
		Expect(got.LastFailure.Container).NotTo(BeNil())
		Expect(got.LastFailure.Container.FailureType).To(Equal(controllers.FailureTypeCrashLoop))
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
		Expect(got.LastFailure.InitContainer.FailureType).To(Equal(controllers.FailureTypeExitError))
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
					LastTerminationReason:   controllers.ReasonCrashLoopBackOff,
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

func cond(condType string, status metav1.ConditionStatus, reason, message string) metav1.Condition {
	return metav1.Condition{Type: condType, Status: status, Reason: reason, Message: message}
}

// Reasons the agent stamps on its summary and conditions. The agent API only
// exports the two the hub branches on, so the rest are named once here.
const (
	reasonDependenciesNotReady = "DependenciesNotReady"
	reasonBuildNotReady        = "BuildNotReady"
	reasonFullyConverged       = "FullyConverged"
	reasonNotConverged         = "NotConverged"
	reasonInvalidSpec          = "InvalidSpec"
)

func withSummary(cr *corev1alpha1.StackResource, state corev1alpha1.StackResourceSummaryState, reason, message string, observedGeneration int64) *corev1alpha1.StackResource {
	cr.Status.Summary = &corev1alpha1.StackResourceStatusSummary{
		State:              state,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: observedGeneration,
	}
	return cr
}

func summaryCR(state corev1alpha1.StackResourceSummaryState, reason, message string, summaryGen, crGen int64) *corev1alpha1.StackResource {
	cr := &corev1alpha1.StackResource{}
	cr.Generation = crGen
	return withSummary(cr, state, reason, message, summaryGen)
}

type resourceEventCase struct {
	crGen       int64 // defaults to 1 via the helper
	condGen     int64 // ObservedGeneration stamped on all conditions; defaults to crGen
	statusGen   int64 // Status.ObservedGeneration; defaults to crGen
	conditions  []metav1.Condition
	portCheck   *corev1alpha1.PortCheckStatus
	wantType    models.ReleaseEventType
	wantReason  string
	wantMessage string
	wantEmit    bool
}

func eventCaseCR(tc resourceEventCase) *corev1alpha1.StackResource {
	if tc.crGen == 0 {
		tc.crGen = 1
	}
	if tc.condGen == 0 {
		tc.condGen = tc.crGen
	}
	if tc.statusGen == 0 {
		tc.statusGen = tc.crGen
	}
	cr := &corev1alpha1.StackResource{}
	cr.Generation = tc.crGen
	cr.Status.ObservedGeneration = tc.statusGen
	cr.Status.PortCheck = tc.portCheck
	for _, c := range tc.conditions {
		c.ObservedGeneration = tc.condGen
		cr.Status.Conditions = append(cr.Status.Conditions, c)
	}
	return cr
}

var _ = Describe("resourceEvent", func() {
	DescribeTable("mapping the agent summary to a workload event",
		func(state corev1alpha1.StackResourceSummaryState, reason string, wantType models.ReleaseEventType) {
			ev := resourceEvent(summaryCR(state, reason, "detail", 3, 3))
			Expect(ev).NotTo(BeNil())
			Expect(ev.Type).To(Equal(wantType))
			Expect(ev.Reason).To(Equal(reason))
			Expect(ev.Message).To(Equal("detail"))
		},
		Entry("waiting", corev1alpha1.SummaryStateWaiting, reasonDependenciesNotReady, models.ReleaseEventTypeResourceWaiting),
		Entry("deploying", corev1alpha1.SummaryStateDeploying, corev1alpha1.ReasonPortNotListening, models.ReleaseEventTypeResourceDeploying),
		Entry("ready", corev1alpha1.SummaryStateReady, reasonFullyConverged, models.ReleaseEventTypeResourceReady),
		Entry("failed", corev1alpha1.SummaryStateFailed, controllers.ReasonCrashLoopBackOff, models.ReleaseEventTypeResourceFailed),
	)

	It("emits nothing while building — build events narrate that phase", func() {
		Expect(resourceEvent(summaryCR(corev1alpha1.SummaryStateBuilding, reasonBuildNotReady, "building", 3, 3))).To(BeNil())
	})

	It("emits nothing for a build failure — build_failed is already on the timeline", func() {
		Expect(resourceEvent(summaryCR(corev1alpha1.SummaryStateFailed, corev1alpha1.ReasonBuildFailed, "boom", 3, 3))).To(BeNil())
	})

	It("emits nothing for a stale summary", func() {
		Expect(resourceEvent(summaryCR(corev1alpha1.SummaryStateReady, reasonFullyConverged, "", 2, 3))).To(BeNil())
	})

	It("emits nothing without a summary", func() {
		Expect(resourceEvent(&corev1alpha1.StackResource{})).To(BeNil())
	})
})

var _ = Describe("portsClosedEvent", func() {
	convergedAndAvailable := []metav1.Condition{
		cond(string(corev1alpha1.StackResourceConverged), metav1.ConditionTrue, "FullyConverged", "all replicas updated"),
		cond(string(corev1alpha1.StackResourceStatusAvailable), metav1.ConditionTrue, "StackResourceAvailable", "StackResource is available"),
	}
	failingDial := &corev1alpha1.PortCheckStatus{
		Status:             corev1alpha1.PortCheckStatusTypeFailure,
		FailingPortNumbers: []int32{80},
	}

	DescribeTable("mapping a landed rollout's port dial to a warning",
		func(tc resourceEventCase) {
			ev := portsClosedEvent(eventCaseCR(tc))
			Expect(ev != nil).To(Equal(tc.wantEmit))
			if !tc.wantEmit {
				return
			}
			Expect(ev.Type).To(Equal(tc.wantType))
			Expect(ev.Reason).To(Equal(tc.wantReason))
			Expect(ev.Message).To(Equal(tc.wantMessage))
		},
		Entry("a ready resource with a failing dial warns that the port is closed", resourceEventCase{
			crGen:       2,
			conditions:  convergedAndAvailable,
			portCheck:   failingDial,
			wantType:    models.ReleaseEventTypeResourcePortsClosed,
			wantReason:  controllers.ReasonPortNotListening,
			wantMessage: "port 80 not accepting connections",
			wantEmit:    true,
		}),
		Entry("a not-converged resource is left to the deploying branch", resourceEventCase{
			conditions: []metav1.Condition{
				cond(string(corev1alpha1.StackResourceConverged), metav1.ConditionFalse, "NotConverged", "rollout not converged"),
				cond(string(corev1alpha1.StackResourceStatusAvailable), metav1.ConditionTrue, "StackResourceAvailable", "StackResource is available"),
			},
			portCheck: failingDial,
			wantEmit:  false,
		}),
		Entry("an unavailable resource does not warn", resourceEventCase{
			conditions: []metav1.Condition{
				cond(string(corev1alpha1.StackResourceConverged), metav1.ConditionTrue, "FullyConverged", "all replicas updated"),
				cond(string(corev1alpha1.StackResourceStatusAvailable), metav1.ConditionFalse, "StackResourceUnavailable", "no available replicas"),
			},
			portCheck: failingDial,
			wantEmit:  false,
		}),
		Entry("a successful dial does not warn", resourceEventCase{
			conditions: convergedAndAvailable,
			portCheck:  &corev1alpha1.PortCheckStatus{Status: corev1alpha1.PortCheckStatusTypeSuccess},
			wantEmit:   false,
		}),
		Entry("no port check does not warn", resourceEventCase{
			conditions: convergedAndAvailable,
			wantEmit:   false,
		}),
		Entry("a dial from a stale status is ignored", resourceEventCase{
			crGen:      2,
			statusGen:  1,
			condGen:    2,
			conditions: convergedAndAvailable,
			portCheck:  failingDial,
			wantEmit:   false,
		}),
	)

	It("does not warn when only a stale Converged=True backs the failing dial", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Generation = 2
		cr.Status.ObservedGeneration = 2
		cr.Status.PortCheck = failingDial
		cr.Status.Conditions = []metav1.Condition{
			{Type: string(corev1alpha1.StackResourceConverged), Status: metav1.ConditionTrue, Reason: "FullyConverged", Message: "all replicas updated", ObservedGeneration: 1},
			{Type: string(corev1alpha1.StackResourceStatusAvailable), Status: metav1.ConditionTrue, Reason: "StackResourceAvailable", Message: "StackResource is available", ObservedGeneration: 2},
		}
		Expect(portsClosedEvent(cr)).To(BeNil())
	})

	It("does not warn when only a stale Available=True backs the failing dial", func() {
		cr := &corev1alpha1.StackResource{}
		cr.Generation = 2
		cr.Status.ObservedGeneration = 2
		cr.Status.PortCheck = failingDial
		cr.Status.Conditions = []metav1.Condition{
			{Type: string(corev1alpha1.StackResourceStatusAvailable), Status: metav1.ConditionTrue, Reason: "StackResourceAvailable", Message: "StackResource is available", ObservedGeneration: 1},
			{Type: string(corev1alpha1.StackResourceConverged), Status: metav1.ConditionTrue, Reason: "FullyConverged", Message: "all replicas updated", ObservedGeneration: 2},
		}
		Expect(portsClosedEvent(cr)).To(BeNil())
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
		{Type: string(corev1alpha1.StackResourceConverged), Status: metav1.ConditionTrue, Reason: reasonFullyConverged, Message: "all replicas updated"},
	}
}

// readyCR is the converged, available CR the agent summarises as Ready.
func readyCR(hash string, extraConditions ...metav1.Condition) *corev1alpha1.StackResource {
	cr := stackResourceCR(hash, append(readyConditions(), extraConditions...))
	return withSummary(cr, corev1alpha1.SummaryStateReady, "", "", 0)
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
					LastTerminationReason:   controllers.ReasonCrashLoopBackOff,
					LastTerminationMessage:  "back-off restarting failed container",
					LastTerminationExitCode: ptr.To(int32(1)),
				},
			},
		},
	}
}

// generationCR builds a CR pinned to an explicit generation and observed
// generation, so a superseded rollout (a summary or condition whose
// ObservedGeneration lags) can be reproduced.
func generationCR(generation, observedGeneration int64, conditions []metav1.Condition) *corev1alpha1.StackResource {
	cr := &corev1alpha1.StackResource{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "web",
			Namespace:  "ns-1",
			Generation: generation,
			Labels:     map[string]string{corev1alpha1.LabelStackID: "stack-1"},
		},
		Status: corev1alpha1.StackResourceStatus{
			Phase:              corev1alpha1.StackResourcePhaseFailed,
			StatusHash:         "gen-hash",
			ObservedGeneration: observedGeneration,
			Conditions:         conditions,
		},
	}
	return cr
}

func newReconcilerForTest(cr *corev1alpha1.StackResource) (
	*stackResourceReconciler,
	*mocks.MockStackResourceService,
	*MockreleaseResolver,
	*MockresourceEventRecorder,
) {
	mockCtrl := gomock.NewController(GinkgoT())

	svc := mocks.NewMockStackResourceService(mockCtrl)
	checker := NewMockreleaseResolver(mockCtrl)
	recorder := NewMockresourceEventRecorder(mockCtrl)
	r := &stackResourceReconciler{
		client:               fake.NewClientBuilder().WithScheme(stackResourceTestScheme()).WithObjects(cr).Build(),
		stackResourceService: svc,
		releaseResolver:      checker,
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
		cr := readyCR("h1")
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
		cr := readyCR("h2")
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
		cr := readyCR("h3")
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
	// and the long termination message the agent put on its summary.
	It("records the summary reason and message for a failed resource", func() {
		cr := withSummary(failedStackResourceCR("h-fail", nil), corev1alpha1.SummaryStateFailed,
			controllers.ReasonCrashLoopBackOff, "back-off restarting failed container", 0)
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceFailed, controllers.ReasonCrashLoopBackOff, "back-off restarting failed container").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// Bug: a failure verdict left over from a superseded release (its summary
	// ObservedGeneration lags the CR generation) must NOT be re-emitted as
	// resource_failed against the newly active release. The status update still
	// persists, but no active-release lookup or event record happens.
	It("does not record a failure carried over from a previous generation", func() {
		cr := withSummary(generationCR(2, 2, nil), corev1alpha1.SummaryStateFailed, reasonInvalidSpec, "invalid workload spec", 1)
		r, svc, _, _ := newReconcilerForTest(cr)

		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		// checker + recorder must not be called.

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// A failure verdict for the current generation still records resource_failed.
	It("records a failure for the current generation", func() {
		cr := withSummary(generationCR(2, 2, nil), corev1alpha1.SummaryStateFailed, reasonInvalidSpec, "invalid workload spec", 2)
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-2"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceFailed, reasonInvalidSpec, "invalid workload spec").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// Happy path: a changed StatusHash on an active release records the mapped event.
	It("records the resource event for an active release", func() {
		cr := readyCR("h4")
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

	// A landed rollout whose declared port is dead records the ports-closed
	// warning alongside the ready event, against the active release.
	It("records a ports-closed warning for a converged resource with a failing dial", func() {
		cr := readyCR("h-ports")
		cr.Status.PortCheck = &corev1alpha1.PortCheckStatus{
			Status:             corev1alpha1.PortCheckStatusTypeFailure,
			FailingPortNumbers: []int32{80},
		}
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1"}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGetActiveByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "", "").
			Return(nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourcePortsClosed, controllers.ReasonPortNotListening, "port 80 not accepting connections").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// The old CR is still on the cluster while the new release deploys: its
	// annotation resolves to a superseded release that is no longer the latest,
	// so its state is dropped instead of landing on the new release's timeline.
	It("does not record events observed from a CR written for a superseded release", func() {
		cr := readyCR("h-stale-release")
		cr.Annotations = map[string]string{corev1alpha1.ReleaseIDAnnotation: "rel-1"}
		cr.Status.PortCheck = &corev1alpha1.PortCheckStatus{
			Status:             corev1alpha1.PortCheckStatusTypeFailure,
			FailingPortNumbers: []int32{80},
		}
		r, svc, checker, _ := newReconcilerForTest(cr)

		superseded := &models.StackRelease{ID: "rel-1", State: models.ReleaseStateSuperseded}
		latest := &models.StackRelease{ID: "rel-2", State: models.ReleaseStateInProgress}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGet(gomock.Any(), "rel-1").Return(superseded, nil)
		checker.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").Return(latest, nil)
		// recorder must not be called.

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	It("records events from a CR written for the active release", func() {
		cr := readyCR("h-matching-release")
		cr.Annotations = map[string]string{corev1alpha1.ReleaseIDAnnotation: "rel-1"}
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1", State: models.ReleaseStateInProgress}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGet(gomock.Any(), "rel-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "", "").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// A cert can finish issuing, or pods flip state, after the release is done:
	// a terminal release still records as long as it is the latest one.
	It("records events on a terminal release that is still the latest", func() {
		cr := readyCR("h-late-event")
		cr.Annotations = map[string]string{corev1alpha1.ReleaseIDAnnotation: "rel-1"}
		r, svc, checker, recorder := newReconcilerForTest(cr)

		release := &models.StackRelease{ID: "rel-1", State: models.ReleaseStateReleased}
		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGet(gomock.Any(), "rel-1").Return(release, nil)
		checker.EXPECT().InternalGetLatestByStackID(gomock.Any(), "stack-1").Return(release, nil)
		recorder.EXPECT().
			RecordResourceEvent(gomock.Any(), release, "web", models.ReleaseEventTypeResourceReady, "", "").
			Return(nil)

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// The annotated release can be gone entirely (GC'd): skip, don't guess.
	It("skips events when the annotated release no longer exists", func() {
		cr := readyCR("h-gone-release")
		cr.Annotations = map[string]string{corev1alpha1.ReleaseIDAnnotation: "rel-gone"}
		r, svc, checker, _ := newReconcilerForTest(cr)

		db := &models.StackResource{ID: "sr-1", Name: "web", Status: &models.StackResourceStatus{LastObservedStatusHash: "old"}}
		svc.EXPECT().InternalGetByStackIDAndResourceName(gomock.Any(), "stack-1", "web").Return(db, nil)
		svc.EXPECT().UpdateStatus(gomock.Any(), "sr-1", gomock.Any()).Return(nil)
		checker.EXPECT().InternalGet(gomock.Any(), "rel-gone").Return(nil, apperrors.NotFound("release not found"))
		// recorder must not be called.

		_, err := r.Reconcile(context.Background(), reconcileRequest())
		Expect(err).NotTo(HaveOccurred())
	})

	// CRs applied before the annotation existed carry no information, and must
	// not go silent.
	It("records events from a CR carrying no release-id annotation", func() {
		cr := readyCR("h-no-annotation")
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
