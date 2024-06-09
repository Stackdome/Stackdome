package validation

import (
	"github.com/ashishmax31/soradev-api-server/pkg/api/openapi"
	"github.com/ashishmax31/soradev-api-server/pkg/errors"
)

func ValidateUserCreate(in *openapi.UserCreateRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
		validateNotEmpty(in, "Name", "name"),
		func() *errors.ServiceError {
			if len(in.GetPassword()) < 8 {
				return errors.Validation("min password length should be 8")
			}
			return nil
		},
	})
}

func ValidateWorkspaceProvisionRequest(in *openapi.WorkspaceProvisionRequest) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrgId", "orgId"),
		validateEmpty(in, "UserId", "userId"),
		validateNotEmpty(in, "SshPublicKey", "sshPublicKey"),
	})
}

func ValidateWorkspaceProvisionRequestStatusUpdate(in *openapi.WorkspaceProvisionRequest) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrgId", "orgId"),
		validateEmpty(in, "UserId", "userId"),
		validateEmpty(in, "SshPublicKey", "sshPublicKey"),
		validateNotEmpty(in, "Status", "status"),
	})
}

func ValidateUserLogin(in *openapi.LoginRequest) Validate {
	return ValidateAll([]Validate{
		validateNotEmpty(in, "Email", "email"),
		validateNotEmpty(in, "Password", "password"),
	})
}
