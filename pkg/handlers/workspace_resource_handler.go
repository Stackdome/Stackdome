package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

// WorkspaceHandlerSpec defines the dependencies for the workspace handler
type WorkspaceResourceHandlerSpec struct {
	WorkspaceResourceService services.WorkspaceResourceService
	Logger                   logger.Logger
	WorkspaceService         services.WorkspaceService
	AuthzClient              auth.AuthorizationClient
}

type workspaceResourceHandler struct {
	workspaceResourceService services.WorkspaceResourceService
	logger                   logger.Logger
	workspaceService         services.WorkspaceService
	authzClient              auth.AuthorizationClient
}

func NewWorkspaceResourceHandler(spec WorkspaceResourceHandlerSpec) *workspaceResourceHandler {
	return &workspaceResourceHandler{
		workspaceResourceService: spec.WorkspaceResourceService,
		logger:                   spec.Logger,
		workspaceService:         spec.WorkspaceService,
		authzClient:              spec.AuthzClient,
	}
}

// GetByID fetches a workspace resource by its ID
func (h *workspaceResourceHandler) GetByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			workspaceID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			workspace, serr := h.workspaceService.GetWorkspace(ctx, workspaceID)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceResource,
				workspaceID,
				workspace.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to get workspace resource '%s' builds", currentUser.ID, resourceName)
			}
			obj, err := h.workspaceResourceService.GetByWorkspaceIDAndResourceName(ctx, workspaceID, resourceName)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspaceResource(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

// List fetches all workspace resources
func (h *workspaceResourceHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			workspaceID := mux.Vars(r)["id"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			workspace, serr := h.workspaceService.GetWorkspace(ctx, workspaceID)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceResource,
				workspaceID,
				workspace.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list workspace '%s' resources", currentUser.ID, workspaceID)
			}
			objs, err := h.workspaceResourceService.GetByWorkspaceID(ctx, workspaceID)
			if err != nil {
				return nil, err
			}

			listResp := openapi.WorkspaceResourceList{
				Items: presenters.PresentWorkspaceResourceList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}
