package int

import (
	"context"
	"fmt"
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

var _ = Describe("Default provisioning", func() {
	It("seeds a Default=true cluster and the platform base domain at boot", func() {
		ctx := context.Background()
		db := GetEnvironment().Database.GetSessionFactory().New(ctx)

		By("finding the platform organisation")
		var platformOrg models.Organisation
		Expect(db.Where("\"default\" = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())

		By("verifying the platform org owns a Default=true cluster")
		var defaultCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&defaultCluster).Error).To(Succeed())
		Expect(defaultCluster.Default).To(BeTrue())

		By("verifying the platform org has the base domain")
		var platformDomain models.OrganisationDomain
		Expect(db.Where(&models.OrganisationDomain{
			OrganisationID: platformOrg.ID,
			Domain:         bootstrap.DefaultProvisioningBaseDomain,
		}).First(&platformDomain).Error).To(Succeed())
	})

	It("seeds <slug>.<base> domain and <slug>-<shortid> registry on fresh signup, and stacks fall back to the default cluster", func() {
		ctx := context.Background()
		db := GetEnvironment().Database.GetSessionFactory().New(ctx)
		projectName := models.DefaultProjectName

		By("resolving the seeded default cluster")
		var platformOrg models.Organisation
		Expect(db.Where("\"default\" = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())
		var defaultCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&defaultCluster).Error).To(Succeed())

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
		expectedDomain := fmt.Sprintf("%s.%s", orgSlug, bootstrap.DefaultProvisioningBaseDomain)

		By("verifying the org was seeded a <slug>.<base> domain")
		var orgDomain models.OrganisationDomain
		Expect(db.Where(&models.OrganisationDomain{OrganisationID: newOrgID}).First(&orgDomain).Error).To(Succeed())
		Expect(orgDomain.Domain).To(Equal(expectedDomain))

		By("verifying the org was seeded a <slug>-<shortid> registry on the default cluster")
		shortID := strings.ReplaceAll(newOrgID, "-", "")
		if len(shortID) > shortOrgIDLength {
			shortID = shortID[:shortOrgIDLength]
		}
		expectedRegistry := fmt.Sprintf("%s-%s", orgSlug, shortID)
		var registry models.ClusterImageRegistry
		Expect(db.Where(&models.ClusterImageRegistry{OrganisationID: newOrgID}).First(&registry).Error).To(Succeed())
		Expect(registry.Name).To(Equal(expectedRegistry))
		Expect(registry.ClusterID).To(Equal(defaultCluster.ID))

		By("creating and deploying a stack with a publicly exposed port in the new org")
		newClient := shared.AuthenticatedClient(newToken)
		created, _ := shared.CreateStackAndDeploy(newClient, newOrgID, projectName, exposedProvisioningStack("dp-web"))
		stackID := created.GetId()
		Expect(stackID).NotTo(BeEmpty())

		DeferCleanup(func() {
			shared.DeleteStack(newClient, newOrgID, projectName, stackID)
			shared.WaitForStackDeleted(newClient, newOrgID, projectName, stackID, 1*time.Minute)
		})

		By("verifying the stack resolved to the default cluster via read-time fallback")
		var stackRow models.Stack
		Expect(db.First(&stackRow, "id = ?", stackID).Error).To(Succeed())
		Expect(stackRow.ClusterID).To(Equal(defaultCluster.ID))

		By("waiting for the stack to become Ready on the default cluster")
		shared.WaitForStackReady(newClient, newOrgID, projectName, stackID, 5*time.Minute)

		By("verifying the exposed-port stack domain uses <slug>.<base>")
		Eventually(func(g Gomega) {
			var stackDomain models.StackDomain
			g.Expect(db.Where(&models.StackDomain{StackID: stackID}).First(&stackDomain).Error).To(Succeed())
			g.Expect(stackDomain.Fqdn).To(HaveSuffix("." + expectedDomain))
		}, 2*time.Minute, 5*time.Second).Should(Succeed())
	})
})
