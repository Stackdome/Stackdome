package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentWorkspaceProvisionRequest(in *models.WorkspaceProvisionRequest) openapi.WorkspaceProvisionRequest {
	res := openapi.WorkspaceProvisionRequest{}
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)
	res.SetOrgId(int32(in.OrganisationID))
	res.SetId(in.ID)
	res.SetUserId(in.UserID)
	res.SetSshPublicKey(in.SshPublicKey)
	res.SetStatus(PresentWorkspaceProvisionRequestStatus(in.Status))
	res.SetState(string(in.State))
	res.SetMessage(in.Message)
	return res
}

func PresentWorkspaceProvisionRequestStatus(in *models.WorkspaceProvisionRequestStatus) openapi.WorkspaceProvisionRequestStatus {
	res := openapi.WorkspaceProvisionRequestStatus{}
	if in == nil {
		return res
	}

	if in.ClusterCACert != nil {
		res.SetClusterCaCert(*in.ClusterCACert)
	}
	if in.ClusterUrl != nil {
		res.SetClusterUrl(*in.ClusterUrl)
	}
	if in.WorkspaceNamespace != nil {
		res.SetWorkspaceNamespace(*in.WorkspaceNamespace)
	}
	if in.WorkspaceServiceAccountName != nil {
		res.SetWorkspaceServiceAccountname(*in.WorkspaceServiceAccountName)
	}
	if in.WorkspaceServiceAccountToken != nil {
		res.SetWorkspaceServiceaccountToken(*in.WorkspaceServiceAccountToken)
	}
	if in.Domain != nil {
		res.SetDomain(*in.Domain)
	}
	return res
}

func ConvertWorkspaceProvisionRequest(in *openapi.WorkspaceProvisionRequest) *models.WorkspaceProvisionRequest {
	res := &models.WorkspaceProvisionRequest{}
	res.SshPublicKey = in.GetSshPublicKey()
	if in.Status != nil {
		status := &models.WorkspaceProvisionRequestStatus{
			ClusterCACert:                in.Status.ClusterCaCert.Get(),
			ClusterUrl:                   in.Status.ClusterUrl.Get(),
			WorkspaceNamespace:           in.Status.WorkspaceNamespace.Get(),
			WorkspaceServiceAccountName:  in.Status.WorkspaceServiceAccountname.Get(),
			WorkspaceServiceAccountToken: in.Status.WorkspaceServiceaccountToken.Get(),
		}
		res.Status = status
	}
	return res
}
