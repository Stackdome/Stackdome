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
	"k8s.io/utils/ptr"
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
			created, serr := h.teamService.CreateTeam(r.Context(), orgID, team)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentTeam(created), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *teamHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teams, serr := h.teamService.ListTeams(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.TeamList{
				Items: presenters.PresentTeamList(teams),
				Total: ptr.To(int32(len(teams))),
			}, nil
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
			memberships, serr := h.teamService.ListUserTeams(r.Context(), identity.UserID)
			if serr != nil {
				return nil, serr
			}
			return openapi.TeamMembershipList{
				Items: presenters.PresentTeamMembershipList(memberships),
				Total: ptr.To(int32(len(memberships))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentTeam(team), nil
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
			updated, serr := h.teamService.UpdateTeam(r.Context(), team.ID, &models.Team{Name: req.GetName()})
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentTeam(updated), nil
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
			membership, serr := h.teamService.AddMember(r.Context(), team.ID, req.GetUserId(), presenters.ConvertTeamRole(req.GetRole()))
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentTeamMembership(membership), nil
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
			memberships, serr := h.teamService.ListMembers(r.Context(), team.ID)
			if serr != nil {
				return nil, serr
			}
			return openapi.TeamMembershipList{
				Items: presenters.PresentTeamMembershipList(memberships),
				Total: ptr.To(int32(len(memberships))),
			}, nil
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
			membership, serr := h.teamService.UpdateMemberRole(r.Context(), membershipID, presenters.ConvertTeamRole(req.GetRole()))
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentTeamMembership(membership), nil
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

func (h *teamHandler) ListTeamRoles(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			return openapi.TeamRoleList{
				Roles: []openapi.TeamRole{openapi.DEVELOPER, openapi.VIEWER},
			}, nil
		},
	}
	handleGet(w, r, cfg)
}
