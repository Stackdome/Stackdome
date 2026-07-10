package release

import (
	"context"
	stderrors "errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
)

const (
	validationTestOrgID   = "org-1"
	validationTestStackID = "stack-1"
)

func validationTestRelease(resources ...*models.StackResource) *models.StackRelease {
	return &models.StackRelease{
		ID:      "release-1",
		StackID: validationTestStackID,
		State:   models.ReleaseStateInProgress,
		Snapshot: models.StackSnapshot{
			Stack: models.StackShellSnapshot{
				ID:             validationTestStackID,
				OrganisationID: validationTestOrgID,
			},
			Resources: resources,
		},
	}
}

func imageResource(name, image string) *models.StackResource {
	return &models.StackResource{
		Name:        name,
		ImageConfig: &models.ImageConfigSpec{Image: image},
	}
}

func buildResourceWithPush(name, externalRef string) *models.StackResource {
	return &models.StackResource{
		Name: name,
		BuildConfig: &models.BuildConfigSpec{
			BuildImageRepository: models.BuildImageRepository{
				ExternalImageRef: externalRef,
			},
		},
	}
}

type validationReconcilerFixture struct {
	ctrl            *gomock.Controller
	releaseService  *MockreleaseService
	resolver        *mocks.MockCredentialResolver
	registryClients *MockregistryClientProvider
	records         *mocks.MockResourceValidationRecordStore
	events          *MockeventRecorder
	reconciler      *validationReconciler
}

// newValidationReconcilerFixtureBare wires all dependencies but sets NO event
// recorder expectations, so tests that assert exact check-event calls have a
// clean slate.
func newValidationReconcilerFixtureBare() *validationReconcilerFixture {
	ctrl := gomock.NewController(GinkgoT())
	svc := NewMockreleaseService(ctrl)
	resolver := mocks.NewMockCredentialResolver(ctrl)
	registryClients := NewMockregistryClientProvider(ctrl)
	records := mocks.NewMockResourceValidationRecordStore(ctrl)
	events := NewMockeventRecorder(ctrl)

	return &validationReconcilerFixture{
		ctrl:            ctrl,
		releaseService:  svc,
		resolver:        resolver,
		registryClients: registryClients,
		records:         records,
		events:          events,
		reconciler: &validationReconciler{
			releaseService:     svc,
			credentialResolver: resolver,
			registryClients:    registryClients,
			validationRecords:  records,
			eventRecorder:      events,
			logger:             testLogger(),
		},
	}
}

