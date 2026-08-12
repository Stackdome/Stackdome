//go:build cloud_e2e

package cloudint

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/computeaccess"
	serviceerrors "github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/bootstrap"
	"github.com/Stackdome/stackdome/test/int/shared"
	registryv1alpha1 "stackdome.io/cluster-agent/api/registry/v1alpha1"
)

var _ = Describe("Stackdome Cloud compute policy", Ordered, func() {
	const projectName = models.DefaultProjectName

	var (
		primaryStack *openapi.Stack
		buildStack   *openapi.Stack
		postgres     *openapi.PostgresAddon
	)

	It("boots in Cloud shared mode with pending infrastructure and no compute grant", func() {
		ctx := context.Background()
		db := cloudEnv.Database.GetSessionFactory().New(ctx)

		var registry models.ClusterImageRegistry
		Expect(db.Where("organisation_id = ? AND cluster_id = ?", cloudEnv.OrgID, cloudEnv.ClusterID).
			Take(&registry).Error).To(Succeed())
		Expect(registry.Name).To(Equal(cloudEnv.RegistryName))
		Expect(registry.BackendStorageSize).To(Equal("2Gi"))
		Expect(registry.BackendStorageClass).To(Equal(bootstrap.CloudTestStorageClass))
		Expect(registry.Status).NotTo(BeNil())
		Expect(registry.Status.State).To(Equal(models.RegistryStatePending))

		clusterRegistry := &registryv1alpha1.ClusterRegistry{}
		err := cloudEnv.Cluster.GetClient().Get(ctx, client.ObjectKey{Name: registry.Name}, clusterRegistry)
		Expect(apierrors.IsNotFound(err)).To(BeTrue(), "signup must not eagerly create the ClusterRegistry CR")

		entitlements, leases := computeAccessCounts(ctx, cloudEnv.OrgID)
		Expect(entitlements).To(BeZero())
		Expect(leases).To(BeZero())
	})

	It("rolls back a rejected first grant and activates exactly once on valid use", func() {
		invalidAddon := postgresAddon("first-use-rejected", 2, "1Gi")
		apiErr := shared.CreatePostgresAddonExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			invalidAddon,
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		ctx := context.Background()
		entitlements, leases := computeAccessCounts(ctx, cloudEnv.OrgID)
		Expect(entitlements).To(BeZero(), "a rejected transaction must not consume the trial")
		Expect(leases).To(BeZero(), "a rejected transaction must not reserve shared capacity")
		var rejectedAddonCount int64
		Expect(cloudEnv.Database.GetSessionFactory().New(ctx).
			Model(&models.PostgresAddon{}).
			Where("organisation_id = ?", cloudEnv.OrgID).
			Count(&rejectedAddonCount).Error).To(Succeed())
		Expect(rejectedAddonCount).To(BeZero())

		stack := shared.CreateSimpleStack("cloud-primary")
		stack.Spec.StackResources[0].SetReplicas(7)
		primaryStack = shared.CreateStack(cloudEnv.Client, cloudEnv.OrgID, projectName, stack)
		Expect(primaryStack.Spec.StackResources).To(HaveLen(1))
		Expect(primaryStack.Spec.StackResources[0].GetReplicas()).To(Equal(int32(1)))

		var entitlement computeaccess.ComputeEntitlement
		Expect(cloudEnv.Database.GetSessionFactory().New(ctx).
			Where("organisation_id = ?", cloudEnv.OrgID).
			Take(&entitlement).Error).To(Succeed())
		Expect(entitlement.Source).To(Equal(computeaccess.ComputeEntitlementSourceTrial))
		Expect(entitlement.Status).To(Equal(computeaccess.ComputeEntitlementStatusActive))
		Expect(entitlement.ExpiresAt).NotTo(BeNil())

		var lease computeaccess.SharedComputeLease
		Expect(cloudEnv.Database.GetSessionFactory().New(ctx).
			Where("organisation_id = ?", cloudEnv.OrgID).
			Take(&lease).Error).To(Succeed())
		Expect(lease.EntitlementID).To(Equal(entitlement.ID))
		Expect(lease.State).To(Equal(computeaccess.SharedComputeLeaseStateActive))
	})

	It("enforces stack and resource limits on creates and replacements", func() {
		second := shared.CreateShellStack(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			shared.CreateSimpleStack("cloud-second"),
		)

		apiErr := shared.CreateStackExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			shared.CreateSimpleStack("cloud-third"),
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		shared.CreateStackResource(cloudEnv.Client, cloudEnv.OrgID, projectName, primaryStack.GetId(), stackResource("worker"))
		shared.CreateStackResource(cloudEnv.Client, cloudEnv.OrgID, projectName, primaryStack.GetId(), stackResource("jobs"))
		apiErr = shared.CreateStackResourceExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			primaryStack.GetId(),
			stackResource("over-limit"),
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		overLimitReplacement := stackWithResources("cloud-primary", "one", "two", "three", "four")
		apiErr = shared.ApplyStackExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			primaryStack.GetId(),
			overLimitReplacement,
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)
		Expect(shared.GetStack(cloudEnv.Client, cloudEnv.OrgID, projectName, primaryStack.GetId()).Spec.StackResources).To(HaveLen(3))

		primaryStack = shared.UpdateStack(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			primaryStack.GetId(),
			stackWithResources("cloud-primary", "web"),
		)
		Expect(primaryStack.Spec.StackResources).To(HaveLen(1))
		Expect(primaryStack.Spec.StackResources[0].GetReplicas()).To(Equal(int32(1)))

		entitlements, leases := computeAccessCounts(context.Background(), cloudEnv.OrgID)
		Expect(entitlements).To(Equal(int64(1)))
		Expect(leases).To(Equal(int64(1)))

		shared.DeleteStack(cloudEnv.Client, cloudEnv.OrgID, projectName, second.GetId())
		shared.WaitForStackDeleted(cloudEnv.Client, cloudEnv.OrgID, projectName, second.GetId(), 2*time.Minute)
	})

	It("enforces the storage class and eagerly provisions volumes within quota", func() {
		created := createStackVolume(primaryStack.GetId(), volume("cloud-data", "1Gi"), http.StatusCreated)
		Expect(created.Spec.GetSize()).To(Equal("1Gi"))
		Expect(created.Spec.GetStorageClass()).To(Equal(bootstrap.CloudTestStorageClass))
		waitForVolumeReady(created.GetId())

		apiErr := createStackVolumeExpectError(primaryStack.GetId(), volume("cloud-too-large", "3Gi"), http.StatusBadRequest)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		second := createStackVolume(primaryStack.GetId(), volume("cloud-cache", "1Gi"), http.StatusCreated)
		waitForVolumeReady(second.GetId())
		apiErr = createStackVolumeExpectError(primaryStack.GetId(), volume("cloud-third-volume", "1Gi"), http.StatusBadRequest)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)
	})

	It("enforces PostgreSQL instance, storage, and addon limits", func() {
		apiErr := shared.CreatePostgresAddonExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			postgresAddon("cloud-too-large-postgres", 1, "3Gi"),
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		postgres = shared.CreatePostgresAddon(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			postgresAddon("cloud-postgres", 1, "1Gi"),
		)
		Expect(postgres.Spec.Instances.GetCount()).To(Equal(int32(1)))
		Expect(postgres.Spec.Storage.GetSize()).To(Equal("1Gi"))
		shared.WaitForAddonReady(cloudEnv.Client, cloudEnv.OrgID, projectName, postgres.GetId(), 10*time.Minute)

		apiErr = shared.CreatePostgresAddonExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			postgresAddon("cloud-second-postgres", 1, "1Gi"),
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeQuotaExceeded)

		oversizeUpdate := postgresAddon(postgres.GetName(), 1, "3Gi")
		apiErr = shared.UpdatePostgresAddonExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			postgres.GetId(),
			oversizeUpdate,
			http.StatusBadRequest,
		)
		expectServiceError(apiErr, serviceerrors.ErrorBadRequest)

		persisted := shared.GetPostgresAddon(cloudEnv.Client, cloudEnv.OrgID, projectName, postgres.GetId())
		Expect(persisted.Spec.Instances.GetCount()).To(Equal(int32(1)))
		Expect(persisted.Spec.Storage.GetSize()).To(Equal("1Gi"))
	})

	It("connects a stack workload to its PostgreSQL addon", func() {
		const database = "cloudapp"
		shared.WaitForConditionTrue(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			postgres.GetId(),
			string(models.PostgresAddonConditionDatabasesApplied),
			2*time.Minute,
		)

		stack := postgresConnectivityStack("cloud-postgres-client", postgres.GetId(), database)
		created, _ := shared.CreateStackAndDeploy(cloudEnv.Client, cloudEnv.OrgID, projectName, stack)
		DeferCleanup(func() {
			shared.DeleteStack(cloudEnv.Client, cloudEnv.OrgID, projectName, created.GetId())
			shared.WaitForStackDeleted(cloudEnv.Client, cloudEnv.OrgID, projectName, created.GetId(), 2*time.Minute)
		})

		shared.WaitForStackReady(cloudEnv.Client, cloudEnv.OrgID, projectName, created.GetId(), 5*time.Minute)
		waitForPostgresQuerySuccess(created.GetNamespace(), "app", 2*time.Minute)
	})

	It("creates the pending registry only when a build release needs it", func() {
		ctx := context.Background()
		clusterRegistry := &registryv1alpha1.ClusterRegistry{}
		err := cloudEnv.Cluster.GetClient().Get(ctx, client.ObjectKey{Name: cloudEnv.RegistryName}, clusterRegistry)
		Expect(apierrors.IsNotFound(err)).To(BeTrue())

		buildStackSpec := shared.CreateStackWithBuildSource("cloud-build", "https://github.com/Stackdome/test-repo.git")
		buildStackSpec.Spec.StackResources[0].Source.Git.SetDockerfilePath("docker/Dockerfile.prod")
		buildStack = shared.CreateStack(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			buildStackSpec,
		)
		release := shared.CreateRelease(cloudEnv.Client, cloudEnv.OrgID, projectName, buildStack.GetId())

		Eventually(func(g Gomega) {
			actual := &registryv1alpha1.ClusterRegistry{}
			g.Expect(cloudEnv.Cluster.GetClient().Get(ctx, client.ObjectKey{Name: cloudEnv.RegistryName}, actual)).To(Succeed())
			g.Expect(actual.Spec.Storage.Size).To(Equal("2Gi"))
			g.Expect(actual.Spec.Storage.StorageClass).NotTo(BeNil())
			g.Expect(*actual.Spec.Storage.StorageClass).To(Equal(bootstrap.CloudTestStorageClass))
		}, 2*time.Minute, 2*time.Second).Should(Succeed())

		shared.WaitForReleaseReleased(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			buildStack.GetId(),
			release.GetId(),
			10*time.Minute,
		)
	})

	It("enforces the platform lease ceiling without partially persisting the rejected request", func() {
		suffix := time.Now().UnixNano()
		signup := shared.SignupNewUser(
			"Capacity Test",
			fmt.Sprintf("cloud-capacity-%d@example.com", suffix),
			"supersecret123",
			fmt.Sprintf("Cloud Capacity %d", suffix),
		)
		secondOrgID := signup.User.GetOrganisationId()
		secondClient := shared.AuthenticatedClient(signup.GetJwtToken())

		apiErr := shared.CreateStackExpectError(
			secondClient,
			secondOrgID,
			projectName,
			shared.CreateSimpleStack("capacity-rejected"),
			http.StatusTooManyRequests,
		)
		expectServiceError(apiErr, serviceerrors.ErrorSharedComputeCapacityReached)

		entitlements, leases := computeAccessCounts(context.Background(), secondOrgID)
		Expect(entitlements).To(BeZero())
		Expect(leases).To(BeZero())
		var stackCount int64
		Expect(cloudEnv.Database.GetSessionFactory().New(context.Background()).
			Model(&models.Stack{}).
			Where("organisation_id = ?", secondOrgID).
			Count(&stackCount).Error).To(Succeed())
		Expect(stackCount).To(BeZero())
	})

	It("blocks new mutations after expiry while allowing deletion", func() {
		ctx := context.Background()
		expiredAt := time.Now().Add(-time.Minute)
		Expect(cloudEnv.Database.GetSessionFactory().New(ctx).
			Model(&computeaccess.ComputeEntitlement{}).
			Where("organisation_id = ?", cloudEnv.OrgID).
			Update("expires_at", expiredAt).Error).To(Succeed())

		apiErr := shared.CreateStackResourceExpectError(
			cloudEnv.Client,
			cloudEnv.OrgID,
			projectName,
			primaryStack.GetId(),
			stackResource("after-expiry"),
			http.StatusGone,
		)
		expectServiceError(apiErr, serviceerrors.ErrorComputeAccessInactive)

		shared.DeletePostgresAddon(cloudEnv.Client, cloudEnv.OrgID, projectName, postgres.GetId())
		if buildStack != nil {
			shared.DeleteStack(cloudEnv.Client, cloudEnv.OrgID, projectName, buildStack.GetId())
		}
		shared.DeleteStack(cloudEnv.Client, cloudEnv.OrgID, projectName, primaryStack.GetId())
	})
})

