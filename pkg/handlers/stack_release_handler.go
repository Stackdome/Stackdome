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

type StackReleaseHandlerSpec struct {
	StackReleaseService services.StackReleaseService
}

type stackReleaseHandler struct {
	releaseService services.StackReleaseService
}

func NewStackReleaseHandler(spec StackReleaseHandlerSpec) *stackReleaseHandler {
	return &stackReleaseHandler{releaseService: spec.StackReleaseService}
}

func (h *stackReleaseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.CreateReleaseRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]

			if req.GetFromReleaseId() != "" {
				release, err := h.releaseService.RollbackRelease(r.Context(), stackID, req.GetFromReleaseId())
				if err != nil {
					return nil, err
				}
				return presenters.PresentStackRelease(release), nil
			}

			identity := auth.GetIdentityFromCtx(r.Context())
			var detail string
			if identity != nil {
				detail = "triggered by " + identity.UserID
			}

			release, err := h.releaseService.CreateRelease(r.Context(), stackID, models.ReleaseCause{
				Kind:   models.ReleaseCauseManual,
				Detail: detail,
			})
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackRelease(release), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackReleaseHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			releases, err := h.releaseService.ListReleases(r.Context(), stackID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackReleaseList(releases), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackReleaseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			releaseID := mux.Vars(r)["release_id"]
			release, err := h.releaseService.GetRelease(r.Context(), releaseID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackRelease(release), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackReleaseHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			releaseID := mux.Vars(r)["release_id"]
			if err := h.releaseService.CancelRelease(r.Context(), releaseID); err != nil {
				return nil, err
			}
			return nil, nil
		},
	}
	handleGet(w, r, cfg)
}
