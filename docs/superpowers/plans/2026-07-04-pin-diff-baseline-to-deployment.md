# Pin form diff baseline to last deployment — plan

**Date:** 2026-07-04
**Bug:** editing a deployed stack shows dirty indicators (row tints, reset arrows, tab dots, changes count) only momentarily — the debounced autosave persists the draft, refreshes the stack, and both diff baselines follow it, so changes stop reading as revertable.
**User decision:** diff must always compare against the **latest release** (config last shipped via Deploy, even while the cluster is still converging). Approved 2026-07-04.

## Root cause

- `baselineResources`/`baselineVolumes` (detail/index.tsx:186,210) derive from `stackToShow` — the live draft stack. Autosave → `onStackRefreshed` → `setFetchedStack` → baseline recomputes with the edit baked in.
- `use-draft-sync.ts` additionally calls `session.rebase(snapshot)` after every successful op batch (2 sites), advancing the session's internal baseline to the just-saved draft.
- The deployed config is already available client-side: `releaseDetail.peek(releaseId).data?.snapshot` (`StackReleaseSnapshot` = resources + volumes + connections), currently used only by the whole-stack Discard flow.

## Design

Autosave's op computation uses its own `mirrorRef` (server state) — untouched. Only the *display/revert* baseline is repinned.

1. **`use-stack-edit-session.ts`** — expose `baseline` from the hook (if not already) and extend `start` so the caller can seed draft and baseline separately (`start(baseline, { draft })` or equivalent). Today `start` clones one value into both.
2. **`detail/index.tsx`** —
   - Extract the existing stack→form mapping (resources + connectionsToEnvRows + connectionsToMounts) into a helper usable for both the live stack spec and a release snapshot.
   - Anchor: `baselineReleaseId = releasesResult.activeRelease?.id ?? stackToShow.status.last_converged.release_id`; `releaseDetail.ensure(baselineReleaseId)`; snapshot forms = mapped `peek(...).data.snapshot`.
   - `baselineResources/Volumes` = snapshot forms when available, else current fallback (never-deployed stacks keep today's behavior).
   - Session auto-start: draft = live-stack forms, baseline = snapshot forms (they differ when autosaved edits already exist).
   - Rebase effect: when the session is active and the snapshot for a *new* `baselineReleaseId` arrives (late fetch, or a fresh deploy creating a new activeRelease), `session.rebase(snapshotForms)` — guarded by a last-rebased-release ref so live-stack recomputes never move the baseline.
   - Pass the session's own baseline down to the canvas while the session is active so per-row diffs and session dirty agree.
3. **`use-draft-sync.ts`** — delete both `session.rebase(snapshot)` calls. Baseline now moves only on deploy (new activeRelease → rebase effect) and revert (session.discard + restart).
4. **Semantics shift (intended):** DRAFT pill, "N changes", staged phase now mean "differs from the latest release", not "unsaved edits". Per-row reset restores the deployed value and autosave syncs the restoration.

## Known edges (accepted)

- Positional resource matching: a post-deploy mid-list delete shifts indices vs baseline — pre-existing session behavior, unchanged by this work.
- Rename post-deploy reads as add + delete (name-keyed env rows aside) — pre-existing.
- Never-deployed stacks: baseline = page-load server state; per-row dirty accumulates until first deploy (strictly more revertable than today).
- Failed latest release still anchors the baseline (user chose latest-release semantics).

## Tests

- `use-draft-sync` tests asserting rebase-on-save → update to assert baseline stays pinned.
- Session tests for split draft/baseline start.
- Detail-level test: edit → simulated autosave refresh → row still dirty vs snapshot; deploy (new activeRelease + snapshot) → dirty clears.
- Existing env/config tab tests stay green.

Gate: vitest green + lint + tsc, then live verify on :5174 (edit deployed stack → autosave completes → row stays tinted with reset arrow → reset restores deployed value → deploy clears).
