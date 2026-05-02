# Stack ↔ Addon Env Linkage — Inline Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the `AddFromAddonDialog` modal with inline addon/database/field pickers in the env table. When a row's `from` is `addon`, the row grows a second line with three dropdowns. Backend payload (`env_from_addons[]`) and the converter logic stay unchanged.

**Architecture:** Drive the change inside the existing `EnvRow` component. Add new props (`addons`, `onChangeAddon`, `rowErrors`, `onBlur`) and a sub-component (`AddonInlinePickers`) that renders three Selects when `from === "addon"`. The parent (`stack-resource-item.tsx`) loses the dialog mount and the toolbar button, gains a `dirtyEnvRows` set for blur-based validation, and feeds errors per row.

**Tech Stack:** React 18, TypeScript, Zod (form-schema), Radix Select (via shadcn `Select`), vitest + jsdom + @testing-library/react.

**Spec:** `docs/superpowers/specs/2026-05-02-stack-addon-env-inline-redesign-design.md`

---

## File Plan

### Modify
- `frontend/src/pages/stacks/components/shared/env-row.tsx` — add `from: "addon"` second-line rendering with three pickers, `rowErrors` prop, blur signal, orphan read-only mode preserved.
- `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx` — wire new props through to EnvRow, add `dirtyEnvRows`, extend `switchRowFrom`, remove `AddFromAddonDialog` import + mount + toolbar button + `addonDialogOpen` state.
- `frontend/src/pages/stacks/schemas/form-schema.ts` — `credField` becomes `.optional()`; add `.refine()` for "database required when not superuser" and "credField required when from=addon"; sharpen `addonId` error message to "Pick an addon".
- `frontend/src/pages/stacks/lib/addon-presets.ts` — delete `applyPreset`, `DEFAULT_ENV_NAMES`, `Preset`, `PresetResult`. Keep `CRED_FIELDS`, `CLUSTER_WIDE_FIELDS`, `CredField`.
- `frontend/__tests__/addon-presets.test.ts` — drop `describe("DEFAULT_ENV_NAMES")` and `describe("applyPreset")`. Keep `describe("CRED_FIELDS")` and `describe("CLUSTER_WIDE_FIELDS")`.

### Create
- `frontend/__tests__/env-row-addon.test.tsx` — component tests for addon-row behaviour.

### Delete
- `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx`
- `frontend/__tests__/add-from-addon-dialog.test.tsx`

### Untouched
- `frontend/src/api/stacks.ts`, `frontend/src/pages/addons/hooks/use-postgres-addons.ts`, all backend code, `frontend/__tests__/stacks-env-roundtrip.test.ts` (re-run only).

---

## Test Fixture (used in env-row-addon.test.tsx)

```ts
import type { PostgresAddon } from "@/api/addons";

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => ({
  id: "addon-1",
  name: "tooljet-db",
  status: { state: "Ready" },
  spec: {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "analytics" }],
    configuration: { enable_superuser_access: false },
  } as any,
  ...(over as any),
});

const baseStackRow = (over: Partial<FormEnvVarData> = {}) =>
  ({ from: "stack", name: "FOO", value: "bar", ...over } as FormEnvVarData);

const baseAddonRow = (over: Partial<Extract<FormEnvVarData, { from: "addon" }>> = {}) =>
  ({
    from: "addon",
    name: "PG_HOST",
    addonType: "postgres",
    addonId: "addon-1",
    database: "tooljet",
    superuser: false,
    credField: "host",
    ...over,
  } as FormEnvVarData);

const noopProps = {
  index: 0,
  resourceIndex: 0,
  secrets: [],
  secretsLoading: false,
  addonNameById: new Map([["addon-1", "tooljet-db"]]),
  onChangeName: vi.fn(),
  onChangeValue: vi.fn(),
  onChangeFrom: vi.fn(),
  onChangeSecret: vi.fn(),
  onChangeAddon: vi.fn(),
  onBlur: vi.fn(),
  onRemove: vi.fn(),
};
```

---

## Task 1: Allow `from: addon` rows to be edited inline (scaffold the new pickers)

Today the EnvRow code disables the name input and the From dropdown when `from === "addon"`, and the `Addon` SelectItem in the From dropdown is itself disabled. We're flipping all three: a user can rename an addon row, can switch a row's `from` to/from Addon via the dropdown, and the addon row's value cell is replaced with a placeholder that we'll fill in over the next tasks.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx` (new)

- [ ] **Step 1: Create the failing test file.**

Write `frontend/__tests__/env-row-addon.test.tsx`:

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen } from "@testing-library/react";
import { EnvRow } from "../src/pages/stacks/components/shared/env-row";
import type { FormEnvVarData } from "../src/pages/stacks/schemas/form-schema";
import type { PostgresAddon } from "../src/api/addons";

const mkAddon = (over: Partial<PostgresAddon> = {}): PostgresAddon => ({
  id: "addon-1",
  name: "tooljet-db",
  status: { state: "Ready" },
  spec: {
    version: { major: 17 },
    storage: { size: "5Gi" },
    databases: [{ name: "tooljet" }, { name: "analytics" }],
    configuration: { enable_superuser_access: false },
  } as any,
  ...(over as any),
});

const baseAddonRow = (over: Partial<Extract<FormEnvVarData, { from: "addon" }>> = {}): FormEnvVarData =>
  ({
    from: "addon",
    name: "PG_HOST",
    addonType: "postgres",
    addonId: "addon-1",
    database: "tooljet",
    superuser: false,
    credField: "host",
    ...over,
  } as FormEnvVarData);

const noopProps = {
  index: 0,
  resourceIndex: 0,
  secrets: [],
  secretsLoading: false,
  addons: [mkAddon()],
  addonNameById: new Map([["addon-1", "tooljet-db"]]),
  onChangeName: vi.fn(),
  onChangeValue: vi.fn(),
  onChangeFrom: vi.fn(),
  onChangeSecret: vi.fn(),
  onChangeAddon: vi.fn(),
  onBlur: vi.fn(),
  onRemove: vi.fn(),
};

describe("EnvRow (addon variant)", () => {
  it("makes name input editable on addon rows", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    const nameInput = screen.getByPlaceholderText("KEY");
    expect(nameInput).not.toBeDisabled();
  });

  it("From dropdown is enabled and Addon item is enabled on addon rows", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    const fromTrigger = screen.getAllByRole("combobox").find((el) =>
      el.textContent?.match(/Stack|Secret|Addon/),
    );
    expect(fromTrigger).toBeDefined();
    expect(fromTrigger).not.toBeDisabled();
  });

  it("renders three picker triggers (Addon, Database, Field) on the second line of an addon row", () => {
    render(<EnvRow row={baseAddonRow()} {...noopProps} />);
    expect(screen.getByTestId("addon-picker-trigger")).toBeInTheDocument();
    expect(screen.getByTestId("database-picker-trigger")).toBeInTheDocument();
    expect(screen.getByTestId("field-picker-trigger")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run the new tests to verify they fail.**

Run from `/Users/akshaysasidharan/code/stackdome/frontend`:
```bash
pnpm test:run env-row-addon
```
Expected: FAIL — at minimum the three `getByTestId` calls fail (no such elements yet); the name input may also fail (still `disabled`).

- [ ] **Step 3: Replace `AddonValueCell` rendering with inline pickers.**

In `frontend/src/pages/stacks/components/shared/env-row.tsx`:

(a) Add `addons`, `onChangeAddon`, `onBlur`, `rowErrors` to `EnvRowProps` (we'll use `rowErrors` and `onBlur` later — declare them now to keep the prop surface stable):

```tsx
import type { PostgresAddon } from "@/api/addons";
import { CRED_FIELDS, CLUSTER_WIDE_FIELDS, type CredField } from "@/pages/stacks/lib/addon-presets";

