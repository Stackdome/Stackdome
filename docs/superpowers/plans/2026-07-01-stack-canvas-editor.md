# Stack Canvas Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the Stack show/edit page's Configuration form with a node-graph canvas editor (React Flow) that is a pure view over the existing edit session, behind a feature flag.

**Architecture:** The canvas reads `useStackEditSession().draft` and writes back through `updateResources`/`updateVolumes`/`setLinkedAddonIds` — the same path the form uses. Graph derivation and layout are pure calculations (`data → data`, unit-tested without React Flow); only the React Flow render layer and session updates are actions. Save is the unchanged atomic PUT (`convertFormStackToApiStack` → `updateStack`).

**Tech Stack:** React 19, Vite, Tailwind v4, `@xyflow/react` (new), `dagre` (new, layout), Vitest, Playwright (verification). Backend unchanged.

## Global Constraints

- **No raw hex / no raw px in components.** Use `index.css` `var(--…)` tokens only. Fonts: Geist (sans) / Geist Mono (mono) only. Single brand color (`--brand`). (memory: brand design system)
- **No magic strings.** Use defined constants/enums (e.g. `EditSessionTab`, `TopologyNodeRef.type` values), not raw literals. Define a constant if none exists.
- **Source of truth = `useStackEditSession` draft.** The canvas never owns stack state; derivation/layout/diff are pure. (memory: GS + PoSD principles)
- **Save contract unchanged:** the full `spec.connections` set is emitted by `convertFormStackToApiStack`; strip read-only `outputs` before PUT (already handled). (memory: stack PUT replaces all connections)
- **React Flow discipline:** `nodeTypes`/`edgeTypes` at module scope; every custom node `React.memo`'d; immutable node-data updates; narrow `useStore(selector, shallow)`; `colorMode="system"`.
- **Verification:** every slice ends with a Playwright check on `http://localhost:5173` against the already-running backend. Login `admin@stackdome.io` / `welcome@123`. (memory: Playwright in-session loop)
- **Frontend test command:** `pnpm --prefix frontend test:run` (Vitest, run-once). Dev server: `pnpm --prefix frontend dev`.
- **Worktree:** all work in `.claude/worktrees/stack-canvas-editor` on branch `stack-canvas-editor`.
- **No mention of comparable third-party PaaS products** in code, comments, or commits.

---

## File Structure

New files (all under `frontend/src/pages/stacks/components/canvas/` unless noted):

| Path | Responsibility |
|---|---|
| `lib/canvas/derive-graph.ts` | **Pure.** `deriveGraph(input) → { nodes, edges }`. Resources+addons→nodes, env-var refs+links→edges, volumes folded into node data. |
| `lib/canvas/derive-graph.ts` types | `CanvasGraph`, `ResourceNodeData`, `CanvasNode`, `CanvasEdge`, `DeriveGraphInput`, `NODE_KIND` const. |
| `lib/canvas/layout-graph.ts` | **Pure.** `layoutGraph(graph) → graph` with positions (dagre). |
| `lib/canvas/tests/derive-graph.test.ts` | Vitest table tests. |
| `lib/canvas/tests/layout-graph.test.ts` | Vitest layout determinism tests. |
| `canvas/StackCanvasTab.tsx` | Flag-gated entry mounted in Configuration `TabsContent`. Owns React Flow provider + selection state. |
| `canvas/CanvasEditor.tsx` | React Flow shell: Background, Controls, nodeTypes, edges, in-session drag, show-connections toggle. |
| `canvas/nodes/ResourceNode.tsx` | `memo`'d node card. |
| `canvas/nodes/node-glyph.tsx` | kind→Lucide glyph mapping (reuse `BlockGlyph` if shape matches). |
| `canvas/ResourceDrawer.tsx` | Slide-in drawer; tab bar + dirty dots; footer; wraps existing sub-tabs via `useResourceTabProps`. |
| `canvas/AddResourcePopover.tsx` | "+ Add resource" → `BlockPicker` → `session.updateResources`. |
| `canvas/CanvasHeader.tsx` | name, status pill, mode tabs, "N unsaved changes", Deploy. |
| `shared/hooks/use-resource-tab-props.ts` | **Extracted** per-resource prop assembly + patch handlers (from `stack-resource-item.tsx`), reused by accordion item AND drawer. |
| `views/DeploymentsView.tsx` / `LogsView.tsx` / `MetricsView.tsx` | Re-skinned mode views wrapping existing tab components. |
| `lib/feature-flags.ts` (`frontend/src/lib/`) | `isCanvasEnabled()` reading `import.meta.env.VITE_STACK_CANVAS` + `localStorage`. |

