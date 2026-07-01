# Canvas-Only Create Surface — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the React Flow canvas the single surface to create *and* edit a stack; every wizard path opens the canvas; the legacy stack form is deleted and unreachable.

**Architecture:** A new `/stacks/new` route renders the existing `StackDetailPage` in *draft mode* (`isDraft = !id`): no fetch, an edit session seeded from `location.state.seed` against an **empty baseline** (so everything reads as "added"), and a **Save that calls `createStack` then redirects to `/stacks/:id`**. Name/labels become page-level state edited in the shell header. Volumes are authored inline in the resource drawer. Managed addons are pickable from the "+ Add resource" popover. The form page + route + the `isCanvasEnabled` flag are removed.

**Tech Stack:** React 19, Vite, Tailwind v4, Radix/shadcn, `@xyflow/react` v12, Zod (`FormStackSchema`), Vitest (jsdom), Playwright MCP.

## Global Constraints

- **The legacy form is gone.** `StackCreatePage` (`frontend/src/pages/stacks/components/create/`) + the `/stacks/create` route are deleted. No path may navigate to, render, or fall back to it. The user must never see the whole-form view.
- The `isCanvasEnabled` feature flag is removed (its off-branch rendered the deleted form). Canvas is unconditional.
- Brand design system only — `index.css` tokens + `branded/`/`ui/` primitives. No raw hex, no off-scale type.
- No third-party-PaaS names ("Railway" etc.) anywhere — code, copy, commits.
- Constants over magic strings (model enums, annotation keys, states).
- Pure calc/data separation: new logic (`buildDraftFormData`, `addInlineVolume`, seed builders) is pure + unit-tested; components stay views.
- Login for manual verification: `admin@stackdome.io` / `welcome@123`. Test stack `tooljet-addon` = `/stacks/d3e497e8-2ec4-4f24-866c-dc6152dee9fa`.
- Frontend commands (run from repo root): tests `pnpm --prefix frontend test:run <path>`; lint `pnpm --prefix frontend lint`; types `pnpm --prefix frontend exec tsc -b`. `git` runs from the worktree root `/Users/akshaysasidharan/code/stackdome/.claude/worktrees/stack-canvas-editor`; `pnpm` runs with `--prefix frontend`.

---

## File Structure

**New files**
- `frontend/src/pages/stacks/lib/canvas/draft-seed.ts` — pure. `emptyDraftSeed()`, `buildDraftFormData()`, the `DraftSeed` type. Consumed by the wizard paths (produce a seed) and `StackDetailPage` (consume a seed → create payload).
- `frontend/src/pages/stacks/lib/canvas/inline-volume.ts` — pure `addInlineVolume()` reducer (create a stack volume + its resource mount together).
- `frontend/src/pages/stacks/components/canvas/DraftTabPlaceholder.tsx` — muted "Available after you save" empty state for ops tabs in draft.
- Co-located `tests/` for each pure module.

**Modified files**
- `frontend/src/App.tsx` — add `/stacks/new`; (Task 6) remove `/stacks/create` + import.
- `frontend/src/components/app-layout.tsx` — full-bleed includes `/stacks/new`; (Task 6) drop `isCanvasEnabled`.
- `frontend/src/pages/stacks/components/detail/index.tsx` — draft branch (no fetch, seed→session, Save=create), draft name/labels state; (Task 6) delete flag-OFF form branch.
- `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx` — editable name + labels row; `isDraft` (primary always Save).
- `frontend/src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx` — optional `onCreateVolume` → inline "Add volume".
- `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx` — thread `onCreateVolume` from `session.updateVolumes`.
- `frontend/src/pages/stacks/components/canvas/AddResourcePopover.tsx` — Managed add-ons section + `onLinkAddon`.
- `frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx` + `StackCanvasTab.tsx` — thread addon list + `onLinkAddon`.
- `frontend/src/pages/stacks/components/wizard/block-composer.tsx`, `hooks/use-template-import.ts`, `hooks/use-docker-compose-import.ts`, `components/wizard/stack-create-wizard.tsx`, `components/nav-stacks.tsx` — navigate to `/stacks/new` with a seed.

**Deleted files (Task 6)**
- `frontend/src/pages/stacks/components/create/` (dir + tests).
- `frontend/src/lib/feature-flags.ts` + `frontend/src/lib/tests/feature-flags.test.ts`.

---

## Task 1: Draft route + draft-mode detail page (blank canvas, Save=create)

**Files:**
- Create: `frontend/src/pages/stacks/lib/canvas/draft-seed.ts`
- Test: `frontend/src/pages/stacks/lib/canvas/tests/draft-seed.test.ts`
- Modify: `frontend/src/App.tsx:41-72` (add route), `frontend/src/components/app-layout.tsx:38-41` (full-bleed), `frontend/src/pages/stacks/components/detail/index.tsx` (draft branch)

**Interfaces:**
- Produces:
  - `interface DraftSeed { name: string; labels: FormStackData["labels"]; resources: FormStackResourceData[]; volumes: FormVolumeExtendedData[]; linkedAddonIds: string[] }`
  - `function emptyDraftSeed(): DraftSeed`
  - `function buildDraftFormData(name: string, labels: FormStackData["labels"], resources: FormStackResourceData[], volumes: FormVolumeExtendedData[]): FormStackData`
- Consumes: `FormStackData`, `FormStackResourceData`, `FormVolumeExtendedData` from `@/pages/stacks/schemas/form-schema`; `createStack` from `@/api/stacks`; `convertFormStackToApiStack`, `FormStackSchema` from the schema.

- [ ] **Step 1: Write the failing test for the pure seed helpers**

Create `frontend/src/pages/stacks/lib/canvas/tests/draft-seed.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { emptyDraftSeed, buildDraftFormData } from "../draft-seed";

describe("emptyDraftSeed", () => {
  it("returns an empty, named-blank seed", () => {
    expect(emptyDraftSeed()).toEqual({
      name: "",
      labels: [],
      resources: [],
      volumes: [],
      linkedAddonIds: [],
    });
  });
});

describe("buildDraftFormData", () => {
  it("assembles FormStackData from draft name/labels/resources/volumes", () => {
    const resources = [{ name: "api", sourceType: "image", image_spec: { image: "nginx:1" } }] as never;
    const out = buildDraftFormData("my-app", [{ key: "k", value: "v" }], resources, []);
    expect(out).toEqual({
      name: "my-app",
      labels: [{ key: "k", value: "v" }],
      spec: { stack_resources: resources, volumes: [] },
    });
  });
});
```