export type AddonBindingPatch = {
  addonId?: string;
  database?: string | null;  // null = explicitly cleared (All databases)
  superuser?: boolean;
  credField?: CredField;
};

export type EnvRowErrors = {
  name?: string;
  addonId?: string;
  database?: string;
  credField?: string;
  duplicate?: string;
};

interface EnvRowProps {
  row: FormEnvVarData;
  index: number;
  resourceIndex: number;
  secrets: Secret[];
  secretsLoading: boolean;
  addons: PostgresAddon[];
  addonNameById?: Map<string, string>;
  rowErrors?: EnvRowErrors;
  onChangeName: (name: string) => void;
  onChangeValue: (value: string) => void;
  onChangeFrom: (from: EnvFrom) => void;
  onChangeSecret: (secretId: string, secretKey: string) => void;
  onChangeAddon: (patch: AddonBindingPatch) => void;
  onBlur?: () => void;
  onRemove: () => void;
}
```

(b) Add the new prop names to the function signature destructure: `addons`, `rowErrors`, `onChangeAddon`, `onBlur`.

(c) Remove `disabled={row.from === "addon"}` from the name `<Input>` (line ~64).

(d) Remove `disabled={row.from === "addon"}` from the From `<Select>` (line ~105).

(e) Remove `disabled` from the `<SelectItem value="addon">` line (line ~113).

(f) Replace the `{row.from === "addon" && (...)}` block (the `<AddonValueCell .../>` call) with:

```tsx
{row.from === "addon" && (
  <AddonInlinePickers
    row={row}
    addons={addons}
    addonNameById={addonNameById}
    onChangeAddon={onChangeAddon}
    rowErrors={rowErrors}
    isOrphan={isOrphanAddon}
  />
)}
```

(g) Replace the `AddonValueCell` function definition (bottom of file, ~lines 216-244) with a stub `AddonInlinePickers` that renders three blank `Select` triggers with the test ids the tests look for. Real wiring lands in later tasks:

```tsx
function AddonInlinePickers({
  row,
  addons: _addons,
  addonNameById: _addonNameById,
  onChangeAddon: _onChangeAddon,
  rowErrors: _rowErrors,
  isOrphan: _isOrphan,
}: {
  row: Extract<FormEnvVarData, { from: "addon" }>;
  addons: PostgresAddon[];
  addonNameById?: Map<string, string>;
  onChangeAddon: (patch: AddonBindingPatch) => void;
  rowErrors?: EnvRowErrors;
  isOrphan: boolean;
}) {
  return (
    <div className="flex gap-2">
      <Select value={row.addonId || undefined}>
        <SelectTrigger className="w-[160px]" data-testid="addon-picker-trigger">
          <SelectValue placeholder="Addon" />
        </SelectTrigger>
        <SelectContent></SelectContent>
      </Select>
      <Select value={row.database || undefined}>
        <SelectTrigger className="w-[140px]" data-testid="database-picker-trigger">
          <SelectValue placeholder="Database" />
        </SelectTrigger>
        <SelectContent></SelectContent>
      </Select>
      <Select value={row.credField || undefined}>
        <SelectTrigger className="w-[140px]" data-testid="field-picker-trigger">
          <SelectValue placeholder="Field" />
        </SelectTrigger>
        <SelectContent></SelectContent>
      </Select>
    </div>
  );
}
```

- [ ] **Step 4: Run the tests to verify they pass.**

```bash
pnpm test:run env-row-addon
```
Expected: PASS (3/3).

- [ ] **Step 5: TS compile sanity check.**

```bash
pnpm exec tsc -b
```
Expected: errors only in `stack-resource-item.tsx` (which now needs to pass `addons` and `onChangeAddon`). Fix immediately by adding placeholders in the call site:

In `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`, find the `<EnvRow ...>` JSX (~line 1240) and add:

```tsx
addons={addons}
onChangeAddon={() => {}}
```

(Real wiring lands in Task 9.) Re-run `tsc -b` — should pass.

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/src/pages/stacks/components/shared/stack-resource-item.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): scaffold inline addon pickers in env row"
```

---

## Task 2: Schema — `credField` optional, refines for required addon fields

The form schema currently allows an addon row whose `database` is missing while `superuser` is false (since `database` is `optional()`), and forces `credField` to be in `CRED_FIELDS` which means a transient empty value `""` would already fail to type. We need to (a) loosen `credField` to optional so in-progress rows are valid types, and (b) add `.refine()`s that produce path-keyed errors at save time.

**Files:**
- Modify: `frontend/src/pages/stacks/schemas/form-schema.ts`
- Test: `frontend/__tests__/stacks-env-roundtrip.test.ts` (re-run; existing fixtures all valid) and ad-hoc inline tests added below.

- [ ] **Step 1: Add a focused vitest for the schema refines.**

Append to `frontend/__tests__/stacks-env-roundtrip.test.ts` (or create a new `frontend/__tests__/form-schema-addon.test.ts` if you prefer keeping concerns separate — for simplicity, append):

