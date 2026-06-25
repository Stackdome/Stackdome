# Per-release Config Diff Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Wire the now-available release `snapshot` into the existing diff UI and extend it to cover volumes and connections, so clicking a release shows what changed vs its predecessor.

**Architecture:** `GET /releases/{id}` returns `StackReleaseDetail.snapshot`. `useReleaseDetail` caches release details; `diffSnapshots(prev, cur)` produces a structured `SnapshotDiff`; `ConfigDiff` renders it. The flow already exists for history rows; this types it, broadens diff coverage, fixes empty-state copy, and adds a collapsed diff to the active release card.

**Tech Stack:** React 19, TypeScript, Tailwind v4, Vitest + Testing Library. Types generated from `config/openapi/stackdome_api.yaml` into `src/api/types/openapi.d.ts`.

## Global Constraints

- Run tests scoped + serial only: `pnpm exec vitest run <file> --maxWorkers=1 --no-file-parallelism` (the full suite crashes the dev laptop).
- Brand design tokens only (`text-success`/`text-danger`/`text-warn`/`text-fg-muted`/`bg-muted`/`border-border`). No raw hex.
- Import depth: files in `timeline/` import `deployments/` siblings with a single `../` (e.g. `../release-snapshot-diff`); test files in `timeline/tests/` use `../../`; `release-snapshot-diff.ts` lives in `deployments/`, its test in `deployments/tests/` uses `../`.
- Secret/output-backed values are references in the snapshot (`ValueRef.output`, env `self_output`); render them as labels, never as raw values.
- Commit after each task. Branch: `feat/stateful-deployments`.
- Dual-write any doc updates to repo `docs/superpowers/` and `~/Documents/Vault/Stackdome/superpowers/`.

## File Structure

- `src/api/releases.ts` — add `StackReleaseDetail`/`StackReleaseSnapshot` exports; `getRelease` returns detail.
- `src/.../deployments/use-release-detail.ts` — `DetailState.data` typed as detail.
- `src/.../deployments/release-snapshot-diff.ts` — `SnapshotDiff` return shape; volume + connection diffs.
- `src/.../deployments/timeline/config-diff.tsx` — render resources + volumes + connections; empty-state copy.
- `src/.../deployments/timeline/release-post-mortem.tsx` — drop casts; pass `SnapshotDiff` + `hasPrev`.
- `src/.../deployments/timeline/current-release-node.tsx` — collapsed "View config changes" diff for the active release.

Relative-path note: in code blocks below, `D/` = `src/pages/stacks/components/detail/deployments/`.

---

### Task 1: Type the release snapshot end-to-end

**Files:**
- Modify: `frontend/src/api/releases.ts`
- Modify: `frontend/D/use-release-detail.ts`
- Modify: `frontend/D/timeline/release-post-mortem.tsx`

**Interfaces:**
- Produces: `StackReleaseDetail` (= `StackRelease` + `snapshot?`), `StackReleaseSnapshot` exported from `@/api/releases`; `getRelease(...)` returns `Promise<StackReleaseDetail>`; `DetailState.data?: StackReleaseDetail`.

- [ ] **Step 1: Add detail type aliases + change `getRelease` return type**

In `frontend/src/api/releases.ts`, add after the existing `StackRelease` alias:

```ts
export type StackReleaseDetail = components["schemas"]["StackReleaseDetail"];
export type StackReleaseSnapshot = components["schemas"]["StackReleaseSnapshot"];
```

Change `getRelease`'s return type:

```ts
export async function getRelease(orgId: string, teamName: string, stackId: string, releaseId: string): Promise<StackReleaseDetail> {
  const response = await api.get<StackReleaseDetail>(`${releasesPath(orgId, teamName, stackId)}/${releaseId}`);
  return response.data;
}
```

- [ ] **Step 2: Type the detail cache**

In `frontend/D/use-release-detail.ts`, change the import and `DetailState`:

```ts
import { getRelease, type StackReleaseDetail } from "@/api/releases";

export interface DetailState { data?: StackReleaseDetail; loading: boolean; error?: string; }
```

- [ ] **Step 3: Drop the `as unknown` casts in the post-mortem**

In `frontend/D/timeline/release-post-mortem.tsx`, replace the two cast lines:

```ts
const data = cur.data;
const outcomes = data?.outcome?.resources ?? {};
const snap = data?.snapshot;
const prevSnap = prev.data?.snapshot;
const diffs = diffSnapshots(prevSnap, snap);
```

