package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type ImageBuildHandlerSpec struct {
	ImageBuildService services.ImageBuildService
	Logger            logger.Logger
}

type imageBuildHandler struct {
	logger            logger.Logger
	imageBuildService services.ImageBuildService
}

func NewImageBuildHandler(spec ImageBuildHandlerSpec) *imageBuildHandler {
	return &imageBuildHandler{
		logger:            spec.Logger,
		imageBuildService: spec.ImageBuildService,
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