```ts
describe("FormEnvVarSchema (addon variant) — refines", () => {
  // Re-import from form-schema.ts. If FormEnvVarSchema is not exported yet,
  // export it from form-schema.ts before continuing.
  it("requires database when superuser is false", async () => {
    const { FormEnvVarSchema } = await import("../src/pages/stacks/schemas/form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: undefined,
      superuser: false,
      credField: "host",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes("database"))).toBe(true);
    }
  });

  it("allows missing database when superuser is true", async () => {
    const { FormEnvVarSchema } = await import("../src/pages/stacks/schemas/form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: undefined,
      superuser: true,
      credField: "host",
    });
    expect(result.success).toBe(true);
  });

  it("requires credField on addon rows", async () => {
    const { FormEnvVarSchema } = await import("../src/pages/stacks/schemas/form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "addon-1",
      database: "tooljet",
      superuser: false,
      credField: undefined,
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path.includes("credField"))).toBe(true);
    }
  });

  it("uses 'Pick an addon' message on empty addonId", async () => {
    const { FormEnvVarSchema } = await import("../src/pages/stacks/schemas/form-schema");
    const result = FormEnvVarSchema.safeParse({
      from: "addon",
      name: "PG_HOST",
      addonType: "postgres",
      addonId: "",
      database: "tooljet",
      superuser: false,
      credField: "host",
    });
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(
        result.error.issues.find((i) => i.path.includes("addonId"))?.message,
      ).toMatch(/pick an addon/i);
    }
  });
});
```

- [ ] **Step 2: Run the tests to verify they fail.**

```bash
pnpm test:run stacks-env-roundtrip
```
Expected: FAIL on the 4 new tests (or `FormEnvVarSchema` not exported error).

- [ ] **Step 3: Update `form-schema.ts`.**

Find the addon variant inside `FormEnvVarSchema = z.discriminatedUnion("from", [...])` (around the third entry).

