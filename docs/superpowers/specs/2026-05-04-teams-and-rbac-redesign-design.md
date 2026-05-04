# Teams & RBAC Redesign

## Problem

The current authorization model has a flat org-level structure where all users in an org can see all resources. This creates several issues:

1. **No visibility scoping** — a developer working on staging can see production stacks, secrets, and addons. Enterprise customers need isolation between environments and teams.
2. **Broken Casbin policies** — the current matcher uses `keyMatch2` which doesn't work with our resource patterns. Base policies using `/*` never actually match resources like `stacks` or `stacks/stack-id`. The system only works because PlatformAdmin bypasses Casbin entirely.
3. **Coarse role model** — only three roles (User, OrganisationAdmin, PlatformAdmin) with no read-only option. No Viewer role for auditors, stakeholders, or new team members who need to observe without modifying.
4. **Per-resource ownership grants are unnecessary** — the GrantAccess/RevokeAllAccess system adds complexity. If teams are the isolation boundary, team roles are sufficient for authorization.
5. **PlatformAdmin is unnecessary** — a cross-org super admin creates trust concerns in cloud deployments and is replaceable by infrastructure tooling.
6. **Default org is unnecessary** — a shared org where unassigned users land creates a leaky multi-tenant boundary.

## Design Decisions

- **Grouping layer:** Teams — a user-facing grouping between Organisation and individual users.
- **Resource visibility:** Team-scoped. A user only sees resources in teams they belong to.
- **Team membership:** A user can belong to multiple teams with a different role in each.
- **Resource ownership:** One team per resource. Resources belong to exactly one team.
- **Ownership grants:** Removed. Team role is sufficient — if you're a Developer in a team, you can modify any resource in that team. Teams ARE the isolation boundary.
- **Infrastructure resources:** Clusters and image registries stay org-scoped. All org members get implicit read access. Only OrgAdmin can create/modify/delete.
- **Object stores:** Team-scoped (tied to postgres addon backups which are team resources).
- **Roles:** Three roles — OrgAdmin, Developer, Viewer. PlatformAdmin removed.
- **Default org:** Removed. Every signup creates a new org.
- **Default team:** Auto-created with every org. Cannot be deleted. New users land here.
- **Policy definition location:** Casbin policies loaded at startup from a dedicated Go file (`pkg/auth/default_policies.go`), not buried in environment bootstrap code.
- **Deployment modes:** Cloud (multi-tenant) and self-hosted (single-tenant) use the same role model. No PlatformAdmin in either mode. Self-hosted first user signs up normally like cloud.

## Architecture

### Role Hierarchy

```
OrgAdmin     — full access within org, manages teams/members/infra
  |-- Developer  — CRUD on team resources, read-only on infra
        |-- Viewer   — read-only on team resources, read-only on infra
```

- **OrgAdmin** is an org-level role, not a team role. OrgAdmins have implicit full access to all teams in their org.
- **Developer** and **Viewer** are team-level roles assigned per team membership.
- Multiple OrgAdmins can exist per org. They are peers with identical access.
- At least one OrgAdmin must exist per org (cannot demote the last one).
- OrgAdmin can promote other org members to OrgAdmin.

### Resource Scoping

| Scope | Resources | Who manages | Who reads |
|-------|-----------|-------------|-----------|
| Team-scoped | stacks, secrets, volumes, postgres addons, object stores, workspace users | Developers in that team | Viewers + Developers in that team |
| Org-scoped | clusters, image registries, org settings | OrgAdmin only | All org members (any team membership) |

### Data Model

**Teams table:**

```sql
CREATE TABLE teams (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR(63) NOT NULL,  -- URL-safe: lowercase alphanumeric and hyphens only
    organisation_id UUID NOT NULL REFERENCES organisations(id),
    default_team    BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMP NOT NULL DEFAULT now(),
    updated_at      TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(name, organisation_id)
);
```

Team names are used in URLs instead of IDs. Names must be lowercase alphanumeric with hyphens (e.g., `backend`, `prod-infra`), validated at the service layer. Max 63 characters (matching K8s naming conventions).

**Team memberships table:**

```sql
CREATE TABLE team_memberships (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id    UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL,  -- 'Viewer' or 'Developer'
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE(team_id, user_id)
);
```

**Team-scoped resources gain `team_id`:**

All team-scoped resources (stacks, secrets, volumes, postgres_addons, object_stores, workspace_users) get a `team_id UUID NOT NULL REFERENCES teams(id)` column. The existing `organisation_id` is kept for data integrity and org-level queries.

