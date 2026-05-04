package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/presenters"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type PostgresAddonHandlerSpec struct {
	PostgresAddonService services.PostgresAddonService
	TeamService          services.TeamService
	Logger               logger.Logger
}

type postgresAddonHandler struct {
	postgresAddonService services.PostgresAddonService
	teamService          services.TeamService
	logger               logger.Logger
}

func NewPostgresAddonHandler(spec PostgresAddonHandlerSpec) *postgresAddonHandler {
	return &postgresAddonHandler{
		postgresAddonService: spec.PostgresAddonService,
		teamService:          spec.TeamService,
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
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}

			currentUser, uerr := auth.GetCurrentUserFromCtx(ctx)
			if uerr != nil {
				return nil, errors.Unauthorized("failed to fetch current user")
			}

			postgresAddon := presenters.ConvertPostgresAddon(&apiPostgresAddon)
			postgresAddon.OrganisationID = orgID
			postgresAddon.TeamID = teamID
			postgresAddon.UserID = currentUser.ID

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
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}

			postgresAddons, err := h.postgresAddonService.ListPostgresAddonsByTeamID(ctx, teamID)
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

			postgresAddon, err := h.postgresAddonService.GetPostgresAddon(ctx, id)
			if err != nil {
				return nil, err
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

			postgresAddon := presenters.ConvertPostgresAddon(&apiPostgresAddon)
			postgresAddon.ID = id
			postgresAddon.UserID = currentUser.ID

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

			postgresAddon, err := h.postgresAddonService.DeletePostgresAddon(ctx, id)
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

			err := h.postgresAddonService.TriggerBackup(ctx, id)
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

			err := h.postgresAddonService.TriggerFence(ctx, id, fenceRequest.GetFence())
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

			err := h.postgresAddonService.TriggerHibernate(ctx, id, hibernateRequest.GetHibernate())
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

func (h *postgresAddonHandler) GetCredentials(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			database := mux.Vars(r)["database"]
			superuser := r.URL.Query().Get("superuser") == "true"

			creds, err := h.postgresAddonService.GetCredentials(ctx, id, database, superuser)
			if err != nil {
				return nil, err
			}

			return creds, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}
