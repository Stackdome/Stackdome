package errors

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/golang/glog"
)

const (

	// InvalidToken occurs when a token is invalid (generally, not found in the database)
	ErrorInvalidToken ServiceErrorCode = 1

	// Forbidden occurs when a user has been blacklisted
	ErrorForbidden ServiceErrorCode = 4

	// Conflict occurs when a database constraint is violated
	ErrorConflict ServiceErrorCode = 6

	// NotFound occurs when a record is not found in the database
	ErrorNotFound ServiceErrorCode = 7

	// Validation occurs when an object fails validation
	ErrorValidation ServiceErrorCode = 8

	// General occurs when an error fails to match any other error code
	ErrorGeneral ServiceErrorCode = 9

	// NotImplemented occurs when an API REST method is not implemented in a handler
	ErrorNotImplemented ServiceErrorCode = 10

	// Unauthorized occurs when the requester is not authorized to perform the specified action
	ErrorUnauthorized ServiceErrorCode = 11

	// Unauthenticated occurs when the provided credentials cannot be validated
	ErrorUnauthenticated ServiceErrorCode = 15

	// MalformedRequest occurs when the request body cannot be read
	ErrorMalformedRequest ServiceErrorCode = 17

	// Bad Request
	ErrorBadRequest ServiceErrorCode = 21

	// Gone
	ErrorGone ServiceErrorCode = 24

	// Too Many Requests
	ErrorTooManyRequests ServiceErrorCode = 25

	// Unprocessable Entity
	ErrorUnprocessableEntity ServiceErrorCode = 26
)

type ServiceErrorCode int

type ServiceErrors []ServiceError

func Find(code ServiceErrorCode) (bool, *ServiceError) {
	for _, err := range Errors() {
		if err.Code == code {
			return true, &err
		}
	}
	return false, nil
}

func Errors() ServiceErrors {
	return ServiceErrors{
		ServiceError{ErrorInvalidToken, "Invalid token provided", http.StatusForbidden},
		ServiceError{ErrorForbidden, "Forbidden to perform this action", http.StatusForbidden},
		ServiceError{ErrorConflict, "An entity with the specified unique values already exists", http.StatusConflict},
		ServiceError{ErrorNotFound, "Resource not found", http.StatusNotFound},
		ServiceError{ErrorValidation, "General validation failure", http.StatusBadRequest},
		ServiceError{ErrorGeneral, "Unspecified error", http.StatusInternalServerError},
		ServiceError{ErrorNotImplemented, "HTTP Method not implemented for this endpoint", http.StatusMethodNotAllowed},
		ServiceError{ErrorUnauthorized, "Account is unauthorized to perform this action", http.StatusForbidden},
		ServiceError{ErrorUnauthenticated, "Account authentication could not be verified", http.StatusUnauthorized},
		ServiceError{ErrorMalformedRequest, "Unable to read request body", http.StatusBadRequest},
		ServiceError{ErrorBadRequest, "Bad request", http.StatusBadRequest},
		ServiceError{ErrorGone, "Access to the target resource is no longer available", http.StatusGone},
		ServiceError{ErrorTooManyRequests, "Too many requests", http.StatusTooManyRequests},
		ServiceError{ErrorUnprocessableEntity, "Unable to process request entity", http.StatusUnprocessableEntity},
	}
}

type ServiceError struct {
	// Code is the numeric and distinct ID for the error
	Code ServiceErrorCode
	// Reason is the context-specific reason the error was generated
	Reason string
	// HttopCode is the HttpCode associated with the error when the error is returned as an API response
	HttpCode int
}

// Reason can be a string with format verbs, which will be replace by the specified values
func New(code ServiceErrorCode, reason string, values ...interface{}) *ServiceError {
	// If the code isn't defined, use the general error code
	var err *ServiceError
	exists, err := Find(code)
	if !exists {
		glog.Errorf("Undefined error code used: %d", code)
		err = &ServiceError{ErrorGeneral, "Unspecified error", 500}
	}

	// If the reason is unspecified, use the default
	if reason != "" {
		err.Reason = fmt.Sprintf(reason, values...)
	}

	return err
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("error: %s", e.Reason)
}

func (e *ServiceError) AsError() error {
	return fmt.Errorf(e.Error())
}

func (e *ServiceError) WithPrefix(format string, a ...interface{}) *ServiceError {
	prefix := fmt.Sprintf(format, a...)
	e.Reason = fmt.Sprintf("%s: %s", prefix, e.Reason)
	return e
}

func (e *ServiceError) Is404() bool {
	return e.Code == NotFound("").Code
}

func (e *ServiceError) IsConflict() bool {
	return e.Code == Conflict("").Code
}

func (e *ServiceError) IsForbidden() bool {
	return e.Code == Forbidden("").Code
}

func (e *ServiceError) AsOpenapiError() openapi.Error {
	return openapi.Error{
		Kind:   openapi.PtrString("Error"),
		Id:     openapi.PtrString(strconv.Itoa(int(e.Code))),
		Code:   openapi.PtrString(fmt.Sprintf("%d", e.Code)),
		Reason: openapi.PtrString(e.Reason),
	}
}

func NotFound(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotFound, reason, values...)
}

func GeneralError(reason string, values ...interface{}) *ServiceError {
	return New(ErrorGeneral, reason, values...)
}

func Unauthorized(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthorized, reason, values...)
}

func Unauthenticated(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthenticated, reason, values...)
}

func Forbidden(reason string, values ...interface{}) *ServiceError {
	return New(ErrorForbidden, reason, values...)
}

func NotImplemented(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotImplemented, reason, values...)
}

func Conflict(reason string, values ...interface{}) *ServiceError {
	return New(ErrorConflict, reason, values...)
}

func Validation(reason string, values ...interface{}) *ServiceError {
	return New(ErrorValidation, reason, values...)
}

func MalformedRequest(reason string, values ...interface{}) *ServiceError {
	return New(ErrorMalformedRequest, reason, values...)
}

func BadRequest(reason string, values ...interface{}) *ServiceError {
	return New(ErrorBadRequest, reason, values...)
}

func Gone(reason string, values ...interface{}) *ServiceError {
	return New(ErrorGone, reason, values...)
}

func TooManyRequests(reason string, values ...interface{}) *ServiceError {
	return New(ErrorTooManyRequests, reason, values...)
}

func UnprocessableEntity(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnprocessableEntity, reason, values...)
}
