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

type PostgresAddonHandlerSpec struct {
	PostgresAddonService services.PostgresAddonService
	AuthzClient          auth.AuthorizationClient
	Logger               logger.Logger
}

type postgresAddonHandler struct {
	postgresAddonService services.PostgresAddonService
	authzClient          auth.AuthorizationClient
	logger               logger.Logger
}

func NewPostgresAddonHandler(spec PostgresAddonHandlerSpec) *postgresAddonHandler {
	return &postgresAddonHandler{
		postgresAddonService: spec.PostgresAddonService,
		authzClient:          spec.AuthzClient,
		logger:               spec.Logger,
	}
}

func (h *postgresAddonHandler) Create(w http.ResponseWriter, r *http.Request) {
	var apiPostgresAddon openapi.PostgresAddon
	cfg := &handlerConfig{
		MarshalInto: &apiPostgresAddon,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Convert API model to domain model
			postgresAddon := presenters.ConvertPostgresAddon(&apiPostgresAddon)
			postgresAddon.OrganisationID = orgID
			postgresAddon.UserID = currentUser.ID

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      "",
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeCreate,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to create postgres addon")
			}

			// Create postgres addon
			createdPostgresAddon, err := h.postgresAddonService.CreatePostgresAddon(ctx, postgresAddon)
			if err != nil {
				return nil, err
			}

			return presenters.PresentPostgresAddon(createdPostgresAddon), nil
		},
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *postgresAddonHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			orgID := mux.Vars(r)["org_id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      orgID,
				ResourceOwnerID: currentUser.ID,
				Action:          models.ResourceAccessModeList,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to list postgres addons")
			}

			// List postgres addons
			postgresAddons, err := h.postgresAddonService.ListPostgresAddonsByOrganisation(ctx, orgID)
			if err != nil {
				return nil, err
			}

			return openapi.PostgresAddonList{
				Items: presenters.PresentPostgresAddonList(postgresAddons),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeRead,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to read postgres addon")
			}

			return presenters.PresentPostgresAddon(postgresAddon), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) Update(w http.ResponseWriter, r *http.Request) {
	var apiPostgresAddon openapi.PostgresAddon
	cfg := &handlerConfig{
		MarshalInto: &apiPostgresAddon,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Convert API model to domain model
			postgresAddon := presenters.ConvertPostgresAddon(&apiPostgresAddon)
			postgresAddon.ID = id

			// Get existing postgres addon for authorization
			existingPostgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			postgresAddon.OrganisationID = existingPostgresAddon.OrganisationID
			postgresAddon.UserID = currentUser.ID

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: existingPostgresAddon.UserID,
				Action:          models.ResourceAccessModeUpdate,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to update postgres addon")
			}

			// Update postgres addon
			updatedPostgresAddon, err := h.postgresAddonService.UpdatePostgresAddon(ctx, id, postgresAddon)
			if err != nil {
				return nil, err
			}

			return presenters.PresentPostgresAddon(updatedPostgresAddon), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon for authorization
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeDelete,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to delete postgres addon")
			}

			// Delete postgres addon
			_, err = h.postgresAddonService.DeletePostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			return presenters.PresentPostgresAddon(postgresAddon), nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) Backup(w http.ResponseWriter, r *http.Request) {
	var backupRequest openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPostRequest
	cfg := &handlerConfig{
		MarshalInto: &backupRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon for authorization
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeExecute,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to backup postgres addon")
			}

			// Trigger backup
			err = h.postgresAddonService.TriggerBackup(ctx, id)
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsBackupPost202Response{
				Message: ptr.To("Backup initiated successfully"),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusAccepted)
}

func (h *postgresAddonHandler) Fence(w http.ResponseWriter, r *http.Request) {
	var fenceRequest openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePostRequest
	cfg := &handlerConfig{
		MarshalInto: &fenceRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon for authorization
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeExecute,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to fence postgres addon")
			}

			// Trigger fence action
			err = h.postgresAddonService.TriggerFence(ctx, id, fenceRequest.GetFence())
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsFencePost200Response{
				Message: ptr.To("Fence action initiated successfully"),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) Hibernate(w http.ResponseWriter, r *http.Request) {
	var hibernateRequest openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsHibernatePostRequest
	cfg := &handlerConfig{
		MarshalInto: &hibernateRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon for authorization
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeExecute,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to hibernate postgres addon")
			}

			// Trigger hibernate action
			err = h.postgresAddonService.TriggerHibernate(ctx, id, hibernateRequest.GetHibernate())
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdAddonsPostgresIdActionsHibernatePost200Response{
				Message: ptr.To("Hibernate action initiated successfully"),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) ListBackups(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			// Get postgres addon for authorization
			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
			}

			// Authorization check
			allowed, accessErr := h.authzClient.AuthorizeResourceAccessRequest(auth.AuthorizationRequest{
				User:            currentUser,
				ResourceType:    auth.PostgresAddon,
				ResourceID:      id,
				ResourceOwnerID: postgresAddon.UserID,
				Action:          models.ResourceAccessModeList,
			})
			if accessErr != nil {
				return nil, errors.Unauthorized("failed to check authorization: %s", accessErr.Error())
			}
			if !allowed {
				return nil, errors.Forbidden("insufficient permissions to list postgres addon backups")
			}

			// List backups
			backups, err := h.postgresAddonService.ListBackups(ctx, id)
			if err != nil {
				return nil, err
			}

			return openapi.PostgresBackupList{
				Items: presenters.PresentPostgresBackupList(backups),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}
