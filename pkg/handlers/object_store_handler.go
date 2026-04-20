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
	"k8s.io/utils/ptr"
)

type ObjectStoreHandlerSpec struct {
	ObjectStoreService services.ObjectStoreService
	AuthzClient        auth.AuthorizationClient
}

type objectStoreHandler struct {
	objectStoreService services.ObjectStoreService
	authzClient        auth.AuthorizationClient
}

func NewObjectStoreHandler(spec ObjectStoreHandlerSpec) *objectStoreHandler {
	return &objectStoreHandler{
		objectStoreService: spec.ObjectStoreService,
		authzClient:        spec.AuthzClient,
	}
}

func (h *objectStoreHandler) Create(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["org_id"]

	var apiObjectStore openapi.ObjectStore
	cfg := &handlerConfig{
		MarshalInto: &apiObjectStore,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, userErr := auth.GetCurrentUserFromCtx(ctx)
			if userErr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Check authorization for creating object stores
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.ObjectStore,
				ResourceID:      "",
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeWrite,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to create object store", currentUser.ID)
			}

			objectStore := presenters.ConvertObjectStore(&apiObjectStore)
			objectStore.OrganisationID = orgID

			createdObjectStore, err := h.objectStoreService.Create(ctx, objectStore)
			if err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(createdObjectStore), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *objectStoreHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["org_id"]

	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, userErr := auth.GetCurrentUserFromCtx(ctx)
			if userErr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Check authorization for listing object stores
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.ObjectStore,
				ResourceID:      "",
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeRead,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list object stores", currentUser.ID)
			}

			objectStores, err := h.objectStoreService.ListByOrganisation(ctx, orgID)
			if err != nil {
				return nil, err
			}

			return openapi.ObjectStoreList{
				Items: presenters.PresentObjectStoreList(objectStores),
				Total: ptr.To(int32(len(objectStores))),
			}, nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *objectStoreHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, userErr := auth.GetCurrentUserFromCtx(ctx)
			if userErr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			objectStore, err := h.objectStoreService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}

			// Check authorization for reading this specific object store
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.ObjectStore,
				ResourceID:      id,
				ResourceOwnerID: objectStore.OrganisationID, // Use organisation ID as owner
				Action:          models.ResourceAccessModeRead,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access object store '%s'", currentUser.ID, id)
			}

			return presenters.PresentObjectStore(objectStore), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *objectStoreHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var apiObjectStore openapi.ObjectStore
	cfg := &handlerConfig{
		MarshalInto: &apiObjectStore,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, userErr := auth.GetCurrentUserFromCtx(ctx)
			if userErr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			objectStore := presenters.ConvertObjectStore(&apiObjectStore)

			// Get existing object store to check authorization
			existingObjectStore, err := h.objectStoreService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}

			// Check authorization for updating this specific object store
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.ObjectStore,
				ResourceID:      id,
				ResourceOwnerID: existingObjectStore.OrganisationID,
				Action:          models.ResourceAccessModeWrite,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update object store '%s'", currentUser.ID, id)
			}

			updatedObjectStore, err := h.objectStoreService.Update(ctx, id, objectStore)
			if err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(updatedObjectStore), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *objectStoreHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			currentUser, userErr := auth.GetCurrentUserFromCtx(ctx)
			if userErr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			objectStore, err := h.objectStoreService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}

			// Check authorization for deleting this specific object store
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.ObjectStore,
				ResourceID:      id,
				ResourceOwnerID: objectStore.OrganisationID,
				Action:          models.ResourceAccessModeWrite,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete object store '%s'", currentUser.ID, id)
			}

			if err := h.objectStoreService.Delete(ctx, id); err != nil {
				return nil, err
			}

			return presenters.PresentObjectStore(objectStore), nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
