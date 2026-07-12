package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
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
				return presenters.PresentStackRelease(release, nil), nil
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
			return presenters.PresentStackRelease(release, nil), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackReleaseHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			params := parseListParams(r, []string{"state"})
			result, err := h.releaseService.ListReleases(r.Context(), stackID, params)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackReleaseList(result), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackReleaseHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			releaseID := mux.Vars(r)["release_id"]
			release, live, err := h.releaseService.GetReleaseDetail(r.Context(), releaseID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackReleaseDetail(release, live), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackReleaseHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			releaseID := mux.Vars(r)["release_id"]

			afterSequence, err := parseOptionalIntQuery(r, "after_sequence", 0)
			if err != nil {
				return nil, errors.MalformedRequest("invalid after_sequence")
			}
			limit, err := parseOptionalIntQuery(r, "limit", 0)
			if err != nil {
				return nil, errors.MalformedRequest("invalid limit")
			}

			page, serr := h.releaseService.ListReleaseEvents(r.Context(), stackID, releaseID, afterSequence, limit)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentReleaseEventList(page.Events, page.NextAfterSequence), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackReleaseHandler) StreamEvents(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			releaseID := mux.Vars(r)["release_id"]

			afterSequence, err := parseOptionalIntQuery(r, "after_sequence", 0)
			if err != nil {
				return nil, errors.MalformedRequest("invalid after_sequence")
			}

			return h.releaseService.StreamReleaseEvents(r.Context(), stackID, releaseID, afterSequence)
		},
	}
	handleServerSideStream(w, r, cfg)
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
