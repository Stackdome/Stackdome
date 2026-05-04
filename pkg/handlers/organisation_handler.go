package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

// OrganisationHandlerSpec is the specification for the OrganisationHandler
type OrganisationHandlerSpec struct {
	OrganisationService services.OrganisationService
}

// organisationHandler is the handler for organisation related operations
type organisationHandler struct {
	organisationService services.OrganisationService
}

// NewOrganisationHandler creates a new OrganisationHandler

func NewOrganisationHandler(spec OrganisationHandlerSpec) *organisationHandler {
	return &organisationHandler{
		organisationService: spec.OrganisationService,
	}
}

// GetByID fetches an organisation by ID
func (h *organisationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			obj, serr := h.organisationService.Get(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrganisation(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

// Get default organisation

func (h *organisationHandler) GetDefault(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			obj, serr := h.organisationService.GetDefaultOrg(ctx)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrganisation(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

// Create an organisation
func (h *organisationHandler) Create(w http.ResponseWriter, r *http.Request) {
	var org openapi.Organisation
	cfg := &handlerConfig{
		MarshalInto: &org,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			spec := presenters.ConvertOrganisation(org)
			obj, serr := h.organisationService.Create(ctx, spec)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrganisation(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

// Update an organisation
func (h *organisationHandler) Update(w http.ResponseWriter, r *http.Request) {
	var org openapi.Organisation
	cfg := &handlerConfig{
		MarshalInto: &org,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			spec := presenters.ConvertOrganisation(org)
			obj, serr := h.organisationService.Update(ctx, id, spec)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentOrganisation(obj), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}
