package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

type WorkspaceHandlerSpec struct {
	WorkspaceService              services.WorkspaceService
	WorkspaceResourceService      services.WorkspaceResourceService
	WorkspaceResourceBuildService services.ResourceBuildService
	Logger                        logger.Logger
}

type workspaceHandler struct {
	workspaceService              services.WorkspaceService
	workspaceResourceService      services.WorkspaceResourceService
	workspaceResourceBuildService services.ResourceBuildService
	logger                        logger.Logger
}

func NewWorkspaceHandler(spec WorkspaceHandlerSpec) *workspaceHandler {
	return &workspaceHandler{
		workspaceResourceService:      spec.WorkspaceResourceService,
		workspaceService:              spec.WorkspaceService,
		workspaceResourceBuildService: spec.WorkspaceResourceBuildService,
		logger:                        spec.Logger,
	}
}

func (h *workspaceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			_, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.workspaceService.GetWorkspace(ctx, id)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspace(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *workspaceHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, serr := h.workspaceService.GetWorkspacesByUserID(ctx, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspaceList(objs), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceHandler) ListByOrganisationID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			objs, serr := h.workspaceService.GetWorkspacesByOrganisationID(ctx, orgID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspaceList(objs), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Workspace
	cfg := &handlerConfig{
		&ws,
		validation.ValidateWorkspace(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertWorkspace(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID
			obj, serr := h.workspaceService.CreateWorkspace(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspace(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *workspaceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Workspace
	cfg := &handlerConfig{
		&ws,
		validation.ValidateWorkspace(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			convertedObject := presenters.ConvertWorkspace(&ws)
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID
			obj, serr := h.workspaceService.UpdateWorkspace(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspace(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *workspaceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			_, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			serr := h.workspaceService.DeleteWorkspace(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
