package presenters

import (
	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func PresentWorkspaceUser(in *models.WorkspaceUser) openapi.WorkspaceUser {
	res := openapi.WorkspaceUser{}
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)
	res.SetOrgId(in.OrganisationID)
	res.SetTeamId(in.TeamID)
	res.SetId(in.ID)
	res.SetUserId(in.UserID)
	res.SetVersion(int32(in.Version))
	res.SetWorkspaces(presentUserWorkspaceNamespaces(in.WorkspaceNamespaces))
	res.SetStatus(PresentWorkspaceUserStatus(in.Status, in))
	res.SetState(PresentWorkspaceUserState(in.Status))
	if in.Status != nil {
		res.SetMessage(in.Status.Message)
	}
	return res
}

func presentUserWorkspaceNamespaces(in []*models.WorkspaceNamespace) []string {
	res := make([]string, 0)
	for _, v := range in {
		if v.Enabled {
			res = append(res, v.Workspace)
		}
	}
	return res
}

func PresentWorkspaceUserState(in *models.WorkspaceUserStatus) openapi.WorkspaceUserState {
	if in == nil {
		return openapi.PENDING
	}
	switch in.State {
	case models.WorkspaceUserProvisionCompleted:
		return openapi.COMPLETED
	case models.WorkspaceUserProvisionPending:
		return openapi.PENDING
	case models.WorkspaceUserProvisionError:
		return openapi.ERROR
	default:
		return openapi.PENDING
	}
}

func PresentWorkspaceUserStatus(in *models.WorkspaceUserStatus, wu *models.WorkspaceUser) openapi.WorkspaceUserStatus {
	res := openapi.WorkspaceUserStatus{}
	if in == nil {
		return res
	}
	res.SetClusterCaCert(in.ClusterCACert)
	res.SetClusterUrl(in.ClusterUrl)
	res.SetServiceAccountName(in.ServiceAccountName)
	res.SetServiceaccountToken(in.ServiceAccountToken)
	res.SetObservedVersion(int32(in.ObservedVersion))
	res.SetProvisionedNamespaces(presentProvisionedNamespaces(in.ProvisionedNamespaces, wu.WorkspaceNamespaces))
	res.SetConditions(presentConditions(in.Conditions))
	return res
}

func presentProvisionedNamespaces(provisionedNamespaces []string, workspaceNamespaces []*models.WorkspaceNamespace) []openapi.WorkspaceUserStatusProvisionedNamespacesInner {
	index := make(map[string]*models.WorkspaceNamespace)
	for _, v := range workspaceNamespaces {
		index[v.Namespace] = v
	}
	res := make([]openapi.WorkspaceUserStatusProvisionedNamespacesInner, 0)
	for _, v := range provisionedNamespaces {
		workspace := index[v]
		if workspace.Enabled {
			res = append(res, openapi.WorkspaceUserStatusProvisionedNamespacesInner{
				Namespace:     &workspace.Namespace,
				WorkspaceName: &workspace.Workspace,
			})
		}
	}
	return res
}

func ConvertWorkspaceUser(in *openapi.WorkspaceUser) *models.WorkspaceUser {
	res := &models.WorkspaceUser{
		Status: &models.WorkspaceUserStatus{},
	}
	res.WorkspaceNamespaces = convertWorkspaceNames(in.GetWorkspaces())
	res.Version = 1
	return res
}

func convertWorkspaceNames(in []string) []*models.WorkspaceNamespace {
	res := make([]*models.WorkspaceNamespace, 0)
	for _, v := range in {
		res = append(res, &models.WorkspaceNamespace{
			Workspace: v,
		})
	}
	return res
}
