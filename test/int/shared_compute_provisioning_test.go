package int

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	cmv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/slug"
	"github.com/Stackdome/stackdome/test/int/bootstrap"
	"github.com/Stackdome/stackdome/test/int/shared"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
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

var _ = Describe("Shared-compute provisioning", func() {
	It("seeds a SharedCompute=true cluster and wildcard TLS at boot", func() {
		ctx := context.Background()
		env := GetEnvironment()
		db := env.Database.GetSessionFactory().New(ctx)

		By("finding the platform organisation")
		var platformOrg models.Organisation
		Expect(db.Where("platform = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())

		By("verifying the platform org owns a SharedCompute=true cluster")
		var sharedComputeCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&sharedComputeCluster).Error).To(Succeed())
		Expect(sharedComputeCluster.SharedCompute).To(BeTrue())

		By("verifying organisation domains remain custom-only")
		var domainCount int64
		Expect(db.Model(&models.OrganisationDomain{}).
			Where(&models.OrganisationDomain{OrganisationID: platformOrg.ID}).
			Count(&domainCount).Error).To(Succeed())
		Expect(domainCount).To(BeZero())

		By("verifying wildcard TLS resources were created without waiting for issuance")
		clusterClient := env.Cluster.GetClient()
		tlsNamespace := config.DefaultPlatformTLSNamespace

		cloudflareSecret := &corev1.Secret{}
		Expect(clusterClient.Get(ctx, types.NamespacedName{
			Namespace: tlsNamespace,
			Name:      models.CloudflareAPITokenSecretName,
		}, cloudflareSecret)).To(Succeed())
		Expect(cloudflareSecret.Data).To(HaveKey(models.CloudflareAPITokenSecretKey))

		issuer := &cmv1.Issuer{}
		Expect(clusterClient.Get(ctx, types.NamespacedName{
			Namespace: tlsNamespace,
			Name:      models.DNSIssuerName,
		}, issuer)).To(Succeed())

		certificate := &cmv1.Certificate{}
		Expect(clusterClient.Get(ctx, types.NamespacedName{
			Namespace: tlsNamespace,
			Name:      models.PlatformWildcardTLSName,
		}, certificate)).To(Succeed())
		Expect(certificate.Spec.SecretName).To(Equal(models.PlatformWildcardTLSName))
		Expect(certificate.Spec.DNSNames).To(Equal([]string{"*." + bootstrap.SharedComputeProvisioningBaseDomain}))
		Expect(certificate.Spec.IssuerRef.Name).To(Equal(models.DNSIssuerName))
		Expect(certificate.Spec.IssuerRef.Kind).To(Equal(cmv1.IssuerKind))

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

	It("keeps organisation domains custom-only, seeds the registry, and falls back to the shared-compute cluster", func() {
		ctx := context.Background()
		db := GetEnvironment().Database.GetSessionFactory().New(ctx)
		projectName := models.DefaultProjectName

		By("resolving the seeded shared-compute cluster")
		var platformOrg models.Organisation
		Expect(db.Where("platform = ?", true).Model(&models.Organisation{}).First(&platformOrg).Error).To(Succeed())
		var sharedComputeCluster models.Cluster
		Expect(db.Where(&models.Cluster{OrganisationID: platformOrg.ID}).First(&sharedComputeCluster).Error).To(Succeed())
		Expect(sharedComputeCluster.SharedCompute).To(BeTrue())

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

		By("verifying signup did not seed an organisation domain")
		var domainCount int64
		Expect(db.Model(&models.OrganisationDomain{}).
			Where(&models.OrganisationDomain{OrganisationID: newOrgID}).
			Count(&domainCount).Error).To(Succeed())
		Expect(domainCount).To(BeZero())

		By("verifying the org was seeded a <slug>-<shortOrgID>-<shortClusterID> registry on the shared-compute cluster")
		expectedRegistry := fmt.Sprintf("%s-%s-%s", orgSlug, shortTestID(newOrgID), shortTestID(sharedComputeCluster.ID))
		var registry models.ClusterImageRegistry
		Expect(db.Where(&models.ClusterImageRegistry{OrganisationID: newOrgID}).First(&registry).Error).To(Succeed())
		Expect(registry.Name).To(Equal(expectedRegistry))
		Expect(registry.ClusterID).To(Equal(sharedComputeCluster.ID))

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

		By("verifying the stack resolved to the shared-compute cluster via read-time fallback")
		var stackRow models.Stack
		Expect(db.First(&stackRow, "id = ?", stackID).Error).To(Succeed())
		Expect(stackRow.ClusterID).To(Equal(sharedComputeCluster.ID))

		By("waiting for the stack to become Ready on the shared-compute cluster")
		shared.WaitForStackReady(newClient, newOrgID, projectName, stackID, 5*time.Minute)

		By("verifying the exposed port received a deterministic platform hostname")
		Eventually(func(g Gomega) {
			var stackDomain models.StackDomain
			g.Expect(db.Where(&models.StackDomain{StackID: stackID}).First(&stackDomain).Error).To(Succeed())
			g.Expect(stackDomain.Fqdn).To(MatchRegexp(
				"^web-[a-f0-9]{8}\\." + regexp.QuoteMeta(bootstrap.SharedComputeProvisioningBaseDomain) + "$",
			))
		}, 2*time.Minute, 5*time.Second).Should(Succeed())
	})

	It("rejects tenant cluster registration when shared compute is enabled", func() {
		ctx := context.Background()
		env := GetEnvironment()
		db := env.Database.GetSessionFactory().New(ctx)

		By("signing up a fresh user with a new organisation")
		ts := time.Now().UnixNano()
		resp := shared.SignupNewUser(
			"Shared Compute Admin",
			fmt.Sprintf("shared-compute-%d@example.com", ts),
			"supersecret123",
			fmt.Sprintf("Shared Compute Org %d", ts),
		)
		orgID := resp.User.GetOrganisationId()
		client := shared.AuthenticatedClient(resp.GetJwtToken())

		By("attempting to register a valid tenant cluster via the public API")
		clusterURL, caData, saToken, err := bootstrap.ExtractAPIServerClusterCredentials(ctx, env.Cluster)
		Expect(err).NotTo(HaveOccurred())
		clusterReq := openapi.NewCluster(fmt.Sprintf("tenant-cluster-%d", ts), clusterURL, caData, saToken)
		createdCluster, httpResp, err := client.DefaultApi.
			ApiV1OrganizationsOrgIdClustersPost(ctx, orgID).
			Cluster(*clusterReq).
			Execute()

		Expect(createdCluster).To(BeNil())
		Expect(err).To(HaveOccurred())
		Expect(httpResp.StatusCode).To(Equal(http.StatusBadRequest))

		var apiErr *openapi.GenericOpenAPIError
		Expect(errors.As(err, &apiErr)).To(BeTrue(), "expected GenericOpenAPIError")
		var errorResponse openapi.Error
		Expect(json.Unmarshal(apiErr.Body(), &errorResponse)).To(Succeed())
		Expect(errorResponse.GetReason()).To(Equal("tenant clusters cannot be added when shared compute is enabled"))

		By("verifying the rejected request did not create a tenant-owned cluster")
		var tenantClusterCount int64
		Expect(db.Model(&models.Cluster{}).
			Where("organisation_id = ? AND shared_compute = ?", orgID, false).
			Count(&tenantClusterCount).Error).To(Succeed())
		Expect(tenantClusterCount).To(BeZero())
	})
})
