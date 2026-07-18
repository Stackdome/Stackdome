package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type ImageBuildHandlerSpec struct {
	ImageBuildService services.ImageBuildService
	LoggingService    services.LoggingService
	Logger            logger.Logger
}

//go:generate mockgen -destination=logging_service_mock.go -package=handlers github.com/Stackdome/stackdome/pkg/services LoggingService

type imageBuildHandler struct {
	logger            logger.Logger
	imageBuildService services.ImageBuildService
	loggingService    services.LoggingService
}

func NewImageBuildHandler(spec ImageBuildHandlerSpec) *imageBuildHandler {
	return &imageBuildHandler{
		logger:            spec.Logger,
		imageBuildService: spec.ImageBuildService,
		loggingService:    spec.LoggingService,
	}
}

// List builds under a stack resource.
func (h *imageBuildHandler) ListByResourceName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			resourceName := mux.Vars(r)["resource_name"]

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

func (h *imageBuildHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			buildID := mux.Vars(r)["build_id"]

			loggingParams, pErr := services.NewLoggingParams(r.URL.Query())
			if pErr != nil {
				return nil, errors.MalformedRequest("invalid logging query params: %s", pErr.Error())
			}

			logStreamer, serr := h.loggingService.StreamLogsForBuild(ctx, orgID, buildID, loggingParams)
			if serr != nil {
				return nil, serr
			}
			return logStreamer, nil
		},
	}
	handleServerSideStream(w, r, cfg)
}
