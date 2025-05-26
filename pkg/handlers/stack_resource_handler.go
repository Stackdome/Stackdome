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
	LoggingService       services.LoggingService
	MetricsService       services.MetricsService
	AuthzClient          auth.AuthorizationClient
}

type stackResourceHandler struct {
	stackResourceService services.StackResourceService
	loggingService       services.LoggingService
	metricsService       services.MetricsService
	logger               logger.Logger
	stackService         services.StackService
	authzClient          auth.AuthorizationClient
}

func NewStackResourceHandler(spec StackResourceHandlerSpec) *stackResourceHandler {
	return &stackResourceHandler{
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
		stackService:         spec.StackService,
		loggingService:       spec.LoggingService,
		metricsService:       spec.MetricsService,
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

func (h *stackResourceHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			orgID := mux.Vars(r)["org_id"]
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
				return nil, errors.Unauthorized("user '%s' is not allowed to stream logs for resource '%s'", currentUser.ID, resourceName)
			}

			loggingParams, pErr := services.NewLoggingParams(r.URL.Query())
			if pErr != nil {
				return nil, errors.MalformedRequest("invalid logging query params: %s", pErr.Error())
			}

			logStreamer, err := h.loggingService.StreamLogsForStackResource(ctx, orgID, stackID, resourceName, loggingParams)
			if err != nil {
				return nil, errors.GeneralError("failed to get logs for resource '%s': %s", resourceName, err.Error())
			}
			return logStreamer, nil
		},
	}
	handleServerSideStream(w, r, cfg)
}

func (h *stackResourceHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			orgID := mux.Vars(r)["org_id"]
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
				return nil, errors.Unauthorized("user '%s' is not allowed to get metrics for resource '%s'", currentUser.ID, resourceName)
			}

			stream := r.URL.Query().Get("stream") == "true"
			if stream {
				streamer, err := h.metricsService.StreamMetricsForStackResource(ctx, orgID, stackID, resourceName)
				if err != nil {
					return nil, errors.GeneralError("failed to get metrics for resource '%s': %s", resourceName, err.Error())
				}
				return streamer, nil
			}
			metrics, err := h.metricsService.GetMetricsForStackResource(ctx, orgID, stackID, resourceName)
			if err != nil {
				return nil, errors.GeneralError("failed to get metrics for resource '%s': %s", resourceName, err.Error())
			}
			return presenters.PresentResourceMetrics(metrics), nil
		},
	}
	handleStreamOrGet(w, r, cfg)
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