(`outcome` and `snapshot` are now typed on `StackReleaseDetail`. If `outcome` is not present on the type, keep a single narrow cast `(data as { outcome?: {...} })` — verify against the generated type before deciding.)

- [ ] **Step 4: Verify types + existing tests still pass**

Run: `cd frontend && pnpm exec tsc -b 2>&1 | grep -E "releases|release-detail|post-mortem|error TS" | grep -v postgres-backups`
Expected: no output (no new errors).

Run: `pnpm exec vitest run src/pages/stacks/components/detail/deployments --maxWorkers=1 --no-file-parallelism`
Expected: all files pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/releases.ts frontend/src/pages/stacks/components/detail/deployments/use-release-detail.ts frontend/src/pages/stacks/components/detail/deployments/timeline/release-post-mortem.tsx
git commit -m "refactor(frontend): type release snapshot end-to-end"
```

---

### Task 2: SnapshotDiff shape + volume diff

**Files:**
- Modify: `frontend/D/release-snapshot-diff.ts`
- Modify: `frontend/D/timeline/config-diff.tsx`
- Modify: `frontend/D/timeline/release-post-mortem.tsx`
- Test: `frontend/D/tests/release-snapshot-diff.test.ts`

**Interfaces:**
- Produces: `interface ItemDiff { name: string; change: "added"|"removed"|"modified"; rows: DiffRow[]; note?: string }`; `interface SnapshotDiff { resources: ResourceDiff[]; volumes: ItemDiff[]; connections: ItemDiff[] }`; `diffSnapshots(prev?, cur?): SnapshotDiff`.
- Consumes: `StackReleaseSnapshot` (has `volumes?: Volume[]`, each `Volume.spec: VolumeSpec { size, storage_class?, access_mode }`).

- [ ] **Step 1: Write the failing volume-diff test**

In `frontend/D/tests/release-snapshot-diff.test.ts`, add:

```ts
function snap(over: Record<string, unknown>) { return over as unknown as import("../release-snapshot-diff").Snap; }

describe("diffSnapshots volumes", () => {
  const vol = (name: string, size: string) => ({ name, spec: { size, access_mode: "ReadWriteOnce" } });

  it("flags an added, a removed, and a resized volume", () => {
    const prev = { resources: [], volumes: [vol("data", "1Gi"), vol("cache", "500Mi")] };
    const cur = { resources: [], volumes: [vol("data", "2Gi"), vol("logs", "1Gi")] };
    const out = diffSnapshots(prev, cur);
    expect(out.volumes).toEqual([
      { name: "data", change: "modified", rows: [{ key: "size", from: "1Gi", to: "2Gi", kind: "changed" }] },
      { name: "cache", change: "removed", rows: [], note: "Volume removed from this release." },
      { name: "logs", change: "added", rows: [{ key: "size", to: "1Gi", kind: "added" }, { key: "access_mode", to: "ReadWriteOnce", kind: "added" }] },
    ]);
  });

  it("returns no volume diff when volumes are unchanged", () => {
    const v = [vol("data", "1Gi")];
    expect(diffSnapshots({ resources: [], volumes: v }, { resources: [], volumes: v }).volumes).toEqual([]);
  });
});
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts --maxWorkers=1 --no-file-parallelism`
Expected: FAIL — `out.volumes` undefined / `diffSnapshots` returns an array, not `{ volumes }`.

- [ ] **Step 3: Change the return shape and add the volume diff**

In `frontend/D/release-snapshot-diff.ts`, add types and a `Snap` alias near the top:

```ts
import type { StackReleaseSnapshot } from "@/api/releases";
export type Snap = StackReleaseSnapshot;
type SnapVolume = NonNullable<StackReleaseSnapshot["volumes"]>[number];

export interface ItemDiff { name: string; change: "added" | "removed" | "modified"; rows: DiffRow[]; note?: string }
export interface SnapshotDiff { resources: ResourceDiff[]; volumes: ItemDiff[]; connections: ItemDiff[] }
```

Add scalar extraction + a generic name-keyed differ:

```ts
function volumeScalars(v: SnapVolume): Record<string, string | undefined> {
  return { size: v.spec?.size, storage_class: v.spec?.storage_class, access_mode: v.spec?.access_mode };
}

