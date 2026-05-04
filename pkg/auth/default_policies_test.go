package auth

import "testing"

func TestDefaultPoliciesNotEmpty(t *testing.T) {
	policies := DefaultPolicies()
	if len(policies) == 0 {
		t.Fatal("DefaultPolicies returned empty list")
	}
	for i, p := range policies {
		if len(p) != 4 {
			t.Errorf("policy %d has %d fields, expected 4", i, len(p))
		}
		if p[0] == "" || p[2] == "" || p[3] == "" {
			t.Errorf("policy %d has empty required fields: %v", i, p)
		}
	}
}

func TestDefaultPoliciesContainsAllRoles(t *testing.T) {
	policies := DefaultPolicies()
	roles := map[string]bool{
		"OrgMember": false,
		"Viewer":    false,
		"Developer": false,
		"OrgAdmin":  false,
	}
	for _, p := range policies {
		if _, exists := roles[p[0]]; exists {
			roles[p[0]] = true
		}
	}
	for role, found := range roles {
		if !found {
			t.Errorf("role %s not found in default policies", role)
		}
	}
}

func TestLoadDefaultPolicies(t *testing.T) {
	loaded := make([][]string, 0)
	addPolicy := func(subject, domain, resource, action string) error {
		loaded = append(loaded, []string{subject, domain, resource, action})
		return nil
	}
	if err := LoadDefaultPolicies(addPolicy); err != nil {
		t.Fatalf("LoadDefaultPolicies failed: %v", err)
	}
	if len(loaded) != len(DefaultPolicies()) {
		t.Errorf("loaded %d policies, expected %d", len(loaded), len(DefaultPolicies()))
	}
}
