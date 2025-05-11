package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

// OrganisationHandlerSpec is the specification for the OrganisationHandler
type OrganisationHandlerSpec struct {
	OrganisationService services.OrganisationService
	AuthzClient         auth.AuthorizationClient
}

// organisationHandler is the handler for organisation related operations
type organisationHandler struct {
	organisationService services.OrganisationService
	authzClient         auth.AuthorizationClient
}

// NewOrganisationHandler creates a new OrganisationHandler

func NewOrganisationHandler(spec OrganisationHandlerSpec) *organisationHandler {
	return &organisationHandler{
		organisationService: spec.OrganisationService,
		authzClient:         spec.AuthzClient,
	}
}

// GetByID fetches an organisation by ID
func (h *organisationHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Organisation,
				id,
				"",
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize organisation access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access organisation '%s'", currentUser.ID, id)
			}
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
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			obj, serr := h.organisationService.GetDefaultOrg(ctx)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccess(
				currentUser,
				auth.Organisation,
				obj.ID,
				"",
				models.ResourceAccessModeRead,
			)
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize organisation access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access organisation '%s'", currentUser.ID, obj.ID)
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
			_, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
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
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			if !(currentUser.Role == models.PlatformAdminRole || currentUser.Role == models.OrganisationAdminRole) {
				return nil, errors.Unauthorized("user '%s' is not allowed to update organisation '%s'", currentUser.ID, id)
			}

			if currentUser.Role == models.OrganisationAdminRole && currentUser.OrganisationID != id {
				return nil, errors.Unauthorized("user '%s' is not allowed to update organisation '%s'", currentUser.ID, id)
			}
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