function scalarRows(p: Record<string, string | undefined>, c: Record<string, string | undefined>): DiffRow[] {
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key]; const to = c[key];
    if (from === to) continue;
    if (from == null) rows.push({ key, to, kind: "added" });
    else if (to == null) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

function diffNamed<T>(prev: T[], cur: T[], nameOf: (t: T) => string, scalars: (t: T) => Record<string, string | undefined>, removedNote: string): ItemDiff[] {
  const p = new Map(prev.map((t) => [nameOf(t), t]));
  const c = new Map(cur.map((t) => [nameOf(t), t]));
  const out: ItemDiff[] = [];
  for (const name of new Set([...p.keys(), ...c.keys()])) {
    const a = p.get(name); const b = c.get(name);
    if (a && !b) { out.push({ name, change: "removed", rows: [], note: removedNote }); continue; }
    if (!a && b) { out.push({ name, change: "added", rows: scalarRows({}, scalars(b)) }); continue; }
    const rows = scalarRows(scalars(a as T), scalars(b as T));
    if (rows.length) out.push({ name, change: "modified", rows });
  }
  return out;
}
```

Replace the `diffSnapshots` signature/body to return `SnapshotDiff`:

```ts
export function diffSnapshots(prev?: Snap, cur?: Snap): SnapshotDiff {
  const empty: SnapshotDiff = { resources: [], volumes: [], connections: [] };
  if (prev == null) return empty; // no predecessor — caller distinguishes "initial"
  const resources = diffResources(prev, cur);
  const volumes = diffNamed(prev.volumes ?? [], cur?.volumes ?? [], (v) => v.name ?? "", volumeScalars, "Volume removed from this release.");
  return { resources, volumes, connections: [] };
}
```

Rename the existing resource-diff body into `function diffResources(prev: Snap, cur?: Snap): ResourceDiff[]` (the current `diffSnapshots` body, reading `resourcesOf(prev)`/`resourcesOf(cur)`).

- [ ] **Step 4: Run the volume test — verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts --maxWorkers=1 --no-file-parallelism`
Expected: PASS. Existing resource tests still pass (they read the array — update them, see Step 5).

- [ ] **Step 5: Update existing consumers to the new shape**

Existing resource tests assert `diffSnapshots(...)` is an array. Change them to read `.resources` (e.g. `expect(diffSnapshots(prev, cur).resources).toEqual([...])`).

In `frontend/D/timeline/config-diff.tsx`, change the prop to the structured diff and render volumes:

```tsx
import type { SnapshotDiff, ItemDiff } from "../release-snapshot-diff";

export interface ConfigDiffProps { diff: SnapshotDiff; hasPrev: boolean; prevSeq?: number }

function ItemCard({ item }: { item: ItemDiff }) {
  const dot = item.change === "added" ? "bg-success" : item.change === "removed" ? "bg-fg-muted" : "bg-warn";
  return (
    <div className={`overflow-hidden rounded-md border border-border ${item.change === "removed" ? "opacity-80" : ""}`}>
      <div className="flex items-center gap-2.5 bg-muted px-3 py-2.5">
        <span className={`h-[7px] w-[7px] flex-none rounded-full ${dot}`} />
        <span className={`font-mono text-[12.5px] font-semibold text-foreground ${item.change === "removed" ? "line-through" : ""}`}>{item.name}</span>
      </div>
      {item.note ? (
        <div className="flex items-start gap-2.5 border-t border-border px-3 py-2.5 text-[12px] text-fg-muted"><span className="flex-none">−</span><span>{item.note}</span></div>
      ) : (
        <div className="border-t border-border py-1">{item.rows.map((row, ri) => <Row key={ri} row={row} />)}</div>
      )}
    </div>
  );
}

function Group({ label, items }: { label: string; items: ItemDiff[] }) {
  if (!items.length) return null;
  return (
    <div className="space-y-2.5">
      <div className="font-mono text-[9px] uppercase tracking-wide text-fg-muted">{label}</div>
      {items.map((it) => <ItemCard key={it.name} item={it} />)}
    </div>
  );
}
```

Change `ConfigDiff` to consume `diff` and render resources (existing card markup, mapping over `diff.resources`) plus `<Group label="Volumes" items={diff.volumes} />` and (Task 3) connections. Empty-state copy handled in Task 4 — for now: `if (!diff.resources.length && !diff.volumes.length && !diff.connections.length) return <div className="text-[12.5px] text-fg-muted">{hasPrev ? \`No configuration changes since #${prevSeq ?? "previous"}.\` : "Initial release — nothing to compare."}</div>;`

