# Canvas Volume Operations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Volume lifecycle management on the stack canvas — create-attached from the add button, drag/context-menu attach for floating volumes, right-click disconnect/delete, drawer mount-editing retired.

**Architecture:** All gestures are pure calculations over the edit-session draft (`lib/canvas/volume-ops.ts`), applied through the existing `session.updateResources` / `session.updateVolumes` setters. Persistence rides the existing draft-sync pipeline (`mountsToConnections`, per-entity CRUD) — no new persistence code. Canvas gestures wire through React Flow v12 handlers (`onNodeContextMenu`, `onNodeDrag*`, `getIntersectingNodes`).

**Tech Stack:** React 19, React Flow (`@xyflow/react` v12), Radix (`ui/dialog`, `ui/dropdown-menu`, `ui/alert-dialog`, `ui/select`), Tailwind v4 semantic tokens, vitest + @testing-library/react.

**Spec:** `docs/superpowers/specs/2026-07-04-canvas-volume-ops-design.md`

## Global Constraints

- **Cardinality:** 1 volume : 1 resource, enforced in UI. Multiple volumes per resource allowed.
- **Volume name immutable after create** — it is the identity key for `alignBaselineToDraft`.
- **No new persistence code** — every mutation goes through `session.updateResources` / `session.updateVolumes`.
- **No magic strings** — use `NODE_KIND`, `NODE_ID_PREFIX`, existing constants. Define new constants where needed.
- **Brand design system only** — semantic tokens (`text-foreground`, `bg-card`, `border-border`, `text-danger`, `bg-danger-bg`, `text-fg-muted`, `border-brand`), existing `ui/` + `branded/` primitives. No raw hex.
- **Mount row shape** is `{ source_volume_name, source_sub_path, target_path }` (`FormStackResourceData["volume_mounts"][number]`). There is no `read_only` field in the form model — the attach dialog offers Target path (required) + Sub path (optional).
- **Tests:** run with `pnpm exec vitest run <path>` from `frontend/` (vitest needs the explicit `run` flag).
- **All file paths below are relative to `frontend/`** unless prefixed with `docs/`.

---

### Task 1: Pure volume operations (`volume-ops.ts`)

**Files:**
- Create: `src/pages/stacks/lib/canvas/volume-ops.ts`
- Test: `src/pages/stacks/lib/canvas/tests/volume-ops.test.ts`

**Interfaces:**
- Consumes: `ResourceArr`, `VolumeArr` from `@/pages/stacks/lib/stack-diff`; volume literal shape from `lib/canvas/inline-volume.ts` (`addInlineVolume`).
- Produces (used by Tasks 4–8):
  - `suggestVolumeName(volumes: VolumeArr): string`
  - `newVolume(input: { name: string; size: string }): VolumeArr[number]`
  - `addMount(resources: ResourceArr, resourceIdx: number, mount: { volumeName: string; targetPath: string; subPath?: string }): ResourceArr`
  - `removeMountsOf(resources: ResourceArr, volumeName: string): ResourceArr`
  - `mountOwner(resources: ResourceArr, volumeName: string): { resourceIdx: number; resourceName: string; targetPath: string } | null`
  - `validateVolumeName(name: string, volumes: VolumeArr): string | undefined`
  - `validateTargetPath(path: string, resource: ResourceArr[number] | undefined): string | undefined`

- [ ] **Step 1: Write the failing tests**

```ts
// src/pages/stacks/lib/canvas/tests/volume-ops.test.ts
import { describe, expect, it } from "vitest";
import {
  addMount,
  mountOwner,
  newVolume,
  removeMountsOf,
  suggestVolumeName,
  validateTargetPath,
  validateVolumeName,
} from "../volume-ops";

const vol = (name: string) => ({ name, spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } });
const res = (name: string, mounts: Array<{ source_volume_name: string; target_path: string }> = []) => ({
  name,
  volume_mounts: mounts.map((m) => ({ ...m, source_sub_path: "" })),
});

describe("suggestVolumeName", () => {
  it("starts at 'volume' and skips taken names", () => {
    expect(suggestVolumeName([])).toBe("volume");
    expect(suggestVolumeName([vol("volume")])).toBe("volume-2");
    expect(suggestVolumeName([vol("volume"), vol("volume-2")])).toBe("volume-3");
  });
});

describe("newVolume", () => {
  it("builds the extended-form volume literal", () => {
    const v = newVolume({ name: "data", size: "2Gi" }) as Record<string, unknown>;
    expect(v.name).toBe("data");
    expect(v.sourceType).toBe("None");
    expect((v.spec as Record<string, unknown>).size).toBe("2Gi");
    expect((v.spec as Record<string, unknown>).access_mode).toBe("ReadWriteOnce");
  });
});

describe("addMount / removeMountsOf / mountOwner", () => {
  it("appends a mount to the target resource only", () => {
    const next = addMount([res("web"), res("api")], 1, { volumeName: "data", targetPath: "/var/data" });
    expect(next[0].volume_mounts).toEqual([]);
    expect(next[1].volume_mounts).toEqual([
      { source_volume_name: "data", source_sub_path: "", target_path: "/var/data" },
    ]);
  });

  it("removeMountsOf strips the volume's mounts from every resource", () => {
    const resources = [
      res("web", [{ source_volume_name: "data", target_path: "/a" }]),
      res("api", [
        { source_volume_name: "data", target_path: "/b" },
        { source_volume_name: "other", target_path: "/c" },
      ]),
    ];
    const next = removeMountsOf(resources, "data");
    expect(next[0].volume_mounts).toEqual([]);
    expect(next[1].volume_mounts).toEqual([{ source_volume_name: "other", source_sub_path: "", target_path: "/c" }]);
  });

  it("mountOwner returns the first mounting resource, or null when unmounted", () => {
    const resources = [res("web"), res("api", [{ source_volume_name: "data", target_path: "/d" }])];
    expect(mountOwner(resources, "data")).toEqual({ resourceIdx: 1, resourceName: "api", targetPath: "/d" });
    expect(mountOwner(resources, "ghost")).toBeNull();
  });
});

describe("validateVolumeName", () => {
  it("requires a value and rejects duplicates", () => {
    expect(validateVolumeName("", [])).toBeTruthy();
    expect(validateVolumeName("data", [vol("data")])).toBeTruthy();
    expect(validateVolumeName("fresh", [vol("data")])).toBeUndefined();
  });
});

describe("validateTargetPath", () => {
  it("requires an absolute path", () => {
    expect(validateTargetPath("", res("web"))).toBeTruthy();
    expect(validateTargetPath("data", res("web"))).toBeTruthy();
    expect(validateTargetPath("/data", res("web"))).toBeUndefined();
  });
  it("rejects a target path already mounted on the resource", () => {
    const r = res("web", [{ source_volume_name: "data", target_path: "/data" }]);
    expect(validateTargetPath("/data", r)).toBeTruthy();
    expect(validateTargetPath("/other", r)).toBeUndefined();
  });
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/canvas/tests/volume-ops.test.ts`
Expected: FAIL — `Cannot find module '../volume-ops'`

- [ ] **Step 3: Write the implementation**

