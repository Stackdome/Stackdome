package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type StackHandlerSpec struct {
	StackService         services.StackService
	StackResourceService services.StackResourceService
	ImageBuildService    services.ImageBuildService
	AuthzClient          auth.AuthorizationClient
	Logger               logger.Logger
}

type stackHandler struct {
	stackService         services.StackService
	stackResourceService services.StackResourceService
	imageBuildService    services.ImageBuildService
	authzClient          auth.AuthorizationClient
	logger               logger.Logger
}

func NewStackHandler(spec StackHandlerSpec) *stackHandler {
	return &stackHandler{
		stackResourceService: spec.StackResourceService,
		stackService:         spec.StackService,
		imageBuildService:    spec.ImageBuildService,
		authzClient:          spec.AuthzClient,
		logger:               spec.Logger,
	}
}

func (h *stackHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.stackService.GetStack(ctx, id)
			if err != nil {
				return nil, err
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				id,
				obj.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access stack '%s'", currentUser.ID, id)
			}

			return presenters.PresentStack(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				"current",
				currentUser.ID,
				models.ResourceAccessModeList,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list stack", currentUser.ID)
			}

			objs, serr := h.stackService.GetStacksByUserID(ctx, currentUser.ID)
			if serr != nil {
				return nil, serr
			}

			listResp := openapi.StackList{
				Items: presenters.PresentStackList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackHandler) ListByOrganisationID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				"",
				"",
				models.ResourceAccessModeList,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list stacks under organistaion '%s'", currentUser.ID, orgID)
			}
			objs, serr := h.stackService.GetStacksByOrganisationID(ctx, orgID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.StackList{
				Items: presenters.PresentStackList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStack(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertStack(&ws)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				"",
				currentUser.ID,
				models.ResourceAccessModeCreate,
			)
			if accessErr != nil {
				h.logger.Errorf("failed to authorize access: %s", accessErr.Error())
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				h.logger.Errorf("user '%s' is not allowed to create stack", currentUser.ID)
				return nil, errors.Unauthorized("user '%s' is not allowed to create stack", currentUser.ID)
			}

			obj, serr := h.stackService.CreateStack(ctx, convertedObject)
			if serr != nil {
				h.logger.Errorf("failed to create workspace: %v", serr)
				return nil, serr
			}
			return presenters.PresentStack(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackHandler) Update(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStack(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			obj, serr := h.stackService.GetStack(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				id,
				obj.UserID,
				models.ResourceAccessModeUpdate,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update stack '%s'", currentUser.ID, id)
			}
			convertedObject := presenters.ConvertStack(&ws)
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID

			obj, serr = h.stackService.UpdateStack(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStack(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, serr := h.stackService.GetStack(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				id,
				obj.UserID,
				models.ResourceAccessModeDelete,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete stack '%s'", currentUser.ID, id)
			}

			serr = h.stackService.DeleteStack(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