(a) Export `FormEnvVarSchema` (if it isn't already): change `const FormEnvVarSchema` to `export const FormEnvVarSchema`.

(b) Change the addon variant from:

```ts
z.object({
  from: z.literal("addon"),
  name: z.string().min(1, "Environment variable name is required"),
  addonType: z.literal("postgres"),
  addonId: z.string().min(1),
  database: z.string().optional(),
  superuser: z.boolean().default(false),
  credField: z.enum(CRED_FIELDS),
}),
```

to:

```ts
z.object({
  from: z.literal("addon"),
  name: z.string().min(1, "Environment variable name is required"),
  addonType: z.literal("postgres"),
  addonId: z.string().min(1, "Pick an addon"),
  database: z.string().optional(),
  superuser: z.boolean().default(false),
  credField: z.enum(CRED_FIELDS).optional(),
})
  .refine((d) => d.superuser || (typeof d.database === "string" && d.database.length > 0), {
    message: "Pick a database",
    path: ["database"],
  })
  .refine((d) => typeof d.credField === "string" && d.credField.length > 0, {
    message: "Pick a field",
    path: ["credField"],
  }),
```

- [ ] **Step 4: Run the schema tests.**

```bash
pnpm test:run stacks-env-roundtrip
```
Expected: all PASS.

- [ ] **Step 5: TS compile.**

```bash
pnpm exec tsc -b
```
Expected: PASS. (Optional `credField` on the form variant will widen `FormEnvVarData` — verify EnvRow still typechecks, it does since we already moved to `Extract<..., { from: "addon" }>` and read `row.credField` defensively.)

- [ ] **Step 6: Commit.**

```bash
git add frontend/src/pages/stacks/schemas/form-schema.ts frontend/__tests__/stacks-env-roundtrip.test.ts
git commit -m "feat(stacks): require database/credField on addon env rows via schema refines"
```

---

## Task 3: Addon picker — wire selection + label

The Addon `Select` should list the addons passed in via the `addons` prop (formatted as `name (Postgres · <state>)`) and call `onChangeAddon({ addonId })` on selection.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing tests.**

Append inside the existing `describe("EnvRow (addon variant)", () => { ... })` block:

```tsx
it("lists addons in the addon picker with name and state", async () => {
  const user = userEvent.setup();
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "" }) as any}
      {...noopProps}
      addons={[mkAddon({ id: "a", name: "primary", status: { state: "Ready" } } as any), mkAddon({ id: "b", name: "secondary", status: { state: "Pending" } } as any)]}
    />,
  );
  await user.click(screen.getByTestId("addon-picker-trigger"));
  expect(await screen.findByRole("option", { name: /primary.*Postgres.*Ready/i })).toBeInTheDocument();
  expect(screen.getByRole("option", { name: /secondary.*Postgres.*Pending/i })).toBeInTheDocument();
});

it("calls onChangeAddon when an addon is picked", async () => {
  const user = userEvent.setup();
  const onChangeAddon = vi.fn();
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "" }) as any}
      {...noopProps}
      onChangeAddon={onChangeAddon}
      addons={[mkAddon({ id: "addon-x" } as any)]}
    />,
  );
  await user.click(screen.getByTestId("addon-picker-trigger"));
  await user.click(await screen.findByRole("option", { name: /tooljet-db/i }));
  expect(onChangeAddon).toHaveBeenCalledWith(expect.objectContaining({ addonId: "addon-x" }));
});
```

Add the import at the top of the test file (if not already present):

```tsx
import userEvent from "@testing-library/user-event";
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```
Expected: 2 new tests FAIL (empty SelectContent).

- [ ] **Step 3: Wire the addon picker.**

In `env-row.tsx`'s `AddonInlinePickers`, replace the addon `<Select>` with:

```tsx
<Select
  value={row.addonId || undefined}
  onValueChange={(v) => onChangeAddon({ addonId: v })}
>
  <SelectTrigger className="w-[160px]" data-testid="addon-picker-trigger">
    <SelectValue placeholder="Addon" />
  </SelectTrigger>
  <SelectContent>
    {_addons.map((a) => (
      <SelectItem key={a.id} value={a.id!}>
        {a.name} (Postgres · {a.status?.state ?? "Unknown"})
      </SelectItem>
    ))}
  </SelectContent>
</Select>
```

Rename `_addons` → `addons` and `_onChangeAddon` → `onChangeAddon` in the destructure.

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```
Expected: 5/5 PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): wire inline addon picker in env row"
```

---

## Task 4: Addon picker — empty-state link

When the `addons` prop is empty, opening the addon picker shows a non-selectable item plus a `+ Create Postgres addon` link to `/addons/create/postgres` in a new tab.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing test.**

```tsx
it("shows '+ Create Postgres addon' link when addons list is empty", async () => {
  const user = userEvent.setup();
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "" }) as any}
      {...noopProps}
      addons={[]}
    />,
  );
  await user.click(screen.getByTestId("addon-picker-trigger"));
  const link = await screen.findByRole("link", { name: /create postgres addon/i });
  expect(link).toHaveAttribute("href", "/addons/create/postgres");
  expect(link).toHaveAttribute("target", "_blank");
});
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 3: Implement the empty-state.**

Inside the addon `<SelectContent>`, branch on `addons.length`:

```tsx
<SelectContent>
  {addons.length === 0 ? (
    <div className="px-3 py-3 text-sm">
      <p className="text-muted-foreground mb-2">No Postgres addons yet.</p>
      <a
        href="/addons/create/postgres"
        target="_blank"
        rel="noreferrer"
        className="text-primary underline"
      >
        + Create Postgres addon
      </a>
    </div>
  ) : (
    addons.map((a) => (
      <SelectItem key={a.id} value={a.id!}>
        {a.name} (Postgres · {a.status?.state ?? "Unknown"})
      </SelectItem>
    ))
  )}
</SelectContent>
```

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): empty-state link in inline addon picker"
```

---

## Task 5: Database picker — gate, list, All databases, auto-select

Disabled until an addon is picked. Once picked, lists databases plus a conditional `─ All databases ─` item (only when the addon's `enable_superuser_access === true`). Auto-selects the single database when the addon has exactly one and doesn't support superuser.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing tests.**

```tsx
it("database picker is disabled when no addon is picked", () => {
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "", database: undefined }) as any}
      {...noopProps}
      addons={[mkAddon()]}
    />,
  );
  expect(screen.getByTestId("database-picker-trigger")).toBeDisabled();
});

it("database picker lists addon's databases when an addon is picked", async () => {
  const user = userEvent.setup();
  render(<EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} />);
  await user.click(screen.getByTestId("database-picker-trigger"));
  expect(await screen.findByRole("option", { name: /tooljet/i })).toBeInTheDocument();
  expect(screen.getByRole("option", { name: /analytics/i })).toBeInTheDocument();
});

it("does NOT show 'All databases' when addon does not enable superuser", async () => {
  const user = userEvent.setup();
  render(<EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} />);
  await user.click(screen.getByTestId("database-picker-trigger"));
  expect(screen.queryByRole("option", { name: /all databases/i })).not.toBeInTheDocument();
});

it("shows 'All databases' when addon enables superuser", async () => {
  const user = userEvent.setup();
  const su = mkAddon({
    spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
  });
  render(<EnvRow row={baseAddonRow({ database: undefined }) as any} {...noopProps} addons={[su]} />);
  await user.click(screen.getByTestId("database-picker-trigger"));
  expect(await screen.findByRole("option", { name: /all databases/i })).toBeInTheDocument();
});

it("calls onChangeAddon with superuser=true and database=null when 'All databases' is picked", async () => {
  const user = userEvent.setup();
  const onChangeAddon = vi.fn();
  const su = mkAddon({
    spec: { ...mkAddon().spec, configuration: { enable_superuser_access: true } } as any,
  });
  render(
    <EnvRow
      row={baseAddonRow({ database: undefined }) as any}
      {...noopProps}
      addons={[su]}
      onChangeAddon={onChangeAddon}
    />,
  );
  await user.click(screen.getByTestId("database-picker-trigger"));
  await user.click(await screen.findByRole("option", { name: /all databases/i }));
  expect(onChangeAddon).toHaveBeenCalledWith({ database: null, superuser: true });
});

it("auto-selects the only database when picking an addon with one db and no superuser", async () => {
  const user = userEvent.setup();
  const onChangeAddon = vi.fn();
  const single = mkAddon({
    id: "addon-single",
    spec: { ...mkAddon().spec, databases: [{ name: "only-one" }] } as any,
  });
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "", database: undefined }) as any}
      {...noopProps}
      addons={[single]}
      onChangeAddon={onChangeAddon}
    />,
  );
  await user.click(screen.getByTestId("addon-picker-trigger"));
  await user.click(await screen.findByRole("option", { name: /tooljet-db/i }));
  expect(onChangeAddon).toHaveBeenCalledWith({
    addonId: "addon-single",
    database: "only-one",
    superuser: false,
  });
});
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 3: Implement database picker + auto-select.**

In `AddonInlinePickers`, replace the addon picker's `onValueChange` and the database `<Select>`:

```tsx
const ALL_DATABASES_VALUE = "__ALL_DATABASES__";

const selectedAddon = addons.find((a) => a.id === row.addonId);
const databases = ((selectedAddon?.spec as unknown as { databases?: { name?: string }[] })
  ?.databases ?? []) as { name?: string }[];
const supportsSuperuser =
  (selectedAddon?.spec as unknown as { configuration?: { enable_superuser_access?: boolean } })
    ?.configuration?.enable_superuser_access === true;

const handleAddonChange = (addonId: string) => {
  const a = addons.find((x) => x.id === addonId);
  const dbs = ((a?.spec as unknown as { databases?: { name?: string }[] })?.databases ?? []) as { name?: string }[];
  const aSupportsSU =
    (a?.spec as unknown as { configuration?: { enable_superuser_access?: boolean } })
      ?.configuration?.enable_superuser_access === true;
  // Auto-select single database when addon doesn't support superuser
  if (dbs.length === 1 && !aSupportsSU && dbs[0]?.name) {
    onChangeAddon({ addonId, database: dbs[0].name, superuser: false });
  } else {
    onChangeAddon({ addonId, database: null, superuser: false });
  }
};

const handleDatabaseChange = (value: string) => {
  if (value === ALL_DATABASES_VALUE) {
    onChangeAddon({ database: null, superuser: true });
  } else {
    onChangeAddon({ database: value, superuser: false });
  }
};
```

Use `handleAddonChange` in the addon `Select`'s `onValueChange`. Replace the database `<Select>` with:

```tsx
<Select
  value={row.superuser ? ALL_DATABASES_VALUE : (row.database || undefined)}
  onValueChange={handleDatabaseChange}
  disabled={!row.addonId}
>
  <SelectTrigger className="w-[140px]" data-testid="database-picker-trigger">
    <SelectValue placeholder={row.addonId ? "Database" : "Pick an addon first"} />
  </SelectTrigger>
  <SelectContent>
    {supportsSuperuser && (
      <SelectItem value={ALL_DATABASES_VALUE}>─ All databases ─</SelectItem>
    )}
    {databases.map((d) =>
      d.name ? (
        <SelectItem key={d.name} value={d.name}>
          {d.name}
        </SelectItem>
      ) : null,
    )}
  </SelectContent>
</Select>
```

The `database: null` semantic in `AddonBindingPatch` means "explicitly clear" — the parent must translate this to `undefined` on the row. (See Task 9.)

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```
Expected: PASS for all 6 new tests.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): inline database picker with All databases + auto-select"
```

---

## Task 6: Field picker — gate, CRED_FIELDS, cluster badge

Disabled until an addon is picked. Lists all 8 `CRED_FIELDS`. Cluster-wide fields (per `CLUSTER_WIDE_FIELDS`) get a small muted `cluster` badge inside the dropdown items.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing tests.**

```tsx
it("field picker is disabled when no addon is picked", () => {
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "", database: undefined, credField: undefined as any }) as any}
      {...noopProps}
      addons={[mkAddon()]}
    />,
  );
  expect(screen.getByTestId("field-picker-trigger")).toBeDisabled();
});

