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

type RegistryCredentialHandlerSpec struct {
	RegistryCredentialService services.RegistryCredentialService
	Logger                    logger.Logger
}

type registryCredentialHandler struct {
	registryCredentialService services.RegistryCredentialService
	logger                    logger.Logger
}

func NewRegistryCredentialHandler(spec RegistryCredentialHandlerSpec) *registryCredentialHandler {
	return &registryCredentialHandler{
		registryCredentialService: spec.RegistryCredentialService,
		logger:                    spec.Logger,
	}
}

func (h *registryCredentialHandler) Create(w http.ResponseWriter, r *http.Request) {
	var credential openapi.RegistryCredential
	cfg := &handlerConfig{
		MarshalInto: &credential,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertRegistryCredential(&credential)
			convertedObject.OrganisationID = mux.Vars(r)["org_id"]

			obj, serr := h.registryCredentialService.Create(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentRegistryCredential(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *registryCredentialHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			objs, serr := h.registryCredentialService.ListByOrgID(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.RegistryCredentialList{
				Items: presenters.PresentRegistryCredentialList(objs),
				Total: ptr.To(int32(len(objs))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *registryCredentialHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, serr := h.registryCredentialService.GetByID(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentRegistryCredential(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *registryCredentialHandler) Update(w http.ResponseWriter, r *http.Request) {
	var credential openapi.RegistryCredential
	cfg := &handlerConfig{
		MarshalInto: &credential,
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			convertedObject := presenters.ConvertRegistryCredential(&credential)

			obj, serr := h.registryCredentialService.Update(r.Context(), id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentRegistryCredential(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *registryCredentialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			result, serr := h.registryCredentialService.Delete(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			response := openapi.RegistryCredentialDeleteResponse{}
			var affected []openapi.AffectedStackRef
			for _, stack := range result.AffectedStacks {
				ref := openapi.AffectedStackRef{}
				ref.SetId(stack.ID)
				ref.SetName(stack.Name)
				affected = append(affected, ref)
			}
			response.SetAffectedStacks(affected)
			return response, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusOK)
}

func (h *registryCredentialHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var request openapi.RegistryCredentialVerifyRequest
	cfg := &handlerConfig{
		MarshalInto: &request,
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			purpose := presenters.ConvertRegistryCredentialVerifyPurpose(&request)
			if serr := h.registryCredentialService.Verify(r.Context(), id, request.GetRepository(), purpose); serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handle(w, r, cfg, http.StatusNoContent)
}
