package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

// OrganisationHandlerSpec is the specification for the OrganisationHandler
type OrganisationHandlerSpec struct {
	OrganisationService services.OrganisationService
	ProjectService      services.ProjectService
}

// organisationHandler is the handler for organisation related operations
type organisationHandler struct {
	organisationService services.OrganisationService
	projectService      services.ProjectService
}

// NewOrganisationHandler creates a new OrganisationHandler

func NewOrganisationHandler(spec OrganisationHandlerSpec) *organisationHandler {
	return &organisationHandler{
		organisationService: spec.OrganisationService,
		projectService:      spec.ProjectService,
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
	var req openapi.DemoteAdminRequest
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			userID := mux.Vars(r)["user_id"]

			project, serr := h.projectService.GetProjectByOrgAndName(r.Context(), orgID, req.GetProjectName())
			if serr != nil {
				return nil, serr
			}

			role := models.ViewerRole
			if req.HasRole() {
				role = presenters.ConvertProjectRole(req.GetRole())
			}

			return nil, h.organisationService.DemoteOrgAdmin(r.Context(), orgID, userID, project.ID, role)
		},
	}
	handle(w, r, cfg, http.StatusNoContent)
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
