# Per-release config diff (snapshot-powered)

**Date:** 2026-06-23
**Branch:** `feat/stateful-deployments`
**Status:** approved

## Problem

The Deploy Timeline shows a config diff between a release and its predecessor
("Config changes · vs #N") in the design. The diff UI was built earlier but was
inert: the `GET /releases/{id}` endpoint returned only the lightweight
`StackRelease`, with no captured spec to diff against.

`main` (#116) now returns `StackReleaseDetail`, which embeds a
`snapshot: StackReleaseSnapshot { stack, resources[], volumes[], connections[], captured_at }`
— the full deploy-time spec for that release. That unblocks a real diff.

## What already exists (post-merge)

The diff flow is wired end-to-end and now live after merging `main`:

- `useReleaseDetail.ensure(id)` → `getRelease` → `StackReleaseDetail` (cached by id).
- `ReleasePostMortem` (rendered on history-row expand) calls `ensure(release.id)`
  and `ensure(prevReleaseId)`, then `diffSnapshots(prevSnap, curSnap)`.
- `ConfigDiff` renders the resulting diff; `OutcomesTable` renders
  `release.outcome.resources`; a "Why it failed" block renders `release.message`.
- `diffSnapshots(prev, cur)` already diffs each resource's image / ports /
  command / args / env vars and classifies whole-resource add/remove.

Confirmed live: clicking history #6 fetches the snapshot and renders
"Config changes · vs #5". It shows the empty state only because every demo
release is the same nginx config (rollbacks), so the diff is genuinely empty.

So this is an **integration + extension**, not a green-field feature.

## Goals

1. Type the snapshot end-to-end (drop `as unknown` casts).
2. Extend the diff to cover **volumes** and **connections**, not just resources.
3. Fix the empty-state copy (distinguish "initial release" from "no changes").
4. Surface the diff on the **active/current** release too (collapsed by default).
5. Verify a real diff renders, live, against a seeded multi-component stack.

Non-goals: pagination/filter UI for the release list (#115 API exists; separate
work), backend changes, diffing read-only/derived fields (status, ids, revision).

## Design

### 1. Types

- `src/api/releases.ts`: `getRelease` returns `StackReleaseDetail`; export
  `StackReleaseDetail` and `StackReleaseSnapshot` type aliases.
- `use-release-detail.ts`: `DetailState.data?: StackReleaseDetail`.
- `release-post-mortem.tsx`: remove the `as unknown as { snapshot?... }` /
  `{ outcome?... }` casts now that the fields are typed.

### 2. Diff engine — `release-snapshot-diff.ts`

Change the return type from `ResourceDiff[]` to a structured `SnapshotDiff`:

```ts
interface ItemDiff { name: string; change: "added" | "removed" | "modified"; rows: DiffRow[]; note?: string }
interface SnapshotDiff {
  resources: ResourceDiff[];   // existing behaviour, unchanged
  volumes: ItemDiff[];         // new
  connections: ItemDiff[];     // new
}
export function diffSnapshots(prev: StackReleaseSnapshot | undefined, cur: StackReleaseSnapshot | undefined): SnapshotDiff
```

- **resources** — keep the current `ResourceDiff[]` logic verbatim (image,
  ports, command, args, env vars; added/removed/modified).
- **volumes** — match `snapshot.volumes[]` by `name`. Scalar rows from
  `VolumeSpec` (size, storage class — exact fields read from the schema during
  implementation). Whole-volume add/remove when a name appears on only one side.
- **connections** — match `snapshot.connections[]` by a stable identity key
  (`kind` + `from` node + `to` node). Rows from `mappings` (env target → value
  source) and `kind`. Whole-connection add/remove by identity. Values that are
  secret/output references render as a label (`(output)` / `(secret)`), never a
  raw value.

`prev == null` (no predecessor) yields an empty `SnapshotDiff` so callers can
tell "initial" from "no changes" by also knowing whether a predecessor exists.

### 3. Rendering — `config-diff.tsx`

- Render `diff.resources` as today (per-resource card with config/env sections).
- Add a **Volumes** group and a **Connections** group, each rendered only when
  its `ItemDiff[]` is non-empty, using the same added/removed/modified visual
  language (green added, struck-through removed, from→to for modified).
- Empty-state copy, driven by a `hasPrev` prop:
  - no predecessor → "Initial release — nothing to compare."
  - predecessor present, all groups empty → "No configuration changes since #N."

### 4. Active-release diff — `current-release-node.tsx`

- Add a collapsed link **"View config changes · vs #N"** on the current card
  (hidden when there is no predecessor). The card stays clean by default —
  consistent with the earlier decision to drop the always-on changelog.
- On expand, render `ConfigDiff` only (diff vs the predecessor). Do **not**
  re-introduce outcomes / why-failed: the live card already shows per-resource
  status rows and the failure banner.
- Thread the predecessor release id/sequence into `CurrentReleaseNode` from
  `TimelineRail` (it already computes `prevIdFor(0)` / `prevSeqFor(0)`), plus the
  `useReleaseDetail` instance so detail fetches share the cache with history rows.

### Data flow

```
list releases (lightweight)            ── timeline rail
  └─ expand history row OR "View changes" on active card
       └─ useReleaseDetail.ensure(N), ensure(N-1)   → GET /releases/{id} (snapshot)
            └─ diffSnapshots(prev.snapshot, cur.snapshot) → SnapshotDiff
                 └─ ConfigDiff renders resources + volumes + connections
```

### Error / edge handling

- Detail still loading → "Loading release detail…" (existing).
- Detail fetch error → inline error (existing).
- No predecessor → "Initial release" copy; no fetch of a non-existent prev.
- Identical snapshots → "No configuration changes since #N".
- Secret/output-backed env or connection values are labelled, never printed.

## Testing

- `release-snapshot-diff.test.ts` — add volume add/remove/modify and connection
  add/remove/modify cases; assert resource diffs still behave; assert
  secret/output values are masked.
- `config-diff.test.tsx` — empty-state copy for initial vs no-changes; volumes
  and connections sections appear only when non-empty.
- `current-release-node.test.tsx` — "View config changes" toggles the diff;
  hidden when no predecessor.
- All scoped/serial (`vitest run <file> --maxWorkers=1 --no-file-parallelism`).

## Live verification

1. Seed a multi-component stack: `hack/run_local.sh samples/tooljet_addon_postgres.json`
   (resources + a postgres connection + a volume).
2. Deploy → release #1.
3. Change the image tag, add/modify an env var, and tweak a connection mapping;
   redeploy → release #2.
4. On the active card, open "View config changes · vs #1" and screenshot the
   rendered diff (resource image/env + connection + volume changes).
5. Delete the throwaway stack to free the cluster.

## Files touched

`src/api/releases.ts`, `use-release-detail.ts`, `release-snapshot-diff.ts`,
`timeline/config-diff.tsx`, `timeline/release-post-mortem.tsx`,
`timeline/current-release-node.tsx`, plus the three test files above.
