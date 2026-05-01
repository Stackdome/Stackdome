# Stack ↔ Addon Linkage via Environment Variables — Frontend Design

**Date:** 2026-05-01
**Status:** Implemented (2026-05-01)
**Author:** Claude

## Context

Postgres addons can be created and listed in the UI as of the previous spec, but stacks cannot yet consume them. The only backend mechanism that connects a stack to an addon is `env_from_addons` on each `ExecutionConfig` — a per-resource array that resolves credentials at deploy time and (as a side effect) writes rows into the `addon_usages` table the backend uses to gate addon deletion. There is no separate "attach addon to stack" API.

Today the stack edit form **silently drops `env_from_addons`** on save. The form converter (`pages/stacks/schemas/form-schema.ts:197-216, 277-294`) reads only `environment_variables` and `environment_variables_from_secret`, and writes only those two arrays back. Any stack created from `samples/tooljet_with_addon.json` (including the running `tooljet-addon` stack) will lose its addon link the first time someone opens it in the UI and clicks Save. This is the round-trip bug we must fix as part of this feature.

This spec covers v1 of the addon-linkage UI: a unified environment-variable list with a `From` column, a mass-action dialog for inserting addon-backed env vars, a count badge on the collapsed resource header, and the converter fix that round-trips the data.

## Goals

1. Let users link a Postgres addon's credentials into a stack resource as environment variables, without copy-pasting plaintext values.
2. Round-trip `env_from_addons` losslessly through the stack edit form.
3. Show at a glance which resources in a stack consume addons.
4. Use only existing backend endpoints — no backend changes.

## Non-Goals

- Reverse view ("which stacks use this addon") on the addon detail page. Deferred until that page exists and a usage endpoint is added.
- Stack-list or stack-header rollups of addon counts.
- Linking non-Postgres addon types — Postgres is the only addon type today.
- Bulk-edit of addon mappings across multiple resources.

## Backend Contract Recap

`PUT /api/v1/organizations/{org}/stacks/{id}` is the only write surface. Each resource's `execution_config` carries three sibling env arrays:

```jsonc
"execution_config": {
  "environment_variables":             [{ "name": "NODE_ENV", "value": "production" }],
  "environment_variables_from_secret": [{ "name": "STRIPE_KEY", "secret_ref": { "secret_id": "…" }, "key": "stripe_live" }],
  "env_from_addons": [{
    "postgres": {
      "addon_id":   "57fa98c8-…",
      "database":   "tooljet",
      "superuser":  false,
      "env_mapping": { "host": "PG_HOST", "port": "PG_PORT", "username": "PG_USER", "password": "PG_PASS", "database": "PG_DB" }
    }
  }]
}
```

The stack reconciler reads `env_from_addons` per resource, resolves credentials JIT from the cluster, materializes the mapped env vars onto the running pods, and upserts `addon_usages` rows. None of that is the frontend's concern. The frontend only has to build the request body correctly and parse it back.

## Design

### User flow

The Environment Variables tab inside each stack resource gains a `From` column and a `⚙ Add from addon` toolbar button. Clicking that button opens a small dialog that resolves the 1:N expansion (one addon source typically produces 5+ env vars). On confirm the dialog inserts those rows into the same env list, each tagged `From: Addon`. The list above is the single source of truth — no separate "Linked Addons" section, no peer tab.

### The unified env table

Three row types share one table. The `From` column drives rendering:

