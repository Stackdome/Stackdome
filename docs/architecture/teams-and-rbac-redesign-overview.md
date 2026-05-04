# Teams & RBAC Redesign — Architecture Overview

*For review and feedback before implementation.*

---

## Why This Change

The current authorization model is flat — all users in an organisation see all resources. This creates problems as customers grow:

- A developer on staging can see production stacks, secrets, and database credentials
- There's no way to scope visibility between environments or teams
- No read-only role for auditors, stakeholders, or onboarding engineers
- The only isolation boundary is the entire organisation

We're introducing **Teams** as a grouping layer between Organisation and Users to solve this.

---

## What's Changing (Summary)

| Area | Current | New |
|------|---------|-----|
| **Isolation boundary** | Organisation | Team (within an Organisation) |
| **Roles** | User, OrganisationAdmin, PlatformAdmin | Viewer, Developer, OrgAdmin |
| **Resource visibility** | All users see all resources in org | Users see only resources in their team(s) |
| **Infrastructure access** | Org-wide (same as resources) | Org-wide read for all members, OrgAdmin manages |
| **PlatformAdmin** | Cross-org super admin | Removed — not needed |
| **Default org** | Shared bucket for unassigned users | Removed — every signup creates a new org |
| **Ownership grants** | Per-resource (creator gets write/delete) | Removed — team role is sufficient |

---

## Organisational Hierarchy

```
Organisation
  ├── OrgAdmin(s)              — org-level role, full access to everything
  ├── Clusters                 — org-scoped infrastructure
  ├── Image Registries         — org-scoped infrastructure
  │
  ├── Team: "default"          — auto-created, cannot be deleted
  │     ├── Members            — each with Viewer or Developer role
  │     ├── Stacks
  │     ├── Secrets
  │     ├── Volumes
  │     ├── Postgres Addons
  │     ├── Object Stores
  │     └── Workspace Users
  │
  ├── Team: "production"
  │     ├── Members
  │     └── ... resources ...
  │
  └── Team: "staging"
        ├── Members
        └── ... resources ...
```

- Each resource belongs to exactly one team
- A user can be in multiple teams with different roles in each
- Infrastructure (clusters, image registries) stays org-level — all members can read, only OrgAdmin can manage

---

## Roles

### OrgAdmin
- **Scope:** Organisation-wide
- **Can do:** Everything — manage teams, manage members, CRUD all resources across all teams, manage clusters and image registries
- **How assigned:** First user (on signup) is automatically OrgAdmin. OrgAdmins can promote others.
- **Constraints:** At least one OrgAdmin must exist per org. Cannot demote the last one.

### Developer
- **Scope:** Per team
- **Can do:** Create, read, update, delete any resource within their team. Read clusters and image registries (org-scoped infrastructure).
- **Cannot do:** Manage teams or members. Manage clusters or image registries. Access resources in teams they're not in.

### Viewer
- **Scope:** Per team
- **Can do:** Read and list resources within their team. Read clusters and image registries.
- **Cannot do:** Create, update, or delete anything.

### OrgMember (implicit, not user-assigned)
- **Scope:** Organisation-wide
- **Purpose:** Grants read-only access to org-scoped infrastructure (clusters, image registries)
- **How assigned:** Automatically added when a user joins any team in the org. Removed when they leave their last team.

---

## Resource Scoping

### Team-Scoped Resources
These belong to a team. Only members of that team can see or manage them.

| Resource | Viewer | Developer | OrgAdmin |
|----------|--------|-----------|----------|
| Stacks | Read, List | Full CRUD | Full CRUD |
| Secrets | Read, List | Full CRUD | Full CRUD |
| Volumes | Read, List | Full CRUD | Full CRUD |
| Postgres Addons | Read, List | Full CRUD | Full CRUD |
| Object Stores | Read, List | Full CRUD | Full CRUD |
| Workspace Users | Read | Full CRUD | Full CRUD |

### Org-Scoped Resources
These belong to the organisation. All org members can read them, only OrgAdmin can manage them.

