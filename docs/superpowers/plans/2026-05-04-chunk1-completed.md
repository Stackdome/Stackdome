# Chunk 1 Completed: Foundation (models, stores, Casbin, migrations)

**Branch:** `feat/teams-rbac-redesign`
**Spec:** `docs/superpowers/specs/2026-05-04-teams-and-rbac-redesign-design.md`
**Full plan:** `docs/superpowers/plans/2026-05-04-teams-and-rbac-redesign.md`

## What Changed

### New Files
- `pkg/models/team.go` — Team and TeamMembership models
- `pkg/stores/team_store.go` — TeamStore interface (CRUD + GetByOrgAndName + GetDefaultTeamForOrg)
- `pkg/stores/team_membership_store.go` — TeamMembershipStore interface (CRUD + ListByTeamID/UserID/UserIDAndOrgID)
- `pkg/stores/pgstore/team_store.go` — TeamStore PostgreSQL implementation
- `pkg/stores/pgstore/team_membership_store.go` — TeamMembershipStore PostgreSQL implementation
- `pkg/auth/default_policies.go` — Base Casbin policies for OrgMember, Viewer, Developer, OrgAdmin roles. `LoadDefaultPolicies()` function used by environment init.
- `pkg/db/migrations/202605040001_create_teams_table.go` — Teams table with unique(name, org_id)
- `pkg/db/migrations/202605040002_create_team_memberships_table.go` — Team memberships with unique(team_id, user_id)
- `pkg/db/migrations/202605040003_add_team_id_to_resources.go` — Adds nullable team_id to: stacks, secrets, volumes, postgres_addons, object_stores, workspace_users
- `pkg/db/migrations/202605040004_update_user_roles.go` — Migrates OrganisationAdmin/PlatformAdmin → OrgAdmin, User → empty, drops default_user column

### Modified Files
- `pkg/models/user.go` — Replaced role constants: `OrgAdminRole`, `DeveloperRole`, `ViewerRole`, `OrgMemberRole`. Removed `DefaultUser` field. Added `IsOrgAdmin()` method. Simplified `ClusterAccessRules()`.
- `pkg/models/organisation.go` — Removed `Default` field and `DefaultOrgName` constant.
- `pkg/models/stack.go` — Added `TeamID` field
- `pkg/models/secret.go` — Added `TeamID` field
- `pkg/models/volume.go` — Added `TeamID` field
- `pkg/models/postgres_addon.go` — Added `TeamID` field
- `pkg/models/object_store.go` — Added `TeamID` field
- `pkg/models/workspace_user.go` — Added `TeamID` field
- `pkg/auth/identity.go` — Replaced `IsPlatformAdmin()` with `IsOrgAdmin()`. Removed `PlatformAdminRole` constant.
- `pkg/auth/permission_service.go` — Changed `IsPlatformAdmin()` bypass to `IsOrgAdmin()` (minimal fix — full redesign in Chunk 2)
- `pkg/resourceaccess/casbin_model.conf` — Changed `keyMatch2` → `keyMatch`, renamed `org` → `domain`
- `pkg/db/migrations/migrations.go` — Registered 4 new migrations
- `pkg/db/migrations/202501260933_create_default_organisation.go` — Inlined Organisation struct to decouple from model changes
- `pkg/stores/user_store.go` — Removed `GetDefaultUser` from interface
- `pkg/stores/pgstore/users.go` — Removed `GetDefaultUser` implementation and `DefaultUser` validation
- `pkg/services/user_service.go` — Removed `GetDefaultUser` from interface and impl. Changed `UserRole` → `DeveloperRole`, `OrganisationAdminRole` → `OrgAdminRole`.
- `pkg/services/organisation_service.go` — Changed role checks from PlatformAdmin/OrganisationAdmin to OrgAdmin
- `pkg/services/cluster_service.go` — Changed `IsPlatformAdmin()` → `IsOrgAdmin()`
- `pkg/services/image_registry_service.go` — Changed `IsPlatformAdmin()` → `IsOrgAdmin()`
- `pkg/presenters/user.go` — Updated role mapping to use `OrgAdminRole`
- `pkg/presenters/organisation.go` — Removed `Default` field reference
- `cmd/environment/development_environment.go` — Replaced inline policies with `auth.LoadDefaultPolicies()`. Removed `ensureDefaultPlatformAdminUser` entirely (no bootstrap users — first user signs up normally).
- `cmd/environment/test_environment.go` — Same changes as development environment.

## What's NOT Done Yet (Chunks 2-4)
- TeamService (CRUD, membership management, OrgMember auto-grouping) — Chunk 2
- PermissionService redesign (remove Grant/Revoke, add OrgAdmin fallback with teamBelongsToOrg) — Chunk 2
- All services migrated to pass teamID to Check, remove GrantAccess/RevokeAllAccess — Chunk 3
- Team handler, routes restructured under /teams/{team_name}/, signup flow rewrite — Chunk 4

### Default Org — still exists, removal deferred to Chunk 4
The old migration `202501260933_create_default_organisation.go` still creates a "DefaultOrganisation" on fresh DBs. `GetDefaultOrg` is still referenced in:
- `pkg/stores/organisation_store.go` (interface)
- `pkg/stores/pgstore/organisations.go` (implementation)
- `pkg/services/organisation_service.go` (interface + implementation)
- `pkg/services/user_service.go` (`CreateOAuthUser` assigns OAuth users to default org)
- `pkg/handlers/organisation_handler.go` (`GetDefault` endpoint)

Per the spec, every signup creates its own org — no shared default org. These references will be removed in Chunk 4 when the signup flow is rewritten. The migration can't be deleted (already committed) but is effectively dead code — nothing in the startup depends on it anymore.

## Build Status
`make binary` compiles cleanly.