Modified:
- `frontend/src/index.css` — add token families (`--fg2/3`, `--r-*`, `--s-*`, type scale, motion) to `:root` (raw) + `.dark` (overrides) + `@theme inline`.
- `frontend/src/pages/stacks/components/detail/index.tsx:683-806` — flag-gate the Configuration `TabsContent` between current panels and `StackCanvasTab`.
- `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` — consume the extracted `useResourceTabProps` (slice 2, DRY).
- `frontend/package.json` — add `@xyflow/react`, `dagre`, `@types/dagre`.

---

## Slices 0–2 are fully detailed below. Slices 3–6 are task-level outlines, expanded into bite-sized steps just-in-time at execution (per YAGNI — they depend on 0–2 outcomes and the spec already locked their shape).

---

# SLICE 0 — Foundations

**Deliverable:** Flag on → Configuration tab shows a dotted-grid canvas with working pan/zoom/fit controls; flag off → current form untouched. New design tokens available.

### Task 0.1: Worktree deps + baseline

**Files:** `frontend/package.json`

- [ ] **Step 1: Install + baseline test**

Run:
```bash
pnpm --prefix frontend install
pnpm --prefix frontend test:run 2>&1 | tail -20
```
Expected: install completes; existing tests PASS (clean baseline). If any fail, STOP and report.

- [ ] **Step 2: Add canvas deps**

Run:
```bash
pnpm --prefix frontend add @xyflow/react dagre
pnpm --prefix frontend add -D @types/dagre
```
Expected: added to `package.json` dependencies.

- [ ] **Step 3: Commit**
```bash
git add frontend/package.json frontend/pnpm-lock.yaml
git commit -m "build(stacks): add @xyflow/react + dagre for canvas editor"
```

### Task 0.2: Design tokens

**Files:** Modify `frontend/src/index.css` (`:root` ~6-85, `.dark` ~87-158, `@theme inline` ~160-240)

**Interfaces — Produces:** CSS custom properties consumed by all canvas components: `--fg2`, `--fg3`, `--r-1..--r-4`, `--s-1..--s-8`, `--amber-soft`, `--shadow-amber`, `--dur-2`, `--ease`. (Components author against existing `--brand*`; we add only genuinely-new families per spec lean.)

- [ ] **Step 1: Add raw tokens to `:root`**

Insert before the closing `}` of `:root` (after `--radius`):
```css
  /* canvas editor — foreground ramp (light) */
  --fg2: oklch(0.45 0.01 258);
  --fg3: oklch(0.58 0.01 258);
  /* radius scale */
  --r-1: 2px; --r-2: 4px; --r-3: 6px; --r-4: 8px;
  /* spacing scale */
  --s-1: 4px; --s-2: 8px; --s-3: 12px; --s-4: 16px; --s-5: 24px; --s-6: 32px; --s-7: 48px; --s-8: 64px;
  /* brand soft + focus ring */
  --amber-soft: color-mix(in oklch, var(--brand) 12%, transparent);
  --shadow-amber: 0 0 0 3px color-mix(in oklch, var(--brand) 20%, transparent);
  /* motion */
  --dur-2: 180ms; --ease: cubic-bezier(0.2, 0, 0, 1);
```

- [ ] **Step 2: Add dark overrides to `.dark`**

Insert before the closing `}` of `.dark`:
```css
  --fg2: oklch(0.83 0.02 257);
  --fg3: oklch(0.70 0.02 257);
```
(radius/spacing/motion are theme-agnostic — no override needed.)

- [ ] **Step 3: Expose in `@theme inline`**

Inside `@theme inline { … }` add Tailwind mappings:
```css
  --color-fg2: var(--fg2);
  --color-fg3: var(--fg3);
```

