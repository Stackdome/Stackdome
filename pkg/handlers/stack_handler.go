package handlers

import (
	"context"
	"net/http"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/auth"
	"github.com/Stackdome/stackdome/pkg/errors"
	"github.com/Stackdome/stackdome/pkg/handlers/validation"
	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
	"github.com/Stackdome/stackdome/pkg/services"
	"github.com/gorilla/mux"
	"k8s.io/utils/ptr"
)

type StackHandlerSpec struct {
	StackService         services.StackService
	StackResourceService services.StackResourceService
	StackReleaseService  services.StackReleaseService
	ImageBuildService    services.ImageBuildService
	LoggingService       services.LoggingService
	MetricsService       services.MetricsService
	TeamService          services.TeamService
	Logger               logger.Logger
}

type stackHandler struct {
	stackService         services.StackService
	stackResourceService services.StackResourceService
	stackReleaseService  services.StackReleaseService
	imageBuildService    services.ImageBuildService
	loggingService       services.LoggingService
	metricsService       services.MetricsService
	teamService          services.TeamService
	logger               logger.Logger
}

func NewStackHandler(spec StackHandlerSpec) *stackHandler {
	return &stackHandler{
		stackResourceService: spec.StackResourceService,
		stackService:         spec.StackService,
		stackReleaseService:  spec.StackReleaseService,
		imageBuildService:    spec.ImageBuildService,
		loggingService:       spec.LoggingService,
		metricsService:       spec.MetricsService,
		teamService:          spec.TeamService,
		logger:               spec.Logger,
	}
}

func (h *stackHandler) presentWithReleases(ctx context.Context, stack *models.Stack) (openapi.Stack, *errors.ServiceError) {
	refs, err := h.stackReleaseService.InternalGetReleaseRefs(ctx, []*models.Stack{stack})
	if err != nil {
		return openapi.Stack{}, err
	}
	return presenters.PresentStack(stack, refs[stack.ID]), nil
}

func (h *stackHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, err := h.stackService.GetStack(r.Context(), id)
			if err != nil {
				return nil, err
			}
			return h.presentWithReleases(r.Context(), obj)
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackHandler) GetTopology(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, err := h.stackService.GetStackTopology(r.Context(), id)
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackTopology(obj), nil
		},
	}
	handleGet(w, r, cfg)
}

