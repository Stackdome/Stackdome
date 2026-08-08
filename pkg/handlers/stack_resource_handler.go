package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/handlers/validation"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

// StackResourceHandlerSpec defines the dependencies for the stackResource handler
type StackResourceHandlerSpec struct {
	StackResourceService services.StackResourceService
	Logger               logger.Logger
	LoggingService       services.LoggingService
	MetricsService       services.MetricsService
}

type stackResourceHandler struct {
	stackResourceService services.StackResourceService
	loggingService       services.LoggingService
	metricsService       services.MetricsService
	logger               logger.Logger
}

func NewStackResourceHandler(spec StackResourceHandlerSpec) *stackResourceHandler {
	return &stackResourceHandler{
		stackResourceService: spec.StackResourceService,
		logger:               spec.Logger,
		loggingService:       spec.LoggingService,
		metricsService:       spec.MetricsService,
	}
}

func (h *stackResourceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var resource openapi.StackResource
	cfg := &handlerConfig{
		MarshalInto: &resource,
		Validate: func() *errors.ServiceError {
			return validation.ValidateStackResource(&resource, nil)
		},
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			modelResource := presenters.ConvertStackResource(&resource)
			modelResource.StackID = stackID
			created, err := h.stackResourceService.Create(r.Context(), modelResource)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackResource(created), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackResourceHandler) Update(w http.ResponseWriter, r *http.Request) {
	var resource openapi.StackResource
	cfg := &handlerConfig{
		MarshalInto: &resource,
		Validate: func() *errors.ServiceError {
			return validation.ValidateStackResource(&resource, nil)
		},
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			modelResource := presenters.ConvertStackResource(&resource)
			updated, err := h.stackResourceService.Update(r.Context(), stackID, resourceName, modelResource)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackResource(updated), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackResourceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			return nil, h.stackResourceService.Delete(r.Context(), stackID, resourceName)
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackResourceHandler) GetByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
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

			loggingParams, pErr := services.NewLoggingParams(r.URL.Query())
			if pErr != nil {
				return nil, errors.MalformedRequest("invalid logging query params: %s", pErr.Error())
			}

			logStreamer, err := h.loggingService.StreamLogsForStackResource(ctx, orgID, stackID, resourceName, loggingParams)
			if err != nil {
				return nil, err
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

			stream := r.URL.Query().Get("stream") == queryValueTrue
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

func (h *stackResourceHandler) Restart(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]
			resource, err := h.stackResourceService.Restart(ctx, stackID, resourceName)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackResource(resource), nil
		},
	}
	handle(w, r, cfg, http.StatusAccepted)
}
