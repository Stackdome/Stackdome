# Design: Restore env source-bindings UI on the connections API

**Date:** 2026-06-04
**Status:** Approved (brainstorming)
**Area:** `frontend/src/pages/stacks` (stack resource Environment tab)

## Context

A stack resource's environment variables can be sourced from more than a literal
value: from a **secret** key, a **postgres addon** credential, **another
resource's** output, or the **resource's own** output. The UI used to support
this with a per-row "From" selector plus an addon-binding group.

Commit `8891fee` ("drop env-from-secret/addon from stacks form") removed that UI
because **main's connections redesign** deleted the inline
`environment_variables_from_secret` / `env_from_addons` fields from a resource's
`execution_config` — addon/secret env injection moved into a separate
**stack-connections** model. The frontend types were regenerated and the now-invalid
UI stripped to literal `KEY/VALUE` only. The commit explicitly deferred the
re-implementation: *"The secret/addon env feature should be re-introduced via the
connections model in a dedicated change."*

This is that change: bring the binding UX back, targeting the connections API.
Today those bindings are only creatable via the raw API (as the ToolJet demo stack
was built).

## Goal

Restore the env-var-centric binding UX (a "From" selector on each env row) so a
user can bind, from the resource's Environment tab:
- **Literal** value
- **Secret** key
- **Addon** credential (postgres)
- **Resource** output (cross-resource)
- **Self** output (the resource's own output)

…where Secret/Addon/Resource rows persist as **stack connections** and
Literal/Self rows persist as the resource's env vars.

## Approach

**Recover + adapt the reverted code.** Restore the old `EnvRow` source selector,
`EnvAddonGroup`, the `FormEnvVar` discriminated union, and the group/expand logic
from `8891fee~1`; then **swap the serialization target** from the deleted inline
fields to stack connections, and add the Resource + Self source arms. The UX layer
is what we want and was proven; the serialize/persist layer is the new work.
(Rejected: rebuild fresh — re-derives an existing UX for more effort.)

## Design

### 1. Form data model — `schemas/form-schema.ts`
Restore the discriminated union, with two new arms:

| `from` | fields |
|--------|--------|
| `stack` (literal) | `name, value` |
| `secret` | `name, secretId, secretKey` |
| `addon` | `name, addonId, database?, superuser, credField` |
| `resource` *(new)* | `name, resourceName, output` |
| `self` *(new)* | `name, selfOutput` |

Validation (mirrors the backend validator):
- addon: requires `database` + `credField` unless `superuser`
- secret: requires `secretKey`
- resource/self: requires `output`
- no duplicate (source → this-resource) group (addon: per `(addonId, database)`)

### 2. UI — `components/shared/env-row.tsx` + `env-addon-group.tsx`
Per-row `From` selector with sub-pickers:
- **Secret**: secret ▸ key (keys from the selected secret's `keys`).
- **Addon**: addon ▸ database ▸ field. Field options = backend addon outputs:
  `host, port, database, username, password, sslmode, ca_certificate, url`.
  Old labels `connectionString` → `url`, `caCertificate` → `ca_certificate`.
  Rows for the same `(addon, database)` are grouped under `EnvAddonGroup`
  (restores the edit/detach state machine).
- **Resource / Self**: resource ▸ output. Output options = the resource's declared
  outputs: `host`, `url.<port>`, `public.<port>.url`, `public.<port>.host`
  (e.g. mailhog `host`, self `public.http.url`).

### 3. Serialization — `schemas/form-schema.ts`
On save, split a resource's env rows:
- **Literal** + **Self** rows → `execution_config.environment_variables`
  (`{name, value}` / `{name, self_output}`) — part of the stack payload.
- **Secret / Addon / Resource** rows → **grouped into stack connections**
  (`kind=env`, `from` = the source, `to` = this resource, `mappings[]` one per row).
  One connection per `(source, database?, resource)`; output accessors:
  addon → field name; secret → `secretOutputAccessor(key)` (`key.<K>` simple,
  `key['<K>']` escaped); resource → the chosen output.

On load, the reverse: expand `stack.spec.connections` back into per-resource env
rows (reuse the old expand logic), and read `self_output` env vars into Self rows.

### 4. Persistence — new `api/connections.ts` + save flow
`StackUpdateRequest` carries **no** connections, so:
- **Create** (new stack): connections ride in `spec.connections` (unchanged).
- **Update** (existing stack): on "Save Changes", diff the desired connections vs
  the loaded set and fire granular CRUD, alongside the stack PUT:
  - a new source group (no matching loaded connection) → `POST /stacks/{id}/connections`
  - an existing group whose mappings/config changed (a row added, removed, or
    re-pointed within the same `(source → resource)`) → `PUT …/connections/{cid}`
    with the full new mapping set
  - a group with no rows left → `DELETE …/connections/{cid}`

  Diff is at **connection (source-group) granularity**, not per-mapping: a
  connection bundles all mappings for one `(from, to, config)`, so a single
  changed field is a `PUT` of that connection, not an add/remove.
- New `api/connections.ts` client (`list/create/update/delete`), team-scoped
  (`…/teams/{defaultTeam}/stacks/{id}/connections[/{cid}]`), types already
  generated (`StackConnection`, `StackConnectionList`). Connection identity for the
  diff uses the backend discriminator (`from` + `to` + `config`).
- Save is best-effort-atomic: stack PUT first, then connection diff; surface a
  partial-failure toast if any connection call fails (the stack still saved).

### 5. Components / boundaries
- `api/connections.ts` — connections CRUD client (team-scoped). One purpose.
- `lib/connection-mapping.ts` — pure functions: env-rows ↔ connections (group on
  save, expand on load) + output accessors. Unit-tested in isolation.
- `env-row.tsx` / `env-addon-group.tsx` — presentational; props in, callbacks out.
- `form-schema.ts` — the union + serialize/deserialize entry points.
- Save orchestration in the stack edit-session save path (`detail/index.tsx` /
  `use-stack-edit-session`): stack PUT + connection diff.

### 6. Testing
- Unit (`lib/connection-mapping`): round-trip rows ↔ {env vars + connections} for
  each source type; grouping (5 addon fields → 1 connection, 5 mappings);
  output-accessor edge cases (secret special chars).
- Unit: connection diff (added/changed/removed → correct CRUD set).
- `EnvRow` vitest: source switch resets sub-fields; validation messages.
- Live: load the existing ToolJet stack → rows render the addon/secret/mailhog/self
  bindings from its connections → edit one → Save → correct connection CRUD fires →
  reload shows the change.

## Out of scope
- Volume-mount / build-artifact-source connections (env `kind` only).
- A standalone connections/topology graph view.
- Backend changes — uses only the existing connections API.

## Risks
- Save is multi-request (stack PUT + N connection calls) → partial failure UX.
- Connection identity/diff must match the backend discriminator to avoid
  duplicate-create or orphan-delete.
- Output-name reconciliation (old field labels vs backend output names).
