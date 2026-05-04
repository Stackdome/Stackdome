# Teams & RBAC Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Introduce Teams as a grouping layer between Organisation and Users, with team-scoped RBAC (Viewer, Developer, OrgAdmin roles), replacing the broken flat org-level permission model.

**Architecture:** Teams own resources (stacks, secrets, volumes, addons, object-stores, workspace-users). Casbin handles team-level role policies with a single Go-level OrgAdmin fallback. Infrastructure resources (clusters, image registries) remain org-scoped with OrgMember read access. Per-resource ownership grants are removed — team roles are sufficient.

**Tech Stack:** Go, Casbin (RBAC), PostgreSQL/GORM, gorilla/mux

**Branch:** `feat/teams-rbac-redesign`
**Spec:** `docs/superpowers/specs/2026-05-04-teams-and-rbac-redesign-design.md`

## Progress

| Chunk | Status | Summary |
|-------|--------|---------|
| 1. Foundation | DONE | See `docs/superpowers/plans/2026-05-04-chunk1-completed.md` for details and deviations |
| 2. Team service + permissions | DONE | TeamService created, PermissionService simplified, GrantAccess/RevokeAllAccess removed from all services |
| 3. Service layer migration | NOT STARTED | |
| 4. API layer | NOT STARTED | |

**Key deviations from plan (Chunk 1):**
- `ensureDefaultPlatformAdminUser` / `ensureDefaultAdminUser` removed entirely (not replaced) — spec says no bootstrap users, first user signs up normally
- Default org migration still runs but is dead code — `GetDefaultOrg` references remain and will be cleaned up in Chunk 4 (signup flow rewrite)

---

## Chunk 1: Foundation (models, stores, Casbin, migrations)

### Step 1.1: Update role constants in user model

- [ ] **Step 1.1: Replace role constants in `pkg/models/user.go`**

Remove the old role constants and add the new ones. Remove PlatformAdminRole entirely. OrgAdmin replaces OrganisationAdmin. The User role is removed — regular members have no role stored on the user model (their team-level roles are in team_memberships). The `Role` field on User is now only for org-level designation (OrgAdmin or empty).

**File:** `pkg/models/user.go`

Replace the role constants block and update `ClusterAccessRules`:

```go
package models

import (
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	CreatedForUserLabel = "user.stackdome.io/id"
)

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	OrgAdminRole Role = "OrgAdmin"
	DeveloperRole Role = "Developer"
	ViewerRole    Role = "Viewer"
	OrgMemberRole Role = "OrgMember"
)

type User struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	Organisation   *Organisation `gorm:"foreignKey:OrganisationID"`
	Role           Role
	OrganisationID string
	GithubID       *string `gorm:"column:github_id;uniqueIndex"`
	AvatarURL      *string `gorm:"column:avatar_url"`
}

func (u *User) IsOrgAdmin() bool {
	return u.Role == OrgAdminRole
}

func (u *User) ClusterAccessRules() []rbacv1.PolicyRule {
	return []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "pods/log"},
			Verbs:     []string{"get", "list"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods/exec", "pods/portforward"},
			Verbs:     []string{"create"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"services"},
			Verbs:     []string{"get"},
		},
	}
}
```

Note: The `DefaultUser` field is removed since there is no default platform admin anymore.

After editing, run:
```bash
go fmt ./pkg/models/user.go
```

---

### Step 1.2: Update Identity in auth package

- [ ] **Step 1.2: Update `pkg/auth/identity.go` to remove PlatformAdmin, add OrgAdmin/OrgMember checks**

**File:** `pkg/auth/identity.go`

```go
package auth

import "context"

type IdentityKey string

const identityContextKey IdentityKey = "AuthIdentity"

type AuthMethod string

const (
	AuthMethodJWT      AuthMethod = "jwt"
	AuthMethodAPIToken AuthMethod = "api_token"
	AuthMethodOAuth    AuthMethod = "oauth"
)

type Identity struct {
	UserID      string
	OrgID       string
	Role        string
	AuthMethod  AuthMethod
	TokenID     string
	TokenScopes []string
	ResourceIDs []string
}

func (i *Identity) IsOrgAdmin() bool {
	return i.Role == "OrgAdmin"
}

func SetIdentityInContext(ctx context.Context, identity *Identity) context.Context {
	return context.WithValue(ctx, identityContextKey, identity)
}

func GetIdentityFromCtx(ctx context.Context) *Identity {
	identity, ok := ctx.Value(identityContextKey).(*Identity)
	if !ok {
		return nil
	}
	return identity
}
```

After editing, run:
```bash
go fmt ./pkg/auth/identity.go
```

---

### Step 1.3: Fix Casbin model matcher

- [ ] **Step 1.3: Update `pkg/resourceaccess/casbin_model.conf` to use `keyMatch` and rename `org` to `domain`**

**File:** `pkg/resourceaccess/casbin_model.conf`

```ini
[request_definition]
r = subject, domain, resource, action

[policy_definition]
p = subject, domain, resource, action

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = ((g(r.subject, p.subject, r.domain) && keyMatch(r.resource, p.resource) && (p.domain == "*" || r.domain == p.domain)) \
    || (r.subject == p.subject && keyMatch(r.resource, p.resource) && r.domain == p.domain)) \
    && (r.action == p.action || p.action == "*")
```

This replaces `keyMatch2` with `keyMatch` to fix the broken pattern matching. The field naming changes from `org` to `domain` (cosmetic only -- Casbin reads positionally).

---

### Step 1.4: Create Team model

- [ ] **Step 1.4: Create `pkg/models/team.go` with Team and TeamMembership models**

**File:** `pkg/models/team.go` (new file)

```go
package models

import "time"

type Team struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name           string `gorm:"not null" json:"name"`
	OrganisationID string `gorm:"not null" json:"organisation_id"`
	DefaultTeam    bool   `gorm:"not null;default:false" json:"default_team"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TeamMembership struct {
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	TeamID    string `gorm:"not null" json:"team_id"`
	UserID    string `gorm:"not null" json:"user_id"`
	Role      Role   `gorm:"not null" json:"role"`
	CreatedAt time.Time

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
```

After creating, run:
```bash
go fmt ./pkg/models/team.go
```

---

### Step 1.5: Create Team store interface

- [ ] **Step 1.5: Create `pkg/stores/team_store.go` with TeamStore interface**

**File:** `pkg/stores/team_store.go` (new file)

```go
package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type TeamStore interface {
	Create(ctx context.Context, team *models.Team) (*models.Team, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.Team, *errors.ServiceError)
	GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError)
	ListByOrgID(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError)
	Update(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
	GetDefaultTeamForOrg(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError)
}
```

After creating, run:
```bash
go fmt ./pkg/stores/team_store.go
```

---

### Step 1.6: Create Team store implementation

- [ ] **Step 1.6: Create `pkg/stores/pgstore/team_store.go` with PostgreSQL implementation**

**File:** `pkg/stores/pgstore/team_store.go` (new file)

```go
package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type dbTeamStore struct {
	sessionFactory db.SessionFactory
}

type TeamStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewTeamStore(spec TeamStoreSpec) stores.TeamStore {
	return &dbTeamStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbTeamStore) Create(ctx context.Context, team *models.Team) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(team).Error; err != nil {
		return nil, errors.GeneralError("failed to create team: %s", err.Error())
	}
	return s.GetByID(ctx, team.ID)
}

func (s *dbTeamStore) GetByID(ctx context.Context, id string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("id = ?", id).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch team: %s", err.Error())
	}
	return &team, nil
}

func (s *dbTeamStore) GetByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("organisation_id = ? AND name = ?", orgID, name).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team '%s' not found in organisation", name)
		}
		return nil, errors.GeneralError("failed to fetch team: %s", err.Error())
	}
	return &team, nil
}

func (s *dbTeamStore) ListByOrgID(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var teams []*models.Team
	if err := grm.Where("organisation_id = ?", orgID).Order("created_at ASC").Find(&teams).Error; err != nil {
		return nil, errors.GeneralError("failed to list teams: %s", err.Error())
	}
	return teams, nil
}

func (s *dbTeamStore) Update(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Model(&models.Team{}).Where("id = ?", id).Updates(team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update team: %s", err.Error())
	}
	return s.GetByID(ctx, id)
}

func (s *dbTeamStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.Team{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("team with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete team: %s", err.Error())
	}
	return nil
}

func (s *dbTeamStore) GetDefaultTeamForOrg(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var team models.Team
	if err := grm.Where("organisation_id = ? AND default_team = ?", orgID, true).First(&team).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("default team not found for organisation '%s'", orgID)
		}
		return nil, errors.GeneralError("failed to fetch default team: %s", err.Error())
	}
	return &team, nil
}
```

After creating, run:
```bash
go fmt ./pkg/stores/pgstore/team_store.go
```

---

### Step 1.7: Create TeamMembership store interface

- [ ] **Step 1.7: Create `pkg/stores/team_membership_store.go` with TeamMembershipStore interface**

**File:** `pkg/stores/team_membership_store.go` (new file)

```go
package stores

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

type TeamMembershipStore interface {
	Create(ctx context.Context, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError)
	GetByID(ctx context.Context, id string) (*models.TeamMembership, *errors.ServiceError)
	GetByTeamAndUser(ctx context.Context, teamID, userID string) (*models.TeamMembership, *errors.ServiceError)
	ListByTeamID(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListByUserID(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError)
	Update(ctx context.Context, id string, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError)
	Delete(ctx context.Context, id string) *errors.ServiceError
}
```

After creating, run:
```bash
go fmt ./pkg/stores/team_membership_store.go
```

---

### Step 1.8: Create TeamMembership store implementation

- [ ] **Step 1.8: Create `pkg/stores/pgstore/team_membership_store.go` with PostgreSQL implementation**

**File:** `pkg/stores/pgstore/team_membership_store.go` (new file)

```go
package pgstore

import (
	"context"

	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"gorm.io/gorm"
)

type dbTeamMembershipStore struct {
	sessionFactory db.SessionFactory
}

type TeamMembershipStoreSpec struct {
	SessionFactory db.SessionFactory
}

func NewTeamMembershipStore(spec TeamMembershipStoreSpec) stores.TeamMembershipStore {
	return &dbTeamMembershipStore{
		sessionFactory: spec.SessionFactory,
	}
}

func (s *dbTeamMembershipStore) Create(ctx context.Context, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Create(membership).Error; err != nil {
		return nil, errors.GeneralError("failed to create team membership: %s", err.Error())
	}
	return s.GetByID(ctx, membership.ID)
}

func (s *dbTeamMembershipStore) GetByID(ctx context.Context, id string) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.TeamMembership
	if err := grm.Preload("Team").Preload("User").Where("id = ?", id).First(&membership).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team membership with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to fetch team membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbTeamMembershipStore) GetByTeamAndUser(ctx context.Context, teamID, userID string) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var membership models.TeamMembership
	if err := grm.Preload("Team").Preload("User").Where("team_id = ? AND user_id = ?", teamID, userID).First(&membership).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("membership not found for user in team")
		}
		return nil, errors.GeneralError("failed to fetch team membership: %s", err.Error())
	}
	return &membership, nil
}