| Resource | OrgMember (any user) | OrgAdmin |
|----------|---------------------|----------|
| Clusters | Read, List | Full CRUD |
| Image Registries | Read, List | Full CRUD |
| Org Settings | Read | Full CRUD |

---

## Data Model Changes

### New Tables

**teams**
| Column | Type | Notes |
|--------|------|-------|
| id | UUID (PK) | Auto-generated |
| name | VARCHAR(63) | URL-safe: lowercase alphanumeric + hyphens. Unique within org. Used in API URLs. |
| organisation_id | UUID (FK) | References organisations |
| default_team | BOOLEAN | True for the auto-created default team |
| created_at | TIMESTAMP | |
| updated_at | TIMESTAMP | |

**team_memberships**
| Column | Type | Notes |
|--------|------|-------|
| id | UUID (PK) | Auto-generated |
| team_id | UUID (FK) | References teams (cascade delete) |
| user_id | UUID (FK) | References users (cascade delete) |
| role | VARCHAR(50) | 'Viewer' or 'Developer' |
| created_at | TIMESTAMP | |

Constraint: unique (team_id, user_id) — a user can only have one role per team.

### Modified Tables

All team-scoped resources gain a `team_id` column:
- stacks
- secrets
- volumes
- postgres_addons
- object_stores
- workspace_users

The existing `organisation_id` is retained for data integrity and org-level queries.

### Users Table

The `role` field changes:
- Old values: `User`, `OrganisationAdmin`, `PlatformAdmin`
- New: `OrgAdmin` (for org admins), empty string (for regular members whose roles are per-team)

---

## Authorization Architecture

### How Permission Checks Work

