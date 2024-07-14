package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/handlers/validation"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
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
	log := logger.NewLogger(ctx)
	// If this is a 400 error, its the user's issue, log as info rather than error
	if err.HttpCode >= 400 && err.HttpCode <= 499 {
		log.Infof(err.Error())
	} else {
		log.Errorf(err.Error())
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
