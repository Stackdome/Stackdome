package workspaceprovisionrequests

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

func WorkspaceNamespaceFor(user *models.User) string {
	return fmt.Sprintf("%s-workspace", WPRClusterObjectName(user))
}

func WPRClusterObjectName(user *models.User) string {
	sanitizedName := sanitizeName(user.Name)

	objectName := fmt.Sprintf("%s-%d", sanitizedName, user.InternalID)
	// Ensure the object name meets the Kubernetes requirements
	return truncateObjectName(objectName)
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
