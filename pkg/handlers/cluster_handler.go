package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
)

type ClusterHandlerSpec struct {
	ClusterService services.ClusterService
}

type clusterHandler struct {
	clusterService services.ClusterService
}

func NewClusterHandler(spec ClusterHandlerSpec) *clusterHandler {
	return &clusterHandler{
		clusterService: spec.ClusterService,
	}
}

// GET /api/v1/organizations/{id}/clusters
func (h *clusterHandler) ListClustersForOrg(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			cluster, serr := h.clusterService.GetOwnedClusterForOrg(ctx, orgID)
			if serr != nil {
				return nil, serr
			}
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
			clusterID := mux.Vars(r)["id"]
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