```ts
// src/pages/stacks/lib/canvas/volume-ops.ts
import type { ResourceArr, VolumeArr } from "@/pages/stacks/lib/stack-diff";

type Resource = ResourceArr[number];
type Mount = NonNullable<Resource["volume_mounts"]>[number];

const BASE_NAME = "volume";

/** Pure calculations for canvas volume gestures. No React, unit-testable. */

export function suggestVolumeName(volumes: VolumeArr): string {
  const taken = new Set(volumes.map((v) => v.name).filter(Boolean));
  if (!taken.has(BASE_NAME)) return BASE_NAME;
  for (let i = 2; ; i++) {
    const candidate = `${BASE_NAME}-${i}`;
    if (!taken.has(candidate)) return candidate;
  }
}

/** Volume literal in the extended form shape (mirrors addInlineVolume's). */
export function newVolume(input: { name: string; size: string }): VolumeArr[number] {
  return {
    name: input.name,
    sourceType: "None",
    labels: [],
    spec: { size: input.size, access_mode: "ReadWriteOnce", needs_sync_before_use: false },
  } as unknown as VolumeArr[number];
}

export function addMount(
  resources: ResourceArr,
  resourceIdx: number,
  mount: { volumeName: string; targetPath: string; subPath?: string },
): ResourceArr {
  return resources.map((r, i) =>
    i === resourceIdx
      ? {
          ...r,
          volume_mounts: [
            ...(r.volume_mounts ?? []),
            {
              source_volume_name: mount.volumeName,
              source_sub_path: mount.subPath ?? "",
              target_path: mount.targetPath,
            } as Mount,
          ],
        }
      : r,
  );
}

export function removeMountsOf(resources: ResourceArr, volumeName: string): ResourceArr {
  return resources.map((r) =>
    (r.volume_mounts ?? []).some((m) => m.source_volume_name === volumeName)
      ? { ...r, volume_mounts: (r.volume_mounts ?? []).filter((m) => m.source_volume_name !== volumeName) }
      : r,
  );
}

export function mountOwner(
  resources: ResourceArr,
  volumeName: string,
): { resourceIdx: number; resourceName: string; targetPath: string } | null {
  for (let i = 0; i < resources.length; i++) {
    const m = (resources[i].volume_mounts ?? []).find((vm) => vm.source_volume_name === volumeName);
    if (m) return { resourceIdx: i, resourceName: resources[i].name ?? "", targetPath: m.target_path };
  }
  return null;
}

export function validateVolumeName(name: string, volumes: VolumeArr): string | undefined {
  if (!name.trim()) return "Required";
  if (volumes.some((v) => v.name === name)) return "Volume name must be unique";
  return undefined;
}

export function validateTargetPath(
  path: string,
  resource: Resource | undefined,
): string | undefined {
  if (!path.trim()) return "Required";
  if (!path.startsWith("/")) return "Must be an absolute path (start with /)";
  if ((resource?.volume_mounts ?? []).some((m) => m.target_path === path)) {
    return "This path is already mounted on the service";
  }
  return undefined;
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/canvas/tests/volume-ops.test.ts`
Expected: PASS (all tests)

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/volume-ops.ts frontend/src/pages/stacks/lib/canvas/tests/volume-ops.test.ts
git commit -m "feat(stacks): add pure volume-op calculations for canvas gestures"
```

---

### Task 2: Guard `revertResource` against dangling mounts

**Files:**
- Modify: `src/pages/stacks/lib/stack-diff.ts:289-311` (`revertResource`)
- Test: `src/pages/stacks/lib/tests/stack-diff.test.ts` (append)

**Interfaces:**
- Consumes: existing `revertResource(draft, baseline, idx)` — signature unchanged.
- Produces: same function; restored resources never carry mounts referencing volumes absent from `draft.volumes`.

- [ ] **Step 1: Write the failing test** (append to `stack-diff.test.ts`)

```ts
describe("revertResource dangling-mount guard", () => {
  it("drops restored mounts whose volume no longer exists in the draft", () => {
    const baseline = {
      resources: [
        { name: "web", volume_mounts: [{ source_volume_name: "data", source_sub_path: "", target_path: "/data" }] },
      ],
      volumes: [{ name: "data" }],
    };
    // Draft deleted the volume (cascade already removed the mount) and edited the resource.
    const draft = {
      resources: [{ name: "web", image_spec: { image: "nginx:2" }, volume_mounts: [] }],
      volumes: [],
    };
    const next = revertResource(draft, baseline, 0);
    expect(next.resources[0].volume_mounts).toEqual([]);
  });

  it("keeps restored mounts whose volume still exists", () => {
    const baseline = {
      resources: [
        { name: "web", volume_mounts: [{ source_volume_name: "data", source_sub_path: "", target_path: "/data" }] },
      ],
      volumes: [{ name: "data" }],
    };
    const draft = {
      resources: [{ name: "web", volume_mounts: [] }],
      volumes: [{ name: "data" }],
    };
    const next = revertResource(draft, baseline, 0);
    expect(next.resources[0].volume_mounts).toEqual([
      { source_volume_name: "data", source_sub_path: "", target_path: "/data" },
    ]);
  });
});
```

Import `revertResource` in the test file's existing import from `../stack-diff` if not already imported.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/tests/stack-diff.test.ts`
Expected: FAIL — first test gets `volume_mounts: [{ source_volume_name: "data", ... }]` instead of `[]`

- [ ] **Step 3: Implement the guard**

In `revertResource`, replace the `baselineEntry != null` branch body:

```ts
  if (baselineEntry != null) {
    // Keep the draft's live status: the baseline's was captured at deploy time
    // and restoring it would show stale telemetry until the next refresh.
    const liveStatus = (draft.resources[idx] as { status?: unknown } | undefined)?.status;
    const restored = {
      ...cloneJson(baselineEntry),
      ...(liveStatus !== undefined ? { status: liveStatus } : {}),
    } as (typeof next.resources)[number];
    // A restored mount may reference a volume the draft has since deleted —
    // reattaching it would create a dangling mount, so drop those rows.
    if (Array.isArray(restored.volume_mounts)) {
      const draftVolumeNames = new Set(draft.volumes.map((v) => v?.name).filter(Boolean));
      restored.volume_mounts = restored.volume_mounts.filter(
        (m) => m?.source_volume_name && draftVolumeNames.has(m.source_volume_name),
      );
    }
    next.resources[idx] = restored;
  } else {
```

- [ ] **Step 4: Run the full stack-diff suite**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/tests/stack-diff.test.ts`
Expected: PASS (new + all pre-existing tests)

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/stack-diff.ts frontend/src/pages/stacks/lib/tests/stack-diff.test.ts
git commit -m "fix(stacks): drop dangling volume mounts when reverting a resource"
```

---

### Task 3: Dangling-mount filters in desired-state and graph chips

**Files:**
- Modify: `src/pages/stacks/lib/draft-sync/desired-state.ts:72-77`
- Modify: `src/pages/stacks/lib/canvas/graph-from-connections.ts:168-173, 216-236`
- Test: `src/pages/stacks/lib/draft-sync/tests/desired-state.test.ts` (append)
- Test: `src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts` (append)

**Interfaces:**
- Consumes: `buildDesiredState(draft)`, `deriveGraph(input)` — signatures unchanged.
- Produces: desired connections never include `volume_mount` rows for volumes missing from the draft; resource-card chips only render volumes present in `volumeNames`.

- [ ] **Step 1: Write the failing tests**

Append to `desired-state.test.ts`:

```ts
it("skips volume_mount connections whose volume is missing from the draft", () => {
  const draft = {
    resources: [
      {
        name: "web",
        image_spec: { image: "nginx" },
        volume_mounts: [
          { source_volume_name: "ghost", source_sub_path: "", target_path: "/g" },
          { source_volume_name: "data", source_sub_path: "", target_path: "/d" },
        ],
      },
    ],
    volumes: [{ name: "data", spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } }],
  };
  const state = buildDesiredState(draft as never);
  const mountConns = [...state.connections.values()].filter((c) => c.kind === "volume_mount");
  expect(mountConns).toHaveLength(1);
  expect(mountConns[0].from?.name).toBe("data");
});
```