func computeAccessCounts(ctx context.Context, organisationID string) (int64, int64) {
	db := cloudEnv.Database.GetSessionFactory().New(ctx)
	var entitlementCount int64
	Expect(db.Model(&computeaccess.ComputeEntitlement{}).
		Where("organisation_id = ?", organisationID).
		Count(&entitlementCount).Error).To(Succeed())
	var leaseCount int64
	Expect(db.Model(&computeaccess.SharedComputeLease{}).
		Where("organisation_id = ?", organisationID).
		Count(&leaseCount).Error).To(Succeed())
	return entitlementCount, leaseCount
}

func expectServiceError(apiErr *openapi.GenericOpenAPIError, code serviceerrors.ServiceErrorCode) {
	var response openapi.Error
	Expect(json.Unmarshal(apiErr.Body(), &response)).To(Succeed())
	Expect(response.GetCode()).To(Equal(fmt.Sprintf("%d", code)))
}

func stackResource(name string) *openapi.StackResource {
	resource := openapi.NewStackResource(name)
	resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource(shared.TestImage)})
	resource.SetReplicas(9)
	return resource
}

func stackWithResources(name string, resourceNames ...string) *openapi.Stack {
	resources := make([]openapi.StackResource, 0, len(resourceNames))
	for _, resourceName := range resourceNames {
		resources = append(resources, *stackResource(resourceName))
	}
	spec := openapi.NewStackSpec()
	spec.SetStackResources(resources)
	return openapi.NewStack(name, *spec)
}

