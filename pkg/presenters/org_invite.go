package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"k8s.io/utils/ptr"
)

func PresentOrgInvite(invite *models.OrgInvite) openapi.OrgInvite {
	resp := openapi.OrgInvite{
		Id:             &invite.ID,
		Email:          &invite.Email,
		OrganisationId: &invite.OrganisationID,
		Role:           ptr.To(string(invite.TeamRole)),
		Status:         ptr.To(openapi.InviteStatus(invite.Status)),
		ExpiresAt:      &invite.ExpiresAt,
		EmailSent:      &invite.EmailSent,
		EmailError:     invite.EmailError,
		CreatedAt:      &invite.CreatedAt,
		AcceptedAt:     invite.AcceptedAt,
	}
	if invite.Team != nil {
		resp.TeamName = &invite.Team.Name
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
		Role:           ptr.To(string(invite.TeamRole)),
		Status:         ptr.To(openapi.InviteStatus(invite.Status)),
		ExpiresAt:      &invite.ExpiresAt,
		EmailSent:      &invite.EmailSent,
		InviteToken:    &rawToken,
		CreatedAt:      &invite.CreatedAt,
	}
	if invite.Team != nil {
		resp.TeamName = &invite.Team.Name
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
	if invite.Team != nil {
		resp.TeamName = &invite.Team.Name
	}
	if invite.InvitedBy != nil {
		resp.InviterName = &invite.InvitedBy.Name
	}
	return resp
}
