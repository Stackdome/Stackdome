package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

func NewWorkspaceStorageHandler(spec WorkspaceStorageHandlerSpec) *workspaceStorageHandler {
	return &workspaceStorageHandler{
		workspaceStorageService: spec.WorkspaceStorageService,
	}
}

type WorkspaceStorageHandlerSpec struct {
	WorkspaceStorageService services.WorkspaceStorageService
}

type workspaceStorageHandler struct {
	workspaceStorageService services.WorkspaceStorageService
	workspaceVolumeService  services.VolumeService
}

func (h *workspaceStorageHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			if id == "current" {
				currentUser, err := auth.GetCurrentUserFromCtx(ctx)
				if err != nil {
					return nil, errors.Unauthorized("failed to fetch current user")
				}
				objs, serr := h.workspaceStorageService.ListByUserID(ctx, currentUser.ID)
				if serr != nil {
					return nil, serr
				}
				return presenters.PresentWorkspaceStorageList(objs), nil
			}
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.workspaceStorageService.Get(ctx, id, currentUser.ID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspaceStorage(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *workspaceStorageHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, serr := h.workspaceStorageService.ListByUserID(ctx, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.WorkspaceStorageList{
				Items: presenters.PresentWorkspaceStorageList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceStorageHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, serr := h.workspaceStorageService.ListByOrganisationID(ctx, currentUser.OrganisationID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.WorkspaceStorageList{
				Items: presenters.PresentWorkspaceStorageList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceStorageHandler) ListVolumes(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			objs, serr := h.workspaceStorageService.ListVolumes(ctx, mux.Vars(r)["id"], currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.VolumeList{
				Items: presenters.PresentVolumeList(objs, true),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *workspaceStorageHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ws openapi.WorkspaceStorage
	cfg := &handlerConfig{
		&ws,
		validation.ValidateWorkspaceStorage(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertWorkspaceStorage(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID
			obj, serr := h.workspaceStorageService.Create(ctx, convertedObject, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspaceStorage(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *workspaceStorageHandler) MarkAsSynced(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			storageID := mux.Vars(r)["id"]
			volumeID := mux.Vars(r)["volume_id"]
			serr := h.workspaceStorageService.MarkAsSynced(ctx, currentUser.ID, storageID, volumeID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *workspaceStorageHandler) Update(w http.ResponseWriter, r *http.Request) {
	var ws openapi.WorkspaceStorage
	cfg := &handlerConfig{
		&ws,
		validation.ValidateWorkspaceStorage(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			convertedObject := presenters.ConvertWorkspaceStorage(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, serr := h.workspaceStorageService.Update(ctx, id, currentUser.ID, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspaceStorage(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *workspaceStorageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			serr := h.workspaceStorageService.Delete(ctx, id, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