- [ ] **Step 2: Run it and watch it fail**

Run: `pnpm --prefix frontend test:run src/pages/stacks/lib/canvas/tests/draft-seed.test.ts`
Expected: FAIL — `Cannot find module '../draft-seed'`.

- [ ] **Step 3: Implement the pure helpers**

Create `frontend/src/pages/stacks/lib/canvas/draft-seed.ts`:

```ts
import type {
  FormStackData,
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

/** A create-flow seed handed to the draft canvas via navigation state. */
export interface DraftSeed {
  name: string;
  labels: FormStackData["labels"];
  resources: FormStackResourceData[];
  volumes: FormVolumeExtendedData[];
  linkedAddonIds: string[];
}

export function emptyDraftSeed(): DraftSeed {
  return { name: "", labels: [], resources: [], volumes: [], linkedAddonIds: [] };
}

/** Assemble the validated create payload shape from the canvas draft state. */
export function buildDraftFormData(
  name: string,
  labels: FormStackData["labels"],
  resources: FormStackResourceData[],
  volumes: FormVolumeExtendedData[],
): FormStackData {
  return {
    name,
    labels,
    spec: { stack_resources: resources, volumes },
  } as FormStackData;
}
```

- [ ] **Step 4: Run the test — it passes**

Run: `pnpm --prefix frontend test:run src/pages/stacks/lib/canvas/tests/draft-seed.test.ts`
Expected: PASS (2 tests).

- [ ] **Step 5: Register the `/stacks/new` route**

In `frontend/src/App.tsx`, add the draft route directly above the `/stacks/:id` route (order matters — `new` must be matched as a literal before the `:id` param, but React Router v6 ranks static segments above params automatically; place it above for clarity):

```tsx
        <Route path="/stacks/create" element={<StackCreatePage />} />
        <Route path="/stacks/new" element={<StackDetailPage />} />
        <Route path="/stacks/:id" element={<StackDetailPage />} />
```

(`/stacks/create` stays for now — deleted in Task 6.)

- [ ] **Step 6: Make `/stacks/new` full-bleed**

In `frontend/src/components/app-layout.tsx:38-41`, change the full-bleed test so the draft route is included (drop the `!endsWith("/new")` exclusion; keep the flag for now):

```tsx
  // Full-bleed layout for the canvas stack editor: /stacks/new (draft) and
  // /stacks/<id> (existing). A single trailing segment only — not /stacks.
  const isFullBleed =
    isCanvasEnabled() && /^\/stacks\/[^/]+$/.test(location.pathname);
```

- [ ] **Step 7: Add draft state + branch to `StackDetailPage`**

In `frontend/src/pages/stacks/components/detail/index.tsx`:

(a) Add imports near the top:

```tsx
import { createStack, getStackById, updateStack } from "@/api/stacks";
import { emptyDraftSeed, buildDraftFormData, type DraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
```

(b) After `const { id } = useParams();` (line ~85) add:

```tsx
  const isDraft = !id;
  const location = useLocation();
  const navigate = useNavigate();
  const seed = useMemo<DraftSeed>(
    () => ((location.state as { seed?: DraftSeed } | null)?.seed) ?? emptyDraftSeed(),
    // read once from the entry navigation; later navigations replace state
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );
  const [draftName, setDraftName] = useState(seed.name);
  const [draftLabels, setDraftLabels] = useState<FormStackData["labels"]>(seed.labels);
```

(Remove the later duplicate `const location = useLocation(); const navigate = useNavigate();` at lines ~188-189 — they now live here.)

(c) Guard the fetch effect (line ~120) so it no-ops in draft — wrap its body:

```tsx
  useEffect(() => {
    if (isDraft) return;
    const path = `/stacks/${id}`;
    // ...existing body unchanged...
  }, [currentStack, id, defaultTeamName, setCustomLabel, setPathLoading, isDraft]);
```

(d) Seed the session once in draft mode (empty baseline → everything reads as added). Add a new effect after `baselineVolumes`:

```tsx
  const draftSeeded = useRef(false);
  useEffect(() => {
    if (!isDraft || draftSeeded.current) return;
    draftSeeded.current = true;
    // Baseline empty so seeded resources/volumes read as "added" and Save is enabled.
    session.start({ resources: [], volumes: [] }, { linkedAddonIds: new Set(seed.linkedAddonIds) });
    if (seed.resources.length) session.updateResources(() => seed.resources);
    if (seed.volumes.length) session.updateVolumes(() => seed.volumes);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isDraft]);
```

(e) Replace the existing `wizardLinkApplied` effect (lines ~188-201) entirely — the wizard now enters via `/stacks/new` + seed, so the `/stacks/:id` + `linkedAddonIds` pre-link path is dead. Delete that effect.

(f) In `performSave` (line ~361), fork for draft. Replace the guard `if (!stackToShow || !session.isActive || !id) return;` and the save call:

```tsx
  const performSave = async () => {
    if (!session.isActive) return;
    if (!isDraft && !stackToShow) return;
    setIsSaving(true);
    setValidationErrors({ resources: {}, volumes: {} });
    try {
      const orgId = getCurrentOrganizationId();
      if (!orgId) throw new Error("Organization ID not found");

      const detachResult = session.pendingDetach.size > 0 ? applyPendingDetach() : null;
      const resources = (detachResult?.resources ?? session.draft.resources) as FormStackResourceData[];

      const formStackData: FormStackData = isDraft
        ? buildDraftFormData(draftName.trim(), draftLabels, resources, session.draft.volumes as VolumeFormData[])
        : {
            name: stackToShow!.name || "",
            labels: stackToShow!.labels || [],
            spec: { stack_resources: resources, volumes: session.draft.volumes as VolumeFormData[] },
          };

      const validation = FormStackSchema.safeParse(formStackData);
      if (!validation.success) {
        // ...existing validation-error mapping + toast + return unchanged...
      }

      const teamName = isDraft ? defaultTeamName : teamNameById(fetchedStack?.team_id ?? currentStack?.team_id);
      if (!teamName) {
        toast({ title: isDraft ? "No team available" : "Failed to update stack",
          description: "Could not resolve a team to save into.", variant: "destructive" });
        setIsSaving(false);
        return;
      }

      const apiData = convertFormStackToApiStack(formStackData);
      if (isDraft) {
        const created = await createStack(orgId, teamName, apiData);
        session.discard();
        navigate(`/stacks/${created.id}`, { replace: true, state: null });
        return;
      }

      const updatedStack = await updateStack(orgId, teamName, id!, apiData);
      // ...existing post-update body unchanged...
    } catch (err) {
      // ...existing catch unchanged (title stays generic)...
    } finally {
      setIsSaving(false);
    }
  };
```