func (h *stackHandler) ListConnections(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, err := h.stackService.ListStackConnections(r.Context(), id)
			if err != nil {
				return nil, err
			}
			return openapi.StackConnectionList{
				Items: presenters.PresentStackConnections(obj),
				Total: ptr.To(int32(len(obj))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackHandler) CreateConnection(w http.ResponseWriter, r *http.Request) {
	var connection openapi.StackConnection
	cfg := &handlerConfig{
		MarshalInto: &connection,
		Validate: func() *errors.ServiceError {
			return validation.ValidateStackConnection(&connection)
		},
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			obj, err := h.stackService.CreateStackConnection(r.Context(), id, presenters.ConvertStackConnection(&connection))
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackConnection(obj), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackHandler) CreateVolume(w http.ResponseWriter, r *http.Request) {
	var volume openapi.Volume
	cfg := &handlerConfig{
		MarshalInto: &volume,
		Validate:    validation.ValidateVolume(&volume),
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			obj, err := h.stackService.CreateStackVolume(r.Context(), stackID, presenters.ConvertVolume(&volume))
			if err != nil {
				return nil, err
			}
			return presenters.PresentVolume(obj, true), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackHandler) UpdateConnection(w http.ResponseWriter, r *http.Request) {
	var connection openapi.StackConnection
	cfg := &handlerConfig{
		MarshalInto: &connection,
		Validate: func() *errors.ServiceError {
			return validation.ValidateStackConnection(&connection)
		},
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			connectionID := mux.Vars(r)["connection_id"]
			obj, err := h.stackService.UpdateStackConnection(r.Context(), stackID, connectionID, presenters.ConvertStackConnection(&connection))
			if err != nil {
				return nil, err
			}
			return presenters.PresentStackConnection(obj), nil
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackHandler) DeleteConnection(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			stackID := mux.Vars(r)["id"]
			connectionID := mux.Vars(r)["connection_id"]
			if err := h.stackService.DeleteStackConnection(r.Context(), stackID, connectionID); err != nil {
				return nil, err
			}
			return nil, nil
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *stackHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			orgID := mux.Vars(r)["org_id"]

			if _, serr := h.stackService.GetStack(ctx, stackID); serr != nil {
				return nil, serr
			}

			loggingParams, pErr := services.NewLoggingParams(r.URL.Query())
			if pErr != nil {
				return nil, errors.MalformedRequest("invalid logging query params: %s", pErr.Error())
			}

			logStreamer, err := h.loggingService.StreamLogsForStack(ctx, orgID, stackID, loggingParams)
			if err != nil {
				return nil, errors.GeneralError("failed to get logs: %s", err.Error())
			}
			return logStreamer, nil
		},
	}
	handleServerSideStream(w, r, cfg)
}

func (h *stackHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			stackID := mux.Vars(r)["id"]
			orgID := mux.Vars(r)["org_id"]

			if _, serr := h.stackService.GetStack(ctx, stackID); serr != nil {
				return nil, serr
			}

			stream := r.URL.Query().Get("stream") == queryValueTrue
			if stream {
				streamer, err := h.metricsService.StreamMetricsForStack(ctx, orgID, stackID)
				if err != nil {
					return nil, errors.GeneralError("failed to stream metrics: %s", err.Error())
				}
				return streamer, nil
			}
			res, err := h.metricsService.GetMetricsForStack(ctx, orgID, stackID)
			if err != nil {
				return nil, errors.GeneralError("failed to get metrics: %s", err.Error())
			}
			return presenters.PresentResourceMetrics(res), nil
		},
	}
	handleStreamOrGet(w, r, cfg)
}

func (h *stackHandler) ListByOrgID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			objs, serr := h.stackService.ListStacksForCurrentUser(r.Context(), orgID)
			if serr != nil {
				return nil, serr
			}
			refs, serr := h.stackReleaseService.InternalGetReleaseRefs(r.Context(), objs)
			if serr != nil {
				return nil, serr
			}
			return openapi.StackList{
				Items: presenters.PresentStackList(objs, refs),
				Total: ptr.To(int32(len(objs))),
			}, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackHandler) ListByTeamName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			objs, serr := h.stackService.GetStacksByTeamID(ctx, teamID)
			if serr != nil {
				return nil, serr
			}
			refs, serr := h.stackReleaseService.InternalGetReleaseRefs(ctx, objs)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.StackList{
				Items: presenters.PresentStackList(objs, refs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}

func (h *stackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStackShell(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			convertedObject := presenters.ConvertStack(&ws)
			identity := auth.GetIdentityFromCtx(ctx)
			if identity == nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = identity.UserID
			convertedObject.StackResources = nil
			convertedObject.Volumes = nil
			convertedObject.Connections = nil

			obj, serr := h.stackService.CreateStack(ctx, convertedObject)
			if serr != nil {
				h.logger.Errorf("failed to create stack: %v", serr)
				return nil, serr
			}
			return h.presentWithReleases(ctx, obj)
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *stackHandler) Update(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStackShell(&ws),
		func() (interface{}, *errors.ServiceError) {
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

			convertedObject := presenters.ConvertStack(&ws)
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = identity.UserID
			convertedObject.StackResources = nil
			convertedObject.Volumes = nil
			convertedObject.Connections = nil

			obj, serr := h.stackService.UpdateStackShell(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return h.presentWithReleases(ctx, obj)
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *stackHandler) Apply(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStack(&ws),
		func() (interface{}, *errors.ServiceError) {
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

			convertedObject := presenters.ConvertStack(&ws)
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = identity.UserID

			// Full object update including children resources.
			obj, serr := h.stackService.UpdateStack(ctx, id, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return h.presentWithReleases(ctx, obj)
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

// ApplyByName is the name-addressed variant of Apply (kubectl-style upsert):
// stack identity comes from the body's name (unique per team), not the URL.
// Responds 201 when the stack was created, 200 when it was updated.
func (h *stackHandler) ApplyByName(w http.ResponseWriter, r *http.Request) {
	var ws openapi.Stack
	created := false
	cfg := &handlerConfig{
		&ws,
		validation.ValidateStack(&ws),
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			identity := auth.GetIdentityFromCtx(ctx)
			if identity == nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}

			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}

			convertedObject := presenters.ConvertStack(&ws)
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = identity.UserID

			obj, wasCreated, serr := h.stackService.ApplyStack(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}
			created = wasCreated
			return h.presentWithReleases(ctx, obj)
		},
		handleError,
	}
	handleWithDynamicStatus(w, r, cfg, func() int {
		if created {
			return http.StatusCreated
		}
		return http.StatusOK
	})
}

func (h *stackHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			id := mux.Vars(r)["id"]
			stack, serr := h.stackService.DeleteStack(r.Context(), id)
			if serr != nil {
				return nil, serr
			}
			return h.presentWithReleases(r.Context(), stack)
		},
	}
	handleDelete(w, r, cfg, http.StatusAccepted)
}
