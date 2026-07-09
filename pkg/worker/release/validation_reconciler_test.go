package release

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/clients"
	"github.com/Stackdome/stackdome/pkg/credentials"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/mocks"
	"github.com/Stackdome/stackdome/pkg/models"
	"go.uber.org/mock/gomock"
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
	reconciler      *validationReconciler
}

func newValidationReconcilerFixture(t *testing.T) *validationReconcilerFixture {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := NewMockreleaseService(ctrl)
	resolver := mocks.NewMockCredentialResolver(ctrl)
	registryClients := NewMockregistryClientProvider(ctrl)
	records := mocks.NewMockResourceValidationRecordStore(ctrl)

	return &validationReconcilerFixture{
		ctrl:            ctrl,
		releaseService:  svc,
		resolver:        resolver,
		registryClients: registryClients,
		records:         records,
		reconciler: &validationReconciler{
			releaseService:     svc,
			credentialResolver: resolver,
			registryClients:    registryClients,
			validationRecords:  records,
			logger:             testLogger(),
		},
	}
}

// 1. Image resource, no record: CheckImage true -> record upserted, resultNil (chain continues).
func TestValidationReconciler_ImagePull_NoRecord_Success(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 2. Image resource, record fingerprint matches -> CheckImage NEVER called (mock expects 0 calls), resultNil.
func TestValidationReconciler_ImagePull_FingerprintMatches_Skipped(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 3. Credential changed (fingerprint mismatch) -> CheckImage called again.
func TestValidationReconciler_ImagePull_FingerprintMismatch_Reprobes(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

//  4. CheckImage returns false -> MarkFailedWithValidationErrors called with
//     code VErrImageNotFound, resource name set; result resultStop.
func TestValidationReconciler_ImagePull_ImageNotFound(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 {
				t.Fatalf("expected 1 validation error, got %d", len(verrs))
			}
			if verrs[0].Code != errors.VErrImageNotFound {
				t.Fatalf("expected code %s, got %s", errors.VErrImageNotFound, verrs[0].Code)
			}
			if verrs[0].ResourceName != "web" {
				t.Fatalf("expected resource name 'web', got %s", verrs[0].ResourceName)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

// 5. CheckImage returns clients.ErrAuthFailed with CONFIGURED credentials
// (SourceIntegration) -> code VErrRegistryAuthFailed, resultStop.
func TestValidationReconciler_ImagePull_AuthFailed(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 || verrs[0].Code != errors.VErrRegistryAuthFailed {
				t.Fatalf("expected VErrRegistryAuthFailed, got %+v", verrs)
			}
			if !strings.Contains(verrs[0].Message, "configured credentials") {
				t.Fatalf("expected message to name the configured-credentials case, got %q", verrs[0].Message)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

// 5b. CheckImage returns clients.ErrAuthFailed with NO credentials resolved
// (SourceAnonymous) -> code VErrRegistryCredentialsRequired naming the
// registry host and image ref, resultStop. Recovers main's
// CredentialsRequired vs CredentialsInvalid distinction in the async check.
func TestValidationReconciler_ImagePull_AnonymousAuthFailed_CredentialsRequired(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 || verrs[0].Code != errors.VErrRegistryCredentialsRequired {
				t.Fatalf("expected VErrRegistryCredentialsRequired, got %+v", verrs)
			}
			msg := verrs[0].Message
			for _, want := range []string{"example.com/app:v1", "example.com", "none are configured"} {
				if !strings.Contains(msg, want) {
					t.Fatalf("expected message to contain %q, got %q", want, msg)
				}
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

//  6. CheckImage returns clients.ErrRateLimited -> check skipped with a
//     warning: no fail, no requeue, no success fingerprint recorded; the
//     release proceeds (resultNil).
func TestValidationReconciler_ImagePull_RateLimited_SkipsCheck(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 6b. CheckPushAccess returns clients.ErrRateLimited -> same skip semantics
//
//	as the pull check: no fail, no requeue, no fingerprint, resultNil.
func TestValidationReconciler_PushAccess_RateLimited_SkipsCheck(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 6c. One resource rate limited, another resource genuinely failing -> the
//
//	rate-limited check is skipped but the real failure still fails the
//	release (resultStop with only the real error).
func TestValidationReconciler_RateLimited_DoesNotMaskOtherFailures(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 || verrs[0].Code != errors.VErrImageNotFound || verrs[0].ResourceName != "api" {
				t.Fatalf("expected single VErrImageNotFound for 'api', got %+v", verrs)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

//  7. Build resource with ExternalImageRef: CheckPushAccess error (auth) with
//     CONFIGURED credentials (SourceIntegration) -> code VErrPushAccessDenied,
//     resultStop.
func TestValidationReconciler_PushAccess_Denied(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 || verrs[0].Code != errors.VErrPushAccessDenied {
				t.Fatalf("expected VErrPushAccessDenied, got %+v", verrs)
			}
			if !strings.Contains(verrs[0].Message, "configured credentials") {
				t.Fatalf("expected message to name the configured-credentials case, got %q", verrs[0].Message)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

// 7b. CheckPushAccess auth failure with NO credentials resolved
// (SourceAnonymous) -> code VErrRegistryCredentialsRequired naming the
// registry host and push ref, resultStop.
func TestValidationReconciler_PushAccess_AnonymousAuthFailed_CredentialsRequired(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 1 || verrs[0].Code != errors.VErrRegistryCredentialsRequired {
				t.Fatalf("expected VErrRegistryCredentialsRequired, got %+v", verrs)
			}
			msg := verrs[0].Message
			for _, want := range []string{"example.com/worker:v1", "example.com", "none are configured"} {
				if !strings.Contains(msg, want) {
					t.Fatalf("expected message to contain %q, got %q", want, msg)
				}
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

// 8. Two bad resources -> ONE MarkFailedWithValidationErrors call with BOTH errors.
func TestValidationReconciler_TwoBadResources_OneCallWithBothErrors(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
			if len(verrs) != 2 {
				t.Fatalf("expected 2 validation errors, got %d: %+v", len(verrs), verrs)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

// 9. release.Manifest != nil (rollback) -> reconciler skips everything, resultNil.
func TestValidationReconciler_Rollback_SkipsValidation(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := imageResource("web", "example.com/app:v1")
	release := validationTestRelease(res)
	release.Manifest = &models.ReleaseManifest{}

	// No mock expectations set on resolver/records/registryClients/releaseService:
	// any unexpected call fails via gomock's controller.

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 10. In-cluster registry image ref -> skipped (no probe).
func TestValidationReconciler_ClusterLocalImage_Skipped(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := imageResource("web", "registry.stackdome-system.svc.cluster.local/app:v1")
	release := validationTestRelease(res)

	// No mock expectations: resolver/records/registryClients must not be called.

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

// 11. In-cluster registry push ref -> skipped (no probe).
func TestValidationReconciler_ClusterLocalPushRef_Skipped(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := buildResourceWithPush("worker", "registry.stackdome-system.svc.cluster.local/worker:v1")
	release := validationTestRelease(res)

	// No mock expectations: resolver/records/registryClients must not be called.

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultNil {
		t.Fatalf("expected resultNil, got %+v", result)
	}
}

//  12. CheckPushAccess returns a non-auth, non-rate-limit error -> treated as
//     transient: reconcile returns the error, MarkFailedWithValidationErrors
//     is never called.
func TestValidationReconciler_PushAccess_TransientError_ReturnsErrorNoMarkFailed(t *testing.T) {
	f := newValidationReconcilerFixture(t)
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
	if err == nil {
		t.Fatalf("expected error, got nil (result: %+v)", result)
	}
}

// 13. Pull credential resolver returns a 404 -> field error VErrRegistryCredentialNotFound (existing behavior).
func TestValidationReconciler_PullCredential_ResolveNotFound_FieldError(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := imageResource("web", "example.com/app:v1")
	release := validationTestRelease(res)

	f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
		credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
		Return(nil, errors.NotFound("registry credential not found"))

	f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), "release-1", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, verrs models.ReleaseValidationErrors) (bool, *errors.ServiceError) {
			if len(verrs) != 1 || verrs[0].Code != errors.VErrRegistryCredentialNotFound {
				t.Fatalf("expected VErrRegistryCredentialNotFound, got %+v", verrs)
			}
			return true, nil
		})

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.resultStop {
		t.Fatalf("expected resultStop, got %+v", result)
	}
}

//  14. Pull credential resolver returns a non-404 error -> reconcile returns the
//     error (worker retry); MarkFailedWithValidationErrors never called.
func TestValidationReconciler_PullCredential_ResolveNonNotFoundError_ReturnsError(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := imageResource("web", "example.com/app:v1")
	release := validationTestRelease(res)

	f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/app:v1",
		credentials.RegistryPurposePull, credentials.RegistryAuthSelector{}).
		Return(nil, errors.InternalServerError("credential store unavailable"))

	f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err == nil {
		t.Fatalf("expected error, got nil (result: %+v)", result)
	}
}

//  15. Push credential resolver returns a non-404 error -> reconcile returns the
//     error (worker retry); MarkFailedWithValidationErrors never called.
func TestValidationReconciler_PushCredential_ResolveNonNotFoundError_ReturnsError(t *testing.T) {
	f := newValidationReconcilerFixture(t)
	res := buildResourceWithPush("worker", "example.com/worker:v1")
	release := validationTestRelease(res)

	f.resolver.EXPECT().RegistryCredentials(gomock.Any(), validationTestOrgID, "example.com/worker:v1",
		credentials.RegistryPurposePush, credentials.RegistryAuthSelector{}).
		Return(nil, errors.InternalServerError("credential store unavailable"))

	f.releaseService.EXPECT().MarkFailedWithValidationErrors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	result, err := f.reconciler.Reconcile(context.Background(), release)
	if err == nil {
		t.Fatalf("expected error, got nil (result: %+v)", result)
	}
}
