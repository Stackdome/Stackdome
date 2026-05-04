package auth

func DefaultPolicies() [][]string {
	return [][]string{
		// OrgMember: auto-assigned, read-only on org infrastructure
		{"OrgMember", "*", "clusters", "list"},
		{"OrgMember", "*", "clusters/*", "read"},
		{"OrgMember", "*", "image-registries", "list"},
		{"OrgMember", "*", "image-registries/*", "read"},
		{"OrgMember", "*", "orgs/*", "read"},

		// Viewer: read-only on team resources
		{"Viewer", "*", "stacks", "list"},
		{"Viewer", "*", "stacks/*", "read"},
		{"Viewer", "*", "secrets", "list"},
		{"Viewer", "*", "secrets/*", "read"},
		{"Viewer", "*", "volumes", "list"},
		{"Viewer", "*", "volumes/*", "read"},
		{"Viewer", "*", "addons/*", "list"},
		{"Viewer", "*", "addons/*/*", "read"},
		{"Viewer", "*", "object-stores", "list"},
		{"Viewer", "*", "object-stores/*", "read"},
		{"Viewer", "*", "workspace-users/*", "read"},

		// Developer: CRUD on team resources
		{"Developer", "*", "stacks", "list"},
		{"Developer", "*", "stacks", "create"},
		{"Developer", "*", "stacks/*", "*"},
		{"Developer", "*", "secrets", "list"},
		{"Developer", "*", "secrets", "create"},
		{"Developer", "*", "secrets/*", "*"},
		{"Developer", "*", "volumes", "list"},
		{"Developer", "*", "volumes", "create"},
		{"Developer", "*", "volumes/*", "*"},
		{"Developer", "*", "addons/*", "list"},
		{"Developer", "*", "addons/*", "create"},
		{"Developer", "*", "addons/*/*", "*"},
		{"Developer", "*", "object-stores", "list"},
		{"Developer", "*", "object-stores", "create"},
		{"Developer", "*", "object-stores/*", "*"},
		{"Developer", "*", "workspace-users", "create"},
		{"Developer", "*", "workspace-users/*", "*"},

		// OrgAdmin: full access to everything
		{"OrgAdmin", "*", "*", "*"},
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
