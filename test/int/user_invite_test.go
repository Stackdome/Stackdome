package int

import (
	"context"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/test/int/shared"
)

var _ = Describe("User Invites", func() {
	var client *openapi.APIClient
	var orgID string
	var teamName = models.DefaultTeamName

	BeforeEach(func() {
		testEnv := GetEnvironment()
		Expect(testEnv).NotTo(BeNil())

		client = testEnv.Client
		orgID = testEnv.OrgID
	})

	Context("Create Invite", func() {
		It("should create an invite and return the token", func() {
			By("Creating an invite for a new email")
			req := shared.NewInviteCreateRequest("newuser@example.com", teamName, "Developer", 7)

			invite := shared.CreateInvite(client, orgID, req)

			Expect(invite.GetId()).NotTo(BeEmpty())
			Expect(invite.GetEmail()).To(Equal("newuser@example.com"))
			Expect(invite.GetTeamName()).To(Equal(teamName))
			Expect(invite.GetRole()).To(Equal("Developer"))
			Expect(invite.GetStatus()).To(Equal(openapi.INVITE_PENDING))
			Expect(invite.GetInviteToken()).NotTo(BeEmpty())
			Expect(invite.GetEmailSent()).To(BeFalse())
		})

		It("should create an invite with Viewer role", func() {
			req := shared.NewInviteCreateRequest("viewer@example.com", teamName, "Viewer", 14)
			invite := shared.CreateInvite(client, orgID, req)

			Expect(invite.GetRole()).To(Equal("Viewer"))
		})

		It("should reject invite with invalid role", func() {
			req := shared.NewInviteCreateRequest("bad-role@example.com", teamName, "Admin", 7)
			shared.CreateInviteExpectError(client, orgID, req, http.StatusBadRequest)
		})

		It("should reject invite with invalid expiry", func() {
			req := shared.NewInviteCreateRequest("bad-expiry@example.com", teamName, "Developer", 60)
			shared.CreateInviteExpectError(client, orgID, req, http.StatusBadRequest)
		})

		It("should reject invite for non-existent team", func() {
			req := shared.NewInviteCreateRequest("no-team@example.com", "nonexistent-team", "Developer", 7)
			shared.CreateInviteExpectError(client, orgID, req, http.StatusBadRequest)
		})

		It("should reject invite for email that already has an account", func() {
			email := fmt.Sprintf("existing-%d@example.com", time.Now().UnixNano())
			shared.SignupNewUser("Existing User", email, "testpassword123", fmt.Sprintf("existing-org-%d", time.Now().UnixNano()))

			req := shared.NewInviteCreateRequest(email, teamName, "Developer", 7)
			shared.CreateInviteExpectError(client, orgID, req, http.StatusConflict)
		})

		It("should reject duplicate pending invite for same email+org", func() {
			req := shared.NewInviteCreateRequest("duplicate@example.com", teamName, "Developer", 7)
			shared.CreateInvite(client, orgID, req)

			shared.CreateInviteExpectError(client, orgID, req, http.StatusConflict)
		})

		It("should reject invite with invalid email format", func() {
			req := shared.NewInviteCreateRequest("notanemail", teamName, "Developer", 7)
			shared.CreateInviteExpectError(client, orgID, req, http.StatusBadRequest)
		})
	})

	Context("List Invites", func() {
		It("should list all invites for the organization", func() {
			By("Creating multiple invites")
			shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("list1@example.com", teamName, "Developer", 7))
			shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("list2@example.com", teamName, "Viewer", 7))

			By("Listing all invites")
			list := shared.ListInvites(client, orgID)

			Expect(list.GetTotal()).To(BeNumerically("==", int32(2)))

			var emails []string
			for _, inv := range list.GetItems() {
				emails = append(emails, inv.GetEmail())
			}
			Expect(emails).To(ContainElement("list1@example.com"))
			Expect(emails).To(ContainElement("list2@example.com"))
		})
	})

	Context("Get Invite", func() {
		It("should retrieve an invite by ID", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("getme@example.com", teamName, "Developer", 7))

			retrieved := shared.GetInvite(client, orgID, created.GetId())

			Expect(retrieved.GetId()).To(Equal(created.GetId()))
			Expect(retrieved.GetEmail()).To(Equal("getme@example.com"))
		})

		It("should return 404 for non-existent invite", func() {
			shared.GetInviteExpectError(client, orgID, "non-existent-id", http.StatusNotFound)
		})
	})

	Context("Revoke Invite", func() {
		It("should revoke a pending invite", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("revokeme@example.com", teamName, "Developer", 7))

			shared.RevokeInvite(client, orgID, created.GetId())

			By("Verifying the invite is revoked")
			retrieved := shared.GetInvite(client, orgID, created.GetId())
			Expect(retrieved.GetStatus()).To(Equal(openapi.INVITE_REVOKED))
		})

		It("should not revoke an already revoked invite", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("double-revoke@example.com", teamName, "Developer", 7))
			shared.RevokeInvite(client, orgID, created.GetId())

			shared.RevokeInviteExpectError(client, orgID, created.GetId(), http.StatusBadRequest)
		})
	})

	Context("Resend Invite", func() {
		It("should resend a pending invite", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("resendme@example.com", teamName, "Developer", 7))
			shared.ResendInvite(client, orgID, created.GetId())
		})

		It("should not resend a revoked invite", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("resend-revoked@example.com", teamName, "Developer", 7))
			shared.RevokeInvite(client, orgID, created.GetId())

			shared.ResendInviteExpectError(client, orgID, created.GetId(), http.StatusBadRequest)
		})
	})

	Context("Async Email Delivery", func() {
		It("should mark email_sent true after worker processes the invite", func() {
			By("Creating an invite")
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("email-worker@example.com", teamName, "Developer", 7))
			Expect(invite.GetEmailSent()).To(BeFalse())

			By("Waiting for the worker to send the email")
			Eventually(func() bool {
				retrieved := shared.GetInvite(client, orgID, invite.GetId())
				return retrieved.GetEmailSent()
			}, 60*time.Second, 2*time.Second).Should(BeTrue())
		})
	})

	Context("Public Invite Info", func() {
		It("should return invite info without authentication", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("info@example.com", teamName, "Developer", 7))

			unauthClient := shared.UnauthenticatedClient()
			info := shared.GetInviteInfo(unauthClient, created.GetInviteToken())

			Expect(info.GetTeamName()).To(Equal(teamName))
			Expect(info.GetInviterName()).NotTo(BeEmpty())
			Expect(info.ExpiresAt).NotTo(BeNil())
		})

		It("should return 404 for invalid token", func() {
			unauthClient := shared.UnauthenticatedClient()
			shared.GetInviteInfoExpectError(unauthClient, "invalid_token_here", http.StatusNotFound)
		})

		It("should return 410 for revoked invite token", func() {
			created := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("info-revoked@example.com", teamName, "Developer", 7))
			shared.RevokeInvite(client, orgID, created.GetId())

			unauthClient := shared.UnauthenticatedClient()
			shared.GetInviteInfoExpectError(unauthClient, created.GetInviteToken(), http.StatusGone)
		})
	})

	Context("Invite-Based Signup", func() {
		It("should sign up a new user with an invite token", func() {
			By("Creating an invite")
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("invited-user@example.com", teamName, "Developer", 7))

			By("Signing up with the invite token")
			signupResp := shared.SignupWithInvite(invite.GetInviteToken(), "Invited User", "invited-user@example.com", "securepassword123")

			Expect(signupResp.GetJwtToken()).NotTo(BeEmpty())
			Expect(signupResp.User.GetOrganisationId()).To(Equal(orgID))

			By("Verifying the invite is marked as accepted")
			retrieved := shared.GetInvite(client, orgID, invite.GetId())
			Expect(retrieved.GetStatus()).To(Equal(openapi.INVITE_ACCEPTED))
			Expect(retrieved.AcceptedAt).NotTo(BeNil())
		})

		It("should reject signup with mismatched email", func() {
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("correct@example.com", teamName, "Developer", 7))
			shared.SignupWithInviteExpectError(invite.GetInviteToken(), "Wrong Email", "wrong@example.com", "securepassword123", http.StatusBadRequest)
		})

		It("should reject signup with revoked invite", func() {
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("revoked-signup@example.com", teamName, "Developer", 7))
			shared.RevokeInvite(client, orgID, invite.GetId())

			shared.SignupWithInviteExpectError(invite.GetInviteToken(), "Revoked User", "revoked-signup@example.com", "securepassword123", http.StatusGone)
		})

		It("should reject signup with already-used invite", func() {
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("used-once@example.com", teamName, "Developer", 7))

			By("First signup should succeed")
			shared.SignupWithInvite(invite.GetInviteToken(), "First User", "used-once@example.com", "securepassword123")

			By("Second signup with same token should fail")
			shared.SignupWithInviteExpectError(invite.GetInviteToken(), "Second User", "second-user@example.com", "securepassword123", http.StatusGone)
		})

		It("should reject signup with invalid token", func() {
			shared.SignupWithInviteExpectError("totally_bogus_token_value", "Bad Token User", "badtoken@example.com", "securepassword123", http.StatusNotFound)
		})

		It("should reject signup without organisation and without invite token", func() {
			unauthClient := shared.UnauthenticatedClient()
			req := openapi.NewUserSignupRequest("No Org User", "noorg@example.com", "securepassword123")

			ctx := context.Background()
			_, httpResp, err := unauthClient.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
			Expect(err).To(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Context("Invited User Access", func() {
		It("should allow invited user to access their team resources", func() {
			By("Creating an invite")
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("access-test@example.com", teamName, "Developer", 7))

			By("Signing up with the invite")
			signupResp := shared.SignupWithInvite(invite.GetInviteToken(), "Access Test User", "access-test@example.com", "securepassword123")

			By("Verifying the invited user can list their teams")
			invitedUserClient := shared.AuthenticatedClient(signupResp.GetJwtToken())
			ctx := context.Background()
			teamList, httpResp, err := invitedUserClient.DefaultApi.ApiV1UsersCurrentTeamsGet(ctx).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			Expect(teamList.GetItems()).NotTo(BeEmpty())
		})
	})

	Context("Cross-Team Invite Flow", func() {
		It("should invite a user to a custom team with correct role and access", func() {
			customTeamName := fmt.Sprintf("backend-team-%d", time.Now().UnixNano())

			By("Org admin creates a new team")
			team := shared.CreateTeam(client, orgID, customTeamName)
			Expect(team.GetName()).To(Equal(customTeamName))

			By("Org admin invites a user to the custom team as Developer")
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("dev@example.com", customTeamName, "Developer", 7))
			Expect(invite.GetTeamName()).To(Equal(customTeamName))
			Expect(invite.GetRole()).To(Equal("Developer"))

			By("Invited user signs up with the invite token")
			signupResp := shared.SignupWithInvite(invite.GetInviteToken(), "Dev User", "dev@example.com", "securepassword123")
			Expect(signupResp.GetJwtToken()).NotTo(BeEmpty())
			Expect(signupResp.User.GetOrganisationId()).To(Equal(orgID))

			invitedClient := shared.AuthenticatedClient(signupResp.GetJwtToken())
			ctx := context.Background()

			By("Invited user appears in the custom team's member list with Developer role")
			members := shared.ListTeamMembers(client, orgID, customTeamName)
			var foundDeveloper bool
			for _, m := range members.GetItems() {
				if m.GetRole() == "Developer" {
					foundDeveloper = true
				}
			}
			Expect(foundDeveloper).To(BeTrue(), "invited user should appear in custom team with Developer role")

			By("Invited user is NOT in the default team")
			defaultMembers := shared.ListTeamMembers(client, orgID, models.DefaultTeamName)
			for _, m := range defaultMembers.GetItems() {
				Expect(m.GetUserId()).NotTo(Equal(signupResp.User.GetId()), "invited user should NOT be in the default team")
			}

			By("Invited user can list org teams (proves OrgMember Casbin grouping)")
			orgTeams, httpResp, err := invitedClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsGet(ctx, orgID).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
			Expect(orgTeams.GetItems()).NotTo(BeEmpty())

			By("Invited user can access team-scoped resources (list stacks for custom team)")
			_, httpResp, err = invitedClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameStacksGet(ctx, orgID, customTeamName).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))

			By("Invited user can list secrets in their team")
			_, httpResp, err = invitedClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsTeamNameSecretsGet(ctx, orgID, customTeamName).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Context("Invited Viewer Access", func() {
		It("should invite a user as Viewer to a custom team with correct role", func() {
			viewerTeamName := fmt.Sprintf("viewer-team-%d", time.Now().UnixNano())

			By("Org admin creates a team")
			shared.CreateTeam(client, orgID, viewerTeamName)

			By("Org admin invites a user as Viewer")
			invite := shared.CreateInvite(client, orgID, shared.NewInviteCreateRequest("viewer-user@example.com", viewerTeamName, "Viewer", 7))

			By("Viewer signs up")
			signupResp := shared.SignupWithInvite(invite.GetInviteToken(), "Viewer User", "viewer-user@example.com", "securepassword123")

			By("Viewer appears in team member list with Viewer role")
			members := shared.ListTeamMembers(client, orgID, viewerTeamName)
			var foundViewer bool
			for _, m := range members.GetItems() {
				if m.GetRole() == "Viewer" {
					foundViewer = true
				}
			}
			Expect(foundViewer).To(BeTrue(), "user should have Viewer role in the team")

			By("Viewer can access org-level teams (OrgMember Casbin grouping)")
			viewerClient := shared.AuthenticatedClient(signupResp.GetJwtToken())
			ctx := context.Background()
			_, httpResp, err := viewerClient.DefaultApi.ApiV1OrganizationsOrgIdTeamsGet(ctx, orgID).Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
		})
	})
})
