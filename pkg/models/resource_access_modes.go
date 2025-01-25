package models

import "net/http"

type ResourceAccessMode string

func (r ResourceAccessMode) String() string {
	return string(r)
}

const (
	// read
	ResourceAccessModeRead ResourceAccessMode = "read"

	// List
	ResourceAccessModeList ResourceAccessMode = "list"

	// write
	ResourceAccessModeWrite ResourceAccessMode = "write"

	// delete
	ResourceAccessModeDelete ResourceAccessMode = "delete"

	// update
	ResourceAccessModeUpdate ResourceAccessMode = "update"

	// create
	ResourceAccessModeCreate ResourceAccessMode = "create"

	// execute
	ResourceAccessModeExecute ResourceAccessMode = "execute"
)

func MapHttpMethodToResourceAccessMode(method string) ResourceAccessMode {
	switch method {
	case http.MethodGet:
		return ResourceAccessModeRead
	case http.MethodPost:
		return ResourceAccessModeCreate
	case http.MethodPut:
		return ResourceAccessModeUpdate
	case http.MethodDelete:
		return ResourceAccessModeDelete
	case http.MethodHead:
		return ResourceAccessModeRead
	case http.MethodPatch:
		return ResourceAccessModeUpdate
	// Default to update for all other methods to be safe
	default:
		return ResourceAccessModeUpdate
	}
}