func volume(name, size string) *openapi.Volume {
	spec := openapi.NewVolumeSpec(size, false, openapi.READ_WRITE_ONCE)
	return openapi.NewVolume(name, *spec)
}

func createStackVolume(stackID string, requested *openapi.Volume, expectedStatus int) *openapi.Volume {
	created, response, err := cloudEnv.Client.DefaultApi.
		ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost(
			context.Background(),
			cloudEnv.OrgID,
			models.DefaultProjectName,
			stackID,
		).
		Volume(*requested).
		Execute()
	Expect(err).NotTo(HaveOccurred())
	Expect(response.StatusCode).To(Equal(expectedStatus))
	Expect(created).NotTo(BeNil())
	return created
}

func createStackVolumeExpectError(stackID string, requested *openapi.Volume, expectedStatus int) *openapi.GenericOpenAPIError {
	_, response, err := cloudEnv.Client.DefaultApi.
		ApiV1OrganizationsOrgIdProjectsProjectNameStacksIdVolumesPost(
			context.Background(),
			cloudEnv.OrgID,
			models.DefaultProjectName,
			stackID,
		).
		Volume(*requested).
		Execute()
	Expect(err).To(HaveOccurred())
	Expect(response.StatusCode).To(Equal(expectedStatus))
	var apiErr *openapi.GenericOpenAPIError
	Expect(stderrors.As(err, &apiErr)).To(BeTrue())
	return apiErr
}

