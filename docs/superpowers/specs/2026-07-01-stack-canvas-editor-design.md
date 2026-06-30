# Design Spec: Stack Canvas Editor

**Date:** 2026-07-01
**Branch / worktree:** `stack-canvas-editor` (`.claude/worktrees/stack-canvas-editor`, branched from `stack-creation-redesign` HEAD)
**Design reference:** `docs/superpowers/design-refs/2026-07-01-stack-creation-redesign-canvas-design-reference.md`
**Status:** Approved shape, pending spec review → `writing-plans`

---

## 1. Summary

Replace the Stack show/edit page's tabbed **Configuration form** with a **node-graph canvas editor** (React Flow / `@xyflow/react`), and redesign the **Deployments / Logs / Metrics** views to match the design bundle. The canvas is a *view layer* over state that already exists — it adds no new source of truth and requires no backend change for the core deliverable.

The create wizard (shipped in PR #129) is kept; only its hand-off target is re-pointed into the canvas (final slice).

## 2. Goals & non-goals

**Goals**
- Canvas renders an existing stack as nodes (resources + addons) and edges (connections).
- A right-side drawer configures a selected node via CONFIGURATION / DEPLOYMENT / ENVIRONMENT tabs.
- Preserve today's behaviour exactly: **dirty/change marking**, **env-var grouping**, atomic **Deploy** (PUT).
- Redesigned Deployments / Logs / Metrics views.
- Each slice independently demoable and Playwright-verified.

**Non-goals (this effort)** — see §13 for the deferred list.

## 3. Locked decisions

| # | Decision | Choice |
|---|---|---|
| D1 | Scope | Stack **show/edit page only**. Keep PR #129 wizard; re-point its hand-off (final slice). |
| D2 | Node layout | **Auto-layout, no persistence** (dagre/ELK each load). Drag is in-session, resets on reload. **Zero backend change.** |
| D3 | Rollout | **Feature flag.** Configuration tab toggles old-form ↔ new-canvas; remove old form once stable. |
| D4 | Connection edges | **Derived & read-only in v1.** Connection *authoring* stays in the drawer's ENVIRONMENT tab (existing addon-group UX). "Draw an edge to connect" is deferred. |
| D5 | Worker / Cron nodes | **Out of scope** — `WorkloadType` is unwired in api-server. Render only kinds that exist today (service + postgres/redis/minio addons). |
| D6 | Dark mode | **Adapt to theme** via `index.css` tokens (`colorMode="system"`); do not force `.dark`. |
| D7 | Volumes | **Chips on the owning node** (design option 1c) + full management in the drawer. Not separate nodes. |

## 4. Architecture

### 4.1 The canvas is a projection (Grokking-Simplicity framing)

We keep a hard line between **data**, **calculations** (pure, `data → data`), and **actions** (effectful). This is the spine of the whole feature.

```
                       ┌──────────────── DATA ────────────────┐
  useStackEditSession  │ draft (StackFormData) · baseline       │   ← single source of truth (unchanged)
                       └───────────────────────────────────────┘
                                     │ (pure)
        ┌────────────────────────────┼────────────────────────────┐
        ▼ CALCULATION                ▼ CALCULATION                 ▼ CALCULATION
  deriveGraph(draft)            layoutGraph(nodes,edges)      diffStack(draft, baseline)
  → { nodes, edges }            → positioned nodes            → dirty marks (EXISTING, reused)
        │                            │                             │
        └──────────────┬─────────────┘                            │
                       ▼ ACTION (React Flow render + interaction)  │
                 CanvasEditor ──── select node ───► ResourceDrawer ┘
                       │                                  │
                       │  edits call session.update(...)  │  (ACTION → mutates draft)
                       ▼                                  ▼
                 Deploy ► performSave ► convertFormStackToApiStack ► updateStack (PUT)   ← EXISTING, reused
```

- **`deriveGraph(draft)` is a pure calculation**: `StackFormData → { nodes, edges }`. No React Flow imports, no I/O, fully unit-testable. Nodes = resources + linked addons; edges = `spec.connections`; volumes folded into their owning node's data as chips.
- **`layoutGraph` is a pure calculation**: graph → graph with `position` set (dagre/ELK). Deterministic; testable without rendering.
- **The React Flow layer is the only action shell** that owns interaction state (selection, viewport, in-session drag positions).
- Editing a node **calls the existing session updater** — the same path the form uses. The drawer never owns stack state; it is a controlled view.

### 4.2 Deep modules (PoSD)

Each unit has a narrow interface hiding substantial implementation — *what it does* is clear without reading internals:

| Module | Interface (what callers see) | Hides |
|---|---|---|
| `deriveGraph` | `(draft) → {nodes, edges}` | node/edge shape, kind→glyph/label mapping, volume folding |
| `layoutGraph` | `(nodes, edges) → nodes` | dagre/ELK config, rank direction, spacing |
| `CanvasEditor` | `(stack session, onSelect)` | React Flow wiring, nodeTypes, memoization, controls |
| `ResourceDrawer` | `(selectedNode, session)` | tab layout; **delegates bodies to existing sub-tab components** |
| `useStackEditSession` | (unchanged) | draft/baseline/diff/save — already a deep module |

PoSD "define errors out of existence": a selected node id that no longer exists in `draft` (e.g. removed) resolves to *no drawer*, not an error path.

## 5. Reuse map — why this is smaller than it looks

| Need | Reuse (existing) | New |
|---|---|---|
| Edit state, dirty diff, revert | `useStackEditSession`, `lib/stack-diff.ts` | — |
| Save (PUT) | `performSave`, `convertFormStackToApiStack`, `updateStack` | — |
| Drawer tab **bodies** | `stack-resource-configuration-tab`, `-deployment-tab`, `-environment-tab` (incl. **env-var grouping**) | drawer shell + tab bar around them |
| Connections → display | `lib/connection-mapping.ts` (`connectionsToEnvRows`) read for edges | `deriveGraph` edge builder |
| Volumes editing | `stack-volumes-form`, `stack-volume-item` | node volume-chip renderer |
| Dirty/deploy bar | `useDeployLifecycle`, `sticky-action-bar` logic | header restyle |
| Add-resource picker | block picker from PR #129 wizard composer (verify reusability) | "+ Add resource" popover wrapper |
| Ops views | `deployments-tab`, `stack-logs-tab`, `stack-metrics-tab` | re-skin into mode views |

## 6. New components

| Component | Responsibility | Key interface |
|---|---|---|
| `lib/canvas/derive-graph.ts` | Pure: `draft → {nodes, edges}` | `deriveGraph(draft): CanvasGraph` |
| `lib/canvas/layout-graph.ts` | Pure: auto-layout positions | `layoutGraph(graph): CanvasGraph` |
| `canvas/CanvasEditor.tsx` | React Flow shell: background, controls, nodeTypes, selection, in-session drag | props: `session`, `selectedId`, `onSelect` |
| `canvas/nodes/ResourceNode.tsx` | `memo`'d card: status dot, glyph, name, kind badge, mono summary, volume chips, dirty mark | `NodeProps<ResourceNodeData>` |
| `canvas/ResourceDrawer.tsx` | Slide-in drawer; tab bar + dirty dots; footer (View logs / Remove); wraps existing sub-tabs | props: `node`, `session`, `openTab` |
| `canvas/AddResourcePopover.tsx` | "+ Add resource" → block picker → `session.addResource` | props: `session`, anchor |
| `canvas/CanvasHeader.tsx` | name, status pill, mode tabs, "N unsaved changes", Deploy | props: `session`, `mode`, `onMode` |
| `canvas/StackCanvasTab.tsx` | Flag-gated entry mounted in Configuration tab | props: `session` |
| `views/{Deployments,Logs,Metrics}View.tsx` | Re-skinned mode views | wrap existing tab components |

## 7. Design tokens (add to `frontend/src/index.css`)

The bundle relies on token families absent from live CSS (per design-ref §2). Add them **theme-aware** (both `:root` light and `.dark`), mapping to existing brand tokens where they match — **no raw hex/px in components** (brand rule):

- **Amber → brand aliases:** the canvas/components use `--amber*`; alias to existing `--brand*` so one brand color holds. (Or author components directly against `--brand*`.)
- **Foreground hierarchy:** add `--fg2`, `--fg3` (the drawer/cards use a 3-step muted ramp; live has only `--foreground` + `--fg-muted`).
- **Radius scale:** `--r-0..--r-4` (4–8px); live has single `--radius`.
- **Spacing scale:** `--s-1..--s-10`.
- **Type scale + tracking:** `--t-*`, `--tr-*`.
- **Misc:** `--amber-soft` (12%), `--shadow-amber` focus ring, `--dur-*`/`--ease` motion.

Decision deferred to slice 0: alias `--amber*`→`--brand*` **vs** author components against `--brand*` directly. Lean toward authoring against existing `--brand*` and adding only the genuinely-new families (fg ramp, scales, motion) to keep the token set lean.

## 8. Feature flag

`VITE_STACK_CANVAS` (build-time) **or** a `localStorage` override for in-app toggling during iteration. Configuration tab: flag off → current form (untouched fallback); flag on → `StackCanvasTab`. Old form deleted once the canvas is stable.

## 9. Node/edge derivation contract

```
nodes:  for r in draft.resources          → { id:r.id, type:'resource', data: { kind, name, summary, status, volumes, dirty } }
        for a in draft.linkedAddonIds      → { id:a,    type:'resource', data: { kind: addonKind, ... } }
edges:  for c in draft.spec.connections    → { id, source:c.from.id, target:c.to.id }   (read-only)
```
- `dirty` per node comes from `diffStack(draft, baseline)` (existing `dirtyResourceIdx` sets) → drives the New/Edited/Removed mark.
- Save is **unchanged**: the canvas only mutates `draft`; `convertFormStackToApiStack` already emits the full `spec.connections` set for the atomic PUT (see memory: *stack PUT replaces all connections*; strip read-only `outputs` before PUT — already handled).

## 10. Design principles applied

**Grokking Simplicity**
- Separate actions / calculations / data (§4.1). Derivation + layout + diff are calculations; only React Flow + session updates are actions.
- Stratified design: `deriveGraph`/`layoutGraph` (graph layer) sit below `CanvasEditor` (render layer) below the page. Each layer speaks only to the one below.
- Copy-on-write at the data boundary: node `data` updates are immutable (React Flow requirement *and* GS discipline) — `updateNodeData`/new objects, never mutate.

**A Philosophy of Software Design**
- Deep modules with narrow interfaces (§4.2); the drawer hides three large existing tab bodies behind one prop surface.
- Pull complexity downward: layout/derivation complexity lives in pure libs, not in components.
- Define errors out of existence: missing/removed selection ⇒ no drawer, not an error.
- Comments explain *why* (e.g. why edges are derived not authored) — not restating code.

## 11. Vertical slices (tracer bullets)

Detailed steps go to the implementation plan; this is the ordering. Each ends in a Playwright-verified demo on `localhost:5173`.

| # | Slice | Demo | Backend |
|---|---|---|---|
| 0 | **Foundations** — `pnpm install` in worktree; add `@xyflow/react`; add tokens; feature flag; blank grid canvas + pan/zoom controls behind flag | flag on → dotted-grid canvas, controls work | none |
| 1 | **Read-only graph** — `deriveGraph` + `layoutGraph` (unit-tested); `ResourceNode` card; drag/fit/show-connections | open a real stack → node graph matches topology | none |
| 2 | **Node drawer** — `ResourceDrawer` wrapping existing sub-tabs; per-tab dirty dots; edits → draft | click node, edit image/port/env; env grouping intact | none |
| 3 | **Header + dirty + deploy** — `CanvasHeader`; node New/Edited/Removed marks; Deploy wiring | edit → unsaved count → Deploy → PUT | none |
| 4 | **Add / remove** — `AddResourcePopover`; remove from drawer; volume chips | add redis, wire env, deploy | none |
| 5 | **Ops views** — re-skin Deployments / Logs / Metrics; mode tab bar | switch modes, redesigned views | none |
| 6 | **Wizard hand-off + polish** — re-point PR #129 composer CTA; perf (memo nodes, narrow `useStore` selectors, `onlyRenderVisibleElements`); empty/edge states | create flow lands in canvas | none |

## 12. Testing & verification

- **Unit (Vitest):** `deriveGraph` and `layoutGraph` are pure → table-driven tests (resources+connections+volumes → expected nodes/edges; layout determinism). Existing `stack-diff` tests still cover dirty/save.
- **Component:** drawer dirty marks; add/remove mutates draft correctly.
- **Playwright per slice:** drive `http://localhost:5173` against the existing running backend. Login `admin@stackdome.io` / `welcome@123`. Each slice's "Demo" column is the acceptance check.
- Save path untouched → existing PUT coverage holds.

## 13. Deferred / out of scope

Node-position **persistence** · **draw-edges-on-canvas** connection authoring · **worker/cron** nodes (backend `WorkloadType` wiring) · multiplayer/presence · re-doing the create wizard's earlier phases (kept as-is) · list-view toggle at scale.

## 14. Open risks

- **Block-picker reuse (slice 4):** confirm the PR #129 composer picker is extractable for the "+ Add resource" popover; if coupled, build a thin shared picker.
- **Token strategy (slice 0):** alias vs author-against-`--brand`; decide before building nodes to avoid churn.
- **Unattached volumes:** rare; handled in drawer, no standalone node. Confirm no stack relies on them visually.
- **Layout stability:** auto-layout must be deterministic so drag doesn't fight re-layout on unrelated re-renders (only re-layout on graph topology change, not on every draft edit).

---

## Appendix A — React Flow best practices (context7 + deepwiki, 2026-07-01)

1. **Declare `nodeTypes`/`edgeTypes` at module scope** (or `useMemo`) — new object identity re-renders all nodes.
2. **`React.memo` every custom node**; memoize handlers with `useCallback`.
3. **Controlled state** via `useNodesState`/`useEdgesState` + `onNodesChange`/`onEdgesChange`.
4. **Immutable data updates** — `updateNodeData` / new node objects; never mutate in place.
5. **Narrow store reads** — `useStore(selector, shallow)`; never select the whole `nodes` array inside a node (one bad selector re-renders every node per drag tick).
6. **Large graphs** — `onlyRenderVisibleElements`, `snapToGrid`/`snapGrid`; keep heavy node content in memoized subcomponents (`useDeferredValue` for non-urgent).
7. **CSS for selected/hover**, not data writes.
8. **Auto-layout** — run dagre/ELK over `nodes`+`edges`, `setNodes` with new positions, then `fitView`. Re-layout only on topology change.
9. **`Handle`** for connection points (hidden/styled since edges are derived in v1); **`NodeToolbar`** for non-scaling per-node actions.
10. **`colorMode="system"`** + `<Background/> <Controls/> <MiniMap/>` from `@xyflow/react`; import `@xyflow/react/dist/style.css`.
11. **`useUpdateNodeInternals`** when node dimensions/handles change programmatically.

## Appendix B — Token mapping

See design-ref §2 for the full bundle→live token mapping and drift notes (amber→brand, fg ramp, radius/spacing scales, dark-first assumption).
