package clusterresource

import "fmt"

type ClusterResourceError struct {
	Message string
	Err     error
}

func (e *ClusterResourceError) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Err.Error())
}

func newError(message string, err error) *ClusterResourceError {
	return &ClusterResourceError{
		Message: message,
		Err:     err,
	}
}