(g) Build a `stackToShow` for draft so the render path below works. After `const stackToShow = currentStack || fetchedStack;` (line ~160) add a draft synthesis:

```tsx
  const draftStackView = useMemo(
    () =>
      isDraft
        ? ({
            name: draftName,
            labels: draftLabels,
            spec: { stack_resources: session.draft.resources, volumes: session.draft.volumes, connections: [] },
          } as unknown as Stack)
        : null,
    [isDraft, draftName, draftLabels, session.draft.resources, session.draft.volumes],
  );
  const effectiveStack = draftStackView ?? stackToShow;
```

Then in the flag-ON render block (line ~688) use `effectiveStack` in place of `stackToShow` for `stackName`/`subtitle`/`statusState`, and pass draft props (Task 2 wires name/labels). For draft, gate the `if (!stackToShow)` "not found" guard (line ~524) so it does not fire in draft:

```tsx
  if (!isDraft && !stackToShow) {
    // ...existing "Stack not found" block...
  }
```

(h) Skip the `loading` guard in draft (draft never fetches): change `if (loading)` to `if (!isDraft && loading)`.

- [ ] **Step 8: Verify types + lint**

Run: `pnpm --prefix frontend exec tsc -b`
Expected: no NEW errors (pre-existing `postgres-backups.ts` error may remain).
Run: `pnpm --prefix frontend lint`
Expected: clean for touched files.

- [ ] **Step 9: Manual Playwright verification (blank draft)**

Set `localStorage.stackCanvas="1"`, login, navigate to `/stacks/new`.
Expected: full-bleed canvas, header title empty, "No resources yet" empty state, primary button present. (Naming + save wired in Tasks 2/next; here confirm the route renders full-bleed with the empty canvas and no fetch/404.)

- [ ] **Step 10: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/draft-seed.ts \
  frontend/src/pages/stacks/lib/canvas/tests/draft-seed.test.ts \
  frontend/src/App.tsx frontend/src/components/app-layout.tsx \
  frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(stacks): draft canvas at /stacks/new — save creates then opens the stack"
```

---

## Task 2: Editable name + labels in the canvas header

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (pass name/labels props + `isDraft`)
- Test: `frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`

**Interfaces:**
- Consumes: `buildDraftFormData` (Task 1) indirectly via detail page; shell owns no state.
- Produces (new `CanvasEditorShellProps` fields): `isDraft?: boolean`, `nameEditable: boolean`, `onNameChange?: (name: string) => void`, `labels: { key: string; value: string }[]`, `labelsEditable: boolean`, `onAddLabel?: (value: string) => void`, `onRemoveLabel?: (index: number) => void`.

- [ ] **Step 1: Write the failing component test**

Create `frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { CanvasEditorShell } from "../CanvasEditorShell";

const base = {
  statusState: null, subtitle: "0 services · 0 volumes",
  activeTab: "configuration", onTabChange: () => {},
  isActive: true, dirtyResourceCount: 0, dirtyTotal: 1, isStaged: false,
  isSaving: false, deployBusy: false, canWrite: true,
  onSave: () => {}, onDeploy: () => {}, onDiscardAll: () => {}, onEdit: () => {}, onDelete: () => {},
  configuration: <div />, deployments: <div />, logs: <div />, metrics: <div />,
  labels: [], labelsEditable: true,
};

describe("CanvasEditorShell header", () => {
  it("renders an editable name input in draft and reports changes", () => {
    const onNameChange = vi.fn();
    render(<CanvasEditorShell {...base} stackName="" isDraft nameEditable onNameChange={onNameChange} />);
    const input = screen.getByPlaceholderText("name-your-stack");
    fireEvent.change(input, { target: { value: "web" } });
    expect(onNameChange).toHaveBeenCalledWith("web");
  });

  it("renders the name as static text when not editable", () => {
    render(<CanvasEditorShell {...base} stackName="tooljet" nameEditable={false} />);
    expect(screen.getByRole("heading", { name: "tooljet" })).toBeInTheDocument();
    expect(screen.queryByPlaceholderText("name-your-stack")).toBeNull();
  });

  it("adds and removes labels", () => {
    const onAddLabel = vi.fn();
    const onRemoveLabel = vi.fn();
    render(
      <CanvasEditorShell {...base} stackName="web" nameEditable
        labels={[{ key: "k", value: "prod" }]} onAddLabel={onAddLabel} onRemoveLabel={onRemoveLabel} />,
    );
    fireEvent.click(screen.getByLabelText("Remove label prod"));
    expect(onRemoveLabel).toHaveBeenCalledWith(0);
    const labelInput = screen.getByPlaceholderText("add label…");
    fireEvent.change(labelInput, { target: { value: "dev" } });
    fireEvent.keyDown(labelInput, { key: "Enter" });
    expect(onAddLabel).toHaveBeenCalledWith("dev");
  });
});
```

- [ ] **Step 2: Run it — fails**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`
Expected: FAIL — `getByPlaceholderText("name-your-stack")` not found (no editable input yet).

- [ ] **Step 3: Add name+labels props and UI to the shell**

In `CanvasEditorShell.tsx`, extend the props interface (after `subtitle: string;`):

```tsx
  /** Draft (unsaved) stack — primary action is always Save (create). */
  isDraft?: boolean;
  /** Render the title as an editable input (draft, or a rename-capable stack). */
  nameEditable: boolean;
  onNameChange?: (name: string) => void;
  labels: { key: string; value: string }[];
  labelsEditable?: boolean;
  onAddLabel?: (value: string) => void;
  onRemoveLabel?: (index: number) => void;
```

Add imports at top: `import { Input } from "@/components/ui/input";` and `X` to the lucide import. In the component signature destructure the new props (`isDraft`, `nameEditable`, `onNameChange`, `labels`, `labelsEditable`, `onAddLabel`, `onRemoveLabel`) and add local label-input state:

```tsx
  const [labelInput, setLabelInput] = useState("");
```

