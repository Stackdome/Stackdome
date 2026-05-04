# Chunk 2 Completed: Team service and permission changes

**Branch:** `feat/teams-rbac-redesign`
**Spec:** `docs/superpowers/specs/2026-05-04-teams-and-rbac-redesign-design.md`
**Full plan:** `docs/superpowers/plans/2026-05-04-teams-and-rbac-redesign.md`

## What Changed

### New Files
- `pkg/services/team_service.go` — TeamService interface and implementation with:
  - Team CRUD (`CreateTeam`, `GetTeam`, `GetTeamByOrgAndName`, `ListTeams`, `UpdateTeam`, `DeleteTeam`)
  - `CreateDefaultTeam` for org bootstrap (default_team=true, name="default", cannot be deleted)
  - Membership management (`AddMember`, `RemoveMember`, `UpdateMemberRole`, `ListMembers`, `ListUserTeams`)
  - OrgAdmin management (`PromoteToOrgAdmin`, `DemoteOrgAdmin`, `ListOrgAdmins`)
  - OrgMember auto-grouping: Casbin `OrgMember` grouping added on first team join, removed when last team membership in org is removed
  - Team name validation: lowercase alphanumeric with hyphens, max 63 chars

### Modified Files

#### Permission System
- `pkg/auth/permission_service.go` — **Rewritten.** Interface simplified to `Check` only. `GrantAccess`, `RevokeAccess`, `RevokeAllAccess` removed. Added `teamBelongsToOrg` OrgAdmin fallback: when Casbin denies access (because OrgAdmin grouping is on orgID but team resources use teamID as domain), a Go-level check verifies the team belongs to the user's org.
- `pkg/auth/permission_helpers.go` — `CheckServicePermission` parameter renamed from `orgID` to `domain` (cosmetic, matches new terminology).
- `pkg/resourceaccess/resource_access_mgr.go` — Added `RemoveGroupingPolicy(subject, role, orgID)` to `ResourceAccessPolicyManager` interface and Casbin implementation. Used by TeamService for membership removal and OrgAdmin demotion.

#### User Store
- `pkg/stores/user_store.go` — Added `Update` and `ListByOrgAndRole` to `UserStore` interface.
- `pkg/stores/pgstore/users.go` — Added `Update` and `ListByOrgAndRole` implementations.

#### Environment / Wiring
- `cmd/environment/environment.go` — Added `TeamService` to `Services` struct.
- `cmd/environment/development_environment.go` — `initializePermissionService` now passes `TeamStore` to `PermissionServiceConfig`. TeamService created in `loadServices` and wired into `Services`.
- `cmd/environment/test_environment.go` — Same changes as development environment.

#### User Service (Signup)
- `pkg/services/user_service.go` — Three changes:
  1. **Signup (`Create`)**: Removed default `DeveloperRole` assignment for empty role. OrgAdmin grouping is now conditional (only when role is OrgAdmin). Added `OrgMember` grouping for all signups.
  2. **OAuth signup (`CreateOAuthUser`)**: No longer calls `GetDefaultOrg`. Instead creates a new org for the OAuth user (matching spec: every signup creates its own org). User gets `OrgAdminRole` + both `OrgAdmin` and `OrgMember` Casbin groupings.
  3. Removed unused `GetDefaultUser` reference (already cleaned in Chunk 1).

#### GrantAccess/RevokeAllAccess Removal (10 call sites)
- `pkg/services/secret_service.go` — Removed `GrantAccess` in `Create` and `RevokeAllAccess` in `Delete`.
- `pkg/services/stack_service.go` — Removed `GrantAccess` in `CreateStack` and `RevokeAllAccess` in `DeleteStack`.
- `pkg/services/volume_service.go` — Removed `GrantAccess` in `Create` and `RevokeAllAccess` in `Delete`.
- `pkg/services/postgres_addon_service.go` — Removed `GrantAccess` in `CreatePostgresAddon` and `RevokeAllAccess` in `DeletePostgresAddon`.
- `pkg/services/workspace_user_service.go` — Removed `GrantAccess` in `Create` and `RevokeAllAccess` in `Delete`.

## Design Decisions / Deviations

### GrantAccess/RevokeAllAccess removed early
The plan placed this removal in Chunk 3 (service layer migration). We removed them in Chunk 2 because the PermissionService interface change (removing these methods) caused compile errors across all 5 services. Removing the dead calls now was cleaner than adding stub implementations.

### OAuth user creates own org
`CreateOAuthUser` previously assigned users to the default org. Updated to create a new org per the spec ("every signup creates its own org"). The full OAuth/invite flow rewrite is deferred to Chunk 4, but this minimal change keeps the code compiling and spec-aligned.

### RemoveGroupingPolicy added early
The plan placed `RemoveGroupingPolicy` in Step 4.9 (Chunk 4). We added it in Chunk 2 because TeamService's `RemoveMember`, `UpdateMemberRole`, `DemoteOrgAdmin`, and `cleanupOrgMemberGrouping` all require it.

## What's NOT Done Yet (Chunks 3-4)
- All services migrated to pass teamID (instead of orgID) to `Check` for team-scoped resources — Chunk 3
- `ListByTeamID` store methods and service methods replacing `ListByOrganisation` — Chunk 3
- Team handler, routes restructured under `/teams/{team_name}/`, signup creates default team — Chunk 4

## Build Status
`make binary` compiles cleanly.