**Users table changes:**

- `role` field stores `OrgAdmin` or is empty (regular member). Old role constants (`User`, `OrganisationAdmin`, `PlatformAdmin`) are replaced.
- `organisation_id` remains — a user belongs to one org.

### Casbin Model

The Casbin model stays 4-tuple. Two changes from the current model:

1. `keyMatch2` replaced with `keyMatch` (fixes broken resource pattern matching)
2. Field names clarified: `org` becomes `domain` (since it holds either teamID or orgID)

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

**Matcher explanation:**

The matcher has two paths joined by `||`, plus an action check:

```
PATH A (role-based):
  g(r.subject, p.subject, r.domain)          -- user has the role in this domain?
  && keyMatch(r.resource, p.resource)         -- resource matches the pattern?
  && (p.domain == "*" || r.domain == p.domain) -- domain matches (or policy is wildcard)?

PATH B (direct grant):
  r.subject == p.subject                      -- policy is for this specific user?
  && keyMatch(r.resource, p.resource)          -- resource matches?
  && r.domain == p.domain                      -- exact domain match?

ACTION CHECK (applies to both paths):
  r.action == p.action || p.action == "*"      -- action matches or wildcard?
```

**Path A** resolves role-based access: "does this user have a role in this domain, and does that role grant this action on this resource?"

**Path B** resolves direct user-level policies: a specific user granted access to a specific resource. Currently unused but available for future sharing features.

**`keyMatch` behavior:**

`keyMatch` treats `*` as a wildcard matching any remaining characters (including `/`). The match works by finding the position of `*` in the pattern and checking that everything before it matches exactly:

- `keyMatch("stacks", "stacks")` → exact match → **true**
- `keyMatch("stacks/stack-123", "stacks/*")` → prefix "stacks/" matches → **true**
- `keyMatch("stacks", "stacks/*")` → "stacks" doesn't have "/" after "stacks" → **false**
- `keyMatch("addons/postgres", "addons/*")` → prefix "addons/" matches → **true**
- `keyMatch("addons/postgres/addon-1", "addons/*/*")` → prefix "addons/" matches (first `*` matches everything after) → **true**
- `keyMatch("stacks", "*")` → `*` at position 0, matches everything → **true**
- `keyMatch("addons/postgres/addon-1", "*")` → **true**

This is important for policy design: org-scoped checks pass the bare resource name (e.g., `stacks`), while item-scoped checks pass `resource/resourceID` (e.g., `stacks/stack-123`). The policy patterns are designed so that list/create policies match the bare name and read/write/delete policies match the `resource/*` pattern.

### Casbin Grouping Policies (Role Assignments)

Created when users join teams or become OrgAdmin:

```
# Team-level roles
g, user-456, Developer, team-abc
g, user-456, Viewer, team-xyz

# Org-level roles
g, user-789, OrgAdmin, org-abc

# Auto-created when user joins ANY team in an org
g, user-456, OrgMember, org-abc
```

**OrgMember** is an implicit role — not assigned by admins, but automatically managed:
- Added when a user joins their first team in an org
- Removed when a user is removed from their last team in an org
- Grants read-only access to org-scoped infrastructure

### Base Role Policies

Loaded at startup. Defined in `pkg/auth/default_policies.go` (not buried in environment bootstrap code).

```
# ── OrgMember (auto-assigned, read-only on org infra) ──
p, OrgMember, *, clusters, list
p, OrgMember, *, clusters/*, read
p, OrgMember, *, image-registries, list
p, OrgMember, *, image-registries/*, read
p, OrgMember, *, orgs/*, read

# ── Viewer (read-only on team resources) ──
p, Viewer, *, stacks, list
p, Viewer, *, stacks/*, read
p, Viewer, *, secrets, list
p, Viewer, *, secrets/*, read
p, Viewer, *, volumes, list
p, Viewer, *, volumes/*, read
p, Viewer, *, addons/*, list
p, Viewer, *, addons/*/*, read
p, Viewer, *, object-stores, list
p, Viewer, *, object-stores/*, read
p, Viewer, *, workspace-users/*, read

# ── Developer (CRUD on team resources) ──
p, Developer, *, stacks, list
p, Developer, *, stacks, create
p, Developer, *, stacks/*, *
p, Developer, *, secrets, list
p, Developer, *, secrets, create
p, Developer, *, secrets/*, *
p, Developer, *, volumes, list
p, Developer, *, volumes, create
p, Developer, *, volumes/*, *
p, Developer, *, addons/*, list
p, Developer, *, addons/*, create
p, Developer, *, addons/*/*, *
p, Developer, *, object-stores, list
p, Developer, *, object-stores, create
p, Developer, *, object-stores/*, *
p, Developer, *, workspace-users, create
p, Developer, *, workspace-users/*, *

# ── OrgAdmin (full access to org-scoped resources) ──
p, OrgAdmin, *, *, *
```

