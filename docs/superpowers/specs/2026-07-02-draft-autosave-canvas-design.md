# Draft Concept + Debounced Autosave for the Stack Canvas — Design

**Date:** 2026-07-02
**Status:** Approved (plan-mode review)
**Branch:** `stack-creation-redesign` (worktree `stack-canvas-editor`)
**Builds on:** 2026-07-02 canvas-only create surface (shipped on this branch)

## Goal

Redefine **draft** as the stack's authored DB state whenever it diverges from
the last deployment (or a never-deployed stack with resources), and persist
canvas edits automatically with a debounce via the backend's thin granular
endpoints. The manual Save button disappears for existing stacks; Deploy
becomes the primary action.

## Decisions (locked with user)

| Question | Decision |
|---|---|
| Autosave transport | Thin granular endpoints (per-resource / per-connection / per-volume) |
| Volume-create gap | Add thin backend endpoint `POST /stacks/{id}/volumes` |
| New-stack first save | Explicit **Create stack** action at `/stacks/new`, then autosave |
| Save button | Removed for existing stacks; autosave indicator + Deploy primary |
| Revert/discard | In scope — revert-to-last-deployed + delete for never-deployed |

## The draft concept

The `stacks` row (+ resources/connections/volumes tables) **is** the draft —
the backend already works this way: Save writes DB only; a release
(`POST /stacks/{id}/releases`) snapshots authored state immutably and applies
CRs. The frontend already derives this as lifecycle phase `staged` in
`frontend/src/pages/stacks/components/detail/deployments/use-deploy-lifecycle.ts`
(saved spec ≠ live snapshot, or never-converged with resources). This work
reframes the UI around that phase (DRAFT pill, staged diff) and collapses the
transient `editing` phase into a background autosave loop.

## Verified constraints (from code)

1. **Connections are id-less in form state.** `splitEnvRows`
   (`frontend/src/pages/stacks/lib/connection-mapping.ts:59-90`) builds
   connections without `id`; ids never round-trip through the form. The
   translator diffs connections by **content identity key**
   (kind | from | to | config discriminator — mirrors backend conflict check
   `pkg/services/stack_service.go:333-343`). The engine's server mirror maps
   identity-key → server id (from `stack.spec.connections`, which carries ids).
2. **Resource rename is delete+create server-side already** —
   `pkg/services/stack_resource_service.go:129` overwrites body name with the
   path name; whole-PUT reconciles by name too. Translator diffs resources by
   name; a rename emits create(new) **before** delete(old).
3. **`POST /volumes` never associates volume↔stack**
   (`pkg/services/volume_service.go:221-246`); association only happens via
   the whole-PUT path (`InternalCreateWithTx`, `volume_service.go:98-114`).
   openapi `Volume` has no `stack_id`. → new backend endpoint (Slice 0).
   Volume **edit/delete deferred**: no canvas affordance exists, and
   `DELETE /volumes/{id}` destroys cluster data synchronously.
4. **Thin writes don't bump `stack.updated_at`** (ReprojectSpec rewrites only
   reference rows) → the engine refetches `getStackById` after every
   successful cycle and writes through to BOTH `setFetchedStack` and the
   stacks context (`stackToShow = currentStack || fetchedStack` at
   `detail/index.tsx:106,154` — a stale context copy wins otherwise).
5. **`pendingDetach` is dead code** (`setPendingDetach` has zero consumers).
   Delete it and the save-time detach dialog rather than relocating.
6. **No cascade on resource delete**
   (`stack_resource_service.go:256-261`) → connection deletes MUST precede
   resource deletes. Translator ordering invariant, encoded as a test.
7. **Node drags cannot thrash autosave** — React Flow positions are local
   state (`StackCanvasTab.tsx:70`), never in `FormStackData`. Only
   per-keystroke drawer edits hit the session; the debounce absorbs them.
8. **No debounce util exists** in `frontend/src` — write a small
   trailing-edge + max-wait scheduler inside the engine.

## Backend contract

| Entity | Create | Update | Delete | Identity |
|---|---|---|---|---|
| Resource | `POST /stacks/{id}/resources` | `PUT …/resources/{name}` | `DELETE …/resources/{name}` | name |
| Connection | `POST …/connections` (201 → id) | `PUT …/connections/{cid}` | `DELETE …/connections/{cid}` | server id (mirror-mapped) |
| Volume | **NEW** `POST /stacks/{id}/volumes` | deferred | deferred | server id |