Replace the `<h1>` title (line ~129) with a conditional:

```tsx
          {nameEditable ? (
            <Input
              aria-label="Stack name"
              value={stackName}
              onChange={(e) => onNameChange?.(e.target.value)}
              placeholder="name-your-stack"
              className="h-auto w-[22ch] border-0 bg-transparent px-0 text-[29px] font-medium tracking-[-0.02em] shadow-none focus-visible:ring-0"
            />
          ) : (
            <h1 className="truncate text-[29px] font-medium tracking-[-0.02em] text-foreground">{stackName}</h1>
          )}
```

Make the primary button honor draft (draft → always Save). Change the `hasUnsaved` line:

```tsx
  const hasUnsaved = isActive && dirtyTotal > 0;
  const primaryIsSave = isDraft || hasUnsaved;
```

and use `primaryIsSave` in the `primaryButton` ternary instead of `hasUnsaved`.

Add a labels row after the subtitle `<p>` (line ~167):

```tsx
        <div className="mt-2 flex flex-wrap items-center gap-1.5">
          {labels.map((l, i) => (
            <span key={`${l.value}-${i}`} className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-[11px] text-muted-foreground">
              {l.value}
              {labelsEditable && (
                <button type="button" aria-label={`Remove label ${l.value}`} onClick={() => onRemoveLabel?.(i)} className="rounded-full hover:text-foreground">
                  <X className="size-3" />
                </button>
              )}
            </span>
          ))}
          {labelsEditable && (
            <Input
              value={labelInput}
              onChange={(e) => setLabelInput(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && labelInput.trim()) {
                  e.preventDefault();
                  onAddLabel?.(labelInput.trim());
                  setLabelInput("");
                }
              }}
              placeholder="add label…"
              className="h-6 w-[14ch] border-0 bg-transparent px-0 text-[11px] shadow-none focus-visible:ring-0"
            />
          )}
        </div>
```

- [ ] **Step 4: Run the shell test — passes**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 5: Wire draft name/labels from the detail page**

In `detail/index.tsx`, in the flag-ON `CanvasEditorShell` render (line ~693), add handlers + props. Define handlers above the return:

```tsx
  const addDraftLabel = useCallback((value: string) => {
    setDraftLabels((prev) => [...(prev ?? []), { key: "stackdome.io/user-defined-value", value }]);
  }, []);
  const removeDraftLabel = useCallback((idx: number) => {
    setDraftLabels((prev) => (prev ?? []).filter((_, i) => i !== idx));
  }, []);
```

Pass to the shell:

```tsx
          stackName={isDraft ? draftName : (effectiveStack?.name ?? "")}
          isDraft={isDraft}
          nameEditable={isDraft}
          onNameChange={setDraftName}
          labels={(isDraft ? draftLabels : effectiveStack?.labels) ?? []}
          labelsEditable={isDraft}
          onAddLabel={addDraftLabel}
          onRemoveLabel={removeDraftLabel}
```

- [ ] **Step 6: Types + lint**

Run: `pnpm --prefix frontend exec tsc -b` — no new errors.
Run: `pnpm --prefix frontend lint` — clean for touched files.

- [ ] **Step 7: Manual Playwright — name a blank draft and save**

`/stacks/new` → type a name in the header → add a resource via "+ Add resource" → fill its image → click Save.
Expected: navigates to `/stacks/<uuid>`; the new stack renders with the typed name; no form ever shown.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/CanvasEditorShell.tsx \
  frontend/src/pages/stacks/components/canvas/tests/canvas-editor-shell.test.tsx \
  frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(stacks): editable stack name + labels in the canvas header"
```

---

## Task 3: Inline volume creation in the resource drawer

**Files:**
- Create: `frontend/src/pages/stacks/lib/canvas/inline-volume.ts`
- Test: `frontend/src/pages/stacks/lib/canvas/tests/inline-volume.test.ts`
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx` (optional `onCreateVolume`)
- Modify: `frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx` (provide `onCreateVolume`)

**Interfaces:**
- Produces:
  - `interface InlineVolumeInput { name: string; size: string; targetPath: string }`
  - `function addInlineVolume(volumes: FormVolumeExtendedData[], mounts: VolumeMount[], input: InlineVolumeInput): { volumes: FormVolumeExtendedData[]; mounts: VolumeMount[] }` where `VolumeMount = NonNullable<FormStackResourceData["volume_mounts"]>[number]`.
- Consumes: `getDefaultVolume()` shape (name/size/access_mode) — replicated purely here (defaults `access_mode: "ReadWriteOnce"`, `needs_sync_before_use: false`).

- [ ] **Step 1: Write the failing pure-reducer test**

Create `frontend/src/pages/stacks/lib/canvas/tests/inline-volume.test.ts`:

```ts
import { describe, it, expect } from "vitest";
import { addInlineVolume } from "../inline-volume";

describe("addInlineVolume", () => {
  it("appends a stack volume and a matching resource mount", () => {
    const { volumes, mounts } = addInlineVolume([], [], { name: "data", size: "2Gi", targetPath: "/var/lib/data" });
    expect(volumes).toEqual([
      {
        name: "data",
        sourceType: "None",
        labels: [],
        spec: { size: "2Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false },
      },
    ]);
    expect(mounts).toEqual([{ source_volume_name: "data", source_sub_path: "", target_path: "/var/lib/data" }]);
  });

  it("does not clobber existing volumes/mounts", () => {
    const { volumes, mounts } = addInlineVolume(
      [{ name: "a", sourceType: "None", labels: [], spec: { size: "1Gi", access_mode: "ReadWriteOnce", needs_sync_before_use: false } }] as never,
      [{ source_volume_name: "a", source_sub_path: "", target_path: "/a" }] as never,
      { name: "b", size: "5Gi", targetPath: "/b" },
    );
    expect(volumes.map((v) => v.name)).toEqual(["a", "b"]);
    expect(mounts.map((m) => m.target_path)).toEqual(["/a", "/b"]);
  });
});
```

- [ ] **Step 2: Run it — fails**

Run: `pnpm --prefix frontend test:run src/pages/stacks/lib/canvas/tests/inline-volume.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the pure reducer**

Create `frontend/src/pages/stacks/lib/canvas/inline-volume.ts`:

```ts
import type {
  FormStackResourceData,
  FormVolumeExtendedData,
} from "@/pages/stacks/schemas/form-schema";

type VolumeMount = NonNullable<FormStackResourceData["volume_mounts"]>[number];

