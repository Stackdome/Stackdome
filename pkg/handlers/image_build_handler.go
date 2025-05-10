package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type ImageBuildHandlerSpec struct {
	ImageBuildService    services.ImageBuildService
	Logger               logger.Logger
	StackService         services.StackService
	AuthzClient          auth.AuthorizationClient
	StackResourceService services.StackResourceService
}

type imageBuildHandler struct {
	stackResourceService services.StackResourceService
	logger               logger.Logger
	stackService         services.StackService
	authzClient          auth.AuthorizationClient
	imageBuildService    services.ImageBuildService
}

func NewImageBuildHandler(spec ImageBuildHandlerSpec) *imageBuildHandler {
	return &imageBuildHandler{
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
		stackService:         spec.StackService,
		imageBuildService:    spec.ImageBuildService,
		authzClient:          spec.AuthzClient,
	}
}

// List builds under a stack resource.
func (h *imageBuildHandler) ListByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			stack, serr := h.stackService.GetStack(ctx, stackID)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				stackID,
				stack.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list stack resource '%s' builds", currentUser.ID, resourceName)
			}

			objs, err := h.imageBuildService.ListByResourceName(ctx, stackID, resourceName)
			if err != nil {
				return nil, err
			}
			listResp := openapi.ImageBuildList{
				Items: presenters.PresentImageBuildList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *imageBuildHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			buildID := mux.Vars(r)["build_id"]
			stackID := mux.Vars(r)["id"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			stack, serr := h.stackService.GetStack(ctx, stackID)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				stackID,
				stack.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to get image build '%s'", currentUser.ID, buildID)
			}
			obj, err := h.imageBuildService.GetByID(ctx, buildID)
			if err != nil {
				return nil, err
			}
			return presenters.PresentImageBuild(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *imageBuildHandler) ListByStackID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			stack, serr := h.stackService.GetStack(ctx, stackID)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				stackID,
				stack.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list stack '%s' builds", currentUser.ID, stackID)
			}
			objs, err := h.imageBuildService.ListByStackID(ctx, stackID)
			if err != nil {
				return nil, err
			}
			listResp := openapi.ImageBuildList{
				Items: presenters.PresentImageBuildList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}
