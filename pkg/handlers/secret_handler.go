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

type SecretHandlerSpec struct {
	SecretService services.SecretService
	TeamService   services.TeamService
	Logger        logger.Logger
}

type secretHandler struct {
	secretService services.SecretService
	teamService   services.TeamService
	logger        logger.Logger
}

func NewSecretHandler(spec SecretHandlerSpec) *secretHandler {
	return &secretHandler{
		secretService: spec.SecretService,
		teamService:   spec.TeamService,
		logger:        spec.Logger,
	}
}

func (h *secretHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			id := mux.Vars(r)["id"]
			obj, err := h.secretService.GetByID(ctx, id)
			if err != nil {
				return nil, err
			}
			return presenters.PresentSecret(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *secretHandler) ListByTeamID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			objs, serr := h.secretService.ListByTeamID(ctx, teamID)
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
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			convertedObject := presenters.ConvertSecret(&secret)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = currentUser.ID

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
			convertedObject := presenters.ConvertSecret(&secret)
			convertedObject.OrganisationID = orgID
			convertedObject.UserID = currentUser.ID

			obj, serr := h.secretService.Update(ctx, id, convertedObject)
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
			serr := h.secretService.Delete(ctx, id)
			if serr != nil {
				return nil, serr
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}