- Whole-stack `PUT /stacks/{id}` (replace-all connections, upsert by id)
  retained ONLY for revert.
- `DELETE /stacks/{id}` (`cmd/server/routes.go:225`) backs never-deployed delete.
- **New endpoint** `POST /organizations/{org_id}/teams/{team_name}/stacks/{id}/volumes`:
  create + associate + cluster-create in one transaction (reuse
  `InternalCreateWithTx`); body `Volume`, 201 → created `Volume` with id.
  OpenAPI spec updated; Go client + frontend types/zod regenerated.

## Architecture

Three layers — calculations separated from actions (Grokking Simplicity):

### 1. Pure calculation layer — `frontend/src/pages/stacks/lib/draft-sync/`

- `constants.ts` — `DEBOUNCE_IDLE_MS = 1200`, `DEBOUNCE_MAX_WAIT_MS = 5000`,
  retry/backoff constants, `SYNC_STATUS = { idle, saving, saved, error }`.
- `server-state.ts` — `ServerStackState { resourcesByName, volumeIdByName,
  volumesByName, connections: Map<identityKey, {id?, conn}> }`;
  `connectionIdentityKey(c)`; `serverStateFromStack(stack)` (single shared
  read-only-field cleaner).
- `desired-state.ts` — `DesiredStackState { resources, held: Set<name>,
  volumes, connections, resourceIssues: Map<idx, FieldErrors> }`;
  `buildDesiredState(draft)` — per-resource `FormStackResourceSchema.safeParse`
  (export it from `form-schema.ts:77`). Invalid-but-present resources are
  **held**: no ops, server copy + its connections untouched — a half-typed
  resource never reads as "deleted". Issues feed live drawer errors.
- `ops.ts` — `SyncOp` union (createVolume / createResource / updateResource /
  deleteResource / createConnection / updateConnection / deleteConnection);
  `computeSyncOps(server, desired, meta) → SyncOp[]`. **Ordering invariant:**
  createVolume → createResource → updateResource → deleteConnection →
  updateConnection → createConnection → deleteResource.
- `snapshot-to-update.ts` — `snapshotToUpdateRequest(snap, current)`
  (connections keep ids so PUT upserts; current name/labels preserved);
  `volumesToDelete(server, snap)` (PUT never deletes volumes).

### 2. Sync engine — `frontend/src/pages/stacks/hooks/use-draft-sync.ts`

`useDraftSync({enabled, stack, session, ids, onStackRefreshed})
 → {status, lastSavedAt, failureCount, flush(): Promise<boolean>}`

- Debounce on `session.dirty` transitions: idle 1200 ms, max-wait 5 s.
- Single-flight; requeue immediately if dirty on completion.
- Cycle: snapshot draft → `buildDesiredState` → `computeSyncOps(mirror, …)` →
  execute sequentially (mirror updated from each response — connection creates
  adopt server ids) → refetch `getStackById` → `onStackRefreshed` →
  `session.rebase(snapshot)` (post-snapshot edits stay dirty).
- Failure: abort cycle → refetch, rebuild mirror, rebase baseline to server
  state, **keep local draft** (user edits win) → exponential backoff
  (1 s · 2ⁿ, cap 30 s) → `status: error`; sticky destructive toast after 3
  consecutive failures.
- `flush()` runs any pending/in-flight cycle to completion; used by Deploy,
  `visibilitychange → hidden`, and best-effort unmount.
- Session gains exactly one primitive: `rebase(baseline)` — server ids never
  enter form state (the mirror owns them).

### 3. UI reframe

- `CanvasEditorShell.tsx` — primary-button matrix:
  - Draft (`/stacks/new`): primary = **Create stack** (`onCreate`).
  - Existing: primary = **Deploy**, enabled when `staged || editing`
    (click = flush → `createRelease`); no Save button.
  - `<AutosaveStatus/>` replaces "N unsaved changes": "Saving…" /
    "All changes saved" / "Save failed — retrying".
  - `DRAFT` StatusPill when staged (replaces "draft saved · undeployed" text).
  - Overflow menu: "Discard draft changes" (staged) ↔ "Delete stack"
    (never-deployed).