func waitForVolumeReady(volumeID string) {
	Eventually(func(g Gomega) {
		volume, response, err := cloudEnv.Client.DefaultApi.
			ApiV1OrganizationsOrgIdProjectsProjectNameVolumesIdGet(
				context.Background(),
				cloudEnv.OrgID,
				models.DefaultProjectName,
				volumeID,
			).
			Execute()
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(response.StatusCode).To(Equal(http.StatusOK))
		g.Expect(volume.Status).NotTo(BeNil())
		g.Expect(volume.Status.GetPhase()).To(Equal(models.VolumePhaseReady))
	}, 3*time.Minute, 2*time.Second).Should(Succeed())
}

func postgresAddon(name string, instances int32, storageSize string) *openapi.PostgresAddon {
	version := openapi.NewPostgresVersion(16)
	version.SetMinor(6)
	spec := openapi.NewPostgresAddonSpec(
		*version,
		*openapi.NewPostgresInstances(instances),
		*openapi.NewPostgresStorage(storageSize),
	)
	spec.SetDatabases([]openapi.PostgresDatabase{*openapi.NewPostgresDatabase("cloudapp")})
	return openapi.NewPostgresAddon(name, *spec)
}

func postgresConnectivityStack(name, addonID, database string) *openapi.Stack {
	stack := shared.CreateStackWithPostgresAddon(name, addonID, database)
	resource := &stack.Spec.StackResources[0]
	resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("postgres:16-alpine")})
	resource.SetPorts([]openapi.Port{})

	execution := openapi.NewExecutionConfig()
	execution.SetCommand([]string{
		"sh",
		"-ec",
		`until result="$(PGPASSWORD="$PG_PASSWORD" psql --host "$PG_HOST" --port "$PG_PORT" --username "$PG_USER" --dbname "$PG_DATABASE" --tuples-only --no-align --command 'SELECT 1' 2>/dev/null)" && [ "$result" = "1" ]; do sleep 2; done
echo stackdome-postgres-select-1=1
exec sleep 3600`,
	})
	resource.SetExecutionConfig(*execution)
	return stack
}

func waitForPostgresQuerySuccess(namespace, resourceName string, timeout time.Duration) {
	ctx := context.Background()
	clientset, err := cloudEnv.Cluster.GetKubeClient()
	Expect(err).NotTo(HaveOccurred())

	Eventually(func(g Gomega) {
		pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: "resource=" + resourceName,
		})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(pods.Items).NotTo(BeEmpty())

		var logs strings.Builder
		for i := range pods.Items {
			for _, container := range pods.Items[i].Spec.Containers {
				output, logErr := clientset.CoreV1().Pods(namespace).
					GetLogs(pods.Items[i].Name, &corev1.PodLogOptions{Container: container.Name}).
					DoRaw(ctx)
				if logErr == nil {
					logs.Write(output)
				}
			}
		}
		g.Expect(logs.String()).To(ContainSubstring("stackdome-postgres-select-1=1"))
	}, timeout, 2*time.Second).Should(Succeed())
}
