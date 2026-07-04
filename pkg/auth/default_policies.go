package auth

import "github.com/Stackdome/stackdome/pkg/models"

func DefaultPolicies() [][]string {
	return [][]string{
		// OrgMember: auto-assigned, read-only on org infrastructure and users.
		{models.OrgMemberRole.String(), "*", "clusters", "list"},
		{models.OrgMemberRole.String(), "*", "clusters/*", "read"},
		{models.OrgMemberRole.String(), "*", "image-registries", "list"},
		{models.OrgMemberRole.String(), "*", "image-registries/*", "read"},
		{models.OrgMemberRole.String(), "*", "registry-credentials", "list"},
		{models.OrgMemberRole.String(), "*", "registry-credentials/*", "read"},
		{models.OrgMemberRole.String(), "*", "git-integrations", "list"},
		{models.OrgMemberRole.String(), "*", "git-integrations/*", "read"},
		{models.OrgMemberRole.String(), "*", "orgs/*", "read"},
		{models.OrgMemberRole.String(), "*", "teams", "list"},
		{models.OrgMemberRole.String(), "*", "teams/*", "read"},
		// Allow org members to list and read user resources so that they can see who else is in the org,
		// but not allow them to manage users in any way.
		{models.OrgMemberRole.String(), "*", "users", "list"},
		{models.OrgMemberRole.String(), "*", "users/*", "read"},

		// Viewer: read-only on team resources
		{models.ViewerRole.String(), "*", "stacks", "list"},
		{models.ViewerRole.String(), "*", "stacks/*", "read"},
		{models.ViewerRole.String(), "*", "secrets", "list"},
		{models.ViewerRole.String(), "*", "secrets/*", "read"},
		{models.ViewerRole.String(), "*", "volumes", "list"},
		{models.ViewerRole.String(), "*", "volumes/*", "read"},
		{models.ViewerRole.String(), "*", "addons/*", "list"},
		{models.ViewerRole.String(), "*", "addons/*/*", "read"},
		{models.ViewerRole.String(), "*", "object-stores", "list"},
		{models.ViewerRole.String(), "*", "object-stores/*", "read"},
		{models.ViewerRole.String(), "*", "workspace-users/*", "read"},

		// Developer: CRUD on team resources
		{models.DeveloperRole.String(), "*", "stacks", "list"},
		{models.DeveloperRole.String(), "*", "stacks", "create"},
		{models.DeveloperRole.String(), "*", "stacks/*", "*"},
		{models.DeveloperRole.String(), "*", "secrets", "list"},
		{models.DeveloperRole.String(), "*", "secrets", "create"},
		{models.DeveloperRole.String(), "*", "secrets/*", "*"},
		{models.DeveloperRole.String(), "*", "volumes", "list"},
		{models.DeveloperRole.String(), "*", "volumes", "create"},
		{models.DeveloperRole.String(), "*", "volumes/*", "*"},
		{models.DeveloperRole.String(), "*", "addons/*", "list"},
		{models.DeveloperRole.String(), "*", "addons/*", "create"},
		{models.DeveloperRole.String(), "*", "addons/*/*", "*"},
		{models.DeveloperRole.String(), "*", "object-stores", "list"},
		{models.DeveloperRole.String(), "*", "object-stores", "create"},
		{models.DeveloperRole.String(), "*", "object-stores/*", "*"},
		{models.DeveloperRole.String(), "*", "workspace-users", "create"},
		{models.DeveloperRole.String(), "*", "workspace-users/*", "*"},

		// OrgAdmin: full access to everything under the org.
		{models.OrgAdminRole.String(), "*", "*", "*"},
	}
}

func LoadDefaultPolicies(addPolicy func(subject, domain, resource, action string) error) error {
	for _, p := range DefaultPolicies() {
		if err := addPolicy(p[0], p[1], p[2], p[3]); err != nil {
			return err
		}
	}
	return nil
}
