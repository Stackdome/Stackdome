package auth

import "github.com/Stackdome/stackdome/pkg/models"

const resourceAddonsAny = ResourceAddons + "/*"

func DefaultPolicies() [][]string {
	return [][]string{
		// OrgMember: auto-assigned, read-only on org infrastructure and users.
		{models.OrgMemberRole.String(), "*", "clusters", ActionList},
		{models.OrgMemberRole.String(), "*", "clusters/*", ActionRead},
		{models.OrgMemberRole.String(), "*", "image-registries", ActionList},
		{models.OrgMemberRole.String(), "*", "image-registries/*", ActionRead},
		{models.OrgMemberRole.String(), "*", "registry-credentials", ActionList},
		{models.OrgMemberRole.String(), "*", "registry-credentials/*", ActionRead},
		{models.OrgMemberRole.String(), "*", "git-integrations", ActionList},
		{models.OrgMemberRole.String(), "*", "git-integrations/*", ActionRead},
		{models.OrgMemberRole.String(), "*", "orgs/*", ActionRead},
		{models.OrgMemberRole.String(), "*", "projects", ActionList},
		{models.OrgMemberRole.String(), "*", "projects/*", ActionRead},
		// Allow org members to list and read user resources so that they can see who else is in the org,
		// but not allow them to manage users in any way.
		{models.OrgMemberRole.String(), "*", "users", ActionList},
		{models.OrgMemberRole.String(), "*", "users/*", ActionRead},

		// Viewer: read-only on project resources
		{models.ViewerRole.String(), "*", ResourceStacks, ActionList},
		{models.ViewerRole.String(), "*", "stacks/*", ActionRead},
		{models.ViewerRole.String(), "*", ResourceSecrets, ActionList},
		{models.ViewerRole.String(), "*", "secrets/*", ActionRead},
		{models.ViewerRole.String(), "*", ResourceVolumes, ActionList},
		{models.ViewerRole.String(), "*", "volumes/*", ActionRead},
		{models.ViewerRole.String(), "*", resourceAddonsAny, ActionList},
		{models.ViewerRole.String(), "*", "addons/*/*", ActionRead},
		{models.ViewerRole.String(), "*", ResourceObjectStores, ActionList},
		{models.ViewerRole.String(), "*", "object-stores/*", ActionRead},
		{models.ViewerRole.String(), "*", "workspace-users/*", ActionRead},

		// Developer: CRUD on project resources
		{models.DeveloperRole.String(), "*", ResourceStacks, ActionList},
		{models.DeveloperRole.String(), "*", ResourceStacks, ActionCreate},
		{models.DeveloperRole.String(), "*", "stacks/*", "*"},
		{models.DeveloperRole.String(), "*", ResourceSecrets, ActionList},
		{models.DeveloperRole.String(), "*", ResourceSecrets, ActionCreate},
		{models.DeveloperRole.String(), "*", "secrets/*", "*"},
		{models.DeveloperRole.String(), "*", ResourceVolumes, ActionList},
		{models.DeveloperRole.String(), "*", ResourceVolumes, ActionCreate},
		{models.DeveloperRole.String(), "*", "volumes/*", "*"},
		{models.DeveloperRole.String(), "*", resourceAddonsAny, ActionList},
		{models.DeveloperRole.String(), "*", resourceAddonsAny, ActionCreate},
		{models.DeveloperRole.String(), "*", "addons/*/*", "*"},
		{models.DeveloperRole.String(), "*", ResourceObjectStores, ActionList},
		{models.DeveloperRole.String(), "*", ResourceObjectStores, ActionCreate},
		{models.DeveloperRole.String(), "*", "object-stores/*", "*"},
		{models.DeveloperRole.String(), "*", "workspace-users", ActionCreate},
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