### PermissionService.Check Flow

```go
func (p *permissionService) Check(ctx, domain, resource, resourceID, action string) error {
    identity := GetIdentityFromCtx(ctx)
    if identity == nil { return ErrUnauthenticated }

    // API token scope check (if applicable)
    if identity.AuthMethod == AuthMethodAPIToken {
        // ... existing scope validation ...
    }

    // Casbin check
    casbinResource := resource
    if resourceID != "" {
        casbinResource = fmt.Sprintf("%s/%s", resource, resourceID)
    }
    allowed, err := p.policyMgr.CheckPermission(identity.UserID, domain, casbinResource, action)
    if err != nil { return ErrAccessDenied }
    if allowed { return nil }

    // OrgAdmin fallback for team-scoped resources
    // OrgAdmin grouping is on orgID, but team resources use teamID as domain.
    // This single fallback handles the domain mismatch.
    if p.isOrgAdmin(identity) && p.teamBelongsToOrg(domain, identity.OrgID) {
        return nil
    }

    return ErrAccessDenied
}
```

**Why the OrgAdmin fallback is necessary:**

OrgAdmin's Casbin grouping is `g, user, OrgAdmin, org-abc`. When checking access to a team-scoped resource, the domain is `team-abc`. Casbin's `g()` function looks for `g(user, OrgAdmin, team-abc)` which doesn't exist — the grouping is for `org-abc`. Rather than syncing OrgAdmin groupings to every team (maintenance overhead), one Go-level fallback cleanly resolves this.

### Casbin Trace-Through Examples

#### Example 1: Developer lists stacks in their team

```
Request:  r = (user-456, team-abc, stacks, list)
Policies: p = (Developer, *, stacks, list)
Grouping: g = (user-456, Developer, team-abc)

Path A:
  g(user-456, Developer, team-abc)  →  grouping exists? YES
  keyMatch("stacks", "stacks")      →  exact match? YES
  p.domain == "*"                   →  YES (policy domain is wildcard)
  action: "list" == "list"          →  YES
→ ALLOWED
```

#### Example 2: Developer reads a specific stack in their team

```
Request:  r = (user-456, team-abc, stacks/stack-123, read)
Policies: p = (Developer, *, stacks/*, *)
Grouping: g = (user-456, Developer, team-abc)

Path A:
  g(user-456, Developer, team-abc)         →  YES
  keyMatch("stacks/stack-123", "stacks/*") →  prefix "stacks/" matches, YES
  p.domain == "*"                          →  YES
  action: "read" matches p.action "*"      →  YES
→ ALLOWED
```

#### Example 3: Developer tries to read a stack in a team they're NOT in

```
Request:  r = (user-456, team-xyz, stacks/stack-789, read)
Policies: p = (Developer, *, stacks/*, *)
Grouping: g = (user-456, Developer, team-abc)   ← only in team-abc, not team-xyz

Path A:
  g(user-456, Developer, team-xyz)  →  NO grouping exists for team-xyz
  → Path A FAILS

Path B:
  No direct policy for user-456 on team-xyz
  → Path B FAILS

OrgAdmin fallback:
  user-456 is not OrgAdmin
  → DENIED
```

#### Example 4: Viewer tries to delete a stack

```
Request:  r = (user-123, team-abc, stacks/stack-123, delete)
Policies: p = (Viewer, *, stacks/*, read)    ← Viewer only has read on items
Grouping: g = (user-123, Viewer, team-abc)

Path A:
  g(user-123, Viewer, team-abc)              →  YES
  keyMatch("stacks/stack-123", "stacks/*")   →  YES
  p.domain == "*"                            →  YES
  action: "delete" == "read"                 →  NO
  → No matching policy with action "delete" or "*" for Viewer
→ DENIED
```

#### Example 5: OrgAdmin reads a stack in any team (Go fallback)

