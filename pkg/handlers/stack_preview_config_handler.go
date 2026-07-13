package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
)

type StackPreviewConfigHandlerSpec struct {
	Service        services.StackPreviewConfigService
	ProjectService services.ProjectService
}

type stackPreviewConfigHandler struct {
	service        services.StackPreviewConfigService
	projectService services.ProjectService
}

func NewStackPreviewConfigHandler(spec StackPreviewConfigHandlerSpec) *stackPreviewConfigHandler {
	return &stackPreviewConfigHandler{
		service:        spec.Service,
		projectService: spec.ProjectService,
	}
}

func (h *stackPreviewConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.StackPreviewConfigCreate
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			projectID, serr := resolveProjectID(r, h.projectService)
			if serr != nil {
				return nil, serr
			}
			identity := auth.GetIdentityFromCtx(r.Context())
			model := presenters.ConvertStackPreviewConfigCreate(&req)
			model.OrganisationID = identity.OrgID
			model.ProjectID = projectID
			model.UserID = identity.UserID
			config, serr := h.service.Create(r.Context(), model)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStackPreviewConfig(config), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackPreviewConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			config, serr := h.service.Get(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStackPreviewConfig(config), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackPreviewConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			projectID, serr := resolveProjectID(r, h.projectService)
			if serr != nil {
				return nil, serr
			}
			params := parseListParams(r, nil)
			result, serr := h.service.List(r.Context(), projectID, params)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStackPreviewConfigList(result), nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackPreviewConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req openapi.StackPreviewConfigUpdate
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			updated := presenters.ConvertStackPreviewConfigUpdate(&req)
			config, serr := h.service.Update(r.Context(), id, updated)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStackPreviewConfig(config), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackPreviewConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			serr := h.service.Delete(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
