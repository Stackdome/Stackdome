package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/handlers/validation"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
)

func NewWorkspaceUserHandler(spec WorkspaceUserHandlerSpec) *workspaceUserHandler {
	return &workspaceUserHandler{
		workspaceUserService: spec.WorkspaceUserService,
		teamService:          spec.TeamService,
	}
}

type WorkspaceUserHandlerSpec struct {
	WorkspaceUserService services.WorkspaceUserService
	TeamService          services.TeamService
}

type workspaceUserHandler struct {
	workspaceUserService services.WorkspaceUserService
	teamService          services.TeamService
}

func (a workspaceUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			obj, serr := a.workspaceUserService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}

			return presenters.PresentWorkspaceUser(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a workspaceUserHandler) Current(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, serr := a.workspaceUserService.GetWorkspaceUser(ctx, currentUser.ID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentWorkspaceUser(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a workspaceUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var wpr openapi.WorkspaceUser
	cfg := &handlerConfig{
		&wpr,
		validation.ValidateWorkspaceUser(&wpr),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, a.teamService)
			if serr != nil {
				return nil, serr
			}
			convertedObject := presenters.ConvertWorkspaceUser(&wpr)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			convertedObject.OrganisationID = currentUser.OrganisationID
			convertedObject.TeamID = teamID
			convertedObject.UserID = currentUser.ID
			obj, serr := a.workspaceUserService.Create(ctx, convertedObject, currentUser)
			if serr != nil {
				return nil, serr
			}

			return presenters.PresentWorkspaceUser(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (a workspaceUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	var wu openapi.WorkspaceUser
	cfg := &handlerConfig{
		&wu,
		validation.ValidateWorkspaceUser(&wu),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			ctx := r.Context()
			convertedObject := presenters.ConvertWorkspaceUser(&wu)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			convertedObject.OrganisationID = currentUser.OrganisationID
			convertedObject.UserID = currentUser.ID
			obj, serr := a.workspaceUserService.Update(ctx, id, convertedObject, currentUser)
			if serr != nil {
				return nil, serr
			}

			return presenters.PresentWorkspaceUser(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (a workspaceUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			serr := a.workspaceUserService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