- [ ] **Step 4: Verify build compiles**

Run: `pnpm --prefix frontend build 2>&1 | tail -5`
Expected: no CSS errors. (Note: `mage build` skips tsc; `pnpm build` runs tsc — pre-existing unrelated TS errors are acceptable, CSS must compile.)

- [ ] **Step 5: Commit**
```bash
git add frontend/src/index.css
git commit -m "style(stacks): add canvas editor design tokens (fg ramp, radius/spacing scale, motion)"
```

### Task 0.3: Feature flag

**Files:** Create `frontend/src/lib/feature-flags.ts`; Test `frontend/src/lib/tests/feature-flags.test.ts`

**Interfaces — Produces:** `isCanvasEnabled(): boolean`

- [ ] **Step 1: Failing test**
```ts
// frontend/src/lib/tests/feature-flags.test.ts
import { describe, it, expect, beforeEach } from "vitest";
import { isCanvasEnabled } from "../feature-flags";

describe("isCanvasEnabled", () => {
  beforeEach(() => localStorage.clear());
  it("is false by default", () => {
    expect(isCanvasEnabled()).toBe(false);
  });
  it("is true when localStorage override set", () => {
    localStorage.setItem("stackCanvas", "1");
    expect(isCanvasEnabled()).toBe(true);
  });
});
```

- [ ] **Step 2: Run — expect FAIL** (`Cannot find module '../feature-flags'`)

Run: `pnpm --prefix frontend test:run feature-flags`

- [ ] **Step 3: Implement**
```ts
// frontend/src/lib/feature-flags.ts
const CANVAS_LS_KEY = "stackCanvas";

/** Canvas editor is opt-in: build-time env OR per-browser localStorage override. */
export function isCanvasEnabled(): boolean {
  if (import.meta.env.VITE_STACK_CANVAS === "true") return true;
  try {
    return localStorage.getItem(CANVAS_LS_KEY) === "1";
  } catch {
    return false; // SSR/no-storage → treat as off (PoSD: define error out of existence)
  }
}
```

- [ ] **Step 4: Run — expect PASS**

Run: `pnpm --prefix frontend test:run feature-flags`

- [ ] **Step 5: Commit**
```bash
git add frontend/src/lib/feature-flags.ts frontend/src/lib/tests/feature-flags.test.ts
git commit -m "feat(stacks): add canvas feature flag (env + localStorage override)"
```

### Task 0.4: Blank canvas shell behind flag

**Files:** Create `canvas/StackCanvasTab.tsx`, `canvas/CanvasEditor.tsx`; Modify `detail/index.tsx:683-806`

**Interfaces:**
- Consumes: `isCanvasEnabled()`; `useStackEditSession` return (`session`).
- Produces: `<StackCanvasTab session={session} />`; `<CanvasEditor nodes={[]} edges={[]} />` (extended in slice 1).

- [ ] **Step 1: CanvasEditor shell**
```tsx
// canvas/CanvasEditor.tsx
import { ReactFlow, Background, Controls, type Node, type Edge } from "@xyflow/react";
import "@xyflow/react/dist/style.css";

interface CanvasEditorProps {
  nodes: Node[];
  edges: Edge[];
}

export function CanvasEditor({ nodes, edges }: CanvasEditorProps) {
  return (
    <div className="h-full w-full" data-testid="stack-canvas">
      <ReactFlow nodes={nodes} edges={edges} fitView colorMode="system" proOptions={{ hideAttribution: true }}>
        <Background />
        <Controls />
      </ReactFlow>
    </div>
  );
}
```

- [ ] **Step 2: StackCanvasTab**
```tsx
// canvas/StackCanvasTab.tsx
import { ReactFlowProvider } from "@xyflow/react";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import { CanvasEditor } from "./CanvasEditor";

interface StackCanvasTabProps {
  session: UseStackEditSession;
}

export function StackCanvasTab(_props: StackCanvasTabProps) {
  // slice 1 derives nodes/edges from session; shell renders empty for now.
  return (
    <div className="h-[calc(100vh-12rem)]">
      <ReactFlowProvider>
        <CanvasEditor nodes={[]} edges={[]} />
      </ReactFlowProvider>
    </div>
  );
}
```

