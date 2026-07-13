package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
)

type OrgInviteHandlerSpec struct {
	OrgInviteService services.OrgInviteService
}

type orgInviteHandler struct {
	inviteService services.OrgInviteService
}

func NewOrgInviteHandler(spec OrgInviteHandlerSpec) *orgInviteHandler {
	return &orgInviteHandler{
		inviteService: spec.OrgInviteService,
	}
}

func (h *orgInviteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.OrgInviteCreateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			role := models.ProjectRole(req.Role)
			invite, rawToken, serr := h.inviteService.Create(r.Context(), req.Email, req.ProjectName, role, int(req.ExpiresInDays))
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrgInviteCreateResponse(invite, rawToken), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *orgInviteHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			params := parseListParams(r, []string{"status"})
			invites, serr := h.inviteService.List(r.Context(), orgID, params)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrgInviteList(invites), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *orgInviteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			id := mux.Vars(r)["id"]
			invite, serr := h.inviteService.GetByID(r.Context(), orgID, id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrgInvite(invite), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *orgInviteHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			id := mux.Vars(r)["id"]
			return nil, h.inviteService.Revoke(r.Context(), orgID, id)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *orgInviteHandler) Resend(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			id := mux.Vars(r)["id"]
			return nil, h.inviteService.Resend(r.Context(), orgID, id)
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *orgInviteHandler) GetInviteInfo(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			token := mux.Vars(r)["token"]
			// Public endpoint.
			invite, serr := h.inviteService.PublicGetInviteInfo(r.Context(), token)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrgInviteInfo(invite), nil
		},
	}
	handleGet(w, r, cfg)
}