- `detail/index.tsx` — mount engine (`enabled: !isDraft && canWrite`);
  `performSave` shrinks to draft-only `performCreate`; detach dialog + gates
  deleted; deploy flushes first; drawer errors derived from `resourceIssues`.
- New `AutosaveStatus.tsx` — presentational, branded tokens only.
- New `use-stack-revert.ts` — ensure live snapshot (releaseDetail.ensure/peek)
  → `updateStack(snapshotToUpdateRequest(…))` → delete extra volumes →
  refetch → `session.discard()` → lifecycle recomputes `clean`.

## Edge cases

| Case | Behavior |
|---|---|
| Edit during in-flight sync | rebase only to the cycle's own snapshot; post-snapshot edits stay dirty → immediate requeue |
| Rename resource | create(new) before delete(old); connections re-key → delete+create in the same cycle |
| Half-typed resource | held — no ops, no deletion, connections preserved; drawer shows field errors |
| Delete resource with connections | connection deletes first (no backend cascade) |
| Add volume (inline/block) | thin `createVolume` first, then dependents |
| Addon link without env rows | no-op until a row binds (`splitEnvRows` skips incomplete rows) |
| Op failure (409/5xx) | abort → refetch/rebuild/rebase, keep draft, backoff, error status |
| Deploy click | `await flush()`; abort with toast on flush failure |
| Revert (staged) | confirm dialog → snapshot PUT → delete extra volumes → refetch → `session.discard()` |
| Never deployed | "Delete stack" instead of Discard → `DELETE /stacks/{id}` → navigate `/stacks` + context removal |
| `/stacks/new` | engine disabled (no id); explicit Create stack → navigate → engine on |

## Error handling

- Autosave failure never loses local edits: draft state is authoritative
  client-side; baseline rebases to server truth; retry with backoff.
- Deploy aborts (with toast) if the flush fails — never deploys stale state.
- Revert and delete are confirm-gated dialogs (destructive variants).

## Testing

- **Unit (table-driven):** ops (~20 cases incl. ordering invariants —
  index(deleteConnection) < index(deleteResource)); desired-state validity
  gating; connection identity keys; snapshot-to-update stripping/id-retention.
- **jsdom + fake timers:** engine — debounce coalescing, max-wait fire,
  single-flight requeue, rebase vs post-snapshot edits,
  failure → backoff → recovery, flush semantics, mirror id adoption.
- **Component:** shell primary-button matrix (draft/staged/editing/clean ×
  canWrite); AutosaveStatus render.
- **Backend (TDD, gomock):** volume create-for-stack service + handler tests.
- **Playwright live:** edit → thin PUT observed → "All changes saved" →
  reload persists; volume add via new endpoint; deploy; revert; delete.

## Slices (implementation order)

0. **Backend thin volume endpoint** — service + handler + route + OpenAPI +
   client regen (Go + frontend types/zod).
1. **Pure core** — `draft-sync/` calculations + `FormStackResourceSchema`
   export + unit tables.
2. **API clients** — `stack-resources.ts`, `connections.ts` fns,
   `volumes.ts`, `deleteStack` in `stacks.ts`.
3. **Engine** — `session.rebase` + `use-draft-sync.ts` + jsdom tests (not mounted).
4. **Wire autosave** — mount in detail page, AutosaveStatus, remove Save
   (existing), "Create stack" (draft), live drawer errors.
5. **Deploy integration** — flush-then-deploy, Deploy enablement, DRAFT pill,
   visibility/unmount flush.
6. **Revert** — snapshot-to-update wiring, `use-stack-revert`, confirm dialog,
   menu item.
7. **Never-deployed delete** — confirm → `deleteStack` → navigate + context
   removal.
8. **Cleanup** — excise `pendingDetach`, detach dialog, stale save-time
   validation plumbing.

## Open items (deferred, logged)

- Volume edit/delete from the canvas (needs confirm-gated UX; delete destroys
  cluster data synchronously).
- Thin endpoint for name/labels (name is display-only on existing stacks today).
- Optimistic-lock / multi-writer conflict handling beyond refetch-and-rebase.