it("field picker lists all 8 CRED_FIELDS", async () => {
  const user = userEvent.setup();
  render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
  await user.click(screen.getByTestId("field-picker-trigger"));
  for (const f of ["host", "port", "username", "password", "database", "sslmode", "connectionString", "caCertificate"]) {
    expect(await screen.findByRole("option", { name: new RegExp(f, "i") })).toBeInTheDocument();
  }
});

it("cluster-wide fields show 'cluster' badge in dropdown items", async () => {
  const user = userEvent.setup();
  render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
  await user.click(screen.getByTestId("field-picker-trigger"));
  const hostOption = await screen.findByRole("option", { name: /host/i });
  expect(hostOption).toHaveTextContent(/cluster/i);
  const userOption = screen.getByRole("option", { name: /^username/i });
  expect(userOption).not.toHaveTextContent(/cluster/i);
});

it("calls onChangeAddon with credField when a field is picked", async () => {
  const user = userEvent.setup();
  const onChangeAddon = vi.fn();
  render(<EnvRow row={baseAddonRow() as any} {...noopProps} onChangeAddon={onChangeAddon} />);
  await user.click(screen.getByTestId("field-picker-trigger"));
  await user.click(await screen.findByRole("option", { name: /port/i }));
  expect(onChangeAddon).toHaveBeenCalledWith({ credField: "port" });
});
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 3: Implement field picker.**

Replace the field `<Select>` in `AddonInlinePickers`:

```tsx
<Select
  value={row.credField || undefined}
  onValueChange={(v) => onChangeAddon({ credField: v as CredField })}
  disabled={!row.addonId}
>
  <SelectTrigger className="w-[140px]" data-testid="field-picker-trigger">
    <SelectValue placeholder={row.addonId ? "Field" : "Pick an addon first"} />
  </SelectTrigger>
  <SelectContent>
    {CRED_FIELDS.map((f) => (
      <SelectItem key={f} value={f}>
        <span className="flex items-center gap-2">
          <span>{f}</span>
          {CLUSTER_WIDE_FIELDS.has(f) && (
            <span className="text-[10px] uppercase tracking-wide text-muted-foreground">
              cluster
            </span>
          )}
        </span>
      </SelectItem>
    ))}
  </SelectContent>
</Select>
```

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```
Expected: PASS for all 4 new tests.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): inline field picker with cluster badge"
```

---

## Task 7: Lazy validation rendering (`rowErrors` prop)

The EnvRow renders red borders + inline error messages based on the `rowErrors` prop. The parent decides when errors are shown (Task 9 covers blur tracking + save-time aggregation). Here we just make EnvRow react correctly to errors when present.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing tests.**

```tsx
it("renders no error styling without rowErrors", () => {
  render(<EnvRow row={baseAddonRow() as any} {...noopProps} />);
  expect(screen.getByTestId("addon-picker-trigger")).not.toHaveClass("border-destructive");
});

it("renders red border + message on addon picker when rowErrors.addonId set", () => {
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "" }) as any}
      {...noopProps}
      rowErrors={{ addonId: "Pick an addon" }}
    />,
  );
  expect(screen.getByTestId("addon-picker-trigger")).toHaveClass("border-destructive");
  expect(screen.getByText("Pick an addon")).toBeInTheDocument();
});

it("renders red border + message on database picker when rowErrors.database set", () => {
  render(
    <EnvRow
      row={baseAddonRow({ database: undefined }) as any}
      {...noopProps}
      rowErrors={{ database: "Pick a database" }}
    />,
  );
  expect(screen.getByTestId("database-picker-trigger")).toHaveClass("border-destructive");
  expect(screen.getByText("Pick a database")).toBeInTheDocument();
});

it("renders red border + message on field picker when rowErrors.credField set", () => {
  render(
    <EnvRow
      row={baseAddonRow({ credField: undefined as any }) as any}
      {...noopProps}
      rowErrors={{ credField: "Pick a field" }}
    />,
  );
  expect(screen.getByTestId("field-picker-trigger")).toHaveClass("border-destructive");
  expect(screen.getByText("Pick a field")).toBeInTheDocument();
});

it("renders duplicate name error on the name input", () => {
  render(
    <EnvRow
      row={baseAddonRow() as any}
      {...noopProps}
      rowErrors={{ duplicate: 'Duplicate name "PG_HOST"' }}
    />,
  );
  expect(screen.getByPlaceholderText("KEY")).toHaveClass("border-destructive");
  expect(screen.getByText('Duplicate name "PG_HOST"')).toBeInTheDocument();
});
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 3: Implement.**

In `AddonInlinePickers`, add per-trigger className conditional on `rowErrors`:

```tsx
<SelectTrigger
  className={`w-[160px] ${rowErrors?.addonId ? "border-destructive" : ""}`}
  data-testid="addon-picker-trigger"