export interface InlineVolumeInput {
  name: string;
  size: string;
  targetPath: string;
}

/** Create a stack-level volume and the resource mount that references it, in one step. */
export function addInlineVolume(
  volumes: FormVolumeExtendedData[],
  mounts: VolumeMount[],
  input: InlineVolumeInput,
): { volumes: FormVolumeExtendedData[]; mounts: VolumeMount[] } {
  const volume = {
    name: input.name,
    sourceType: "None" as const,
    labels: [],
    spec: { size: input.size, access_mode: "ReadWriteOnce", needs_sync_before_use: false },
  } as unknown as FormVolumeExtendedData;
  const mount = { source_volume_name: input.name, source_sub_path: "", target_path: input.targetPath } as VolumeMount;
  return { volumes: [...volumes, volume], mounts: [...mounts, mount] };
}
```

- [ ] **Step 4: Run the test — passes**

Run: `pnpm --prefix frontend test:run src/pages/stacks/lib/canvas/tests/inline-volume.test.ts`
Expected: PASS (2 tests).

- [ ] **Step 5: Add an optional `onCreateVolume` path to the config tab**

In `stack-resource-configuration-tab.tsx`, add to `StackResourceConfigurationTabProps`:

```tsx
  /** When provided, the drawer offers inline volume creation (name+size+path)
   *  instead of only selecting a pre-existing volume. */
  onCreateVolume?: (input: { name: string; size: string; targetPath: string }) => void;
```

Destructure it in the impl signature. Replace the "Add mount" button block (lines ~661-673) so that when `onCreateVolume` is present it renders an inline create form (name + size + path) that calls `onCreateVolume`; otherwise it keeps the existing disabled-when-empty "Add mount" button:

```tsx
          {onCreateVolume ? (
            <InlineVolumeAdder onCreate={onCreateVolume} />
          ) : (
            <div>
              <Button variant="ghost" size="sm" onClick={addVolumeMount} disabled={(volumes || []).length === 0}>
                <PlusCircle className="h-4 w-4 mr-2" />Add mount
              </Button>
              {(volumes || []).length === 0 && (
                <p className="text-sm text-muted-foreground mt-2">No volumes available.</p>
              )}
            </div>
          )}
```

Add the small local component at the bottom of the file (above the `React.memo` export):

```tsx
function InlineVolumeAdder({ onCreate }: { onCreate: (i: { name: string; size: string; targetPath: string }) => void }) {
  const [name, setName] = React.useState("");
  const [size, setSize] = React.useState("1Gi");
  const [targetPath, setTargetPath] = React.useState("/mnt/data");
  const canAdd = name.trim() !== "" && size.trim() !== "" && targetPath.trim() !== "";
  return (
    <div className="grid grid-cols-1 md:grid-cols-[1fr_1fr_1fr_auto] gap-4 items-end border border-dashed p-3 rounded-md">
      <FieldShell label="Volume name" htmlFor="inline-vol-name">
        <Input id="inline-vol-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g., data" />
      </FieldShell>
      <FieldShell label="Size" htmlFor="inline-vol-size">
        <Input id="inline-vol-size" value={size} onChange={(e) => setSize(e.target.value)} placeholder="e.g., 1Gi" />
      </FieldShell>
      <FieldShell label="Mount path" htmlFor="inline-vol-path">
        <Input id="inline-vol-path" value={targetPath} onChange={(e) => setTargetPath(e.target.value)} placeholder="/mnt/data" />
      </FieldShell>
      <Button
        variant="ghost" size="sm" disabled={!canAdd}
        onClick={() => { onCreate({ name: name.trim(), size: size.trim(), targetPath: targetPath.trim() }); setName(""); setSize("1Gi"); setTargetPath("/mnt/data"); }}
      >
        <PlusCircle className="h-4 w-4 mr-2" />Add volume
      </Button>
    </div>
  );
}
```

- [ ] **Step 6: Thread `onCreateVolume` from the drawer**

In `ResourceDrawer.tsx`, build the callback from `session.updateVolumes` + `onChange` (patch this resource's mounts) using `addInlineVolume`, and pass it into `configurationProps`. Add import:

```tsx
import { addInlineVolume } from "@/pages/stacks/lib/canvas/inline-volume";
```

Add the handler inside the component (after `onChange`):

```tsx
  const onCreateVolume = useCallback(
    (input: { name: string; size: string; targetPath: string }) => {
      const currentMounts = session.draft.resources[resourceIndex]?.volume_mounts ?? [];
      const { volumes, mounts } = addInlineVolume(
        session.draft.volumes as never,
        currentMounts as never,
        input,
      );
      session.updateVolumes(() => volumes as never);
      session.updateResources((prev) =>
        prev.map((r, i) => (i === resourceIndex ? { ...r, volume_mounts: mounts } : r)),
      );
    },
    [session, resourceIndex],
  );
```

`useResourceTabProps` builds `configurationProps` from the passed `context`. Thread `onCreateVolume` through by adding it to the `context` object passed to `useResourceTabProps` (add `onCreateVolume` to that call), and in `use-resource-tab-props.ts` include `onCreateVolume: context.onCreateVolume` on the returned `configurationProps`. (Check `use-resource-tab-props.ts` for the `configurationProps` assembly and add the field.)

- [ ] **Step 7: Run the full unit suite for touched areas**

Run: `pnpm --prefix frontend test:run src/pages/stacks/lib/canvas src/pages/stacks/components/shared`
Expected: PASS (no regressions; existing config-tab tests still green since `onCreateVolume` is optional).

Run: `pnpm --prefix frontend exec tsc -b` — no new errors.

- [ ] **Step 8: Manual Playwright — author a volume in the drawer**

`/stacks/new` → add a Postgres block → click its node → Configuration tab → "Add volume" (name `data`, size `2Gi`, path `/var/lib/postgresql/data`) → Save.
Expected: node card shows the volume; saved stack has one volume + the mount.

- [ ] **Step 9: Commit**

```bash
git add frontend/src/pages/stacks/lib/canvas/inline-volume.ts \
  frontend/src/pages/stacks/lib/canvas/tests/inline-volume.test.ts \
  frontend/src/pages/stacks/components/shared/stack-resource-configuration-tab.tsx \
  frontend/src/pages/stacks/components/shared/hooks/use-resource-tab-props.ts \
  frontend/src/pages/stacks/components/canvas/ResourceDrawer.tsx
