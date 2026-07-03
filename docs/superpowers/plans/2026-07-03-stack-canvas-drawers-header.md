# Stack Canvas Drawer Stack, Collapsible Header, Labels & PUBLIC Row Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Floating stackable resource drawers, a collapsible editor header, always-editable stack labels, and a PUBLIC endpoint row in the stack canvas editor.

**Architecture:** All frontend, inside the existing canvas editor (`frontend/src/pages/stacks/components/canvas/`). Drawer state becomes a pure `DrawerEntry[]` stack owned by `StackCanvasTab`; a new presentational `DrawerStack` renders floating panels. `CanvasEditorShell` gains a collapsed 44px mode and the PUBLIC row. Labels persist for deployed stacks via a full-stack PUT built from the live server stack (reusing draft-sync cleaning helpers).

**Tech Stack:** React 19, Tailwind v4 (brand tokens from `frontend/src/index.css`), Radix primitives via `frontend/src/components/ui/`, vitest + @testing-library/react (jsdom), lucide-react icons.

**Spec:** `docs/superpowers/specs/2026-07-03-stack-canvas-drawers-header-design.md`
**Design reference:** `docs/superpowers/design-refs/2026-07-03-stack-creation-redesign-design-reference.md`

## Global Constraints

- **No raw hex/rgb values.** Use `index.css` tokens / existing Tailwind token classes (`bg-background`, `border-border`, `text-fg-muted`, `bg-brand`, `text-danger`, `bg-muted`, `shadow-lg`…). The design's scrim `rgba(10,14,20,0.55)` maps to `bg-background/55`.
- **No magic strings** — reuse existing constants (`SYNC_STATUS` from `@/pages/stacks/lib/draft-sync/constants`, `USER_DEFINED_LABEL_KEY` in `detail/index.tsx`); define new constants where needed.
- All frontend commands run from repo root with `pnpm --prefix frontend …`.
- Run a single test file: `pnpm --prefix frontend test -- --run src/path/to/file.test.ts`
- Type-check: `pnpm --prefix frontend exec tsc -b` · Lint: `pnpm --prefix frontend lint`
- Existing test conventions: see `frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx` — `// @vitest-environment jsdom` header, `@testing-library/jest-dom/vitest`, `afterEach(cleanup)`.
- Commit after each task with a `feat(stacks):`/`refactor(stacks):` message.

---

### Task 1: Drawer-stack pure module

**Files:**
- Create: `frontend/src/pages/stacks/lib/canvas/drawer-stack.ts`
- Test: `frontend/src/pages/stacks/lib/canvas/tests/drawer-stack.test.ts`

**Interfaces:**
- Consumes: nothing.
- Produces (used by Tasks 2, 3, 4, 5):
  - `type DrawerEntry = { kind: "resource"; index: number } | { kind: "volume"; name: string }`
  - `entryKey(e: DrawerEntry): string`
  - `replaceStack(entry: DrawerEntry): DrawerEntry[]`
  - `pushEntry(stack: DrawerEntry[], entry: DrawerEntry): DrawerEntry[]` (dedupe-truncate)
  - `truncateTo(stack: DrawerEntry[], depth: number): DrawerEntry[]`
  - `popEntry(stack: DrawerEntry[]): DrawerEntry[]`

- [ ] **Step 1: Write the failing test**

```ts
// frontend/src/pages/stacks/lib/canvas/tests/drawer-stack.test.ts
import { describe, it, expect } from "vitest";
import {
  entryKey,
  replaceStack,
  pushEntry,
  truncateTo,
  popEntry,
  type DrawerEntry,
} from "../drawer-stack";

const r = (index: number): DrawerEntry => ({ kind: "resource", index });
const v = (name: string): DrawerEntry => ({ kind: "volume", name });

describe("drawer-stack", () => {
  it("entryKey distinguishes kinds and identities", () => {
    expect(entryKey(r(0))).not.toBe(entryKey(r(1)));
    expect(entryKey(r(0))).not.toBe(entryKey(v("0")));
    expect(entryKey(v("data"))).toBe(entryKey(v("data")));
  });

  it("replaceStack yields a single-entry stack", () => {
    expect(replaceStack(r(2))).toEqual([r(2)]);
  });

  it("pushEntry appends a new entry", () => {
    expect(pushEntry([r(0)], v("data"))).toEqual([r(0), v("data")]);
  });

  it("pushEntry truncates to an existing entry instead of duplicating", () => {
    const stack = [r(0), v("data"), r(1)];
    expect(pushEntry(stack, v("data"))).toEqual([r(0), v("data")]);
    expect(pushEntry(stack, r(0))).toEqual([r(0)]);
  });

  it("truncateTo keeps entries up to and including depth", () => {
    const stack = [r(0), v("data"), r(1)];
    expect(truncateTo(stack, 1)).toEqual([r(0), v("data")]);
    expect(truncateTo(stack, 0)).toEqual([r(0)]);
  });

  it("popEntry removes the front entry; empty stays empty", () => {
    expect(popEntry([r(0), v("data")])).toEqual([r(0)]);
    expect(popEntry([])).toEqual([]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/canvas/tests/drawer-stack.test.ts`
Expected: FAIL — cannot resolve `../drawer-stack`.

- [ ] **Step 3: Write minimal implementation**

```ts
// frontend/src/pages/stacks/lib/canvas/drawer-stack.ts

/** One open panel in the floating drawer stack. */
export type DrawerEntry =
  | { kind: "resource"; index: number }
  | { kind: "volume"; name: string };

export function entryKey(e: DrawerEntry): string {
  return e.kind === "resource" ? `resource:${e.index}` : `volume:${e.name}`;
}

export function replaceStack(entry: DrawerEntry): DrawerEntry[] {
  return [entry];
}

/** Push, or truncate back to the entry if it is already open (no duplicates). */
export function pushEntry(stack: DrawerEntry[], entry: DrawerEntry): DrawerEntry[] {
  const existing = stack.findIndex((e) => entryKey(e) === entryKey(entry));
  return existing >= 0 ? stack.slice(0, existing + 1) : [...stack, entry];
}

export function truncateTo(stack: DrawerEntry[], depth: number): DrawerEntry[] {
  return stack.slice(0, depth + 1);
}

export function popEntry(stack: DrawerEntry[]): DrawerEntry[] {
  return stack.slice(0, -1);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/canvas/tests/drawer-stack.test.ts`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/drawer-stack.ts frontend/src/pages/stacks/lib/canvas/tests/drawer-stack.test.ts
git commit -m "feat(stacks): pure drawer-stack model for floating canvas drawers"
```

---

### Task 2: DrawerStack presentational component

**Files:**
- Create: `frontend/src/pages/stacks/components/canvas/DrawerStack.tsx`
- Test: `frontend/src/pages/stacks/components/canvas/tests/drawer-stack-component.test.tsx`

**Interfaces:**
- Consumes: `DrawerEntry`, `entryKey` from Task 1.
- Produces (used by Task 3):

```ts
export interface DrawerPanelDescriptor {
  entry: DrawerEntry;
  title: string;
  /** Small leading glyph for the behind-panel header. */
  icon: ReactNode;
}
interface DrawerStackProps {
  /** Bottom → top; last item is the front (interactive) panel. */
  panels: DrawerPanelDescriptor[];
  /** Body for the front panel only. */
  front: ReactNode;
  onTruncate: (depth: number) => void;
  onPop: () => void;
  onCloseAll: () => void;
}
export function DrawerStack(props: DrawerStackProps): ReactNode
```

Geometry (from design reference): front panel `fixed top-3 bottom-3 right-3 w-[600px] z-[200] rounded-lg border border-border bg-background shadow-lg`, entry animation 260ms translateX(34px)→0. Behind panel at depth `d` back from front: `top: 12 + 10·d px`, `bottom: 12 + 10·d px`, `right: 12 + 16·d px`, `zIndex: 200 − d`, covered by a `bg-background/55` scrim, header-only. Keyboard: `Escape` → `onPop()`, `Shift+Escape` → `onCloseAll()`; both skipped when `event.defaultPrevented` (a Radix layer already consumed it). No page backdrop.

- [ ] **Step 1: Write the failing test**

```tsx
// frontend/src/pages/stacks/components/canvas/tests/drawer-stack-component.test.tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { DrawerStack } from "../DrawerStack";
import type { DrawerEntry } from "@/pages/stacks/lib/canvas/drawer-stack";