In `frontend/D/timeline/release-post-mortem.tsx`, pass the structured diff:

```tsx
const diff = diffSnapshots(prevSnap, snap);
// ...
<ConfigDiff diff={diff} hasPrev={!!prevReleaseId} prevSeq={prevSeq} />
```

- [ ] **Step 6: Run deployments tests — verify all pass**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments --maxWorkers=1 --no-file-parallelism`
Expected: all pass (config-diff + release-snapshot-diff + post-mortem updated).

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-snapshot-diff.ts frontend/src/pages/stacks/components/detail/deployments/timeline/config-diff.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/release-post-mortem.tsx frontend/src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts
git commit -m "feat(frontend): SnapshotDiff shape + volume diff"
```

---

### Task 3: Connection diff

**Files:**
- Modify: `frontend/D/release-snapshot-diff.ts`
- Test: `frontend/D/tests/release-snapshot-diff.test.ts`

**Interfaces:**
- Consumes: `StackReleaseSnapshot.connections?: StackConnection[]` where `StackConnection { kind, from: {type,id?,name?}, to: {type,id?,name?}, mappings?: [{ target: { type, name?, path? }, value: { output?, template? } }] }`.
- Produces: fills `SnapshotDiff.connections: ItemDiff[]`.

- [ ] **Step 1: Write the failing connection-diff test**

```ts
describe("diffSnapshots connections", () => {
  const conn = (env: string, output: string) => ({
    kind: "env",
    from: { type: "addon/postgres", name: "db" },
    to: { type: "stack_resource", name: "api" },
    mappings: [{ target: { type: "env", name: env }, value: { output } }],
  });

  it("flags a changed mapping value on an existing connection", () => {
    const prev = { resources: [], connections: [conn("DATABASE_URL", "url")] };
    const cur = { resources: [], connections: [conn("DATABASE_URL", "public.url")] };
    const out = diffSnapshots(prev, cur);
    expect(out.connections).toEqual([
      { name: "env · db → api", change: "modified", rows: [{ key: "DATABASE_URL", from: "url", to: "public.url", kind: "changed" }] },
    ]);
  });

  it("flags an added connection", () => {
    const out = diffSnapshots({ resources: [], connections: [] }, { resources: [], connections: [conn("DATABASE_URL", "url")] });
    expect(out.connections).toEqual([
      { name: "env · db → api", change: "added", rows: [{ key: "DATABASE_URL", to: "url", kind: "added" }] },
    ]);
  });
});
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts --maxWorkers=1 --no-file-parallelism`
Expected: FAIL — `out.connections` is `[]`.

- [ ] **Step 3: Implement the connection diff**

In `frontend/D/release-snapshot-diff.ts` add:

```ts
type SnapConn = NonNullable<StackReleaseSnapshot["connections"]>[number];

function nodeLabel(n?: { type?: string; name?: string; id?: string }): string {
  return n?.name ?? n?.id ?? n?.type ?? "?";
}
function connName(c: SnapConn): string {
  return `${c.kind ?? "?"} · ${nodeLabel(c.from)} → ${nodeLabel(c.to)}`;
}
function valueLabel(v?: { output?: string; template?: string }): string {
  return v?.output ?? v?.template ?? "(reference)";
}
function connScalars(c: SnapConn): Record<string, string | undefined> {
  const out: Record<string, string | undefined> = {};
  for (const m of c.mappings ?? []) {
    const key = m.target?.name ?? m.target?.path ?? "(target)";
    out[key] = valueLabel(m.value);
  }
  return out;
}
```

In `diffSnapshots`, compute connections and include them:

```ts
const connections = diffNamed(prev.connections ?? [], cur?.connections ?? [], connName, connScalars, "Connection removed from this release.");
return { resources, volumes, connections };
```