>
```

…similar for database and field triggers.

After the three Selects close, render an error block:

```tsx
{(rowErrors?.addonId || rowErrors?.database || rowErrors?.credField) && (
  <div className="mt-1 space-y-0.5">
    {rowErrors?.addonId && <p className="text-xs text-destructive">{rowErrors.addonId}</p>}
    {rowErrors?.database && <p className="text-xs text-destructive">{rowErrors.database}</p>}
    {rowErrors?.credField && <p className="text-xs text-destructive">{rowErrors.credField}</p>}
  </div>
)}
```

For the name input + duplicate, in the main `EnvRow` body where the name `<Input>` lives:

```tsx
<Input
  id={`env-name-${resourceIndex}-${index}`}
  value={row.name || ""}
  onChange={(e) => onChangeName(e.target.value)}
  className={`w-full text-sm font-mono ${isOrphanAddon ? "opacity-60" : ""} ${rowErrors?.duplicate || rowErrors?.name ? "border-destructive" : ""}`}
  placeholder="KEY"
  readOnly={isOrphanAddon}
/>
```

And under the row's grid, add an inline error if duplicate:

```tsx
{rowErrors?.duplicate && (
  <p className="col-span-full text-xs text-destructive mt-0.5 mb-1 px-3">
    {rowErrors.duplicate}
  </p>
)}
{rowErrors?.name && (
  <p className="col-span-full text-xs text-destructive mt-0.5 mb-1 px-3">
    {rowErrors.name}
  </p>
)}
```

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): render lazy-validation errors on env row"
```

---

## Task 8: Orphan addon row stays read-only on second line