- [ ] **Step 3: Flag-gate in detail/index.tsx**

At the top of the Configuration `<TabsContent value="configuration">` (`:683`), wrap the existing three `<Panel>`s so the canvas replaces them when enabled:
```tsx
import { isCanvasEnabled } from "@/lib/feature-flags";
import { StackCanvasTab } from "../canvas/StackCanvasTab";
// ...
<TabsContent value="configuration" /* existing props */>
  {isCanvasEnabled() ? (
    <StackCanvasTab session={session} />
  ) : (
    <>
      {/* existing Stack Resources / Volumes / Addons panels unchanged */}
    </>
  )}
</TabsContent>
```

- [ ] **Step 4: Manual + Playwright verify**

Run dev server (background): `pnpm --prefix frontend dev`
Playwright: navigate `http://localhost:5173`, login `admin@stackdome.io` / `welcome@123`, open any stack, run `localStorage.setItem("stackCanvas","1")`, reload, open Configuration tab.
Expected: dotted-grid canvas with pan/zoom/fit controls; flag off → original form. Screenshot.

- [ ] **Step 5: Commit**
```bash
git add frontend/src/pages/stacks/components/canvas/ frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(stacks): flag-gated blank canvas shell in Configuration tab"
```

---

# SLICE 1 — Read-only graph

**Deliverable:** Opening a real stack renders its resources/addons as node cards wired by edges (derived from env-var references + addon links); auto-laid-out; draggable in-session; fit + show-connections controls.

### Task 1.1: `deriveGraph` (pure calculation) — TDD

**Files:** Create `lib/canvas/derive-graph.ts`, `lib/canvas/tests/derive-graph.test.ts`

**Interfaces — Produces:**
```ts
export const NODE_KIND = { service: "service", addon: "addon" } as const;
export type NodeKind = (typeof NODE_KIND)[keyof typeof NODE_KIND];
export interface VolumeChip { name: string; mountPath?: string; sizeGb?: number; }
export interface ResourceNodeData {
  kind: NodeKind; name: string; summary: string;
  status?: string; volumes: VolumeChip[]; resourceIdx?: number;
}
export interface CanvasNode { id: string; type: "resource"; data: ResourceNodeData; position: { x: number; y: number }; }
export interface CanvasEdge { id: string; source: string; target: string; }
export interface CanvasGraph { nodes: CanvasNode[]; edges: CanvasEdge[]; }
export interface DeriveGraphInput {
  resources: Partial<FormStackResourceData>[];
  linkedAddonIds: ReadonlySet<string>;
  addonNameById: ReadonlyMap<string, string>;
}
export function deriveGraph(input: DeriveGraphInput): CanvasGraph;
```
- Consumes: `FormStackResourceData`, `FormEnvVarData` (from `schemas/form-schema`).

**Derivation rules** (the implementer must satisfy these tests):
- One `service` node per resource (`id = "resource:" + name`), `resourceIdx` = array index.
- One `addon` node per id in `linkedAddonIds` (`id = "addon:" + addonId`, name via `addonNameById`).
- Edge for each env var with `from:"addon"` → `source=resource node, target="addon:"+addonId`.
- Edge for each env var with `from:"resource"` → `source=resource node, target="resource:"+resourceName`.
- Edges deduped by `id = source+"->"+target`.
- `volumes` chip per `resource.volume_mounts` entry (name + mountPath).
- `position` = `{x:0,y:0}` (layout assigns real positions later).

