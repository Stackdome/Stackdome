package int

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Resource Reference Delete Protection", func() {
	var client *openapi.APIClient
	var orgID string
	teamName := models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")
		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Implicit references", func() {
		It("should block deletion of a secret used as an image pull secret", func() {
			By("Creating a DockerRegistry secret for image pull")
			secret := openapi.NewSecret("test-pull-secret", openapi.DOCKER_REGISTRY, []openapi.SecretData{
				*openapi.NewSecretData("registry", "docker.io"),
				*openapi.NewSecretData("username", "user"),
				*openapi.NewSecretData("password", "pass"),
			})
			created := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := created.GetId()

			By("Creating a stack with a resource that uses the secret as a pull secret")
			resource := openapi.NewStackResource("web")
			imageSpec := openapi.NewImageSpec("nginx:1.25-alpine")
			pullSecret := openapi.NewSecretRef(secretID)
			imageSpec.SetPullSecret(*pullSecret)
			resource.SetImageSpec(*imageSpec)
			resource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})

			stack := shared.CreateSkipProvisioningStack("test-pull-ref", []openapi.StackResource{*resource})
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
				shared.DeleteSecret(client, orgID, teamName, secretID)
			})

			By("Attempting to delete the secret — expecting 409")
			httpResp, err := shared.DeleteSecretRaw(client, orgID, teamName, secretID)
			Expect(err).To(HaveOccurred(), "expected delete to fail")
			Expect(httpResp.StatusCode).To(Equal(http.StatusConflict))
		})
	})

	Context("Explicit connection references", func() {
		It("should block deletion while connected and allow after disconnection", func() {
			By("Creating a secret")
			secret := shared.CreateGenericSecret("test-conn-secret", map[string]string{
				"api_key": "test-key-value",
			})
			created := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := created.GetId()

			By("Creating a stack with an env connection to the secret")
			resource := openapi.NewStackResource("web")
			imageSpec := openapi.NewImageSpec("nginx:1.25-alpine")
			resource.SetImageSpec(*imageSpec)
			resource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})

			conn := shared.SecretEnvConnection(secretID, "web", "API_KEY", "api_key")

			stack := shared.CreateSkipProvisioningStackWithConnections(
				"test-conn-ref",
				[]openapi.StackResource{*resource},
				[]openapi.StackConnection{conn},
			)
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Attempting to delete the secret — expecting 409")
			httpResp, err := shared.DeleteSecretRaw(client, orgID, teamName, secretID)
			Expect(err).To(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusConflict))

			By("Removing the connection")
			connections := shared.ListStackConnections(client, orgID, teamName, stackID)
			Expect(connections).To(HaveLen(1))
			shared.DeleteStackConnection(client, orgID, teamName, stackID, connections[0].GetId())

			By("Deleting the secret — expecting success")
			shared.DeleteSecret(client, orgID, teamName, secretID)
		})
	})

	Context("Release-retained references", func() {
		It("should block deletion when a Released release still references the secret", func() {
			By("Creating a secret")
			secret := shared.CreateGenericSecret("test-release-secret", map[string]string{
				"db_url": "postgres://localhost/test",
			})
			created := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := created.GetId()

			By("Creating a stack with a connection to the secret (simulated Released)")
			resource := openapi.NewStackResource("web")
			imageSpec := openapi.NewImageSpec(shared.TestImage)
			resource.SetImageSpec(*imageSpec)
			resource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})

			conn := shared.SecretEnvConnection(secretID, "web", "DB_URL", "db_url")

			stack := shared.CreateSimulatedReleaseStackWithConnections(
				"test-release-ref",
				[]openapi.StackResource{*resource},
				[]openapi.StackConnection{conn},
				string(models.ReleaseStateReleased),
			)
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
				shared.DeleteSecret(client, orgID, teamName, secretID)
			})

			By("Deploying the stack")
			release := shared.CreateRelease(client, orgID, teamName, stackID)

			By("Waiting for the simulated release to reach Released state")
			shared.WaitForReleaseReleased(client, orgID, teamName, stackID, release.GetId(), 1*time.Minute)

			By("Updating the stack to remove the secret connection")
			updatedResource := openapi.NewStackResource("web")
			updatedImageSpec := openapi.NewImageSpec(shared.TestImage)
			updatedResource.SetImageSpec(*updatedImageSpec)
			updatedResource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})
			updatedSpec := openapi.NewStackSpec([]openapi.StackResource{*updatedResource})
			updatedStack := openapi.NewStack("test-release-ref", *updatedSpec)
			updatedStack.SetAnnotations([]openapi.Annotation{
				*openapi.NewAnnotation(shared.SkipProvisioningAnnotationKey, "true"),
				*openapi.NewAnnotation(shared.SimulateReleaseStateAnnotationKey, string(models.ReleaseStateReleased)),
			})
			shared.UpdateStack(client, orgID, teamName, stackID, updatedStack)

			By("Verifying the spec no longer references the secret")
			fetchedStack := shared.GetStack(client, orgID, teamName, stackID)
			Expect(fetchedStack.Spec.GetConnections()).To(BeEmpty())

			By("Attempting to delete the secret — expecting 409 (Released release still grips)")
			httpResp, err := shared.DeleteSecretRaw(client, orgID, teamName, secretID)
			Expect(err).To(HaveOccurred(), "expected delete to fail because Released release still references the secret")
			Expect(httpResp.StatusCode).To(Equal(http.StatusConflict))

			By("Deleting the stack (cascades releases + reference rows)")
			shared.DeleteStack(client, orgID, teamName, stackID)
			shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)

			By("Deleting the secret — expecting success after stack deletion")
			shared.DeleteSecret(client, orgID, teamName, secretID)
		})

		It("should NOT block deletion when only a Failed release references the secret", func() {
			By("Creating a secret")
			secret := shared.CreateGenericSecret("test-failed-secret", map[string]string{
				"db_url": "postgres://localhost/test",
			})
			created := shared.CreateSecret(client, orgID, teamName, secret)
			secretID := created.GetId()

			By("Creating a stack with a connection to the secret (simulated Failed)")
			resource := openapi.NewStackResource("web")
			imageSpec := openapi.NewImageSpec(shared.TestImage)
			resource.SetImageSpec(*imageSpec)
			resource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})

			conn := shared.SecretEnvConnection(secretID, "web", "DB_URL", "db_url")

			stack := shared.CreateSimulatedReleaseStackWithConnections(
				"test-failed-ref",
				[]openapi.StackResource{*resource},
				[]openapi.StackConnection{conn},
				string(models.ReleaseStateFailed),
			)
			createdStack := shared.CreateStack(client, orgID, teamName, stack)
			stackID := createdStack.GetId()

			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, teamName, stackID)
				shared.WaitForStackDeleted(client, orgID, teamName, stackID, 2*time.Minute)
			})

			By("Deploying the stack")
			release := shared.CreateRelease(client, orgID, teamName, stackID)

			By("Waiting for the simulated release to reach Failed state")
			shared.WaitForReleaseState(client, orgID, teamName, stackID, release.GetId(), string(models.ReleaseStateFailed), 1*time.Minute)

			By("Updating the stack to remove the secret connection")
			updatedResource := openapi.NewStackResource("web")
			updatedImageSpec := openapi.NewImageSpec(shared.TestImage)
			updatedResource.SetImageSpec(*updatedImageSpec)
			updatedResource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, false)})
			updatedSpec := openapi.NewStackSpec([]openapi.StackResource{*updatedResource})
			updatedStack := openapi.NewStack("test-failed-ref", *updatedSpec)
			updatedStack.SetAnnotations([]openapi.Annotation{
				*openapi.NewAnnotation(shared.SkipProvisioningAnnotationKey, "true"),
				*openapi.NewAnnotation(shared.SimulateReleaseStateAnnotationKey, string(models.ReleaseStateFailed)),
			})
			shared.UpdateStack(client, orgID, teamName, stackID, updatedStack)

			By("Deleting the secret — expecting success (Failed release does NOT grip)")
			shared.DeleteSecret(client, orgID, teamName, secretID)
		})
	})
})