```
Request:  r = (user-789, team-abc, stacks/stack-123, read)
Policies: p = (OrgAdmin, *, *, *)
Grouping: g = (user-789, OrgAdmin, org-abc)

Path A:
  g(user-789, OrgAdmin, team-abc)  →  NO (grouping is for org-abc, not team-abc)
  → Path A FAILS

Path B:
  No direct policy for user-789 on team-abc
  → Path B FAILS

OrgAdmin fallback (Go code):
  isOrgAdmin(user-789)? YES
  teamBelongsToOrg(team-abc, org-abc)? YES
→ ALLOWED
```

#### Example 6: OrgAdmin manages a cluster (org-scoped resource)

```
Request:  r = (user-789, org-abc, clusters/cluster-1, delete)
Policies: p = (OrgAdmin, *, *, *)
Grouping: g = (user-789, OrgAdmin, org-abc)

Path A:
  g(user-789, OrgAdmin, org-abc)          →  YES
  keyMatch("clusters/cluster-1", "*")     →  YES (wildcard matches everything)
  p.domain == "*"                         →  YES
  action: "delete" matches "*"            →  YES
→ ALLOWED
```

#### Example 7: Regular user reads a cluster (OrgMember infra access)

```
Request:  r = (user-456, org-abc, clusters/cluster-1, read)
Policies: p = (OrgMember, *, clusters/*, read)
Grouping: g = (user-456, OrgMember, org-abc)    ← auto-created on team join

Path A:
  g(user-456, OrgMember, org-abc)              →  YES
  keyMatch("clusters/cluster-1", "clusters/*") →  YES
  p.domain == "*"                              →  YES
  action: "read" == "read"                     →  YES
→ ALLOWED
```

#### Example 8: Regular user tries to delete a cluster

```
Request:  r = (user-456, org-abc, clusters/cluster-1, delete)
Policies: p = (OrgMember, *, clusters/*, read)   ← OrgMember only has read
Grouping: g = (user-456, OrgMember, org-abc)

Path A:
  g(user-456, OrgMember, org-abc)              →  YES
  keyMatch("clusters/cluster-1", "clusters/*") →  YES
  p.domain == "*"                              →  YES
  action: "delete" == "read"                   →  NO
  → No OrgMember policy for delete
→ DENIED (only OrgAdmin can modify clusters)
```

#### Example 9: Developer reads a postgres addon (hierarchical resource)

```
Request:  r = (user-456, team-abc, addons/postgres/addon-1, read)
Policies: p = (Developer, *, addons/*/*, *)
Grouping: g = (user-456, Developer, team-abc)

Path A:
  g(user-456, Developer, team-abc)                      →  YES
  keyMatch("addons/postgres/addon-1", "addons/*/*")      →  prefix "addons/" matches, YES
  p.domain == "*"                                        →  YES
  action: "read" matches "*"                             →  YES
→ ALLOWED
```

## Signup & User Flows

### User Signup (Normal or OAuth)

Same flow for both cloud and self-hosted deployments. The first user on a fresh installation uses the same signup flow:

```
User signs up
  → New Organisation created
  → User becomes OrgAdmin of that org
  → Casbin grouping: g, user, OrgAdmin, org-new
  → Casbin grouping: g, user, OrgMember, org-new
  → Default team auto-created in the org
```

### Invite to Existing Org

```
OrgAdmin invites user to org
  → User account created (or existing user linked)
  → Added to specified team with specified role (Developer/Viewer)
  → Casbin groupings:
      g, user, Developer, team-specified
      g, user, OrgMember, org-abc   (if first team in this org)
```

### Promote to OrgAdmin

```
OrgAdmin promotes user
  → Casbin grouping added: g, user, OrgAdmin, org-abc
  → User now has implicit full access to all teams in the org
  → User's existing team memberships remain (but are effectively superseded)
```

### Demote OrgAdmin

```
OrgAdmin demotes another OrgAdmin
  → Check: is this the last OrgAdmin? If yes, deny.
  → Casbin grouping removed: g, user, OrgAdmin, org-abc
  → User retains their team memberships and OrgMember grouping
```

### Remove from Team

```
OrgAdmin removes user from team
  → Team membership deleted
  → Casbin grouping removed: g, user, role, team-id
  → If last team in org:
      → OrgMember grouping removed: g, user, OrgMember, org-abc
      → User can still log in but has no access to anything
```

## Default Team

- Auto-created when an org is created, with `default_team = true`
- Cannot be deleted (enforced in service layer)
- New users invited without a specific team go to the Default team
- The Default team behaves identically to any other team — same roles, same permissions