- [ ] **Step 1: Failing tests**
```ts
// lib/canvas/tests/derive-graph.test.ts
import { describe, it, expect } from "vitest";
import { deriveGraph, NODE_KIND } from "../derive-graph";

const baseInput = { resources: [], linkedAddonIds: new Set<string>(), addonNameById: new Map<string,string>() };

describe("deriveGraph", () => {
  it("makes one service node per resource", () => {
    const g = deriveGraph({ ...baseInput, resources: [{ name: "web" }, { name: "api" }] });
    expect(g.nodes.map(n => n.id)).toEqual(["resource:web", "resource:api"]);
    expect(g.nodes[0].data.kind).toBe(NODE_KIND.service);
  });

  it("adds addon nodes from linkedAddonIds with names", () => {
    const g = deriveGraph({ ...baseInput, linkedAddonIds: new Set(["a1"]), addonNameById: new Map([["a1","db"]]) });
    const addon = g.nodes.find(n => n.id === "addon:a1");
    expect(addon?.data).toMatchObject({ kind: NODE_KIND.addon, name: "db" });
  });

  it("derives addon edge from env var ref", () => {
    const g = deriveGraph({
      ...baseInput,
      resources: [{ name: "web", execution_config: { environment_variables: [{ from: "addon", name: "DB", addonId: "a1" }] } }],
      linkedAddonIds: new Set(["a1"]), addonNameById: new Map([["a1","db"]]),
    });
    expect(g.edges).toContainEqual(expect.objectContaining({ source: "resource:web", target: "addon:a1" }));
  });

  it("derives resource→resource edge and dedupes", () => {
    const g = deriveGraph({
      ...baseInput,
      resources: [{ name: "web", execution_config: { environment_variables: [
        { from: "resource", name: "U1", resourceName: "api", output: "URL" },
        { from: "resource", name: "U2", resourceName: "api", output: "URL" },
      ] } }, { name: "api" }],
    });
    expect(g.edges.filter(e => e.source === "resource:web" && e.target === "resource:api")).toHaveLength(1);
  });

  it("folds volume_mounts into node volume chips", () => {
    const g = deriveGraph({ ...baseInput, resources: [{ name: "db", volume_mounts: [{ name: "data", mount_path: "/var/lib" }] }] });
    expect(g.nodes[0].data.volumes).toEqual([{ name: "data", mountPath: "/var/lib" }]);
  });
});
```

- [ ] **Step 2: Run — expect FAIL** (`pnpm --prefix frontend test:run derive-graph`)

- [ ] **Step 3: Implement `deriveGraph`** to satisfy the rules above (pure; no React/React Flow imports). Use `NODE_KIND` constants, not literals.

- [ ] **Step 4: Run — expect PASS**

- [ ] **Step 5: Commit** (`feat(stacks): pure deriveGraph — resources/addons → canvas nodes/edges`)

### Task 1.2: `layoutGraph` (pure) — TDD

**Files:** Create `lib/canvas/layout-graph.ts`, `lib/canvas/tests/layout-graph.test.ts`

**Interfaces — Produces:** `layoutGraph(graph: CanvasGraph, opts?: { direction?: "LR" | "TB" }): CanvasGraph` (returns new graph with non-`{0,0}` positions; deterministic for same input).

- [ ] **Step 1: Failing test** — assert all nodes get distinct positions and output is deterministic across two calls; assert no input mutation.
```ts
import { describe, it, expect } from "vitest";
import { layoutGraph } from "../layout-graph";
const graph = { nodes: [
  { id: "resource:web", type: "resource" as const, data: { kind: "service" as const, name: "web", summary: "", volumes: [] }, position: { x: 0, y: 0 } },
  { id: "addon:a1", type: "resource" as const, data: { kind: "addon" as const, name: "db", summary: "", volumes: [] }, position: { x: 0, y: 0 } },
], edges: [{ id: "resource:web->addon:a1", source: "resource:web", target: "addon:a1" }] };

describe("layoutGraph", () => {
  it("assigns distinct deterministic positions without mutating input", () => {
    const a = layoutGraph(graph); const b = layoutGraph(graph);
    expect(a.nodes[0].position).not.toEqual({ x: 0, y: 0 });
    expect(a.nodes.map(n => n.position)).toEqual(b.nodes.map(n => n.position));
    expect(graph.nodes[0].position).toEqual({ x: 0, y: 0 }); // input untouched
  });
});
```

- [ ] **Step 2: Run — expect FAIL**
- [ ] **Step 3: Implement with `dagre`** (`LR` default, fixed node size ~216×88, fixed ranksep/nodesep; return new node objects — copy-on-write).
- [ ] **Step 4: Run — expect PASS**
- [ ] **Step 5: Commit** (`feat(stacks): pure layoutGraph (dagre auto-layout)`)

### Task 1.3: `ResourceNode` card