func (s *dbTeamMembershipStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("User").Where("team_id = ?", teamID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list team memberships: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) ListByUserID(ctx context.Context, userID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("Team").Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) ListByUserIDAndOrgID(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var memberships []*models.TeamMembership
	if err := grm.Preload("Team").
		Joins("JOIN teams ON teams.id = team_memberships.team_id").
		Where("team_memberships.user_id = ? AND teams.organisation_id = ?", userID, orgID).
		Find(&memberships).Error; err != nil {
		return nil, errors.GeneralError("failed to list memberships for user in org: %s", err.Error())
	}
	return memberships, nil
}

func (s *dbTeamMembershipStore) Update(ctx context.Context, id string, membership *models.TeamMembership) (*models.TeamMembership, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Model(&models.TeamMembership{}).Where("id = ?", id).Updates(map[string]interface{}{
		"role": membership.Role,
	}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("team membership with id '%s' not found", id)
		}
		return nil, errors.GeneralError("failed to update team membership: %s", err.Error())
	}
	return s.GetByID(ctx, id)
}

func (s *dbTeamMembershipStore) Delete(ctx context.Context, id string) *errors.ServiceError {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Where("id = ?", id).Delete(&models.TeamMembership{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NotFound("team membership with id '%s' not found", id)
		}
		return errors.GeneralError("failed to delete team membership: %s", err.Error())
	}
	return nil
}
```

After creating, run:
```bash
go fmt ./pkg/stores/pgstore/team_membership_store.go
```

---

### Step 1.9: Create default_policies.go

- [ ] **Step 1.9: Create `pkg/auth/default_policies.go` with all base role policies**

This file defines the base Casbin policies loaded at startup. Policies use `*` for the domain field so they apply across all teams/orgs -- access scoping is handled by the grouping policies (which tie a user+role to a specific team/org).

**File:** `pkg/auth/default_policies.go` (new file)

```go
package auth

// DefaultPolicies returns the base Casbin role policies loaded at startup.
// Format: [subject, domain, resource, action]
// Domain is "*" because policies define what a role CAN do;
// grouping policies (g) define WHERE a user has that role.
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

// LoadDefaultPolicies adds all base role policies to the policy manager.
// Called during environment initialization.
func LoadDefaultPolicies(addPolicy func(subject, domain, resource, action string) error) error {
	for _, p := range DefaultPolicies() {
		if err := addPolicy(p[0], p[1], p[2], p[3]); err != nil {
			return err
		}
	}
	return nil
}
```

After creating, run:
```bash
go fmt ./pkg/auth/default_policies.go
```

---

### Step 1.10: Create database migration for teams table

- [ ] **Step 1.10: Create `pkg/db/migrations/202605040001_create_teams_table.go`**

**File:** `pkg/db/migrations/202605040001_create_teams_table.go` (new file)

```go
package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040001_create_teams_table",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				CREATE TABLE IF NOT EXISTS teams (
					id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					name            VARCHAR(63) NOT NULL,
					organisation_id UUID NOT NULL REFERENCES organisations(id),
					default_team    BOOLEAN NOT NULL DEFAULT false,
					created_at      TIMESTAMP NOT NULL DEFAULT now(),
					updated_at      TIMESTAMP NOT NULL DEFAULT now(),
					UNIQUE(name, organisation_id)
				);
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create teams table: %w", err)
			}

			indexSQL := `CREATE INDEX IF NOT EXISTS idx_teams_organisation_id ON teams(organisation_id);`
			if err := tx.Exec(indexSQL).Error; err != nil {
				return fmt.Errorf("failed to create teams organisation_id index: %w", err)
			}

			return nil
		},
	}
}
```

After creating, run:
```bash
go fmt ./pkg/db/migrations/202605040001_create_teams_table.go
```

---

### Step 1.11: Create database migration for team_memberships table

- [ ] **Step 1.11: Create `pkg/db/migrations/202605040002_create_team_memberships_table.go`**

**File:** `pkg/db/migrations/202605040002_create_team_memberships_table.go` (new file)

```go
package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func createTeamMembershipsTable() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040002_create_team_memberships_table",
		Migrate: func(tx *gorm.DB) error {
			sql := `
				CREATE TABLE IF NOT EXISTS team_memberships (
					id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
					team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
					user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					role       VARCHAR(50) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT now(),
					UNIQUE(team_id, user_id)
				);
			`
			if err := tx.Exec(sql).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships table: %w", err)
			}

			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_team_memberships_team_id ON team_memberships(team_id);`).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships team_id index: %w", err)
			}
			if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_team_memberships_user_id ON team_memberships(user_id);`).Error; err != nil {
				return fmt.Errorf("failed to create team_memberships user_id index: %w", err)
			}

			return nil
		},
	}
}
```

After creating, run:
```bash
go fmt ./pkg/db/migrations/202605040002_create_team_memberships_table.go
```

---

### Step 1.12: Create database migration to add team_id to resource tables

- [ ] **Step 1.12: Create `pkg/db/migrations/202605040003_add_team_id_to_resources.go`**

This adds `team_id` to all team-scoped resource tables. The column is added as nullable first, then in a future data migration step it will be populated and made NOT NULL. For now we keep it nullable to allow the migration to run on existing data.

**File:** `pkg/db/migrations/202605040003_add_team_id_to_resources.go` (new file)

```go
package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func addTeamIDToResources() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040003_add_team_id_to_resources",
		Migrate: func(tx *gorm.DB) error {
			tables := []string{
				"stacks",
				"secrets",
				"volumes",
				"postgres_addons",
				"object_stores",
				"workspace_users",
			}

			for _, table := range tables {
				addCol := fmt.Sprintf(
					"ALTER TABLE %s ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id);",
					table,
				)
				if err := tx.Exec(addCol).Error; err != nil {
					return fmt.Errorf("failed to add team_id to %s: %w", table, err)
				}

				addIdx := fmt.Sprintf(
					"CREATE INDEX IF NOT EXISTS idx_%s_team_id ON %s(team_id);",
					table, table,
				)
				if err := tx.Exec(addIdx).Error; err != nil {
					return fmt.Errorf("failed to create team_id index on %s: %w", table, err)
				}
			}

			return nil
		},
	}
}
```

After creating, run:
```bash
go fmt ./pkg/db/migrations/202605040003_add_team_id_to_resources.go
```

---

### Step 1.13: Create migration to update user roles

- [ ] **Step 1.13: Create `pkg/db/migrations/202605040004_update_user_roles.go`**

Migrate existing role values and remove the `default_user` column.

**File:** `pkg/db/migrations/202605040004_update_user_roles.go` (new file)

```go
package migrations

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func updateUserRoles() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202605040004_update_user_roles",
		Migrate: func(tx *gorm.DB) error {
			// Migrate OrganisationAdmin and PlatformAdmin to OrgAdmin
			if err := tx.Exec(`UPDATE users SET role = 'OrgAdmin' WHERE role IN ('OrganisationAdmin', 'PlatformAdmin');`).Error; err != nil {
				return fmt.Errorf("failed to migrate admin roles: %w", err)
			}

			// Clear the User role (regular members have empty role)
			if err := tx.Exec(`UPDATE users SET role = '' WHERE role = 'User';`).Error; err != nil {
				return fmt.Errorf("failed to clear User role: %w", err)
			}

			// Remove default_user column
			if err := tx.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS default_user;`).Error; err != nil {
				return fmt.Errorf("failed to drop default_user column: %w", err)
			}

			return nil
		},
	}
}
```

After creating, run:
```bash
go fmt ./pkg/db/migrations/202605040004_update_user_roles.go
```

---

### Step 1.14: Register new migrations

- [ ] **Step 1.14: Add new migrations to `pkg/db/migrations/migrations.go`**

**File:** `pkg/db/migrations/migrations.go`

Add these four entries at the end of the `MigrationList` slice:

```go
	addGitHubOAuthFields(),
	createTeamsTable(),
	createTeamMembershipsTable(),
	addTeamIDToResources(),
	updateUserRoles(),
```

After editing, run:
```bash
go fmt ./pkg/db/migrations/migrations.go
```

---

### Step 1.15: Add TeamID field to all team-scoped resource models

- [ ] **Step 1.15: Add `TeamID string` to Stack, Secret, Volume, PostgresAddon, ObjectStore, WorkspaceUser models**

Add `TeamID string \`gorm:"index"\` json:"team_id"` to each model struct. Exact locations:

**`pkg/models/stack.go`** -- add to the `Stack` struct after `OrganisationID`:
```go
TeamID         string      `gorm:"index" json:"team_id"`
```

**`pkg/models/secret.go`** -- add to the `Secret` struct after `OrganisationID`:
```go
TeamID         string `gorm:"index" json:"team_id"`
```

**`pkg/models/volume.go`** -- add to the `Volume` struct after `OrganisationID`:
```go
TeamID         string           `gorm:"index" json:"team_id"`
```

**`pkg/models/postgres_addon.go`** -- add to the `PostgresAddon` struct after `OrganisationID`:
```go
TeamID         string `gorm:"index" json:"team_id"`
```

**`pkg/models/object_store.go`** -- add to the `ObjectStore` struct after `OrganisationID`:
```go
TeamID         string `gorm:"index" json:"team_id"`
```

**`pkg/models/workspace_user.go`** -- add to the `WorkspaceUser` struct after `OrganisationID`:
```go
TeamID         string `gorm:"index" json:"team_id"`
```

After editing all six files, run:
```bash
go fmt ./pkg/models/...
```

---

### Step 1.16: Remove Organisation Default field and related code

- [ ] **Step 1.16: Remove the `Default` field from Organisation model and `DefaultOrgName` constant**

**File:** `pkg/models/organisation.go`

Remove the `DefaultOrgName` constant and the `Default bool` field from the Organisation struct:

```go
package models

import (
	"time"
)

type Organisation struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name      string
	Domains   []*OrganisationDomain `gorm:"foreignKey:OrganisationID"`
}
```

After editing, run:
```bash
go fmt ./pkg/models/organisation.go
```

---

### Step 1.17: Build verification

- [ ] **Step 1.17: Verify the project compiles after all Chunk 1 changes**

```bash
make binary
```

This will likely surface compilation errors from code still referencing old constants (`PlatformAdminRole`, `UserRole`, `OrganisationAdminRole`, `DefaultOrgName`, `DefaultUser`). Fix each one:

- Any reference to `models.PlatformAdminRole` -> remove or replace with `models.OrgAdminRole`
- Any reference to `models.UserRole` -> remove (empty string for regular members)
- Any reference to `models.OrganisationAdminRole` -> `models.OrgAdminRole`
- Any reference to `identity.IsPlatformAdmin()` -> `identity.IsOrgAdmin()`
- Any reference to `models.DefaultOrgName` -> remove
- Any reference to `user.DefaultUser` -> remove
- Any reference to `GetDefaultUser` -> remove the method and its usages

Fix each file until `make binary` succeeds. Key files that will need fixes:
- `pkg/auth/permission_service.go` (line 44: `identity.IsPlatformAdmin()` call)
- `pkg/services/user_service.go` (references to `PlatformAdminRole`, `UserRole`, `GetDefaultUser`, `OrganisationAdminRole`)
- `cmd/environment/development_environment.go` (`ensureDefaultPlatformAdminUser`, `initializeBaseResourceAccessPolicies`)
- `cmd/environment/test_environment.go` (same)
- `pkg/stores/pgstore/organisations.go` (`GetDefaultOrg`)
- `pkg/auth/jwt_user_claims.go` (if it references old roles)

For now, make minimal edits to fix compilation. The full rewrites happen in Chunks 2-4.

After fixing, run:
```bash
make binary
```

Expected: clean compilation.

---

### Step 1.18: Commit Chunk 1

- [ ] **Step 1.18: Commit all Chunk 1 changes**

```bash
git add pkg/models/team.go \
  pkg/stores/team_store.go \
  pkg/stores/team_membership_store.go \
  pkg/stores/pgstore/team_store.go \
  pkg/stores/pgstore/team_membership_store.go \
  pkg/auth/default_policies.go \
  pkg/db/migrations/202605040001_create_teams_table.go \
  pkg/db/migrations/202605040002_create_team_memberships_table.go \
  pkg/db/migrations/202605040003_add_team_id_to_resources.go \
  pkg/db/migrations/202605040004_update_user_roles.go \
  pkg/db/migrations/migrations.go \
  pkg/models/user.go \
  pkg/models/organisation.go \
  pkg/models/stack.go \
  pkg/models/secret.go \
  pkg/models/volume.go \
  pkg/models/postgres_addon.go \
  pkg/models/object_store.go \
  pkg/models/workspace_user.go \
  pkg/auth/identity.go \
  pkg/resourceaccess/casbin_model.conf

git commit -m "feat: add teams foundation - models, stores, Casbin fix, migrations

Introduce Team and TeamMembership models with stores. Fix Casbin
matcher (keyMatch2 -> keyMatch). Add team_id to all team-scoped
resource models. Create default_policies.go for base role policies.
Update role constants (OrgAdmin replaces OrganisationAdmin/PlatformAdmin).
Remove default org concept."
```

---

## Chunk 2: Team service and permission changes

### Step 2.1: Create the TeamService interface and implementation

- [ ] **Step 2.1: Create `pkg/services/team_service.go`**

This service handles team CRUD, membership management, OrgMember auto-grouping, and OrgAdmin promotion/demotion.

**File:** `pkg/services/team_service.go` (new file)

```go
package services

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ashishmax31/stackdome-api-server/pkg/auth"
	"github.com/ashishmax31/stackdome-api-server/pkg/db"
	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores/pgstore"
)

var teamNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`)

