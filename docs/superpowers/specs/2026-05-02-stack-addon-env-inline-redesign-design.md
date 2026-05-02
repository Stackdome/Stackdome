# Stack ↔ Addon Env Linkage — Inline Redesign

**Status:** Implemented
**Date:** 2026-05-02
**Supersedes (UI portion only):** `docs/superpowers/specs/2026-05-01-stack-addon-env-linkage-design.md`

## Context

The first cut of stack ↔ addon env linkage shipped behind a modal (`AddFromAddonDialog`). The modal lets the user pick an addon, a database, optionally a superuser flag, and tick credential fields with editable env-var names. On confirm it inserts N flat `from: "addon"` rows into the env table.

In practice the modal feels like a detour. For the common case of "I want one or two envs from an addon" the user has to leave the env table, fill out a form-within-a-form, and come back. The 1:N expansion the modal optimised for is rarely needed — most stacks pull `DATABASE_URL` (or a small ad-hoc subset) and that's it. The presets baked into the modal are guesses we don't have evidence anyone uses.

This spec replaces the modal with **inline pickers in the existing env table**. Every env var stays one row. When `from === "addon"`, the row grows a second line containing three dropdowns. The user types the env name themselves, exactly like a stack literal; the binding to the addon's credentials happens via clicks.

Backend contract is untouched — `env_from_addons[]` and the form-side discriminated union both stay as today.

## Goals

- Remove the `Add from Addon` modal.
- Keep one env var = one row.
- Make addon-backed envs feel like a small variant of secret-backed envs (which already use inline pickers in the value cell).
- Preserve today's data flow: form-side flat rows ↔ API-side `env_from_addons[]` groups.
- Keep the existing addon-count badge on the collapsed resource header.

## Non-Goals

- Adding new addon types (Redis, S3, etc.). The schema and dispatch pattern are ready for them; the UI work is a localised addition when the time comes.
- Cross-resource env-name uniqueness. Scope stays per-resource as today.
- Bulk import or paste-multiple-rows flows.
- Backend changes.

## Backend Contract Recap

A stack resource's env config has three arrays:

- `environment_variables[]` — literal `{ name, value }` pairs.
- `environment_variables_from_secret[]` — `{ name, secret_id, secret_key }`.
- `env_from_addons[]` — `{ addon_id, database?, superuser?, env_mapping: Record<string, CredField> }`. One entry per `(addon, database, superuser)` triple; the `env_mapping` keys are the env-var names visible to the container; values are credential field names (`host`, `port`, `username`, `password`, `database`, `sslmode`, `connectionString`, `caCertificate`).

The form schema discriminates on `from` (`"stack" | "secret" | "addon"`) and the converter is the only place that translates between flat form rows and the three API arrays. Both directions are covered by `frontend/__tests__/stacks-env-roundtrip.test.ts`.

## Design

### Row anatomy

Stack and Secret rows are unchanged. The `From` column and trailing `[×]` align with line 1 of every row.

```
NAME             VALUE                            FROM
─────────────    ──────────────────────────       ────────
NODE_ENV         production                       Stack ▾    [×]
API_KEY          [my-secret ▾] [api_key ▾]        Secret ▾   [×]
PG_HOST          (line 2 below)                   Addon ▾    [×]
                 [my-pg ▾]  [tooljet ▾]  [host ▾]
```

The addon row's second line contains exactly three pickers:

1. **Addon** — the source addon.
2. **Database** — which logical database within the addon. Includes a special `─ All databases ─` item when the addon supports superuser access.
3. **Field** — which credential field on that addon (`host`, `port`, `username`, `password`, `database`, `sslmode`, `connectionString`, `caCertificate`).

The two-line treatment is intentional: addon rows are 1:N references, not literals, and they carry strictly more state than stack/secret rows. Making them visually fatter matches their semantic weight and avoids cramming three dropdowns into a single value cell.

### Picker behaviour

**Addon picker.**

- Sourced from `usePostgresAddons(orgId)` (already lifted to the stack resources form).
- Each item shows `<name> (Postgres · <state>)`. Linking a non-Ready addon is allowed; the reconciler resolves credentials at deploy time.
- **Empty addon list:** the `SelectContent` shows a single non-selectable item `No Postgres addons yet.` with a `+ Create Postgres addon` link below that opens `/addons/create/postgres` in a new tab. The `From: Addon` option itself stays selectable — discoverability matters.