**Files:** Create `canvas/nodes/ResourceNode.tsx`, `canvas/nodes/node-glyph.tsx`

**Interfaces:**
- Consumes: `ResourceNodeData` (Task 1.1); React Flow `NodeProps`.
- Produces: default-exported `memo` component; `nodeTypes = { resource: ResourceNode }` declared at module scope in `CanvasEditor`.

- [ ] **Step 1: node-glyph** — map `NODE_KIND.service`→`Box`/`Globe`, `NODE_KIND.addon`→`Database` Lucide icons (reuse `BlockGlyph` if its props fit; else thin wrapper). No raw hex — `text-[color:var(--brand)]` etc.
- [ ] **Step 2: ResourceNode** — `memo` card (216px): status dot, glyph, name, kind badge (mono uppercase), summary (mono), volume chips with fill bar; hidden `Handle`s (source+target) since edges are derived. Selected/hover via CSS only. Tokens only.
- [ ] **Step 3: Storybook-free visual check** — rendered in slice via dev server (no isolated test). Type-check: `pnpm --prefix frontend exec tsc -b --noEmit` on the file path region passes.
- [ ] **Step 4: Commit** (`feat(stacks): ResourceNode canvas card (memoized)`)

### Task 1.4: Wire graph into CanvasEditor + StackCanvasTab

**Files:** Modify `canvas/CanvasEditor.tsx`, `canvas/StackCanvasTab.tsx`

**Interfaces:**
- Consumes: `deriveGraph`, `layoutGraph`, `ResourceNode`, `session` (`draft.resources`, `linkedAddonIds`), addon name map (from `detail/index.tsx` — pass as prop `addonNameById`).
- Produces: `CanvasEditor` props extended: `{ session, addonNameById, showConnections, onToggleConnections }`.

- [ ] **Step 1: Derive + layout in StackCanvasTab**
```tsx
const graph = useMemo(
  () => layoutGraph(deriveGraph({ resources: session.draft.resources, linkedAddonIds: session.linkedAddonIds, addonNameById })),
  [session.draft.resources, session.linkedAddonIds, addonNameById],
);
```
Feed `graph.nodes`/`graph.edges` to `CanvasEditor` via `useNodesState`/`useEdgesState` (controlled, for in-session drag). Sync external graph changes with an effect that resets nodes when topology id-set changes (not on every drag).

- [ ] **Step 2: nodeTypes at module scope**; `<ReactFlow nodeTypes={nodeTypes} onNodesChange={onNodesChange} …>`.
- [ ] **Step 3: Show-connections toggle** — control that hides/shows edges (CSS/`edges=[]`), wired to a `useState` in StackCanvasTab.
- [ ] **Step 4: Pass `addonNameById` from detail/index.tsx** into `StackCanvasTab` (it already builds `addonNameById` for the env tab).
- [ ] **Step 5: Playwright verify** — open `tooljet-addon`-like stack with the flag on; assert N node cards = N resources+addons, edges present, drag moves a node, fit button frames graph. Screenshot vs design topology.
- [ ] **Step 6: Commit** (`feat(stacks): render stack as node graph (read-only)`)

---

# SLICE 2 — Node drawer

**Deliverable:** Clicking a node opens a slide-in drawer with CONFIGURATION / DEPLOYMENT / ENVIRONMENT tabs (reusing existing sub-tab bodies), per-tab dirty dots, edits flowing into `draft`; env-var grouping intact.

### Task 2.1: Extract `useResourceTabProps` (DRY deep module)

**Files:** Create `shared/hooks/use-resource-tab-props.ts`; Modify `shared/stack-resource-item.tsx` to consume it; Test `shared/hooks/tests/use-resource-tab-props.test.tsx`

**Why:** `stack-resource-item.tsx` already assembles the rich prop sets + patch handlers the three sub-tabs need. Extract that into one hook so the drawer reuses it verbatim (no duplicated wiring). PoSD deep module: narrow `(session, index, context) → { configurationProps, deploymentProps, environmentProps }`.

