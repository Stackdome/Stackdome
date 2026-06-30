# Stack Creation Wizard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a guided stack-creation wizard — a chooser modal + block composer — that runs before the existing form page, killing blank-slate paralysis by letting users assemble a stack from recognizable building blocks (and folding template + docker-compose import into the same modal).

**Architecture:** A single `StackCreateWizard` Radix `Dialog`, launched from the Stacks list "New Stack" button, with an internal phase machine (`chooser | composer | template | compose`). Every path produces a `Partial<FormStackData>` and hands off to the **unchanged** existing form at `/stacks/create` via `navigate(..., { state: { importedData, importSource } })` — the same prefill mechanism templates/compose already use. Blocks are partial `FormStackResourceData` built by reusing the existing docker-compose → form pipeline. The canvas editor, the ResourceConfig restyle, auto-wiring, and Worker/Cron/Job blocks are explicitly **out of scope** (separate efforts).

**Tech Stack:** React 19, Vite, TypeScript, Tailwind v4, Radix UI (shadcn-style `components/ui/`), `components/branded/` primitives, React Router v6, Zod, Vitest + @testing-library/react.

## Global Constraints

- **No magic strings.** Use defined constants for import sources, block ids, env-var `from` values, etc. Add a constant if none exists. (Repo rule — applies to prod + tests.)
- **Brand system.** Use `components/ui/` + `components/branded/` primitives and existing Tailwind/`index.css` tokens (`bg-card`, `border`, `text-muted-foreground`, `bg-primary`, `text-primary-foreground`). No raw hex, no off-scale type. Amber `#f97316` is the primary accent and already maps to `primary`.
- **Land in the existing form.** Do NOT modify `components/create/index.tsx`'s form rendering. The wizard only produces `importedData` + `importSource`. The form is the shared destination.
- **Scope = wizard only.** No canvas, no react-flow, no ResourceConfig restyle, no auto-wiring, no Worker/Cron/Job blocks. v1 catalog = Web (Service) + Postgres/Redis/MySQL/MongoDB (image+port+volume+env) + Custom.
- **Tests:** Vitest + @testing-library/react. Add `// @vitest-environment jsdom` pragma at the top of every component/hook test. Tests live in a sibling `tests/` dir, import via `../name`, cross-module via `@/`. Mock `react-router-dom`'s `useNavigate` with `vi.mock`. Shim `Element.prototype.scrollIntoView = vi.fn()` in `beforeAll` for Radix component tests.
- **Branch:** All work on `stack-creation-redesign` (create at Task 0).
- **Copy is verbatim from the design.** Exact strings are given per task; do not paraphrase. Source of truth: `scratchpad/design/DESIGN-NOTES.md` + the `.dc.html` files.

---

## File Structure

**New:**
- `frontend/src/pages/stacks/data/blocks/types.ts` — `BlockCategory`, `BlockId` consts, `BlockPreset` interface.
- `frontend/src/pages/stacks/data/blocks/registry.ts` — `blockCatalog` (v1 blocks) + `getBlockById`.
- `frontend/src/pages/stacks/lib/block-to-form.ts` — `blockToResources(block)`, `addBlockToStack(stack, block)` (accumulate + dedupe names).
- `frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx` — the modal shell + phase machine.
- `frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx` — Step-0 chooser (5 tiles).
- `frontend/src/pages/stacks/components/wizard/block-composer.tsx` — composer phase (palette + "your stack so far" + Open editor). Contains a small in-file `StackSoFar` list.
- `frontend/src/pages/stacks/components/wizard/block-picker.tsx` — palette (categories → cards).
- `frontend/src/pages/stacks/components/wizard/templates-browser-panel.tsx` — extracted inner UI of the old templates dialog (no Dialog chrome).
- `frontend/src/pages/stacks/components/wizard/docker-compose-import-panel.tsx` — extracted inner UI of the old compose dialog (no Dialog chrome).
- Sibling `tests/` dirs for each of the above with logic/components.

**Modified:**
- `frontend/src/pages/stacks/lib/import-source.ts` — add `Blocks` source + include in prefill set.
- `frontend/src/pages/stacks/components/list/index.tsx` — replace "New Stack" button + `DockerComposeImportDropdown` + the two dialogs with a single "New Stack" button that opens `StackCreateWizard`.

**Deleted (Task 9 cleanup, after wizard fully replaces them):**
- `frontend/src/pages/stacks/components/shared/import-dropdown.tsx` + its test.
- `frontend/src/pages/stacks/components/shared/import-options.tsx` + its test.
- `frontend/src/pages/stacks/components/shared/templates-browser-dialog.tsx` + its test (content moved to `templates-browser-panel.tsx`).
- `frontend/src/pages/stacks/components/shared/docker-compose-import-dialog.tsx` + its test (content moved to `docker-compose-import-panel.tsx`).

---

## Task 0: Branch

- [ ] **Step 1: Create the working branch**

```bash
cd /Users/akshaysasidharan/code/stackdome
git checkout -b stack-creation-redesign
```

- [ ] **Step 2: Commit this plan**

```bash
git add docs/superpowers/plans/2026-06-30-stack-creation-wizard.md
git commit -m "docs: add stack creation wizard implementation plan"
```

---

## Task 1: Add the `Blocks` import source

**Files:**
- Modify: `frontend/src/pages/stacks/lib/import-source.ts`
- Test: `frontend/src/pages/stacks/lib/tests/import-source.test.ts`

**Interfaces:**
- Consumes: nothing.
- Produces: `ImportSource.Blocks === "blocks"`; `isPrefillSource("blocks") === true`. Consumed by Task 4 (composer navigation) and the existing create page.

- [ ] **Step 1: Write the failing test**

```ts
// frontend/src/pages/stacks/lib/tests/import-source.test.ts
import { describe, it, expect } from "vitest";
import { ImportSource, isPrefillSource } from "../import-source";

describe("import-source", () => {
  it("exposes Blocks as a source", () => {
    expect(ImportSource.Blocks).toBe("blocks");
  });

  it("treats blocks, template, and docker-compose as prefill sources", () => {
    expect(isPrefillSource(ImportSource.Blocks)).toBe(true);
    expect(isPrefillSource(ImportSource.Template)).toBe(true);
    expect(isPrefillSource(ImportSource.DockerCompose)).toBe(true);
  });

  it("rejects unknown sources", () => {
    expect(isPrefillSource("github")).toBe(false);
    expect(isPrefillSource(undefined)).toBe(false);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- import-source`
Expected: FAIL — `ImportSource.Blocks` is `undefined`.

- [ ] **Step 3: Add the source**

In `frontend/src/pages/stacks/lib/import-source.ts`, extend the const map and the prefill list:

```ts
export const ImportSource = {
  DockerCompose: "docker-compose",
  Template: "template",
  Blocks: "blocks",
} as const;
export type ImportSource = (typeof ImportSource)[keyof typeof ImportSource];

export const PREFILL_IMPORT_SOURCES: ImportSource[] = [
  ImportSource.DockerCompose,
  ImportSource.Template,
  ImportSource.Blocks,
];

export function isPrefillSource(source: unknown): source is ImportSource {
  return PREFILL_IMPORT_SOURCES.includes(source as ImportSource);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- import-source`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/import-source.ts frontend/src/pages/stacks/lib/tests/import-source.test.ts
git commit -m "feat(stacks): add 'blocks' import source"
```

---

## Task 2: Block catalog types + registry

**Files:**
- Create: `frontend/src/pages/stacks/data/blocks/types.ts`
- Create: `frontend/src/pages/stacks/data/blocks/registry.ts`
- Test: `frontend/src/pages/stacks/data/blocks/tests/registry.test.ts`

**Interfaces:**
- Consumes: `FormStackData` types (via Task 3's converter, not here).
- Produces:
  - `BlockCategory = "services" | "data"` and a `BLOCK_CATEGORY_META` ordered list `{ id: BlockCategory; label: string; note: string }[]`.
  - `BlockId` const map (`Web`, `Custom`, `Postgres`, `Redis`, `Mysql`, `Mongo`).
  - `BlockPreset` interface: `{ id: string; name: string; category: BlockCategory; icon: string; summary: string; compose?: string }` — `compose` is a 1-service docker-compose YAML for known-software blocks; absent for generic blocks (Web/Custom).
  - `blockCatalog: BlockPreset[]`, `getBlockById(id: string): BlockPreset | undefined`.

- [ ] **Step 1: Write the failing test**

```ts
// frontend/src/pages/stacks/data/blocks/tests/registry.test.ts
import { describe, it, expect } from "vitest";
import { blockCatalog, getBlockById, BlockId, BLOCK_CATEGORY_META } from "../registry";

