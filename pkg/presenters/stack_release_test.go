package presenters_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"k8s.io/utils/ptr"
)

var _ = Describe("PresentStackRelease", func() {
	baseRelease := func() *models.StackRelease {
		return &models.StackRelease{
			ID:      "r1",
			StackID: "s1",
			State:   models.ReleaseStateReleased,
			Cause:   models.ReleaseCause{Kind: models.ReleaseCauseManual},
		}
	}

	It("presents a nil LiveStatus when no live status is given", func() {
		out := presenters.PresentStackRelease(baseRelease(), nil)
		Expect(out.LiveStatus).To(BeNil())
	})

	It("round-trips health, revisions, conditions, and resources from a populated live status", func() {
		restartTime := ptr.To(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		runTime := ptr.To(time.Date(2024, 1, 1, 0, 5, 0, 0, time.UTC))
		live := &models.ReleaseLiveStatus{
			Health:           models.ReleaseHealthOK,
			TargetRevision:   "5",
			ObservedRevision: "5",
			Conditions: []models.Condition{
				{Type: "Available", Status: "True", Reason: "AllReady", Message: "all resources ready"},
			},
			Resources: map[string]*models.StackResourceStatus{
				"web": {
					State:                         models.StackResourcePhaseReady,
					Message:                       "ready",
					ObservedCrRevision:            "5",
					Replicas:                      2,
					AvailableReplicas:             2,
					UpdatedReplicas:               2,
					PublicIngresses:               []models.Ingress{{URL: "https://web.example.com", TargetPort: 8080}},
					InternalServiceName:           ptr.To("web-svc"),
					LastRestartRequestProcessedAt: restartTime,
					LastRunTime:                   runTime,
					LastRunSucceeded:              ptr.To(true),
					Conditions:                    []models.Condition{{Type: "Ready", Status: "True", Reason: "PodsReady", Message: "all pods ready"}},
				},
			},
		}

		out := presenters.PresentStackRelease(baseRelease(), live)

		Expect(out.LiveStatus).NotTo(BeNil())
		Expect(*out.LiveStatus.Health).To(Equal(openapi.RELEASE_HEALTH_OK))
		Expect(*out.LiveStatus.TargetRevision).To(Equal("5"))
		Expect(*out.LiveStatus.ObservedRevision).To(Equal("5"))
		Expect(out.LiveStatus.Conditions).To(HaveLen(1))
		Expect(*out.LiveStatus.Conditions[0].Type).To(Equal("Available"))
		Expect(*out.LiveStatus.Conditions[0].Reason).To(Equal("AllReady"))
		Expect(*out.LiveStatus.Conditions[0].Status).To(Equal("True"))
		Expect(*out.LiveStatus.Conditions[0].Message).To(Equal("all resources ready"))

		Expect(out.LiveStatus.Resources).NotTo(BeNil())
		resources := *out.LiveStatus.Resources
		Expect(resources).To(HaveKey("web"))
		web := resources["web"]
		Expect(*web.State).To(Equal(string(models.StackResourcePhaseReady)))
		Expect(*web.Message).To(Equal("ready"))
		Expect(*web.ObservedRevision).To(Equal("5"))
		Expect(*web.Replicas).To(Equal(int32(2)))
		Expect(*web.AvailableReplicas).To(Equal(int32(2)))
		Expect(*web.UpdatedReplicas).To(Equal(int32(2)))
		Expect(web.PublicIngress).To(HaveLen(1))
		Expect(*web.PublicIngress[0].Url).To(Equal("https://web.example.com"))
		Expect(*web.PublicIngress[0].TargetPort).To(Equal(int32(8080)))
		Expect(*web.InternalServiceName).To(Equal("web-svc"))
		Expect(web.LastRestartRequestProcessedAt).To(Equal(restartTime))
		Expect(web.LastRunTime).To(Equal(runTime))
		Expect(*web.LastRunSucceeded).To(Equal(true))
		Expect(web.Conditions).To(HaveLen(1))
		Expect(*web.Conditions[0].Type).To(Equal("Ready"))
		Expect(*web.Conditions[0].Status).To(Equal("True"))
		Expect(*web.Conditions[0].Reason).To(Equal("PodsReady"))
		Expect(*web.Conditions[0].Message).To(Equal("all pods ready"))
	})

	It("maps a resource's LastFailure with container, init container, and build detail", func() {
		live := &models.ReleaseLiveStatus{
			Health: models.ReleaseHealthDegraded,
			Resources: map[string]*models.StackResourceStatus{
				"web": {
					State: models.StackResourcePhaseFailed,
					LastFailure: &models.StackResourceFailure{
						Type: models.FailureTypeBuildFailure,
						Container: &models.ContainerFailureDetail{
							FailureType:  "OOMKilled",
							Reason:       "OOMKilled",
							Message:      "container was OOM killed",
							RestartCount: 3,
							ExitCode:     ptr.To(int32(137)),
						},
						InitContainer: &models.ContainerFailureDetail{
							FailureType:  "Error",
							Reason:       "Error",
							Message:      "init container failed",
							RestartCount: 1,
							ExitCode:     ptr.To(int32(1)),
						},
						Build: &models.BuildFailureDetail{
							FailureType:  "exit_error",
							Reason:       "Error",
							Message:      "COPY failed: file not found",
							RestartCount: 0,
							ExitCode:     ptr.To(int32(1)),
						},
					},
				},
			},
		}

		out := presenters.PresentStackRelease(baseRelease(), live)

		resources := *out.LiveStatus.Resources
		failure := resources["web"].LastFailure
		Expect(failure).NotTo(BeNil())
		Expect(*failure.Type).To(Equal(string(models.FailureTypeBuildFailure)))

		Expect(failure.Container).NotTo(BeNil())
		Expect(*failure.Container.FailureType).To(Equal("OOMKilled"))
		Expect(*failure.Container.Reason).To(Equal("OOMKilled"))
		Expect(*failure.Container.Message).To(Equal("container was OOM killed"))
		Expect(*failure.Container.RestartCount).To(Equal(int32(3)))
		Expect(*failure.Container.ExitCode).To(Equal(int32(137)))

		Expect(failure.InitContainer).NotTo(BeNil())
		Expect(*failure.InitContainer.FailureType).To(Equal("Error"))
		Expect(*failure.InitContainer.Reason).To(Equal("Error"))
		Expect(*failure.InitContainer.Message).To(Equal("init container failed"))
		Expect(*failure.InitContainer.RestartCount).To(Equal(int32(1)))
		Expect(*failure.InitContainer.ExitCode).To(Equal(int32(1)))

		Expect(failure.Build).NotTo(BeNil())
		Expect(*failure.Build.FailureType).To(Equal("exit_error"))
		Expect(*failure.Build.Reason).To(Equal("Error"))
		Expect(*failure.Build.Message).To(Equal("COPY failed: file not found"))
		Expect(*failure.Build.RestartCount).To(Equal(int32(0)))
		Expect(*failure.Build.ExitCode).To(Equal(int32(1)))
	})

	It("presents a nil LastFailure when the resource status has none", func() {
		live := &models.ReleaseLiveStatus{
			Health: models.ReleaseHealthOK,
			Resources: map[string]*models.StackResourceStatus{
				"web": {
					State:       models.StackResourcePhaseReady,
					LastFailure: nil,
				},
			},
		}

		out := presenters.PresentStackRelease(baseRelease(), live)

		resources := *out.LiveStatus.Resources
		Expect(resources["web"].LastFailure).To(BeNil())
	})
})