**Interfaces — Produces:**
```ts
export function useResourceTabProps(args: {
  session: UseStackEditSession;
  index: number;
  context: { secrets: UseSecretsReturn; addons: PostgresAddon[]; addonNameById: Map<string,string>;
             allResources: { name: string; index: number }[]; resourceOptions: { name: string; outputs: string[] }[];
             errors: Record<string, string | undefined>; };
}): {
  configurationProps: StackResourceConfigurationTabProps;
  deploymentProps: StackResourceDeploymentTabProps;
  environmentProps: StackResourceEnvironmentTabProps;
};
```

- [ ] **Step 1:** Read `stack-resource-item.tsx` fully; identify the prop-assembly + `onPatchResource`/`onPatchInitSpec`/`onPatchExecCommandArgs`/`onChangeEnvVars`/`onDiscard*` closures.
- [ ] **Step 2: Failing test** — mount a probe that calls the hook with a fake session + one resource; assert `configurationProps.draft.name` matches the resource and `onPatchResource({image_spec:{image:"x"}})` triggers `session.updateResources`.
- [ ] **Step 3: Implement** the hook by moving (not copying) the assembly out of `stack-resource-item.tsx`.
- [ ] **Step 4:** Refactor `stack-resource-item.tsx` to consume the hook; run existing stacks tests + type-check — must stay green (behavior-preserving).
- [ ] **Step 5: Commit** (`refactor(stacks): extract useResourceTabProps for reuse`)

### Task 2.2: `ResourceDrawer`

**Files:** Create `canvas/ResourceDrawer.tsx`

**Interfaces:**
- Consumes: `useResourceTabProps`; the three sub-tab components; `session.dirty` (`dirtyTabsForResource` via `StackDiff`); selected node id.
- Produces: `<ResourceDrawer node={selectedNode} session={session} context={…} openTab onClose onRemove />`.

- [ ] **Step 1:** Drawer shell (right slide-in, ~`--bg-card`, tokens only) using existing `ui/` primitives (Sheet/Drawer if present, else absolute panel). Tab bar = mono-uppercase CONFIGURATION/DEPLOYMENT/ENVIRONMENT with brand underline active.
- [ ] **Step 2:** Resolve `resourceIdx` from node data; call `useResourceTabProps`; render the matching sub-tab body per active tab.
- [ ] **Step 3:** Dirty dots — `dirtyTabsForResource(draft.resources[idx], baseline.resources[idx])` → amber dot per dirty tab; header "N change" badge from `perResourceDirty`.
- [ ] **Step 4:** Footer — "View logs" (switch to Logs mode, slice 5 stub ok) + "Remove resource" (`onRemove` → slice 4 wires deletion; here just close + callback).
- [ ] **Step 5:** Selection wiring in `CanvasEditor`/`StackCanvasTab` — `onNodeClick` sets `selectedId`; missing/removed id → no drawer (PoSD).
- [ ] **Step 6: Playwright verify** — click a node, edit image + a port + add an env var; assert dirty dots appear on the right tabs, env grouping (addon group) renders, `draft` updated. Screenshot.
- [ ] **Step 7: Commit** (`feat(stacks): node config drawer reusing existing sub-tabs`)

---

# SLICE 3 — Header + dirty + deploy (task-level; expand at execution)

**Deliverable:** Restyled editor header; node cards show New/Edited/Removed marks; Deploy persists via existing PUT.

- **Task 3.1 `CanvasHeader.tsx`** — name, `StatusPill`, mode tab bar (Configuration/Deployments/Logs/Metrics), "N unsaved changes" (from `session.dirty` totals), Deploy button. Reuse `useDeployLifecycle` + `performSave`/`handleSave` already in `detail/index.tsx` (lift or pass down). Replaces the page's current header only when flag on.
- **Task 3.2 Node dirty marks** — extend `ResourceNodeData` with `dirtyState: "new"|"edited"|"removed"|undefined`; compute in `deriveGraph`/StackCanvasTab from `session.dirty.dirtyResourceIdx` + `pendingDetach` + added-but-unsaved (index ≥ baseline length). Add a unit test to `derive-graph.test.ts` for the mark mapping. Card renders brand border/badge.
- **Task 3.3 Deploy wiring** — Deploy button → `handleSave()`; verify the unchanged PUT fires with full `spec.connections`. Playwright: edit → unsaved count increments → Deploy → network PUT 200 → marks clear.
- **Verify:** Playwright end-to-end edit→deploy; assert PUT payload connections present.