type TeamService interface {
	// Team CRUD
	CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError)
	GetTeam(ctx context.Context, id string) (*models.Team, *errors.ServiceError)
	GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError)
	ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError)
	UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError)
	DeleteTeam(ctx context.Context, id string) *errors.ServiceError
	CreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError)

	// Membership management
	AddMember(ctx context.Context, teamID, userID string, role models.Role) (*models.TeamMembership, *errors.ServiceError)
	RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError
	UpdateMemberRole(ctx context.Context, membershipID string, role models.Role) (*models.TeamMembership, *errors.ServiceError)
	ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError)
	ListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError)

	// OrgAdmin management
	PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	DemoteOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError
	ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError)
}

type TeamServiceSpec struct {
	SessionFactory db.SessionFactory
	PolicyManager  resourceaccess.ResourceAccessPolicyManager
	Permissions    auth.PermissionService
	Logger         logger.Logger
}

type teamService struct {
	teamStore           stores.TeamStore
	membershipStore     stores.TeamMembershipStore
	userStore           stores.UserStore
	policyMgr           resourceaccess.ResourceAccessPolicyManager
	permissions         auth.PermissionService
	logger              logger.Logger
}

func NewTeamService(spec TeamServiceSpec) TeamService {
	return &teamService{
		teamStore: pgstore.NewTeamStore(pgstore.TeamStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		membershipStore: pgstore.NewTeamMembershipStore(pgstore.TeamMembershipStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		userStore: pgstore.NewUserStore(pgstore.UserStoreSpec{
			SessionFactory: spec.SessionFactory,
		}),
		policyMgr:   spec.PolicyManager,
		permissions: spec.Permissions,
		logger:      spec.Logger,
	}
}

func (s *teamService) CreateTeam(ctx context.Context, orgID string, team *models.Team) (*models.Team, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if err := validateTeamName(team.Name); err != nil {
		return nil, err
	}

	team.OrganisationID = orgID
	created, serr := s.teamStore.Create(ctx, team)
	if serr != nil {
		return nil, serr
	}
	return created, nil
}

func (s *teamService) CreateDefaultTeam(ctx context.Context, orgID string) (*models.Team, *errors.ServiceError) {
	team := &models.Team{
		Name:           "default",
		OrganisationID: orgID,
		DefaultTeam:    true,
	}
	created, serr := s.teamStore.Create(ctx, team)
	if serr != nil {
		return nil, serr
	}
	return created, nil
}

func (s *teamService) GetTeam(ctx context.Context, id string) (*models.Team, *errors.ServiceError) {
	return s.teamStore.GetByID(ctx, id)
}

func (s *teamService) GetTeamByOrgAndName(ctx context.Context, orgID, name string) (*models.Team, *errors.ServiceError) {
	return s.teamStore.GetByOrgAndName(ctx, orgID, name)
}

func (s *teamService) ListTeams(ctx context.Context, orgID string) ([]*models.Team, *errors.ServiceError) {
	// Any org member can list teams
	return s.teamStore.ListByOrgID(ctx, orgID)
}

func (s *teamService) UpdateTeam(ctx context.Context, id string, team *models.Team) (*models.Team, *errors.ServiceError) {
	existing, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, existing.OrganisationID, auth.ResourceOrgs, existing.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if team.Name != "" {
		if err := validateTeamName(team.Name); err != nil {
			return nil, err
		}
	}
	return s.teamStore.Update(ctx, id, team)
}

func (s *teamService) DeleteTeam(ctx context.Context, id string) *errors.ServiceError {
	team, serr := s.teamStore.GetByID(ctx, id)
	if serr != nil {
		return serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return permErr
	}
	if team.DefaultTeam {
		return errors.BadRequest("cannot delete the default team")
	}
	return s.teamStore.Delete(ctx, id)
}

func (s *teamService) AddMember(ctx context.Context, teamID, userID string, role models.Role) (*models.TeamMembership, *errors.ServiceError) {
	team, serr := s.teamStore.GetByID(ctx, teamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}

	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("team membership role must be Developer or Viewer")
	}

	membership := &models.TeamMembership{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}
	created, serr := s.membershipStore.Create(ctx, membership)
	if serr != nil {
		return nil, serr
	}

	// Add Casbin grouping for team role
	if err := s.policyMgr.AddGroupingPolicy(userID, string(role), teamID); err != nil {
		s.logger.Errorf("failed to add team role grouping: %s", err.Error())
	}

	// Auto-add OrgMember grouping if this is the user's first team in the org
	s.ensureOrgMemberGrouping(ctx, userID, team.OrganisationID)

	return created, nil
}

func (s *teamService) RemoveMember(ctx context.Context, membershipID string) *errors.ServiceError {
	membership, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return serr
	}
	team, serr := s.teamStore.GetByID(ctx, membership.TeamID)
	if serr != nil {
		return serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	if err := s.membershipStore.Delete(ctx, membershipID); err != nil {
		return err
	}

	// Remove Casbin grouping for team role
	if err := s.policyMgr.RemoveFilteredPolicy(0, membership.UserID, string(membership.Role), membership.TeamID); err != nil {
		s.logger.Errorf("failed to remove team role grouping: %s", err.Error())
	}

	// Check if user has any remaining teams in this org, remove OrgMember if not
	s.cleanupOrgMemberGrouping(ctx, membership.UserID, team.OrganisationID)

	return nil
}

func (s *teamService) UpdateMemberRole(ctx context.Context, membershipID string, role models.Role) (*models.TeamMembership, *errors.ServiceError) {
	existing, serr := s.membershipStore.GetByID(ctx, membershipID)
	if serr != nil {
		return nil, serr
	}
	team, serr := s.teamStore.GetByID(ctx, existing.TeamID)
	if serr != nil {
		return nil, serr
	}
	if permErr := auth.CheckServicePermission(s.permissions, ctx, team.OrganisationID, auth.ResourceOrgs, team.OrganisationID, auth.ActionWrite); permErr != nil {
		return nil, permErr
	}
	if role != models.DeveloperRole && role != models.ViewerRole {
		return nil, errors.BadRequest("team membership role must be Developer or Viewer")
	}

	// Remove old grouping, add new one
	_ = s.policyMgr.RemoveFilteredPolicy(0, existing.UserID, string(existing.Role), existing.TeamID)
	if err := s.policyMgr.AddGroupingPolicy(existing.UserID, string(role), existing.TeamID); err != nil {
		s.logger.Errorf("failed to update team role grouping: %s", err.Error())
	}

	return s.membershipStore.Update(ctx, membershipID, &models.TeamMembership{Role: role})
}

func (s *teamService) ListMembers(ctx context.Context, teamID string) ([]*models.TeamMembership, *errors.ServiceError) {
	return s.membershipStore.ListByTeamID(ctx, teamID)
}

func (s *teamService) ListUserTeams(ctx context.Context, userID, orgID string) ([]*models.TeamMembership, *errors.ServiceError) {
	return s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
}

func (s *teamService) PromoteToOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}
	if user.OrganisationID != orgID {
		return errors.BadRequest("user does not belong to this organisation")
	}

	user.Role = models.OrgAdminRole
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgAdminRole), orgID); err != nil {
		s.logger.Errorf("failed to add OrgAdmin grouping: %s", err.Error())
	}

	return nil
}