- [ ] **Step 4: Run the connection test — verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts --maxWorkers=1 --no-file-parallelism`
Expected: PASS.

- [ ] **Step 5: Render the Connections group**

In `frontend/D/timeline/config-diff.tsx`, after the Volumes group add: `<Group label="Connections" items={diff.connections} />`.

- [ ] **Step 6: Run deployments tests**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments --maxWorkers=1 --no-file-parallelism`
Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-snapshot-diff.ts frontend/src/pages/stacks/components/detail/deployments/timeline/config-diff.tsx frontend/src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts
git commit -m "feat(frontend): connection diff between release snapshots"
```

---

### Task 4: Empty-state copy (initial vs no-changes)

**Files:**
- Modify: `frontend/D/timeline/config-diff.tsx`
- Test: `frontend/D/timeline/tests/config-diff.test.tsx`

**Interfaces:**
- Consumes: `ConfigDiff({ diff, hasPrev, prevSeq })` from Task 2.

- [ ] **Step 1: Write the failing copy test**

In `frontend/D/timeline/tests/config-diff.test.tsx`:

```tsx
it("shows the initial-release copy when there is no predecessor", () => {
  render(<ConfigDiff diff={{ resources: [], volumes: [], connections: [] }} hasPrev={false} />);
  expect(screen.getByText(/initial release/i)).toBeInTheDocument();
});

it("shows the no-changes copy when a predecessor exists but nothing changed", () => {
  render(<ConfigDiff diff={{ resources: [], volumes: [], connections: [] }} hasPrev prevSeq={12} />);
  expect(screen.getByText(/no configuration changes since #12/i)).toBeInTheDocument();
});
```

(Update the existing config-diff test render calls to the new `diff={...}` prop shape.)

- [ ] **Step 2: Run it to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/timeline/tests/config-diff.test.tsx --maxWorkers=1 --no-file-parallelism`
Expected: FAIL on the no-changes assertion (copy currently always "Initial release").

- [ ] **Step 3: Implement the conditional copy**

Confirm the empty branch in `ConfigDiff` reads:

```tsx
if (!diff.resources.length && !diff.volumes.length && !diff.connections.length)
  return <div className="text-[12.5px] text-fg-muted">{hasPrev ? `No configuration changes since #${prevSeq ?? "previous"}.` : "Initial release — nothing to compare."}</div>;
```

- [ ] **Step 4: Run the test — verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/timeline/tests/config-diff.test.tsx --maxWorkers=1 --no-file-parallelism`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/config-diff.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/config-diff.test.tsx
git commit -m "fix(frontend): distinguish initial release from no-config-changes"
```

---

### Task 5: Active-release "View config changes" diff

**Files:**
- Modify: `frontend/D/timeline/current-release-node.tsx`
- Modify: `frontend/D/timeline/timeline-rail.tsx`
- Test: `frontend/D/timeline/tests/current-release-node.test.tsx`

**Interfaces:**
- Consumes: `useReleaseDetail` (`ReleaseDetail`), `diffSnapshots`, `ConfigDiff`.
- Produces: `CurrentReleaseNode` accepts `detail?: ReleaseDetail`, `prevReleaseId?: string`, `prevSeq?: number` and renders a collapsed "View config changes" toggle when `prevReleaseId` is set.

- [ ] **Step 1: Write the failing toggle test**

In `frontend/D/timeline/tests/current-release-node.test.tsx`, add (Wrap passes a real `useReleaseDetail`):

```tsx
import { useReleaseDetail } from "../../use-release-detail";
vi.mock("@/api/releases", async (orig) => ({ ...(await orig()), getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, snapshot: { resources: [], volumes: [], connections: [] } }) }));

function WrapDiff() {
  const detail = useReleaseDetail("o", "t", "s");
  return <CurrentReleaseNode release={{ id: "r1", sequence: 14, state: "Released" } as StackRelease} stack={stack()} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} detail={detail} prevReleaseId="r0" prevSeq={13} />;
}