# SLICE 4 — Add / remove resource (task-level)

**Deliverable:** Add resources/addons from a popover; remove from drawer; volume chips editable.

- **Task 4.1 `AddResourcePopover.tsx`** — "+ Add resource" button (canvas top-left) → popover hosting the **extracted `BlockPicker`** (`catalog`, `categories`, `addedIds`, `onAdd`, `query`). Reuse block catalog from `@/pages/stacks/data/blocks`.
- **Task 4.2 Add → draft** — `onAdd(blockId)` maps preset → new `FormStackResourceData` (or addon link via `setLinkedAddonIds`); append through `session.updateResources(prev => [...prev, newResource])`. New node appears marked "new" (slice 3 mark).
- **Task 4.3 Remove** — drawer "Remove resource" → `session.updateResources(prev => prev.filter)` (or `pendingDetach` for linked addons); node marked removed / dropped per existing semantics.
- **Task 4.4 Volume chips** — add/edit volumes via drawer Volumes section (reuse `stack-volumes-form`); chips reflect `volume_mounts`. Unit test: deriveGraph volume folding already covers display.
- **Verify:** Playwright: add redis, wire an env var to it (edge appears), deploy.

# SLICE 5 — Redesigned ops views (task-level)

**Deliverable:** Deployments / Logs / Metrics re-skinned; mode tab bar switches.

- **Task 5.1 Mode switching** — `CanvasHeader` mode tabs drive which view renders (canvas vs Deployments/Logs/Metrics). Reuse existing `deployments-tab`, `stack-logs-tab`, `stack-metrics-tab` data hooks.
- **Task 5.2 `DeploymentsView` / `LogsView` / `MetricsView`** — re-skin to design (tokens, mono labels) wrapping existing tab components; keep data wiring.
- **Verify:** Playwright: switch each mode, assert content + redesigned styling.

# SLICE 6 — Wizard hand-off + polish (task-level)

**Deliverable:** Create flow lands in canvas; perf hardened; edge cases handled.

- **Task 6.1 Re-point composer CTA** — PR #129 `block-composer` "Drag, wire, and configure on the canvas" CTA navigates to the stack detail with canvas flag active for new stacks.
- **Task 6.2 Perf** — confirm `nodeTypes` module-scope, `ResourceNode` memoized, narrow `useStore` selectors, enable `onlyRenderVisibleElements` only if measured beneficial, `snapGrid`. Re-layout only on topology change.
- **Task 6.3 Edge cases** — empty stack (no resources) empty-state; unattached volumes in drawer; flag-off path still pristine.
- **Task 6.4 Cleanup** — once stable, plan removal of old form (separate follow-up; keep behind flag until sign-off).
- **Verify:** Playwright full create→canvas→deploy.

---

## Self-Review

- **Spec coverage:** D1 (scope) → all slices stay on show/edit page + slice 6 hand-off; D2 (auto-layout) → Task 1.2; D3 (flag) → Tasks 0.3/0.4; D4 (derived edges) → Task 1.1 rules; D5 (worker out) → `NODE_KIND` has only service/addon; D6 (theme adapt) → `colorMode="system"` + token dark overrides; D7 (volume chips) → Task 1.1 + 4.4. Env grouping → Task 2.1/2.2 reuse. Dirty/deploy → Slice 3. Ops views → Slice 5. Testing → Vitest pure tests + Playwright per slice.
- **Type consistency:** `NODE_KIND`, `ResourceNodeData`, `CanvasGraph`, `deriveGraph`, `layoutGraph`, `useResourceTabProps` names used consistently across tasks. Sub-tab prop interfaces match the verbatim signatures extracted from the codebase.
- **Placeholders:** Slices 3–6 are intentionally task-level (YAGNI — expanded just-in-time at execution with full bite-sized steps); slices 0–2 (execute-now) are fully detailed with real code and commands. No "TBD/implement later" inside an execute-now task.
- **Risk carried from spec:** token alias-vs-author resolved (author against `--brand`, add new families only — Task 0.2); block-picker confirmed extractable (Task 4.1).
