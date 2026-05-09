package auth

import "testing"

func TestSplitScope(t *testing.T) {
	tests := []struct {
		input   string
		wantNil bool
		wantRes string
		wantAct string
	}{
		{"stacks:read", false, "stacks", "read"},
		{"addons/postgres:*", false, "addons/postgres", "*"},
		{"*:*", false, "*", "*"},
		{"stacks:*", false, "stacks", "*"},
		{"stacks", true, "", ""},
		{"", true, "", ""},
		{"a:b:c", true, "", ""},
		{":", false, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitScope(tt.input)
			if tt.wantNil {
				if result != nil {
					t.Errorf("splitScope(%q) = %v, want nil", tt.input, result)
				}
				return
			}
			if result == nil {
				t.Fatalf("splitScope(%q) = nil, want [%q, %q]", tt.input, tt.wantRes, tt.wantAct)
			}
			if result[0] != tt.wantRes || result[1] != tt.wantAct {
				t.Errorf("splitScope(%q) = [%q, %q], want [%q, %q]", tt.input, result[0], result[1], tt.wantRes, tt.wantAct)
			}
		})
	}
}

func TestValidateScope(t *testing.T) {
	tests := []struct {
		scope string
		want  bool
	}{
		// Global wildcard
		{"*:*", true},

		// Exact resource:action matches
		{"stacks:read", true},
		{"stacks:write", true},
		{"stacks:delete", true},
		{"stacks:list", true},
		{"stacks:create", true},
		{"stacks:logs", true},
		{"stacks:exec", true},
		{"secrets:read", true},
		{"secrets:create", true},
		{"volumes:delete", true},
		{"clusters:list", true},
		{"orgs:write", true},
		{"object-stores:read", true},
		{"users:read", true},
		{"image-builds:create", true},
		{"domains:list", true},

		// Resource wildcard (all actions)
		{"stacks:*", true},
		{"secrets:*", true},
		{"volumes:*", true},
		{"clusters:*", true},
		{"orgs:*", true},
		{"addons/postgres:*", true},
		{"object-stores:*", true},
		{"users:*", true},
		{"image-builds:*", true},
		{"addons:*", true},
		{"domains:*", true},

		// Addon-specific actions
		{"addons/postgres:read", true},
		{"addons/postgres:logs", true},
		{"addons/postgres:exec", true},

		// Parent scope covers child resource
		{"addons:read", true},
		{"addons:logs", true},
		{"addons:exec", true},

		// Invalid resource
		{"nonexistent:read", false},
		{"stakcs:read", false},
		{"STACKS:read", false},

		// Invalid action on valid resource
		{"stacks:fly", false},
		{"secrets:exec", false},
		{"secrets:logs", false},
		{"clusters:exec", false},
		{"volumes:logs", false},

		// Malformed scopes
		{"", false},
		{"stacks", false},
		{"read", false},
		{":read", false},
		{"stacks:", false},
		{"a:b:c", false},
	}

	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			got := ValidateScope(tt.scope)
			if got != tt.want {
				t.Errorf("ValidateScope(%q) = %v, want %v", tt.scope, got, tt.want)
			}
		})
	}
}

func TestScopeCovers(t *testing.T) {
	tests := []struct {
		name          string
		grantedScoped string
		resource      string
		action        string
		approved      bool
	}{
		// Global wildcard covers everything
		{"global wildcard covers any resource/action", "*:*", "stacks", "read", true},
		{"global wildcard covers addons", "*:*", "addons/postgres", "delete", true},

		// Exact match
		{"exact match stacks:read", "stacks:read", "stacks", "read", true},
		{"exact match secrets:create", "secrets:create", "secrets", "create", true},
		{"exact match addons/postgres:write", "addons/postgres:write", "addons/postgres", "write", true},

		// Action wildcard on same resource
		{"stacks:* covers stacks:read", "stacks:*", "stacks", "read", true},
		{"stacks:* covers stacks:delete", "stacks:*", "stacks", "delete", true},
		{"stacks:* covers stacks:logs", "stacks:*", "stacks", "logs", true},
		{"secrets:* covers secrets:create", "secrets:*", "secrets", "create", true},
		{"addons/postgres:* covers addons/postgres:read", "addons/postgres:*", "addons/postgres", "read", true},

		// Parent resource covers child
		{"addons:* covers addons/postgres:read", "addons:*", "addons/postgres", "read", true},
		{"addons:* covers addons/postgres:write", "addons:*", "addons/postgres", "write", true},
		{"addons:read covers addons/postgres:read", "addons:read", "addons/postgres", "read", true},

		// Wrong action
		{"stacks:read does not cover stacks:write", "stacks:read", "stacks", "write", false},
		{"stacks:read does not cover stacks:delete", "stacks:read", "stacks", "delete", false},
		{"secrets:list does not cover secrets:create", "secrets:list", "secrets", "create", false},
		{"addons:read does not cover addons/postgres:write", "addons:read", "addons/postgres", "write", false},

		// Wrong resource
		{"stacks:read does not cover secrets:read", "stacks:read", "secrets", "read", false},
		{"stacks:* does not cover secrets:read", "stacks:*", "secrets", "read", false},
		{"secrets:* does not cover stacks:read", "secrets:*", "stacks", "read", false},

		// Child does not cover parent
		{"addons/postgres:* does not cover addons:read", "addons/postgres:*", "addons", "read", false},

		// Unrelated resources
		{"volumes:read does not cover clusters:read", "volumes:read", "clusters", "read", false},
		{"orgs:* does not cover users:read", "orgs:*", "users", "read", false},

		// Malformed granted scope
		{"malformed scope returns false", "stacks", "stacks", "read", false},
		{"empty scope returns false", "", "stacks", "read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ScopeCovers(tt.grantedScoped, tt.resource, tt.action)
			if got != tt.approved {
				t.Errorf("ScopeCovers(%q, %q, %q) = %v, want %v", tt.grantedScoped, tt.resource, tt.action, got, tt.approved)
			}
		})
	}
}
