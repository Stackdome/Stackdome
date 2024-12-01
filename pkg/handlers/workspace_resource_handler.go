package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

// WorkspaceHandlerSpec defines the dependencies for the workspace handler
type WorkspaceResourceHandlerSpec struct {
	WorkspaceResourceService services.WorkspaceResourceService
	Logger                   logger.Logger
	WorkspaceService         services.WorkspaceService
}

type workspaceResourceHandler struct {
	workspaceResourceService services.WorkspaceResourceService
	logger                   logger.Logger
	workspaceService         services.WorkspaceService
}

func NewWorkspaceResourceHandler(spec WorkspaceResourceHandlerSpec) *workspaceResourceHandler {
	return &workspaceResourceHandler{
		workspaceResourceService: spec.WorkspaceResourceService,
		logger:                   spec.Logger,
		workspaceService:         spec.WorkspaceService,
	}
}

// GetByID fetches a workspace resource by its ID
func (h *workspaceResourceHandler) GetByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			workspaceID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
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
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, err := h.workspaceResourceService.GetByWorkspaceID(ctx, workspaceID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspaceResourceList(objs), nil
		},
	}
	handleList(w, r, cfg)
}