**Database picker.**

- Disabled until an addon is picked. Placeholder: `Pick an addon first`.
- After an addon is picked, items are:
  - `─ All databases ─` — only if `addon.spec.configuration.enable_superuser_access === true`. Picking it sets `superuser: true` and `database: undefined` under the hood. The user-facing wording is plain English; the postgres "superuser" detail is a backend concern.
  - One item per `addon.spec.databases[].name`.
- Auto-select rule: if there's exactly one database **and** the addon does not support superuser, auto-select it. Otherwise leave empty.
- Changing the addon dropdown to a different addon resets `database` to undefined and `superuser` to false. `credField` is preserved (it's addon-type-scoped; today every addon is postgres).

**Field picker.**

- Disabled until an addon is picked. Placeholder: `Pick an addon first`.
- Items: each entry in `CRED_FIELDS`. Cluster-wide fields (`host`, `port`, `sslmode`, `caCertificate`, sourced from `CLUSTER_WIDE_FIELDS`) get a small muted `cluster` badge inside the dropdown item.

### State transitions on `from` change

| Transition | Effect |
|---|---|
| `stack → secret` | Clear `value`. Init `secretId=""`, `secretKey=""`. Name preserved. |
| `stack → addon` | Clear `value`. Init `addonType="postgres"`, `addonId=""`, `database=undefined`, `superuser=false`, `credField=""`. Name preserved. |
| `secret → stack` | Drop secret fields. Init `value=""`. |
| `secret → addon` | Drop secret fields. Init addon fields as above. |
| `addon → stack` / `addon → secret` | Drop all addon fields. Init the target variant. |

No confirmation prompt — switching destroys transient form state only; the row hasn't been persisted.

### Future addon types

The row's second line is rendered by a small `<AddonInlinePickers />` sub-component dispatching on `addonType`. Today only `"postgres"` is implemented. Adding e.g. Redis later is a localised change: a new branch that renders whichever pickers Redis needs (probably just `[Addon ▾] [Field ▾]` with no database). The `FormEnvVarSchema` discriminated union extends the same way — add a new variant alongside the existing postgres one.

### Validation

**Model: lazy + strict.** Per-row required-field checks fire on **blur out of the row** or on **save attempt**, whichever comes first. Until then, the row stays clean visually.

| Field | Required when | Error wording |
|---|---|---|
| `name` | Always (existing) | `Environment variable name is required` |
| `addonId` | `from === "addon"` | `Pick an addon` |
| `database` | `from === "addon"` AND `superuser === false` | `Pick a database` |
| `credField` | `from === "addon"` | `Pick a field` |

**Duplicate env names.** Live (every keystroke). Scope: within this resource's env list only. Cross-resource collisions are allowed. All rows sharing a duplicate name get a red border; an inline message under the first one reads `Duplicate name "<name>"`. Save is blocked while any duplicate exists. The check is `from`-agnostic — a stack literal `NODE_ENV` collides with an addon row named `NODE_ENV`.

**Save-time error surfacing.** Field-level errors (red border, inline message) on the offending row. The existing summary banner at the top of the env section reads `Fix N issue(s) before saving.` and aggregates the count.

### Save (form → API)

- Stack rows → `environment_variables[]` (unchanged).
- Secret rows → `environment_variables_from_secret[]` (unchanged).
- Addon rows → grouped by `(addonId, database, superuser)`; each group → one `env_from_addons[]` entry (unchanged converter).

Two rows binding to the same `(addon, database, field)` triple with different env names are allowed: the resulting `env_mapping` has multiple keys pointing to the same source field (e.g., `{ PG_HOST: "host", DB_HOST: "host" }`). Backend handles this fine.

### Edge cases

**Orphan (addon was deleted).** On form load, an addon row whose `addonId` isn't in the addon list renders the second line read-only with a soft-yellow strip:

```
PG_HOST    ⚠ <missing addon> · tooljet · host                  Addon ▾    [×]
           Addon was deleted. This variable won't resolve. Remove to clean up.
```

The `From` dropdown and `[×]` still work. Switching `from` away from Addon clears the orphan binding. Save preserves orphan rows as-is so a stale form never destroys references.

**Addon picker shows non-Ready states.** Linking is allowed regardless of state (`Ready` / `Pending` / `Error`).

**Addon supported "All databases" at form load, no longer supports it now.** Row loads as-is (`superuser: true`, `database: undefined`); the dropdown still surfaces `─ All databases ─` as the selected value. User can re-bind manually. Not worth a special warning.

**No addons exist.** `From: Addon` is selectable; opening the addon dropdown shows the `+ Create Postgres addon` empty-state link. Save is blocked with `Pick an addon` if the user attempts to save without resolving.

## Data Model

No schema change. `FormEnvVarSchema` already has all three discriminated-union variants. The converter logic in both directions is unchanged.

## File Plan

**Delete:**

- `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx`
- `frontend/__tests__/add-from-addon-dialog.test.tsx`

**Modify:**

| File | Change |
|---|---|
| `frontend/src/pages/stacks/components/shared/env-row.tsx` | Add the two-line `from: "addon"` rendering: addon/database/field pickers, `─ All databases ─` item, disabled-until-addon-picked behaviour, blur-trigger validation hooks, orphan warning strip. |
| `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` | Remove `AddFromAddonDialog` import and mount; remove the `Add from addon` toolbar button; drop the `addonDialogOpen` state. Keep `addonCount` badge and `existingEnvNames` memo. |
| `frontend/src/pages/stacks/lib/addon-presets.ts` | Delete `applyPreset()`, `DEFAULT_ENV_NAMES`, `Preset`, `PresetResult`. Keep `CRED_FIELDS`, `CLUSTER_WIDE_FIELDS`. |
| `frontend/__tests__/addon-presets.test.ts` | Trim — drop the `DEFAULT_ENV_NAMES` and `applyPreset` describe blocks; keep `CRED_FIELDS` and `CLUSTER_WIDE_FIELDS` blocks. |
| `frontend/__tests__/stacks-env-roundtrip.test.ts` | Mostly intact — converter unchanged. Re-verify addon-row fan-out and group reassembly cases still pass. |

**Add:**

| File | Purpose |
|---|---|
| `frontend/__tests__/env-row-addon.test.tsx` | Component test for the new addon row. Covers: two-line layout when `from === "addon"`; addon picker empty-state link when no addons; database picker disabled until addon picked; `─ All databases ─` only shown when `enable_superuser_access === true`; switching `from` resets variant fields; orphan row renders read-only with warning; lazy+strict validation fires on blur and on save attempt. |

**Untouched:**

- `frontend/src/api/stacks.ts` — existing PUT/GET cover the request shape.
- `frontend/src/pages/addons/hooks/use-postgres-addons.ts` — already lifted and reused.
- Backend.

## Testing

### Unit (vitest)

- `frontend/__tests__/env-row-addon.test.tsx` (new) — see purpose above.
- `frontend/__tests__/stacks-env-roundtrip.test.ts` (existing) — converter round-trip, addon-row fan-out, group reassembly, orphan passthrough.

### Manual (run against the local cluster)

1. Stack with one resource, no env vars. Add a stack literal → save → reload → persists.
2. Add an addon-backed env: type `PG_HOST`, set `From: Addon`, pick addon, pick database, pick `host` → save → reload → row reappears intact.
3. Add 4 more addon-backed envs from the same `(addon, db)` → save → on reload, all 5 rows reappear; on the API side, one `env_from_addons[]` entry with 5 `env_mapping` keys.
4. Pick an addon that supports superuser → confirm `─ All databases ─` shows; pick it → save → reload → row reappears with database empty and "All databases" still highlighted.
5. Type a duplicate name (`NODE_ENV` twice across stack + addon rows) → live red border + save blocked.
6. Flip an existing addon row's `From` to `Stack` → addon fields cleared; row becomes a literal with empty value.
7. Delete the addon backing an existing row externally; reload form → orphan warning strip renders, save still works (orphan preserved).
8. Empty addon list: `From: Addon` → open addon dropdown → see "+ Create Postgres addon" link.

### Out of test scope

- Backend `env_from_addons` reconciliation (covered elsewhere).
- Addon CRUD flows (covered by addon page tests).

## Open Questions

None at design time.

## Future Work (Not In This Spec)

- New addon types (Redis, S3, etc.) and the per-type inline picker dispatch.
- Cross-resource env-name validation.
- A "duplicate this env to another resource" affordance.