func (s *teamService) DemoteOrgAdmin(ctx context.Context, orgID, userID string) *errors.ServiceError {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, orgID, auth.ResourceOrgs, orgID, auth.ActionWrite); permErr != nil {
		return permErr
	}

	// Check this is not the last OrgAdmin
	admins, serr := s.ListOrgAdmins(ctx, orgID)
	if serr != nil {
		return serr
	}
	if len(admins) <= 1 {
		return errors.BadRequest("cannot demote the last OrgAdmin")
	}

	user, serr := s.userStore.GetByID(ctx, userID)
	if serr != nil {
		return serr
	}
	user.Role = ""
	if _, serr := s.userStore.Update(ctx, userID, user); serr != nil {
		return serr
	}

	_ = s.policyMgr.RemoveFilteredPolicy(0, userID, string(models.OrgAdminRole), orgID)

	return nil
}

func (s *teamService) ListOrgAdmins(ctx context.Context, orgID string) ([]*models.User, *errors.ServiceError) {
	return s.userStore.ListByOrgAndRole(ctx, orgID, models.OrgAdminRole)
}

// ensureOrgMemberGrouping adds the OrgMember grouping if not already present.
func (s *teamService) ensureOrgMemberGrouping(ctx context.Context, userID, orgID string) {
	if err := s.policyMgr.AddGroupingPolicy(userID, string(models.OrgMemberRole), orgID); err != nil {
		s.logger.Errorf("failed to add OrgMember grouping: %s", err.Error())
	}
}

// cleanupOrgMemberGrouping removes OrgMember grouping if user has no teams left in org.
func (s *teamService) cleanupOrgMemberGrouping(ctx context.Context, userID, orgID string) {
	memberships, err := s.membershipStore.ListByUserIDAndOrgID(ctx, userID, orgID)
	if err != nil {
		s.logger.Errorf("failed to check remaining memberships: %s", err.Error())
		return
	}
	if len(memberships) == 0 {
		_ = s.policyMgr.RemoveFilteredPolicy(0, userID, string(models.OrgMemberRole), orgID)
	}
}

func validateTeamName(name string) *errors.ServiceError {
	if len(name) == 0 {
		return errors.BadRequest("team name is required")
	}
	if len(name) > 63 {
		return errors.BadRequest("team name must be at most 63 characters")
	}
	if !teamNameRegex.MatchString(name) {
		return errors.BadRequest("team name must be lowercase alphanumeric with hyphens (e.g. 'backend', 'prod-infra')")
	}
	return nil
}
```

After creating, run:
```bash
go fmt ./pkg/services/team_service.go
```

This step requires a `ListByOrgAndRole` method on UserStore. Add it in the next step.

---

### Step 2.2: Add ListByOrgAndRole to UserStore

- [ ] **Step 2.2: Add `ListByOrgAndRole` method to user store interface and implementation**

**File:** `pkg/stores/user_store.go` -- add to the `UserStore` interface:
```go
ListByOrgAndRole(ctx context.Context, orgID string, role models.Role) ([]*models.User, *errors.ServiceError)
```

**File:** `pkg/stores/pgstore/users.go` -- add the implementation method:
```go
func (s *dbUserStore) ListByOrgAndRole(ctx context.Context, orgID string, role models.Role) ([]*models.User, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var users []*models.User
	if err := grm.Where("organisation_id = ? AND role = ?", orgID, string(role)).Find(&users).Error; err != nil {
		return nil, errors.GeneralError("failed to list users by org and role: %s", err.Error())
	}
	return users, nil
}
```

Also add an `Update` method if it doesn't exist:
```go
func (s *dbUserStore) Update(ctx context.Context, id string, user *models.User) (*models.User, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	if err := grm.Model(&models.User{}).Where("id = ?", id).Updates(user).Error; err != nil {
		return nil, errors.GeneralError("failed to update user: %s", err.Error())
	}
	return s.GetByID(ctx, id)
}
```

And add to the UserStore interface:
```go
Update(ctx context.Context, id string, user *models.User) (*models.User, *errors.ServiceError)
```

After editing, run:
```bash
go fmt ./pkg/stores/user_store.go ./pkg/stores/pgstore/users.go
```

---

### Step 2.3: Simplify PermissionService

- [ ] **Step 2.3: Rewrite `pkg/auth/permission_service.go` to remove Grant/Revoke methods and add OrgAdmin fallback**

**File:** `pkg/auth/permission_service.go`

```go
package auth

import (
	"context"
	"fmt"
	"slices"

	"github.com/ashishmax31/stackdome-api-server/pkg/logger"
	"github.com/ashishmax31/stackdome-api-server/pkg/resourceaccess"
	"github.com/ashishmax31/stackdome-api-server/pkg/stores"
)

type PermissionService interface {
	Check(ctx context.Context, domain, resource, resourceID, action string) error
}

var ErrAccessDenied = fmt.Errorf("access denied")
var ErrUnauthenticated = fmt.Errorf("unauthenticated")

type permissionService struct {
	policyMgr resourceaccess.ResourceAccessPolicyManager
	teamStore stores.TeamStore
	logger    logger.Logger
}

type PermissionServiceConfig struct {
	PolicyManager resourceaccess.ResourceAccessPolicyManager
	TeamStore     stores.TeamStore
	Logger        logger.Logger
}

func NewPermissionService(cfg PermissionServiceConfig) PermissionService {
	return &permissionService{
		policyMgr: cfg.PolicyManager,
		teamStore: cfg.TeamStore,
		logger:    cfg.Logger,
	}
}

func (p *permissionService) Check(ctx context.Context, domain, resource, resourceID, action string) error {
	identity := GetIdentityFromCtx(ctx)
	if identity == nil {
		return ErrUnauthenticated
	}

	// API token scope check
	if identity.AuthMethod == AuthMethodAPIToken {
		scopeAllowed := false
		for _, scope := range identity.TokenScopes {
			if ScopeCovers(scope, resource, action) {
				scopeAllowed = true
				break
			}
		}
		if !scopeAllowed {
			return ErrAccessDenied
		}
		if len(identity.ResourceIDs) > 0 && resourceID != "" {
			if !slices.Contains(identity.ResourceIDs, resourceID) {
				return ErrAccessDenied
			}
		}
	}

	// Build Casbin resource string
	casbinResource := resource
	if resourceID != "" {
		casbinResource = fmt.Sprintf("%s/%s", resource, resourceID)
	}

	// Casbin check
	allowed, err := p.policyMgr.CheckPermission(identity.UserID, domain, casbinResource, action)
	if err != nil {
		p.logger.Errorf("permission check failed: %s", err.Error())
		return ErrAccessDenied
	}
	if allowed {
		return nil
	}

	// OrgAdmin fallback for team-scoped resources.
	// OrgAdmin grouping is on orgID, but team resources use teamID as domain.
	// Check if user is OrgAdmin and the domain (teamID) belongs to their org.
	if identity.IsOrgAdmin() && p.teamBelongsToOrg(ctx, domain, identity.OrgID) {
		return nil
	}

	return ErrAccessDenied
}