(Match the resource literal shape used by the file's existing tests — if existing tests build resources through a helper, reuse it; the resource must pass `FormStackResourceSchema.safeParse`.)

Append to `graph-from-connections.test.ts`:

```ts
it("omits chips for mounts whose volume is missing from volumeNames", () => {
  const graph = deriveGraph({
    resources: [
      {
        name: "web",
        volume_mounts: [
          { source_volume_name: "data", source_sub_path: "", target_path: "/d" },
          { source_volume_name: "ghost", source_sub_path: "", target_path: "/g" },
        ],
      },
    ],
    linkedAddonIds: new Set(),
    addonNameById: new Map(),
    secretNameById: new Map(),
    volumeNames: ["data"],
  });
  const web = graph.nodes.find((n) => n.id === "resource:web");
  expect((web?.data as { volumes: { name: string }[] }).volumes.map((v) => v.name)).toEqual(["data"]);
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/draft-sync/tests/desired-state.test.ts src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts`
Expected: FAIL — 2 mount connections / 2 chips where 1 expected

- [ ] **Step 3: Implement both filters**

`desired-state.ts` — the mounts→connections loop becomes:

```ts
  // Add volume mount connections for each valid resource. Mounts referencing
  // volumes absent from the draft are dangling — never emit them.
  for (const { name, mounts } of validForMounts) {
    const liveMounts = mounts.filter((m) => volumes.has((m as { source_volume_name?: string }).source_volume_name ?? ""));
    for (const conn of mountsToConnections(name, liveMounts)) {
      connections.set(connectionIdentityKey(conn), conn);
    }
  }
```

`graph-from-connections.ts` — give `volumeChips` a known-volumes filter:

```ts
function volumeChips(
  resource: Partial<FormStackResourceData>,
  knownVolumes: ReadonlySet<string> | null,
): VolumeChip[] {
  return (resource.volume_mounts ?? [])
    .filter((m) => !knownVolumes || knownVolumes.has((m.source_volume_name ?? "") as string))
    .map((m) => ({
      name: (m.source_volume_name ?? "") as string,
      mountPath: m.target_path as string | undefined,
    }));
}
```

In `deriveGraph`, before the resources loop:

```ts
  // null = caller didn't supply volumeNames (tests, older callers): no filtering.
  const knownVolumes = input.volumeNames ? new Set(input.volumeNames) : null;
```

and change the call site to `volumes: volumeChips(resource, knownVolumes),`.

- [ ] **Step 4: Run both suites plus existing canvas tests**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/lib/draft-sync/tests src/pages/stacks/lib/canvas/tests`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/draft-sync/desired-state.ts frontend/src/pages/stacks/lib/canvas/graph-from-connections.ts frontend/src/pages/stacks/lib/draft-sync/tests/desired-state.test.ts frontend/src/pages/stacks/lib/canvas/tests/graph-from-connections.test.ts
git commit -m "fix(stacks): filter dangling volume mounts from sync and canvas chips"
```

---

### Task 4: AddVolumeDialog + "Volume" entry in the add popover

**Files:**
- Create: `src/pages/stacks/components/canvas/AddVolumeDialog.tsx`
- Modify: `src/pages/stacks/components/canvas/AddResourcePopover.tsx`
- Modify: `src/pages/stacks/components/canvas/CanvasEditor.tsx` (pass-through props)
- Modify: `src/pages/stacks/components/canvas/StackCanvasTab.tsx` (state + handler + render)
- Test: `src/pages/stacks/components/canvas/tests/add-volume-dialog.test.tsx`

**Interfaces:**
- Consumes: Task 1's `suggestVolumeName`, `newVolume`, `addMount`, `validateVolumeName`, `validateTargetPath`.
- Produces:
  - `AddVolumeDialog` props: `{ open: boolean; onOpenChange: (open: boolean) => void; resources: ResourceArr; volumes: VolumeArr; initialResourceIdx?: number | null; onCreate: (input: { name: string; size: string; resourceIdx: number; targetPath: string }) => void }`
  - `AddResourcePopover` new props: `{ canAddVolume: boolean; onAddVolume: () => void }`
  - `CanvasEditor` new props: `{ canAddVolume: boolean; onAddVolume: () => void }`
  - `StackCanvasTab` gains `applyDraft(fn: (draft: { resources; volumes }) => { resources; volumes })` — the session-or-start mutation helper reused by Tasks 5–7.

- [ ] **Step 1: Write the failing component test**

```tsx
// src/pages/stacks/components/canvas/tests/add-volume-dialog.test.tsx
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { AddVolumeDialog } from "../AddVolumeDialog";

const resources = [{ name: "web", volume_mounts: [{ source_volume_name: "old", source_sub_path: "", target_path: "/taken" }] }];
const volumes = [{ name: "old" }];

function renderDialog(onCreate = vi.fn()) {
  render(
    <AddVolumeDialog
      open
      onOpenChange={() => {}}
      resources={resources}
      volumes={volumes}
      initialResourceIdx={0}
      onCreate={onCreate}
    />,
  );
  return onCreate;
}

describe("AddVolumeDialog", () => {
  it("suggests a unique name and creates with valid input", () => {
    const onCreate = renderDialog();
    expect(screen.getByLabelText(/name/i)).toHaveValue("volume");
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(onCreate).toHaveBeenCalledWith({ name: "volume", size: "1Gi", resourceIdx: 0, targetPath: "/data" });
  });

  it("blocks duplicate names", () => {
    const onCreate = renderDialog();
    fireEvent.change(screen.getByLabelText(/name/i), { target: { value: "old" } });
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(onCreate).not.toHaveBeenCalled();
    expect(screen.getByText(/must be unique/i)).toBeInTheDocument();
  });

  it("blocks relative and already-taken mount paths", () => {
    const onCreate = renderDialog();
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "data" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(screen.getByText(/absolute path/i)).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText(/mount path/i), { target: { value: "/taken" } });
    fireEvent.click(screen.getByRole("button", { name: /add volume/i }));
    expect(screen.getByText(/already mounted/i)).toBeInTheDocument();
    expect(onCreate).not.toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/canvas/tests/add-volume-dialog.test.tsx`
Expected: FAIL — `Cannot find module '../AddVolumeDialog'`

- [ ] **Step 3: Implement `AddVolumeDialog`**

```tsx
// src/pages/stacks/components/canvas/AddVolumeDialog.tsx
import { useEffect, useState } from "react";
import { HardDrive } from "lucide-react";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import type { ResourceArr, VolumeArr } from "@/pages/stacks/lib/stack-diff";
import {
  suggestVolumeName,
  validateTargetPath,
  validateVolumeName,
} from "@/pages/stacks/lib/canvas/volume-ops";

const DEFAULT_SIZE = "1Gi";

interface AddVolumeDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  resources: ResourceArr;
  volumes: VolumeArr;
  /** Pre-select a service (resource-card "Add volume…" menu entry). */
  initialResourceIdx?: number | null;
  onCreate: (input: { name: string; size: string; resourceIdx: number; targetPath: string }) => void;
}

/** Create-a-volume dialog: the volume is born attached to the chosen service. */
export function AddVolumeDialog({ open, onOpenChange, resources, volumes, initialResourceIdx, onCreate }: AddVolumeDialogProps) {
  const [name, setName] = useState("");
  const [size, setSize] = useState(DEFAULT_SIZE);
  const [resourceIdx, setResourceIdx] = useState<number | null>(null);
  const [targetPath, setTargetPath] = useState("");
  const [errors, setErrors] = useState<{ name?: string; resource?: string; targetPath?: string }>({});

  // Re-seed the form each time the dialog opens.
  useEffect(() => {
    if (!open) return;
    setName(suggestVolumeName(volumes));
    setSize(DEFAULT_SIZE);
    setResourceIdx(initialResourceIdx ?? (resources.length === 1 ? 0 : null));
    setTargetPath("");
    setErrors({});
    // eslint-disable-next-line react-hooks/exhaustive-deps -- seed on open only
  }, [open]);

  const submit = () => {
    const next = {
      name: validateVolumeName(name, volumes),
      resource: resourceIdx == null ? "Required" : undefined,
      targetPath: validateTargetPath(targetPath, resourceIdx == null ? undefined : resources[resourceIdx]),
    };
    setErrors(next);
    if (next.name || next.resource || next.targetPath) return;
    onCreate({ name, size: size.trim() || DEFAULT_SIZE, resourceIdx: resourceIdx!, targetPath });
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="size-[18px] text-brand" /> Add volume
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          <div className="grid grid-cols-2 gap-4">
            <FieldShell label="Name" htmlFor="add-volume-name" required error={errors.name}>
              <Input
                id="add-volume-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                className={`font-mono ${errors.name ? "border-danger" : ""}`}
                aria-invalid={!!errors.name}
              />
            </FieldShell>
            <FieldShell label="Size" htmlFor="add-volume-size" hint="e.g., 1Gi, 500Mi.">
              <Input id="add-volume-size" value={size} onChange={(e) => setSize(e.target.value)} className="font-mono" />
            </FieldShell>
          </div>
          <FieldShell label="Attach to service" htmlFor="add-volume-service" required error={errors.resource}>
            <Select
              value={resourceIdx == null ? "" : String(resourceIdx)}
              onValueChange={(v) => setResourceIdx(Number(v))}
            >
              <SelectTrigger id="add-volume-service" className={errors.resource ? "border-danger" : ""}>
                <SelectValue placeholder="Select service" />
              </SelectTrigger>
              <SelectContent>
                {resources.map((r, i) =>
                  r.name ? (
                    <SelectItem key={r.name} value={String(i)}>
                      {r.name}
                    </SelectItem>
                  ) : null,
                )}
              </SelectContent>
            </Select>
          </FieldShell>
          <FieldShell
            label="Mount path"
            htmlFor="add-volume-path"
            required
            hint="Absolute path inside the service, e.g., /var/lib/data."
            error={errors.targetPath}
          >
            <Input
              id="add-volume-path"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="/var/lib/data"
              className={`font-mono ${errors.targetPath ? "border-danger" : ""}`}
              aria-invalid={!!errors.targetPath}
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={submit}>Add volume</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

If `FieldShell` doesn't associate `label` with `htmlFor` in a way testing-library resolves, check how `volume-fields.tsx` renders and mirror it; the tests query by `getByLabelText`.

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/canvas/tests/add-volume-dialog.test.tsx`
Expected: PASS

- [ ] **Step 5: Add the popover entry**

`AddResourcePopover.tsx` — add props and a Storage section. Extend the interface:

```ts
  /** False while the stack has no services — volumes are born attached. */
  canAddVolume: boolean;
  onAddVolume: () => void;
```

Destructure them in the component signature, then insert between `<BlockPicker …/>` and the `visibleAddons` block:

```tsx
          {"volume".includes(query.trim().toLowerCase()) || !query.trim() ? (
            <div className="mt-5">
              <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
                Storage
              </div>
              <button
                type="button"
                disabled={!canAddVolume}
                title={canAddVolume ? undefined : "Add a service first — volumes attach to a service."}
                onClick={() => {
                  setOpen(false);
                  onAddVolume();
                }}
                className="flex min-h-[60px] w-full items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary disabled:cursor-not-allowed disabled:opacity-50"
              >
                <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                  <HardDrive className="size-[18px]" />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block text-sm font-medium text-foreground">Volume</span>
                  <span className="block truncate font-mono text-[11px] text-muted-foreground">persistent storage</span>
                </span>
                <Plus className="h-[17px] w-[17px] text-primary" />
              </button>
            </div>
          ) : null}
```

Add `HardDrive` to the lucide import.

- [ ] **Step 6: Wire through CanvasEditor and StackCanvasTab**

`CanvasEditor.tsx`: add `canAddVolume: boolean; onAddVolume: () => void;` to `CanvasEditorProps`, destructure, and pass both to `<AddResourcePopover canAddVolume={canAddVolume} onAddVolume={onAddVolume} …/>`.

`StackCanvasTab.tsx` — inside `StackCanvasFlow`:

```tsx
import { AddVolumeDialog } from "./AddVolumeDialog";
import { addMount, newVolume } from "@/pages/stacks/lib/canvas/volume-ops";
import type { EditSessionDraft } from "@/pages/stacks/hooks/use-stack-edit-session";
```

```tsx
  const [addVolumeOpen, setAddVolumeOpen] = useState(false);
  const [addVolumeResourceIdx, setAddVolumeResourceIdx] = useState<number | null>(null);

  /** Apply a pure draft mutation, starting a session lazily when needed. */
  const applyDraft = useCallback(
    (fn: (draft: EditSessionDraft) => EditSessionDraft) => {
      const current: EditSessionDraft = session.isActive
        ? { resources: session.draft.resources, volumes: session.draft.volumes }
        : { resources: draftResources, volumes: draftVolumes };
      const next = fn(current);
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set(connectionAddonIds), draft: next },
        );
      } else {
        session.updateResources(() => next.resources);
        session.updateVolumes(() => next.volumes);
      }
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const onCreateVolume = useCallback(
    (input: { name: string; size: string; resourceIdx: number; targetPath: string }) => {
      applyDraft((draft) => ({
        resources: addMount(draft.resources, input.resourceIdx, {
          volumeName: input.name,
          targetPath: input.targetPath,
        }),
        volumes: [...draft.volumes, newVolume({ name: input.name, size: input.size })],
      }));
    },
    [applyDraft],
  );
```

Render after `<DrawerStack …/>`:

```tsx
      <AddVolumeDialog
        open={addVolumeOpen}
        onOpenChange={(o) => {
          setAddVolumeOpen(o);
          if (!o) setAddVolumeResourceIdx(null);
        }}
        resources={resources}
        volumes={volumes}
        initialResourceIdx={addVolumeResourceIdx}
        onCreate={onCreateVolume}
      />
```

Pass to `CanvasEditor`:

```tsx
          canAddVolume={resources.length > 0}
          onAddVolume={() => setAddVolumeOpen(true)}
```

- [ ] **Step 7: Type-check, lint, existing canvas tests**

Run: `cd frontend && pnpm exec tsc -b --noEmit 2>&1 | grep -i "canvas\|volume" ; pnpm exec vitest run src/pages/stacks/components/canvas/tests`
Expected: no new type errors in touched files; tests PASS (including `add-resource-popover.test.tsx` — fix its render call by adding the two new required props).

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas
git commit -m "feat(stacks): create attached volumes from the canvas add button"
```

---

### Task 5: Context menus (chip / floating volume / resource card) + delete confirm

**Files:**
- Create: `src/pages/stacks/components/canvas/CanvasContextMenu.tsx`
- Modify: `src/pages/stacks/components/canvas/nodes/ResourceNode.tsx` (chip `data-volume-chip` attribute)
- Modify: `src/pages/stacks/components/canvas/CanvasEditor.tsx` (`onNodeContextMenu` prop)
- Modify: `src/pages/stacks/components/canvas/StackCanvasTab.tsx` (menu state, actions, confirm dialog)
- Modify: `src/pages/stacks/components/canvas/VolumeDrawer.tsx` (cascade on remove)
- Test: `src/pages/stacks/components/canvas/tests/canvas-context-menu.test.tsx`

**Interfaces:**
- Consumes: Task 1's `removeMountsOf`, `mountOwner`; Task 4's `applyDraft`, `setAddVolumeOpen`, `setAddVolumeResourceIdx`; existing `openVolume`, `removeResource`, `onNodeClick` logic.
- Produces:

```ts
export type CanvasMenuTarget =
  | { kind: "resource"; resourceIdx: number; x: number; y: number }
  | { kind: "volume-chip"; volumeName: string; x: number; y: number }
  | { kind: "volume-node"; volumeName: string; x: number; y: number };

interface CanvasContextMenuProps {
  target: CanvasMenuTarget | null;
  onClose: () => void;
  onOpenResource: (resourceIdx: number) => void;
  onAddVolumeToResource: (resourceIdx: number) => void;
  onDeleteResource: (resourceIdx: number) => void;
  onDisconnectVolume: (volumeName: string) => void;
  onOpenVolume: (volumeName: string) => void;
  onRequestDeleteVolume: (volumeName: string) => void;
  onRequestAttach: (volumeName: string) => void; // Task 6 opens the attach dialog; wire a no-op until then
}
```

- [ ] **Step 1: Write the failing component test**

```tsx
// src/pages/stacks/components/canvas/tests/canvas-context-menu.test.tsx
import { describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import { CanvasContextMenu, type CanvasMenuTarget } from "../CanvasContextMenu";

const handlers = () => ({
  onClose: vi.fn(),
  onOpenResource: vi.fn(),
  onAddVolumeToResource: vi.fn(),
  onDeleteResource: vi.fn(),
  onDisconnectVolume: vi.fn(),
  onOpenVolume: vi.fn(),
  onRequestDeleteVolume: vi.fn(),
  onRequestAttach: vi.fn(),
});

const renderMenu = (target: CanvasMenuTarget) => {
  const h = handlers();
  render(<CanvasContextMenu target={target} {...h} />);
  return h;
};

describe("CanvasContextMenu", () => {
  it("shows chip items and fires disconnect", () => {
    const h = renderMenu({ kind: "volume-chip", volumeName: "data", x: 10, y: 10 });
    expect(screen.getByText("Disconnect volume")).toBeInTheDocument();
    expect(screen.getByText("Volume settings")).toBeInTheDocument();
    expect(screen.getByText("Delete volume")).toBeInTheDocument();
    expect(screen.queryByText("Attach to service…")).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("Disconnect volume"));
    expect(h.onDisconnectVolume).toHaveBeenCalledWith("data");
  });

  it("shows floating-volume items and fires attach", () => {
    const h = renderMenu({ kind: "volume-node", volumeName: "data", x: 10, y: 10 });
    expect(screen.getByText("Attach to service…")).toBeInTheDocument();
    expect(screen.queryByText("Disconnect volume")).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("Attach to service…"));
    expect(h.onRequestAttach).toHaveBeenCalledWith("data");
  });

  it("shows resource items and fires delete", () => {
    const h = renderMenu({ kind: "resource", resourceIdx: 2, x: 10, y: 10 });
    expect(screen.getByText("Open settings")).toBeInTheDocument();
    expect(screen.getByText("Add volume…")).toBeInTheDocument();
    fireEvent.click(screen.getByText("Delete service"));
    expect(h.onDeleteResource).toHaveBeenCalledWith(2);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/canvas/tests/canvas-context-menu.test.tsx`
Expected: FAIL — `Cannot find module '../CanvasContextMenu'`

- [ ] **Step 3: Implement `CanvasContextMenu`**

Radix DropdownMenu with a virtual anchor pinned at the cursor:

```tsx
// src/pages/stacks/components/canvas/CanvasContextMenu.tsx
import { HardDrive, Link2, Plug, Settings2, Trash2 } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

export type CanvasMenuTarget =
  | { kind: "resource"; resourceIdx: number; x: number; y: number }
  | { kind: "volume-chip"; volumeName: string; x: number; y: number }
  | { kind: "volume-node"; volumeName: string; x: number; y: number };

interface CanvasContextMenuProps {
  target: CanvasMenuTarget | null;
  onClose: () => void;
  onOpenResource: (resourceIdx: number) => void;
  onAddVolumeToResource: (resourceIdx: number) => void;
  onDeleteResource: (resourceIdx: number) => void;
  onDisconnectVolume: (volumeName: string) => void;
  onOpenVolume: (volumeName: string) => void;
  onRequestDeleteVolume: (volumeName: string) => void;
  onRequestAttach: (volumeName: string) => void;
}

const DANGER_ITEM = "text-danger focus:text-danger focus:bg-danger-bg";

/** Right-click menu for canvas nodes: one component, item-set keyed by target kind. */
export function CanvasContextMenu({
  target,
  onClose,
  onOpenResource,
  onAddVolumeToResource,
  onDeleteResource,
  onDisconnectVolume,
  onOpenVolume,
  onRequestDeleteVolume,
  onRequestAttach,
}: CanvasContextMenuProps) {
  if (!target) return null;

  const volumeItems = (volumeName: string, mounted: boolean) => (
    <>
      {mounted ? (
        <DropdownMenuItem onSelect={() => onDisconnectVolume(volumeName)}>
          <Plug className="size-4" /> Disconnect volume
        </DropdownMenuItem>
      ) : (
        <DropdownMenuItem onSelect={() => onRequestAttach(volumeName)}>
          <Link2 className="size-4" /> Attach to service…
        </DropdownMenuItem>
      )}
      <DropdownMenuItem onSelect={() => onOpenVolume(volumeName)}>
        <Settings2 className="size-4" /> Volume settings
      </DropdownMenuItem>
      <DropdownMenuSeparator />
      <DropdownMenuItem className={DANGER_ITEM} onSelect={() => onRequestDeleteVolume(volumeName)}>
        <Trash2 className="size-4" /> Delete volume
      </DropdownMenuItem>
    </>
  );

  return (
    <DropdownMenu open onOpenChange={(open) => !open && onClose()}>
      <DropdownMenuTrigger asChild>
        <span style={{ position: "fixed", left: target.x, top: target.y, width: 0, height: 0 }} aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="start" sideOffset={2} className="w-52">
        {target.kind === "resource" ? (
          <>
            <DropdownMenuItem onSelect={() => onOpenResource(target.resourceIdx)}>
              <Settings2 className="size-4" /> Open settings
            </DropdownMenuItem>
            <DropdownMenuItem onSelect={() => onAddVolumeToResource(target.resourceIdx)}>
              <HardDrive className="size-4" /> Add volume…
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem className={DANGER_ITEM} onSelect={() => onDeleteResource(target.resourceIdx)}>
              <Trash2 className="size-4" /> Delete service
            </DropdownMenuItem>
          </>
        ) : (
          volumeItems(target.volumeName, target.kind === "volume-chip")
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/canvas/tests/canvas-context-menu.test.tsx`
Expected: PASS

- [ ] **Step 5: Wire right-click through the canvas**

`ResourceNode.tsx` — tag each chip row so a node-level handler can tell chip from card (add the attribute to the chip `div`):

```tsx
        <div
          key={v.name}
          title={v.mountPath}
          data-volume-chip={v.name}
          className="flex items-center gap-2 border-t border-border bg-background px-[13px] py-2"
        >
```

`CanvasEditor.tsx` — add to props and pass to `<ReactFlow>`:

```ts
  onNodeContextMenu?: NodeMouseHandler<CanvasFlowNode>;
```

`StackCanvasTab.tsx` — menu state + dispatcher + actions:

```tsx
import { CanvasContextMenu, type CanvasMenuTarget } from "./CanvasContextMenu";
import { removeMountsOf } from "@/pages/stacks/lib/canvas/volume-ops";
import { NODE_KIND, type AttachmentNodeData, type ResourceNodeData } from "@/pages/stacks/lib/canvas/graph-from-connections";
import {
  AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent,
  AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle,
} from "@/components/ui/alert-dialog";
```

```tsx
  const [menuTarget, setMenuTarget] = useState<CanvasMenuTarget | null>(null);
  const [pendingDeleteVolume, setPendingDeleteVolume] = useState<string | null>(null);

  const onNodeContextMenu = useCallback<NodeMouseHandler<CanvasFlowNode>>(
    (event, node) => {
      event.preventDefault();
      const { clientX: x, clientY: y } = event as React.MouseEvent;
      const chipEl = (event.target as HTMLElement).closest("[data-volume-chip]");
      if (chipEl) {
        setMenuTarget({ kind: "volume-chip", volumeName: chipEl.getAttribute("data-volume-chip")!, x, y });
        return;
      }
      if (node.type === "attachment" && (node.data as AttachmentNodeData).kind === NODE_KIND.volume) {
        setMenuTarget({ kind: "volume-node", volumeName: (node.data as AttachmentNodeData).name, x, y });
        return;
      }
      if (node.type === "resource") {
        const idx = (node.data as ResourceNodeData).resourceIdx;
        if (idx != null) setMenuTarget({ kind: "resource", resourceIdx: idx, x, y });
      }
    },
    [],
  );

  const openResourceDrawer = useCallback(
    (idx: number) => {
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          {
            linkedAddonIds: new Set(connectionAddonIds),
            openResourceIdx: idx,
            openTab: "configuration",
            draft: { resources: draftResources, volumes: draftVolumes },
          },
        );
      }
      setDrawerStack(replaceStack({ kind: "resource", index: idx }));
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );

  const onDisconnectVolume = useCallback(
    (volumeName: string) => {
      applyDraft((draft) => ({ ...draft, resources: removeMountsOf(draft.resources, volumeName) }));
    },
    [applyDraft],
  );

  const onDeleteVolumeConfirmed = useCallback(
    (volumeName: string) => {
      applyDraft((draft) => ({
        resources: removeMountsOf(draft.resources, volumeName),
        volumes: draft.volumes.filter((v) => v.name !== volumeName),
      }));
      setPendingDeleteVolume(null);
    },
    [applyDraft],
  );

  const openVolumeFromCanvas = useCallback(
    (volumeName: string) => {
      // The volume drawer reads session.draft — make sure a session exists first.
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set(connectionAddonIds), draft: { resources: draftResources, volumes: draftVolumes } },
        );
      }
      setDrawerStack((s) => pushEntry(s, { kind: "volume", name: volumeName }));
    },
    [session, baselineResources, baselineVolumes, draftResources, draftVolumes, connectionAddonIds],
  );
```

Refactor the existing `onNodeClick` to call `openResourceDrawer(idx)` (same body — dedupe), and give floating volume nodes a click action too (the spec closes today's "attachment nodes ignore clicks" gap):

```tsx
  const onNodeClick = useCallback<NodeMouseHandler<CanvasFlowNode>>(
    (_event, node) => {
      if (node.type === "attachment") {
        const data = node.data as AttachmentNodeData;
        if (data.kind === NODE_KIND.volume) openVolumeFromCanvas(data.name);
        return; // secret/object-store attachments stay display-only
      }
      const idx = (node.data as ResourceNodeData).resourceIdx;
      if (idx == null) return; // addon node — managed via the Environment tab, no drawer in v1
      openResourceDrawer(idx);
    },
    [openResourceDrawer, openVolumeFromCanvas],
  );
```

Also update `AttachmentNode.tsx`: change the wrapper's `cursor-default` to `cursor-pointer` and add `hover:border-brand/60 transition-colors` so volume nodes read as clickable (the class change applies to all attachment kinds; acceptable — secrets may get drawers later).

Render (next to `AddVolumeDialog`):

```tsx
      <CanvasContextMenu
        target={menuTarget}
        onClose={() => setMenuTarget(null)}
        onOpenResource={openResourceDrawer}
        onAddVolumeToResource={(idx) => {
          setAddVolumeResourceIdx(idx);
          setAddVolumeOpen(true);
        }}
        onDeleteResource={(idx) => {
          if (!session.isActive) return; // deleting requires a session; resource menus only render with resourceIdx from the draft
          removeResource(idx);
        }}
        onDisconnectVolume={onDisconnectVolume}
        onOpenVolume={openVolumeFromCanvas}
        onRequestDeleteVolume={setPendingDeleteVolume}
        onRequestAttach={() => {}}
      />
      <AlertDialog open={pendingDeleteVolume != null} onOpenChange={(o) => !o && setPendingDeleteVolume(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete volume “{pendingDeleteVolume}”?</AlertDialogTitle>
            <AlertDialogDescription>
              The volume and its data are removed when the stack deploys. If it is mounted, the mount is removed too.
              This cannot be undone after deploy.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              className="bg-danger text-white hover:bg-danger/90"
              onClick={() => pendingDeleteVolume && onDeleteVolumeConfirmed(pendingDeleteVolume)}
            >
              Delete volume
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
```

Note on `onDeleteResource`: when the session is inactive there is no draft to delete from; start one via `applyDraft((draft) => ({ ...draft, resources: draft.resources.filter((_, i) => i !== idx) }))` instead of the early-return above, and clear the drawer stack. Use this `applyDraft` form — it covers both active and inactive sessions:

```tsx
        onDeleteResource={(idx) => {
          applyDraft((draft) => ({ ...draft, resources: draft.resources.filter((_, i) => i !== idx) }));
          setDrawerStack([]);
        }}
```

Pass `onNodeContextMenu={onNodeContextMenu}` to `<CanvasEditor …/>`.

- [ ] **Step 6: Cascade the VolumeDrawer's Remove button**

`VolumeDrawer.tsx` — deleting from the drawer must also strip mounts (same cascade as the menu). Add a `session.updateResources` call:

```tsx
import { removeMountsOf } from "@/pages/stacks/lib/canvas/volume-ops";
```

```tsx
  const onRemove = useCallback(
    (idx: number) => {
      const name = volumes[idx]?.name;
      session.updateVolumes((prev) => prev.filter((_, i) => i !== idx));
      if (name) session.updateResources((prev) => removeMountsOf(prev, name));
      onClose();
    },
    [session, volumes, onClose],
  );
```

- [ ] **Step 7: Run canvas test suites + type-check**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/canvas/tests && pnpm exec tsc -b --noEmit 2>&1 | grep -i canvas`
Expected: PASS, no new type errors (update `volume-drawer.test.tsx` if its mock session lacks `updateResources`).

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas frontend/src/pages/stacks/lib/canvas
git commit -m "feat(stacks): right-click context menus for canvas volumes and services"
```

---

### Task 6: MountPathDialog + attach from context menu

**Files:**
- Create: `src/pages/stacks/components/canvas/MountPathDialog.tsx`
- Modify: `src/pages/stacks/components/canvas/StackCanvasTab.tsx` (attach state + handler; replace the `onRequestAttach` no-op)

**Interfaces:**
- Consumes: Task 1's `addMount`, `validateTargetPath`; Task 4's `applyDraft`.
- Produces:

```ts
interface MountPathDialogProps {
  /** null = closed. */
  volumeName: string | null;
  resources: ResourceArr;
  /** Fixed target (drag-drop). null = show the service picker (menu attach). */
  resourceIdx: number | null;
  onCancel: () => void;
  onAttach: (input: { volumeName: string; resourceIdx: number; targetPath: string }) => void;
}
```

- `StackCanvasTab` produces `attachRequest: { volumeName: string; resourceIdx: number | null } | null` state and `onAttachConfirm` — Task 7 reuses both for drag-drop.

- [ ] **Step 1: Implement `MountPathDialog`**

(Validation logic is already unit-tested via `volume-ops`; the dialog is thin composition — no dedicated component test, matching the popover precedent.)

```tsx
// src/pages/stacks/components/canvas/MountPathDialog.tsx
import { useEffect, useState } from "react";
import { HardDrive } from "lucide-react";
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { FieldShell } from "@/components/branded";
import type { ResourceArr } from "@/pages/stacks/lib/stack-diff";
import { validateTargetPath } from "@/pages/stacks/lib/canvas/volume-ops";

interface MountPathDialogProps {
  volumeName: string | null;
  resources: ResourceArr;
  resourceIdx: number | null;
  onCancel: () => void;
  onAttach: (input: { volumeName: string; resourceIdx: number; targetPath: string }) => void;
}

/** Attach an existing volume: fixed target after a drag-drop, picker from the menu. */
export function MountPathDialog({ volumeName, resources, resourceIdx, onCancel, onAttach }: MountPathDialogProps) {
  const open = volumeName != null;
  const [pickedIdx, setPickedIdx] = useState<number | null>(null);
  const [targetPath, setTargetPath] = useState("");
  const [errors, setErrors] = useState<{ resource?: string; targetPath?: string }>({});

  useEffect(() => {
    if (!open) return;
    setPickedIdx(resourceIdx ?? (resources.length === 1 ? 0 : null));
    setTargetPath("");
    setErrors({});
    // eslint-disable-next-line react-hooks/exhaustive-deps -- seed on open only
  }, [open, resourceIdx]);

  const submit = () => {
    const idx = resourceIdx ?? pickedIdx;
    const next = {
      resource: idx == null ? "Required" : undefined,
      targetPath: validateTargetPath(targetPath, idx == null ? undefined : resources[idx]),
    };
    setErrors(next);
    if (next.resource || next.targetPath || idx == null || !volumeName) return;
    onAttach({ volumeName, resourceIdx: idx, targetPath });
  };

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onCancel()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <HardDrive className="size-[18px] text-brand" /> Attach “{volumeName}”
          </DialogTitle>
        </DialogHeader>
        <div className="grid gap-4">
          {resourceIdx == null && (
            <FieldShell label="Service" htmlFor="attach-service" required error={errors.resource}>
              <Select value={pickedIdx == null ? "" : String(pickedIdx)} onValueChange={(v) => setPickedIdx(Number(v))}>
                <SelectTrigger id="attach-service" className={errors.resource ? "border-danger" : ""}>
                  <SelectValue placeholder="Select service" />
                </SelectTrigger>
                <SelectContent>
                  {resources.map((r, i) =>
                    r.name ? (
                      <SelectItem key={r.name} value={String(i)}>
                        {r.name}
                      </SelectItem>
                    ) : null,
                  )}
                </SelectContent>
              </Select>
            </FieldShell>
          )}
          <FieldShell
            label="Mount path"
            htmlFor="attach-path"
            required
            hint="Absolute path inside the service."
            error={errors.targetPath}
          >
            <Input
              id="attach-path"
              value={targetPath}
              onChange={(e) => setTargetPath(e.target.value)}
              placeholder="/var/lib/data"
              className={`font-mono ${errors.targetPath ? "border-danger" : ""}`}
              aria-invalid={!!errors.targetPath}
              autoFocus
            />
          </FieldShell>
        </div>
        <DialogFooter>
          <Button variant="ghost" onClick={onCancel}>
            Cancel
          </Button>
          <Button onClick={submit}>Attach</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 2: Wire attach state in `StackCanvasTab`**

```tsx
import { MountPathDialog } from "./MountPathDialog";
import { addMount } from "@/pages/stacks/lib/canvas/volume-ops"; // extend existing import
```

```tsx
  const [attachRequest, setAttachRequest] = useState<{ volumeName: string; resourceIdx: number | null } | null>(null);

  const onAttachConfirm = useCallback(
    (input: { volumeName: string; resourceIdx: number; targetPath: string }) => {
      applyDraft((draft) => ({
        ...draft,
        resources: addMount(draft.resources, input.resourceIdx, {
          volumeName: input.volumeName,
          targetPath: input.targetPath,
        }),
      }));
      setAttachRequest(null);
    },
    [applyDraft],
  );
```

Replace the `onRequestAttach={() => {}}` no-op from Task 5:

```tsx
        onRequestAttach={(volumeName) => setAttachRequest({ volumeName, resourceIdx: null })}
```

Render next to the other dialogs:

```tsx
      <MountPathDialog
        volumeName={attachRequest?.volumeName ?? null}
        resources={resources}
        resourceIdx={attachRequest?.resourceIdx ?? null}
        onCancel={() => setAttachRequest(null)}
        onAttach={onAttachConfirm}
      />
```

- [ ] **Step 3: Verify by type-check and full canvas suite**

Run: `cd frontend && pnpm exec tsc -b --noEmit 2>&1 | grep -i canvas ; pnpm exec vitest run src/pages/stacks/components/canvas/tests src/pages/stacks/lib/canvas/tests`
Expected: no new type errors; PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas
git commit -m "feat(stacks): attach floating volumes from the context menu"
```

---

### Task 7: Drag a floating volume onto a service card

**Files:**
- Modify: `src/pages/stacks/components/canvas/CanvasEditor.tsx` (drag prop pass-through)
- Modify: `src/pages/stacks/components/canvas/StackCanvasTab.tsx` (drag handlers)
- Modify: `src/pages/stacks/components/canvas/nodes/ResourceNode.tsx` (drop-target ring)
- Modify: `src/pages/stacks/lib/canvas/graph-from-connections.ts` (`dropTarget` flag on `ResourceNodeData`)

**Interfaces:**
- Consumes: Task 6's `setAttachRequest` (drop opens the same dialog with a fixed `resourceIdx`); React Flow v12 `useReactFlow().getIntersectingNodes`; `OnNodeDrag` type from `@xyflow/react`.
- Produces: `ResourceNodeData.dropTarget?: boolean` (transient, set only during drag).

- [ ] **Step 1: Add the `dropTarget` flag and ring**

`graph-from-connections.ts` — add to `ResourceNodeData`:

```ts
  /** Transient: true while a dragged volume hovers this card as a valid drop target. */
  dropTarget?: boolean;
```

`ResourceNode.tsx` — extend the card wrapper `cn(...)` call:

```tsx
        data.dropTarget && "border-brand ring-[3px] ring-brand/30",
```

(Place it after the `selected && …` line so the drop ring reads over selection.)

- [ ] **Step 2: Pass drag handlers through `CanvasEditor`**

Add to `CanvasEditorProps` and forward to `<ReactFlow>`:

```ts
import type { OnNodeDrag } from "@xyflow/react"; // extend existing type import

  onNodeDragStart?: OnNodeDrag<CanvasFlowNode>;
  onNodeDrag?: OnNodeDrag<CanvasFlowNode>;
  onNodeDragStop?: OnNodeDrag<CanvasFlowNode>;
```

- [ ] **Step 3: Implement drag logic in `StackCanvasTab`**

```tsx
import { useRef } from "react"; // extend existing react import
import type { OnNodeDrag, XYPosition } from "@xyflow/react"; // extend existing import
import { NODE_ID_PREFIX } from "@/pages/stacks/lib/canvas/graph-from-connections"; // extend existing import
```

```tsx
  const { fitView, getIntersectingNodes } = useReactFlow(); // extend existing destructure
  const dragStartPos = useRef<XYPosition | null>(null);

  const isFloatingVolume = (node: CanvasFlowNode) =>
    node.type === "attachment" && (node.data as AttachmentNodeData).kind === NODE_KIND.volume;

  /** First intersecting service card (addons have no resourceIdx and never qualify). */
  const dropTargetFor = useCallback(
    (node: CanvasFlowNode): CanvasFlowNode | null => {
      const hit = getIntersectingNodes(node).find(
        (n) => n.type === "resource" && (n.data as ResourceNodeData).resourceIdx != null,
      );
      return (hit as CanvasFlowNode) ?? null;
    },
    [getIntersectingNodes],
  );

  const onNodeDragStart = useCallback<OnNodeDrag<CanvasFlowNode>>(
    (_event, node) => {
      dragStartPos.current = isFloatingVolume(node) ? { ...node.position } : null;
    },
    [],
  );

  const onNodeDrag = useCallback<OnNodeDrag<CanvasFlowNode>>(
    (_event, node) => {
      if (!isFloatingVolume(node)) return;
      const target = dropTargetFor(node);
      setNodes((prev) =>
        prev.map((n) => {
          const isTarget = n.id === target?.id;
          const current = (n.data as ResourceNodeData).dropTarget ?? false;
          if (current === isTarget) return n;
          return { ...n, data: { ...n.data, dropTarget: isTarget } } as CanvasFlowNode;
        }),
      );
    },
    [dropTargetFor, setNodes],
  );

  const onNodeDragStop = useCallback<OnNodeDrag<CanvasFlowNode>>(
    (_event, node) => {
      if (!isFloatingVolume(node)) return;
      const target = dropTargetFor(node);
      // Clear all rings.
      setNodes((prev) =>
        prev.map((n) =>
          (n.data as ResourceNodeData).dropTarget ? ({ ...n, data: { ...n.data, dropTarget: false } } as CanvasFlowNode) : n,
        ),
      );
      if (!target) {
        dragStartPos.current = null;
        return; // plain reposition
      }
      const volumeName = node.id.slice(NODE_ID_PREFIX.volume.length);
      const resourceIdx = (target.data as ResourceNodeData).resourceIdx!;
      setAttachRequest({ volumeName, resourceIdx });
    },
    [dropTargetFor, setNodes],
  );
```

Cancel path — restore the node's pre-drag position. Extend the `MountPathDialog` `onCancel`:

```tsx
        onCancel={() => {
          const req = attachRequest;
          setAttachRequest(null);
          const start = dragStartPos.current;
          dragStartPos.current = null;
          if (req && start) {
            setNodes((prev) =>
              prev.map((n) => (n.id === NODE_ID_PREFIX.volume + req.volumeName ? { ...n, position: start } : n)),
            );
          }
        }}
```

And clear `dragStartPos.current = null` inside `onAttachConfirm` too (successful attach removes the floating node — the derived graph re-renders it as a chip).

Pass the three handlers to `<CanvasEditor …/>`:

```tsx
          onNodeDragStart={onNodeDragStart}
          onNodeDrag={onNodeDrag}
          onNodeDragStop={onNodeDragStop}
```

- [ ] **Step 4: Verify types + suites**

Run: `cd frontend && pnpm exec tsc -b --noEmit 2>&1 | grep -i canvas ; pnpm exec vitest run src/pages/stacks/lib/canvas/tests src/pages/stacks/components/canvas/tests`
Expected: no new type errors; PASS

- [ ] **Step 5: Manual drag check (Playwright MCP)**

With `pnpm dev` (or the `:8000` single server) running: open a stack canvas, add a volume via the add button, right-click its chip → Disconnect (it floats), drag the floating node over a service card (ring appears), drop, enter `/var/lib/test`, Attach. Verify the chip docks and the pending-changes badge shows the resource dirty. Then drag it over empty canvas — plain reposition, no dialog.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas frontend/src/pages/stacks/lib/canvas
git commit -m "feat(stacks): drag floating volumes onto service cards to attach"
```

---

### Task 8: Retire mount editing in the canvas resource drawer

**Files:**
- Modify: `src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx` (add `mountsReadOnly` prop; read-only rows branch)
- Modify: `src/pages/stacks/components/shared/hooks/use-resource-tab-props.ts` (thread the prop)
- Modify: `src/pages/stacks/components/canvas/ResourceDrawer.tsx` (set `mountsReadOnly`; drop `onCreateVolume`)

**Interfaces:**
- Consumes: existing `onOpenVolume` (kept — read-only rows still open the volume drawer).
- Produces: `StackResourceConfigurationTab` prop `mountsReadOnly?: boolean` (default false — wizard/detail forms keep editing); context field `mountsReadOnly?: boolean` in `use-resource-tab-props.ts`.

- [ ] **Step 1: Add the prop and the read-only branch**

`stack-resource-configuration-tab.tsx`:
- Add `mountsReadOnly?: boolean;` to the props interface (near `onOpenVolume`) and destructure it (default `false`).
- Wrap the Volume Mounts section: when `mountsReadOnly`, render compact rows instead of the editable grid, and hide the add controls:

```tsx
      {/* Volume Mounts Section */}
      <div>
        <h3 className="text-sm font-semibold text-foreground mb-3">Volume Mounts</h3>
        {mountsReadOnly ? (
          <div className="grid gap-1.5 max-w-3xl">
            {(draft.volume_mounts || []).length === 0 && (
              <p className="text-sm text-muted-foreground">
                No volumes mounted. Add one from the canvas — “+ Add resource → Volume”.
              </p>
            )}
            {(draft.volume_mounts || []).map((vm: VolumeMount, vmIdx: number) => (
              <div
                key={vmIdx}
                className="flex items-center gap-2 rounded-md border border-border bg-muted/10 px-3 py-2"
              >
                <HardDrive className="h-3.5 w-3.5 shrink-0 text-fg-muted" aria-hidden />
                <span className="truncate font-mono text-[12.5px] text-foreground">{vm.source_volume_name}</span>
                <code className="ml-auto shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[11px] text-muted-foreground">
                  {vm.target_path}
                </code>
                {onOpenVolume && vm.source_volume_name && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="size-7 shrink-0 text-fg-muted hover:text-brand"
                    aria-label={`Open volume ${vm.source_volume_name}`}
                    title="Open volume settings"
                    onClick={() => onOpenVolume(vm.source_volume_name!)}
                  >
                    <ArrowUpRight className="size-3.5" />
                  </Button>
                )}
              </div>
            ))}
            <p className="mt-1 text-[12.5px] text-muted-foreground">
              Manage mounts on the canvas: right-click a volume to disconnect, drag a floating volume onto a service to attach.
            </p>
          </div>
        ) : (
          <div className="grid gap-5 max-w-3xl">
            {/* …existing editable rows + add controls, unchanged… */}
          </div>
        )}
      </div>
```

Add `HardDrive` to the file's lucide import.

- [ ] **Step 2: Thread the prop**

`use-resource-tab-props.ts`: add `mountsReadOnly?: boolean;` next to `onCreateVolume`/`onOpenVolume` in the context interface, and `mountsReadOnly: context.mountsReadOnly,` next to lines 186-187.

`ResourceDrawer.tsx`: in the context object (around line 130) replace `onCreateVolume,` with `mountsReadOnly: true,` and delete the now-unused `onCreateVolume` callback (lines 93-110) and its `addInlineVolume` import if unreferenced.

Keep `InlineVolumeAdder` and the `onCreateVolume` prop on the configuration tab itself — the wizard/detail surfaces still use them; only the canvas drawer opts out.

- [ ] **Step 3: Verify — type-check + full stacks suite**

Run: `cd frontend && pnpm exec tsc -b --noEmit 2>&1 | grep -iE "resource-(drawer|tab|configuration)" ; pnpm exec vitest run src/pages/stacks`
Expected: no new type errors; PASS

- [ ] **Step 4: Commit**

```bash
git add frontend/src/pages/stacks/components/shared frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx
git commit -m "refactor(stacks): make canvas drawer volume mounts read-only"
```

---

### Task 9: Full verification

**Files:** none new.

- [ ] **Step 1: Full frontend gate**

Run: `cd frontend && pnpm lint && pnpm exec tsc -b --noEmit && pnpm exec vitest run`
Expected: all green (pre-existing unrelated type errors, if any, must match `main`'s baseline — compare before blaming this branch).

- [ ] **Step 2: End-to-end manual pass (Playwright MCP against the dev server)**

Walk the whole lifecycle on a real stack, verifying after each step that the canvas, drawer, and pending-changes state agree:
1. "+ Add resource → Volume" → dialog → create attached → chip docks, resource + volume both read dirty.
2. Right-click chip → Volume settings → drawer opens.
3. Right-click chip → Disconnect volume → chip gone, floating node appears, resource dirty.
4. Drag floating node onto another service → ring highlight → drop → mount path → Attach → chip docks there.
5. Right-click floating/chip → Delete volume → confirm dialog → volume gone everywhere, no dangling mounts (check the resource drawer's read-only rows).
6. Resource card right-click → Open settings / Add volume… / Delete service all work.
7. Deploy (or autosave-sync) and confirm the server accepts the connection set — watch the network tab for connection CRUD calls.

- [ ] **Step 3: Update spec status + commit any fixes**

```bash
git add -A
git commit -m "test(stacks): verify canvas volume lifecycle end-to-end"
```

(Skip the commit if the tree is clean.)
