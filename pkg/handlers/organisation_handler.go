package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
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

func (h *organisationHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	var req openapi.PromoteAdminRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			return nil, h.organisationService.PromoteToOrgAdmin(r.Context(), orgID, req.GetUserId())
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *organisationHandler) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			userID := mux.Vars(r)["user_id"]
			return nil, h.organisationService.DemoteOrgAdmin(r.Context(), orgID, userID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *organisationHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			admins, serr := h.organisationService.ListOrgAdmins(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			presented := make([]openapi.User, len(admins))
			for i, admin := range admins {
				presented[i] = presenters.PresentUser(admin)
			}
			return openapi.UserList{
				Items: presented,
				Total: ptr.To(int32(len(admins))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}
