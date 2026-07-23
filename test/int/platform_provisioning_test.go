package int

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/slug"
	"github.com/Stackdome/stackdome/test/int/bootstrap"
	"github.com/Stackdome/stackdome/test/int/shared"
)

// shortOrgIDLength mirrors the unexported services.shortOrgIDLength used by
// signup seeding to derive the <slug>-<shortid> registry name.
const shortOrgIDLength = 8

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

		By("verifying the org was seeded a <slug>-<shortid> registry on the platform cluster")
		shortID := strings.ReplaceAll(newOrgID, "-", "")
		if len(shortID) > shortOrgIDLength {
			shortID = shortID[:shortOrgIDLength]
		}
		expectedRegistry := fmt.Sprintf("%s-%s", orgSlug, shortID)
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
		orgName := fmt.Sprintf("Owned Cluster Org %d", ts)
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

		By("attaching a custom domain to the organisation")
		ownedDomain := fmt.Sprintf("owned-%d.test", ts)
		orgResp, httpResp, err := client.DefaultApi.ApiV1OrganizationsIdGet(ctx, orgID).Execute()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
		domains := append(orgResp.GetDomains(), openapi.DomainName{Fqdn: &ownedDomain})
		orgResp.SetDomains(domains)
		updatedOrg, httpResp, err := client.DefaultApi.ApiV1OrganizationsIdPut(ctx, orgID).Organisation(*orgResp).Execute()
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
	})
})
