package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
)

type ClusterImageRegistryHandlerSpec struct {
	ClusterImageRegistryService services.ImageRegistryService
}

type clusterImageRegistryHandler struct {
	clusterImageRegistryService services.ImageRegistryService
}

func NewClusterImageRegistryHandler(spec ClusterImageRegistryHandlerSpec) *clusterImageRegistryHandler {
	return &clusterImageRegistryHandler{
		clusterImageRegistryService: spec.ClusterImageRegistryService,
	}
}

// GET /api/v1/organizations/{org_id}/clusters/{cluster_id}/registries
func (h *clusterImageRegistryHandler) ListRegistriesForCluster(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			clusterID := mux.Vars(r)["cluster_id"]
			orgID := mux.Vars(r)["org_id"]
			registries, serr := h.clusterImageRegistryService.ListByClusterID(ctx, orgID, clusterID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentClusterImageRegistryList(registries), nil
		},
	}
	handleGet(w, r, cfg)
}

// GET /api/v1/organizations/{org_id}/clusters/{cluster_id}/registries/{id}
func (h *clusterImageRegistryHandler) GetRegistry(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["cluster_id"]
			registryID := mux.Vars(r)["id"]
			registry, serr := h.clusterImageRegistryService.Get(ctx, registryID)
			if serr != nil {
				return nil, serr
			}
			if registry.OrganisationID != orgID || registry.ClusterID != clusterID {
				return nil, errors.NotFound("registry '%s' not found under organization '%s' and cluster '%s'", registryID, orgID, clusterID)
			}
			return presenters.PresentClusterImageRegistry(registry), nil
		},
	}
	handleGet(w, r, cfg)
}

// POST /api/v1/organizations/{org_id}/clusters/{cluster_id}/registries
func (h *clusterImageRegistryHandler) CreateRegistry(w http.ResponseWriter, r *http.Request) {
	var registry openapi.ClusterImageRegistry
	cfg := &handlerConfig{
		MarshalInto: &registry,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["cluster_id"]
			convertedRegistry := presenters.ConvertClusterImageRegistry(&registry)
			convertedRegistry.OrganisationID = orgID
			convertedRegistry.ClusterID = clusterID
			createdRegistry, serr := h.clusterImageRegistryService.Create(ctx, convertedRegistry)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentClusterImageRegistry(createdRegistry), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

// DELETE /api/v1/organizations/{org_id}/clusters/{cluster_id}/registries/{id}
func (h *clusterImageRegistryHandler) DeleteRegistry(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["cluster_id"]
			registryID := mux.Vars(r)["id"]
			// Get the registry first to check ownership
			registry, serr := h.clusterImageRegistryService.Get(ctx, registryID)
			if serr != nil {
				return nil, serr
			}
			if registry.OrganisationID != orgID || registry.ClusterID != clusterID {
				return nil, errors.NotFound("registry '%s' not found under organization '%s' and cluster '%s'", registryID, orgID, clusterID)
			}
			serr = h.clusterImageRegistryService.Delete(ctx, orgID, registryID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
