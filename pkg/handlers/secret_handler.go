package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type SecretHandlerSpec struct {
	SecretService services.SecretService
	AuthzClient   auth.AuthorizationClient
	Logger        logger.Logger
}

type secretHandler struct {
	secretService services.SecretService
	authzClient   auth.AuthorizationClient
	logger        logger.Logger
}

func NewSecretHandler(spec SecretHandlerSpec) *secretHandler {
	return &secretHandler{
		secretService: spec.SecretService,
		authzClient:   spec.AuthzClient,
		logger:        spec.Logger,
	}
}

func (h *secretHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			obj, err := h.secretService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.Secret,
				ResourceID:      id,
				ResourceOwnerID: obj.UserID,
				Action:          models.ResourceAccessModeRead,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to access secret '%s'", currentUser.ID, id)
			}

			return presenters.PresentSecret(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *secretHandler) ListByOrganisationID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:         currentUser,
				ResourceType: auth.Secret,
				Action:       models.ResourceAccessModeList,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to list secrets '%s' under organisation %s", currentUser.ID, orgID)
			}
			objs, serr := h.secretService.ListByOrganisation(ctx, orgID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.SecretList{
				Items: presenters.PresentSecretList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *secretHandler) Create(w http.ResponseWriter, r *http.Request) {
	var secret openapi.Secret
	cfg := &handlerConfig{
		&secret,
		nil,
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			convertedObject := presenters.ConvertSecret(&secret)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.Secret,
				ResourceID:      "",
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeCreate,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to create secret '%s'", currentUser.ID, secret.Name)
			}

			obj, serr := h.secretService.Create(ctx, convertedObject)
			if serr != nil {
				h.logger.Errorf("failed to create secret: %v", serr)
				return nil, serr
			}
			return presenters.PresentSecret(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *secretHandler) Update(w http.ResponseWriter, r *http.Request) {
	var secret openapi.Secret
	cfg := &handlerConfig{
		&secret,
		nil,
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			orgID := mux.Vars(r)["org_id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			obj, serr := h.secretService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.Secret,
				ResourceID:      obj.ID,
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeUpdate,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to update secret '%s'", currentUser.ID, secret.Name)
			}
			convertedObject := presenters.ConvertSecret(&secret)
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID

			obj, serr = h.secretService.Update(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentSecret(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *secretHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}
			secret, serr := h.secretService.GetByID(ctx, id)
			if serr != nil {
				return nil, serr
			}
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.Secret,
				ResourceID:      secret.ID,
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeUpdate,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to authorize access: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Unauthorized("user '%s' is not allowed to delete secret '%s'", currentUser.ID, secret.Name)
			}
			serr = h.secretService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
