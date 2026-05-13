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
	"k8s.io/utils/ptr"
)

type APITokenHandlerSpec struct {
	APITokenService services.APITokenService
}

type apiTokenHandler struct {
	apiTokenService services.APITokenService
}

func NewAPITokenHandler(spec APITokenHandlerSpec) *apiTokenHandler {
	return &apiTokenHandler{
		apiTokenService: spec.APITokenService,
	}
}

func (h *apiTokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req openapi.APITokenCreateRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Validate:    validation.ValidateAPITokenCreate(&req),
		Action: func() (interface{}, *errors.ServiceError) {
			token, rawToken, serr := h.apiTokenService.Create(
				r.Context(), req.GetName(), req.GetScopes(), req.GetResourceIds(), req.ExpiresAt,
			)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentAPITokenCreateResponse(token, rawToken), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *apiTokenHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			tokens, serr := h.apiTokenService.List(r.Context())
			if serr != nil {
				return nil, serr
			}
			return openapi.APITokenList{
				Items: presenters.PresentAPITokenList(tokens),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *apiTokenHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			token, serr := h.apiTokenService.GetByID(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentAPIToken(token), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *apiTokenHandler) ListScopes(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			var scopes []openapi.ScopeResource
			for _, rt := range auth.ResourceTypes {
				scopes = append(scopes, openapi.ScopeResource{
					Resource: ptr.To(rt.Name),
					Actions:  rt.Actions,
				})
			}
			return openapi.ScopeList{
				FullAccessScope: ptr.To(auth.ScopeFullAccess),
				Items:           scopes,
				Total:           ptr.To(int32(len(scopes))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *apiTokenHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			return nil, h.apiTokenService.Revoke(r.Context(), id)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
