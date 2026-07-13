package shared

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
)

const testServerPort = 8987

func CreateInvite(client *openapi.APIClient, orgID string, req openapi.OrgInviteCreateRequest) *openapi.OrgInviteCreateResponse {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesPost(ctx, orgID).OrgInviteCreateRequest(req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create invite")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
	Expect(resp).NotTo(BeNil())
	return resp
}

func CreateInviteExpectError(client *openapi.APIClient, orgID string, req openapi.OrgInviteCreateRequest, expectedStatus int) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesPost(ctx, orgID).OrgInviteCreateRequest(req).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func ListInvites(client *openapi.APIClient, orgID string) *openapi.OrgInviteList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesGet(ctx, orgID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list invites")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp).NotTo(BeNil())
	return resp
}

func GetInvite(client *openapi.APIClient, orgID, inviteID string) *openapi.OrgInvite {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdGet(ctx, orgID, inviteID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get invite")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp).NotTo(BeNil())
	return resp
}

func GetInviteExpectError(client *openapi.APIClient, orgID, inviteID string, expectedStatus int) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdGet(ctx, orgID, inviteID).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func RevokeInvite(client *openapi.APIClient, orgID, inviteID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdDelete(ctx, orgID, inviteID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to revoke invite")
	Expect(httpResp.StatusCode).To(Equal(http.StatusNoContent))
}

func RevokeInviteExpectError(client *openapi.APIClient, orgID, inviteID string, expectedStatus int) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdDelete(ctx, orgID, inviteID).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func ResendInvite(client *openapi.APIClient, orgID, inviteID string) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdResendPost(ctx, orgID, inviteID).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to resend invite")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
}

func ResendInviteExpectError(client *openapi.APIClient, orgID, inviteID string, expectedStatus int) {
	ctx := context.Background()
	httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdInvitesIdResendPost(ctx, orgID, inviteID).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func GetInviteInfo(client *openapi.APIClient, token string) *openapi.OrgInviteInfo {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1InvitesTokenInfoGet(ctx, token).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to get invite info")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp).NotTo(BeNil())
	return resp
}

func GetInviteInfoExpectError(client *openapi.APIClient, token string, expectedStatus int) {
	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1InvitesTokenInfoGet(ctx, token).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func SignupWithInvite(inviteToken, name, email, password string) *openapi.UserSignupResponse {
	client := UnauthenticatedClient()
	req := openapi.NewUserSignupRequest(name, email, password)
	req.SetInviteToken(inviteToken)

	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to signup with invite")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
	Expect(resp).NotTo(BeNil())
	return resp
}

func SignupWithInviteExpectError(inviteToken, name, email, password string, expectedStatus int) {
	client := UnauthenticatedClient()
	req := openapi.NewUserSignupRequest(name, email, password)
	req.SetInviteToken(inviteToken)

	ctx := context.Background()
	_, httpResp, err := client.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
	Expect(err).To(HaveOccurred())
	Expect(httpResp.StatusCode).To(Equal(expectedStatus))
}

func SignupNewUser(name, email, password, orgName string) *openapi.UserSignupResponse {
	client := UnauthenticatedClient()
	org := openapi.NewOrganisation()
	org.SetName(orgName)
	req := openapi.NewUserSignupRequest(name, email, password)
	req.SetOrganisation(*org)

	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1UserSignupPost(ctx).UserSignupRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to signup")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
	Expect(resp).NotTo(BeNil())
	return resp
}

func NewInviteCreateRequest(email, projectName, role string, expiresInDays int) openapi.OrgInviteCreateRequest {
	return *openapi.NewOrgInviteCreateRequest(email, projectName, role, int32(expiresInDays))
}

func CreateProject(client *openapi.APIClient, orgID, projectName string) *openapi.Project {
	ctx := context.Background()
	req := openapi.NewProjectCreateRequest(projectName)
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdProjectsPost(ctx, orgID).ProjectCreateRequest(*req).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to create project")
	Expect(httpResp.StatusCode).To(Equal(http.StatusCreated))
	Expect(resp).NotTo(BeNil())
	return resp
}

func ListProjectMembers(client *openapi.APIClient, orgID, projectName string) *openapi.ProjectMembershipList {
	ctx := context.Background()
	resp, httpResp, err := client.DefaultApi.ApiV1OrganizationsOrgIdProjectsProjectNameMembersGet(ctx, orgID, projectName).Execute()
	Expect(err).NotTo(HaveOccurred(), "failed to list project members")
	Expect(httpResp.StatusCode).To(Equal(http.StatusOK))
	Expect(resp).NotTo(BeNil())
	return resp
}

func UnauthenticatedClient() *openapi.APIClient {
	config := openapi.NewConfiguration()
	config.Host = fmt.Sprintf("localhost:%d", testServerPort)
	config.Scheme = "http"
	return openapi.NewAPIClient(config)
}

func AuthenticatedClient(token string) *openapi.APIClient {
	config := openapi.NewConfiguration()
	config.Host = fmt.Sprintf("localhost:%d", testServerPort)
	config.Scheme = "http"
	config.DefaultHeader = map[string]string{
		"Authorization": "Bearer " + token,
	}
	return openapi.NewAPIClient(config)
}