git commit -m "feat(stacks): author volumes inline in the resource drawer"
```

---

## Task 4: Managed addons pickable from "+ Add resource"

**Files:**
- Modify: `frontend/src/pages/stacks/components/canvas/AddResourcePopover.tsx`
- Modify: `frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx` (thread props)
- Modify: `frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx` (`onLinkAddon` → `session.setLinkedAddonIds`)
- Test: `frontend/src/pages/stacks/components/canvas/tests/add-resource-popover.test.tsx`

**Interfaces:**
- Produces (new `AddResourcePopoverProps` fields): `addons: { id: string; name: string }[]`, `linkedAddonIds: ReadonlySet<string>`, `onLinkAddon: (addonId: string) => void`.
- Consumes: `usePostgresAddons()` → `{ addons: PostgresAddon[] }` (already available in `StackCanvasTab` via props `addonNameById`; pass the raw list too).

- [ ] **Step 1: Write the failing popover test**

Create `frontend/src/pages/stacks/components/canvas/tests/add-resource-popover.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { AddResourcePopover } from "../AddResourcePopover";

describe("AddResourcePopover managed addons", () => {
  it("lists linked-able addons and links on click", () => {
    const onLinkAddon = vi.fn();
    render(
      <AddResourcePopover
        addedIds={[]}
        onAdd={() => {}}
        addons={[{ id: "a1", name: "prod-db" }]}
        linkedAddonIds={new Set()}
        onLinkAddon={onLinkAddon}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: /Add resource/i }));
    fireEvent.click(screen.getByRole("button", { name: /prod-db/i }));
    expect(onLinkAddon).toHaveBeenCalledWith("a1");
  });
});
```

- [ ] **Step 2: Run it — fails**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/canvas/tests/add-resource-popover.test.tsx`
Expected: FAIL — `addons` prop not accepted / no "prod-db" button.

- [ ] **Step 3: Add the Managed add-ons section**

In `AddResourcePopover.tsx`, extend props and render a section under the `BlockPicker`. New imports: `import { Check, Plus, Search } from "lucide-react";` and `import { AddonTypeIcon } from "@/pages/addons/components/addon-type-icon";`.

```tsx
interface AddResourcePopoverProps {
  addedIds: string[];
  onAdd: (blockId: string) => void;
  addons: { id: string; name: string }[];
  linkedAddonIds: ReadonlySet<string>;
  onLinkAddon: (addonId: string) => void;
}
```

Inside `PopoverContent`, after the `BlockPicker` block, add (respect the search `query`):

```tsx
          {addons.length > 0 && (
            <div className="mt-5">
              <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
                Managed add-ons
              </div>
              <div className="grid grid-cols-2 gap-2.5">
                {addons
                  .filter((a) => !query.trim() || a.name.toLowerCase().includes(query.trim().toLowerCase()))
                  .map((a) => {
                    const linked = linkedAddonIds.has(a.id);
                    return (
                      <button
                        type="button" key={a.id} onClick={() => onLinkAddon(a.id)}
                        className="flex min-h-[60px] items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary"
                      >
                        <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                          <AddonTypeIcon type="postgres" size={18} />
                        </span>
                        <span className="min-w-0 flex-1">
                          <span className="block text-sm font-medium text-foreground">{a.name}</span>
                          <span className="block truncate font-mono text-[11px] text-muted-foreground">managed postgres</span>
                        </span>
                        {linked ? <Check className="h-[17px] w-[17px] text-success" /> : <Plus className="h-[17px] w-[17px] text-primary" />}
                      </button>
                    );
                  })}
              </div>
            </div>
          )}
```

- [ ] **Step 4: Run the popover test — passes**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/canvas/tests/add-resource-popover.test.tsx`
Expected: PASS (1 test).

- [ ] **Step 5: Thread props through CanvasEditor**

In `CanvasEditor.tsx`, add to `CanvasEditorProps`: `addons`, `linkedAddonIds`, `onLinkAddon` (same types as the popover). Pass them into `<AddResourcePopover ... addons={addons} linkedAddonIds={linkedAddonIds} onLinkAddon={onLinkAddon} />`.

- [ ] **Step 6: Provide addons + link handler from StackCanvasTab**

In `StackCanvasTab.tsx`, add `usePostgresAddons` (import it) or accept an `addons` prop. Simplest: import and use inside `StackCanvasFlow`:

```tsx
import { usePostgresAddons } from "@/pages/addons/hooks/use-postgres-addons";
```

Inside `StackCanvasFlow`:

```tsx
  const { addons: allAddons } = usePostgresAddons();
  const pickableAddons = useMemo(
    () => allAddons.filter((a) => a.id && a.name).map((a) => ({ id: a.id!, name: a.name! })),
    [allAddons],
  );
  const onLinkAddon = useCallback(
    (addonId: string) => {
      if (!session.isActive) {
        session.start(
          { resources: baselineResources, volumes: baselineVolumes },
          { linkedAddonIds: new Set([...connectionAddonIds, addonId]) },
        );
      } else {
        session.setLinkedAddonIds((prev) => new Set(prev).add(addonId));
      }
    },
    [session, baselineResources, baselineVolumes, connectionAddonIds],
  );
```

Pass to `<CanvasEditor ... addons={pickableAddons} linkedAddonIds={linkedAddonIds} onLinkAddon={onLinkAddon} />`.

- [ ] **Step 7: Suite + types**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/canvas`
Expected: PASS.
Run: `pnpm --prefix frontend exec tsc -b` — no new errors.

- [ ] **Step 8: Manual Playwright — link an addon on a blank draft**

`/stacks/new` → "+ Add resource" → scroll to "Managed add-ons" → pick one → confirm an addon node renders → add a web resource that binds it → Save.
Expected: saved stack has the addon connection.

- [ ] **Step 9: Commit**

```bash
git add frontend/src/pages/stacks/components/canvas/AddResourcePopover.tsx \
  frontend/src/pages/stacks/components/canvas/CanvasEditor.tsx \
  frontend/src/pages/stacks/components/canvas/StackCanvasTab.tsx \
  frontend/src/pages/stacks/components/canvas/tests/add-resource-popover.test.tsx
git commit -m "feat(stacks): pick managed addons from the canvas + Add resource popover"
```

---

## Task 5: Wizard rewire — every path opens `/stacks/new` with a seed