// teamBelongsToOrg checks if the given teamID belongs to the given orgID.
// Returns false if domain is empty, equals orgID (already an org domain), or lookup fails.
func (p *permissionService) teamBelongsToOrg(ctx context.Context, teamID, orgID string) bool {
	if teamID == "" || teamID == orgID {
		return false
	}
	if p.teamStore == nil {
		return false
	}
	team, err := p.teamStore.GetByID(ctx, teamID)
	if err != nil {
		return false
	}
	return team.OrganisationID == orgID
}
```

After editing, run:
```bash
go fmt ./pkg/auth/permission_service.go
```

---

### Step 2.4: Update permission_helpers.go

- [ ] **Step 2.4: Update `pkg/auth/permission_helpers.go` -- rename parameter from `orgID` to `domain`**

The function signature and parameter names change to match the new terminology. The logic stays the same.

**File:** `pkg/auth/permission_helpers.go`

```go
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
```

After editing, run:
```bash
go fmt ./pkg/auth/permission_helpers.go
```

---

### Step 2.5: Update environment files to use default_policies.go

- [ ] **Step 2.5: Update `cmd/environment/development_environment.go`**

Replace the `initializeBaseResourceAccessPolicies` method body and remove `ensureDefaultPlatformAdminUser` from the init steps.

In the `Init` method, change the initializer steps list:
- Remove `d.initializeBaseResourceAccessPolicies` (replaced with new implementation)
- Remove `d.ensureDefaultPlatformAdminUser` (no more default platform admin)

Replace the `initializeBaseResourceAccessPolicies` method:

```go
func (d *developmentEnvironment) initializeBaseResourceAccessPolicies(ctx context.Context) error {
	d.Logger.Debugf("Initializing base resource access policies")
	if err := auth.LoadDefaultPolicies(d.ResourceAccessPolicyManager.AddPolicy); err != nil {
		return fmt.Errorf("failed to load default policies: %w", err)
	}
	d.Logger.Debugf("Base resource access policies initialized")
	return nil
}
```

Remove the `ensureDefaultPlatformAdminUser` method entirely.

Update the `Init` method's initializer steps to remove `d.ensureDefaultPlatformAdminUser`:

```go
initializerSteps := []func(context.Context) error{
	d.loadEnvAndConfigs,
	d.setupLogger,
	d.setupDatabase,
	d.initializeResourceAccessPolicyManager,
	d.initializePermissionService,
	d.loadServices,
	d.initializeClusterManager,
	d.initializeWorkerManager,
	d.injectClusterResourceServices,
	d.initializeBaseResourceAccessPolicies,
	d.startManagers,
}
```

Update the `initializePermissionService` method to pass TeamStore:

```go
func (d *developmentEnvironment) initializePermissionService(ctx context.Context) error {
	teamStore := pgstore.NewTeamStore(pgstore.TeamStoreSpec{
		SessionFactory: d.DBSession,
	})
	d.PermissionService = auth.NewPermissionService(auth.PermissionServiceConfig{
		PolicyManager: d.ResourceAccessPolicyManager,
		TeamStore:     teamStore,
		Logger:        d.Logger,
	})
	return nil
}
```

Add the TeamService to the `loadServices` method:

```go
teamService := services.NewTeamService(services.TeamServiceSpec{
	SessionFactory: d.DBSession,
	PolicyManager:  d.ResourceAccessPolicyManager,
	Permissions:    d.PermissionService,
	Logger:         d.Logger,
})
```

And add it to the Services struct assignment:

```go
d.Services = Services{
	// ... existing services ...
	TeamService: teamService,
}
```

After editing, run:
```bash
go fmt ./cmd/environment/development_environment.go
```

---

### Step 2.6: Update test environment

- [ ] **Step 2.6: Apply the same changes from Step 2.5 to `cmd/environment/test_environment.go`**

Apply identical changes:
- Remove `ensureDefaultPlatformAdminUser` from init steps
- Replace `initializeBaseResourceAccessPolicies` to use `auth.LoadDefaultPolicies`
- Update `initializePermissionService` to pass TeamStore
- Add TeamService creation in `loadServices`

After editing, run:
```bash
go fmt ./cmd/environment/test_environment.go
```

---

### Step 2.7: Add TeamService to the Services struct

- [ ] **Step 2.7: Add TeamService to `cmd/environment/environment.go` Services struct**

**File:** `cmd/environment/environment.go`

Add to the `Services` struct:
```go
TeamService             services.TeamService
```

After editing, run:
```bash
go fmt ./cmd/environment/environment.go
```

---

### Step 2.8: Update user_service.go signup flow

- [ ] **Step 2.8: Update `pkg/services/user_service.go` to create default team on signup and use new roles**

Update the `Create` method:
- When a user creates an org (no `OrganisationID` provided), set role to `OrgAdminRole`
- Add OrgAdmin and OrgMember Casbin groupings
- Remove reference to `UserRole` as default
- Remove `GetDefaultUser` method
- Remove `CreateOAuthUser`'s reference to `GetDefaultOrg`

The key changes in `Create`:

```go
func (u usersService) Create(ctx context.Context, user *models.User) (*openapi.UserSignupResponse, *errors.ServiceError) {
	u.logger.Infof("Creating user with email: %s", user.Email)
	if len(user.Password) < 8 {
		return nil, errors.BadRequest("password must be at least 8 characters")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		u.logger.Errorf("failed to hash password, %s", err.Error())
		return nil, errors.GeneralError("failed to create user")
	}
	user.Password = string(hashedPassword)

	if user.OrganisationID == "" {
		if user.Organisation == nil {
			return nil, errors.BadRequest("organisation is required")
		}
		createdOrganisation, err := u.organisationService.Create(ctx, user.Organisation)
		if err != nil {
			u.logger.Errorf("failed to create organisation, %s", err.Error())
			return nil, errors.GeneralError("failed to create user")
		}
		user.OrganisationID = createdOrganisation.ID
		user.Role = models.OrgAdminRole
	}

	createdUser, serr := u.userStore.Create(ctx, user)
	if serr != nil {
		return nil, serr
	}

	// Add role grouping
	if createdUser.Role == models.OrgAdminRole {
		if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
			createdUser.ID,
			string(models.OrgAdminRole),
			createdUser.OrganisationID,
		); policyAddErr != nil {
			u.logger.Errorf("failed to add OrgAdmin policy for user: %s", policyAddErr.Error())
			return nil, errors.GeneralError("failed to create user")
		}
	}

	// Add OrgMember grouping
	if policyAddErr := u.resourceAccessPolicyMgr.AddGroupingPolicy(
		createdUser.ID,
		string(models.OrgMemberRole),
		createdUser.OrganisationID,
	); policyAddErr != nil {
		u.logger.Errorf("failed to add OrgMember policy for user: %s", policyAddErr.Error())
	}

	expirationTime := time.Now().UTC().Add(10 * 24 * time.Hour)
	claims := u.jwtClaimsBuilder.BuildClaims(createdUser, expirationTime)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, tokenErr := token.SignedString([]byte(u.jwtSecretKey))
	if tokenErr != nil {
		return nil, errors.GeneralError("failed to generate token: %s", tokenErr.Error())
	}
	res := openapi.UserSignupResponse{
		User:     ptr.To(presenters.PresentUser(createdUser)),
		JwtToken: &tokenString,
	}
	return &res, nil
}
```

Remove the `GetDefaultUser` method. Update `CreateOAuthUser` to not rely on `GetDefaultOrg` (this method will need a different flow -- for now, require an org to be specified or create one).

After editing, run:
```bash
go fmt ./pkg/services/user_service.go
```

---

### Step 2.9: Build verification

- [ ] **Step 2.9: Verify the project compiles after all Chunk 2 changes**

```bash
make binary
```

Fix any remaining compilation errors. Common issues:
- References to `PermissionService.GrantAccess` or `PermissionService.RevokeAllAccess` that still exist (will be removed in Chunk 3, but may cause compile errors now). Temporarily comment them or add stub implementations.
- References to `GetDefaultUser` or `DefaultUser` field
- Missing imports for `auth` or `pgstore` packages in environment files

After fixing, run:
```bash
make binary
```

Expected: clean compilation.

---

### Step 2.10: Commit Chunk 2

- [ ] **Step 2.10: Commit all Chunk 2 changes**

```bash
git add pkg/services/team_service.go \
  pkg/auth/permission_service.go \
  pkg/auth/permission_helpers.go \
  pkg/services/user_service.go \
  pkg/stores/user_store.go \
  pkg/stores/pgstore/users.go \
  cmd/environment/development_environment.go \
  cmd/environment/test_environment.go \
  cmd/environment/environment.go

git commit -m "feat: add TeamService and simplify PermissionService

