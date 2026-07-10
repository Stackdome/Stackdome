package handlers

import (
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
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

func (h *postgresAddonHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			addons, serr := h.postgresAddonService.ListPostgresAddonsForCurrentUser(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			return openapi.PostgresAddonList{
				Items: presenters.PresentPostgresAddonList(addons),
			}, nil
		},
	}
	handleList(w, r, cfg)
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

			identity := auth.GetIdentityFromCtx(ctx)
			if identity == nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}

			postgresAddon := presenters.ConvertPostgresAddon(&apiPostgresAddon)
			postgresAddon.ID = id
			postgresAddon.UserID = identity.UserID
			postgresAddon.TeamID = teamID

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
	var backupRequest openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPostRequest
	cfg := &handlerConfig{
		MarshalInto: &backupRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			err := h.postgresAddonService.TriggerBackup(ctx, id)
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsBackupPost202Response{
				Message: ptr.To("Backup initiated successfully"),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusAccepted)
}

func (h *postgresAddonHandler) Fence(w http.ResponseWriter, r *http.Request) {
	var fenceRequest openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePostRequest
	cfg := &handlerConfig{
		MarshalInto: &fenceRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			err := h.postgresAddonService.TriggerFence(ctx, id, fenceRequest.GetFence())
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsFencePost200Response{
				Message: ptr.To("Fence action initiated successfully"),
			}, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *postgresAddonHandler) Hibernate(w http.ResponseWriter, r *http.Request) {
	var hibernateRequest openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePostRequest
	cfg := &handlerConfig{
		MarshalInto: &hibernateRequest,
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]

			err := h.postgresAddonService.TriggerHibernate(ctx, id, hibernateRequest.GetHibernate())
			if err != nil {
				return nil, err
			}

			return openapi.ApiV1OrganizationsOrgIdTeamsTeamNameAddonsPostgresIdActionsHibernatePost200Response{
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
			superuser := r.URL.Query().Get("superuser") == queryValueTrue

			creds, err := h.postgresAddonService.GetCredentials(ctx, id, database, superuser)
			if err != nil {
				return nil, err
			}

			return creds, nil
		},
	}
	handle(w, r, cfg, http.StatusOK)
}
