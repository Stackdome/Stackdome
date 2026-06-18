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

func NewVolumeHandler(spec VolumeHandlerSpec) *volumeHandler {
	return &volumeHandler{
		volumeService: spec.VolumeService,
		teamService:   spec.TeamService,
	}
}

type VolumeHandlerSpec struct {
	VolumeService services.VolumeService
	TeamService   services.TeamService
}

type volumeHandler struct {
	volumeService services.VolumeService
	teamService   services.TeamService
}

func (h *volumeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			if id == "current" {
				teamID, serr := resolveTeamID(r, h.teamService)
				if serr != nil {
					return nil, serr
				}
				currentUser, err := auth.GetCurrentUserFromCtx(ctx)
				if err != nil {
					return nil, errors.Unauthorized("failed to fetch current user")
				}
				objs, serr := h.volumeService.ListByUserID(ctx, teamID, currentUser.ID)
				if serr != nil {
					return nil, serr
				}
				listResp := openapi.VolumeList{
					Items: presenters.PresentVolumeList(objs, true),
					Total: ptr.To(int32(len(objs))),
				}
				return listResp, nil
			}
			obj, err := h.volumeService.Get(ctx, id)
			if err != nil {
				return nil, err
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
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			convertedObject := presenters.ConvertVolume(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = currentUser.ID

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

func (h *volumeHandler) ListByStackID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			objs, err := h.volumeService.ListVolumesUsedByStack(ctx, stackID)
			if err != nil {
				return nil, err
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

func (h *volumeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			serr := h.volumeService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