| `From`   | Key column      | Value column                                     | Editable structurally?            |
|----------|-----------------|--------------------------------------------------|-----------------------------------|
| `Stack`  | env var name    | text input (literal value)                       | yes (today's flow)                |
| `Secret` | env var name    | secret picker + key picker (today's flow)        | yes (today's flow)                |
| `Addon`  | env var name    | read-only pill `⚙ <addon> · <db> · <field>`      | name editable; binding read-only  |

To re-bind an addon row to a different addon/database/credField, remove it and re-run the dialog. Switching the `From` value away from `Addon` clears the binding (with a small confirm).

### `From` column wording

Locked: `Stack | Secret | Addon`. ("Stack" reads as "the literal value defined on this stack" — distinct from values that come from a secret store or an addon. Reviewable; spec can swap to `Inline` or `Value` if the team prefers.)

### Toolbar buttons (above the env list)

```
[ × Clear ]   [ ▣ Paste ]   [ ⚙ Add from addon ]
```

The bottom-right `+ Add Variable` button stays for adding a single Stack row. `Paste` retains today's behavior (parses KEY=VALUE lines into Stack rows).

### `Add from addon` dialog

```
┌────────────────────────────────────────────────────────────────────┐
│  Add from Addon                                                ×  │
│  Inject credentials from an addon as environment variables.       │
├────────────────────────────────────────────────────────────────────┤
│  Addon       [ tooljet-db   (Postgres · Ready)               ▾ ]  │
│  Database    [ tooljet                                       ▾ ]  │
│  Superuser   ○ off          (only shown if addon allows it)       │
│                                                                    │
│  Inject credentials                            [ Apply preset ▾ ] │
│  ┌──────────────────────────────────────────────────────────┐     │
│  │ ✓  host              →  [ PG_HOST              ] cluster │     │
│  │ ✓  port              →  [ PG_PORT              ] cluster │     │
│  │ ✓  username          →  [ PG_USER              ]         │     │
│  │ ✓  password          →  [ PG_PASS              ]         │     │
│  │ ✓  database          →  [ PG_DB                ]         │     │
│  │ ☐  sslmode           →  [                      ] cluster │     │
│  │ ☐  connectionString  →  [                      ]         │     │
│  │ ☐  caCertificate     →  [                      ] cluster │     │
│  └──────────────────────────────────────────────────────────┘     │
│                                                                    │
│  Adds 5 environment variables.                                     │
│                                          [ Cancel ]   [ Add ]     │
└────────────────────────────────────────────────────────────────────┘
```

- **Addon picker**. Sourced from `listPostgresAddons(orgId)`. Shows name + type + status (`Ready` / `Pending` / `Error`). Linking a non-Ready addon is allowed — by stack deploy time the reconciler will pull credentials in.
- **Database picker**. Sourced from the chosen addon's `spec.databases[].name`. Mandatory unless `Superuser` is on. If the addon has only one database, auto-select it.
- **Superuser toggle**. Hidden unless `addon.spec.configuration.enable_superuser_access === true`. When on, `Database` field is disabled (defaults to `postgres` server-side).
- **Credential rows**. Eight fixed rows — one per credential field the backend supports. Toggling a row on prefills the env-var name with a sensible default (`host → PG_HOST`, `port → PG_PORT`, `username → PG_USER`, `password → PG_PASS`, `database → PG_DB`, `sslmode → PG_SSLMODE`, `connectionString → DATABASE_URL`, `caCertificate → PG_CA_CERT`). Editable by the user.
- **`cluster` annotation**. Tiny right-aligned muted label on the four cluster-wide fields (`host`, `port`, `sslmode`, `caCertificate`). Tooltip: "This value is the same across every database of this addon." Helps users understand why they still have to pick a database when they only want host/port.
- **Apply preset** menu — three options: **Postgres conventions** (ticks the standard 5 with default names), **Connection string only** (`connectionString → DATABASE_URL`, all others off), **Clear**.
- **Validation**. At least one credential row must be ticked; every ticked row must have a non-empty env-var name; env-var names within the dialog must be unique; on confirm, env-var names must not already exist anywhere in the resource's env list (collision check).
- **Confirm**. Dialog closes; one row per ticked field gets inserted into the env list above, each marked `From: Addon` with `(addon_id, database, superuser, credField)` carried in form state.
- **Empty addon list**. If `listPostgresAddons` returns zero addons, the picker shows an empty state with a `+ Create Postgres addon` action that opens `/addons/create/postgres` in a new tab.

### Collapsed resource header badge

```
●  tooljet                                       ⚙ 1 addon       ⌄
   tooljet/tooljet-ce:latest
```

Badge text = count of **unique addon IDs** referenced by addon-source rows in this resource's env list. Two databases of the same Postgres addon = `1 addon`. Hidden when count is zero. Position: right side of the collapsed header, before the chevron.

### Edge cases

#### Addon was deleted

Form load detects an addon-source row whose `addon_id` is not in `listPostgresAddons`. The row renders read-only with a soft-yellow strip, a single-line warning beneath it, and only the `×` button works:

```
⚠ PG_HOST   ⚙ <missing addon> · tooljet · host           Addon ▾    ×
            Addon was deleted. This variable won't resolve. Remove to clean up.
```

The warning is informational — the form still saves orphaned rows back to the API as-is, so the user can review before removing. Orphans count toward the collapsed header's addon badge (they are still references); the orphan visual surfaces only on expansion. On save, orphaned rows are still emitted; that matches today's silent-passthrough behavior and avoids destroying user data on a stale form.

#### Addon not Ready yet

The picker still shows it. Status pill makes the state visible. Linking it is allowed. The stack reconciler requeues every 30s waiting for the addon to become Ready (`addon_env_reconciler.go:12`). No special UI handling needed — the addon list already polls and the picker label updates as the state transitions.

#### Same env-var name appears twice

Validation rejects on submit:
- Within an addon dialog: enforced before confirm.
- Across the resource's env list (mixing Stack + Addon rows that produce the same name): inline error on the second occurrence; save disabled.

#### `connectionString` is on, plus individual fields

Allowed — backend resolves both, user gets `DATABASE_URL` and `PG_HOST`, `PG_PASS`, etc. side by side.

## Data Model

### Form-side schema additions

`pages/stacks/schemas/form-schema.ts`

```ts
const FormEnvVarSchema = z.discriminatedUnion("from", [
  z.object({
    from: z.literal("stack"),
    name: z.string().min(1),
    value: z.string(),
  }),
  z.object({
    from: z.literal("secret"),
    name: z.string().min(1),
    secretId: z.string().min(1),
    secretKey: z.string().min(1),
  }),
  z.object({
    from: z.literal("addon"),
    name: z.string().min(1),
    addonType: z.literal("postgres"),  // discriminator for future addon types
    addonId: z.string().min(1),
    database: z.string().optional(),    // omitted only when superuser
    superuser: z.boolean().default(false),
    credField: z.enum(["host","port","username","password","database","sslmode","connectionString","caCertificate"]),
  }),
])
```

The existing `FormEnvVarSchema` is replaced with this discriminated union. The migration drops the boolean `useSecret`/`selectedSecretId`/`selectedSecretKey` fields; their data is conveyed through the `from` discriminator instead.

### Converter changes

`convertApiResourceToFormResource` (load):

1. Map `environment_variables[]` → rows with `from: "stack"`.
2. Map `environment_variables_from_secret[]` → rows with `from: "secret"`.
3. **New**: Map `env_from_addons[]` → fan out each entry's `env_mapping` into one row per `(credField → envName)` pair, all sharing the same `addonId`/`database`/`superuser`/`addonType`. Order within an addon group: by credential field (host, port, username, password, database, sslmode, connectionString, caCertificate).

`convertFormStackToApiStack` (save):

1. Rows with `from === "stack"` → `environment_variables[]`.
2. Rows with `from === "secret"` → `environment_variables_from_secret[]`.
3. **New**: Rows with `from === "addon"` → group by `(addonId, database, superuser)`; each group becomes one `env_from_addons[]` entry whose `env_mapping` is rebuilt from the rows. Empty groups (no fields) are dropped. When `superuser === true`, the entry's `database` field is omitted (the backend defaults to `postgres`). Order of entries: sorted by `addonId` then `database` for deterministic output.

The discriminated union plus per-direction tests guarantee the round trip.

## File Plan

**New**

- `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx` — the dialog. Receives the resource's existing env var names for collision checks, the addon list, and an `onAdd(rows: AddonRow[])` callback. ~200 LOC.
- `frontend/src/pages/stacks/components/shared/env-row.tsx` — renders one row in the unified table; switches on `from`. Extracted from the current inline JSX to keep `stack-resource-item.tsx` from growing further. ~120 LOC.
- `frontend/src/pages/stacks/lib/addon-presets.ts` — preset definitions and default env-var-name mapping per credential field.
- `frontend/__tests__/stacks-env-roundtrip.test.ts` — vitest covering the converter both directions, addon row fan-out, group reassembly, orphan passthrough.
- `frontend/__tests__/add-from-addon-dialog.test.tsx` — component test covering presets, validation, superuser hiding, single-database auto-select.

**Modified**

- `frontend/src/pages/stacks/schemas/form-schema.ts` — replace `FormEnvVarSchema` with the discriminated union; rewrite both converters per "Converter changes" above.
- `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` — replace the existing env-vars rendering (~250 LOC) with the unified table that delegates row rendering to `env-row.tsx`. Add the `Add from addon` toolbar button. Add the addon-count badge in the accordion trigger header.
- `frontend/src/pages/stacks/hooks/` — add `use-postgres-addons` (or reuse the existing addons hook from `pages/addons/hooks/use-postgres-addons.ts`) to fetch the addon list once per stack form session.

**Untouched**

- `frontend/src/api/stacks.ts` — no changes; existing PUT/GET cover the entire request shape already.
- Backend — no changes.

## Testing

### Unit (vitest)

- **Round-trip converter**: 12+ cases. Pure literal env vars; pure secret-backed; pure addon-backed (one DB; multiple DBs of same addon; multi-addon); mixed; orphan addon row passthrough; empty arrays; superuser entry; an entry with all 8 credential fields.
- **Group reassembly**: when two addon rows share `(addonId, database, superuser)`, save produces one `env_from_addons` entry; when they differ in `database`, two entries.
- **Dialog component**: presets fill correct defaults; superuser hides when `enable_superuser_access` is false; single-database auto-selects; collision against existing env list blocks confirm; ticking with empty env name blocks confirm.

### Manual (run against the local cluster)

1. Open the running `tooljet-addon` stack in the edit form.
2. Confirm the env list shows the existing `env_from_addons` rows (PG_HOST, …, TOOLJET_DB_HOST, …) marked `From: Addon` with the right captions.
3. Save without changes; verify via PUT body in DevTools that `env_from_addons` is preserved exactly. Verify the running pod env unchanged.
4. Click `Add from addon`. Pick `tooljet-db`. Pick `app` (third database). Apply Postgres conventions, rename `PG_USER → APP_USER`, save.
5. Verify the new `env_from_addons` entry appears in the PUT body, addon_usages row updates after reconcile, pod env reflects new vars after rollout.
6. Delete the linked addon while the stack still references it; reload the form. Verify the orphan banner renders; save still works; row persists.
7. Re-create an addon with the same name (different ID); reload. Orphan banner persists (different ID). Remove the row, link the new addon, save.
8. Collapse a resource accordion: badge shows `⚙ 2 addons` for a service that uses two distinct addons; `⚙ 1 addon` for a service with two databases of the same addon.

### Out of test scope

Reconciler behavior (covered by backend integration tests, indirectly).

## Open Questions

None blocking. Two stylistic calls flagged for review during implementation:

1. **`From` value for literals**: `Stack` vs `Inline` vs `Value`. Spec uses `Stack`; trivial to swap.
2. **Cluster-wide annotation in dialog**: spec includes the small `cluster` label; can be removed if it adds noise during real-use testing.

## Future Work (Not In This Spec)

- Reverse view "Used by" panel on the addon detail page (requires a backend `GET /addons/postgres/{id}/stacks` endpoint and the addon detail page itself, both deferred).
- Bulk-import env vars from an addon across multiple resources at once.
- Inline editing of an addon row's binding (addon, database, credField) without remove + re-add.
- Connection-string-only preset baked as a default for stacks created from templates.
- Surfacing `addon_usages` reconciler health in the UI (e.g. requeue counter when an addon is stuck Pending).
