package int

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/test/int/shared"
)

var _ = Describe("Volume", func() {
	var client *openapi.APIClient
	var orgID string

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil(), "Test environment should be initialized")

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Team dependency check", func() {
		It("should allow deleting a team with no volumes (ListByTeamID ORDER BY created_at)", func() {
			By("Creating a non-default team")
			team := shared.CreateTeam(client, orgID, "vol-test-team")

			By("Deleting the team — exercises volumeStore.ListByTeamID which orders by created_at")
			shared.DeleteTeam(client, orgID, team.GetName())
		})
	})
})
