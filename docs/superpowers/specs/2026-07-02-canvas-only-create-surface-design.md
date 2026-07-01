# Canvas-Only Create Surface — Design

**Date:** 2026-07-02
**Status:** Approved-in-conversation, pending written review
**Branch:** `stack-creation-redesign` (worktree `stack-canvas-editor`)

## Goal

Make the React Flow canvas editor the **single** surface for creating *and*
editing a stack. Every wizard entry point ends in the canvas. The legacy stack
form is deleted and can never render or be fallen back to.

## Hard constraints (global)

- **The legacy form is gone.** `StackCreatePage` (`frontend/src/pages/stacks/components/create/`) and the `/stacks/create` route are deleted. No code path may navigate to it, render it, or fall back to it. The user must never see the whole-form view again.
- Brand design system only (`index.css` tokens + `branded/`/`ui/` primitives). No raw hex, no off-scale type.
- No third-party-PaaS names ("Railway" etc.) anywhere — code, copy, commits.
- Pure calc/data separation kept: `deriveGraph`, `layoutGraph`, `nodePresentation`, seed adapters stay pure + unit-tested; components stay views.
- Constants over magic strings.
- Canvas feature flag stays default-ON; this work assumes it on. (`VITE_STACK_CANVAS=false` opt-out still routes to… nothing now — see "Flag off" below.)

## The core problem

The canvas lives on `StackDetailPage` (`/stacks/:id`), built around an
**existing** stack id — it fetches by id. The form (`/stacks/create`) is the
only surface that works on a stack that does not exist yet. Deleting the form
forces the canvas to handle an **unsaved draft** (no id).

## Architecture

### One route, draft-aware

- **New route** `/stacks/new` → same `StackDetailPage`, rendered in **draft mode** (`isDraft = !id`, from `useParams`).
- Draft mode:
  - **No fetch** (`getStackById` skipped).
  - Edit session seeded from `location.state.seed` (a `FormStackData`) and `location.state.linkedAddonIds` (`string[]`). Empty seed → empty canvas (blank slate).
  - **Baseline is empty** — every resource/volume/link reads as "new/added" (drives dirty marks + counts naturally).
  - Session starts automatically on mount (no "Edit" click needed — a draft is always in edit mode).
- **Save (draft)** → validate via `FormStackSchema.safeParse` → `createStack(orgId, defaultTeamName, convertFormStackToApiStack(...))` → `navigate('/stacks/:id', { replace: true })`. Nothing persists until Save. This removes the throwaway-stack litter created by the old eager-create path.
- **Deploy** stays disabled until the first Save (reuses the existing "must save before deploy" lifecycle — a draft has no id to release).

### Data flow (create)

```
wizard path → builds FormStackData seed
            → navigate('/stacks/new', { state: { seed, linkedAddonIds } })
            → StackDetailPage(draft) → session.start({resources, volumes}, {linkedAddonIds})
            → user edits on canvas (name, labels, resources, volumes, addons)
            → Save → createStack → /stacks/:id (real stack, session ends, page refetches)
```

### Feature flag removed (required, not optional)

`isCanvasEnabled()`'s **off-branch rendered the legacy form** — which is being
deleted. So the flag cannot stay: its opt-out (`VITE_STACK_CANVAS=false` /
`localStorage.stackCanvas="0"`) has nothing to render. `isCanvasEnabled` and its
call sites are removed and collapsed to unconditional canvas:
- `StackDetailPage` — drop the flag branch; always render `CanvasEditorShell`.
- `app-layout.tsx` — the full-bleed detection `isCanvasEnabled() && /^\/stacks\/[^/]+$/.test(pathname) && !pathname.endsWith("/new")` currently **excludes** `/new`. Flip it: the draft canvas at `/stacks/new` **must** be full-bleed. New rule: full-bleed for `/^\/stacks\/(new|[^/]+)$/` (both the draft route and `/stacks/:id`), no flag gate.
- `feature-flags.ts` + its tests — delete `isCanvasEnabled` (and the file if nothing else lives there).

## Components

### 1. `StackDetailPage` — draft mode (`detail/index.tsx`)

- Reads `isDraft = !id`.
- Draft branch: skip fetch/loader/error; derive `stackToShow` from the session's seed; empty `baselineResources`/`baselineVolumes`; auto-start session from `location.state.seed`.
- Save handler forks: `isDraft ? createStack(...) : updateStack(...)`. On draft-save success, `navigate('/stacks/:id', {replace:true})` and clear nav state.
- Ops tabs (Deployments/Logs/Metrics) in draft render a muted empty state (see component 5). Configuration is the only live tab.
- Status pill: none in draft (no `Ready`/`Pending` state yet); optional muted `DRAFT` marker.

### 2. Canvas header — editable name + labels (`CanvasEditorShell.tsx`)

- The `h1` title becomes an **inline-editable input** styled to match (`text-[29px] font-medium`), placeholder `name-your-stack` in draft. Value flows to session (new `session.setName` or a page-level `draftName` state threaded through).
- A compact **labels** row beneath the title: existing labels as removable chips + an "add label" input (Enter to add), mirroring the form's label UX. Backed by session/page label state.
- **Name editability:** editable in draft always. On an existing saved stack, name is **display-only** for this iteration (renaming a live stack renames k8s resources — out of scope; can be a follow-up). Labels editable in both draft and existing.
- New props: `stackName` becomes editable → add `onNameChange`, `nameEditable`, `labels`, `onAddLabel`, `onRemoveLabel`. Shell stays presentation-only (owns no stack state).

### 3. Resource drawer — inline volume create (`stack-resource-configuration-tab.tsx` + drawer)

