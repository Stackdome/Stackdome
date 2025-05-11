package auth

import (
	"fmt"

	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type AuthorizableResource string

const (
	// User
	User AuthorizableResource = "User"

	// Stack
	Stack AuthorizableResource = "Stack"

	// StackResource
	StackResource AuthorizableResource = "StackResource"

	// ImageBuildBuild
	ImageBuild AuthorizableResource = "ImageBuild"

	// Volume
	Volume AuthorizableResource = "Volume"

	// WorkspaceUser
	WorkspaceUser AuthorizableResource = "WorkspaceUser"

	// Organisation
	Organisation AuthorizableResource = "Organisation"

	// Cluster
	Cluster AuthorizableResource = "Cluster"
)

type AuthorizationClient interface {
	AuthorizeResourceAccess(user *models.User, resourceType AuthorizableResource, resourceID string, resourceOwnerID string, action models.ResourceAccessMode) (bool, error)
}

type AuthorizerBackend interface {
	CheckPermission(subject, orgID, resource, action, resourceOwnerID string) (bool, error)
}

type authorizationHandler struct {
	authorizerBackend AuthorizerBackend
}

type AuthorizationHanlderSpec struct {
	AuthorizerBackend AuthorizerBackend
}

func NewAuthorizationHandler(spec AuthorizationHanlderSpec) AuthorizationClient {
	return &authorizationHandler{
		authorizerBackend: spec.AuthorizerBackend,
	}
}

func (a *authorizationHandler) AuthorizeResourceAccess(
	user *models.User,
	resourceType AuthorizableResource,
	resourceID string,
	resourceOwnerID string,
	action models.ResourceAccessMode) (bool, error) {
	var resourceURL string
	if resourceID != "" {
		resourceURL = fmt.Sprintf("/%s/%s", resourceType, resourceID)
	} else {
		// For all other resources, the resource name is the same as the resourceType
		// We also set the resourceOwnerID to the current user ID.
		// Ex POST (Create) a new workspace / any other resource.
		resourceURL = fmt.Sprintf("/%s", resourceType)
	}

	allowed, accessErr := a.authorizerBackend.CheckPermission(user.ID, user.OrganisationID, resourceURL, action.String(), resourceOwnerID)
	if accessErr != nil {
		return false, fmt.Errorf("failed to authorize request: %w", accessErr)
	}
	return allowed, nil
}
