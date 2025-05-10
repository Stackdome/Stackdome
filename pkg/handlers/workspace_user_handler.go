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
)

func NewWorkspaceUserHandler(spec WorkspaceUserHandlerSpec) *workspaceUserHandler {
	return &workspaceUserHandler{
		workspaceUserService: spec.WorkspaceUserService,
		authzClient:          spec.AuthzClient,
	}
}

type WorkspaceUserHandlerSpec struct {
	WorkspaceUserService services.WorkspaceUserService
	AuthzClient          auth.AuthorizationClient
}

type workspaceUserHandler struct {
	workspaceUserService services.WorkspaceUserService
	authzClient          auth.AuthorizationClient
}

func (a workspaceUserHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			id := mux.Vars(r)["id"]

			obj, serr := a.workspaceUserService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := a.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceUser,
				id,
				obj.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to read workspaceuser '%s'", currentUser.ID, id)
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
			convertedObject := presenters.ConvertWorkspaceUser(&wpr)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			allowed, accessErr := a.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceUser,
				"",
				currentUser.ID,
				models.ResourceAccessModeCreate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to create workspaceuser", currentUser.ID)
			}
			convertedObject.OrganisationID = currentUser.OrganisationID
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

			obj, serr := a.workspaceUserService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := a.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceUser,
				id,
				obj.UserID,
				models.ResourceAccessModeUpdate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update workspaceuser '%s'", currentUser.ID, id)
			}

			convertedObject.OrganisationID = currentUser.OrganisationID
			convertedObject.UserID = currentUser.ID
			obj, serr = a.workspaceUserService.Update(ctx, id, convertedObject, currentUser)
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
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			id := mux.Vars(r)["id"]

			obj, serr := a.workspaceUserService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := a.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.WorkspaceUser,
				id,
				obj.UserID,
				models.ResourceAccessModeDelete,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete workspaceuser '%s'", currentUser.ID, id)
			}

			serr = a.workspaceUserService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
