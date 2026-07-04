package validation

import "github.com/Stackdome/stackdome/pkg/api/openapi"

func ValidateWorkspaceUser(in *openapi.WorkspaceUser) Validate {
	return ValidateAll([]Validate{
		validateEmpty(in, "Id", "id"),
		validateEmpty(in, "OrgId", "org_id"),
		validateEmpty(in, "UserId", "user_id"),
		validateEmpty(in, "State", "state"),
		validateEmpty(in, "Message", "message"),
	})
}
