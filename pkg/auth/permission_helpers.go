package auth

import (
	"context"
	stderrors "errors"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
)

func CheckServicePermission(permissions PermissionService, ctx context.Context, domain, resource, resourceID, action string) *errors.ServiceError {
	if permissions == nil {
		return nil
	}
	err := permissions.Check(ctx, domain, resource, resourceID, action)
	if err == nil {
		return nil
	}
	if stderrors.Is(err, ErrUnauthenticated) {
		return nil
	}
	return errors.Forbidden("insufficient permissions")
}
