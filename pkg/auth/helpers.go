package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Stackdome/stackdome/pkg/errors"
)

func handleError(w http.ResponseWriter, code errors.ServiceErrorCode, reason string) {
	err := errors.New(code, "%s", reason)

	writeJSONResponse(w, err.HttpCode, err.AsOpenapiError())
}

func writeJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if payload != nil {
		response, _ := json.Marshal(payload)
		_, _ = w.Write(response)
	}
}
