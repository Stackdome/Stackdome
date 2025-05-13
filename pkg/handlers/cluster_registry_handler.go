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

type ClusterImageRegistryHandlerSpec struct {
	ClusterImageRegistryService services.ClusterImageRegistryService
	AuthzClient                 auth.AuthorizationClient
}

type clusterImageRegistryHandler struct {
	clusterImageRegistryService services.ClusterImageRegistryService
	authzClient                 auth.AuthorizationClient
}

func NewClusterImageRegistryHandler(spec ClusterImageRegistryHandlerSpec) *clusterImageRegistryHandler {
	return &clusterImageRegistryHandler{
		clusterImageRegistryService: spec.ClusterImageRegistryService,
		authzClient:                 spec.AuthzClient,
	}
}

// GET /api/v1/organizations/{org_id}/clusters/{cluster_id}/registries
func (h *clusterImageRegistryHandler) ListRegistriesForCluster(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			clusterID := mux.Vars(r)["cluster_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Cluster,
				orgID,
				currentUser.ID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize cluster access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access cluster '%s'", currentUser.ID, clusterID)
			}
			registries, serr := h.clusterImageRegistryService.ListByClusterID(ctx, clusterID)
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
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Cluster,
				orgID,
				currentUser.ID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize cluster access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access cluster '%s'", currentUser.ID, clusterID)
			}
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
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Cluster,
				orgID,
				currentUser.ID,
				models.ResourceAccessModeWrite,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize cluster access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to modify cluster '%s'", currentUser.ID, clusterID)
			}
			if orgID != currentUser.OrganisationID && currentUser.Role != models.PlatformAdminRole {
				return nil, errors.Unauthorized("user '%s' is not allowed to modify cluster '%s'", currentUser.ID, clusterID)
			}
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
				return nil, errors.Unauthorized("failed to authorize cluster access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to modify cluster '%s'", currentUser.ID, clusterID)
			}
			if orgID != currentUser.OrganisationID && currentUser.Role != models.PlatformAdminRole {
				return nil, errors.Unauthorized("user '%s' is not allowed to modify cluster '%s'", currentUser.ID, clusterID)
			}
			// Get the registry first to check ownership
			registry, serr := h.clusterImageRegistryService.Get(ctx, registryID)
			if serr != nil {
				return nil, serr
			}
			if registry.OrganisationID != orgID || registry.ClusterID != clusterID {
				return nil, errors.NotFound("registry '%s' not found under organization '%s' and cluster '%s'", registryID, orgID, clusterID)
			}
			serr = h.clusterImageRegistryService.Delete(ctx, registryID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
