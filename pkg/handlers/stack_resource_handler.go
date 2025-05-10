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

// StackHandlerSpec defines the dependencies for the stackResource handler
type StackResourceHandlerSpec struct {
	StackResourceService services.StackResourceService
	Logger               logger.Logger
	StackService         services.StackService
	AuthzClient          auth.AuthorizationClient
}

type stackResourceHandler struct {
	stackResourceService services.StackResourceService
	logger               logger.Logger
	stackService         services.StackService
	authzClient          auth.AuthorizationClient
}

func NewStackResourceHandler(spec StackResourceHandlerSpec) *stackResourceHandler {
	return &stackResourceHandler{
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
		stackService:         spec.StackService,
		authzClient:          spec.AuthzClient,
	}
}

// GetByID fetches a workspace resource by its ID
func (h *stackResourceHandler) GetByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			workspace, serr := h.stackService.GetStack(ctx, stackID)
			if serr != nil {
				return nil, serr
			}

			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Stack,
				stackID,
				workspace.UserID,
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to get workspace resource '%s'", currentUser.ID, resourceName)
			}
			obj, err := h.stackResourceService.GetByStackIDAndResourceName(ctx, stackID, resourceName)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackResource(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

// List fetches all workspace resources
func (h *stackResourceHandler) List(w http.ResponseWriter, r *http.Request) {
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
				return nil, errors.Unauthorized("user '%s' is not allowed to list workspace '%s' resources", currentUser.ID, stackID)
			}
			objs, err := h.stackResourceService.GetByStackID(ctx, stackID)
			if err != nil {
				return nil, err
			}

			listResp := openapi.StackResourceList{
				Items: presenters.PresentStackResourceList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}
