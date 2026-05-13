package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/interfaces"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/gorilla/mux"
)

type handlerConfig struct {
	MarshalInto  interface{}
	Validate     validation.Validate
	Action       httpAction
	ErrorHandler errorHandlerFunc
}

type errorHandlerFunc func(ctx context.Context, w http.ResponseWriter, err *errors.ServiceError)
type httpAction func() (interface{}, *errors.ServiceError)

func handleError(ctx context.Context, w http.ResponseWriter, err *errors.ServiceError) {
	logger := logger.GetLoggerFromContext(ctx)
	// If this is a 400 error, its the user's issue, log as info rather than error
	if err.HttpCode >= 400 && err.HttpCode <= 499 {
		logger.Infof(err.Error())
	} else {
		logger.Errorf(err.Error())
	}
	writeJSONResponse(w, err.HttpCode, err.AsOpenapiError())
}

func handle(w http.ResponseWriter, r *http.Request, cfg *handlerConfig, httpStatus int) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		handleError(r.Context(), w, errors.MalformedRequest("Unable to read request body: %s", err))
		return
	}

	emptyBody := len(bytes) == 0
	if !emptyBody {
		err = json.Unmarshal(bytes, &cfg.MarshalInto)
		if err != nil {
			handleError(r.Context(), w, errors.MalformedRequest("Invalid request format: %s", err))
			return
		}
	}

	if cfg.Validate != nil {
		if err := cfg.Validate(); err != nil {
			cfg.ErrorHandler(r.Context(), w, err)
			return
		}
	}

	result, serviceErr := cfg.Action()

	switch {
	case serviceErr != nil:
		cfg.ErrorHandler(r.Context(), w, serviceErr)
	default:
		writeJSONResponse(w, httpStatus, result)
	}

}

func handleDelete(w http.ResponseWriter, r *http.Request, cfg *handlerConfig, httpStatus int) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	if cfg.Validate != nil {
		if err := cfg.Validate(); err != nil {
			cfg.ErrorHandler(r.Context(), w, err)
			return
		}
	}

	result, serviceErr := cfg.Action()

	switch {
	case serviceErr != nil:
		cfg.ErrorHandler(r.Context(), w, serviceErr)
	default:
		writeJSONResponse(w, httpStatus, result)
	}

}

func handleGet(w http.ResponseWriter, r *http.Request, cfg *handlerConfig) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	result, serviceErr := cfg.Action()
	switch {
	case serviceErr == nil:
		writeJSONResponse(w, http.StatusOK, result)
	default:
		cfg.ErrorHandler(r.Context(), w, serviceErr)
	}
}

func handleStreamOrGet(w http.ResponseWriter, r *http.Request, cfg *handlerConfig) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	result, serviceErr := cfg.Action()
	if serviceErr != nil {
		cfg.ErrorHandler(r.Context(), w, serviceErr)
		return
	}
	streamable, ok := result.(interfaces.ServerSideStreamable)
	if ok {
		internalStreamHandler(w, r, streamable, cfg)
		return
	}
	writeJSONResponse(w, http.StatusOK, result)
}

func handleServerSideStream(w http.ResponseWriter, r *http.Request, cfg *handlerConfig) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	result, serviceErr := cfg.Action()
	if serviceErr != nil {
		cfg.ErrorHandler(r.Context(), w, serviceErr)
		return
	}

	streamable, ok := result.(interfaces.ServerSideStreamable)
	if !ok {
		cfg.ErrorHandler(r.Context(), w, errors.InternalServerError("Invalid response type"))
		return
	}

	internalStreamHandler(w, r, streamable, cfg)
}

func internalStreamHandler(w http.ResponseWriter, r *http.Request, streamable interfaces.ServerSideStreamable, cfg *handlerConfig) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		cfg.ErrorHandler(r.Context(), w, errors.InternalServerError("Streaming unsupported"))
		return
	}

	streamer, err := streamable.Stream(r.Context())
	if err != nil {
		cfg.ErrorHandler(r.Context(), w, errors.InternalServerError("Unable to create streamer: %s", err))
		return
	}
	addStreamHeaders(w)
	for {
		select {
		case <-r.Context().Done():
			// Client disconnected
			return
		case streamObject, ok := <-streamer:
			if !ok {
				// Stream ended
				return
			}
			if err := streamObject.Error(); err != nil {
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
				flusher.Flush()
				return
			}
			data := streamObject.Data()
			if data != "" {
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}
}

func handleList(w http.ResponseWriter, r *http.Request, cfg *handlerConfig) {
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = handleError
	}

	results, serviceError := cfg.Action()
	if serviceError != nil {
		cfg.ErrorHandler(r.Context(), w, serviceError)
		return
	}

	writeJSONResponse(w, http.StatusOK, results)
}

func writeJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", "Authorization")

	w.WriteHeader(code)

	if payload != nil {
		response, _ := json.Marshal(payload)
		_, _ = w.Write(response)
	}
}

// Prepare a 'list' of non-db-backed resources
func determineListRange(obj interface{}, page int, size int) (list []interface{}, total int) {
	items := reflect.ValueOf(obj)
	total = items.Len()
	low := (page - 1) * size
	high := low + size
	if low < 0 || low >= total || high >= total {
		low = 0
		high = total
	}
	for i := low; i < high; i++ {
		list = append(list, items.Index(i).Interface())
	}

	return list, total
}

func addStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.WriteHeader(http.StatusOK)
}

func parsePaginationParams(r *http.Request) stores.PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	return stores.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

func resolveTeamID(r *http.Request, teamService services.TeamService) (string, *errors.ServiceError) {
	orgID := mux.Vars(r)["org_id"]
	teamName := mux.Vars(r)["team_name"]
	if teamName == "" {
		return "", errors.BadRequest("team_name is required")
	}
	team, serr := teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
	if serr != nil {
		return "", serr
	}
	return team.ID, nil
}
