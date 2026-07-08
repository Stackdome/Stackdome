package int

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

// Fixture values for the validation specs below. Kept local to this file
// since nothing else in the suite references them yet.
const (
	// validationUnpullableImageRef is a syntactically valid image reference
	// that cannot be pulled. Fat stack create must not probe it.
	validationUnpullableImageRef = "docker.io/stackdome-e2e/definitely-not-a-real-image:v1"

	// validationMissingImageRef 404s anonymously on GHCR without hitting
	// Docker Hub's anonymous-pull rate limit.
	validationMissingImageRef = "ghcr.io/stackdome/e2e-missing:v1"

	// validationDefaultBranchRepoURL is a small public repo already used by
	// other build-from-source fixtures in this suite (see samples/build_from_source.json).
	validationDefaultBranchRepoURL = "https://github.com/Stackdome/test-repo.git"

	validationBadBranchName = "does-not-exist-xyz"

	// validationSkipProbeResourceName is the fixed resource name CreateSimpleStack
	// always uses; the "skip re-probing" spec relies on this to look up the
	// resource_validation_records row.
	validationSkipProbeResourceName = "web"
)

var _ = Describe("Stack validation", func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	It("rejects an invalid thin resource create with aggregated field errors", func() {
		By("Creating a valid stack first")
		stack := shared.CreateSimpleStack("test-validation-thin-agg")
		created := shared.CreateStack(client, orgID, teamName, stack)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})

		By("Adding a resource with a tcp+public port, a valueless env var, and an unknown dependency")
		resource := openapi.NewStackResource("bad-resource")
		resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource(shared.TestImage)})

		badPort := openapi.NewPort("db", 5432, true)
		badPort.SetProtocol("tcp")
		resource.SetPorts([]openapi.Port{*badPort})

		exec := openapi.NewExecutionConfig()
		exec.SetEnvironmentVariables([]openapi.EnvVar{*openapi.NewEnvVar("MISSING_VALUE")})
		resource.SetExecutionConfig(*exec)

		resource.SetDependsOn([]string{"nonexistent"})

		apiErr := shared.CreateStackResourceExpectError(client, orgID, teamName, stackID, resource, 400)

		By("Verifying the aggregated error details")
		codes := shared.ErrorValidationCodes(apiErr)
		Expect(len(codes)).To(BeNumerically(">=", 3), "expected at least 3 aggregated field errors, got: %v", codes)
		Expect(codes).To(ContainElements(
			errors.VErrPublicPortNotHTTP,
			errors.VErrEnvValueMissing,
			errors.VErrDependencyUnknown,
		))
	})

	It("accepts a fat stack create with an unpullable image (no network probes)", func() {
		resource := openapi.NewStackResource("app")
		resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource(validationUnpullableImageRef)})
		spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
		stack := openapi.NewStack("test-validation-unpullable", *spec)

		By("Creating the stack — should succeed without probing the registry")
		created := shared.CreateStack(client, orgID, teamName, stack)
		Expect(created.GetId()).NotTo(BeEmpty())
		Expect(created.Spec.StackResources).To(HaveLen(1))
		Expect(created.Spec.StackResources[0].Source.Image.GetRef()).To(Equal(validationUnpullableImageRef))

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, created.GetId())
			shared.WaitForStackDeleted(client, orgID, teamName, created.GetId(), 1*time.Minute)
		})
	})

	It("fails a release with structured validation errors for a missing image", func() {
		resourceName := "missing-image-app"
		resource := openapi.NewStackResource(resourceName)
		resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource(validationMissingImageRef)})
		spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
		stack := openapi.NewStack("test-validation-missing-image", *spec)

		By("Creating the stack (fat create — no probing at create time)")
		created := shared.CreateStack(client, orgID, teamName, stack)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})

		By("Creating a release — should start Pending")
		release := shared.CreateRelease(client, orgID, teamName, stackID)
		Expect(string(release.GetState())).To(Equal(string(models.ReleaseStatePending)))

		By("Waiting for the release worker's async image probe to fail the release")
		failed := shared.WaitForReleaseState(client, orgID, teamName, stackID, release.GetId(),
			string(models.ReleaseStateFailed), 2*time.Minute)

		Expect(failed.ValidationErrors).NotTo(BeEmpty(), "expected structured validation_errors on the failed release")
		verr := failed.ValidationErrors[0]
		Expect(verr.GetCode()).To(Equal(errors.VErrImageNotFound))
		Expect(verr.GetResourceName()).To(Equal(resourceName))
	})

	It("resolves the default branch at release time", func() {
		resourceName := "default-branch-app"
		resource := openapi.NewStackResource(resourceName)
		gitSource := openapi.NewGitSource(validationDefaultBranchRepoURL)
		resource.SetSource(openapi.SourceSpec{Git: gitSource})
		spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
		stack := openapi.NewStack("test-validation-default-branch", *spec)

		By("Creating the stack (create succeeds without any git call)")
		created := shared.CreateStack(client, orgID, teamName, stack)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})

		By("Creating a release — resolvePins resolves the repo's default branch synchronously")
		release := shared.CreateRelease(client, orgID, teamName, stackID)

		Expect(release.Pins).NotTo(BeNil(), "expected release.pins to be set")
		pins := release.Pins.GetResources()
		resourcePins, ok := pins[resourceName]
		Expect(ok).To(BeTrue(), "expected pins for resource %s", resourceName)
		Expect(resourcePins.GetGitSha()).NotTo(BeEmpty(), "expected a resolved git_sha")

		By("Fetching the release detail and verifying the snapshot carries the resolved branch")
		detail := shared.GetRelease(client, orgID, teamName, stackID, release.GetId())
		Expect(detail.Snapshot).NotTo(BeNil())

		var snapshotResource *openapi.StackResource
		for i := range detail.Snapshot.Resources {
			if detail.Snapshot.Resources[i].Name == resourceName {
				snapshotResource = &detail.Snapshot.Resources[i]
				break
			}
		}
		Expect(snapshotResource).NotTo(BeNil(), "expected snapshot to contain resource %s", resourceName)
		Expect(snapshotResource.Source).NotTo(BeNil())
		Expect(snapshotResource.Source.Git).NotTo(BeNil())
		Expect(snapshotResource.Source.Git.GetBranch()).NotTo(BeEmpty(), "expected the resolved default branch to be pinned")
		Expect(snapshotResource.Source.Git.GetCommit()).NotTo(BeEmpty(), "expected the resolved commit SHA to be pinned")
	})

	It("returns structured 400 from POST /releases for a bad branch", func() {
		resourceName := "bad-branch-app"
		resource := openapi.NewStackResource(resourceName)
		gitSource := openapi.NewGitSource(validationDefaultBranchRepoURL)
		gitSource.SetBranch(validationBadBranchName)
		resource.SetSource(openapi.SourceSpec{Git: gitSource})
		spec := openapi.NewStackSpec([]openapi.StackResource{*resource})
		stack := openapi.NewStack("test-validation-bad-branch", *spec)

		created := shared.CreateStack(client, orgID, teamName, stack)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})

		apiErr := shared.CreateReleaseExpectError(client, orgID, teamName, stackID, 400)

		codes := shared.ErrorValidationCodes(apiErr)
		Expect(codes).To(ContainElement(errors.VErrGitBranchNotFound))
	})

	It("skips re-probing an unchanged image on the next release", func() {
		testEnv := GetEnvironment()

		By("Creating a stack with a real pullable public image and deploying it")
		stack := shared.CreateSimpleStack("test-validation-skip-probe")
		created, release1 := shared.CreateStackAndDeploy(client, orgID, teamName, stack)
		stackID := created.GetId()

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 1*time.Minute)
		})

		By("Waiting for the first release to converge (generous timeout: anonymous pulls can be rate-limited and requeued)")
		shared.WaitForReleaseReleased(client, orgID, teamName, stackID, release1.GetId(), 5*time.Minute)

		By("Reading the validation record written by the first release's image probe")
		rec1, err := testEnv.Database.GetResourceValidationRecord(context.Background(), stackID,
			validationSkipProbeResourceName, models.ValidationCheckImagePull)
		Expect(err).NotTo(HaveOccurred(), "expected a resource_validation_records row after the first release")
		Expect(rec1).NotTo(BeNil())
		firstValidatedAt := rec1.ValidatedAt

		By("Creating a second release for the same, unchanged image")
		release2 := shared.CreateRelease(client, orgID, teamName, stackID)
		shared.WaitForReleaseReleased(client, orgID, teamName, stackID, release2.GetId(), 5*time.Minute)

		By("Verifying the validation record's validated_at is unchanged — the probe was skipped")
		rec2, err := testEnv.Database.GetResourceValidationRecord(context.Background(), stackID,
			validationSkipProbeResourceName, models.ValidationCheckImagePull)
		Expect(err).NotTo(HaveOccurred())
		Expect(rec2).NotTo(BeNil())
		Expect(rec2.ValidatedAt.Equal(firstValidatedAt)).To(BeTrue(),
			"expected validated_at to be unchanged (probe skipped), first=%v second=%v", firstValidatedAt, rec2.ValidatedAt)
	})
})
