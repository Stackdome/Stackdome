package int

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/slug"
	"github.com/Stackdome/stackdome/test/int/bootstrap"
	"github.com/Stackdome/stackdome/test/int/shared"
	"k8s.io/utils/ptr"
)

// shortOrgIDLength mirrors the unexported services.shortOrgIDLength used by
// registry naming to derive <slug>-<shortOrgID>-<shortClusterID>.
const shortOrgIDLength = 8

func shortTestID(id string) string {
	s := strings.ReplaceAll(id, "-", "")
	if len(s) > shortOrgIDLength {
		s = s[:shortOrgIDLength]
	}
	return s
}

func exposedProvisioningStack(name string) *openapi.Stack {
	resource := openapi.NewStackResource("web")
	resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})
	resource.SetPorts([]openapi.Port{*openapi.NewPort("http", 80, true)})

	spec := openapi.NewStackSpec()
	spec.SetStackResources([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

func basicProvisioningStack(name string) *openapi.Stack {
	resource := openapi.NewStackResource("web")
	resource.SetSource(openapi.SourceSpec{Image: openapi.NewImageSource("nginx:1.25-alpine")})

	spec := openapi.NewStackSpec()
	spec.SetStackResources([]openapi.StackResource{*resource})
	return openapi.NewStack(name, *spec)
}

var _ = Describe("Platform provisioning", func() {
	It("seeds a Platform=true cluster and the platform base domain at boot", func() {
		ctx := context.Background()
		db := GetEnvironment().Database.GetSessionFactory().New(ctx)

		By("finding the platform organisation")
		var platformOrg models.Organisation
		Expect(db.Where("platform = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())

		By("verifying the platform org owns a Platform=true cluster")
		var platformCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&platformCluster).Error).To(Succeed())
		Expect(platformCluster.Platform).To(BeTrue())

		By("verifying the platform org has the base domain")
		var platformDomain models.OrganisationDomain
		Expect(db.Where(&models.OrganisationDomain{
			OrganisationID: platformOrg.ID,
			Domain:         bootstrap.PlatformProvisioningBaseDomain,
		}).First(&platformDomain).Error).To(Succeed())

		By("verifying the platform org is infrastructure-only: no users, no projects, no registries")
		var userCount int64
		Expect(db.Model(&models.User{}).Where(&models.User{OrganisationID: platformOrg.ID}).Count(&userCount).Error).To(Succeed())
		Expect(userCount).To(BeZero())
		var projectCount int64
		Expect(db.Model(&models.Project{}).Where(&models.Project{OrganisationID: platformOrg.ID}).Count(&projectCount).Error).To(Succeed())
		Expect(projectCount).To(BeZero())
		var registryCount int64
		Expect(db.Model(&models.ClusterImageRegistry{}).Where(&models.ClusterImageRegistry{OrganisationID: platformOrg.ID}).Count(&registryCount).Error).To(Succeed())
		Expect(registryCount).To(BeZero())
	})

	It("seeds <slug>.<base> domain and <slug>-<shortid> registry on fresh signup, and stacks fall back to the platform cluster", func() {
		ctx := context.Background()
		db := GetEnvironment().Database.GetSessionFactory().New(ctx)
		projectName := models.DefaultProjectName

		By("resolving the seeded platform cluster")
		var platformOrg models.Organisation
		Expect(db.Where("platform = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())
		var platformCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&platformCluster).Error).To(Succeed())

		By("signing up a fresh user with a new organisation")
		ts := time.Now().UnixNano()
		orgName := fmt.Sprintf("Acme %d", ts)
		email := fmt.Sprintf("acme-%d@example.com", ts)
		resp := shared.SignupNewUser("Acme Admin", email, "supersecret123", orgName)
		newOrgID := resp.User.GetOrganisationId()
		newToken := resp.GetJwtToken()
		Expect(newOrgID).NotTo(BeEmpty())
		Expect(newToken).NotTo(BeEmpty())

		orgSlug := slug.FromOrgName(orgName)
		expectedDomain := fmt.Sprintf("%s.%s", orgSlug, bootstrap.PlatformProvisioningBaseDomain)

		By("verifying the org was seeded a <slug>.<base> domain")
		var orgDomain models.OrganisationDomain
		Expect(db.Where(&models.OrganisationDomain{OrganisationID: newOrgID}).First(&orgDomain).Error).To(Succeed())
		Expect(orgDomain.Domain).To(Equal(expectedDomain))

		By("verifying the org was seeded a <slug>-<shortOrgID>-<shortClusterID> registry on the platform cluster")
		expectedRegistry := fmt.Sprintf("%s-%s-%s", orgSlug, shortTestID(newOrgID), shortTestID(platformCluster.ID))
		var registry models.ClusterImageRegistry
		Expect(db.Where(&models.ClusterImageRegistry{OrganisationID: newOrgID}).First(&registry).Error).To(Succeed())
		Expect(registry.Name).To(Equal(expectedRegistry))
		Expect(registry.ClusterID).To(Equal(platformCluster.ID))

		By("creating and deploying a stack with a publicly exposed port in the new org")
		newClient := shared.AuthenticatedClient(newToken)
		created, _ := shared.CreateStackAndDeploy(newClient, newOrgID, projectName, exposedProvisioningStack("dp-web"))
		stackID := created.GetId()
		Expect(stackID).NotTo(BeEmpty())

		DeferCleanup(func() {
			shared.DeleteStack(newClient, newOrgID, projectName, stackID)
			shared.WaitForStackDeleted(newClient, newOrgID, projectName, stackID, 1*time.Minute)
			// The seeded registry is a real workload + PVC on the shared
			// cluster — drop it so specs don't accumulate registries.
			resp, derr := newClient.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete(
				ctx, newOrgID, registry.ClusterID, registry.ID).Execute()
			Expect(derr).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		By("verifying the stack resolved to the platform cluster via read-time fallback")
		var stackRow models.Stack
		Expect(db.First(&stackRow, "id = ?", stackID).Error).To(Succeed())
		Expect(stackRow.ClusterID).To(Equal(platformCluster.ID))

		By("waiting for the stack to become Ready on the platform cluster")
		shared.WaitForStackReady(newClient, newOrgID, projectName, stackID, 5*time.Minute)

		By("verifying the exposed-port stack domain uses <slug>.<base>")
		Eventually(func(g Gomega) {
			var stackDomain models.StackDomain
			g.Expect(db.Where(&models.StackDomain{StackID: stackID}).First(&stackDomain).Error).To(Succeed())
			g.Expect(stackDomain.Fqdn).To(HaveSuffix("." + expectedDomain))
		}, 2*time.Minute, 5*time.Second).Should(Succeed())
	})

	It("deploys onto an org-owned cluster registered via the API, beating the platform fallback", func() {
		ctx := context.Background()
		env := GetEnvironment()
		db := env.Database.GetSessionFactory().New(ctx)
		projectName := models.DefaultProjectName

		By("signing up a fresh user with a new organisation")
		ts := time.Now().UnixNano()
		orgName := fmt.Sprintf("Owned Org %d", ts)
		email := fmt.Sprintf("owned-%d@example.com", ts)
		resp := shared.SignupNewUser("Owned Admin", email, "supersecret123", orgName)
		orgID := resp.User.GetOrganisationId()
		client := shared.AuthenticatedClient(resp.GetJwtToken())

		By("registering the org's own cluster via the API")
		clusterURL, caData, saToken, err := bootstrap.ExtractAPIServerClusterCredentials(ctx, env.Cluster)
		Expect(err).NotTo(HaveOccurred())
		// Cluster URLs are globally unique, so the single Kind cluster is
		// re-registered under a hostname alias (its API cert covers both SANs).
		switch {
		case strings.Contains(clusterURL, "127.0.0.1"):
			clusterURL = strings.Replace(clusterURL, "127.0.0.1", "localhost", 1)
		case strings.Contains(clusterURL, "localhost"):
			clusterURL = strings.Replace(clusterURL, "localhost", "127.0.0.1", 1)
		default:
			Fail(fmt.Sprintf("cannot derive an alias for cluster URL %q", clusterURL))
		}
		clusterReq := openapi.NewCluster(fmt.Sprintf("owned-cluster-%d", ts), clusterURL, caData, saToken)
		createdCluster, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdClustersPost(ctx, orgID).Cluster(*clusterReq).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
		ownedClusterID := createdCluster.GetId()
		Expect(ownedClusterID).NotTo(BeEmpty())
		DeferCleanup(func() {
			// Deregister the second registration of the shared Kind cluster so
			// its duplicate controller set doesn't keep reconciling for the
			// rest of the suite.
			resp, derr := client.DefaultApi.ApiV1OrganizationsOrgIdClustersIdDelete(ctx, orgID, ownedClusterID).Execute()
			Expect(derr).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		By("attaching a custom domain to the organisation")
		ownedDomain := fmt.Sprintf("owned-%d.test", ts)
		orgResp, httpResp, err := client.DefaultApi.ApiV1OrganizationsIdGet(ctx, orgID).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
		// Round-tripping the GET response fails request validation (readOnly
		// DomainName.id), and the update reconciles domains by fqdn — send
		// fqdn-only entries for every domain the org should end up with.
		domains := []openapi.DomainName{}
		for _, d := range orgResp.GetDomains() {
			domains = append(domains, openapi.DomainName{Fqdn: ptr.To(d.GetFqdn())})
		}
		domains = append(domains, openapi.DomainName{Fqdn: &ownedDomain})
		orgUpdate := openapi.Organisation{Name: orgResp.Name, Domains: domains}
		updatedOrg, httpResp, err := client.DefaultApi.ApiV1OrganizationsIdPut(ctx, orgID).Organisation(orgUpdate).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
		updatedFqdns := []string{}
		for _, d := range updatedOrg.GetDomains() {
			updatedFqdns = append(updatedFqdns, d.GetFqdn())
		}
		Expect(updatedFqdns).To(ContainElement(ownedDomain))

		By("creating and deploying a basic stack in the new org")
		created, _ := shared.CreateStackAndDeploy(client, orgID, projectName, basicProvisioningStack("owned-web"))
		stackID := created.GetId()
		Expect(stackID).NotTo(BeEmpty())

		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, projectName, stackID)
			shared.WaitForStackDeleted(client, orgID, projectName, stackID, 1*time.Minute)
		})

		By("verifying the stack resolved to the org's own cluster, not the platform cluster")
		var stackRow models.Stack
		Expect(db.First(&stackRow, "id = ?", stackID).Error).To(Succeed())
		Expect(stackRow.ClusterID).To(Equal(ownedClusterID))
		var platformCluster models.Cluster
		Expect(db.Where(&models.Cluster{Platform: true}).First(&platformCluster).Error).To(Succeed())
		Expect(stackRow.ClusterID).NotTo(Equal(platformCluster.ID))

		By("waiting for the stack to become Ready on the owned cluster")
		shared.WaitForStackReady(client, orgID, projectName, stackID, 5*time.Minute)

		By("verifying AddCluster auto-created a <slug>-<shortOrgID>-<shortClusterID> registry on the owned cluster")
		expectedRegistry := fmt.Sprintf("%s-%s-%s", slug.FromOrgName(orgName), shortTestID(orgID), shortTestID(ownedClusterID))
		var ownedRegistry models.ClusterImageRegistry
		Expect(db.Where(&models.ClusterImageRegistry{OrganisationID: orgID, ClusterID: ownedClusterID}).
			First(&ownedRegistry).Error).To(Succeed())
		Expect(ownedRegistry.Name).To(Equal(expectedRegistry))

		if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
			By("building from source: the build must push to the owned cluster's registry")
			integration := shared.CreateGitCredentialsIntegration(client, orgID, shared.BuildSourceGitHost, shared.BuildSourceGitUsername, githubToken)
			DeferCleanup(func() {
				shared.DeleteGitIntegration(client, orgID, integration.GetId())
			})

			// Distinct resource name: image_builds rows are keyed by the CR
			// name <resource>-<commit>, and the shared-fixture name at the
			// same commit collides with the platform-cluster build spec.
			const ownedBuildResourceName = "owned-app"
			buildStack := shared.CreateStackWithNamedBuildSource("owned-build", ownedBuildResourceName, shared.BuildSourceRepoURL)
			buildCreated, _ := shared.CreateStackAndDeploy(client, orgID, projectName, buildStack)
			buildStackID := buildCreated.GetId()
			DeferCleanup(func() {
				shared.DeleteStack(client, orgID, projectName, buildStackID)
				shared.WaitForStackDeleted(client, orgID, projectName, buildStackID, 2*time.Minute)
			})

			var buildStackRow models.Stack
			Expect(db.First(&buildStackRow, "id = ?", buildStackID).Error).To(Succeed())
			Expect(buildStackRow.ClusterID).To(Equal(ownedClusterID))

			By("verifying the build resource resolved the owned cluster's registry")
			var buildResource models.StackResource
			Expect(db.Where(&models.StackResource{StackID: buildStackID, Name: ownedBuildResourceName}).
				First(&buildResource).Error).To(Succeed())
			Expect(buildResource.BuildConfig).NotTo(BeNil())
			Expect(buildResource.BuildConfig.BuildImageRepository.ClusterRegistryName).To(Equal(ownedRegistry.Name))

			By("waiting for the built stack to become Ready on the owned cluster")
			shared.WaitForStackReady(client, orgID, projectName, buildStackID, 10*time.Minute)

			By("verifying the deployment runs an image from the owned cluster's registry")
			clusterClient := env.Cluster.GetClient()
			deploy, derr := shared.GetDeploymentForStackResource(ctx, clusterClient, buildCreated.GetNamespace(), ownedBuildResourceName)
			Expect(derr).NotTo(HaveOccurred())
			Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deploy.Spec.Template.Spec.Containers[0].Image).To(ContainSubstring(ownedRegistry.Name))
		} else {
			By("skipping the build case: GITHUB_TOKEN not set")
		}

		By("verifying builds fail fast once the owned cluster has no registry, even though another cluster's registry row exists")
		// DB-only row on the platform cluster: resolution must ignore it.
		otherClusterRegistry := &models.ClusterImageRegistry{
			OrganisationID:     orgID,
			ClusterID:          platformCluster.ID,
			Name:               fmt.Sprintf("other-cluster-reg-%d", ts),
			BackendStorageSize: models.DefaultRegistryStorageSize,
		}
		Expect(db.Create(otherClusterRegistry).Error).To(Succeed())
		httpResp, err = client.DefaultApi.ApiV1OrganizationsOrgIdClustersClusterIdImageRegistriesIdDelete(ctx, orgID, ownedClusterID, ownedRegistry.ID).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent))

		noRegStack := shared.CreateStackWithNamedBuildSource("owned-noreg", "owned-noreg-app", shared.BuildSourceRepoURL)
		thin, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameStacksPost(ctx, orgID, projectName).Stack(*noRegStack).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
		DeferCleanup(func() {
			shared.DeleteStack(client, orgID, projectName, thin.GetId())
		})
		_, httpResp, err = client.DefaultApi.ApplyStack(ctx, orgID, projectName, thin.GetId()).Stack(*noRegStack).Execute()
		Expect(err).To(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusBadRequest))
	})
})
