package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/gorilla/mux"
)

type PreviewStackHandlerSpec struct {
	Service     services.PreviewStackService
	TeamService services.TeamService
}

type previewStackHandler struct {
	service     services.PreviewStackService
	teamService services.TeamService
}

func NewPreviewStackHandler(spec PreviewStackHandlerSpec) *previewStackHandler {
	return &previewStackHandler{
		service:     spec.Service,
		teamService: spec.TeamService,
	}
}

func (h *previewStackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.PreviewStackCreate
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (any, *errors.ServiceError) {
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			identity := auth.GetIdentityFromCtx(r.Context())
			model := presenters.ConvertPreviewStackCreate(&req)
			model.UserID = identity.UserID
			model.TeamID = teamID
			preview, serr := h.service.Create(r.Context(), model)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentPreviewStack(preview), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusAccepted)
}

func (h *previewStackHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (any, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			preview, serr := h.service.Get(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentPreviewStack(preview), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *previewStackHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (any, *errors.ServiceError) {
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			params := parseListParams(r, nil)
			if configID := r.URL.Query().Get("config_id"); configID != "" {
				params.Filters = append(params.Filters, stores.Filter{Field: "stack_preview_config_id", Value: configID})
			}
			result, serr := h.service.List(r.Context(), teamID, params)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentPreviewStackList(result), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *previewStackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (any, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			serr := h.service.Delete(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *previewStackHandler) Sync(w http.ResponseWriter, r *http.Request) {
	var req openapi.PreviewStackSync
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (any, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			serr := h.service.Sync(r.Context(), id, services.PreviewSyncOpts{
				Commit:           req.GetCommit(),
				StackfileContent: req.StackfileContent,
				ForceSync:        req.GetForceSync(),
				ImageOverrides:   req.GetImageOverrides(),
			})
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusNoContent)
}