// newValidationReconcilerFixture is the default fixture for tests that focus on
// validation behavior rather than events: it permits (but does not assert) any
// check-event recorder calls so those tests need no event bookkeeping.
func newValidationReconcilerFixture() *validationReconcilerFixture {
	f := newValidationReconcilerFixtureBare()
	f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	f.events.EXPECT().RecordReleaseChecksPassed(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	return f
}

var _ = Describe("ValidationReconciler", func() {
	var f *validationReconcilerFixture

	BeforeEach(func() {
		f = newValidationReconcilerFixture()
	})

	// 1. Image resource, no record: CheckImage true -> record upserted, resultNil (chain continues).
	It("probes and upserts a record for an image resource with no prior record", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)

		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(true, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 2. Image resource, record fingerprint matches -> CheckImage NEVER called (mock expects 0 calls), resultNil.
	It("skips the probe when the record fingerprint matches", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)

		fp := checkFingerprint("example.com/app:v1", "", "hash-1")
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(&models.ResourceValidationRecord{Fingerprint: fp}, nil)

		// CheckImage must never be called.
		f.registryClients.EXPECT().ClientFor(gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 3. Credential changed (fingerprint mismatch) -> CheckImage called again.
	It("re-probes when the credential fingerprint changed", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-2"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)

		staleFP := checkFingerprint("example.com/app:v1", "", "hash-1")
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(&models.ResourceValidationRecord{Fingerprint: staleFP}, nil)

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(true, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	//  4. CheckImage returns false -> MarkFailedWithValidationErrors called with
	//     code VErrImageNotFound, resource name set; result resultStop.
	It("fails the release with VErrImageNotFound when the image does not exist", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrImageNotFound))
				Expect(verrs[0].ResourceName).To(Equal("web"))
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// 5. CheckImage returns clients.ErrAuthFailed with CONFIGURED credentials
	// (SourceIntegration) -> code VErrRegistryAuthFailed, resultStop.
	It("fails with VErrRegistryAuthFailed when configured pull credentials are rejected", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceIntegration, DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, clients.ErrAuthFailed)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrRegistryAuthFailed))
				Expect(verrs[0].Message).To(ContainSubstring("configured credentials"),
					"expected message to name the configured-credentials case")
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// 5b. CheckImage returns clients.ErrAuthFailed with NO credentials resolved
	// (SourceAnonymous) -> code VErrRegistryCredentialsRequired naming the
	// registry host and image ref, resultStop. Recovers main's
	// CredentialsRequired vs CredentialsInvalid distinction in the async check.
	It("fails with VErrRegistryCredentialsRequired when anonymous pull auth is rejected", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, clients.ErrAuthFailed)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrRegistryCredentialsRequired))
				for _, want := range []string{"example.com/app:v1", "example.com", "none are configured"} {
					Expect(verrs[0].Message).To(ContainSubstring(want))
				}
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	//  6. CheckImage returns clients.ErrRateLimited -> check skipped with a
	//     warning: no fail, no requeue, no success fingerprint recorded; the
	//     release proceeds (resultNil).
	It("skips the pull check on rate limit without failing or recording", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, clients.ErrRateLimited)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		// No fail call, and no success fingerprint recorded (nothing was verified).
		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 6b. CheckPushAccess returns clients.ErrRateLimited -> same skip semantics
	//
	//	as the pull check: no fail, no requeue, no fingerprint, resultNil.
	It("skips the push check on rate limit without failing or recording", func() {
		res := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "worker", models.ValidationCheckPushAccess).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckPushAccess(gomock.Any(), "example.com/worker:v1").Return(clients.ErrRateLimited)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 6c. One resource rate limited, another resource genuinely failing -> the
	//
	//	rate-limited check is skipped but the real failure still fails the
	//	release (resultStop with only the real error).
	It("does not let a rate-limited check mask another genuine failure", func() {
		res1 := imageResource("web", "example.com/app:v1")
		res2 := imageResource("api", "example.com/api:v1")
		release := validationTestRelease(res1, res2)

		resolved1 := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved1, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client1 := mocks.NewMockRegistryClient(f.ctrl)
		client1.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, clients.ErrRateLimited)
		f.registryClients.EXPECT().ClientFor(resolved1).Return(client1, nil)

		resolved2 := &credentials.ResolvedRegistryCredential{DataHash: "hash-2"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/api:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolved2, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "api", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client2 := mocks.NewMockRegistryClient(f.ctrl)
		client2.EXPECT().CheckImage(gomock.Any(), "example.com/api:v1").Return(false, nil)
		f.registryClients.EXPECT().ClientFor(resolved2).Return(client2, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1), "expected single VErrImageNotFound for 'api', got %+v", verrs)
				Expect(verrs[0].Code).To(Equal(errors.VErrImageNotFound))
				Expect(verrs[0].ResourceName).To(Equal("api"))
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	//  7. Build resource with ExternalImageRef: CheckPushAccess error (auth) with
	//     CONFIGURED credentials (SourceIntegration) -> code VErrPushAccessDenied,
	//     resultStop.
	It("fails with VErrPushAccessDenied when configured push credentials are rejected", func() {
		res := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceIntegration, DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "worker", models.ValidationCheckPushAccess).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckPushAccess(gomock.Any(), "example.com/worker:v1").Return(clients.ErrAuthFailed)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrPushAccessDenied))
				Expect(verrs[0].Message).To(ContainSubstring("configured credentials"),
					"expected message to name the configured-credentials case")
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// 7b. CheckPushAccess auth failure with NO credentials resolved
	// (SourceAnonymous) -> code VErrRegistryCredentialsRequired naming the
	// registry host and push ref, resultStop.
	It("fails with VErrRegistryCredentialsRequired when anonymous push auth is rejected", func() {
		res := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{Source: credentials.SourceAnonymous}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "worker", models.ValidationCheckPushAccess).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckPushAccess(gomock.Any(), "example.com/worker:v1").Return(clients.ErrAuthFailed)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrRegistryCredentialsRequired))
				for _, want := range []string{"example.com/worker:v1", "example.com", "none are configured"} {
					Expect(verrs[0].Message).To(ContainSubstring(want))
				}
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// 8. Two bad resources -> ONE MarkFailedWithValidationErrors call with BOTH errors.
	It("reports both bad resources in a single MarkFailedWithValidationErrors call", func() {
		res1 := imageResource("web", "example.com/app:v1")
		res2 := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res1, res2)

		resolvedPull := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(resolvedPull, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		pullClient := mocks.NewMockRegistryClient(f.ctrl)
		pullClient.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, nil)
		f.registryClients.EXPECT().ClientFor(resolvedPull).Return(pullClient, nil)

		resolvedPush := &credentials.ResolvedRegistryCredential{Source: credentials.SourceIntegration, DataHash: "hash-2"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(resolvedPush, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "worker", models.ValidationCheckPushAccess).
			Return(nil, errors.NotFound("no record"))
		pushClient := mocks.NewMockRegistryClient(f.ctrl)
		pushClient.EXPECT().CheckPushAccess(gomock.Any(), "example.com/worker:v1").Return(clients.ErrAuthFailed)
		f.registryClients.EXPECT().ClientFor(resolvedPush).Return(pushClient, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			Times(1).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(2), "expected 2 validation errors, got %+v", verrs)
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// 9. release.Manifest != nil (rollback) -> reconciler skips everything, resultNil.
	It("skips validation entirely for a rollback release", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)
		release.Manifest = &models.ReleaseManifest{}

		// No mock expectations set on resolver/records/registryClients/releaseService:
		// any unexpected call fails via gomock's controller.

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 10. In-cluster registry image ref -> skipped (no probe).
	It("skips in-cluster registry image refs", func() {
		res := imageResource("web", "registry.stackdome-system.svc.cluster.local/app:v1")
		release := validationTestRelease(res)

		// No mock expectations: resolver/records/registryClients must not be called.

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// 11. In-cluster registry push ref -> skipped (no probe).
	It("skips in-cluster registry push refs", func() {
		res := buildResourceWithPush("worker", "registry.stackdome-system.svc.cluster.local/worker:v1")
		release := validationTestRelease(res)

		// No mock expectations: resolver/records/registryClients must not be called.

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	//  12. CheckPushAccess returns a non-auth, non-rate-limit error -> treated as
	//     transient: reconcile returns the error, MarkFailedWithValidationErrors
	//     is never called.
	It("returns transient push errors without marking the release failed", func() {
		res := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "worker", models.ValidationCheckPushAccess).
			Return(nil, errors.NotFound("no record"))

		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckPushAccess(gomock.Any(), "example.com/worker:v1").
			Return(stderrors.New("connection reset by peer"))
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).To(HaveOccurred(), "expected error, got nil (result: %+v)", result)
	})

	// 13. Pull credential resolver returns a 404 -> field error VErrRegistryCredentialNotFound (existing behavior).
	It("fails with VErrRegistryCredentialNotFound when the pull credential resolves to a 404", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(nil, errors.NotFound("registry credential not found"))

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
				Expect(verrs).To(HaveLen(1))
				Expect(verrs[0].Code).To(Equal(errors.VErrRegistryCredentialNotFound))
				return true, nil
			})

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	//  14. Pull credential resolver returns a non-404 error -> reconcile returns the
	//     error (worker retry); MarkFailedWithValidationErrors never called.
	It("returns a non-404 pull credential resolver error for worker retry", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
			Return(nil, errors.InternalServerError("credential store unavailable"))

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).To(HaveOccurred(), "expected error, got nil (result: %+v)", result)
	})

	//  15. Push credential resolver returns a non-404 error -> reconcile returns the
	//     error (worker retry); MarkFailedWithValidationErrors never called.
	It("returns a non-404 push credential resolver error for worker retry", func() {
		res := buildResourceWithPush("worker", "example.com/worker:v1")
		release := validationTestRelease(res)

		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
			credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
			Return(nil, errors.InternalServerError("credential store unavailable"))

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).To(HaveOccurred(), "expected error, got nil (result: %+v)", result)
	})
})

