package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type WorkspaceResourceBuildHandlerSpec struct {
	WorkspaceResourceService     services.WorkspaceResourceService
	Logger                       logger.Logger
	WorkspaceService             services.WorkspaceService
	WorkspaceResouceBuildService services.ResourceBuildService
}

type workspaceResourceBuildHandler struct {
	workspaceResourceService     services.WorkspaceResourceService
	logger                       logger.Logger
	workspaceService             services.WorkspaceService
	workspaceResouceBuildService services.ResourceBuildService
}

func NewWorkspaceResourceBuildHandler(spec WorkspaceResourceBuildHandlerSpec) *workspaceResourceBuildHandler {
	return &workspaceResourceBuildHandler{
		workspaceResourceService:     spec.WorkspaceResourceService,
		logger:                       spec.Logger,
		workspaceService:             spec.WorkspaceService,
		workspaceResouceBuildService: spec.WorkspaceResouceBuildService,
	}
}

// List builds under a workspace resource
func (h *workspaceResourceBuildHandler) ListByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			workspaceID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, err := h.workspaceResouceBuildService.ListByResourceName(ctx, workspaceID, resourceName)
			if err != nil {
				return nil, err
			}
			listResp := openapi.WorkspaceResourceBuildList{
				Items: presenters.PresentWorkspaceResourceBuildList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceResourceBuildHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			buildID := mux.Vars(r)["build_id"]
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.workspaceResouceBuildService.GetByID(ctx, buildID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspaceResourceBuild(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *workspaceResourceBuildHandler) ListByWorkspaceID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			workspaceID := mux.Vars(r)["id"]
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, err := h.workspaceResouceBuildService.ListByWorkspaceID(ctx, workspaceID)
			if err != nil {
				return nil, err
			}
			listResp := openapi.WorkspaceResourceBuildList{
				Items: presenters.PresentWorkspaceResourceBuildList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}
