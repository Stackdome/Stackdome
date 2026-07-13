package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stores"
	"k8s.io/utils/ptr"
)

func PresentOrgInvite(invite *models.OrgInvite) openapi.OrgInvite {
	resp := openapi.OrgInvite{
		Id:             &invite.ID,
		Email:          &invite.Email,
		OrganisationId: &invite.OrganisationID,
		Role:           ptr.To(string(invite.ProjectRole)),
		Status:         ptr.To(openapi.InviteStatus(invite.Status)),
		ExpiresAt:      &invite.ExpiresAt,
		EmailSent:      &invite.EmailSent,
		EmailError:     invite.EmailError,
		CreatedAt:      &invite.CreatedAt,
		AcceptedAt:     invite.AcceptedAt,
	}
	if invite.Project != nil {
		resp.ProjectName = &invite.Project.Name
	}
	if invite.InvitedBy != nil {
		resp.InvitedBy = &invite.InvitedBy.Name
	}
	return resp
}

func PresentOrgInviteCreateResponse(invite *models.OrgInvite, rawToken string) openapi.OrgInviteCreateResponse {
	resp := openapi.OrgInviteCreateResponse{
		Id:             &invite.ID,
		Email:          &invite.Email,
		OrganisationId: &invite.OrganisationID,
		Role:           ptr.To(string(invite.ProjectRole)),
		Status:         ptr.To(openapi.InviteStatus(invite.Status)),
		ExpiresAt:      &invite.ExpiresAt,
		EmailSent:      &invite.EmailSent,
		InviteToken:    &rawToken,
		CreatedAt:      &invite.CreatedAt,
	}
	if invite.Project != nil {
		resp.ProjectName = &invite.Project.Name
	}
	if invite.InvitedBy != nil {
		resp.InvitedBy = &invite.InvitedBy.Name
	}
	return resp
}

func PresentOrgInviteList(result *stores.PaginatedResult[*models.OrgInvite]) openapi.OrgInviteList {
	items := make([]openapi.OrgInvite, len(result.Items))
	for i, invite := range result.Items {
		items[i] = PresentOrgInvite(invite)
	}
	total := int32(result.Total)
	return openapi.OrgInviteList{
		Items: items,
		Total: &total,
	}
}

func PresentOrgInviteInfo(invite *models.OrgInvite) openapi.OrgInviteInfo {
	resp := openapi.OrgInviteInfo{
		ExpiresAt: &invite.ExpiresAt,
	}
	if invite.Organisation != nil {
		resp.OrgName = &invite.Organisation.Name
	}
	if invite.Project != nil {
		resp.ProjectName = &invite.Project.Name
	}
	if invite.InvitedBy != nil {
		resp.InviterName = &invite.InvitedBy.Name
	}
	return resp
}
