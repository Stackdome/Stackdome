package clusterresource

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

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

func sanitizeName(name string) string {
	// Replace spaces and special characters with hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	sanitized := reg.ReplaceAllString(name, "-")

	// Remove leading and trailing hyphens
	sanitized = strings.TrimPrefix(sanitized, "-")
	sanitized = strings.TrimSuffix(sanitized, "-")

	return strings.ToLower(sanitized)
}

func truncateObjectName(name string) string {
	// Truncate the object name if it exceeds the maximum length
	maxLength := 63
	if len(name) > maxLength {
		name = name[:maxLength]
	}

	name = strings.TrimSuffix(name, "-")
	return name
}

func WrapErrAsServiceError(err error) *errors.ServiceError {
	if err == nil {
		return nil
	}
	return errors.GeneralError(err.Error())
}

type DBClusterService interface {
	GetClusterForOrg(ctx context.Context, orgID string) (*models.Cluster, *errors.ServiceError)
}

type DBUserService interface {
	Get(ctx context.Context, userID string) (*models.User, *errors.ServiceError)
}

type DBWorkspaceUserService interface {
	GetWorkspaceUser(ctx context.Context, userID string) (*models.WorkspaceUser, *errors.ServiceError)
}
