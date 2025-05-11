package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

type ClusterHandlerSpec struct {
	ClusterService services.ClusterService
	AuthzClient    auth.AuthorizationClient
}

type clusterHandler struct {
	clusterService services.ClusterService
	authzClient    auth.AuthorizationClient
}

func NewClusterHandler(spec ClusterHandlerSpec) *clusterHandler {
	return &clusterHandler{
		clusterService: spec.ClusterService,
		authzClient:    spec.AuthzClient,
	}
}

// GET /api/v1/organizations/{id}/clusters
func (h *clusterHandler) ListClustersForOrg(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Organisation,
				orgID,
				"",
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize organisation access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access organisation '%s'", currentUser.ID, orgID)
			}
			cluster, serr := h.clusterService.GetClusterForOrg(ctx, orgID)
			if serr != nil {
				return nil, serr
			}
			// Return as a list
			return presenters.PresentClusterList([]*models.Cluster{cluster}), nil
		},
	}
	handleGet(w, r, cfg)
}

// Delete cluster
// DELETE /api/v1/organizations/{org_id}/clusters/{id}
func (h *clusterHandler) DeleteClusterForOrg(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Cluster,
				orgID,
				"",
				models.ResourceAccessModeWrite,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize cluster delete access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete cluster '%s'", currentUser.ID, orgID)
			}
			if orgID != currentUser.OrganisationID && currentUser.Role != models.PlatformAdminRole {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete cluster '%s'", currentUser.ID, orgID)
			}
			serr := h.clusterService.Delete(ctx, clusterID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

// Add cluster
// POST /api/v1/organizations/{org_id}/clusters
func (h *clusterHandler) AddClusterForOrg(w http.ResponseWriter, r *http.Request) {
	var cluster openapi.Cluster
	cfg := &handlerConfig{
		MarshalInto: &cluster,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Cluster,
				orgID,
				"",
				models.ResourceAccessModeWrite,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize cluster add access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to add cluster '%s'", currentUser.ID, orgID)
			}
			if orgID != currentUser.OrganisationID && currentUser.Role != models.PlatformAdminRole {
				return nil, errors.Unauthorized("user '%s' is not allowed to add cluster '%s'", currentUser.ID, orgID)
			}
			convertedCluster := presenters.ConvertCluster(&cluster)
			convertedCluster.OrganisationID = orgID
			cluster, serr := h.clusterService.AddCluster(ctx, convertedCluster)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentCluster(cluster), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

// GET /api/v1/organizations/{org_id}/clusters/{id}
func (h *clusterHandler) GetClusterForOrg(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Organisation,
				orgID,
				"",
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize organisation access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access organisation '%s'", currentUser.ID, orgID)
			}
			cluster, serr := h.clusterService.Get(ctx, clusterID)
			if serr != nil {
				return nil, serr
			}
			if cluster.OrganisationID != orgID {
				return nil, errors.NotFound("cluster '%s' not found under organisation '%s'", clusterID, orgID)
			}
			return presenters.PresentCluster(cluster), nil
		},
	}
	handleGet(w, r, cfg)
}