Create TeamService with CRUD, membership, OrgAdmin management.
Simplify PermissionService to Check-only with OrgAdmin fallback.
Remove GrantAccess/RevokeAccess/RevokeAllAccess from interface.
Update environment init to use default_policies.go.
Remove default platform admin user bootstrap."
```

---

## Chunk 3: Service layer migration

### Step 3.1: Update stack_service.go

- [ ] **Step 3.1: Update `pkg/services/stack_service.go` -- pass teamID to Check, remove GrantAccess/RevokeAllAccess, add ListByTeamID**

Changes:
1. Replace `spec.OrganisationID` with `spec.TeamID` in `CreateStack`'s permission check domain
2. Remove the `GrantAccess` block after stack creation (lines 168-174)
3. Replace `stack.OrganisationID` with `stack.TeamID` in `GetStack`, `UpdateStack`, `DeleteStack` permission checks
4. Remove `RevokeAllAccess` block from `DeleteStack` (lines 430-434)
5. Replace `GetStacksByOrganisationID` with `GetStacksByTeamID`
6. Add `ListByTeamID` to the `StackStore` interface and implementation

**Key method changes in stack_service.go:**

`CreateStack` permission check:
```go
if permErr := auth.CheckServicePermission(s.permissions, ctx, spec.TeamID, auth.ResourceStacks, "", auth.ActionCreate); permErr != nil {
    return nil, permErr
}
```

Remove the `GrantAccess` block entirely (the 7-line block starting with `if s.permissions != nil`).

`GetStack` permission check:
```go
if permErr := auth.CheckServicePermission(s.permissions, ctx, stack.TeamID, auth.ResourceStacks, ID, auth.ActionRead); permErr != nil {
    return nil, permErr
}
```

`UpdateStack` permission check:
```go
if permErr := auth.CheckServicePermission(s.permissions, ctx, existingStack.TeamID, auth.ResourceStacks, ID, auth.ActionWrite); permErr != nil {
    return nil, permErr
}
```

`DeleteStack` permission check:
```go
if permErr := auth.CheckServicePermission(s.permissions, ctx, stack.TeamID, auth.ResourceStacks, ID, auth.ActionDelete); permErr != nil {
    return nil, permErr
}
```

Remove the `RevokeAllAccess` block from `DeleteStack`.

Replace `GetStacksByOrganisationID` with:
```go
func (s *stackService) GetStacksByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, teamID, auth.ResourceStacks, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	stacks, err := s.stackStore.ListByTeamID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	return stacks, nil
}
```

Update the `StackQueryService` interface:
```go
type StackQueryService interface {
	GetStack(ctx context.Context, ID string) (*models.Stack, *errors.ServiceError)
	GetStackByName(ctx context.Context, name string, userID string) (*models.Stack, *errors.ServiceError)
	GetStacksByUserID(ctx context.Context, userID string) ([]*models.Stack, *errors.ServiceError)
	GetStacksByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError)
}
```

After editing, run:
```bash
go fmt ./pkg/services/stack_service.go
```

---

### Step 3.2: Add ListByTeamID to StackStore

- [ ] **Step 3.2: Add `ListByTeamID` to `pkg/stores/stack_store.go` interface and pgstore implementation**

**File:** `pkg/stores/stack_store.go` -- add to interface:
```go
ListByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError)
```

**File:** `pkg/stores/pgstore/stacks.go` (or whatever the stack store implementation file is) -- add:
```go
func (s *dbStackStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.Stack, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var stacks []*models.Stack
	if err := grm.Where("team_id = ? AND deletion_timestamp IS NULL", teamID).
		Preload("StackResources").
		Order("created_at DESC").
		Find(&stacks).Error; err != nil {
		return nil, errors.GeneralError("failed to list stacks by team: %s", err.Error())
	}
	return stacks, nil
}
```

After editing, run:
```bash
go fmt ./pkg/stores/stack_store.go
```

---

### Step 3.3: Update secret_service.go

- [ ] **Step 3.3: Update `pkg/services/secret_service.go` -- pass teamID to Check, remove GrantAccess/RevokeAllAccess, add ListByTeamID**

Changes:
1. `Create` -- permission check uses `secret.TeamID` instead of `secret.OrganisationID`. Remove `GrantAccess` block (lines 107-113).
2. `GetByID` -- permission check uses `secret.TeamID`
3. `Update` -- permission check uses `existingSecret.TeamID`
4. `Delete` -- permission check uses `secret.TeamID`. Remove `RevokeAllAccess` block (lines 247-250).
5. Replace `ListByOrganisation` with `ListByTeamID`:

```go
func (s *secretService) ListByTeamID(ctx context.Context, teamID string) ([]*models.Secret, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, teamID, auth.ResourceSecrets, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.secretStore.ListByTeamID(ctx, teamID)
}
```

Update the `SecretService` interface: replace `ListByOrganisation` with `ListByTeamID`.

Add `ListByTeamID` to `SecretStore` interface and pgstore implementation:

**`pkg/stores/secret.go`:**
```go
ListByTeamID(ctx context.Context, teamID string) ([]*models.Secret, *errors.ServiceError)
```

**pgstore implementation:**
```go
func (s *dbSecretStore) ListByTeamID(ctx context.Context, teamID string) ([]*models.Secret, *errors.ServiceError) {
	grm := s.sessionFactory.New(ctx)
	var secrets []*models.Secret
	if err := grm.Where("team_id = ?", teamID).Order("created_at DESC").Find(&secrets).Error; err != nil {
		return nil, errors.GeneralError("failed to list secrets by team: %s", err.Error())
	}
	return secrets, nil
}
```

After editing, run:
```bash
go fmt ./pkg/services/secret_service.go ./pkg/stores/secret.go
```

---

### Step 3.4: Update volume_service.go

- [ ] **Step 3.4: Update `pkg/services/volume_service.go` -- pass teamID to Check, remove GrantAccess/RevokeAllAccess**

Changes:
1. `Get` -- permission check uses `volume.TeamID`
2. `Create` -- permission check uses `spec.TeamID`. Remove `GrantAccess` block (lines 253-259).
3. `Delete` -- permission check uses `volume.TeamID`. Remove `RevokeAllAccess` block (lines 357-360).
4. `ListByUserID` -- this method can remain, or add `ListByTeamID` if needed. For now, update the permission check domain.

After editing, run:
```bash
go fmt ./pkg/services/volume_service.go
```

---

### Step 3.5: Update workspace_user_service.go

- [ ] **Step 3.5: Update `pkg/services/workspace_user_service.go` -- pass teamID to Check, remove GrantAccess/RevokeAllAccess**

Changes:
1. `GetByID` -- permission check uses `request.TeamID`
2. `Create` -- permission check uses `spec.TeamID`. Remove `GrantAccess` block (lines 137-140).
3. `Update` -- permission check uses `current.TeamID`
4. `Delete` -- permission check uses `workspaceUser.TeamID`. Remove `RevokeAllAccess` block (lines 228-229).

After editing, run:
```bash
go fmt ./pkg/services/workspace_user_service.go
```

---

### Step 3.6: Update postgres_addon_service.go

- [ ] **Step 3.6: Update `pkg/services/postgres_addon_service.go` -- pass teamID to Check, remove GrantAccess/RevokeAllAccess, add ListByTeamID**

Changes:
1. `CreatePostgresAddon` -- permission check uses `postgresAddon.TeamID`. Remove `GrantAccess` block (lines 252-258).
2. `GetPostgresAddon` -- permission check uses `addon.TeamID`
3. `UpdatePostgresAddon` -- permission check uses addon's TeamID
4. `DeletePostgresAddon` -- permission check uses addon's TeamID. Remove `RevokeAllAccess` block (lines 436-440).
5. Replace `ListPostgresAddonsByOrganisation` with `ListPostgresAddonsByTeamID`:

```go
func (s *postgresAddonService) ListPostgresAddonsByTeamID(ctx context.Context, teamID string) ([]*models.PostgresAddon, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, teamID, auth.ResourceAddonsPostgres, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.postgresAddonStore.ListByTeamID(ctx, teamID)
}
```

Update the `PostgresAddonService` interface accordingly.

Add `ListByTeamID` to `PostgresAddonStore` interface and implementation.

After editing, run:
```bash
go fmt ./pkg/services/postgres_addon_service.go
```

---

### Step 3.7: Update object_store_service.go

- [ ] **Step 3.7: Update `pkg/services/object_store_service.go` -- pass teamID to Check, add ListByTeamID**

Changes:
1. `Create` -- permission check uses `objectStore.TeamID`
2. `GetByID` -- permission check uses `objectStore.TeamID`
3. `Update` -- permission check uses existing object store's TeamID
4. `Delete` -- permission check uses object store's TeamID
5. Replace `ListByOrganisation` with `ListByTeamID`:

```go
func (s *objectStoreService) ListByTeamID(ctx context.Context, teamID string) ([]*models.ObjectStore, *errors.ServiceError) {
	if permErr := auth.CheckServicePermission(s.permissions, ctx, teamID, auth.ResourceObjectStores, "", auth.ActionList); permErr != nil {
		return nil, permErr
	}
	return s.objectStoreStore.ListByTeamID(ctx, teamID)
}
```

Update the interface and add `ListByTeamID` to `ObjectStoreStore` and pgstore implementation.

After editing, run:
```bash
go fmt ./pkg/services/object_store_service.go
```

---

### Step 3.8: Update cluster_service.go (org-scoped, no team changes)

- [ ] **Step 3.8: Verify `pkg/services/cluster_service.go` uses orgID for permission checks**

Cluster service should already use `orgID` as the domain. Verify that all `CheckServicePermission` calls pass `orgID` (not teamID). No GrantAccess/RevokeAllAccess to remove since clusters were always org-scoped.

After verifying, run:
```bash
go fmt ./pkg/services/cluster_service.go
```

---

### Step 3.9: Update image_registry_service.go (org-scoped, no team changes)

- [ ] **Step 3.9: Verify `pkg/services/image_registry_service.go` uses orgID for permission checks**

Same as cluster_service -- verify org-scoped permission checks. No changes needed if already correct.

After verifying, run:
```bash
go fmt ./pkg/services/image_registry_service.go
```

---

### Step 3.10: Add all ListByTeamID store methods

- [ ] **Step 3.10: Add `ListByTeamID` to all remaining store interfaces and pgstore implementations**

For each store that needs it, add the interface method and pgstore implementation. The pattern is identical -- query with `WHERE team_id = ?`.

Stores to update:
- `pkg/stores/postgres_addon_store.go` + `pkg/stores/pgstore/postgres_addon.go`
- `pkg/stores/object_store_store.go` + pgstore implementation
- `pkg/stores/volume_store.go` + pgstore implementation (add `ListByTeamID` if needed for future use)

After editing all stores, run:
```bash
go fmt ./pkg/stores/... ./pkg/stores/pgstore/...
```

---

### Step 3.11: Build verification

- [ ] **Step 3.11: Verify the project compiles after all Chunk 3 changes**

```bash
make binary
```

Fix any compilation errors. The most likely issues:
- Handler code still calling old service methods (e.g., `ListByOrganisation`, `GetStacksByOrganisationID`)
- These will be fixed in Chunk 4, but if they block compilation, update the handler to call the new method name with the same parameter for now.

Expected: clean compilation.

---

### Step 3.12: Commit Chunk 3

- [ ] **Step 3.12: Commit all Chunk 3 changes**

```bash
git add pkg/services/stack_service.go \
  pkg/services/secret_service.go \
  pkg/services/volume_service.go \
  pkg/services/workspace_user_service.go \
  pkg/services/postgres_addon_service.go \
  pkg/services/object_store_service.go \
  pkg/services/cluster_service.go \
  pkg/services/image_registry_service.go \
  pkg/stores/stack_store.go \
  pkg/stores/secret.go \
  pkg/stores/postgres_addon_store.go \
  pkg/stores/object_store_store.go \
  pkg/stores/volume_store.go

git commit -m "feat: migrate services to team-scoped permission checks

Pass teamID to permission checks for team-scoped resources.
Remove GrantAccess/RevokeAllAccess from all services.
Replace ListByOrganisation with ListByTeamID for team resources.
Cluster and image registry services remain org-scoped."
```

---

## Chunk 4: API layer

### Step 4.1: Create team handler

- [ ] **Step 4.1: Create `pkg/handlers/team_handler.go` with team CRUD and membership endpoints**

**File:** `pkg/handlers/team_handler.go` (new file)

```go
package handlers

import (
	"net/http"

	"github.com/ashishmax31/stackdome-api-server/pkg/errors"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
	"github.com/ashishmax31/stackdome-api-server/pkg/services"
	"github.com/gorilla/mux"
)

type TeamHandlerSpec struct {
	TeamService services.TeamService
}

type teamHandler struct {
	teamService services.TeamService
}

func NewTeamHandler(spec TeamHandlerSpec) *teamHandler {
	return &teamHandler{
		teamService: spec.TeamService,
	}
}

// Team CRUD