afterEach(cleanup);

const r = (index: number): DrawerEntry => ({ kind: "resource", index });
const v = (name: string): DrawerEntry => ({ kind: "volume", name });

function setup(overrides: Partial<Parameters<typeof DrawerStack>[0]> = {}) {
  const onTruncate = vi.fn();
  const onPop = vi.fn();
  const onCloseAll = vi.fn();
  render(
    <DrawerStack
      panels={[
        { entry: r(0), title: "web", icon: <span /> },
        { entry: v("data"), title: "data", icon: <span /> },
      ]}
      front={<div>volume body</div>}
      onTruncate={onTruncate}
      onPop={onPop}
      onCloseAll={onCloseAll}
      {...overrides}
    />,
  );
  return { onTruncate, onPop, onCloseAll };
}

describe("DrawerStack", () => {
  it("renders the front panel body and a header-only behind panel", () => {
    setup();
    expect(screen.getByText("volume body")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Bring web panel to front" })).toBeInTheDocument();
  });

  it("staggers behind panels and layers z-indices", () => {
    setup();
    const behind = screen.getByTestId("drawer-panel-0");
    const front = screen.getByTestId("drawer-panel-1");
    expect(behind.style.right).toBe("28px"); // 12 + 16·1
    expect(behind.style.zIndex).toBe("199");
    expect(front.style.right).toBe("12px");
    expect(front.style.zIndex).toBe("200");
  });

  it("clicking a behind panel truncates to its depth", () => {
    const { onTruncate } = setup();
    fireEvent.click(screen.getByRole("button", { name: "Bring web panel to front" }));
    expect(onTruncate).toHaveBeenCalledWith(0);
  });

  it("Escape pops, Shift+Escape closes all", () => {
    const { onPop, onCloseAll } = setup();
    fireEvent.keyDown(window, { key: "Escape" });
    expect(onPop).toHaveBeenCalledTimes(1);
    fireEvent.keyDown(window, { key: "Escape", shiftKey: true });
    expect(onCloseAll).toHaveBeenCalledTimes(1);
  });

  it("renders nothing when the stack is empty", () => {
    render(
      <DrawerStack panels={[]} front={null} onTruncate={vi.fn()} onPop={vi.fn()} onCloseAll={vi.fn()} />,
    );
    expect(screen.queryByTestId("drawer-panel-0")).toBeNull();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/drawer-stack-component.test.tsx`
Expected: FAIL — cannot resolve `../DrawerStack`.

- [ ] **Step 3: Write the implementation**

```tsx
// frontend/src/pages/stacks/components/canvas/DrawerStack.tsx
import { useEffect, type ReactNode } from "react";
import type { DrawerEntry } from "@/pages/stacks/lib/canvas/drawer-stack";
import { entryKey } from "@/pages/stacks/lib/canvas/drawer-stack";

const BASE_INSET_PX = 12;
const STAGGER_Y_PX = 10;
const STAGGER_X_PX = 16;
const FRONT_Z = 200;

export interface DrawerPanelDescriptor {
  entry: DrawerEntry;
  title: string;
  icon: ReactNode;
}

interface DrawerStackProps {
  /** Bottom → top; last item is the front (interactive) panel. */
  panels: DrawerPanelDescriptor[];
  /** Body for the front panel only. */
  front: ReactNode;
  onTruncate: (depth: number) => void;
  onPop: () => void;
  onCloseAll: () => void;
}

/**
 * Floating, stackable drawer panels over the canvas (no backdrop — the canvas
 * stays interactive). Behind panels are header-only and dimmed; clicking one
 * truncates the stack back to it. Esc pops the front panel, Shift+Esc closes all.
 */
export function DrawerStack({ panels, front, onTruncate, onPop, onCloseAll }: DrawerStackProps) {
  const open = panels.length > 0;

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key !== "Escape" || e.defaultPrevented) return;
      if (e.shiftKey) onCloseAll();
      else onPop();
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onPop, onCloseAll]);

  if (!open) return null;
  const frontIdx = panels.length - 1;

  return (
    <>
      {panels.map(({ entry, title, icon }, i) => {
        const depth = frontIdx - i; // 0 = front
        const isFront = depth === 0;
        return (
          <aside
            key={entryKey(entry)}
            data-testid={`drawer-panel-${i}`}
            className="fixed flex w-[600px] max-w-[calc(100vw-24px)] flex-col overflow-hidden rounded-lg border border-border bg-background shadow-lg animate-in slide-in-from-right-8 fade-in duration-[260ms]"
            style={{
              top: BASE_INSET_PX + STAGGER_Y_PX * depth,
              bottom: BASE_INSET_PX + STAGGER_Y_PX * depth,
              right: BASE_INSET_PX + STAGGER_X_PX * depth,
              zIndex: FRONT_Z - depth,
            }}
          >
            {isFront ? (
              front
            ) : (
              <button
                type="button"
                aria-label={`Bring ${title} panel to front`}
                onClick={() => onTruncate(i)}
                className="relative flex h-full w-full flex-col items-stretch text-left"
              >
                <span className="flex items-center gap-2.5 border-b border-border px-4 py-[15px]">
                  <span className="size-[19px] shrink-0 text-brand">{icon}</span>
                  <span className="truncate text-base font-medium text-foreground">{title}</span>
                </span>
                {/* Scrim: dims the parked panel (maps design rgba(10,14,20,.55)). */}
                <span aria-hidden className="absolute inset-0 bg-background/55" />
              </button>
            )}
          </aside>
        );
      })}
    </>
  );
}
```

Note: if `animate-in`/`slide-in-from-right-8` utilities are unavailable in this Tailwind setup (check `pnpm --prefix frontend exec tsc -b` won't catch CSS — verify visually in Task 10), drop the animation classes rather than adding raw keyframes.

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/drawer-stack-component.test.tsx`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/DrawerStack.tsx frontend/src/pages/stacks/components/canvas/tests/drawer-stack-component.test.tsx
git commit -m "feat(stacks): DrawerStack floating panel component"
```

---

### Task 3: Wire StackCanvasTab to the drawer stack (resource panels)

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx` (chrome only)

**Interfaces:**
- Consumes: Task 1 model, Task 2 `DrawerStack`/`DrawerPanelDescriptor`.
- Produces (used by Tasks 4, 5): `StackCanvasFlow` holds `drawerStack: DrawerEntry[]` state plus `openVolume(name: string)` callback (defined here, threaded into the drawer in Task 5).

Behavior:
- `selectedIndex` state → `drawerStack` state.
- Canvas node click → `setDrawerStack(replaceStack({ kind: "resource", index }))`.
- Remove the canvas-resize `useEffect` (`drawerWasOpen` ref + `fitView` timer, currently lines 101–110) — panels float, canvas width never changes.
- `ResourceDrawer` loses its own fixed width/border chrome (`aside` → plain `div`, `w-[496px] border-l shadow-lg` dropped) since `DrawerStack` owns panel chrome; `onClose` maps to pop.
- Resource removal closes the whole stack (indices shift).
- Reconcile: drop entries whose resource index or volume name no longer exists in the draft.

- [ ] **Step 1: Update ResourceDrawer chrome**

In `ResourceDrawer.tsx`, change the root element (currently `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx:156-159`):

```tsx
    <div
      className="flex h-full w-full flex-col bg-background"
      data-testid="resource-drawer"
    >
```

and the matching closing tag `</aside>` → `</div>`. Nothing else in the file changes.

- [ ] **Step 2: Rewrite drawer state in StackCanvasTab**

In `StackCanvasTab.tsx`:

Replace the import of `ResourceDrawer` block with additional imports:

```tsx
import { ResourceDrawer } from "./ResourceDrawer";
import { DrawerStack, type DrawerPanelDescriptor } from "./DrawerStack";
import {
  replaceStack,
  pushEntry,
  truncateTo,
  popEntry,
  entryKey,
  type DrawerEntry,
} from "@/pages/stacks/lib/canvas/drawer-stack";
import { nodePresentation } from "@/pages/stacks/lib/canvas/node-presentation";
import { NodeGlyph } from "./nodes/node-glyph";
import { HardDrive } from "lucide-react";
```

Replace `const [selectedIndex, setSelectedIndex] = useState<number | null>(null);` with:

```tsx
  const [drawerStack, setDrawerStack] = useState<DrawerEntry[]>([]);
```

Delete the whole drawer-resize effect and its comment (the `drawerWasOpen` ref + `useEffect` keyed on `selectedIndex`, currently lines 101–110).

In `onNodeClick`, replace `setSelectedIndex(idx);` with:

```tsx
      setDrawerStack(replaceStack({ kind: "resource", index: idx }));
```

Replace `closeDrawer`/`removeResource` with:

```tsx
  const popDrawer = useCallback(() => setDrawerStack((s) => popEntry(s)), []);
  const closeAllDrawers = useCallback(() => setDrawerStack([]), []);
  const truncateDrawers = useCallback((depth: number) => setDrawerStack((s) => truncateTo(s, depth)), []);
  const openVolume = useCallback(
    (name: string) => setDrawerStack((s) => pushEntry(s, { kind: "volume", name })),
    [],
  );
  void openVolume; // threaded into the drawer body in the mount-row task
  const removeResource = useCallback(
    (idx: number) => {
      session.updateResources((prev) => prev.filter((_, i) => i !== idx));
      setDrawerStack([]);
    },
    [session],
  );

  // Drop panels whose target no longer exists in the draft (deleted resource/volume).
  useEffect(() => {
    setDrawerStack((s) =>
      s.filter((e) =>
        e.kind === "resource"
          ? e.index < resources.length
          : (session.isActive ? session.draft.volumes : baselineVolumes).some((v) => v.name === e.name),
      ),
    );
  }, [resources.length, session.isActive, session.draft.volumes, baselineVolumes]);
```

Build panel descriptors + front body (place after `removeResource`):

```tsx
  const panels: DrawerPanelDescriptor[] = useMemo(
    () =>
      drawerStack.map((entry) => {
        if (entry.kind === "resource") {
          const r = resources[entry.index] ?? {};
          const pres = nodePresentation({
            isAddon: false,
            image: r.image_spec?.image,
            hasBuild: !!r.build_spec,
            ports: (r.ports ?? []).map((p) => ({
              number: p.number,
              protocol: p.protocol,
              exposedToPublic: p.exposed_to_public,
            })),
          });
          return {
            entry,
            title: r.name || `Resource ${entry.index + 1}`,
            icon: <NodeGlyph glyph={pres.glyph} className="size-[19px]" />,
          };
        }
        return { entry, title: entry.name, icon: <HardDrive className="size-[19px]" /> };
      }),
    [drawerStack, resources],
  );

  const frontEntry = drawerStack[drawerStack.length - 1];
  const frontBody =
    frontEntry?.kind === "resource" ? (
      <ResourceDrawer
        key={entryKey(frontEntry)}
        resourceIndex={frontEntry.index}
        session={session}
        baselineResources={baselineResources}
        connectionAddonIds={connectionAddonIds}
        errors={errors[frontEntry.index] ?? {}}
        onClose={popDrawer}
        onRemove={removeResource}
        onViewLogs={onViewLogs}
      />
    ) : null; // volume front body arrives with VolumeDrawer (next task)
```

Replace the JSX return's drawer segment (`{selectedIndex != null && (<ResourceDrawer …/>)}`) with:

```tsx
      <DrawerStack
        panels={panels}
        front={frontBody}
        onTruncate={truncateDrawers}
        onPop={popDrawer}
        onCloseAll={closeAllDrawers}
      />
```

(The wrapping `<div className="flex h-full w-full">` and canvas child stay.)

- [ ] **Step 3: Type-check, lint, run canvas tests**

Run:
```
pnpm --prefix frontend exec tsc -b
pnpm --prefix frontend lint
pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas
```
Expected: clean tsc/lint; existing canvas tests PASS (drawer stack tests from Tasks 1–2 included).

- [ ] **Step 4: Manual smoke (dev server)**

Run `pnpm --prefix frontend dev`, open a stack at http://localhost:5173: node click opens floating drawer (canvas doesn't shrink), second node click replaces it, Esc closes, edits still autosave.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx
git commit -m "feat(stacks): float resource drawer over canvas via drawer stack"
```

---

### Task 4: VolumeFields extraction + VolumeDrawer body

**Files:**
- Create: `frontend/src/pages/stacks/components/shared/volume-fields.tsx`
- Modify: `frontend/src/pages/stacks/components/shared/stack-volume-item.tsx`
- Create: `frontend/src/pages/stacks/components/canvas/VolumeDrawer.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx` (front body for volume entries)
- Test: `frontend/src/pages/stacks/components/canvas/tests/volume-drawer.test.tsx`

**Interfaces:**
- Consumes: Task 3 state (`drawerStack`, `popDrawer`, `removeVolume` added here).
- Produces:

```ts
// volume-fields.tsx — the spec/mounts/remove form body shared by accordion + drawer
export function VolumeFields(props: {
  volume: Partial<VolumeFormData>;
  index: number;
  onChange: (index: number, updated: Partial<VolumeFormData>) => void;
  onRemove: (index: number) => void;
  errors: { [field: string]: string | undefined };
  allVolumes: Partial<VolumeFormData>[];
  allStackResources?: Partial<FormStackResourceData>[];
}): ReactNode

// VolumeDrawer.tsx — drawer-shaped wrapper (header + VolumeFields)
export function VolumeDrawer(props: {
  volumeName: string;
  session: UseStackEditSession;
  onClose: () => void;
}): ReactNode
```

- [ ] **Step 1: Extract VolumeFields from StackVolumeItem**

Create `volume-fields.tsx` by moving the *content* of `stack-volume-item.tsx`'s `<AccordionContent>` inner `<div className="px-4 space-y-6">…</div>` (the "Volume specification" grid, "Mount details" list, and "Remove volume" footer — currently `stack-volume-item.tsx:108-195`) into the new component verbatim, along with the helpers it uses (`update`, `isDuplicate`, `mountingInfo` derivation — currently lines 36–61). Imports move with them (`FieldShell`, `Input`, `Button`, `CornerDownRight`, `Trash2`, types).

`stack-volume-item.tsx` keeps: status dot + accordion header (lines 42–106) and renders inside `<AccordionContent>`:

```tsx
        <div className="px-4 space-y-6">
          <VolumeFields
            volume={volume}
            index={index}
            onChange={onChange}
            onRemove={onRemove}
            errors={errors}
            allVolumes={allVolumes}
            allStackResources={allStackResources}
          />
        </div>
```

(`isDuplicate`/`mountingInfo` move into `VolumeFields`; the accordion header keeps its own `mountingInfo` copy only if it references it — it does, via `mountedBy`. To stay DRY, export the derivation from `volume-fields.tsx`:)

```ts
export function volumeMountingInfo(
  volume: Partial<VolumeFormData>,
  allStackResources: Partial<FormStackResourceData>[],
): { resourceName: string; targetPath: string }[]
```

and use it in both files.

- [ ] **Step 2: Write the failing VolumeDrawer test**

```tsx
// frontend/src/pages/stacks/components/canvas/tests/volume-drawer.test.tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { VolumeDrawer } from "../VolumeDrawer";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";

afterEach(cleanup);

function makeSession(volumes: { name: string; spec?: { size?: string } }[]) {
  const updateVolumes = vi.fn();
  const updateResources = vi.fn();
  return {
    session: {
      isActive: true,
      draft: { resources: [{ name: "web", volume_mounts: [{ source_volume_name: "data", target_path: "/data" }] }], volumes },
      updateVolumes,
      updateResources,
    } as unknown as UseStackEditSession,
    updateVolumes,
  };
}

describe("VolumeDrawer", () => {
  it("renders the volume's fields and mount details", () => {
    const { session } = makeSession([{ name: "data", spec: { size: "1Gi" } }]);
    render(<VolumeDrawer volumeName="data" session={session} onClose={vi.fn()} />);
    expect(screen.getByDisplayValue("data")).toBeInTheDocument();
    expect(screen.getByDisplayValue("1Gi")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument(); // mounted-by row
  });

  it("edits flow into session.updateVolumes", () => {
    const { session, updateVolumes } = makeSession([{ name: "data", spec: { size: "1Gi" } }]);
    render(<VolumeDrawer volumeName="data" session={session} onClose={vi.fn()} />);
    fireEvent.change(screen.getByDisplayValue("1Gi"), { target: { value: "2Gi" } });
    expect(updateVolumes).toHaveBeenCalled();
  });

  it("close button calls onClose", () => {
    const { session } = makeSession([{ name: "data" }]);
    const onClose = vi.fn();
    render(<VolumeDrawer volumeName="data" session={session} onClose={onClose} />);
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    expect(onClose).toHaveBeenCalled();
  });
});
```

- [ ] **Step 3: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/volume-drawer.test.tsx`
Expected: FAIL — cannot resolve `../VolumeDrawer`.

- [ ] **Step 4: Implement VolumeDrawer**

```tsx
// frontend/src/pages/stacks/components/canvas/VolumeDrawer.tsx
import { useCallback } from "react";
import { X, HardDrive } from "lucide-react";
import type { UseStackEditSession } from "@/pages/stacks/hooks/use-stack-edit-session";
import type { FormVolumeExtendedData as VolumeFormData } from "@/pages/stacks/schemas/form-schema";
import { VolumeFields } from "@/pages/stacks/components/shared/volume-fields";

interface VolumeDrawerProps {
  /** Draft volume name — the stack entry's identity. */
  volumeName: string;
  session: UseStackEditSession;
  onClose: () => void;
}

/** Drawer body for a volume pushed from a service's mount row. */
export function VolumeDrawer({ volumeName, session, onClose }: VolumeDrawerProps) {
  const volumes = session.draft.volumes;
  const index = volumes.findIndex((v) => v.name === volumeName);
  const volume = (volumes[index] ?? {}) as Partial<VolumeFormData>;

  const onChange = useCallback(
    (idx: number, updated: Partial<VolumeFormData>) => {
      session.updateVolumes((prev) => prev.map((v, i) => (i === idx ? updated : v)));
    },
    [session],
  );

  const onRemove = useCallback(
    (idx: number) => {
      session.updateVolumes((prev) => prev.filter((_, i) => i !== idx));
      onClose();
    },
    [session, onClose],
  );

  if (index < 0) return null;

  return (
    <div className="flex h-full w-full flex-col bg-background" data-testid="volume-drawer">
      <div className="flex flex-none items-center gap-3 border-b border-border px-4 py-[15px]">
        <HardDrive className="size-[19px] shrink-0 text-brand" />
        <div className="min-w-0 flex-1 leading-tight">
          <div className="truncate text-base font-medium text-foreground">{volume.name || volumeName}</div>
          <div className="truncate font-mono text-[11px] text-fg-muted">
            {volume.spec?.size || "size unset"} · {volume.spec?.access_mode || "ReadWriteOnce"}
          </div>
        </div>
        <button
          type="button"
          onClick={onClose}
          aria-label="Close"
          className="shrink-0 rounded p-1 text-fg-muted hover:bg-muted hover:text-foreground"
        >
          <X className="size-[18px]" />
        </button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-4 py-4">
        <VolumeFields
          volume={volume}
          index={index}
          onChange={onChange}
          onRemove={onRemove}
          errors={{}}
          allVolumes={volumes}
          allStackResources={session.draft.resources}
        />
      </div>
    </div>
  );
}
```

- [ ] **Step 5: Render volume front bodies in StackCanvasTab**

In `StackCanvasTab.tsx`, import `VolumeDrawer` and replace the `frontBody` ternary's `null` volume branch:

```tsx
    ) : frontEntry ? (
      <VolumeDrawer key={entryKey(frontEntry)} volumeName={frontEntry.name} session={session} onClose={popDrawer} />
    ) : null;
```

- [ ] **Step 6: Run tests, type-check, lint**

Run:
```
pnpm --prefix frontend test -- --run src/pages/stacks
pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```
Expected: all stacks tests PASS (including untouched `stack-volume-item` consumers), clean tsc/lint.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/shared/volume-fields.tsx frontend/src/pages/stacks/components/shared/stack-volume-item.tsx frontend/src/pages/stacks/components/canvas/VolumeDrawer.tsx frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx frontend/src/pages/stacks/components/canvas/tests/volume-drawer.test.tsx
git commit -m "feat(stacks): volume drawer body reusing extracted VolumeFields"
```

---

### Task 5: Mount-row push trigger

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/hooks/use-resource-tab-props.ts` (context type + pass-through)
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx` (mount-row open button)
- Modify: `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx` (accept + forward `onOpenVolume`)
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx` (pass `openVolume`)

**Interfaces:**
- Consumes: `openVolume(name: string)` from Task 3.
- Produces: `onOpenVolume?: (name: string) => void` threaded: StackCanvasTab → ResourceDrawer prop → `useResourceTabProps` context → configuration tab.

- [ ] **Step 1: Thread the callback**

1. `use-resource-tab-props.ts`: add `onOpenVolume?: (name: string) => void` to the `context` parameter type and include it in the returned `configurationProps` (follow how `onCreateVolume` flows — same shape, same destination).
2. `ResourceDrawer.tsx`: add prop `onOpenVolume?: (name: string) => void` to `ResourceDrawerProps`, pass into `useResourceTabProps` context alongside `onCreateVolume`.
3. `StackCanvasTab.tsx`: remove the `void openVolume;` placeholder line and pass `onOpenVolume={openVolume}` to `<ResourceDrawer …>`.

- [ ] **Step 2: Add the open button to mount rows**

In `stack-resource-configuration-tab.tsx`, the mount rows render around line 585 (`{(draft.volume_mounts || []).map((vm, vmIdx) => (…))}`), each with a source-volume `FieldShell` select. Destructure `onOpenVolume` from the configuration props (same place `onCreateVolume` is received). Inside each mount row, immediately after the source-volume select's `FieldShell`, add:

```tsx
                {onOpenVolume && vm.source_volume_name && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-7 shrink-0 self-end text-fg-muted hover:text-brand"
                    aria-label={`Open volume ${vm.source_volume_name}`}
                    title="Open volume settings"
                    onClick={() => onOpenVolume(vm.source_volume_name!)}
                  >
                    <ArrowUpRight className="size-3.5" />
                  </Button>
                )}
```

Add `ArrowUpRight` to the file's existing `lucide-react` import. Match the row's existing flex layout (the button sits beside the select; adjust wrapper to `flex items-end gap-1.5` if the select isn't already in a flex row — keep the diff minimal).

- [ ] **Step 3: Verify behavior + regression**

Run:
```
pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
pnpm --prefix frontend test -- --run src/pages/stacks
```
Expected: clean. Manual dev-server check: open service drawer → click mount row's open button → volume drawer pushes on top, service panel parks behind dimmed; clicking parked panel returns; Esc pops; pushing the same volume twice doesn't duplicate.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/stacks/components/shared/hooks/use-resource-tab-props.ts frontend/src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx
git commit -m "feat(stacks): push volume drawer from service mount rows"
```

---

### Task 6: Collapsible header

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (pass `stackId`)
- Test: extend `frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`

**Interfaces:**
- Consumes: nothing new.
- Produces: `CanvasEditorShellProps` gains `stackId?: string` (collapse-persistence key; drafts fall back to a shared draft key).

Behavior: `collapsed` boolean, lazy-init from `localStorage`, key `stackdome.editor-header-collapsed.<stackId|draft>`. Toggle via chevron button (`aria-label="Collapse header"` / `"Expand header"`) and ⌘./Ctrl+. window shortcut. Collapsed renders ONE 44px bar: chevron, name at 14px, status dot, compact tabs (icon + 12px label), unsaved summary, autosave status, primary button. Expanded renders today's full header (name row, subtitle, labels — PUBLIC row joins in Task 9) plus the chevron.

- [ ] **Step 1: Write failing tests** (append to `canvas-editor-shell.test.tsx`)

```tsx
describe("CanvasEditorShell collapse", () => {
  afterEach(() => localStorage.clear());

  it("collapses to a compact bar and hides labels/subtitle", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1"
      labels={[{ key: "user", value: "prod" }]} />);
    expect(screen.getByText("prod")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    expect(screen.queryByText("prod")).toBeNull();
    expect(screen.queryByText("0 services · 0 volumes")).toBeNull();
    expect(screen.getByText("acme")).toBeInTheDocument(); // compact bar name
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("persists collapsed state per stack id", () => {
    localStorage.setItem("stackdome.editor-header-collapsed.s1", "1");
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("toggles via Cmd+.", () => {
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" />);
    fireEvent.keyDown(window, { key: ".", metaKey: true });
    expect(screen.getByRole("button", { name: "Expand header" })).toBeInTheDocument();
  });

  it("keeps tabs clickable while collapsed", () => {
    const onTabChange = vi.fn();
    render(<CanvasEditorShell {...base} stackName="acme" nameEditable={false} stackId="s1" onTabChange={onTabChange} />);
    fireEvent.click(screen.getByRole("button", { name: "Collapse header" }));
    fireEvent.click(screen.getByRole("button", { name: /Logs/ }));
    expect(onTabChange).toHaveBeenCalledWith("logs");
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`
Expected: new tests FAIL (no collapse button).

- [ ] **Step 3: Implement**

In `CanvasEditorShell.tsx`:

```tsx
// imports: add ChevronDown, ChevronRight to the lucide-react import; add useEffect, useCallback to react import.

const COLLAPSE_KEY_PREFIX = "stackdome.editor-header-collapsed.";
const DRAFT_COLLAPSE_ID = "draft";
```

Props: add `stackId?: string` to `CanvasEditorShellProps` (doc: "Persistence key for header collapse; falls back to a shared draft key.").

State + shortcut inside the component:

```tsx
  const collapseKey = `${COLLAPSE_KEY_PREFIX}${stackId ?? DRAFT_COLLAPSE_ID}`;
  const [collapsed, setCollapsed] = useState<boolean>(() => {
    try {
      return localStorage.getItem(collapseKey) === "1";
    } catch {
      return false;
    }
  });
  const toggleCollapsed = useCallback(() => {
    setCollapsed((c) => {
      const next = !c;
      try {
        localStorage.setItem(collapseKey, next ? "1" : "0");
      } catch {
        /* storage unavailable — collapse stays session-local */
      }
      return next;
    });
  }, [collapseKey]);

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "." && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggleCollapsed();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [toggleCollapsed]);
```

Chevron button (shared by both modes):

```tsx
  const chevron = (
    <button
      type="button"
      onClick={toggleCollapsed}
      aria-label={collapsed ? "Expand header" : "Collapse header"}
      title={`${collapsed ? "Expand" : "Collapse"} header (⌘.)`}
      className="flex size-6 flex-none items-center justify-center rounded text-fg-muted hover:bg-muted hover:text-foreground"
    >
      {collapsed ? <ChevronRight className="size-4" /> : <ChevronDown className="size-4" />}
    </button>
  );
```

Render: wrap the current header + tab row in `{!collapsed && (<>…</>)}` (add `chevron` as the first child of the name row's flex container). Add the collapsed bar before it:

```tsx
      {collapsed && (
        <div className="flex h-11 flex-none items-center gap-3 border-b border-border px-4">
          {chevron}
          <span className="truncate text-[14px] font-medium text-foreground">{stackName}</span>
          {statusState && (
            <span
              aria-label={`status ${statusState}`}
              className={cn(
                "size-2 flex-none rounded-full",
                variantFromState(statusState) === "success" ? "bg-success" : "bg-warn",
              )}
            />
          )}
          <div className="mx-2 flex items-center gap-1">
            {EDITOR_TABS.map(({ id, label, Icon }) => (
              <button
                key={id}
                type="button"
                onClick={() => onTabChange(id)}
                className={cn(
                  "flex items-center gap-1.5 rounded-md border px-2 py-1 text-[12px] font-medium transition-colors",
                  activeTab === id
                    ? "border-brand bg-brand-bg text-brand"
                    : "border-transparent text-muted-foreground hover:text-foreground",
                )}
              >
                <Icon className="size-3.5" />
                {label}
              </button>
            ))}
          </div>
          <div className="flex-1" />
          {hasUnsaved && (
            <span className="font-mono text-[11px] text-brand">
              {dirtyTotal} unsaved {dirtyTotal === 1 ? "change" : "changes"}
            </span>
          )}
          {!isDraft && <AutosaveStatus status={syncStatus} />}
          {primaryButton}
        </div>
      )}
```

(The expanded tab row keeps its Configuration badge; the collapsed bar intentionally omits it — the unsaved summary covers dirty state.)

In `detail/index.tsx`, pass `stackId={effectiveStack?.id}` to `<CanvasEditorShell …>`.

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`
Expected: PASS, including pre-existing shell tests.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx frontend/src/pages/stacks/components/detail/index.tsx frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx
git commit -m "feat(stacks): collapsible canvas editor header with compact bar"
```

---

### Task 7: Labels — normalization + always editable (deployed persistence)

**Files:**
- Create: `frontend/src/pages/stacks/lib/labels.ts`
- Test: `frontend/src/pages/stacks/lib/tests/labels.test.ts`
- Modify: `frontend/src/pages/stacks/lib/draft-sync/snapshot-to-update.ts` (add `stackToUpdateRequest`)
- Test: extend `frontend/src/pages/stacks/lib/draft-sync/tests/snapshot-to-update.test.ts` (create if absent)
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx`

**Interfaces:**
- Consumes: `updateStack` (`frontend/src/api/stacks.ts:56`), `cleanServerResource`/`cleanVolume`/`withExplicitEmptyCollections` (already in `snapshot-to-update.ts` / `server-state.ts`), `SYNC_STATUS`.
- Produces:
  - `normalizeLabel(raw: string): string`
  - `stackToUpdateRequest(stack: Stack, labels: Stack["labels"]): StackUpdateRequest`

**CRITICAL:** a stack PUT replaces ALL connections from `spec.connections` — the request body MUST carry the live stack's full resource/volume/connection set or bindings get wiped (see memory + `snapshotToUpdateRequest`'s replace-all comment). The test below locks this in.

- [ ] **Step 1: Write failing label-normalization test**

```ts
// frontend/src/pages/stacks/lib/tests/labels.test.ts
import { describe, it, expect } from "vitest";
import { normalizeLabel } from "../labels";

describe("normalizeLabel", () => {
  it("lowercases and trims", () => {
    expect(normalizeLabel("  Prod  ")).toBe("prod");
  });
  it("collapses whitespace runs to a single dash", () => {
    expect(normalizeLabel("payments  team core")).toBe("payments-team-core");
  });
  it("returns empty string for whitespace-only input", () => {
    expect(normalizeLabel("   ")).toBe("");
  });
});
```

- [ ] **Step 2: Run to verify it fails, then implement**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/tests/labels.test.ts` → FAIL (module missing). Then:

```ts
// frontend/src/pages/stacks/lib/labels.ts

/** Design rule: labels are lowercase, whitespace becomes dashes. */
export function normalizeLabel(raw: string): string {
  return raw.trim().toLowerCase().replace(/\s+/g, "-");
}
```

Re-run → PASS.

- [ ] **Step 3: Write failing stackToUpdateRequest test**

Add to `frontend/src/pages/stacks/lib/draft-sync/tests/snapshot-to-update.test.ts` (create the file with vitest imports if it doesn't exist; if it exists, append the describe block):

```ts
import { describe, it, expect } from "vitest";
import { stackToUpdateRequest } from "../snapshot-to-update";
import type { Stack } from "@/api/stacks";

describe("stackToUpdateRequest", () => {
  const stack = {
    id: "s1",
    name: "acme",
    labels: [{ key: "user", value: "old" }],
    spec: {
      stack_resources: [
        {
          id: "r1",
          name: "web",
          revision: "3",
          status: { state: "Ready" },
          outputs: [{ name: "url" }],
          ports: [{ number: 80, exposed_to_public: true }],
        },
      ],
      volumes: [{ name: "data", spec: { size: "1Gi" } }],
      connections: [{ id: "c1", kind: "volume_mount" }],
    },
  } as unknown as Stack;

  it("carries the full spec INCLUDING connections (PUT is replace-all)", () => {
    const req = stackToUpdateRequest(stack, [{ key: "user", value: "prod" }]);
    expect(req.spec?.connections).toEqual([{ id: "c1", kind: "volume_mount" }]);
    expect(req.spec?.stack_resources).toHaveLength(1);
    expect(req.spec?.volumes).toHaveLength(1);
  });

  it("applies the new labels and keeps the name", () => {
    const req = stackToUpdateRequest(stack, [{ key: "user", value: "prod" }]);
    expect(req.name).toBe("acme");
    expect(req.labels).toEqual([{ key: "user", value: "prod" }]);
  });

  it("strips server-only resource fields (id/revision/status/outputs)", () => {
    const req = stackToUpdateRequest(stack, []);
    const r = req.spec!.stack_resources![0] as Record<string, unknown>;
    expect(r.id).toBeUndefined();
    expect(r.revision).toBeUndefined();
    expect(r.status).toBeUndefined();
    expect(r.outputs).toBeUndefined();
  });
});
```

- [ ] **Step 4: Run to verify it fails, then implement**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/draft-sync/tests/snapshot-to-update.test.ts` → FAIL (`stackToUpdateRequest` not exported). Then add to `snapshot-to-update.ts` (below `snapshotToUpdateRequest`, reusing its private helpers):

```ts
/**
 * Whole-stack PUT body from the LIVE server stack with only the labels swapped.
 * PUT replace-all semantics require the full resource/volume/connection set —
 * an incomplete body silently wipes bindings.
 */
export function stackToUpdateRequest(stack: Stack, labels: Stack["labels"]): StackUpdateRequest {
  const connections = stack.spec?.connections ?? [];
  return {
    name: stack.name,
    labels,
    spec: {
      stack_resources: (stack.spec?.stack_resources ?? []).map((r) =>
        withExplicitEmptyCollections(cleanServerResource(r as StackResource)),
      ),
      volumes:
        (stack.spec?.volumes ?? []).length > 0
          ? (stack.spec?.volumes ?? []).map((v) => cleanVolume(v as Volume))
          : undefined,
      ...(connections.length > 0 ? { connections } : {}),
    },
  };
}
```

(Import `Stack` type if not already imported in the file.) Re-run → PASS.

- [ ] **Step 5: Wire detail page**

In `detail/index.tsx`:

1. Imports: `updateStack` from `@/api/stacks`, `stackToUpdateRequest` from `@/pages/stacks/lib/draft-sync/snapshot-to-update`, `normalizeLabel` from `@/pages/stacks/lib/labels`, `SYNC_STATUS` + `SyncStatus` type (already imported for the shell — verify).
2. Local state near `draftLabels` (line ~88): `const [labelSync, setLabelSync] = useState<SyncStatus>(SYNC_STATUS.idle);`
3. Handlers (replace the existing `addDraftLabel`/`removeDraftLabel` block at lines 465–470):

```tsx
  const addDraftLabel = useCallback((value: string) => {
    const normalized = normalizeLabel(value);
    if (!normalized) return;
    setDraftLabels((prev) => {
      const cur = prev ?? [];
      if (cur.some((l) => l.value === normalized)) return cur;
      return [...cur, { key: USER_DEFINED_LABEL_KEY, value: normalized }];
    });
  }, []);
  const removeDraftLabel = useCallback((idx: number) => {
    setDraftLabels((prev) => (prev ?? []).filter((_, i) => i !== idx));
  }, []);

  // Deployed stacks: persist the new label set immediately via a full PUT
  // (replace-all body built from the live server stack, labels swapped).
  const persistLabels = useCallback(
    async (next: NonNullable<Stack["labels"]>) => {
      if (!stackToShow?.id || !deployIds) return;
      setLabelSync(SYNC_STATUS.saving);
      try {
        const fresh = await updateStack(
          deployIds.orgId,
          deployIds.teamName,
          deployIds.stackId,
          stackToUpdateRequest(stackToShow, next),
        );
        setFetchedStack(fresh);
        setStacks(stacks.map((s) => (s.id === fresh.id ? fresh : s)));
        setLabelSync(SYNC_STATUS.saved);
        setTimeout(() => setLabelSync(SYNC_STATUS.idle), 2000);
      } catch {
        setLabelSync(SYNC_STATUS.error);
      }
    },
    [stackToShow, deployIds, setStacks, stacks],
  );

  const addStackLabel = useCallback(
    (value: string) => {
      const normalized = normalizeLabel(value);
      if (!normalized) return;
      const cur = stackToShow?.labels ?? [];
      if (cur.some((l) => l.value === normalized)) return;
      void persistLabels([...cur, { key: USER_DEFINED_LABEL_KEY, value: normalized }]);
    },
    [stackToShow, persistLabels],
  );
  const removeStackLabel = useCallback(
    (idx: number) => {
      const cur = stackToShow?.labels ?? [];
      void persistLabels(cur.filter((_, i) => i !== idx));
    },
    [stackToShow, persistLabels],
  );
```

(Adjust `deployIds` guard to the actual shape at line ~250 — it's an object with `orgId`/`teamName`/`stackId`; gate on `deployIds.stackId` like `useDraftSync` does.)

4. Shell props (lines 564–567) become:

```tsx
        labels={(isDraft ? draftLabels : effectiveStack?.labels) ?? []}
        labelsEditable={isDraft || canWriteStack}
        onAddLabel={isDraft ? addDraftLabel : addStackLabel}
        onRemoveLabel={isDraft ? removeDraftLabel : removeStackLabel}
```

5. Autosave indicator surfaces label saves: `syncStatus={isDraft ? SYNC_STATUS.idle : labelSync !== SYNC_STATUS.idle ? labelSync : draftSync.status}`.

Known limitation (accepted in spec review): a label PUT built from the live server stack can interleave with in-flight per-resource autosave ops; both are last-write-wins per entity and labels don't touch resources, so drift is bounded to the debounce window.

- [ ] **Step 6: Verify**

Run:
```
pnpm --prefix frontend test -- --run src/pages/stacks
pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```
Expected: PASS/clean. Manual: on a deployed stack add label "  Payments Core " → chip `payments-core` appears, autosave indicator flashes saving→saved; duplicate add is a no-op; remove works; refresh keeps labels; **stack connections/bindings intact after label edit** (check a service with a volume mount still shows it).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/lib/labels.ts frontend/src/pages/stacks/lib/tests/labels.test.ts frontend/src/pages/stacks/lib/draft-sync/snapshot-to-update.ts frontend/src/pages/stacks/lib/draft-sync/tests/snapshot-to-update.test.ts frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(stacks): normalized stack labels editable on deployed stacks"
```

---

### Task 8: Endpoint URL heuristic + org domains hook

**Files:**
- Create: `frontend/src/pages/stacks/lib/public-endpoints.ts`
- Test: `frontend/src/pages/stacks/lib/tests/public-endpoints.test.ts`
- Create: `frontend/src/hooks/use-org-domains.ts`

**Interfaces:**
- Consumes: `getOrganization` (`frontend/src/api/organizations.ts:7`), generated `Ingress` type (`url`, `target_port`).
- Produces (used by Task 9):

```ts
export type UrlClass = "custom" | "prefix" | "generated";
export function classifyIngressUrl(url: string, orgDomains: string[]): UrlClass
export function pickBestIngress(
  ingresses: { url?: string; target_port?: number }[],
  orgDomains: string[],
): { url: string; target_port?: number } | null
export function useOrgDomains(orgId: string | undefined): string[]
```

Generated-host rule: first DNS label matches `/^[a-z2-7]{16}$/` (lowercased no-pad std base32 of an md5, per `pkg/services/exposed_port_domain.go:32-44`).

- [ ] **Step 1: Write the failing test**

```ts
// frontend/src/pages/stacks/lib/tests/public-endpoints.test.ts
import { describe, it, expect } from "vitest";
import { classifyIngressUrl, pickBestIngress } from "../public-endpoints";

const ORG = ["acme.stackdome.app"];
const GENERATED = "https://k7x2m9qp4rt8w3ab.web.acme.stackdome.app";
const PREFIX = "https://web.acme.stackdome.app";
const CUSTOM = "https://app.mycompany.com";

describe("classifyIngressUrl", () => {
  it("host outside org domains → custom", () => {
    expect(classifyIngressUrl(CUSTOM, ORG)).toBe("custom");
  });
  it("16-char base32 first label under org domain → generated", () => {
    expect(classifyIngressUrl(GENERATED, ORG)).toBe("generated");
  });
  it("friendly first label under org domain → prefix", () => {
    expect(classifyIngressUrl(PREFIX, ORG)).toBe("prefix");
  });
  it("unparseable url → custom (shown as-is, never hidden)", () => {
    expect(classifyIngressUrl("not a url", ORG)).toBe("custom");
  });
});

describe("pickBestIngress", () => {
  it("prefers custom > prefix > generated", () => {
    const ingresses = [
      { url: GENERATED, target_port: 80 },
      { url: PREFIX, target_port: 80 },
      { url: CUSTOM, target_port: 80 },
    ];
    expect(pickBestIngress(ingresses, ORG)?.url).toBe(CUSTOM);
    expect(pickBestIngress(ingresses.slice(0, 2), ORG)?.url).toBe(PREFIX);
    expect(pickBestIngress(ingresses.slice(0, 1), ORG)?.url).toBe(GENERATED);
  });
  it("breaks ties by array order and skips url-less entries", () => {
    expect(pickBestIngress([{ target_port: 1 }, { url: PREFIX }, { url: "https://web2.acme.stackdome.app" }], ORG)?.url).toBe(PREFIX);
  });
  it("returns null when nothing has a url", () => {
    expect(pickBestIngress([{}, { target_port: 80 }], ORG)).toBeNull();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/tests/public-endpoints.test.ts`
Expected: FAIL — module missing.

- [ ] **Step 3: Implement**

```ts
// frontend/src/pages/stacks/lib/public-endpoints.ts

/** Matches EncodeStackResourceSubdomainPrefix output: 16-char lowercase no-pad std base32. */
const GENERATED_LABEL = /^[a-z2-7]{16}$/;

export type UrlClass = "custom" | "prefix" | "generated";

export function classifyIngressUrl(url: string, orgDomains: string[]): UrlClass {
  let host: string;
  try {
    host = new URL(url).hostname;
  } catch {
    return "custom"; // show unparseable urls as-is rather than hide them
  }
  const under = orgDomains.find((d) => host === d || host.endsWith(`.${d}`));
  if (!under) return "custom";
  const firstLabel = host.slice(0, host.length - under.length).replace(/\.$/, "").split(".")[0] ?? "";
  return GENERATED_LABEL.test(firstLabel) ? "generated" : "prefix";
}

const CLASS_RANK: Record<UrlClass, number> = { custom: 0, prefix: 1, generated: 2 };

/** Best URL for the header pill: custom domain > subdomain-prefix > generated hash. */
export function pickBestIngress(
  ingresses: { url?: string; target_port?: number }[],
  orgDomains: string[],
): { url: string; target_port?: number } | null {
  let best: { url: string; target_port?: number } | null = null;
  let bestRank = Number.POSITIVE_INFINITY;
  for (const ing of ingresses) {
    if (!ing.url) continue;
    const rank = CLASS_RANK[classifyIngressUrl(ing.url, orgDomains)];
    if (rank < bestRank) {
      best = { url: ing.url, target_port: ing.target_port };
      bestRank = rank;
    }
  }
  return best;
}
```

```ts
// frontend/src/hooks/use-org-domains.ts
import { useEffect, useState } from "react";
import { getOrganization } from "@/api/organizations";

/** FQDNs owned by the organisation — used to classify public ingress URLs. */
export function useOrgDomains(orgId: string | undefined): string[] {
  const [domains, setDomains] = useState<string[]>([]);
  useEffect(() => {
    if (!orgId) return;
    let cancelled = false;
    getOrganization(orgId)
      .then((org) => {
        if (cancelled) return;
        setDomains((org.domains ?? []).map((d) => d.fqdn).filter((f): f is string => !!f));
      })
      .catch(() => {
        // Classification degrades gracefully: with no org domains every URL
        // reads as "custom", which still renders correctly in the pill.
      });
    return () => {
      cancelled = true;
    };
  }, [orgId]);
  return domains;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/lib/tests/public-endpoints.test.ts`
Expected: PASS (7 tests). Also `pnpm --prefix frontend exec tsc -b`.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/public-endpoints.ts frontend/src/pages/stacks/lib/tests/public-endpoints.test.ts frontend/src/hooks/use-org-domains.ts
git commit -m "feat(stacks): public endpoint URL classification and org domains hook"
```

---

### Task 9: PublicEndpointRow + header wiring

**Files:**
- Create: `frontend/src/pages/stacks/components/canvas/PublicEndpointRow.tsx`
- Test: `frontend/src/pages/stacks/components/canvas/tests/public-endpoint-row.test.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx` (render row when expanded)
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (compute endpoints)

**Interfaces:**
- Consumes: Task 8 `pickBestIngress` + `useOrgDomains`.
- Produces:

```ts
export interface PublicEndpoint { service: string; url: string; port?: number; }
// CanvasEditorShellProps gains: publicEndpoints?: PublicEndpoint[]
```

- [ ] **Step 1: Write the failing test**

```tsx
// frontend/src/pages/stacks/components/canvas/tests/public-endpoint-row.test.tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, fireEvent, cleanup, act } from "@testing-library/react";
import { PublicEndpointRow } from "../PublicEndpointRow";

afterEach(cleanup);

const endpoints = [
  { service: "web", url: "https://web.acme.stackdome.app", port: 80 },
  { service: "api", url: "https://api.mycompany.com", port: 8080 },
];

describe("PublicEndpointRow", () => {
  beforeEach(() => {
    Object.assign(navigator, { clipboard: { writeText: vi.fn().mockResolvedValue(undefined) } });
  });

  it("renders one pill per endpoint with service chip and host", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    expect(screen.getByText("PUBLIC")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("web.acme.stackdome.app")).toBeInTheDocument();
    expect(screen.getByText("api.mycompany.com")).toBeInTheDocument();
  });

  it("link opens in a new tab", () => {
    render(<PublicEndpointRow endpoints={endpoints} />);
    const link = screen.getByRole("link", { name: /web\.acme\.stackdome\.app/ });
    expect(link).toHaveAttribute("href", "https://web.acme.stackdome.app");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noreferrer");
  });

  it("copy button writes the url and flashes a check", async () => {
    vi.useFakeTimers();
    render(<PublicEndpointRow endpoints={[endpoints[0]]} />);
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Copy https://web.acme.stackdome.app" }));
    });
    expect(navigator.clipboard.writeText).toHaveBeenCalledWith("https://web.acme.stackdome.app");
    expect(screen.getByRole("button", { name: "Copied" })).toBeInTheDocument();
    act(() => vi.advanceTimersByTime(1400));
    expect(screen.getByRole("button", { name: "Copy https://web.acme.stackdome.app" })).toBeInTheDocument();
    vi.useRealTimers();
  });

  it("renders nothing without endpoints", () => {
    const { container } = render(<PublicEndpointRow endpoints={[]} />);
    expect(container).toBeEmptyDOMElement();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/public-endpoint-row.test.tsx`
Expected: FAIL — module missing.

- [ ] **Step 3: Implement PublicEndpointRow**

```tsx
// frontend/src/pages/stacks/components/canvas/PublicEndpointRow.tsx
import { useCallback, useEffect, useRef, useState } from "react";
import { Globe, ExternalLink, Copy, Check } from "lucide-react";
import { Tooltip, TooltipTrigger, TooltipContent } from "@/components/ui/tooltip";

const COPY_FLASH_MS = 1400;

export interface PublicEndpoint {
  service: string;
  url: string;
  port?: number;
}

function hostOf(url: string): string {
  try {
    return new URL(url).hostname;
  } catch {
    return url;
  }
}

async function copyText(text: string): Promise<void> {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  // Clipboard API unavailable (insecure context): textarea fallback.
  const ta = document.createElement("textarea");
  ta.value = text;
  document.body.appendChild(ta);
  ta.select();
  document.execCommand("copy");
  ta.remove();
}

/** Header row mapping each publicly exposed service to its best live URL. */
export function PublicEndpointRow({ endpoints }: { endpoints: PublicEndpoint[] }) {
  const [copiedUrl, setCopiedUrl] = useState<string | null>(null);
  const timer = useRef<ReturnType<typeof setTimeout>>(null);
  useEffect(() => () => { if (timer.current) clearTimeout(timer.current); }, []);

  const onCopy = useCallback((url: string) => {
    void copyText(url).then(() => {
      setCopiedUrl(url);
      if (timer.current) clearTimeout(timer.current);
      timer.current = setTimeout(() => setCopiedUrl(null), COPY_FLASH_MS);
    });
  }, []);

  if (endpoints.length === 0) return null;

  return (
    <div className="mt-2.5 flex flex-wrap items-center gap-2">
      <span className="font-mono text-[9.5px] uppercase tracking-[0.14em] text-fg-muted">PUBLIC</span>
      {endpoints.map(({ service, url, port }) => (
        <span
          key={`${service}-${url}`}
          className="inline-flex items-stretch overflow-hidden rounded-md border border-border bg-muted/40 text-[12px]"
        >
          <Tooltip delayDuration={300}>
            <TooltipTrigger asChild>
              <span className="flex items-center gap-1.5 border-r border-border px-2 py-1 font-mono text-fg-muted">
                <Globe className="size-3" />
                {service}
              </span>
            </TooltipTrigger>
            <TooltipContent side="top">
              Mapped to {service}{port != null ? ` · :${port}` : ""}
            </TooltipContent>
          </Tooltip>
          <a
            href={url}
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-1.5 px-2 py-1 font-mono text-foreground hover:bg-muted"
          >
            <span aria-hidden className="size-[5px] rounded-full bg-success" />
            {hostOf(url)}
            <ExternalLink className="size-3 text-fg-muted" />
          </a>
          <button
            type="button"
            onClick={() => onCopy(url)}
            aria-label={copiedUrl === url ? "Copied" : `Copy ${url}`}
            className="flex items-center border-l border-border px-1.5 text-fg-muted hover:bg-muted hover:text-foreground"
          >
            {copiedUrl === url ? <Check className="size-3 text-success" /> : <Copy className="size-3" />}
          </button>
        </span>
      ))}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- --run src/pages/stacks/components/canvas/tests/public-endpoint-row.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 5: Wire shell + detail page**

1. `CanvasEditorShell.tsx`: add `publicEndpoints?: PublicEndpoint[]` to props (import type from `./PublicEndpointRow`). In the **expanded** header only, render after the labels block (inside the same `px-7` container):

```tsx
        <PublicEndpointRow endpoints={publicEndpoints ?? []} />
```

2. `detail/index.tsx`: compute endpoints from the live stack + org domains:

```tsx
  const orgDomains = useOrgDomains(effectiveStack?.organisation_id ?? getCurrentOrganizationId() ?? undefined);
  const publicEndpoints = useMemo(() => {
    if (isDraft) return [];
    return (effectiveStack?.spec.stack_resources ?? []).flatMap((r) => {
      const best = pickBestIngress(r.status?.public_ingress ?? [], orgDomains);
      return best && r.name ? [{ service: r.name, url: best.url, port: best.target_port }] : [];
    });
  }, [isDraft, effectiveStack, orgDomains]);
```

(Imports: `useOrgDomains` from `@/hooks/use-org-domains`, `pickBestIngress` from `@/pages/stacks/lib/public-endpoints`. If the resource type's `status` isn't typed with `public_ingress`, cast via the generated `StackResourceStatus` type — do NOT use `any`.) Pass `publicEndpoints={publicEndpoints}` to the shell.

- [ ] **Step 6: Verify**

Run:
```
pnpm --prefix frontend test -- --run src/pages/stacks
pnpm --prefix frontend exec tsc -b && pnpm --prefix frontend lint
```
Expected: PASS/clean. Manual: deployed stack with exposed port shows `PUBLIC [service] [dot host ↗] [copy]`; copy flashes check; row absent on drafts and collapsed header.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/PublicEndpointRow.tsx frontend/src/pages/stacks/components/canvas/tests/public-endpoint-row.test.tsx frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(stacks): PUBLIC endpoint row mapping services to live ingress URLs"
```

---

### Task 10: Full verification pass

**Files:** none new.

- [ ] **Step 1: Full frontend gate**

Run:
```
pnpm --prefix frontend test -- --run
pnpm --prefix frontend exec tsc -b
pnpm --prefix frontend lint
```
Expected: all PASS/clean. Fix anything that fails before proceeding.

- [ ] **Step 2: Manual end-to-end (Playwright MCP against http://localhost:5173)**

With `mage run` + `pnpm --prefix frontend dev` (or the embedded build):

1. Open a deployed stack → header expanded: name, status pill, labels, subtitle, PUBLIC row, tabs.
2. Click service node → floating drawer, canvas doesn't resize; edit a field → autosave indicator cycles.
3. Click mount row's open-volume button → volume drawer pushes; parked service panel dimmed at 10px/16px stagger; click parked panel → returns; Esc pops; Shift+Esc closes all.
4. Collapse header via chevron → 44px bar with compact tabs + Deploy; ⌘. toggles; reload → collapsed state persists.
5. Add label `Payments Core` on the deployed stack → chip `payments-core`; reload → persists; **verify volume mounts/addon bindings survived** (PUT replace-all regression check).
6. PUBLIC pill: link opens new tab, copy flashes check; drafts show no row.

- [ ] **Step 3: Commit any fixes; final commit**

```bash
git add -A && git commit -m "fix(stacks): canvas editor polish from verification pass" # only if fixes were needed
```

---

## Self-Review Notes

- Spec §1 drawers → Tasks 1–5; §2 header → Task 6; §3 labels → Task 7; §4 PUBLIC row → Tasks 8–9; testing/error handling folded into each task + Task 10.
- Deviation from spec: drawer width uses design's 600px (spec matches); behind-panel scrim uses `bg-background/55` token instead of raw rgba per Global Constraints.
- The `void openVolume;` placeholder in Task 3 is intentional wiring order (removed in Task 5).
