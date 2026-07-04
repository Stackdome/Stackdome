# Canvas Editor Header — Declutter Redesign

**Date:** 2026-07-04
**Status:** Approved (design)
**Component:** `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx` + supporting
**Visual reference:** artifact `redesign-variants` (Variant A · Action rail)

## Problem

The editor top bar is cluttered. The title row carries name, editable name input, two
status pills (live status **and** DRAFT), autosave text, the Deploy button, and an actions
menu; below it a labels row and a spread-out public-endpoints strip. Users flagged: too many
competing status signals, Deploy buried in the title row, labels/edit rarely wanted, endpoints
visually noisy.

## Goals

1. Title row = **identity only**: name + single status pill.
2. Move **Deploy** to the tab rail.
3. Drop **labels** and existing-stack **name editing** (draft keeps a name field).
4. Collapse **two status pills into one** (remove DRAFT pill).
5. Funnel every undeployed change into one **"View changes (N)"** entry that opens a
   reviewable, per-change-discardable diff modal.
6. Make **public endpoints** one clean horizontal row.

Non-goals: changing autosave/deploy/session logic; touching the ops tab bodies
(deployments/logs/metrics); redesigning the canvas body or drawers.

## Chosen design — Variant A (Action rail)

### Layout (expanded, existing stack)

```
[▾] payments-api  ● READY                                    ✓ All changes saved
────────────────────────────────────────────────────────────────────────────────
▦ Configuration   🚀 Deployments   ▤ Logs   📈 Metrics    [View changes (3)] [Deploy] [⋯]
────────────────────────────────────────────────────────────────────────────────
PUBLIC   🌐 web → app.example.com ↗ ⧉   🌐 api → api.example.com ↗ ⧉
```

- **Title row:** chevron · name (read-only `<h1>`, or `<Input>` when `isDraft`) · single
  status pill · spacer · autosave readout. Nothing else. Fixed height across clean/dirty.
- **Tab rail (right cluster):** `View changes (N)` (warn-toned) → `Deploy` (primary) → `⋯`.
  Order is deliberate: review reads left-to-right into commit.
- **Endpoints row:** one `PUBLIC` label + compact `svc → host ↗ ⧉` chips (one border each,
  copy on hover, port kept in tooltip).

### State rules

| Signal | Clean | Dirty (unsaved or staged) |
|---|---|---|
| `View changes (N)` entry | hidden | shown, `N = dirtyTotal` |
| Deploy button | **disabled** (kept visible) | enabled |
| Status pill | live status only | live status only (never DRAFT) |
| Autosave | `✓ All changes saved` | `⟳ Saving…` / `⚠ Save failed — retrying` |

"Dirty" for the entry/badge = `hasUnsaved || isStaged` (an open session with pending edits, or
a saved-but-undeployed diff). Deploy enable condition is unchanged from today
(`!deployBusy && canWrite && (isStaged || hasUnsaved)`).

### Draft mode

Nothing exists server-side, so there are no changes to view. Draft keeps:
- Title row: **editable name `<Input>`** + spacer (no status pill, no autosave).
- Tab rail right: single **`Create stack`** button (no View changes, no Deploy, no ⋯).

### Actions menu (`⋯`)

Slimmed and moved to the tab rail. Retains **Delete stack** (destructive; gated by
`canDeleteStack`) and **Discard draft changes** (gated by `canDiscardDraft`, when distinct from
the modal's discard). "Discard all changes" **leaves** the menu — it now lives in the modal
footer.

## View-changes modal

Opened by `View changes (N)`. Reuses the deployments diff visual language.

```
┌ Undeployed changes  (3)                              payments-api ┐
│ ● web       Modified                              [ Discard ]     │
│     image     nginx:1.25 → nginx:1.27                             │
│     replicas  2 → 3                                               │
│ ● worker     Added                                [ Discard ]     │
│     image     — → redis:7                                         │
│ ● pgdata     Removed                              [ Discard ]     │
│     volume    10Gi · /var/lib/pg                                  │
├──────────────────────────────────────────────────────────────────┤
│ [Discard all]                       [Close]  [🚀 Deploy 3 changes]│
└──────────────────────────────────────────────────────────────────┘
```

- **Cards:** status dot (green added / info modified / red removed) · resource name (mono) ·
  Added/Modified/Removed badge · field rows `from → to` — the same structure as
  `deployments/timeline/config-diff.tsx` (`ResourceDiff` / `ItemDiff` / `DiffRow`).
- **Per-card Discard:** wired to the existing revert functions in `lib/stack-diff.ts`
  (`revertResource`, `revertVolume`, `revertEnvRow`, `revertResourceField`). No new revert logic.
- **Source of the diff:** the live edit-session diff (`lib/stack-diff.ts`), adapted to the
  `ConfigDiff` render shape. (The deployments `ConfigDiff` is display-only today; this modal is
  the first consumer to pair it with the session's granular revert.)
- **Footer Deploy:** calls the same `onDeploy` the header uses. While the modal is open it is
  the active deploy surface — the header Deploy is **replaced** (not shown as a second live
  button).
- **Empty state:** discarding the last change shows "No pending changes"; on close the
  `View changes` entry is gone and Deploy returns to disabled — matching clean state.

## Component changes

- **`CanvasEditorShell.tsx`**
  - Remove: label props/state/render (`labels`, `labelsEditable`, `onAddLabel`, `onRemoveLabel`,
    `labelInput`, the labels row) and the DRAFT `StatusPill` (`:290–292`).
  - Move Deploy button out of the title row into the tab rail right cluster; add the
    `View changes (N)` control and relocate the `⋯` menu beside it.
  - Title row reduces to chevron · name · status pill · autosave.
  - Keep: collapse chevron + ⌘/Ctrl-. shortcut + `localStorage` persistence (unchanged).
  - Collapsed 44px bar mirrors the rule: name · status dot · tabs · `View changes` · Deploy.
- **New `ViewChangesModal.tsx`** (canvas dir): renders the session diff via ConfigDiff-style
  cards with per-card Discard + footer Discard-all / Deploy.
- **`PublicEndpointRow.tsx`:** restyle to single-row compact chips (`svc → host`), copy on
  hover. Same data (`PublicEndpoint[]`), same tooltip.
- **`detail/index.tsx` (`:716–746`):** drop label wiring; pass the session diff + revert
  callbacks and modal open-state to the shell/modal.

## Testing

- Shell renders draft (name input + Create) vs existing (read-only title + Deploy).
- Clean state: no `View changes`, Deploy disabled. Dirty: entry shows `N`, Deploy enabled.
- Only one status pill ever renders; no DRAFT pill.
- Modal lists changes, per-card Discard reverts that resource/volume/env row and updates `N`;
  discarding the last one empties the modal and hides the entry.
- Endpoints render one row; copy flashes; port tooltip intact.
- Existing collapse/persistence tests still pass.

## Open assumptions (confirm at plan time)

- Draft = name input + single `Create stack` (inferred — no server-side changes to view).
- `⋯` retains Delete + Discard-draft; Discard-all moves to modal only.
- `View changes` count `N = dirtyTotal`.