func (h *teamHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			team := &models.Team{Name: req.Name}
			return h.teamService.CreateTeam(r.Context(), orgID, team)
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *teamHandler) List(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			return h.teamService.ListTeams(r.Context(), orgID)
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) GetByName(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			return h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
		},
	}
	handleGet(w, r, cfg)
}

func (h *teamHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.UpdateTeam(r.Context(), team.ID, &models.Team{Name: req.Name})
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *teamHandler) Delete(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return nil, h.teamService.DeleteTeam(r.Context(), team.ID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

// Team Membership

func (h *teamHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.AddMember(r.Context(), team.ID, req.UserID, models.Role(req.Role))
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}

func (h *teamHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			teamName := mux.Vars(r)["team_name"]
			team, serr := h.teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
			if serr != nil {
				return nil, serr
			}
			return h.teamService.ListMembers(r.Context(), team.ID)
		},
	}
	handleList(w, r, cfg)
}

func (h *teamHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role string `json:"role"`
	}
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			return h.teamService.UpdateMemberRole(r.Context(), membershipID, models.Role(req.Role))
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *teamHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			membershipID := mux.Vars(r)["id"]
			return nil, h.teamService.RemoveMember(r.Context(), membershipID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

// OrgAdmin management

func (h *teamHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	cfg := &handlerConfig{
		MarshalInto: &req,
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			return nil, h.teamService.PromoteToOrgAdmin(r.Context(), orgID, req.UserID)
		},
		ErrorHandler: handleError,
	}
	handle(w, r, cfg, http.StatusOK)
}

func (h *teamHandler) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			userID := mux.Vars(r)["user_id"]
			return nil, h.teamService.DemoteOrgAdmin(r.Context(), orgID, userID)
		},
	}
	handleDelete(w, r, cfg, http.StatusNoContent)
}

func (h *teamHandler) ListAdmins(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			orgID := mux.Vars(r)["org_id"]
			return h.teamService.ListOrgAdmins(r.Context(), orgID)
		},
	}
	handleList(w, r, cfg)
}
```

After creating, run:
```bash
go fmt ./pkg/handlers/team_handler.go
```

---

### Step 4.2: Create team resolution helper

- [ ] **Step 4.2: Create a helper function to resolve team_name to team ID in handlers**

Add a helper function that handlers can use to look up the team from the URL path. This avoids duplicating the lookup logic in every handler.

**File:** `pkg/handlers/handler_helper.go` -- add at the end:

```go
// resolveTeamID looks up a team by org_id and team_name from mux vars.
// Returns the team ID or a ServiceError.
func resolveTeamID(r *http.Request, teamService services.TeamService) (string, *errors.ServiceError) {
	orgID := mux.Vars(r)["org_id"]
	teamName := mux.Vars(r)["team_name"]
	if teamName == "" {
		return "", errors.BadRequest("team_name is required")
	}
	team, serr := teamService.GetTeamByOrgAndName(r.Context(), orgID, teamName)
	if serr != nil {
		return "", serr
	}
	return team.ID, nil
}
```

Add the necessary import for `services` package and `mux` if not already present.

After editing, run:
```bash
go fmt ./pkg/handlers/handler_helper.go
```

---

### Step 4.3: Update stack_handler.go for team-scoped routes

- [ ] **Step 4.3: Update `pkg/handlers/stack_handler.go` to extract team_name and resolve team ID**

Add `TeamService` to the `StackHandlerSpec` and `stackHandler` struct. Update handler methods to resolve the team.

Add to spec and struct:
```go
type StackHandlerSpec struct {
	// ... existing fields ...
	TeamService      services.TeamService
}

type stackHandler struct {
	// ... existing fields ...
	teamService      services.TeamService
}
```

Update `NewStackHandler` to wire it.

Update `Create` to resolve teamID and set it on the stack:
```go
func (h *stackHandler) Create(w http.ResponseWriter, r *http.Request) {
	var stack openapi.Stack
	cfg := &handlerConfig{
		&stack,
		nil,
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject := presenters.ConvertStack(&stack)
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = currentUser.ID

			obj, serr := h.stackService.CreateStack(ctx, convertedObject)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStack(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}
```

Update `ListByOrganisationID` to `ListByTeamID`:
```go
func (h *stackHandler) ListByTeamID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			objs, serr := h.stackService.GetStacksByTeamID(r.Context(), teamID)
			if serr != nil {
				return nil, serr
			}
			return presenters.PresentStackList(objs), nil
		},
	}
	handleList(w, r, cfg)
}
```

After editing, run:
```bash
go fmt ./pkg/handlers/stack_handler.go
```

---

### Step 4.4: Update secret_handler.go for team-scoped routes

- [ ] **Step 4.4: Update `pkg/handlers/secret_handler.go` to extract team_name**

Add `TeamService` to spec/struct. Update `Create` to set `TeamID`. Update `ListByOrganisationID` to `ListByTeamID`.

```go
type SecretHandlerSpec struct {
	SecretService services.SecretService
	TeamService   services.TeamService
	Logger        logger.Logger
}

type secretHandler struct {
	secretService services.SecretService
	teamService   services.TeamService
	logger        logger.Logger
}
```

Update `Create`:
```go
func (h *secretHandler) Create(w http.ResponseWriter, r *http.Request) {
	var secret openapi.Secret
	cfg := &handlerConfig{
		&secret,
		nil,
		func() (interface{}, *errors.ServiceError) {
			ctx := r.Context()
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			convertedObject := presenters.ConvertSecret(&secret)
			currentUser, err := auth.GetCurrentUserFromCtx(ctx)
			if err != nil {
				return nil, errors.Unauthorized("failed to fetch user")
			}
			orgID := mux.Vars(r)["org_id"]
			convertedObject.OrganisationID = orgID
			convertedObject.TeamID = teamID
			convertedObject.UserID = currentUser.ID

			obj, serr := h.secretService.Create(ctx, convertedObject)
			if serr != nil {
				h.logger.Errorf("failed to create secret: %v", serr)
				return nil, serr
			}
			return presenters.PresentSecret(obj), nil
		},
		handleError,
	}
	handle(w, r, cfg, http.StatusCreated)
}
```

Rename `ListByOrganisationID` to `ListByTeamID`:
```go
func (h *secretHandler) ListByTeamID(w http.ResponseWriter, r *http.Request) {
	cfg := &handlerConfig{
		Action: func() (interface{}, *errors.ServiceError) {
			teamID, serr := resolveTeamID(r, h.teamService)
			if serr != nil {
				return nil, serr
			}
			objs, serr := h.secretService.ListByTeamID(r.Context(), teamID)
			if serr != nil {
				return nil, serr
			}
			listResp := openapi.SecretList{
				Items: presenters.PresentSecretList(objs),
				Total: ptr.To(int32(len(objs))),
			}
			return listResp, nil
		},
	}
	handleList(w, r, cfg)
}
```

After editing, run:
```bash
go fmt ./pkg/handlers/secret_handler.go
```

---

### Step 4.5: Update remaining handlers for team-scoped routes

- [ ] **Step 4.5: Update volume, postgres addon, object store, and workspace user handlers**

For each handler:
1. Add `TeamService` to spec and struct
2. Update `Create` methods to resolve `teamID` and set it on the model
3. Update `List` methods to use `ListByTeamID` where applicable
4. Wire `TeamService` in the constructor

**Files to update:**
- `pkg/handlers/volume_handler.go`
- `pkg/handlers/postgres_addon_handler.go`
- `pkg/handlers/object_store_handler.go`
- `pkg/handlers/workspace_user_handler.go`

The pattern is the same as Steps 4.3 and 4.4.

After editing all handlers, run:
```bash
go fmt ./pkg/handlers/...
```

---

### Step 4.6: Restructure routes in routes.go

- [ ] **Step 4.6: Rewrite `cmd/server/routes.go` to add team routes and restructure resource routes under `/teams/{team_name}/`**

**File:** `cmd/server/routes.go`

Add the team handler initialization and create team-scoped subrouters. The key changes:

```go
// Create team handler
teamHandler := handlers.NewTeamHandler(handlers.TeamHandlerSpec{
	TeamService: services.TeamService,
})

// Team CRUD routes (under organizations)
teamRouter := organizationsRouter.PathPrefix("/{org_id}/teams").Subrouter()
teamRouter.Use(authenticationMiddleware.AuthenticateUser)
teamRouter.HandleFunc("", teamHandler.Create).Methods(http.MethodPost)
teamRouter.HandleFunc("", teamHandler.List).Methods(http.MethodGet)
teamRouter.HandleFunc("/{team_name}", teamHandler.GetByName).Methods(http.MethodGet)
teamRouter.HandleFunc("/{team_name}", teamHandler.Update).Methods(http.MethodPut)
teamRouter.HandleFunc("/{team_name}", teamHandler.Delete).Methods(http.MethodDelete)

// Team membership routes
teamRouter.HandleFunc("/{team_name}/members", teamHandler.AddMember).Methods(http.MethodPost)
teamRouter.HandleFunc("/{team_name}/members", teamHandler.ListMembers).Methods(http.MethodGet)
teamRouter.HandleFunc("/{team_name}/members/{id}", teamHandler.UpdateMemberRole).Methods(http.MethodPut)
teamRouter.HandleFunc("/{team_name}/members/{id}", teamHandler.RemoveMember).Methods(http.MethodDelete)

// OrgAdmin management routes
organizationsRouter.HandleFunc("/{org_id}/admins", teamHandler.PromoteToAdmin).Methods(http.MethodPost)
organizationsRouter.HandleFunc("/{org_id}/admins", teamHandler.ListAdmins).Methods(http.MethodGet)
organizationsRouter.HandleFunc("/{org_id}/admins/{user_id}", teamHandler.DemoteAdmin).Methods(http.MethodDelete)
```

Move team-scoped resources under team path:

```go
// Team-scoped resource routes (under teams)
teamResourceRouter := teamRouter.PathPrefix("/{team_name}").Subrouter()

