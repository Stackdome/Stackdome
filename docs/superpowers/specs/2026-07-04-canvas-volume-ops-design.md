# Canvas Volume Operations — Design

**Date:** 2026-07-04
**Status:** Approved
**Branch:** `stack-canvas-editor`

## Goal

Move volume lifecycle management onto the stack canvas. Volumes are created from the canvas add button (born attached to a service), floating (unmounted) volumes attach via drag-onto-card or context menu, and docked volumes disconnect/delete via right-click context menus. The resource drawer's volume add/mount form is retired.

## Background

Today volumes are added from within the resource drawer. On canvas, mounted volumes render as chips docked on their owning resource card (`volumeChips()` in `graph-from-connections.ts`), while unmounted volumes float as `attachment` nodes. Persistence plumbing is complete: mounts live as `volume_mount` connections (`mountsToConnections()` / `connectionsToMounts()` in `connection-mapping.ts`), draft-sync diffs and syncs per-entity, and `volumesToDelete()` handles removal at deploy. This feature is purely UI-gesture work wired to existing edit-session state.

## Decisions

| Decision | Choice |
|---|---|
| Create flow | Dialog prompts for service + mount path upfront; volume is born attached |
| Attach gesture for floating volumes | Drag onto resource card (primary) + context-menu "Attach to service…" (fallback/accessibility) |
| Context menu scope | Volume chips, floating volume nodes, and resource cards |
| Cardinality | 1 volume : 1 resource enforced in UI (multiple volumes per resource allowed) |
| Drawer add flow | Retired; drawer keeps read-only mounted-volume rows that open the volume drawer |
| Volume name | Immutable after create (identity key for baseline alignment) |

## UX Flows

### Add volume
`AddResourcePopover` gains a "Volume" entry (HardDrive icon). Opens `AddVolumeDialog`: name (auto-suggested, editable), size, service picker (required), mount path (required). Confirm creates volume + mount in one session update → docked chip appears on the chosen card. Entry disabled with tooltip when the stack has no resources.

### Attach floating volume
1. **Drag:** dragging a floating volume node highlights valid resource cards (ring). Drop on a valid card opens `MountPathDialog` (mount path + read-only toggle); confirm attaches, cancel returns the node to its pre-drag position. Already-mounted volumes are not attachable elsewhere (disconnect first). Addon/secret/object-store nodes never highlight.
2. **Context menu:** right-click floating node → "Attach to service…" → same dialog with a service picker prepended.

### Disconnect
Right-click docked chip → "Disconnect volume". No confirmation (reversible draft edit). Chip disappears; volume reappears as a floating node near its former card. Volume entity and data survive; only the mount is removed.

### Delete volume
Destructive red item in both chip and floating-node menus. Confirmation dialog warns that data is destroyed on deploy (draft model — nothing is deleted immediately). Deleting a mounted volume cascades a disconnect in the same session update (one confirm covers both).

### Context menus
| Target | Items |
|---|---|
| Docked chip | Disconnect volume / Volume settings / Delete volume |
| Floating volume node | Attach to service… / Volume settings / Delete volume |
| Resource card | Open settings / Add volume… (dialog pre-picked to this service) / Delete service |

"Volume settings" opens the existing `VolumeDrawer`. Plain click on a floating volume node also opens it (fixes current gap: attachment nodes ignore clicks).

### Drawer cleanup
Resource drawer loses its volume add/mount form; keeps read-only mounted-volume rows that open `VolumeDrawer`.

## Architecture

All gestures mutate existing edit-session state (`session.updateVolumes` / `session.updateResources` in `use-stack-edit-session.ts`); draft-sync and `mountsToConnections` persist as today. **No new persistence code.**

### New components (under `frontend/src/pages/stacks/components/canvas/`)
- `AddVolumeDialog.tsx` — create form; emits one session update (append volume + append mount on chosen resource).
- `MountPathDialog.tsx` — mount path + read-only toggle; reused by drag-drop and "Attach to service…" (adds service picker in that mode).
- `CanvasContextMenu.tsx` — single menu component; three item-sets keyed by target kind (chip / floating volume / resource card). Radix menu primitive positioned at cursor via React Flow `onNodeContextMenu` and chip-level `onContextMenu` (`stopPropagation` so chip wins over card).

### Changed
- `AddResourcePopover.tsx` — Volume entry.
- `ResourceNode.tsx` / `VolumeChip` — context-menu wiring, drop-highlight ring.
- `AttachmentNode.tsx` — context menu, click → drawer.
- `StackCanvasTab.tsx` — drag-attach handlers: `onNodeDrag` → `getIntersectingNodes` (React Flow v12) → highlight valid target; `onNodeDragStop` → open `MountPathDialog` or restore position (position captured at `onNodeDragStart`).
- Resource drawer — strip add/mount form.

## Data flow & stack-diff coverage

Mounts live in `resource.volume_mounts`; volumes in `volumes[]`. `diffStack` (positional diff + `alignBaselineToDraft` by name) already covers most states:

| Gesture | Diff outcome | Status |
|---|---|---|
| Create attached | volume added (`dirtyVolumeIdx`) + resource dirty (`volume_mounts` → configuration bucket) | works today |
| Disconnect | resource dirty only; volume stays clean and floats | works today |
| Re-attach elsewhere | both affected resources dirty | works today |
| Delete unmounted volume | baseline-only volume reads dirty; `volumesToDelete()` at deploy | works today |
| Delete mounted volume | must cascade-remove the mount from the resource in the same update | **gap — build it** |
| Revert resource after its volume was deleted | `revertResource` restores baseline mounts that may point at a volume absent from the draft | **gap — guard it** |

Gap fixes:
1. Volume delete cascades disconnect in one session update.
2. Defensive filters: `volumeChips()` in `deriveGraph` and `buildDesiredState()` skip mounts whose volume name is missing from the draft. Unit-tested.

## Error handling & edge cases

- **Duplicate volume name** — validated in `AddVolumeDialog` against `volumes[]`; inline error, confirm disabled.
- **Mount path** — required, absolute (`/` prefix), no duplicate path on the same resource. Reuse existing zod schemas where present.
- **Drop on invalid target** — no highlight, drop is a plain reposition; no dialog.
- **Dialog cancel after drop** — node restored to pre-drag position.
- **Draft-sync failure** — existing autosave error path; gestures are optimistic session-state edits like all form edits.
- **Empty stack** — Volume add entry disabled with tooltip.
- **Pane right-click** — no menu in v1; `preventDefault` scoped to nodes/chips.

## Testing

- **Unit (vitest):** `stack-diff.test.ts` — cascade delete leaves no dangling mounts; revert-after-delete guard. `graph-from-connections` — chip skips missing volume; disconnected volume becomes floating node. `connection-mapping` — attach/detach round-trip.
- **Component:** `AddVolumeDialog` validation (duplicate name, bad path); context-menu item sets per target kind.
- **Manual (Playwright loop during implementation):** add → drag-attach → disconnect → re-attach → delete on `localhost:5173`, checking dirty badges after each step.

## Out of scope

- Volume metrics.
- Pane/background right-click menu.
- Multi-mount (one volume on several resources) — data model allows it, UI enforces 1:1 (access mode is hardcoded ReadWriteOnce; multi-resource mounts break at the K8s level unless pods co-schedule).
- Resource card menu beyond Open settings / Add volume… / Delete service.