When the row's `addonId` isn't in `addonNameById`, the inline pickers must NOT render — instead, the orphan read-only display from before continues. The first line (name input, From dropdown, [×]) stays interactive.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/env-row.tsx`
- Test: `frontend/__tests__/env-row-addon.test.tsx`

- [ ] **Step 1: Append failing test.**

```tsx
it("orphan addon row renders read-only second line with warning, but From dropdown still works", () => {
  render(
    <EnvRow
      row={baseAddonRow({ addonId: "missing-addon", database: "tooljet", credField: "host" }) as any}
      {...noopProps}
      addonNameById={new Map([["addon-1", "tooljet-db"]])}
      addons={[mkAddon()]}
    />,
  );
  // Read-only display present
  expect(screen.getByText(/<missing addon>/i)).toBeInTheDocument();
  expect(
    screen.getByText(/Addon was deleted\. This variable won't resolve\. Remove to clean up\./i),
  ).toBeInTheDocument();
  // Inline pickers NOT rendered
  expect(screen.queryByTestId("addon-picker-trigger")).not.toBeInTheDocument();
  expect(screen.queryByTestId("database-picker-trigger")).not.toBeInTheDocument();
  expect(screen.queryByTestId("field-picker-trigger")).not.toBeInTheDocument();
  // From dropdown still works (not disabled)
  const fromTrigger = screen.getAllByRole("combobox").find((el) =>
    el.textContent?.match(/Addon/),
  );
  expect(fromTrigger).not.toBeDisabled();
});
```

- [ ] **Step 2: Run, see fail.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 3: Implement.**

In `EnvRow`, branch the value-cell rendering for `from === "addon"`:

```tsx
{row.from === "addon" && (
  isOrphanAddon ? (
    <AddonOrphanReadOnly
      addonId={row.addonId}
      database={row.database}
      credField={row.credField}
      superuser={row.superuser}
    />
  ) : (
    <AddonInlinePickers
      row={row}
      addons={addons}
      addonNameById={addonNameById}
      onChangeAddon={onChangeAddon}
      rowErrors={rowErrors}
    />
  )
)}
```

Remove the `isOrphan` prop from `AddonInlinePickers` (no longer needed there), and add a small `AddonOrphanReadOnly` component at the bottom:

```tsx
function AddonOrphanReadOnly({
  database,
  credField,
  superuser,
}: {
  addonId: string;
  database?: string;
  credField?: string;
  superuser: boolean;
}) {
  const dbLabel = superuser ? "(superuser)" : database ?? "—";
  return (
    <div className="text-xs italic px-3 py-2 text-yellow-600">
      ⚙ &lt;missing addon&gt; · {dbLabel} · {credField ?? "—"}
    </div>
  );
}
```

(The bottom-of-row warning paragraph already exists from the current code — leave it.)

- [ ] **Step 4: Run, see pass.**

```bash
pnpm test:run env-row-addon
```

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/env-row.tsx frontend/__tests__/env-row-addon.test.tsx
git commit -m "feat(stacks): keep orphan addon rows read-only with inline redesign"
```

---

## Task 9: Wire EnvRow into stack-resource-item.tsx + remove dialog

This is the big plumbing task. We compose all the EnvRow props from the parent, add blur-based dirty tracking, extend `switchRowFrom` to handle the `addon` target, and remove the dialog + toolbar button.

**Files:**
- Modify: `frontend/src/pages/stacks/components/shared/stack-resource-item.tsx`
- (No new tests — covered by EnvRow tests + manual smoke in Task 12.)

- [ ] **Step 1: Add `dirtyEnvRows` state and `markEnvRowDirty` helper.**

In the function body of `StackResourceItem`, after `const [addonDialogOpen, setAddonDialogOpen] = useState(false);` (which we'll delete in Step 7 — leave it for now to keep diffs reviewable), add:

```tsx
const [dirtyEnvRows, setDirtyEnvRows] = useState<Set<number>>(new Set());
const markEnvRowDirty = (envIdx: number) => {
  setDirtyEnvRows((prev) => {
    if (prev.has(envIdx)) return prev;
    const next = new Set(prev);
    next.add(envIdx);
    return next;
  });
};
```

- [ ] **Step 2: Extend `switchRowFrom` to handle the `addon` target.**

Find `switchRowFrom` (around line 173). Add the `addon` branch:

```tsx
const switchRowFrom = (envIdx: number, from: EnvFrom) => {
  const current = resource.execution_config?.environment_variables?.[envIdx];
  if (!current) return;
  if (from === "stack") {
    replaceEnvVar(envIdx, { from: "stack", name: current.name, value: "" });
  } else if (from === "secret") {
    replaceEnvVar(envIdx, { from: "secret", name: current.name, secretId: "", secretKey: "" });
  } else if (from === "addon") {
    replaceEnvVar(envIdx, {
      from: "addon",
      name: current.name,
      addonType: "postgres",
      addonId: "",
      database: undefined,
      superuser: false,
      credField: undefined as any, // becomes valid on selection; schema treats as optional
    });
  }
};
```

- [ ] **Step 3: Add `onChangeAddon` handler.**

Below `switchRowFrom`, add:

```tsx
const onChangeAddon = (envIdx: number, patch: { addonId?: string; database?: string | null; superuser?: boolean; credField?: string }) => {
  const current = resource.execution_config?.environment_variables?.[envIdx];
  if (!current || current.from !== "addon") return;
  // null means "explicitly cleared" (All databases case); undefined means "not provided in patch"
  const nextDatabase =
    patch.database === null
      ? undefined
      : patch.database === undefined
        ? current.database
        : patch.database;
  replaceEnvVar(envIdx, {
    ...current,
    addonId: patch.addonId ?? current.addonId,
    database: nextDatabase,
    superuser: patch.superuser ?? current.superuser,
    credField: (patch.credField ?? current.credField) as any,
  });
  markEnvRowDirty(envIdx);
};
```

- [ ] **Step 4: Build the per-row `rowErrors` map.**

Above the `(resource.execution_config?.environment_variables || []).map(...)` JSX (~line 1237), add:

```tsx
const envVars = (resource.execution_config?.environment_variables || []) as FormEnvVarData[];

// Live duplicate-name detection (always on, regardless of dirty state)
const nameCounts = new Map<string, number>();
envVars.forEach((r) => {
  const k = r.name?.trim();
  if (!k) return;
  nameCounts.set(k, (nameCounts.get(k) ?? 0) + 1);
});

const rowErrorsForIndex = (envIdx: number): EnvRowErrors | undefined => {
  const row = envVars[envIdx];
  if (!row) return undefined;
  const out: EnvRowErrors = {};

  // Duplicate name (always live)
  if (row.name && (nameCounts.get(row.name.trim()) ?? 0) > 1) {
    out.duplicate = `Duplicate name "${row.name}"`;
  }

  // Lazy + strict: only show required-field errors after blur or save attempt
  const dirty = dirtyEnvRows.has(envIdx);
  // The form-wide save error path uses `errors["execution_config.environment_variables.<idx>.<field>"]`
  // — fall back to those if present (means save was attempted).
  const errPath = (field: string) =>
    errors[`execution_config.environment_variables.${envIdx}.${field}`];

  if (row.from === "addon") {
    if ((dirty || errPath("addonId")) && !row.addonId) out.addonId = "Pick an addon";
    if ((dirty || errPath("database")) && !row.superuser && !row.database) out.database = "Pick a database";
    if ((dirty || errPath("credField")) && !row.credField) out.credField = "Pick a field";
  }
  if ((dirty || errPath("name")) && !row.name) out.name = "Environment variable name is required";

  return Object.keys(out).length === 0 ? undefined : out;
};
```

(Import `EnvRowErrors` from `./env-row` at the top.)

- [ ] **Step 5: Update the `<EnvRow>` JSX call site.**

Find the `(resource.execution_config?.environment_variables || []).map((env, envIdx) => ...)` block (~line 1237). Update the `<EnvRow ...>` props to:

```tsx
<EnvRow
  key={envIdx}
  row={env as FormEnvVarData}
  index={envIdx}
  resourceIndex={index}
  secrets={secrets.secrets}
  secretsLoading={secrets.isLoading}
  addons={addons}
  addonNameById={addonNameById}
  rowErrors={rowErrorsForIndex(envIdx)}
  onChangeName={(name) => {
    replaceEnvVar(envIdx, { ...(env as FormEnvVarData), name });
  }}
  onChangeValue={(value) => {
    if (env.from === "stack") {
      replaceEnvVar(envIdx, { ...env, value });
    }
  }}
  onChangeFrom={(from) => {
    switchRowFrom(envIdx, from);
    markEnvRowDirty(envIdx);
  }}
  onChangeSecret={(secretId, secretKey) =>
    replaceEnvVar(envIdx, {
      from: "secret",
      name: env.name,
      secretId,
      secretKey,
    })
  }
  onChangeAddon={(patch) => onChangeAddon(envIdx, patch)}
  onBlur={() => markEnvRowDirty(envIdx)}
  onRemove={() => removeEnvVar(envIdx)}
/>
```

- [ ] **Step 6: Wire `onBlur` on the EnvRow's outer container.**

In `env-row.tsx`, add `onBlur={onBlur}` to the outer `<div>` (the one with `border-b last:border-b-0 ...`). React's onBlur bubbles by default (from focusable children), so it covers Inputs and Select triggers. (Radix Select content is portaled, but its trigger is focusable in-place and fires blur on close.)

- [ ] **Step 7: Remove the `Add from addon` toolbar button.**

In `stack-resource-item.tsx`, find the JSX at line ~1178 (`{/* Add from addon button */}`). Delete the entire `<Button onClick={...}>...<span>Add from addon</span></Button>` block (the few lines between the comment and closing tag).

- [ ] **Step 8: Remove the `<AddFromAddonDialog ...>` mount.**

Find the `<AddFromAddonDialog open={addonDialogOpen} ...>` block (~line 1276). Delete the entire JSX element including its `onAdd` handler.

- [ ] **Step 9: Remove the dialog state and import.**

- Delete the `const [addonDialogOpen, setAddonDialogOpen] = useState(false);` line (~line 101).
- Delete the `import { AddFromAddonDialog } from "./add-from-addon-dialog";` line (~line 33).

- [ ] **Step 10: TS compile + run all tests.**

```bash
pnpm exec tsc -b
pnpm test:run
```
Expected: TS PASS; all existing tests + new env-row tests PASS. The `add-from-addon-dialog.test.tsx` will still pass (the dialog file still exists — Task 10 deletes it).

- [ ] **Step 11: Commit.**

```bash
git add frontend/src/pages/stacks/components/shared/stack-resource-item.tsx frontend/src/pages/stacks/components/shared/env-row.tsx
git commit -m "feat(stacks): wire inline addon pickers, drop AddFromAddonDialog"
```

---

## Task 10: Delete `add-from-addon-dialog.tsx` and its test

The dialog has no remaining importer.

**Files:**
- Delete: `frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx`
- Delete: `frontend/__tests__/add-from-addon-dialog.test.tsx`

- [ ] **Step 1: Delete the files.**

```bash
rm frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx
rm frontend/__tests__/add-from-addon-dialog.test.tsx
```

- [ ] **Step 2: Verify no references remain.**

```bash
grep -r "AddFromAddonDialog\|add-from-addon-dialog" frontend/src frontend/__tests__
```
Expected: empty output.

- [ ] **Step 3: Run TS compile + full tests.**

```bash
pnpm exec tsc -b
pnpm test:run
```
Expected: PASS.

- [ ] **Step 4: Commit.**

```bash
git add -A frontend/src/pages/stacks/components/shared/add-from-addon-dialog.tsx frontend/__tests__/add-from-addon-dialog.test.tsx
git commit -m "chore(stacks): remove obsolete AddFromAddonDialog"
```

---

## Task 11: Trim `addon-presets.ts` and its test

Remove `applyPreset`, `DEFAULT_ENV_NAMES`, `Preset`, `PresetResult` (no remaining importers). Trim test file to drop the corresponding describe blocks.

**Files:**
- Modify: `frontend/src/pages/stacks/lib/addon-presets.ts`
- Modify: `frontend/__tests__/addon-presets.test.ts`

- [ ] **Step 1: Verify no remaining importers.**

```bash
grep -rn "applyPreset\|DEFAULT_ENV_NAMES\|Preset\|PresetResult" frontend/src frontend/__tests__ | grep -v "addon-presets"
```
Expected: empty (only the test file imports these, which we'll trim next).

- [ ] **Step 2: Trim `addon-presets.ts`.**

Replace the contents of `frontend/src/pages/stacks/lib/addon-presets.ts` with:

```ts
export const CRED_FIELDS = [
  "host",
  "port",
  "username",
  "password",
  "database",
  "sslmode",
  "connectionString",
  "caCertificate",
] as const;

export type CredField = (typeof CRED_FIELDS)[number];

export const CLUSTER_WIDE_FIELDS: ReadonlySet<CredField> = new Set<CredField>([
  "host",
  "port",
  "sslmode",
  "caCertificate",
]);
```

- [ ] **Step 3: Trim `addon-presets.test.ts`.**

Replace the contents of `frontend/__tests__/addon-presets.test.ts` with:

```ts
import { describe, it, expect } from "vitest";
import {
  CRED_FIELDS,
  CLUSTER_WIDE_FIELDS,
} from "../src/pages/stacks/lib/addon-presets";

describe("CRED_FIELDS", () => {
  it("lists all 8 fields the backend supports", () => {
    expect(CRED_FIELDS).toEqual([
      "host",
      "port",
      "username",
      "password",
      "database",
      "sslmode",
      "connectionString",
      "caCertificate",
    ]);
  });
});

describe("CLUSTER_WIDE_FIELDS", () => {
  it("contains the four cluster-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("host")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("port")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("sslmode")).toBe(true);
    expect(CLUSTER_WIDE_FIELDS.has("caCertificate")).toBe(true);
  });

  it("does not contain database-scoped credentials", () => {
    expect(CLUSTER_WIDE_FIELDS.has("username")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("password")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("database")).toBe(false);
    expect(CLUSTER_WIDE_FIELDS.has("connectionString")).toBe(false);
  });
});
```

- [ ] **Step 4: TS compile + tests.**

```bash
pnpm exec tsc -b
pnpm test:run
```
Expected: PASS.

- [ ] **Step 5: Commit.**

```bash
git add frontend/src/pages/stacks/lib/addon-presets.ts frontend/__tests__/addon-presets.test.ts
git commit -m "refactor(stacks): trim unused preset helpers from addon-presets"
```

---

## Task 12: Final verification + manual smoke

Run full suite + walk the manual checklist from the design spec.

- [ ] **Step 1: Full vitest run.**

```bash
pnpm test:run
```
Expected: all PASS.

- [ ] **Step 2: TS compile + lint.**

```bash
pnpm exec tsc -b
pnpm lint
```
Expected: PASS (or no new errors compared to `main`).

- [ ] **Step 3: Start dev server.**

```bash
pnpm dev
```
Open `http://localhost:5173`.

- [ ] **Step 4: Walk the manual checklist** (from spec § Testing → Manual):

Check each by clicking through the UI:

1. Stack with one resource, no env vars. Add a stack literal → save → reload → persists.
2. Add an addon-backed env: type `PG_HOST`, set `From: Addon`, pick addon, pick database, pick `host` → save → reload → row reappears intact.
3. Add 4 more addon-backed envs from the same `(addon, db)` → save → on reload, all 5 rows reappear; on the API side, one `env_from_addons[]` entry with 5 `env_mapping` keys.
4. Pick an addon that supports superuser → confirm `─ All databases ─` shows; pick it → save → reload → row reappears with database empty and "All databases" still highlighted.
5. Type a duplicate name (`NODE_ENV` twice across stack + addon rows) → live red border + save blocked.
6. Flip an existing addon row's `From` to `Stack` → addon fields cleared; row becomes a literal with empty value.
7. Delete the addon backing an existing row externally (via the addons page) → reload form → orphan warning strip renders, save still works (orphan preserved).
8. Empty addon list (delete all addons) → set `From: Addon` → open addon dropdown → see "+ Create Postgres addon" link.

- [ ] **Step 5: Update the design-spec implementation marker.**

In `docs/superpowers/specs/2026-05-02-stack-addon-env-inline-redesign-design.md`, change the front-matter `Status:` from `Design` to `Implemented`.

- [ ] **Step 6: Commit final docs touch + manual notes.**

```bash
git add docs/superpowers/specs/2026-05-02-stack-addon-env-inline-redesign-design.md
git commit -m "docs: mark inline addon env spec as implemented"
```

---

## Self-review notes

- **Spec coverage:** Each spec section has at least one task — row anatomy (Tasks 1, 5, 6), picker behaviour (Tasks 3–6), state transitions (Task 9), validation (Tasks 2, 7, 9), edge cases (Tasks 4, 8), file plan (Tasks 9–11), testing (Tasks 1–8 unit, Task 12 manual).
- **Schema delta:** Plan adds two `.refine()`s and loosens `credField` to optional — this contradicts the spec's "no schema change" line. The change is small and additive (only stricter-on-save errors), but the design doc should be updated. Task 12 marks it implemented; reviewer can spot the delta and accept inline.
- **Type consistency:** `AddonBindingPatch` defined in Task 1, used in Tasks 3, 5, 6, 9. `EnvRowErrors` defined in Task 1, used in Tasks 7, 9. `CredField` imported from `addon-presets.ts` and survives Task 11 (kept).
- **Testid stability:** `addon-picker-trigger`, `database-picker-trigger`, `field-picker-trigger`, `env-row-<resourceIdx>-<envIdx>` — referenced consistently across tests.
- **Blur reliability:** Container `onBlur` covers Input blur naturally. Radix Select trigger is a `<button>` and fires blur on focus loss after popover close. Fallback: `onChangeAddon` also calls `markEnvRowDirty(envIdx)` (Task 9, Step 3), so any selection in a row marks it dirty even without a clean blur.