**Growth path:**
1. **Day 1:** Just the OrgAdmin. Creates resources in Default team. Feels like a flat org.
2. **Small team:** OrgAdmin invites developers. Everyone's in Default team. Still feels flat.
3. **Growing:** OrgAdmin creates "Production", "Staging" teams. Moves resources and people. Default team becomes shared/sandbox.
4. **Enterprise:** Default team is a landing zone. New hires arrive here, OrgAdmin assigns them to proper teams.

## API Endpoints

### Team CRUD (OrgAdmin only)

```
POST   /api/v1/organizations/{org_id}/teams                     -- create team
GET    /api/v1/organizations/{org_id}/teams                     -- list teams
GET    /api/v1/organizations/{org_id}/teams/{team_name}         -- get team
PUT    /api/v1/organizations/{org_id}/teams/{team_name}         -- update team
DELETE /api/v1/organizations/{org_id}/teams/{team_name}         -- delete team (not default)
```

### Team Membership (OrgAdmin only)

```
POST   /api/v1/organizations/{org_id}/teams/{team_name}/members       -- add member
GET    /api/v1/organizations/{org_id}/teams/{team_name}/members       -- list members
PUT    /api/v1/organizations/{org_id}/teams/{team_name}/members/{id}  -- update role
DELETE /api/v1/organizations/{org_id}/teams/{team_name}/members/{id}  -- remove member
```

### OrgAdmin Management

```
POST   /api/v1/organizations/{org_id}/admins                -- promote to OrgAdmin
DELETE /api/v1/organizations/{org_id}/admins/{user_id}      -- demote OrgAdmin
GET    /api/v1/organizations/{org_id}/admins                -- list OrgAdmins
```

### Team-Scoped Resource Endpoints

Resources move under the team path, using team name (not UUID) in URLs:

```
# Before (org-scoped):
POST /api/v1/organizations/{org_id}/stacks
GET  /api/v1/organizations/{org_id}/stacks

# After (team-scoped):
POST   /api/v1/organizations/{org_id}/teams/{team_name}/stacks
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks
GET    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}
PUT    /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}
DELETE /api/v1/organizations/{org_id}/teams/{team_name}/stacks/{id}
```

Same pattern for: secrets, volumes, postgres addons, object stores, workspace users.

Handlers resolve `{team_name}` to the internal team UUID via `(org_id, team_name)` lookup. All internal references and foreign keys use the UUID.

### Org-Scoped Resource Endpoints (unchanged)

```
# Clusters (OrgAdmin manages, all members read)
GET    /api/v1/organizations/{org_id}/clusters
POST   /api/v1/organizations/{org_id}/clusters
GET    /api/v1/organizations/{org_id}/clusters/{id}
DELETE /api/v1/organizations/{org_id}/clusters/{id}

# Image registries (unchanged)
GET    /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries
POST   /api/v1/organizations/{org_id}/clusters/{cluster_id}/image_registries
...
```

## Service Layer Changes

### Remove GrantAccess / RevokeAllAccess

Since team role is sufficient, per-resource ownership grants are removed from all services. The `GrantAccess` and `RevokeAllAccess` calls in Create/Delete methods of stack, secret, volume, postgres addon, and workspace user services are deleted.

The `PermissionService` interface simplifies to:

```go
type PermissionService interface {
    Check(ctx context.Context, domain, resource, resourceID, action string) error
}
```

`GrantAccess`, `RevokeAccess`, and `RevokeAllAccess` methods are removed.

### Service Check Calls

Services pass `teamID` for team-scoped resources and `orgID` for org-scoped resources:

```go
// Team-scoped (stack service)
func (s *stackService) GetStack(ctx context.Context, id string) (*models.Stack, error) {
    stack, err := s.store.GetByID(ctx, id)
    if err != nil { return nil, err }
    if err := s.permissions.Check(ctx, stack.TeamID, auth.ResourceStacks, id, auth.ActionRead); err != nil {
        return nil, err
    }
    return stack, nil
}

// Org-scoped (cluster service)
func (s *clusterService) GetCluster(ctx context.Context, id string) (*models.Cluster, error) {
    cluster, err := s.store.GetByID(ctx, id)
    if err != nil { return nil, err }
    if err := s.permissions.Check(ctx, cluster.OrganisationID, auth.ResourceClusters, id, auth.ActionRead); err != nil {
        return nil, err
    }
    return cluster, nil
}
```

### List Operations

Team-scoped list operations query by team:

```go
func (s *stackService) ListByTeamID(ctx context.Context, teamID string) ([]*models.Stack, error) {
    if err := s.permissions.Check(ctx, teamID, auth.ResourceStacks, "", auth.ActionList); err != nil {
        return nil, err
    }
    return s.store.ListByTeamID(ctx, teamID)
}
```

For "list all my resources across teams," the handler fetches the user's team memberships and calls `ListByTeamID` for each.

## Database Changes

### New Tables

- `teams` (schema above)
- `team_memberships` (schema above)

### Modified Tables

- `stacks` — add `team_id UUID NOT NULL REFERENCES teams(id)`
- `secrets` — add `team_id`
- `volumes` — add `team_id`
- `postgres_addons` — add `team_id`
- `object_stores` — add `team_id`
- `workspace_users` — add `team_id`
- `users` — update role values (remove PlatformAdmin, User, OrganisationAdmin; use OrgAdmin for admins, empty for regular members)

### Removed

- Default org concept (no `default` column needed on organisations)

## Files Changed

### New Files

- `pkg/models/team.go` — Team and TeamMembership models
- `pkg/stores/team_store.go` — Team store interface
- `pkg/stores/pgstore/team_store.go` — Team store implementation
- `pkg/stores/team_membership_store.go` — TeamMembership store interface
- `pkg/stores/pgstore/team_membership_store.go` — TeamMembership store implementation
- `pkg/services/team_service.go` — Team CRUD + membership management + OrgMember auto-grouping
- `pkg/handlers/team_handler.go` — Team and membership HTTP handlers
- `pkg/auth/default_policies.go` — Base role policies (moved from environment bootstrap)
- `pkg/db/migrations/YYYYMMDDHHMM_create_teams.go` — Teams and memberships tables
- `pkg/db/migrations/YYYYMMDDHHMM_add_team_id_to_resources.go` — Add team_id to all team-scoped resources

### Modified Files

- `pkg/auth/permission_service.go` — Simplified Check (remove GrantAccess/RevokeAllAccess, add OrgAdmin fallback)
- `pkg/auth/identity.go` — Remove PlatformAdmin, add OrgAdmin/OrgMember checks
- `pkg/resourceaccess/casbin_model.conf` — keyMatch2 → keyMatch, field naming
- `pkg/models/user.go` — Update role constants (OrgAdmin only, remove PlatformAdmin/User/OrganisationAdmin)
- `cmd/server/routes.go` — Add team/membership routes, restructure resource routes under teams
- `cmd/environment/development_environment.go` — Use default_policies.go, remove PlatformAdmin setup
- `cmd/environment/test_environment.go` — Same
- `pkg/services/*.go` — All services: remove GrantAccess/RevokeAllAccess, pass teamID to Check, add TeamID to create inputs
- `pkg/handlers/*.go` — Extract teamID from URL path
- `pkg/services/user_service.go` — Signup creates org + default team, remove PlatformAdmin logic
- `pkg/presenters/user.go` — Update role presentation

### Deleted Code

- Per-resource GrantAccess/RevokeAllAccess calls from all services
- PlatformAdmin role constant and bypass logic
- Default org creation logic
- `initializeBaseResourceAccessPolicies` from environment files (moved to default_policies.go)

## Implementation Chunks

The work is split into logical chunks for context management, not for incremental shipping. There are no backward compatibility concerns — the product is not released.

### Chunk 1: Foundation (models, stores, Casbin)
- Create Team and TeamMembership models
- Create team and team_membership stores (interface + pgstore)
- Fix Casbin matcher (keyMatch2 → keyMatch)
- Create `pkg/auth/default_policies.go` with all base policies
- Update role constants in user model
- Update identity.go (remove PlatformAdmin, add OrgAdmin/OrgMember)
- Create database migrations (teams table, memberships table, add team_id to resources)

### Chunk 2: Team service and permission changes
- Create TeamService (CRUD, membership, OrgMember auto-grouping, OrgAdmin promotion)
- Simplify PermissionService (remove Grant/Revoke, add OrgAdmin fallback)
- Update environment files to use default_policies.go

### Chunk 3: Service layer migration
- Update all services to pass teamID/orgID to Check
- Remove GrantAccess/RevokeAllAccess from all services
- Add TeamID to resource creation inputs
- Update list operations to ListByTeamID

### Chunk 4: API layer
- Create team and membership handlers
- Restructure resource routes under teams
- Update existing handlers to extract teamID from URL
- Update signup flow (create org + default team)