The system uses [Casbin](https://casbin.org/) for policy evaluation. Each permission check is a 4-tuple:

```
(subject, domain, resource, action)
```

- **subject:** The user ID
- **domain:** Either a team ID (for team-scoped resources) or org ID (for org-scoped resources)
- **resource:** The resource type and optional ID (e.g., `stacks` for list, `stacks/stack-123` for specific item)
- **action:** `read`, `write`, `delete`, `create`, `list`, `logs`, `exec`

### Permission Resolution Flow

```
1. Is the user authenticated?
   └── No → 401 Unauthenticated

2. Is it an API token request?
   └── Yes → Check token scopes cover the requested resource:action
              If denied → 403

3. Does Casbin allow it?
   └── Casbin checks:
       a) Does the user have a role in this domain that grants this action?
       b) Does a direct user policy exist for this resource?
   └── If allowed → 200 OK

4. OrgAdmin fallback:
   └── Is the user an OrgAdmin for the org that owns this team?
       └── Yes → 200 OK (OrgAdmin has full access to all teams in their org)
       └── No → 403 Access Denied
```

### Why the OrgAdmin Fallback Exists

OrgAdmin's role is registered at the org level (e.g., "user-789 is OrgAdmin of org-abc"). But team-scoped resources are checked against a team domain (e.g., "team-abc"). Casbin can't match "org-abc" to "team-abc" natively, so a single check in application code bridges this: "is this user an OrgAdmin for the org that owns this team?"

This is the only special-case logic. Everything else is handled by Casbin policies.

### Role Policies

These are the predefined policies loaded at application startup. They define what each role can do:

**OrgMember** (auto-assigned to all org members):
- Read and list: clusters, image registries, org settings

**Viewer** (per team):
- Read and list: stacks, secrets, volumes, addons, object stores, workspace users

**Developer** (per team):
- Everything Viewer can do, plus:
- Create, update, delete: stacks, secrets, volumes, addons, object stores, workspace users

**OrgAdmin** (per org):
- Full access to everything (wildcard policy)

### Role Assignment Storage

Casbin stores role assignments as "grouping policies":
```
user-456 is Developer in team-abc
user-456 is Viewer in team-xyz
user-789 is OrgAdmin in org-abc
user-456 is OrgMember in org-abc  (auto-managed)
```

---

## Worked Examples

### Example 1: Developer lists stacks in their team
- User-456 is a Developer in team-abc
- Calls: `GET /api/v1/organizations/{org}/teams/backend/stacks`
- Casbin checks: Does user-456 have Developer role in team-abc? Yes. Does Developer have "list" on "stacks"? Yes.
- **Result: Allowed.** Returns stacks belonging to team-abc only.

### Example 2: Developer tries to see another team's stacks
- User-456 is a Developer in team-abc, but NOT in team-xyz
- Calls: `GET /api/v1/organizations/{org}/teams/production/stacks`
- Casbin checks: Does user-456 have any role in team-xyz? No.
- OrgAdmin fallback: Is user-456 an OrgAdmin? No.
- **Result: Denied (403).**

### Example 3: Viewer tries to delete a stack
- User-123 is a Viewer in team-abc
- Calls: `DELETE /api/v1/organizations/{org}/teams/backend/stacks/{id}`
- Casbin checks: Does Viewer have "delete" on "stacks/{id}"? No (Viewer only has "read").
- **Result: Denied (403).**

### Example 4: OrgAdmin accesses any team's resources
- User-789 is OrgAdmin of org-abc. Team-abc belongs to org-abc.
- Calls: `GET /api/v1/organizations/{org}/teams/backend/stacks/{id}`
- Casbin checks: Does user-789 have a role in team-abc? No (OrgAdmin is org-level).
- OrgAdmin fallback: Is user-789 OrgAdmin for org-abc? Yes. Does team-abc belong to org-abc? Yes.
- **Result: Allowed.**

### Example 5: Developer reads a cluster (org-scoped infra)
- User-456 is a Developer in team-abc within org-abc
- Calls: `GET /api/v1/organizations/{org}/clusters/{id}`
- Casbin checks: Does user-456 have OrgMember role in org-abc? Yes (auto-assigned on team join). Does OrgMember have "read" on "clusters/{id}"? Yes.
- **Result: Allowed.**

### Example 6: Developer tries to delete a cluster
- Same user, calls: `DELETE /api/v1/organizations/{org}/clusters/{id}`
- Casbin checks: Does OrgMember have "delete" on clusters? No (OrgMember only has "read" and "list").
- **Result: Denied (403).** Only OrgAdmin can modify clusters.

---

## API Changes

### New Endpoints

**Team management** (OrgAdmin only):
```
POST   /organizations/{org_id}/teams                            Create team
GET    /organizations/{org_id}/teams                            List teams
GET    /organizations/{org_id}/teams/{team_name}                Get team
PUT    /organizations/{org_id}/teams/{team_name}                Update team
DELETE /organizations/{org_id}/teams/{team_name}                Delete team
```

**Team membership** (OrgAdmin only):
```
POST   /organizations/{org_id}/teams/{team_name}/members        Add member
GET    /organizations/{org_id}/teams/{team_name}/members        List members
PUT    /organizations/{org_id}/teams/{team_name}/members/{id}   Update role
DELETE /organizations/{org_id}/teams/{team_name}/members/{id}   Remove member
```

**OrgAdmin management:**
```
POST   /organizations/{org_id}/admins                           Promote to OrgAdmin
DELETE /organizations/{org_id}/admins/{user_id}                 Demote OrgAdmin
GET    /organizations/{org_id}/admins                           List OrgAdmins
```

### Changed Endpoints

All team-scoped resources move under `/teams/{team_name}/`:

| Before | After |
|--------|-------|
| `/{org_id}/stacks` | `/{org_id}/teams/{team_name}/stacks` |
| `/{org_id}/secrets` | `/{org_id}/teams/{team_name}/secrets` |
| `/{org_id}/volumes` | `/{org_id}/teams/{team_name}/volumes` |
| `/{org_id}/addons/postgres` | `/{org_id}/teams/{team_name}/addons/postgres` |
| `/{org_id}/object-stores` | `/{org_id}/teams/{team_name}/object-stores` |

Team name is used in URLs (not UUID) for readability. Names must be lowercase alphanumeric with hyphens, max 63 characters.

### Unchanged Endpoints

Org-scoped resources stay at their current paths:
- `/{org_id}/clusters`
- `/{org_id}/clusters/{cluster_id}/image_registries`

---

## User Flows

### New User Signup (Normal or OAuth)

```
User signs up
  → New organisation created
  → User becomes OrgAdmin
  → "default" team auto-created in the org
  → User gets OrgMember role (auto-assigned)
```

Same flow for cloud and self-hosted. The first user on a fresh self-hosted installation uses the normal signup — no bootstrap scripts or special setup.

### OrgAdmin Invites a Team Member

```
OrgAdmin invites user (email, team, role)
  → User account created
  → Added to specified team with specified role (Developer or Viewer)
  → OrgMember role auto-assigned for the org
```

### OrgAdmin Promotes Another User

```
OrgAdmin promotes user to OrgAdmin
  → User gains OrgAdmin role at org level
  → Full access to all teams in the org
  → Existing team memberships remain (but are superseded by OrgAdmin access)
```

### User Removed from Last Team

```
OrgAdmin removes user from their last team
  → OrgMember role auto-removed
  → User can still log in but has no access to anything
  → Must be re-added to a team to regain access
```

---

## Default Team

Every organisation has a "default" team that:
- Is auto-created when the org is created
- Cannot be deleted
- Behaves identically to any other team

**Why it exists:** So users can start deploying immediately without setting up team structure. For small teams, the default team is all you need — it feels like the current flat model.

**Growth path:**
1. **Day 1:** OrgAdmin and a few developers, all in default team. No friction.
2. **Small team:** More developers join the default team. Still feels flat.
3. **Growing:** OrgAdmin creates "production" and "staging" teams. Moves resources and people.
4. **Enterprise:** Default team becomes a landing zone. New hires arrive here, OrgAdmin assigns them to proper teams.

---

## Deployment Modes

| | Cloud (multi-tenant) | Self-hosted (single-tenant) |
|---|---|---|
| **Signup** | Creates new org | Same — creates new org |
| **Top-level role** | OrgAdmin (per org) | OrgAdmin (per org) |
| **Cross-org access** | None — complete tenant isolation | None — same model |
| **Platform operations** | Infrastructure tooling (Grafana, K8s, DB) | Same |

No PlatformAdmin in either mode. Platform operations (monitoring, debugging) happen through infrastructure tools, not application-level roles. This eliminates trust concerns in cloud deployments where customers don't want the platform operator browsing their resources.

---

## What's Being Removed

| Concept | Why |
|---------|-----|
| **PlatformAdmin role** | Cross-org super admin creates trust concerns in cloud. Infrastructure tools handle platform ops. |
| **Default organisation** | Shared org for unassigned users is a leaky tenant boundary. Every signup creates its own org. |
| **Per-resource ownership grants** | Teams are the isolation boundary. If you're a Developer in a team, you can manage any resource in that team. No need for per-resource creator grants. |

---

## Future Extensibility

This design explicitly supports:

- **Custom roles:** The policy engine works with any role name. Adding a "QA" role with specific permissions is just new policy entries — no architecture changes.
- **Invites system:** Separate spec, but the team membership model is ready for it.
- **Cross-team sharing:** Casbin's "direct grant" path (currently unused) can grant a specific user access to a specific resource in another team.
- **Multi-org users:** The model supports a user being OrgAdmin in multiple orgs. Not implemented now but the data model allows it.

---

## Questions for Review

1. Does the team-scoped model cover the isolation needs you foresee from enterprise customers?
2. Is OrgAdmin having full access to all teams the right default, or should there be a way to restrict OrgAdmin to specific teams?
3. Should the invite system be designed before this ships, or can it follow as a separate feature?
4. Are there any resources missing from the team-scoped or org-scoped lists?
5. Is the default team concept clear and useful, or is it adding unnecessary complexity?
