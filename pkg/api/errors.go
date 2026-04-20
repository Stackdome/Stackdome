package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"

	"github.com/golang/glog"
)

// SendNotFound sends a 404 response with some details about the non existing resource.
func SendNotFound(w http.ResponseWriter, r *http.Request) {
	// Set the content type:
	w.Header().Set("Content-Type", "application/json")

	// Prepare the body:
	id := "404"
	reason := fmt.Sprintf(
		"The requested resource '%s' doesn't exist",
		r.URL.Path,
	)
	body := Error{
		Type:   ErrorType,
		ID:     id,
		Reason: reason,
	}
	data, err := json.Marshal(body)
	if err != nil {
		SendPanic(w, r)
		return
	}

	// Send the response:
	w.WriteHeader(http.StatusNotFound)
	_, err = w.Write(data)
	if err != nil {
		err = fmt.Errorf("can't send response body for request '%s'", r.URL.Path)
		glog.Error(err)
		return
	}
}

func SendUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "application/json")

	// Prepare the body:
	apiError := errors.Unauthorized("%s", message)
	data, err := json.Marshal(apiError)
	if err != nil {
		SendPanic(w, r)
		return
	}

	// Send the response:
	w.WriteHeader(http.StatusUnauthorized)
	_, err = w.Write(data)
	if err != nil {
		err = fmt.Errorf("can't send response body for request '%s'", r.URL.Path)
		glog.Error(err)
		return
	}
}

var panicBody []byte

// SendPanic sends a panic error response to the client, but it doesn't end the process.
func SendPanic(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(panicBody)
	if err != nil {
		err = fmt.Errorf(
			"can't send panic response for request '%s': %s",
			r.URL.Path,
			err.Error(),
		)
		glog.Error(err)
	}
}