it("toggles a config-changes diff for the active release", async () => {
  render(<WrapDiff />);
  await userEvent.click(screen.getByRole("button", { name: /view config changes/i }));
  expect(await screen.findByText(/no configuration changes since #13/i)).toBeInTheDocument();
});

it("hides the diff toggle when there is no predecessor", () => {
  render(<CurrentReleaseNode release={{ id: "r1", sequence: 1, state: "Released" } as StackRelease} stack={stack()} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} />);
  expect(screen.queryByRole("button", { name: /view config changes/i })).not.toBeInTheDocument();
});
```

- [ ] **Step 2: Run it to verify it fails**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/timeline/tests/current-release-node.test.tsx --maxWorkers=1 --no-file-parallelism`
Expected: FAIL — no "View config changes" button.

- [ ] **Step 3: Add the collapsed diff to the current card**

In `frontend/D/timeline/current-release-node.tsx`:

```tsx
import { useState } from "react";
import { diffSnapshots } from "../release-snapshot-diff";
import type { ReleaseDetail } from "../use-release-detail";
import { ConfigDiff } from "./config-diff";
```

Extend props: `detail?: ReleaseDetail; prevReleaseId?: string; prevSeq?: number;`. In the body:

```tsx
const [showDiff, setShowDiff] = useState(false);
const canDiff = !!prevReleaseId && !!detail;
// when opening, ensure both detail fetches
const onToggleDiff = () => {
  if (!detail || !prevReleaseId) return;
  if (release.id) detail.ensure(release.id);
  detail.ensure(prevReleaseId);
  setShowDiff((v) => !v);
};
const curSnap = detail?.peek(release.id).data?.snapshot;
const prevSnap = detail?.peek(prevReleaseId).data?.snapshot;
const diff = diffSnapshots(prevSnap, curSnap);
```

Render after the resources/recovered blocks, before the card closes:

```tsx
{canDiff && (
  <div className="mt-4">
    <button onClick={onToggleDiff} className="font-sans text-[12.5px] font-medium text-primary">
      {showDiff ? "Hide config changes" : `View config changes · vs #${prevSeq ?? "previous"}`}
    </button>
    {showDiff && <div className="mt-3"><ConfigDiff diff={diff} hasPrev prevSeq={prevSeq} /></div>}
  </div>
)}
```

- [ ] **Step 4: Thread detail + predecessor from the rail**

In `frontend/D/timeline/timeline-rail.tsx`, pass to `CurrentReleaseNode`:

```tsx
<CurrentReleaseNode release={activeRelease} stack={stack} logContext={logContext} onOpenLogs={onOpenLogs} onCancel={onCancel}
  detail={detail} prevReleaseId={prevIdFor(0)} prevSeq={prevSeqFor(0)} />
```

(`detail`, `prevIdFor`, `prevSeqFor` already exist in `timeline-rail.tsx`.)

- [ ] **Step 5: Run the test — verify it passes**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments/timeline/tests/current-release-node.test.tsx --maxWorkers=1 --no-file-parallelism`
Expected: PASS.

- [ ] **Step 6: Run the whole deployments dir + typecheck**

Run: `cd frontend && pnpm exec vitest run src/pages/stacks/components/detail/deployments --maxWorkers=1 --no-file-parallelism`
Run: `pnpm exec tsc -b 2>&1 | grep -E "error TS" | grep -v postgres-backups`
Expected: all tests pass; no new type errors.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/current-release-node.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/timeline-rail.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/current-release-node.test.tsx
git commit -m "feat(frontend): config-changes diff on the active release card"
```

---

### Task 6: Live verification against a seeded stack

**Files:** none (verification only).

- [ ] **Step 1: Seed a multi-component stack**

Run: `hack/run_local.sh samples/tooljet_addon_postgres.json` (resources + a postgres connection + a volume). If it expects an already-running env, ensure `docker ps` shows `psql-stackdome-dev` and the API server on :8000 first.

- [ ] **Step 2: Deploy release #1, then change config and deploy #2**

Via the authed UI (or API): deploy once; then edit the stack — bump the app image tag, add/modify one env var, and tweak a connection mapping — and deploy again.

- [ ] **Step 3: Open the active card diff and screenshot**

In the Deployments tab, click "View config changes · vs #1" on the active card. Confirm the rendered diff shows the image bump, the env change, and the connection/volume change. Take a Playwright screenshot.

- [ ] **Step 4: Clean up**

Delete the throwaway stack (UI ⋮ → delete, or `DELETE /releases`-style stack delete) to free the cluster.

- [ ] **Step 5: Record the result**

Note the verified diff (with screenshot path) in the PR description / a short handoff under `docs/superpowers/handoffs/`.

---

## Self-Review

- **Spec coverage:** types (T1), volumes (T2), connections (T3), empty-state copy (T4), active-release diff (T5), live verify (T6). All spec sections mapped.
- **Type consistency:** `SnapshotDiff`/`ItemDiff` defined in T2, consumed identically in T3/T4/T5; `diffSnapshots(prev?, cur?): SnapshotDiff` stable across tasks; `ConfigDiff({ diff, hasPrev, prevSeq })` consistent T2→T5.
- **Placeholders:** none — every code step shows code; the one judgement call (whether `outcome` is typed on `StackReleaseDetail`) is explicit with a fallback.