// --- Check-event production (Task 8) ---
var _ = Describe("ValidationReconciler check events", func() {
	var f *validationReconcilerFixture

	BeforeEach(func() {
		f = newValidationReconcilerFixtureBare()
	})

	// E1. Fresh release, all checks pass -> checks_started and checks_passed each
	// recorded once; reconciler returns resultNil.
	It("records checks_started and checks_passed once when all checks pass", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(true, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)
		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)

		f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), release).Return(nil).Times(1)
		f.events.EXPECT().RecordReleaseChecksPassed(gomock.Any(), release).Return(nil).Times(1)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// E3. Validation error but the MarkFailedWithValidationErrors CAS is lost (a
	// concurrent reconcile already moved the release) -> checks_started is still
	// recorded and the reconciler still stops.
	It("stops on a validation failure even when the CAS is lost", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), release).Return(nil).Times(1)

		f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
			Return(false, nil)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultStop).To(BeTrue(), "expected resultStop, got %+v", result)
	})

	// E4. Rollback (release.Manifest != nil) -> zero recorder calls of any kind.
	It("records nothing for a rollback release", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)
		release.Manifest = &models.ReleaseManifest{}

		f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), gomock.Any()).Times(0)
		f.events.EXPECT().RecordReleaseChecksPassed(gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// E5. Rate-limited check -> checks_started recorded, checks_passed NOT recorded
	// (the skipped check was never verified), release still proceeds (resultNil).
	It("records checks_started but not checks_passed when rate limited", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(false, clients.ErrRateLimited)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)

		f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), release).Return(nil).Times(1)
		f.events.EXPECT().RecordReleaseChecksPassed(gomock.Any(), gomock.Any()).Times(0)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil, got %+v", result)
	})

	// E6. A recorder error is log-only: checks_passed returning an error does not
	// change the reconciler's (successful) outcome.
	It("treats a checks_passed recorder error as log-only", func() {
		res := imageResource("web", "example.com/app:v1")
		release := validationTestRelease(res)

		resolved := &credentials.ResolvedRegistryCredential{DataHash: "hash-1"}
		f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
			credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).Return(resolved, nil)
		f.records.EXPECT().Get(gomock.Any(), validationTestStackID, "web", models.ValidationCheckImagePull).
			Return(nil, errors.NotFound("no record"))
		client := mocks.NewMockRegistryClient(f.ctrl)
		client.EXPECT().CheckImage(gomock.Any(), "example.com/app:v1").Return(true, nil)
		f.registryClients.EXPECT().ClientFor(resolved).Return(client, nil)
		f.records.EXPECT().Upsert(gomock.Any(), gomock.Any()).Return(nil)

		f.events.EXPECT().RecordReleaseChecksStarted(gomock.Any(), release).Return(nil).Times(1)
		f.events.EXPECT().RecordReleaseChecksPassed(gomock.Any(), release).
			Return(errors.InternalServerError("recorder unavailable")).Times(1)

		result, err := f.reconciler.Reconcile(context.Background(), release)
		Expect(err).ToNot(HaveOccurred())
		Expect(result.resultNil).To(BeTrue(), "expected resultNil despite recorder error, got %+v", result)
	})
})
