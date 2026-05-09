package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

type TeamHandlerSpec struct {
	TeamService services.TeamService
}

type teamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(spec TeamHandlerSpec) *teamHandler {
	return &teamHandler{
		teamService: spec.TeamService,
	}
}

func (h *teamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.TeamCreateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			team := &models.Team{Name: req.GetName()}
			return h.teamService.CreateTeam(r.Context(), orgID, team)
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *teamHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			return h.teamService.ListTeams(r.Context(), orgID)
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) ListCurrentUserTeams(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			identity := auth.GetIdentityFromCtx(r.Context())
			if identity == nil {
				return nil, errors.Unauthorized("user identity not found in context")
			}
			return h.teamService.ListUserTeams(r.Context(), identity.UserID)
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			return h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
		},
	}
	handleGet(w, r, cfg)
}

func (h *teamHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req openapi.TeamUpdateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.UpdateTeam(r.Context(), team.ID, &models.Team{Name: req.GetName()})
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *teamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return nil, h.teamService.DeleteTeam(r.Context(), team.ID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *teamHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	var req openapi.AddTeamMemberRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.AddMember(r.Context(), team.ID, req.GetUserId(), models.TeamRole(req.GetRole()))
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *teamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.ListMembers(r.Context(), team.ID)
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req openapi.UpdateTeamMemberRoleRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			return h.teamService.UpdateMemberRole(r.Context(), membershipID, models.TeamRole(req.GetRole()))
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *teamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			return nil, h.teamService.RemoveMember(r.Context(), membershipID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
