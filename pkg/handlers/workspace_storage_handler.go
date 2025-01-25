package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

func NewWorkspaceStorageHandler(spec WorkspaceStorageHandlerSpec) *workspaceStorageHandler {
	return &workspaceStorageHandler{
		workspaceStorageService: spec.WorkspaceStorageService,
		authzClient:             spec.AuthzClient,
	}
}

type WorkspaceStorageHandlerSpec struct {
	WorkspaceStorageService services.WorkspaceStorageService
	AuthzClient             auth.AuthorizationClient
}

type workspaceStorageHandler struct {
	workspaceStorageService services.WorkspaceStorageService
	workspaceVolumeService  services.VolumeService
	authzClient             auth.AuthorizationClient
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

			obj, err := h.workspaceStorageService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				id,
				obj.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to get workspace storage '%s'", currentUser.ID, id)
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
			orgID := mux.Vars(r)["org_id"]

			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				"",
				"",
				models.ResourceAccessModeList,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list workspace storages in org '%s'", currentUser.ID, orgID)
			}
			objs, serr := h.workspaceStorageService.ListByOrganisationID(ctx, orgID)
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
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				"",
				"",
				models.ResourceAccessModeList,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list volumes under workspace storage '%s'", currentUser.ID, id)
			}

			objs, serr := h.workspaceStorageService.ListVolumes(ctx, id, currentUser.ID)
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

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				"",
				currentUser.ID,
				models.ResourceAccessModeCreate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to create workspace storages in org '%s'", currentUser.ID, orgID)
			}
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

			storage, serr := h.workspaceStorageService.GetByID(ctx, storageID)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				storageID,
				storage.UserID,
				models.ResourceAccessModeUpdate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update workspace storage '%s'", currentUser.ID, storageID)
			}
			serr = h.workspaceStorageService.MarkAsSynced(ctx, currentUser.ID, storageID, volumeID)
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
			storage, serr := h.workspaceStorageService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				id,
				storage.UserID,
				models.ResourceAccessModeUpdate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update workspace storage '%s'", currentUser.ID, id)
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
			storage, serr := h.workspaceStorageService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceStorageResource,
				id,
				storage.UserID,
				models.ResourceAccessModeDelete,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete workspace storage '%s'", currentUser.ID, id)
			}
			serr = h.workspaceStorageService.Delete(ctx, id, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