// Stacks (team-scoped)
teamResourceRouter.HandleFunc("/stacks", stackHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/stacks", stackHandler.ListByTeamID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.GetByID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.Update).Methods(http.MethodPut)
teamResourceRouter.HandleFunc("/stacks/{id}", stackHandler.Delete).Methods(http.MethodDelete)
teamResourceRouter.HandleFunc("/stacks/{id}/logs", stackHandler.StreamLogs).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/metrics", stackHandler.GetMetrics).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/resources", stackResourceHandler.List).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}", stackResourceHandler.GetByResourceName).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/logs", stackResourceHandler.StreamLogs).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/metrics", stackResourceHandler.GetMetrics).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/resources/{resource_name}/builds", imageBuildHandler.ListByResourceName).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/builds", imageBuildHandler.ListByStackID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/stacks/{id}/builds/{build_id}", imageBuildHandler.GetByID).Methods(http.MethodGet)

// Secrets (team-scoped)
teamResourceRouter.HandleFunc("/secrets", secretHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/secrets", secretHandler.ListByTeamID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.GetByID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.Update).Methods(http.MethodPut)
teamResourceRouter.HandleFunc("/secrets/{id}", secretHandler.Delete).Methods(http.MethodDelete)

// Volumes (team-scoped)
teamResourceRouter.HandleFunc("/volumes", volumeHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/volumes/{id}", volumeHandler.GetByID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/volumes/{id}", volumeHandler.Delete).Methods(http.MethodDelete)

// Postgres addons (team-scoped)
teamResourceRouter.HandleFunc("/addons/postgres", postgresAddonHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/addons/postgres", postgresAddonHandler.List).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.GetByID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.Update).Methods(http.MethodPut)
teamResourceRouter.HandleFunc("/addons/postgres/{id}", postgresAddonHandler.Delete).Methods(http.MethodDelete)
teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/backup", postgresAddonHandler.Backup).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/fence", postgresAddonHandler.Fence).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/addons/postgres/{id}/actions/hibernate", postgresAddonHandler.Hibernate).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/addons/postgres/{id}/backups", postgresAddonHandler.ListBackups).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/addons/postgres/{id}/credentials/{database}", postgresAddonHandler.GetCredentials).Methods(http.MethodGet)

// Object stores (team-scoped)
teamResourceRouter.HandleFunc("/object-stores", objectStoreHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/object-stores", objectStoreHandler.List).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.GetByID).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.Update).Methods(http.MethodPut)
teamResourceRouter.HandleFunc("/object-stores/{id}", objectStoreHandler.Delete).Methods(http.MethodDelete)
```

Remove the old org-scoped routes for these resources (the old `secretRouter`, `stackRouter`, `postgresAddonRouter`, `objectStoreRouter`, `stackStorageRouter` blocks).

Keep org-scoped routes unchanged:
- Cluster routes stay at `/organizations/{org_id}/clusters`
- Image registry routes stay nested under clusters

Also update handler constructor calls to pass `TeamService`:

```go
stackHandler := handlers.NewStackHandler(handlers.StackHandlerSpec{
	StackService:         services.StackService,
	StackResourceService: services.StackResourceService,
	ImageBuildService:    services.ImageBuildService,
	LoggingService:       services.LoggingService,
	MetricsService:       services.MetricsService,
	TeamService:          services.TeamService,
	Logger:               logger,
})

secretHandler := handlers.NewSecretHandler(handlers.SecretHandlerSpec{
	SecretService: services.SecretService,
	TeamService:   services.TeamService,
	Logger:        logger,
})
```

And similarly for volume, postgres addon, object store, workspace user handlers.

Move workspace user routes under teams:
```go
teamResourceRouter.HandleFunc("/workspace-users", workspaceUserHandler.Create).Methods(http.MethodPost)
teamResourceRouter.HandleFunc("/workspace-users/current", workspaceUserHandler.Current).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Get).Methods(http.MethodGet)
teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Update).Methods(http.MethodPut)
teamResourceRouter.HandleFunc("/workspace-users/{id}", workspaceUserHandler.Delete).Methods(http.MethodDelete)
```

Remove old `workspaceUsersRouter` block.

Remove the `organizationsRouter.HandleFunc("/default", ...)` route (no more default org).

After editing, run:
```bash
go fmt ./cmd/server/routes.go
```

---

### Step 4.7: Update user signup to create default team

- [ ] **Step 4.7: Update user_service.go Create method to create default team after org creation**

The `Create` method needs to call the TeamService to create a default team when a new org is created. Add `TeamService` dependency to the user service.

**File:** `pkg/services/user_service.go`

Add `TeamService` to `UserServiceSpec` and `usersService`:

```go
type UserServiceSpec struct {
	// ... existing fields ...
	TeamService             TeamService
}

type usersService struct {
	// ... existing fields ...
	teamService             TeamService
}
```

Update `NewUserService` to wire it. Update `Create` to create default team:

After `user.OrganisationID = createdOrganisation.ID` and `user.Role = models.OrgAdminRole`, add:

```go
// Create default team for the new org
if u.teamService != nil {
	if _, teamErr := u.teamService.CreateDefaultTeam(ctx, createdOrganisation.ID); teamErr != nil {
		u.logger.Errorf("failed to create default team: %s", teamErr.Error())
		return nil, errors.GeneralError("failed to create default team")
	}
}
```

After editing, run:
```bash
go fmt ./pkg/services/user_service.go
```

Wire `TeamService` in the environment files where `userService` is created:

In `cmd/environment/development_environment.go` `loadServices`:
- Create teamService before userService
- Pass it to `UserServiceSpec`

Note: This creates a circular dependency issue if TeamService depends on UserStore and UserService depends on TeamService. Since TeamService uses UserStore directly (not UserService), this is fine.

---

### Step 4.8: Update presenters for new roles

- [ ] **Step 4.8: Update `pkg/presenters/user.go` to present new roles**

If `PresentUser` references old role names, update them. Check if there are any role-specific presentation logic that needs changing.

After editing, run:
```bash
go fmt ./pkg/presenters/...
```

---

### Step 4.9: Remove RemoveFilteredPolicy from grouping (fix RemoveMember)

- [ ] **Step 4.9: Add `RemoveGroupingPolicy` to ResourceAccessPolicyManager interface**

The `RemoveMember` method in TeamService needs to remove grouping policies. `RemoveFilteredPolicy` removes regular policies, not grouping policies. Add `RemoveGroupingPolicy` method.

**File:** `pkg/resourceaccess/resource_access_mgr.go`

Add to the `ResourceAccessPolicyManager` interface:
```go
RemoveGroupingPolicy(subject, role, orgID string) error
```

Add implementation:
```go
func (r *casbinResourceAccessPolicyManager) RemoveGroupingPolicy(subject, role, orgID string) error {
	_, err := r.enforcer.RemoveGroupingPolicy(subject, role, orgID)
	return err
}
```

Then update `TeamService.RemoveMember` and related methods to use `RemoveGroupingPolicy` instead of `RemoveFilteredPolicy`:

```go
// In RemoveMember:
if err := s.policyMgr.RemoveGroupingPolicy(membership.UserID, string(membership.Role), membership.TeamID); err != nil {
    s.logger.Errorf("failed to remove team role grouping: %s", err.Error())
}
```

```go
// In UpdateMemberRole:
_ = s.policyMgr.RemoveGroupingPolicy(existing.UserID, string(existing.Role), existing.TeamID)
```

```go
// In DemoteOrgAdmin:
_ = s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgAdminRole), orgID)
```

```go
// In cleanupOrgMemberGrouping:
_ = s.policyMgr.RemoveGroupingPolicy(userID, string(models.OrgMemberRole), orgID)
```

After editing, run:
```bash
go fmt ./pkg/resourceaccess/resource_access_mgr.go ./pkg/services/team_service.go
```

---

### Step 4.10: Final build verification

- [ ] **Step 4.10: Full build verification and test**

```bash
make binary
```

Fix all remaining compilation errors. At this point, the entire codebase should compile cleanly.

Verify the server starts:
```bash
# Start the server (will need a running database)
./bin/api-server
```

Check that /health returns 200:
```bash
curl http://localhost:8080/health
```

---

### Step 4.11: Write a basic test for default_policies.go

- [ ] **Step 4.11: Add a basic test for `pkg/auth/default_policies.go`**

**File:** `pkg/auth/default_policies_test.go` (new file)

```go
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
```

Run the test:
```bash
go test ./pkg/auth/ -run TestDefault -v
```

Expected: all tests pass.

---

### Step 4.12: Commit Chunk 4

- [ ] **Step 4.12: Commit all Chunk 4 changes**

```bash
git add pkg/handlers/team_handler.go \
  pkg/handlers/handler_helper.go \
  pkg/handlers/stack_handler.go \
  pkg/handlers/secret_handler.go \
  pkg/handlers/volume_handler.go \
  pkg/handlers/postgres_addon_handler.go \
  pkg/handlers/object_store_handler.go \
  pkg/handlers/workspace_user_handler.go \
  cmd/server/routes.go \
  pkg/services/user_service.go \
  pkg/resourceaccess/resource_access_mgr.go \
  pkg/services/team_service.go \
  pkg/auth/default_policies_test.go \
  pkg/presenters/

git commit -m "feat: add team routes and restructure API under /teams/{team_name}/

Create team and membership HTTP handlers.
Restructure all team-scoped resource routes under /teams/{team_name}/.
Update handlers to resolve team_name to team ID.
Add OrgAdmin management endpoints.
User signup now creates default team with org.
Add RemoveGroupingPolicy to ResourceAccessPolicyManager."
```

---

## Post-Implementation Verification

### Step P.1: Full integration test

- [ ] **Step P.1: Run integration tests**

```bash
make test-integration
```

Review output in `test/int/last-run.log`. Fix any failures.

---

### Step P.2: Manual smoke test

- [ ] **Step P.2: Manual smoke test of the complete flow**

1. Sign up a new user -- verify org and default team are created
2. Create a team under the org
3. Add a member to the team
4. Create a stack under the team
5. Verify the stack is accessible by team members
6. Verify the stack is NOT accessible by users in other teams
7. Verify OrgAdmin can access stacks in all teams
8. Create a cluster (org-scoped) and verify all org members can read it
9. Promote a user to OrgAdmin and verify access
10. Demote OrgAdmin (verify last-admin check)