describe("block registry", () => {
  it("ships exactly the v1 catalog (web, custom, postgres, redis, mysql, mongo)", () => {
    expect(blockCatalog.map((b) => b.id).sort()).toEqual(
      [BlockId.Custom, BlockId.Mongo, BlockId.Mysql, BlockId.Postgres, BlockId.Redis, BlockId.Web].sort(),
    );
  });

  it("gives known-software blocks a compose snippet and generic blocks none", () => {
    expect(getBlockById(BlockId.Postgres)?.compose).toContain("postgres:");
    expect(getBlockById(BlockId.Web)?.compose).toBeUndefined();
    expect(getBlockById(BlockId.Custom)?.compose).toBeUndefined();
  });

  it("only uses the two v1 categories", () => {
    const cats = new Set(blockCatalog.map((b) => b.category));
    expect([...cats].sort()).toEqual(["data", "services"]);
    expect(BLOCK_CATEGORY_META.map((c) => c.id)).toEqual(["services", "data"]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- blocks/tests/registry`
Expected: FAIL — module not found.

- [ ] **Step 3: Write the types**

```ts
// frontend/src/pages/stacks/data/blocks/types.ts
export type BlockCategory = "services" | "data";

export interface BlockCategoryMeta {
  id: BlockCategory;
  label: string; // uppercase marker label
  note: string;  // muted sub-note
}

export const BlockId = {
  Web: "web",
  Custom: "custom",
  Postgres: "postgres",
  Redis: "redis",
  Mysql: "mysql",
  Mongo: "mongo",
} as const;
export type BlockId = (typeof BlockId)[keyof typeof BlockId];

export interface BlockPreset {
  id: string;
  name: string;
  category: BlockCategory;
  icon: string;      // lucide icon name (e.g. "globe", "database", "zap", "box")
  summary: string;   // mono one-liner shown on the card, e.g. "postgres:16 · :5432 · pgdata"
  compose?: string;  // 1-service docker-compose YAML; omitted for generic blocks
}
```

- [ ] **Step 4: Write the registry**

```ts
// frontend/src/pages/stacks/data/blocks/registry.ts
import { BlockId, type BlockCategoryMeta, type BlockPreset } from "./types";

export const BLOCK_CATEGORY_META: BlockCategoryMeta[] = [
  { id: "services", label: "SERVICES", note: "your code, running" },
  { id: "data", label: "DATA STORES", note: "run in your cluster" },
];

export const blockCatalog: BlockPreset[] = [
  { id: BlockId.Web, name: "Web service", category: "services", icon: "globe", summary: "your image · :8080" },
  { id: BlockId.Custom, name: "Custom", category: "services", icon: "box", summary: "empty container shape" },
  {
    id: BlockId.Postgres, name: "Postgres", category: "data", icon: "database", summary: "postgres:16 · :5432 · pgdata",
    compose: [
      "services:",
      "  postgres:",
      "    image: postgres:16",
      '    ports: ["5432:5432"]',
      '    volumes: ["pgdata:/var/lib/postgresql/data"]',
      "    environment:",
      '      POSTGRES_PASSWORD: ""',
      "volumes:",
      "  pgdata: {}",
      "",
    ].join("\n"),
  },
  {
    id: BlockId.Redis, name: "Redis", category: "data", icon: "zap", summary: "redis:7 · :6379",
    compose: ["services:", "  redis:", "    image: redis:7", '    ports: ["6379:6379"]', ""].join("\n"),
  },
  {
    id: BlockId.Mysql, name: "MySQL", category: "data", icon: "database", summary: "mysql:8 · :3306 · mysql-data",
    compose: [
      "services:", "  mysql:", "    image: mysql:8", '    ports: ["3306:3306"]',
      '    volumes: ["mysql-data:/var/lib/mysql"]', "    environment:", '      MYSQL_ROOT_PASSWORD: ""',
      "volumes:", "  mysql-data: {}", "",
    ].join("\n"),
  },
  {
    id: BlockId.Mongo, name: "MongoDB", category: "data", icon: "database", summary: "mongo:7 · :27017 · mongo-data",
    compose: [
      "services:", "  mongo:", "    image: mongo:7", '    ports: ["27017:27017"]',
      '    volumes: ["mongo-data:/data/db"]', "volumes:", "  mongo-data: {}", "",
    ].join("\n"),
  },
];

export function getBlockById(id: string): BlockPreset | undefined {
  return blockCatalog.find((b) => b.id === id);
}

export { BlockId } from "./types";
```

- [ ] **Step 5: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- blocks/tests/registry`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/stacks/data/blocks/
git commit -m "feat(stacks): add v1 block catalog (web, custom, postgres, redis, mysql, mongo)"
```

---

## Task 3: Convert a block into form resources (reuse the compose pipeline)

**Files:**
- Create: `frontend/src/pages/stacks/lib/block-to-form.ts`
- Test: `frontend/src/pages/stacks/lib/tests/block-to-form.test.ts`

**Interfaces:**
- Consumes: `BlockPreset` (Task 2); `parseAndValidateDockerCompose` + `convertDockerComposeToStackData` from `@/pages/stacks/lib/docker-compose-parser` / `@/pages/stacks/lib/docker-compose-converter`; `FormStackData`, `FormStackResourceData`, `FormVolumeData` from `@/pages/stacks/schemas/form-schema`.
- Produces:
  - `blockToResources(block: BlockPreset): { resources: FormStackResourceData[]; volumes: FormVolumeData[] }`.
  - `emptyStack(): Pick<FormStackData, "name" | "labels" | "spec">` with `spec.stack_resources: []`, `spec.volumes: []`.
  - `addBlockToStack(stack, block)` → new stack object with the block's resources/volumes appended and resource names de-duplicated (`postgres`, `postgres-2`, …). Used by Task 4.

- [ ] **Step 1: Write the failing test**

```ts
// frontend/src/pages/stacks/lib/tests/block-to-form.test.ts
import { describe, it, expect } from "vitest";
import { blockToResources, addBlockToStack, emptyStack } from "../block-to-form";
import { getBlockById, BlockId } from "@/pages/stacks/data/blocks/registry";

describe("block-to-form", () => {
  it("converts the postgres block into one resource + one volume", () => {
    const { resources, volumes } = blockToResources(getBlockById(BlockId.Postgres)!);
    expect(resources).toHaveLength(1);
    expect(resources[0].name).toBe("postgres");
    expect(resources[0].image_spec?.image).toBe("postgres:16");
    expect(volumes).toHaveLength(1);
    expect(volumes[0].name).toBe("pgdata");
  });

  it("converts a generic web block into an empty-image resource and no volumes", () => {
    const { resources, volumes } = blockToResources(getBlockById(BlockId.Web)!);
    expect(resources).toHaveLength(1);
    expect(resources[0].name).toBe("web");
    expect(resources[0].sourceType).toBe("image");
    expect(resources[0].image_spec?.image).toBe("");
    expect(volumes).toHaveLength(0);
  });

  it("de-duplicates resource names when the same block is added twice", () => {
    let stack = emptyStack();
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    stack = addBlockToStack(stack, getBlockById(BlockId.Postgres)!);
    expect(stack.spec.stack_resources.map((r) => r.name)).toEqual(["postgres", "postgres-2"]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- block-to-form`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the converter**

```ts
// frontend/src/pages/stacks/lib/block-to-form.ts
import type { BlockPreset } from "@/pages/stacks/data/blocks/types";
import { BlockId } from "@/pages/stacks/data/blocks/types";
import { parseAndValidateDockerCompose } from "@/pages/stacks/lib/docker-compose-parser";
import { convertDockerComposeToStackData } from "@/pages/stacks/lib/docker-compose-converter";
import type {
  FormStackData,
  FormStackResourceData,
  FormVolumeData,
} from "@/pages/stacks/schemas/form-schema";

type WorkingStack = Pick<FormStackData, "name" | "labels" | "spec">;

export function emptyStack(): WorkingStack {
  return { name: "", labels: [], spec: { stack_resources: [], volumes: [] } };
}

/** Generic blocks have no compose snippet — produce a minimal resource skeleton. */
function genericResource(block: BlockPreset): FormStackResourceData {
  const base = {
    name: block.id,
    sourceType: "image" as const,
    image_spec: { image: "" },
    ports: block.id === BlockId.Web ? [{ container_port: 8080, protocol: "TCP" }] : [],
  };
  return base as unknown as FormStackResourceData;
}

export function blockToResources(block: BlockPreset): {
  resources: FormStackResourceData[];
  volumes: FormVolumeData[];
} {
  if (!block.compose) {
    return { resources: [genericResource(block)], volumes: [] };
  }
  const parsed = parseAndValidateDockerCompose(block.compose);
  const result = convertDockerComposeToStackData(parsed);
  if (!result.success || !result.data) {
    throw new Error(`Block "${block.id}" failed to convert: ${result.errors?.[0]?.message ?? "unknown"}`);
  }
  return {
    resources: result.data.spec.stack_resources ?? [],
    volumes: result.data.spec.volumes ?? [],
  };
}

function uniqueName(base: string, taken: Set<string>): string {
  if (!taken.has(base)) return base;
  let i = 2;
  while (taken.has(`${base}-${i}`)) i += 1;
  return `${base}-${i}`;
}

export function addBlockToStack(stack: WorkingStack, block: BlockPreset): WorkingStack {
  const { resources, volumes } = blockToResources(block);
  const takenResources = new Set((stack.spec.stack_resources ?? []).map((r) => r.name));
  const takenVolumes = new Set((stack.spec.volumes ?? []).map((v) => v.name));

  const renamedResources = resources.map((r) => {
    const name = uniqueName(r.name, takenResources);
    takenResources.add(name);
    return { ...r, name };
  });
  const renamedVolumes = volumes.map((v) => {
    const name = uniqueName(v.name, takenVolumes);
    takenVolumes.add(name);
    return { ...v, name };
  });

  return {
    ...stack,
    spec: {
      stack_resources: [...(stack.spec.stack_resources ?? []), ...renamedResources],
      volumes: [...(stack.spec.volumes ?? []), ...renamedVolumes],
    },
  };
}
```

> Note: if `convertDockerComposeToStackData` produces a resource name different from the service key (it `sanitizeKubernetesName`s), the postgres test asserts `"postgres"` which is already RFC-1123 safe. If MySQL/Mongo names diverge, adjust the registry service keys to match — do not weaken the assertion.

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- block-to-form`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/lib/block-to-form.ts frontend/src/pages/stacks/lib/tests/block-to-form.test.ts
git commit -m "feat(stacks): convert blocks to form resources via compose pipeline"
```

---

## Task 4: Block composer phase

**Files:**
- Create: `frontend/src/pages/stacks/components/wizard/block-picker.tsx`
- Create: `frontend/src/pages/stacks/components/wizard/block-composer.tsx`
- Test: `frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx`

**Interfaces:**
- Consumes: `blockCatalog`, `BLOCK_CATEGORY_META`, `getBlockById` (Task 2); `addBlockToStack`, `emptyStack` (Task 3); `ImportSource` (Task 1); `useNavigate`.
- Produces:
  - `BlockPicker` props: `{ catalog: BlockPreset[]; categories: BlockCategoryMeta[]; addedIds: string[]; onAdd: (id: string) => void; query: string }`.
  - `BlockComposer` props: `{ onBack: () => void; onClose: () => void }`. On "Open editor" it calls `navigate("/stacks/create", { state: { importedData, importSource: ImportSource.Blocks } })` then `onClose()`.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { BlockComposer } from "../block-composer";
import { ImportSource } from "@/pages/stacks/lib/import-source";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));
beforeAll(() => { Element.prototype.scrollIntoView = vi.fn(); });
afterEach(() => { cleanup(); navigate.mockReset(); });

describe("BlockComposer", () => {
  it("adds blocks and navigates to the form with importSource=blocks", async () => {
    const user = userEvent.setup();
    render(<BlockComposer onBack={vi.fn()} onClose={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Web service/i }));
    await user.click(screen.getByRole("button", { name: /Postgres/i }));

    // "your stack so far" shows both
    const panel = screen.getByTestId("stack-so-far");
    expect(within(panel).getByText("web")).toBeInTheDocument();
    expect(within(panel).getByText("postgres")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: /Open editor/i }));

    expect(navigate).toHaveBeenCalledTimes(1);
    const [path, opts] = navigate.mock.calls[0];
    expect(path).toBe("/stacks/create");
    expect(opts.state.importSource).toBe(ImportSource.Blocks);
    expect(opts.state.importedData.spec.stack_resources.map((r: { name: string }) => r.name)).toEqual([
      "web",
      "postgres",
    ]);
  });

  it("disables Open editor until at least one block is added", () => {
    render(<BlockComposer onBack={vi.fn()} onClose={vi.fn()} />);
    expect(screen.getByRole("button", { name: /Open editor/i })).toBeDisabled();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- block-composer`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `BlockPicker`**

```tsx
// frontend/src/pages/stacks/components/wizard/block-picker.tsx
import { Plus, Check } from "lucide-react";
import type { BlockCategoryMeta, BlockPreset } from "@/pages/stacks/data/blocks/types";
import { cn } from "@/lib/utils";
import { BlockGlyph } from "./block-glyph";

interface BlockPickerProps {
  catalog: BlockPreset[];
  categories: BlockCategoryMeta[];
  addedIds: string[];
  onAdd: (id: string) => void;
  query: string;
}

export function BlockPicker({ catalog, categories, addedIds, onAdd, query }: BlockPickerProps) {
  const q = query.trim().toLowerCase();
  const match = (b: BlockPreset) => !q || b.name.toLowerCase().includes(q) || b.summary.toLowerCase().includes(q);
  const visible = catalog.filter(match);

  if (visible.length === 0) {
    return (
      <p className="px-1 py-6 text-sm text-muted-foreground">
        No matches for “{query}”
      </p>
    );
  }

  return (
    <div className="flex flex-col gap-5">
      {categories.map((cat) => {
        const blocks = visible.filter((b) => b.category === cat.id);
        if (blocks.length === 0) return null;
        return (
          <div key={cat.id}>
            <div className="mb-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
              {cat.label}
              <span className="font-normal normal-case tracking-normal text-muted-foreground/70"> · {cat.note}</span>
            </div>
            <div className="grid grid-cols-2 gap-2.5">
              {blocks.map((b) => {
                const added = addedIds.includes(b.id);
                return (
                  <button
                    type="button"
                    key={b.id}
                    onClick={() => onAdd(b.id)}
                    className={cn(
                      "flex items-center gap-3 rounded-md border bg-card px-3 py-3 text-left transition-colors hover:border-primary",
                      added && "border-primary/60",
                    )}
                  >
                    <span className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded bg-muted text-muted-foreground">
                      <BlockGlyph icon={b.icon} size={18} />
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block text-sm font-medium text-foreground">{b.name}</span>
                      <span className="block truncate font-mono text-[11px] text-muted-foreground">{b.summary}</span>
                    </span>
                    {added ? <Check className="h-[17px] w-[17px] text-green-500" /> : <Plus className="h-[17px] w-[17px] text-primary" />}
                  </button>
                );
              })}
            </div>
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Add the block glyph helper**

The design uses a `Glyph` component keyed by name. Map block icon names to lucide-react icons (already a dependency, used across the list page).

```tsx
// frontend/src/pages/stacks/components/wizard/block-glyph.tsx
import { Globe, Database, Zap, Box, type LucideIcon } from "lucide-react";

const ICONS: Record<string, LucideIcon> = { globe: Globe, database: Database, zap: Zap, box: Box };

export function BlockGlyph({ icon, size = 18 }: { icon: string; size?: number }) {
  const Icon = ICONS[icon] ?? Box;
  return <Icon style={{ width: size, height: size }} />;
}
```

- [ ] **Step 5: Implement `BlockComposer`**

```tsx
// frontend/src/pages/stacks/components/wizard/block-composer.tsx
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ArrowLeft, ArrowRight, Search, X } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { blockCatalog, BLOCK_CATEGORY_META, getBlockById } from "@/pages/stacks/data/blocks/registry";
import { addBlockToStack, emptyStack } from "@/pages/stacks/lib/block-to-form";
import { ImportSource } from "@/pages/stacks/lib/import-source";
import { BlockPicker } from "./block-picker";
import { BlockGlyph } from "./block-glyph";

interface BlockComposerProps {
  onBack: () => void;
  onClose: () => void;
}

export function BlockComposer({ onBack, onClose }: BlockComposerProps) {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [stack, setStack] = useState(emptyStack);

  // addedIds = block ids whose resource name (or a -N variant) is present
  const addedIds = useMemo(() => {
    const names = new Set(stack.spec.stack_resources.map((r) => r.name));
    return blockCatalog.filter((b) => names.has(b.id) || [...names].some((n) => n.startsWith(`${b.id}-`))).map((b) => b.id);
  }, [stack]);

  const addBlock = (id: string) => {
    const block = getBlockById(id);
    if (block) setStack((s) => addBlockToStack(s, block));
  };

  const removeResource = (index: number) =>
    setStack((s) => ({
      ...s,
      spec: { ...s.spec, stack_resources: s.spec.stack_resources.filter((_, i) => i !== index) },
    }));

  const openEditor = () => {
    navigate("/stacks/create", { state: { importedData: stack, importSource: ImportSource.Blocks } });
    onClose();
  };

  const count = stack.spec.stack_resources.length;

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-3 border-b px-5 py-3">
        <button type="button" onClick={onBack} className="flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> back
        </button>
      </div>

      <div className="grid flex-1 grid-cols-[1fr_360px] gap-0 overflow-hidden">
        {/* LEFT: palette */}
        <div className="overflow-y-auto p-6">
          <div className="mb-1 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">COMPOSE</div>
          <h2 className="mb-1 text-2xl font-medium tracking-tight">What's in your stack?</h2>
          <p className="mb-5 text-sm text-muted-foreground">
            Add blocks below. Known software lands fully configured.
          </p>
          <div className="relative mb-5">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search services, data stores, addons…"
              className="pl-9"
            />
          </div>
          <BlockPicker catalog={blockCatalog} categories={BLOCK_CATEGORY_META} addedIds={addedIds} onAdd={addBlock} query={query} />
        </div>

        {/* RIGHT: your stack so far */}
        <div className="flex flex-col border-l bg-card/40">
          <div className="border-b px-4 py-3 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">
            Your stack so far · {count}
          </div>
          <div data-testid="stack-so-far" className="flex-1 space-y-1.5 overflow-y-auto p-4">
            {count === 0 ? (
              <p className="px-1 py-6 text-sm text-muted-foreground">Pick blocks on the left to start.</p>
            ) : (
              stack.spec.stack_resources.map((r, i) => (
                <div key={`${r.name}-${i}`} className="flex items-center gap-3 rounded border bg-card px-3 py-2">
                  <BlockGlyph icon="box" size={16} />
                  <span className="min-w-0 flex-1">
                    <span className="block text-sm text-foreground">{r.name}</span>
                    <span className="block truncate font-mono text-[11px] text-muted-foreground">
                      {r.image_spec?.image || "configure source"}
                    </span>
                  </span>
                  <button type="button" aria-label={`Remove ${r.name}`} onClick={() => removeResource(i)} className="text-muted-foreground hover:text-foreground">
                    <X className="h-4 w-4" />
                  </button>
                </div>
              ))
            )}
          </div>
          <div className="border-t p-4">
            <Button className="w-full" disabled={count === 0} onClick={openEditor}>
              Open editor <ArrowRight className="ml-1 h-4 w-4" />
            </Button>
            <p className="mt-2 text-center text-xs text-muted-foreground">Review &amp; configure your resources</p>
          </div>
        </div>
      </div>
    </div>
  );
}
```

> Auto-wiring is deferred — the design's "AUTO-WIRED" panel is intentionally omitted, and the subcopy drops "connections get wired automatically".

- [ ] **Step 6: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- block-composer`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/block-picker.tsx frontend/src/pages/stacks/components/wizard/block-glyph.tsx frontend/src/pages/stacks/components/wizard/block-composer.tsx frontend/src/pages/stacks/components/wizard/tests/block-composer.test.tsx
git commit -m "feat(stacks): add block composer phase"
```

---

## Task 5: Wizard chooser (Step 0)

**Files:**
- Create: `frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx`
- Test: `frontend/src/pages/stacks/components/wizard/tests/wizard-chooser.test.tsx`

**Interfaces:**
- Consumes: nothing beyond UI primitives + lucide icons.
- Produces: `WizardChooser` props `{ onPickBlocks: () => void; onPickTemplate: () => void; onPickCompose: () => void; onPickBlank: () => void }`. Renders 5 tiles; the GitHub tile is disabled with a "soon" pill and has no handler.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/wizard-chooser.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { WizardChooser } from "../wizard-chooser";

afterEach(cleanup);

describe("WizardChooser", () => {
  it("routes each start option to its handler", async () => {
    const user = userEvent.setup();
    const onPickBlocks = vi.fn(), onPickTemplate = vi.fn(), onPickCompose = vi.fn(), onPickBlank = vi.fn();
    render(<WizardChooser onPickBlocks={onPickBlocks} onPickTemplate={onPickTemplate} onPickCompose={onPickCompose} onPickBlank={onPickBlank} />);

    await user.click(screen.getByRole("button", { name: /Compose blocks/i }));
    await user.click(screen.getByRole("button", { name: /From template/i }));
    await user.click(screen.getByRole("button", { name: /Docker compose/i }));
    await user.click(screen.getByRole("button", { name: /blank slate/i }));

    expect(onPickBlocks).toHaveBeenCalledOnce();
    expect(onPickTemplate).toHaveBeenCalledOnce();
    expect(onPickCompose).toHaveBeenCalledOnce();
    expect(onPickBlank).toHaveBeenCalledOnce();
  });

  it("renders the GitHub tile disabled with a 'soon' marker", () => {
    render(<WizardChooser onPickBlocks={vi.fn()} onPickTemplate={vi.fn()} onPickCompose={vi.fn()} onPickBlank={vi.fn()} />);
    const github = screen.getByRole("button", { name: /GitHub repo/i });
    expect(github).toBeDisabled();
    expect(screen.getByText(/soon/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- wizard-chooser`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `WizardChooser`**

```tsx
// frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx
import { Grid3x3, LayoutTemplate, GitBranch, Container, Code, ArrowRight, type LucideIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface WizardChooserProps {
  onPickBlocks: () => void;
  onPickTemplate: () => void;
  onPickCompose: () => void;
  onPickBlank: () => void;
}

interface AltStart {
  icon: LucideIcon;
  label: string;
  desc: string;
  onClick?: () => void;
  disabled?: boolean;
  soon?: boolean;
}

export function WizardChooser({ onPickBlocks, onPickTemplate, onPickCompose, onPickBlank }: WizardChooserProps) {
  const alts: AltStart[] = [
    { icon: LayoutTemplate, label: "From template", desc: "A curated, ready-made stack.", onClick: onPickTemplate },
    { icon: GitBranch, label: "GitHub repo", desc: "Auto-detect build & start.", disabled: true, soon: true },
    { icon: Container, label: "Docker compose", desc: "Import a compose.yml.", onClick: onPickCompose },
    { icon: Code, label: "Start from a blank slate", desc: "Build it up yourself.", onClick: onPickBlank },
  ];

  return (
    <div className="p-8">
      <h2 className="mb-1 text-2xl font-medium tracking-tight">How do you want to start?</h2>
      <p className="mb-6 text-sm text-muted-foreground">
        Let's get something running. Pick a starting point — you can change anything later.
      </p>

      {/* Primary tile */}
      <button
        type="button"
        onClick={onPickBlocks}
        className="mb-6 flex w-full items-center gap-4 rounded-lg border border-primary bg-card p-5 text-left transition-colors hover:bg-card/80"
      >
        <span className="flex h-11 w-11 flex-none items-center justify-center rounded bg-primary/10 text-primary">
          <Grid3x3 className="h-5 w-5" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="mb-0.5 flex items-center gap-2">
            <span className="text-base font-medium text-foreground">Build from blocks</span>
            <span className="rounded-full bg-primary/10 px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-primary">
              Recommended
            </span>
          </span>
          <span className="block text-sm text-muted-foreground">
            Assemble from recognizable building blocks — web, Postgres, Redis, workers. Known software lands fully configured.
          </span>
        </span>
        <Button asChild={false} variant="default" className="pointer-events-none flex-none">
          Compose blocks <ArrowRight className="ml-1 h-4 w-4" />
        </Button>
      </button>

      <div className="mb-4 font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">OR START FROM</div>
      <div className="grid grid-cols-2 gap-2.5">
        {alts.map((a) => (
          <button
            type="button"
            key={a.label}
            onClick={a.onClick}
            disabled={a.disabled}
            className={cn(
              "flex items-start gap-3 rounded-md border bg-card p-4 text-left transition-colors",
              a.disabled ? "cursor-not-allowed opacity-50" : "hover:border-primary",
            )}
          >
            <span className="flex h-9 w-9 flex-none items-center justify-center rounded bg-muted text-muted-foreground">
              <a.icon className="h-[18px] w-[18px]" />
            </span>
            <span className="min-w-0 flex-1">
              <span className="mb-0.5 flex items-center gap-2">
                <span className="text-sm font-medium text-foreground">{a.label}</span>
                {a.soon && (
                  <span className="rounded-full border px-1.5 py-0.5 font-mono text-[9px] uppercase tracking-wider text-muted-foreground">
                    soon
                  </span>
                )}
              </span>
              <span className="block text-xs text-muted-foreground">{a.desc}</span>
            </span>
          </button>
        ))}
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- wizard-chooser`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/wizard-chooser.tsx frontend/src/pages/stacks/components/wizard/tests/wizard-chooser.test.tsx
git commit -m "feat(stacks): add wizard chooser step"
```

---

## Task 6: Extract the templates browser panel + template phase

**Files:**
- Create: `frontend/src/pages/stacks/components/wizard/templates-browser-panel.tsx`
- Modify: `frontend/src/pages/stacks/components/shared/templates-browser-dialog.tsx` (re-export the panel inside its Dialog body — keeps the old dialog working until Task 9 deletes it)
- Test: `frontend/src/pages/stacks/components/wizard/tests/templates-browser-panel.test.tsx`

**Interfaces:**
- Consumes: `Template` type, `templates` registry.
- Produces: `TemplatesBrowserPanel` props `{ templates: Template[]; onUse: (template: Template) => void }` — the split-pane content WITHOUT Dialog chrome.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/templates-browser-panel.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { TemplatesBrowserPanel } from "../templates-browser-panel";
import type { Template } from "@/pages/stacks/data/templates/types";

const tooljet: Template = {
  id: "tooljet", name: "ToolJet", initials: "TJ", icon: "box", category: "Website",
  shortDescription: "Low-code", longDescription: "Low-code platform", website: "https://tooljet.com",
  docs: "https://docs.tooljet.com", version: "1.0", stackYaml: "services:\n  tooljet:\n    image: tooljet/tooljet:latest\n",
};
beforeAll(() => { Element.prototype.scrollIntoView = vi.fn(); });
afterEach(cleanup);

describe("TemplatesBrowserPanel", () => {
  it("calls onUse with the selected template", async () => {
    const user = userEvent.setup();
    const onUse = vi.fn();
    render(<TemplatesBrowserPanel templates={[tooljet]} onUse={onUse} />);
    await user.click(screen.getByRole("button", { name: /Use template/i }));
    expect(onUse).toHaveBeenCalledWith(tooljet);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- templates-browser-panel`
Expected: FAIL — module not found.

- [ ] **Step 3: Move the dialog's inner JSX into `TemplatesBrowserPanel`**

Open `components/shared/templates-browser-dialog.tsx`. Cut everything inside `<DialogContent>` (the search input + listbox + detail pane + "Use template" button, plus the local selection/keyboard state) into the new panel component. The panel takes `{ templates, onUse }` and contains all the state. Then have the dialog render the panel:

```tsx
// frontend/src/pages/stacks/components/wizard/templates-browser-panel.tsx
import { useState } from "react";
import type { Template } from "@/pages/stacks/data/templates/types";
// ... (move the existing imports the dialog used for its body: Input, TemplateBadge, ExternalLinkButton, Button, icons)

interface TemplatesBrowserPanelProps {
  templates: Template[];
  onUse: (template: Template) => void;
}

export function TemplatesBrowserPanel({ templates, onUse }: TemplatesBrowserPanelProps) {
  const [selectedId, setSelectedId] = useState(templates[0]?.id ?? "");
  const [query, setQuery] = useState("");
  // ... move the existing filtering, keyboard nav, and split-pane JSX here verbatim,
  // replacing the outer <DialogContent> wrapper with a plain <div className="flex h-full">.
  // The "Use template" button calls onUse(selected).
  // (Preserve all existing behavior and markup from the dialog body.)
  return (/* moved split-pane JSX */ null);
}
```

```tsx
// components/shared/templates-browser-dialog.tsx (now a thin wrapper, deleted in Task 9)
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { TemplatesBrowserPanel } from "@/pages/stacks/components/wizard/templates-browser-panel";
import type { Template } from "@/pages/stacks/data/templates/types";

export default function TemplatesBrowserDialog(props: {
  open: boolean; onOpenChange: (o: boolean) => void; templates: Template[]; onUse: (t: Template) => void;
}) {
  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent className="block gap-0 overflow-hidden p-0 sm:max-w-[1040px]">
        <TemplatesBrowserPanel templates={props.templates} onUse={props.onUse} />
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 4: Run both the new panel test and the existing dialog test**

Run: `pnpm --prefix frontend test -- templates-browser`
Expected: PASS (new panel test + the existing `templates-browser-dialog.test.tsx` still green — behavior unchanged).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/templates-browser-panel.tsx frontend/src/pages/stacks/components/wizard/tests/templates-browser-panel.test.tsx frontend/src/pages/stacks/components/shared/templates-browser-dialog.tsx
git commit -m "refactor(stacks): extract TemplatesBrowserPanel from dialog"
```

---

## Task 7: Extract the docker-compose import panel + compose phase

**Files:**
- Create: `frontend/src/pages/stacks/components/wizard/docker-compose-import-panel.tsx`
- Modify: `frontend/src/pages/stacks/components/shared/docker-compose-import-dialog.tsx` (thin wrapper, deleted in Task 9)
- Test: `frontend/src/pages/stacks/components/wizard/tests/docker-compose-import-panel.test.tsx`

**Interfaces:**
- Consumes: nothing new.
- Produces: `DockerComposeImportPanel` props `{ onImport: (yaml: string) => Promise<void>; isLoading: boolean; error: string | null; onClearError: () => void }` — file upload + paste UI WITHOUT Dialog chrome.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/docker-compose-import-panel.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DockerComposeImportPanel } from "../docker-compose-import-panel";

afterEach(cleanup);

describe("DockerComposeImportPanel", () => {
  it("imports pasted YAML", async () => {
    const user = userEvent.setup();
    const onImport = vi.fn().mockResolvedValue(undefined);
    render(<DockerComposeImportPanel onImport={onImport} isLoading={false} error={null} onClearError={vi.fn()} />);
    await user.type(screen.getByRole("textbox"), "services:\n  web:\n    image: nginx");
    await user.click(screen.getByRole("button", { name: /^Import$/i }));
    expect(onImport).toHaveBeenCalledWith("services:\n  web:\n    image: nginx");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- docker-compose-import-panel`
Expected: FAIL — module not found.

- [ ] **Step 3: Move the dialog's inner JSX into `DockerComposeImportPanel`**

Same pattern as Task 6: cut the file-input + `Textarea` (via `FieldShell`) + error + Import/Cancel buttons + local `yamlContent` state out of `docker-compose-import-dialog.tsx` into the panel. Panel calls `onImport(yamlContent)`. Replace the dialog's body with `<DockerComposeImportPanel {...} />` inside `DialogContent`.

```tsx
// frontend/src/pages/stacks/components/wizard/docker-compose-import-panel.tsx
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { FieldShell } from "@/components/branded";

interface DockerComposeImportPanelProps {
  onImport: (yaml: string) => Promise<void>;
  isLoading: boolean;
  error: string | null;
  onClearError: () => void;
}

export function DockerComposeImportPanel({ onImport, isLoading, error, onClearError }: DockerComposeImportPanelProps) {
  const [yamlContent, setYamlContent] = useState("");
  // ... move the existing hidden file <input type="file"> + FileReader logic, the FieldShell/Textarea,
  // the "Choose file…" button, error display, and the Cancel/Import footer here verbatim.
  // Import button: onClick={() => onImport(yamlContent)} disabled={isLoading || !yamlContent.trim()}.
  return (/* moved upload + paste JSX */ null);
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- docker-compose-import`
Expected: PASS (panel test + existing dialog test still green).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/docker-compose-import-panel.tsx frontend/src/pages/stacks/components/wizard/tests/docker-compose-import-panel.test.tsx frontend/src/pages/stacks/components/shared/docker-compose-import-dialog.tsx
git commit -m "refactor(stacks): extract DockerComposeImportPanel from dialog"
```

---

## Task 8: Wizard shell + list-page integration

**Files:**
- Create: `frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx`
- Modify: `frontend/src/pages/stacks/components/list/index.tsx`
- Test: `frontend/src/pages/stacks/components/wizard/tests/stack-create-wizard.test.tsx`

**Interfaces:**
- Consumes: `WizardChooser`, `BlockComposer`, `TemplatesBrowserPanel`, `DockerComposeImportPanel`, `useTemplateImport`, `useDockerComposeImport`, `templates`, `useNavigate`, `Dialog`.
- Produces: `StackCreateWizard` props `{ open: boolean; onOpenChange: (open: boolean) => void }`. Internal phase state `"chooser" | "composer" | "template" | "compose"` starting at `"chooser"`. Blank → `navigate("/stacks/create")` (no state) + close.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
// frontend/src/pages/stacks/components/wizard/tests/stack-create-wizard.test.tsx
import "@testing-library/jest-dom/vitest";
import { describe, it, expect, vi, beforeAll, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StackCreateWizard } from "../stack-create-wizard";

const navigate = vi.fn();
vi.mock("react-router-dom", () => ({ useNavigate: () => navigate }));
beforeAll(() => { Element.prototype.scrollIntoView = vi.fn(); });
afterEach(() => { cleanup(); navigate.mockReset(); });

describe("StackCreateWizard", () => {
  it("opens on the chooser and advances to the composer", async () => {
    const user = userEvent.setup();
    render(<StackCreateWizard open onOpenChange={vi.fn()} />);
    expect(screen.getByText(/How do you want to start\?/i)).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /Compose blocks/i }));
    expect(screen.getByText(/What's in your stack\?/i)).toBeInTheDocument();
  });

  it("blank slate navigates straight to the empty form", async () => {
    const user = userEvent.setup();
    const onOpenChange = vi.fn();
    render(<StackCreateWizard open onOpenChange={onOpenChange} />);
    await user.click(screen.getByRole("button", { name: /blank slate/i }));
    expect(navigate).toHaveBeenCalledWith("/stacks/create");
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test -- stack-create-wizard`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the shell**

```tsx
// frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { Layers } from "lucide-react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { templates } from "@/pages/stacks/data/templates/registry";
import { useTemplateImport } from "@/pages/stacks/hooks/use-template-import";
import { useDockerComposeImport } from "@/pages/stacks/hooks/use-docker-compose-import";
import { WizardChooser } from "./wizard-chooser";
import { BlockComposer } from "./block-composer";
import { TemplatesBrowserPanel } from "./templates-browser-panel";
import { DockerComposeImportPanel } from "./docker-compose-import-panel";

type Phase = "chooser" | "composer" | "template" | "compose";

interface StackCreateWizardProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function StackCreateWizard({ open, onOpenChange }: StackCreateWizardProps) {
  const navigate = useNavigate();
  const [phase, setPhase] = useState<Phase>("chooser");
  const tpl = useTemplateImport();
  const compose = useDockerComposeImport();

  const close = () => {
    onOpenChange(false);
    setPhase("chooser"); // reset for next open
  };

  // The hooks navigate + close their own dialog state; here we just close the wizard.
  const onUseTemplate = (t: Parameters<typeof tpl.useTemplate>[0]) => { tpl.useTemplate(t); close(); };
  const onImportCompose = async (yaml: string) => { await compose.handleImport(yaml); close(); };

  const wide = phase === "composer" || phase === "template";

  return (
    <Dialog open={open} onOpenChange={(o) => (o ? onOpenChange(true) : close())}>
      <DialogContent className={`block gap-0 overflow-hidden p-0 ${wide ? "sm:max-w-[1000px]" : "sm:max-w-[640px]"}`}>
        <div className="flex items-center gap-3 border-b px-5 py-3.5">
          <span className="flex h-6 w-6 items-center justify-center text-primary"><Layers className="h-5 w-5" /></span>
          <span className="font-mono text-[11px] uppercase tracking-[1.5px] text-muted-foreground">New Stack</span>
        </div>

        <div className="max-h-[80vh] overflow-y-auto">
          {phase === "chooser" && (
            <WizardChooser
              onPickBlocks={() => setPhase("composer")}
              onPickTemplate={() => setPhase("template")}
              onPickCompose={() => setPhase("compose")}
              onPickBlank={() => { navigate("/stacks/create"); close(); }}
            />
          )}
          {phase === "composer" && <BlockComposer onBack={() => setPhase("chooser")} onClose={close} />}
          {phase === "template" && (
            <div className="h-[70vh]"><TemplatesBrowserPanel templates={templates} onUse={onUseTemplate} /></div>
          )}
          {phase === "compose" && (
            <DockerComposeImportPanel
              onImport={onImportCompose}
              isLoading={compose.isLoading}
              error={compose.error}
              onClearError={compose.clearError}
            />
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test -- stack-create-wizard`
Expected: PASS.

- [ ] **Step 5: Wire it into the list page**

In `frontend/src/pages/stacks/components/list/index.tsx`: remove the `DockerComposeImportDropdown` + `DockerComposeImportDialog` + `TemplatesBrowserDialog` usage and the import-hook plumbing currently in the header. Replace the header action with a single "New Stack" button that opens the wizard:

```tsx
// add near other imports
import { StackCreateWizard } from "@/pages/stacks/components/wizard/stack-create-wizard";
// in component state
const [wizardOpen, setWizardOpen] = useState(false);
// in JSX header (replace the New Stack button + Import dropdown)
<Button onClick={() => setWizardOpen(true)}>
  <PlusCircle className="mr-1.5 h-4 w-4" /> New Stack
</Button>
<StackCreateWizard open={wizardOpen} onOpenChange={setWizardOpen} />
```

Remove now-unused imports from `list/index.tsx`: `DockerComposeImportDropdown`, `DockerComposeImportDialog`, `TemplatesBrowserDialog`, `useDockerComposeImport`, `useTemplateImport`, `templates`, `GitBranch`/`Box`/`ChevronDown` if only used by the dropdown. (Leave the empty-state "New Stack" button — point it at `setWizardOpen(true)` too.)

- [ ] **Step 6: Run the list page test + typecheck**

Run: `pnpm --prefix frontend test -- stacks/components/list` then `pnpm --prefix frontend exec tsc -b`
Expected: list tests PASS; tsc reports no errors in `pages/stacks`.

- [ ] **Step 7: Manual smoke (Playwright MCP) against the dev server**

With `pnpm --prefix frontend dev` running, open `http://localhost:5173`, go to Stacks → New Stack. Verify: chooser appears → "Compose blocks" → add Web + Postgres → "Open editor" lands on `/stacks/create` with the two resources prefilled; and Template / Docker compose phases import and land prefilled. (Per repo preference, drive the UI via Playwright MCP.)

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/wizard/stack-create-wizard.tsx frontend/src/pages/stacks/components/wizard/tests/stack-create-wizard.test.tsx frontend/src/pages/stacks/components/list/index.tsx
git commit -m "feat(stacks): wire StackCreateWizard into the stacks list"
```

---

## Task 9: Remove dead code + tests

**Files:**
- Delete: `frontend/src/pages/stacks/components/shared/import-dropdown.tsx` (+ its test under `shared/tests/`)
- Delete: `frontend/src/pages/stacks/components/shared/import-options.tsx` (+ its test)
- Delete: `frontend/src/pages/stacks/components/shared/templates-browser-dialog.tsx` (+ its test — superseded by the panel + its test)
- Delete: `frontend/src/pages/stacks/components/shared/docker-compose-import-dialog.tsx` (+ its test — superseded by the panel + its test)

**Interfaces:**
- Consumes: nothing.
- Produces: nothing. Pure removal; the wizard is the only entry point now.

- [ ] **Step 1: Confirm nothing else imports the doomed files**

Run:
```bash
cd /Users/akshaysasidharan/code/stackdome/frontend
grep -rEn "import-dropdown|import-options|templates-browser-dialog|docker-compose-import-dialog|DockerComposeImportDropdown|buildImportOptions" src --include=*.ts --include=*.tsx | grep -v "/wizard/"
```
Expected: no matches outside `wizard/` (the panels live in `wizard/`, not the deleted files). If any non-test source still imports them, fix that first.

- [ ] **Step 2: Delete the files**

```bash
cd /Users/akshaysasidharan/code/stackdome
git rm frontend/src/pages/stacks/components/shared/import-dropdown.tsx \
       frontend/src/pages/stacks/components/shared/import-options.tsx \
       frontend/src/pages/stacks/components/shared/templates-browser-dialog.tsx \
       frontend/src/pages/stacks/components/shared/docker-compose-import-dialog.tsx
git rm frontend/src/pages/stacks/components/shared/tests/import-dropdown.test.tsx \
       frontend/src/pages/stacks/components/shared/tests/import-options.test.tsx \
       frontend/src/pages/stacks/components/shared/tests/templates-browser-dialog.test.tsx \
       frontend/src/pages/stacks/components/shared/tests/docker-compose-import-dialog.test.tsx
```
(If a listed test file does not exist, drop it from the command — verify with `ls frontend/src/pages/stacks/components/shared/tests/` first.)

- [ ] **Step 3: Full frontend gate**

Run:
```bash
pnpm --prefix frontend test
pnpm --prefix frontend exec tsc -b
pnpm --prefix frontend lint
```
Expected: all green; no references to deleted modules.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore(stacks): remove import dropdown + dialogs superseded by wizard"
```

---

## Self-Review

- **Spec coverage:**
  - "Wizard before the form page" → Tasks 5,8 (chooser + shell), lands in unchanged form via `importedData`. ✓
  - "Fold template + docker compose into the chooser modal" → Tasks 6,7 (extracted panels) + Task 8 (phases). ✓
  - "Build from blocks" → Tasks 2,3,4 (catalog, conversion, composer). ✓
  - "Don't need both form and canvas; canvas separate" → canvas explicitly out of scope; lands in existing form. ✓
  - "Remove dead code + tests" → Task 9. ✓
  - "New branch stack-creation-redesign" → Task 0. ✓
  - GitHub tile disabled "soon" → Task 5. ✓
  - Auto-wiring deferred → noted in Task 4 (panel omitted). ✓
  - v1 catalog Web+DB+Custom, Worker/Cron/Job deferred → Task 2. ✓
- **Placeholder scan:** Tasks 6 & 7 say "move the existing JSX verbatim" rather than reprinting the dialogs' full bodies — this is a deliberate *refactor-move*, not a placeholder; the source is the named file and behavior must be preserved (its existing test stays green as the guard). All net-new code (Tasks 1–5, 8) is shown in full.
- **Type consistency:** `ImportSource.Blocks` (Task 1) used in Task 4 & 8. `BlockPreset`/`BlockId` (Task 2) used in Tasks 3,4. `blockToResources`/`addBlockToStack`/`emptyStack` (Task 3) used in Task 4. `WizardChooser`/`BlockComposer`/`TemplatesBrowserPanel`/`DockerComposeImportPanel` props match their consumers in Task 8. ✓
