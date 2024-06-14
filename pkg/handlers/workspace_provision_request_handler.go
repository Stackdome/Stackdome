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
)

func NewWorkspaceProvisionRequestServiceHandler(spec WorkspaceProvisionRequestServiceHandlerSpec) *provisionRequestHandler {
	return &provisionRequestHandler{
		provisionRequestService: spec.WorkspaceProvisionRequestService,
	}
}

type WorkspaceProvisionRequestServiceHandlerSpec struct {
	WorkspaceProvisionRequestService services.WorkspaceProvisionRequestService
}

type provisionRequestHandler struct {
	provisionRequestService services.WorkspaceProvisionRequestService
}

func (a provisionRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()

			id := mux.Vars(r)["id"]

			obj, err := a.provisionRequestService.Get(ctx, id)
			if err != nil {
				return nil, err
			}
			return presenters.PresentWorkspaceProvisionRequest(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (a provisionRequestHandler) Create(w http.ResponseWriter, r *http.Request) {
	var wpr openapi.WorkspaceProvisionRequest
	cfg := &handlerConfig{
		&wpr,
		validation.ValidateWorkspaceProvisionRequest(&wpr),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertWorkspaceProvisionRequest(&wpr)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user from handler")
			}
			convertedObject.OrganisationID = currentUser.OrganisationID
			convertedObject.UserID = currentUser.ID
			obj, serr := a.provisionRequestService.Create(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}

			return presenters.PresentWorkspaceProvisionRequest(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (a provisionRequestHandler) Update(w http.ResponseWriter, r *http.Request) {
	var wpr openapi.WorkspaceProvisionRequest
	cfg := &handlerConfig{
		&wpr,
		validation.ValidateWorkspaceProvisionRequest(&wpr),
		func() (_ interface{}, returnErr *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			ctx := r.Context()
			convertedObject := presenters.ConvertWorkspaceProvisionRequest(&wpr)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			convertedObject.OrganisationID = currentUser.OrganisationID
			convertedObject.UserID = currentUser.ID
			obj, serr := a.provisionRequestService.Update(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}

			return presenters.PresentWorkspaceProvisionRequest(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (a provisionRequestHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (_ interface{}, returnErr *errors.ServiceError) {
			ctx := r.Context()

			id := mux.Vars(r)["id"]

			err := a.provisionRequestService.Delete(ctx, id)
			if err != nil {
				return nil, err
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