**Files:**
- Modify: `frontend/src/pages/stacks/components/wizard/block-composer.tsx`
- Modify: `frontend/src/pages/stacks/hooks/use-template-import.ts`
- Modify: `frontend/src/pages/stacks/hooks/use-docker-compose-import.ts`
- Modify: `frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx` (blank slate)
- Modify: `frontend/src/components/nav-stacks.tsx`
- Test: `frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx` (update existing)

**Interfaces:**
- Consumes: `DraftSeed` (Task 1). Every path builds `{ seed: DraftSeed }` and calls `navigate("/stacks/new", { state: { seed } })`.
- The seed's `name` may be pre-filled (template name); blank/blocks leave it `""` (named in canvas).

- [ ] **Step 1: Update the block-composer test to expect `/stacks/new`**

In `block-composer.test.tsx`, replace assertions that expect a `createStack` call / navigation to `/stacks/:id` with an expectation that `Continue` navigates to `/stacks/new` carrying a seed. Core assertion:

```tsx
// after adding a block and clicking Continue:
expect(navigateMock).toHaveBeenCalledWith(
  "/stacks/new",
  expect.objectContaining({
    state: expect.objectContaining({
      seed: expect.objectContaining({
        resources: expect.arrayContaining([expect.objectContaining({ name: expect.any(String) })]),
        linkedAddonIds: expect.any(Array),
      }),
    }),
  }),
);
```