- Replace the drawer's "Volume Mounts → select existing volume" with **create-or-pick**:
  - "Add volume" spawns a row: **name** + **size** + **mount path** (target_path), with storage class / access mode behind an "advanced" disclosure (defaults: `ReadWriteOnce`, default storage class).
  - Writing a row creates BOTH a stack-level `volume` (name/size/spec) *and* the resource's `volume_mount` (source_volume_name = that name, target_path). Kept in sync on edit/remove.
- The select-existing dropdown remains available (pick an already-defined volume) but no longer blocks when the volume list is empty — you can author one inline.
- Model note: volumes remain stack-level in the API; this is purely a UX relocation. A volume authored here is named for its resource by default (e.g. `<resource>-data`) but the name is editable. RWO means one mount in practice; the UI does not hard-forbid reuse.
- Deletes the standalone `StackVolumesForm` from the create surface (it was only reachable via the form).

### 4. Managed-addon linking in "+ Add" (`AddResourcePopover.tsx`)

- Extend the popover (which today only lists block catalog via `BlockPicker`) with a **"Managed add-ons"** section below the blocks, sourced from `usePostgresAddons()` — same list the wizard composer shows.
- Selecting a managed addon calls a new `onLinkAddon(addonId)` → the canvas adds it to the session's linked-addon set (→ `connectionAddonIds`) and renders it as an addon node. Drawer env tab then binds its creds (already works).
- Without this, a blank-canvas user cannot attach an addon at all (no form fallback). Required, not optional.

### 5. Draft ops-tab empty states

- A small shared `DraftTabPlaceholder` (or inline blocks) rendered for Deployments/Logs/Metrics when `isDraft`: muted icon + "Available after you save" + a Save hint. No API calls fire in draft.

### 6. Wizard rewire (`stack-create-wizard.tsx`, `block-composer.tsx`, `use-template-import.ts`, `use-docker-compose-import.ts`, `wizard-chooser` blank)

- Every path converges on: build `FormStackData` seed → `navigate('/stacks/new', { state: { seed, linkedAddonIds } })`.
  - **Blocks** (`block-composer`): remove `createAndOpen` eager-create, the name `Input`, `creating` state, the `openFormEditor` fallback, and the toast-on-fail. `Continue` → seed (`emptyStack`-shaped `FormStackData` from composed resources) + `Array.from(selectedAddonIds)` → `/stacks/new`. Name now entered in canvas.
  - **Template** (`use-template-import`): `templateToFormData(t)` → seed → `/stacks/new` (drop `ImportSource`/`importWarnings` plumbing to the form; carry warnings via nav state if still surfaced, else drop).
  - **Compose** (`use-docker-compose-import`): converted data → seed → `/stacks/new`.
  - **Blank** (`wizard-chooser` → `stack-create-wizard`): empty seed → `/stacks/new`.
- `nav-stacks.tsx:22` active-state check `/stacks/create` → `/stacks/new`.

### 7. Deletions

- Route `/stacks/create` + `StackCreatePage` import in `App.tsx`.
- `frontend/src/pages/stacks/components/create/` directory (page + any create-only helpers) and its tests.
- `ImportSource` prefill plumbing that only served the form (`isPrefillSource` consumers in the create page). Lean to removing `ImportSource` unless the seed still needs provenance tags in the canvas.
- `isCanvasEnabled` + `feature-flags.ts` + its tests (see "Feature flag removed").

## Error handling

- Draft Save validation failure → same field-error mapping the form used (`FormStackSchema` issues → per-resource errors surfaced on the relevant node/drawer), plus a toast. No navigation.
- `createStack` API failure → toast (destructive), stay in draft (no navigation, no data loss). **No fallback to the form** — the draft remains editable.
- Missing org/team → toast, stay in draft.
- Empty draft Save (no resources) → blocked by `FormStackSchema` (stack needs ≥1 resource); surfaced as a validation toast.

## Testing

- **Unit (pure):** seed adapters (template/compose/blocks/blank → `FormStackData`) — table-driven. Inline-volume reducer (add row → produces matching `volume` + `volume_mount`; remove → drops both). Name/label reducers.
- **Component (jsdom):** `CanvasEditorShell` renders editable name in draft, display-only on existing; label add/remove. `AddResourcePopover` lists + links a managed addon. Draft ops-tab placeholders render.
- **Integration/Playwright (per slice):** each wizard path → `/stacks/new` → canvas renders seeded graph → name it → Save → lands `/stacks/:id`. Blank slate → add resource + inline volume + link addon → Save. Assert `/stacks/create` is unreachable (route removed).
- Existing 515-test suite stays green; feature-flag tests updated/removed with the flag.

## Slices (implementation order)

1. **Draft-mode detail page** — `/stacks/new` route, `isDraft` branch, seed→session, Save=create, no fetch. (Blank seed only.)
2. **Header name + labels** — editable title + labels row in shell; wire to session/page state; draft-save uses the typed name.
3. **Drawer inline volume create** — create-or-pick in Configuration tab; sync volume + mount.
4. **Addon in "+ Add"** — extend popover + `onLinkAddon`; addon node in draft.
5. **Wizard rewire** — all four paths → `/stacks/new`; simplify `block-composer`; update `nav-stacks`.
6. **Delete the form** — remove route, page, tests, prefill plumbing, flag opt-out. Verify `/stacks/create` 404s and the form is unreachable from every path.

## Open items (deferred, logged)

- Rename an already-saved stack from the canvas header (needs rename-safe backend semantics).
- Non-postgres managed addon types in the "+ Add" list (only postgres today).
- Template import warnings surfacing in the canvas (were shown on the form).
