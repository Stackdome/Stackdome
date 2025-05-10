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

func NewVolumeHandler(spec VolumeHandlerSpec) *volumeHandler {
	return &volumeHandler{
		volumeService: spec.VolumeService,
		authzClient:   spec.AuthzClient,
	}
}

type VolumeHandlerSpec struct {
	VolumeService services.VolumeService
	AuthzClient   auth.AuthorizationClient
}

type volumeHandler struct {
	volumeService services.VolumeService
	authzClient   auth.AuthorizationClient
}

func (h *volumeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			if id == "current" {
				currentUser, err := auth.GetCurrentUserFromCtx(ctx)
				if err != nil {
					return nil, errors.Unauthorized("failed to fetch current user")
				}
				objs, serr := h.volumeService.ListByUserID(ctx, currentUser.ID)
				if serr != nil {
					return nil, serr
				}
				listResp := openapi.VolumeList{
					Items: presenters.PresentVolumeList(objs, true),
					Total: ptr.To(int32(len(objs))),
				}
				return listResp, nil
			}
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.volumeService.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Volume,
				id,
				obj.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to get volume '%s'", currentUser.ID, id)
			}
			return presenters.PresentVolume(obj, true), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *volumeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Volume
	cfg := &handlerConfig{
		&ws,
		validation.ValidateVolume(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertVolume(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Volume,
				"",
				currentUser.ID,
				models.ResourceAccessModeCreate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to create volume in org '%s'", currentUser.ID, orgID)
			}
			obj, serr := h.volumeService.Create(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentVolume(obj, true), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

// func (h *volumeHandler) MarkAsSynced(w http.ResponseWriter, r *http.Request) {
// 	cfg := &handlerConfig{
// 		Action: func() (interface{}, *errors.ServiceError) {
// 			ctx := r.Context()
// 			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
// 			if err != nil {
// 				return nil, errors.Unauthorized("failed to fetch user")
// 			}
// 			storageID := mux.Vars(r)["id"]
// 			volumeID := mux.Vars(r)["volume_id"]

// 			storage, serr := h.storageService.GetByID(ctx, storageID)
// 			if serr != nil {
// 				return nil, serr
// 			}
// 			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
// 				currentUser,
// 				auth.StackStorage,
// 				storageID,
// 				storage.UserID,
// 				models.ResourceAccessModeUpdate,
// 			)
// 			if accessErr != nil {
// 				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
// 			}
// 			if !allowed {
// 				return nil, errors.Unauthorized("user '%s' is not allowed to update storage '%s'", currentUser.ID, storageID)
// 			}
// 			serr = h.storageService.MarkAsSynced(ctx, currentUser.ID, storageID, volumeID)
// 			if serr != nil {
// 				return nil, serr
// 			}
// 			return nil, nil
// 		},
// 	}
// 	handle(w, r, cfg, http.StatusOK)
// }

// func (h *volumeHandler) Update(w http.ResponseWriter, r *http.Request) {
// 	var ws openapi.StackStorage
// 	cfg := &handlerConfig{
// 		&ws,
// 		validation.ValidateStackStorage(&ws),
// 		func() (interface{}, *errors.ServiceError) {
// 			ctx := r.Context()
// 			id := mux.Vars(r)["id"]
// 			convertedObject := presenters.ConvertStorage(&ws)
// 			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
// 			if err != nil {
// 				return nil, errors.Unauthorized("failed to fetch current user")
// 			}
// 			storage, serr := h.storageService.GetByID(ctx, id)
// 			if serr != nil {
// 				return nil, serr
// 			}
// 			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
// 				currentUser,
// 				auth.StackStorage,
// 				id,
// 				storage.UserID,
// 				models.ResourceAccessModeUpdate,
// 			)
// 			if accessErr != nil {
// 				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
// 			}
// 			if !allowed {
// 				return nil, errors.Unauthorized("user '%s' is not allowed to update storage '%s'", currentUser.ID, id)
// 			}
// 			obj, serr := h.storageService.Update(ctx, id, currentUser.ID, convertedObject)
// 			if serr != nil {
// 				return nil, serr
// 			}
// 			return presenters.PresentStorage(obj), nil
// 		},
// 		handleError,
// 	}
// 	handle(w, r, cfg, http.StatusOK)
// }

func (h *volumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			volume, serr := h.volumeService.Get(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Volume,
				id,
				volume.UserID,
				models.ResourceAccessModeDelete,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete volume '%s'", currentUser.ID, id)
			}
			serr = h.volumeService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