(Match the file's existing `useNavigate` mock pattern; if it mocks `react-router-dom`, keep that mock.)

- [ ] **Step 2: Run it — fails**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/wizard/tests/block-composer.test.tsx`
Expected: FAIL — composer still calls the old create path.

- [ ] **Step 3: Simplify block-composer to seed-and-navigate**

In `block-composer.tsx`: remove imports `createStack`, `convertFormStackToApiStack`, `FormStackSchema`, `FormStackData`, `useResourceTeams`, `getCurrentOrganizationId`, `useToast`; remove the `creating` state, the Stack-name `Input` block (lines ~192-203), `createAndOpen`, and `openFormEditor`. Replace with:

```tsx
import { emptyDraftSeed } from "@/pages/stacks/lib/canvas/draft-seed";
// ...
  const openCanvas = () => {
    const seed = {
      ...emptyDraftSeed(),
      resources: stack.spec.stack_resources as never,
      volumes: (stack.spec.volumes ?? []) as never,
      labels: stack.labels,
      linkedAddonIds: Array.from(selectedAddonIds),
    };
    onClose();
    navigate("/stacks/new", { state: { seed } });
  };
```

Update `WizardFooter`:

```tsx
      <WizardFooter
        onBack={onBack}
        onContinue={openCanvas}
        continueDisabled={count === 0}
        hint="Open in the canvas editor"
      />
```

(The right-panel "Stack name" field is removed — naming happens in the canvas. Keep the "Your stack so far" list.)

- [ ] **Step 4: Run the composer test — passes**

Run: `pnpm --prefix frontend test:run src/pages/stacks/components/wizard/tests/block-composer.test.tsx`
Expected: PASS.

- [ ] **Step 5: Template import → `/stacks/new`**

In `use-template-import.ts`, replace the `useTemplate` body:

```ts
  const useTemplate = (template: Template) => {
    const { data } = templateToFormData(template);
    setIsDialogOpen(false);
    navigate("/stacks/new", {
      state: {
        seed: {
          name: data.name ?? "",
          labels: data.labels ?? [],
          resources: data.spec?.stack_resources ?? [],
          volumes: data.spec?.volumes ?? [],
          linkedAddonIds: [],
        },
      },
    });
  };
```

- [ ] **Step 6: Docker compose import → `/stacks/new`**

In `use-docker-compose-import.ts`, replace the `navigate('/stacks/create', {...})` (lines ~69-75) with:

```ts
      setIsDialogOpen(false);
      navigate("/stacks/new", {
        state: {
          seed: {
            name: conversionResult.data.name ?? "",
            labels: conversionResult.data.labels ?? [],
            resources: conversionResult.data.spec?.stack_resources ?? [],
            volumes: conversionResult.data.spec?.volumes ?? [],
            linkedAddonIds: [],
          },
        },
      });
```

- [ ] **Step 7: Blank slate → `/stacks/new`**

In `stack-create-wizard.tsx`, change the blank-slate handler (line ~74-77):

```tsx
                onPickBlank={() => {
                  navigate("/stacks/new");
                  close();
                }}
```

- [ ] **Step 8: Nav active-state**

In `nav-stacks.tsx:22`, change `!location.pathname.includes('/stacks/create')` → `!location.pathname.includes('/stacks/new')`.

- [ ] **Step 9: Types + lint + full suite**

Run: `pnpm --prefix frontend exec tsc -b` — no new errors.
Run: `pnpm --prefix frontend lint` — clean for touched files.
Run: `pnpm --prefix frontend test:run` — full suite green.

- [ ] **Step 10: Manual Playwright — all four live paths**

For each of Build-from-blocks, From-template, Docker-compose, Blank-slate: open the wizard, complete the path, confirm it lands on `/stacks/new` with the seeded graph (template/compose show their resources; blank shows empty), name it, Save → `/stacks/:id`.

- [ ] **Step 11: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/block-composer.tsx \
  frontend/src/pages/stacks/hooks/use-template-import.ts \
  frontend/src/pages/stacks/hooks/use-docker-compose-import.ts \
  frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx \
  frontend/src/components/nav-stacks.tsx \
  frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx
git commit -m "feat(stacks): every wizard path opens the draft canvas at /stacks/new"
```

---

## Task 6: Delete the form + the feature flag; draft ops-tab placeholders

**Files:**
- Create: `frontend/src/pages/stacks/components/canvas/DraftTabPlaceholder.tsx`
- Delete: `frontend/src/pages/stacks/components/create/` (dir + tests), `frontend/src/lib/feature-flags.ts`, `frontend/src/lib/tests/feature-flags.test.ts`
- Modify: `frontend/src/App.tsx` (remove route + import), `frontend/src/components/app-layout.tsx` (drop flag), `frontend/src/pages/stacks/components/detail/index.tsx` (remove flag-OFF branch + `isCanvasEnabled`, use draft placeholders)

**Interfaces:**
- Consumes: nothing new. This task removes dead code and gates ops-tab bodies on `isDraft`.

- [ ] **Step 1: Add the draft ops-tab placeholder**

Create `frontend/src/pages/stacks/components/canvas/DraftTabPlaceholder.tsx`:

```tsx
import { Save } from "lucide-react";

/** Shown for Deployments/Logs/Metrics while the stack is an unsaved draft. */
export function DraftTabPlaceholder({ label }: { label: string }) {
  return (
    <div className="flex h-full flex-col items-center justify-center gap-2 py-24 text-center">
      <Save className="size-5 text-muted-foreground" aria-hidden />
      <p className="text-sm font-medium text-foreground">{label} available after you save</p>
      <p className="text-[13px] text-muted-foreground">Save this stack to create it, then deploy to see live data.</p>
    </div>
  );
}
```

- [ ] **Step 2: Gate ops bodies on `isDraft` in the detail page**

In `detail/index.tsx`, import the placeholder and swap the three ops bodies (lines ~612-650) to render it in draft:

```tsx
import { DraftTabPlaceholder } from "@/pages/stacks/components/canvas/DraftTabPlaceholder";
// ...
  const deploymentsBody = isDraft ? <DraftTabPlaceholder label="Deployments" /> : (stackToShow?.id ? (/* existing DeploymentsTab */) : (/* existing fallback */));
  const logsBody = isDraft ? <DraftTabPlaceholder label="Logs" /> : (/* existing */);
  const metricsBody = isDraft ? <DraftTabPlaceholder label="Metrics" /> : (/* existing */);
```

- [ ] **Step 3: Remove the feature flag and flag-OFF form branch**

In `detail/index.tsx`: remove `import { isCanvasEnabled } from "@/lib/feature-flags";`. Change `if (isCanvasEnabled()) {` (line ~688) to always take the shell path — delete the guard and delete the entire flag-OFF `return (...)` block below it (lines ~733-end, the `<div className="p-8 space-y-8">` form with `StackResourcesForm`/`StackVolumesForm`/`AddonsInStackPanel` accordions and the flag-OFF `TabsContent`s). Remove now-unused imports (`Tabs`/`TabsContent`/`TabsList`/`TabsTrigger`, `StackResourcesForm`, `StackVolumesForm`, `StackResourcesDetail`, `StackVolumesDetail`, `PageHeader`, `StickyActionBar` if unused, `AddonsInStackPanel` if unused) — run lint to find them.

In `app-layout.tsx`: remove `import { isCanvasEnabled } from "@/lib/feature-flags";` and simplify:

```tsx
  const isFullBleed = /^\/stacks\/[^/]+$/.test(location.pathname);
```

- [ ] **Step 4: Delete the form + flag files**

```bash
git rm -r frontend/src/pages/stacks/components/create
git rm frontend/src/lib/feature-flags.ts frontend/src/lib/tests/feature-flags.test.ts
```

Remove the `/stacks/create` route + `StackCreatePage` import in `App.tsx` (lines 6 and 48). Grep for any remaining references:

```bash
grep -rn "isCanvasEnabled\|stacks/create\|components/create\|StackCreatePage\|ImportSource\|isPrefillSource" frontend/src
```

Resolve each hit: delete `ImportSource`/`isPrefillSource` plumbing if only the form used it (`frontend/src/pages/stacks/lib/import-source.ts` + its consumers) — keep only if a seed path still tags provenance (it does not, per Task 5). Remove `importSource`/`importWarnings`/`importedData` handling that pointed at the create page.

- [ ] **Step 5: Types + lint + full suite**

Run: `pnpm --prefix frontend exec tsc -b`
Expected: no new errors (only the pre-existing `postgres-backups.ts` one, if still present).
Run: `pnpm --prefix frontend lint`
Expected: clean (no unused imports left behind).
Run: `pnpm --prefix frontend test:run`
Expected: full suite green (feature-flag tests are gone; block-composer + others updated).

- [ ] **Step 6: Manual Playwright — the form is unreachable**

- Navigate directly to `/stacks/create`.
  Expected: NotFound page (route removed) — the whole-form view never renders.
- Run every wizard path once more end-to-end; confirm each lands in the canvas and none ever shows the accordion form.
- Open an existing stack (`/stacks/d3e497e8-2ec4-4f24-866c-dc6152dee9fa`) → confirm the canvas still renders and Save (edit) still works.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "refactor(stacks): delete the legacy form + canvas feature flag; canvas is the only surface"
```

---

## Self-Review

**Spec coverage**
- Draft-in-canvas / `/stacks/new` / Save=create → Task 1. ✅
- Name + labels in header → Task 2. ✅
- Volumes authored in the drawer → Task 3. ✅
- Managed addon in "+ Add" → Task 4. ✅
- Wizard rewire (blocks/template/compose/blank) + nav → Task 5. ✅
- Form deletion + flag removal + full-bleed `/stacks/new` + draft ops placeholders → Task 6 (+ app-layout in Task 1). ✅
- Error handling (validation stays in draft, createStack failure stays in draft, no form fallback) → Task 1 performSave fork. ✅
- Deferred (rename existing, non-postgres addons, template warnings) → left out by design, noted in spec. ✅

**Placeholder scan** — Task 6 steps 2/4 reference "existing DeploymentsTab / existing fallback" rather than repeating those large JSX blocks; that is a deliberate "keep the existing body, wrap in an `isDraft ?` ternary" instruction, not a missing implementation. All new code (helpers, components, tests) is shown in full.

**Type consistency** — `DraftSeed` fields (`name`, `labels`, `resources`, `volumes`, `linkedAddonIds`) are used identically in Tasks 1 and 5. `buildDraftFormData(name, labels, resources, volumes)` signature matches its call in Task 1 performSave. `addInlineVolume(volumes, mounts, input)` matches its drawer call in Task 3. `onLinkAddon(addonId: string)` consistent across Tasks 4. `onCreateVolume({name,size,targetPath})` consistent between config-tab prop, drawer handler, and `InlineVolumeInput`.

**Risk notes for the implementer**
- `detail/index.tsx` is large and heavily branched; make the Task 1 edits surgically and run `tsc -b` after each sub-step (a-h) rather than all at once.
- `use-resource-tab-props.ts` must forward `onCreateVolume` into `configurationProps` (Task 3 step 6) or the drawer's inline adder won't fire — verify by reading that file before editing.
- Keep `session.start` empty-baseline invariant in draft (Task 1 step 7d): baseline must stay `{resources:[],volumes:[]}` so seeded items read as dirty/added and Save enables.
