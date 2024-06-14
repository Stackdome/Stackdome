package validation

import "github.com/ashishmax31/soradev-api-server/pkg/api/openapi"

func ValidateWorkspaceProvisionRequest(in *openapi.WorkspaceProvisionRequest) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrgId", "org_id"),
		validateEmpty(in, "UserId", "user_id"),
		validateEmpty(in, "State", "state"),
		validateEmpty(in, "Message", "message"),
		validateNotEmpty(in, "SshPublicKey", "sshPublicKey"),
	})
}

func ValidateWorkspaceProvisionRequestStatusUpdate(in *openapi.WorkspaceProvisionRequest) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrgId", "orgId"),
		validateEmpty(in, "UserId", "userId"),
		validateEmpty(in, "SshPublicKey", "sshPublicKey"),
		validateEmpty(in, "State", "state"),
		validateEmpty(in, "Message", "message"),
		validateNotEmpty(in, "Status", "status"),
	})
}
