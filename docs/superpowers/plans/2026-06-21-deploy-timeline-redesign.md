# Deploy Timeline — Deployments Tab Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the Deployments tab as a vertical "deploy timeline" rail — the newest release is a live current node, every earlier release expands inline into a post-mortem (resource outcomes + resource-grouped config diff) — replacing the card + flat history + Sheet drawer.

**Architecture:** One container (`DeploymentsTab`, reusing `useReleases`) renders a `TimelineRail`. The rail draws a dot/connector spine: a `CurrentReleaseNode` (status pill, stage tracker, expandable resource rows, recovered note, release-error block, own changelog) then `HistoryRow[]` each lazily fetching `getRelease` on expand to show a post-mortem. New pure logic: `release-snapshot-diff.ts` (resource-grouped diff) + `use-release-detail.ts` (lazy fetch + cache). All status/stage/failure/log/banner visuals reuse existing `branded/` primitives + `index.css` tokens (no raw hex). `derive.ts` and `use-releases.ts` are reused unchanged (plus small additive helpers).

**Tech Stack:** React 19, TypeScript, Vite, Tailwind v4, Radix (DropdownMenu), lucide-react, Vitest + @testing-library/react (jsdom), @testing-library/user-event v14, axios (`@/api/client`).

**Spec:** `docs/superpowers/specs/2026-06-21-deploy-timeline-redesign-design.md`

## Global Constraints

- **No raw hex / off-scale type.** Use `index.css` tokens via Tailwind classes (`text-foreground`, `text-muted-foreground`, `text-fg-muted`, `bg-card`, `bg-muted`, `border-border`, `text-primary`, `bg-warn-bg`/`text-warn`/`border-warn-border`, `text-danger`/`bg-danger-bg`/`border-danger-border`, `text-success`/`bg-success-bg`/`border-success-border`, `font-mono`, `font-sans`) and reuse `branded/`/`ui/` primitives. (memory: stick to the Stackdome brand design system)
- **TDD always:** every task is test-first — write the failing test, run it red, implement, run it green, commit. One logical change per commit.
- **TS clean:** only the pre-existing `postgres-backups.ts` error is tolerated; touched files must add no new `tsc -b` errors. `pnpm --prefix frontend lint` stays at zero new errors.
- **Reuse, don't reinvent:** `derive.ts` (deriveStages, deriveFailingResources, deriveRecovered, releaseGitSha, causeLabel, formatDuration, humanizeFailureType) and `use-releases.ts` are reused unchanged. Branded primitives: `StatusPill`/`variantFromState`, `StageTracker`/`Stages`, `FailureCard`, `LogSnapshot`, `AlertBanner`, `EmptyState`, `Panel`.
- **Folder:** new components live in `frontend/src/pages/stacks/components/detail/deployments/timeline/`; tests in `.../timeline/tests/`. Test file header: `// @vitest-environment jsdom`.
- **API:** `@/api/releases` — `listReleases/getRelease/createRelease/rollbackRelease/cancelRelease(orgId, teamName, stackId, [releaseId])`. `getRelease` returns the full release incl. `snapshot` (untyped JSONB) + `outcome`. List payload omits both.

---

## File Structure

**New — logic:**
- `deployments/release-snapshot-diff.ts` — resource-grouped snapshot diff (replaces `release-diff.ts`).
- `deployments/use-release-detail.ts` — lazy `getRelease` + `Map` cache hook.
- `deployments/derive.ts` — **extend** with `phaseTone`, `toneTextClass`, `toneDotClass` (additive; existing exports untouched).

**New — components (`deployments/timeline/`):**
- `rail-node.tsx` — dot + vertical connector + content slot.
- `outcomes-table.tsx` — resource outcomes grid.
- `config-diff.tsx` — renders `ResourceDiff[]`.
- `release-post-mortem.tsx` — lazy detail: why-failed + outcomes + diff.
- `resource-row.tsx` — one live resource; expand failing → FailureCard + LogSnapshot + Open in Logs.
- `release-menu.tsx` — ⋮ dropdown, state-gated.
- `current-release-node.tsx` — current release card body.
- `history-row.tsx` — one history line + menu + post-mortem toggle.
- `banners.tsx` — `DriftBanner` + `ReleaseErrorBanner`.
- `timeline-rail.tsx` — composes current node + earlier marker + history rows + show-more + empty.

**Modified:**
- `deployments/deployments-tab.tsx` — rewritten container (same props; renders `TimelineRail`; adds optional `onOpenLogs`).
- `stacks/components/detail/index.tsx` — pass `onOpenLogs` (switch to Logs tab) to `DeploymentsTab`.

**Removed (Task 15):** `current-deployment-card.tsx`, `release-history.tsx`, `release-row.tsx`, `unreleased-changes-banner.tsx`, `failing-resources-accordion.tsx`, `release-detail-drawer.tsx`, `release-diff.ts` + their tests; `src/pages/__demo/`; the `App.tsx` demo route + import; `deploy-*.jpeg`; `docs/superpowers/handoffs/2026-06-21-deployments-LOCAL-VERIFICATION.md`.

---

### Task 1: Resource-grouped snapshot diff

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/release-snapshot-diff.ts`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts`

**Interfaces:**
- Consumes: `components["schemas"]["StackResource"]` from `@/api/types/openapi`.
- Produces:
  ```ts
  export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
  export interface DiffSection { kind: "configuration" | "environment"; rows: DiffRow[]; }
  export interface ResourceDiff { name: string; change: "added" | "removed" | "modified"; sections: DiffSection[]; note?: string; }
  export function diffSnapshots(prev: unknown, cur: unknown): ResourceDiff[];
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect } from "vitest";
import { diffSnapshots } from "../release-snapshot-diff";

const snap = (resources: unknown[]) => ({ resources });
const web = (over: Record<string, unknown> = {}) => ({
  name: "web",
  image_config: { image: "web:1" },
  ports: [{ number: 3000 }],
  execution_config: { command: ["node", "a.js"], env: [{ name: "LOG", value: "info" }] },
  ...over,
});

describe("diffSnapshots", () => {
  it("returns [] when nothing changed", () => {
    expect(diffSnapshots(snap([web()]), snap([web()]))).toEqual([]);
  });

  it("flags a modified image as a changed configuration row", () => {
    const out = diffSnapshots(snap([web()]), snap([web({ image_config: { image: "web:2" } })]));
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "web", change: "modified" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows).toContainEqual({ key: "image_config.image", from: "web:1", to: "web:2", kind: "changed" });
  });

  it("splits env changes into an environment section", () => {
    const out = diffSnapshots(
      snap([web()]),
      snap([web({ execution_config: { command: ["node", "a.js"], env: [{ name: "LOG", value: "debug" }, { name: "NEW", value: "1" }] } })]),
    );
    const env = out[0].sections.find((s) => s.kind === "environment")!;
    expect(env.rows).toContainEqual({ key: "LOG", from: "info", to: "debug", kind: "changed" });
    expect(env.rows).toContainEqual({ key: "NEW", to: "1", kind: "added" });
  });

  it("marks an added resource with all rows as added", () => {
    const out = diffSnapshots(snap([]), snap([web()]));
    expect(out[0]).toMatchObject({ name: "web", change: "added" });
    const cfg = out[0].sections.find((s) => s.kind === "configuration")!;
    expect(cfg.rows.every((r) => r.kind === "added")).toBe(true);
  });

  it("marks a removed resource with a note and no sections", () => {
    const out = diffSnapshots(snap([web()]), snap([]));
    expect(out[0]).toMatchObject({ name: "web", change: "removed", sections: [] });
    expect(out[0].note).toMatch(/removed/i);
  });

  it("returns [] when there is no previous snapshot", () => {
    expect(diffSnapshots(undefined, snap([web()]))).toEqual([]);
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- release-snapshot-diff`
Expected: FAIL — `diffSnapshots is not a function`.

- [ ] **Step 3: Write minimal implementation**

```ts
import type { components } from "@/api/types/openapi";

type SnapResource = components["schemas"]["StackResource"];

export interface DiffRow { key: string; from?: string; to?: string; kind: "added" | "removed" | "changed"; }
export interface DiffSection { kind: "configuration" | "environment"; rows: DiffRow[]; }
export interface ResourceDiff { name: string; change: "added" | "removed" | "modified"; sections: DiffSection[]; note?: string; }

function resourcesOf(snap: unknown): SnapResource[] {
  const r = (snap as { resources?: SnapResource[] } | null | undefined)?.resources;
  return Array.isArray(r) ? r : [];
}

function configScalars(r: SnapResource): Record<string, string | undefined> {
  return {
    "image_config.image": r.image_config?.image,
    "ports": (r.ports ?? []).map((p) => p.number).join(", ") || undefined,
    "execution_config.command": (r.execution_config?.command ?? []).join(" ") || undefined,
    "execution_config.args": (r.execution_config?.args ?? []).join(" ") || undefined,
  };
}

function envMap(r: SnapResource): Record<string, string> {
  const out: Record<string, string> = {};
  for (const e of r.execution_config?.env ?? []) {
    if (e?.name) out[e.name] = e.value ?? (e.secret_key_ref ? "(secret)" : "");
  }
  return out;
}

function configRows(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffRow[] {
  const p = prev ? configScalars(prev) : {};
  const c = cur ? configScalars(cur) : {};
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key];
    const to = c[key];
    if (from === to) continue;
    if (from == null) rows.push({ key, to, kind: "added" });
    else if (to == null) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

function envRows(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffRow[] {
  const p = prev ? envMap(prev) : {};
  const c = cur ? envMap(cur) : {};
  const rows: DiffRow[] = [];
  for (const key of new Set([...Object.keys(p), ...Object.keys(c)])) {
    const from = p[key];
    const to = c[key];
    if (from === to) continue;
    if (from === undefined) rows.push({ key, to, kind: "added" });
    else if (to === undefined) rows.push({ key, from, kind: "removed" });
    else rows.push({ key, from, to, kind: "changed" });
  }
  return rows;
}

function sectionsFor(prev: SnapResource | undefined, cur: SnapResource | undefined): DiffSection[] {
  const sections: DiffSection[] = [];
  const cfg = configRows(prev, cur);
  if (cfg.length) sections.push({ kind: "configuration", rows: cfg });
  const env = envRows(prev, cur);
  if (env.length) sections.push({ kind: "environment", rows: env });
  return sections;
}

export function diffSnapshots(prev: unknown, cur: unknown): ResourceDiff[] {
  if (prev == null) return []; // initial release — nothing to compare
  const prevByName = new Map(resourcesOf(prev).map((r) => [r.name ?? "", r]));
  const curByName = new Map(resourcesOf(cur).map((r) => [r.name ?? "", r]));
  const out: ResourceDiff[] = [];
  for (const name of new Set([...prevByName.keys(), ...curByName.keys()])) {
    const p = prevByName.get(name);
    const c = curByName.get(name);
    if (p && !c) { out.push({ name, change: "removed", sections: [], note: "Resource removed from this release — workload and config deleted from the stack." }); continue; }
    if (!p && c) { out.push({ name, change: "added", sections: sectionsFor(undefined, c) }); continue; }
    const sections = sectionsFor(p, c);
    if (sections.length) out.push({ name, change: "modified", sections });
  }
  return out;
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- release-snapshot-diff`
Expected: PASS (6 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-snapshot-diff.ts frontend/src/pages/stacks/components/detail/deployments/tests/release-snapshot-diff.test.ts
git commit -m "feat(frontend): resource-grouped snapshot diff for release post-mortem"
```

---

### Task 2: Tone helpers in derive.ts

**Files:**
- Modify: `frontend/src/pages/stacks/components/detail/deployments/derive.ts` (append; do not touch existing exports)
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts` (append cases)

**Interfaces:**
- Produces:
  ```ts
  export type Tone = "ok" | "amber" | "err" | "muted";
  export function phaseTone(phase: string): Tone;
  export function toneTextClass(t: Tone): string;   // Tailwind text color class
  export function toneDotClass(t: Tone): string;     // Tailwind bg color class for a dot
  ```

- [ ] **Step 1: Write the failing test** (append to `tests/derive.test.ts`)

```ts
import { phaseTone, toneTextClass, toneDotClass } from "../derive";

describe("phaseTone", () => {
  it("maps phases to tones", () => {
    expect(phaseTone("Ready")).toBe("ok");
    expect(phaseTone("Progressing")).toBe("amber");
    expect(phaseTone("Building")).toBe("amber");
    expect(phaseTone("CrashLoopBackOff")).toBe("err");
    expect(phaseTone("OOMKilled")).toBe("err");
    expect(phaseTone("ImagePullBackOff")).toBe("err");
    expect(phaseTone("Pending")).toBe("muted");
  });
  it("maps tones to brand token classes", () => {
    expect(toneTextClass("ok")).toBe("text-success");
    expect(toneTextClass("err")).toBe("text-danger");
    expect(toneTextClass("amber")).toBe("text-warn");
    expect(toneDotClass("muted")).toBe("bg-fg-muted");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- deployments/tests/derive`
Expected: FAIL — `phaseTone is not a function`.

- [ ] **Step 3: Write minimal implementation** (append to `derive.ts`)

```ts
export type Tone = "ok" | "amber" | "err" | "muted";

export function phaseTone(phase: string): Tone {
  if (/ready|available|running|healthy/i.test(phase)) return "ok";
  if (/progress|building|deploying|pending.*build/i.test(phase)) return "amber";
  if (/crash|oom|error|failed|imagepull|backoff/i.test(phase)) return "err";
  return "muted";
}

export function toneTextClass(t: Tone): string {
  return { ok: "text-success", amber: "text-warn", err: "text-danger", muted: "text-fg-muted" }[t];
}

export function toneDotClass(t: Tone): string {
  return { ok: "bg-success", amber: "bg-warn", err: "bg-danger", muted: "bg-fg-muted" }[t];
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- deployments/tests/derive`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/derive.ts frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts
git commit -m "feat(frontend): tone helpers for timeline phase coloring"
```

---

### Task 3: useReleaseDetail hook (lazy getRelease + cache)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/use-release-detail.ts`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/use-release-detail.test.tsx`

**Interfaces:**
- Consumes: `getRelease` from `@/api/releases`.
- Produces:
  ```ts
  export interface DetailState { data?: StackRelease; loading: boolean; error?: string; }
  export interface ReleaseDetail { ensure: (id: string) => void; peek: (id?: string) => DetailState; }
  export function useReleaseDetail(orgId: string, teamName: string, stackId: string): ReleaseDetail;
  ```
  `ensure(id)` fetches once and caches by id; re-calls are no-ops while cached. `peek(id)` returns the cached state (`{ loading: false }` if never requested, so the previous-release lookup reuses an already-fetched release).

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, waitFor, act } from "@testing-library/react";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { useReleaseDetail } from "../use-release-detail";

afterEach(cleanup);

function Harness({ ids }: { ids: string[] }) {
  const detail = useReleaseDetail("o", "t", "s");
  ids.forEach((id) => detail.ensure(id));
  return <div>{ids.map((id) => <span key={id} data-id={id}>{detail.peek(id).data?.sequence ?? "—"}</span>)}</div>;
}

describe("useReleaseDetail", () => {
  it("fetches once per id and caches", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r1", sequence: 12 });
    render(<Harness ids={["r1"]} />);
    await waitFor(() => expect(screen.getByText("12")).toBeInTheDocument());
    expect(getRelease).toHaveBeenCalledTimes(1);
    expect(getRelease).toHaveBeenCalledWith("o", "t", "s", "r1");
  });

  it("reuses a fetched release as another row's previous (no double fetch)", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation((_o, _t, _s, id) => Promise.resolve({ id, sequence: id === "r1" ? 12 : 11 }));
    const { rerender } = render(<Harness ids={["r1"]} />);
    await waitFor(() => expect(getRelease).toHaveBeenCalledTimes(1));
    rerender(<Harness ids={["r1", "r2"]} />);
    await waitFor(() => expect(getRelease).toHaveBeenCalledTimes(2)); // only r2 added
    rerender(<Harness ids={["r1", "r2"]} />);
    expect(getRelease).toHaveBeenCalledTimes(2); // r1/r2 already cached
  });

  it("captures error", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("boom"));
    function E() { const d = useReleaseDetail("o", "t", "s"); d.ensure("rx"); return <span>{d.peek("rx").error ?? ""}</span>; }
    render(<E />);
    await waitFor(() => expect(screen.getByText("boom")).toBeInTheDocument());
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- use-release-detail`
Expected: FAIL — `useReleaseDetail is not a function`.

- [ ] **Step 3: Write minimal implementation**

```ts
import { useCallback, useRef, useState } from "react";
import { getRelease, type StackRelease } from "@/api/releases";

export interface DetailState { data?: StackRelease; loading: boolean; error?: string; }
export interface ReleaseDetail { ensure: (id: string) => void; peek: (id?: string) => DetailState; }

const EMPTY: DetailState = { loading: false };

export function useReleaseDetail(orgId: string, teamName: string, stackId: string): ReleaseDetail {
  const [cache, setCache] = useState<Record<string, DetailState>>({});
  const inFlight = useRef<Set<string>>(new Set());

  const ensure = useCallback((id: string) => {
    if (!id || inFlight.current.has(id)) return;
    setCache((c) => (c[id] ? c : { ...c, [id]: { loading: true } }));
    if (cache[id]?.data || cache[id]?.error) return;
    inFlight.current.add(id);
    getRelease(orgId, teamName, stackId, id)
      .then((data) => setCache((c) => ({ ...c, [id]: { loading: false, data } })))
      .catch((e) => setCache((c) => ({ ...c, [id]: { loading: false, error: e instanceof Error ? e.message : "Failed to load release" } })))
      .finally(() => inFlight.current.delete(id));
  }, [orgId, teamName, stackId, cache]);

  const peek = useCallback((id?: string): DetailState => (id ? cache[id] ?? EMPTY : EMPTY), [cache]);
  return { ensure, peek };
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- use-release-detail`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/use-release-detail.ts frontend/src/pages/stacks/components/detail/deployments/tests/use-release-detail.test.tsx
git commit -m "feat(frontend): lazy getRelease cache hook for inline post-mortem"
```

---

### Task 4: RailNode (dot + connector + content)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/rail-node.tsx`
- Test: `.../timeline/tests/rail-node.test.tsx`

**Interfaces:**
- Consumes: `Tone`, `toneDotClass` from `../derive`.
- Produces:
  ```ts
  export interface RailNodeProps { tone: Tone; big?: boolean; pulse?: boolean; isLast?: boolean; children: React.ReactNode; }
  export function RailNode(props: RailNodeProps): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { RailNode } from "../rail-node";

afterEach(cleanup);

describe("RailNode", () => {
  it("renders content and a dot", () => {
    render(<RailNode tone="ok"><span>node body</span></RailNode>);
    expect(screen.getByText("node body")).toBeInTheDocument();
    expect(screen.getByTestId("rail-dot")).toHaveClass("bg-success");
  });
  it("hides the connector on the last node", () => {
    render(<RailNode tone="muted" isLast><span>x</span></RailNode>);
    expect(screen.getByTestId("rail-connector")).toHaveClass("invisible");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- rail-node`
Expected: FAIL — cannot find `../rail-node`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import type { Tone } from "../derive";
import { toneDotClass } from "../derive";

export interface RailNodeProps { tone: Tone; big?: boolean; pulse?: boolean; isLast?: boolean; children: React.ReactNode; }

export function RailNode({ tone, big, pulse, isLast, children }: RailNodeProps) {
  return (
    <div className="flex items-stretch gap-3.5">
      <div className="flex w-[34px] flex-none flex-col items-center">
        <span
          data-testid="rail-dot"
          className={[
            "mt-0.5 flex-none rounded-full",
            big ? "h-[15px] w-[15px] border-2 border-current bg-background" : "h-2.5 w-2.5",
            big ? toneDotClass(tone).replace("bg-", "text-") : toneDotClass(tone),
            pulse ? "animate-pulse" : "",
          ].join(" ")}
        />
        <span
          data-testid="rail-connector"
          className={["mt-1 w-px flex-1 bg-border", isLast ? "invisible" : "visible"].join(" ")}
          style={{ minHeight: isLast ? 0 : 16 }}
        />
      </div>
      <div className={["min-w-0 flex-1", isLast ? "pb-1" : "pb-5"].join(" ")}>{children}</div>
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- rail-node`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/rail-node.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/rail-node.test.tsx
git commit -m "feat(frontend): timeline RailNode (dot + connector)"
```

---

### Task 5: OutcomesTable

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/outcomes-table.tsx`
- Test: `.../timeline/tests/outcomes-table.test.tsx`

**Interfaces:**
- Consumes: `phaseTone`, `toneTextClass` from `../../derive`.
- Produces:
  ```ts
  export interface OutcomeRow { phase?: string; ready_replicas?: number; replicas?: number; message?: string; }
  export interface OutcomesTableProps { outcomes: Record<string, OutcomeRow>; }
  export function OutcomesTable(props: OutcomesTableProps): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { OutcomesTable } from "../outcomes-table";

afterEach(cleanup);

describe("OutcomesTable", () => {
  it("renders a row per resource with phase, replicas, message", () => {
    render(<OutcomesTable outcomes={{
      web: { phase: "Ready", ready_replicas: 1, replicas: 1, message: "" },
      worker: { phase: "CrashLoopBackOff", ready_replicas: 0, replicas: 1, message: "exit 1" },
    }} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("1/1")).toBeInTheDocument();
    expect(screen.getByText("CrashLoopBackOff")).toHaveClass("text-danger");
    expect(screen.getByText("exit 1")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- outcomes-table`
Expected: FAIL — cannot find `../outcomes-table`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { phaseTone, toneTextClass } from "../../derive";

export interface OutcomeRow { phase?: string; ready_replicas?: number; replicas?: number; message?: string; }
export interface OutcomesTableProps { outcomes: Record<string, OutcomeRow>; }

export function OutcomesTable({ outcomes }: OutcomesTableProps) {
  const rows = Object.entries(outcomes);
  return (
    <div>
      <div className="grid grid-cols-[1fr_90px_60px_1fr] gap-2 pb-1.5 font-mono text-[10px] uppercase tracking-wide text-fg-muted">
        <span>Resource</span><span>Phase</span><span>Repl.</span><span>Message</span>
      </div>
      {rows.map(([name, o]) => {
        const tone = phaseTone(o.phase ?? "");
        return (
          <div key={name} className="grid grid-cols-[1fr_90px_60px_1fr] items-center gap-2 border-t border-border py-2">
            <span className="font-mono text-[12px] font-semibold text-foreground">{name}</span>
            <span className={`text-[12.5px] ${toneTextClass(tone)}`}>{o.phase}</span>
            <span className="font-mono text-[11px] text-fg-muted">{o.ready_replicas ?? 0}/{o.replicas ?? 0}</span>
            <span className={`font-mono text-[11px] ${tone === "err" ? "text-danger" : "text-fg-muted"}`}>{o.message || "—"}</span>
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- outcomes-table`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/outcomes-table.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/outcomes-table.test.tsx
git commit -m "feat(frontend): OutcomesTable for release post-mortem"
```

---

### Task 6: ConfigDiff

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/config-diff.tsx`
- Test: `.../timeline/tests/config-diff.test.tsx`

**Interfaces:**
- Consumes: `ResourceDiff` from `../../release-snapshot-diff`.
- Produces:
  ```ts
  export interface ConfigDiffProps { diffs: ResourceDiff[]; prevSeq?: number; }
  export function ConfigDiff(props: ConfigDiffProps): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ConfigDiff } from "../config-diff";
import type { ResourceDiff } from "../../release-snapshot-diff";

afterEach(cleanup);

const diffs: ResourceDiff[] = [
  { name: "web", change: "modified", sections: [{ kind: "configuration", rows: [{ key: "image_config.image", from: "web:1", to: "web:2", kind: "changed" }] }] },
  { name: "worker", change: "added", sections: [{ kind: "configuration", rows: [{ key: "image_config.image", to: "worker:1", kind: "added" }] }] },
  { name: "mailhog", change: "removed", sections: [], note: "Resource removed from this release — workload and config deleted from the stack." },
];

describe("ConfigDiff", () => {
  it("renders changed, added and removed resources", () => {
    render(<ConfigDiff diffs={diffs} prevSeq={12} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("web:1")).toHaveClass("line-through");
    expect(screen.getByText("web:2")).toHaveClass("text-success");
    expect(screen.getByText("ADDED")).toBeInTheDocument();
    expect(screen.getByText(/removed from this release/i)).toBeInTheDocument();
  });
  it("shows an empty note when there are no diffs", () => {
    render(<ConfigDiff diffs={[]} />);
    expect(screen.getByText(/nothing to compare/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- config-diff`
Expected: FAIL — cannot find `../config-diff`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import type { ResourceDiff, DiffRow } from "../../release-snapshot-diff";

function Row({ row }: { row: DiffRow }) {
  return (
    <div className="flex items-start gap-2.5 px-3 pb-2 pt-1">
      <span className="w-[150px] flex-none font-mono text-[11px] text-fg-muted">{row.key}</span>
      <span className="flex flex-wrap items-center gap-1.5 font-mono text-[11px]">
        {row.kind === "added" && <span className="text-success">— → {row.to}</span>}
        {row.kind === "removed" && <span className="text-danger line-through">{row.from}</span>}
        {row.kind === "changed" && (
          <>
            <span className="text-danger line-through opacity-80">{row.from}</span>
            <span className="text-fg-muted">→</span>
            <span className="text-success">{row.to}</span>
          </>
        )}
      </span>
    </div>
  );
}

export interface ConfigDiffProps { diffs: ResourceDiff[]; prevSeq?: number; }

export function ConfigDiff({ diffs }: ConfigDiffProps) {
  if (!diffs.length) return <div className="text-[12.5px] text-fg-muted">Initial release — nothing to compare.</div>;
  return (
    <div className="space-y-2.5">
      {diffs.map((d) => {
        const dot = d.change === "added" ? "bg-success" : d.change === "removed" ? "bg-fg-muted" : "bg-warn";
        return (
          <div key={d.name} className={`overflow-hidden rounded-md border border-border ${d.change === "removed" ? "opacity-80" : ""}`}>
            <div className="flex items-center gap-2.5 bg-muted px-3 py-2.5">
              <span className={`h-[7px] w-[7px] flex-none rounded-full ${dot}`} />
              <span className={`font-mono text-[12.5px] font-semibold text-foreground ${d.change === "removed" ? "line-through" : ""}`}>{d.name}</span>
              {d.change === "added" && <span className="rounded border border-success px-1.5 py-0.5 font-mono text-[9px] uppercase text-success">Added</span>}
              {d.change === "removed" && <span className="rounded border border-fg-muted px-1.5 py-0.5 font-mono text-[9px] uppercase text-fg-muted">Removed</span>}
            </div>
            {d.change === "removed" ? (
              <div className="flex items-start gap-2.5 border-t border-border px-3 py-2.5 text-[12px] text-fg-muted">
                <span className="flex-none">−</span><span>{d.note}</span>
              </div>
            ) : (
              d.sections.map((sec, si) => (
                <div key={si} className="border-t border-border">
                  <div className="px-3 pb-0.5 pt-2 font-mono text-[9px] uppercase tracking-wide text-fg-muted">{sec.kind}</div>
                  {sec.rows.map((row, ri) => <Row key={ri} row={row} />)}
                </div>
              ))
            )}
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- config-diff`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/config-diff.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/config-diff.test.tsx
git commit -m "feat(frontend): ConfigDiff renderer for resource-grouped diff"
```

---

### Task 7: ReleasePostMortem (lazy detail)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/release-post-mortem.tsx`
- Test: `.../timeline/tests/release-post-mortem.test.tsx`

**Interfaces:**
- Consumes: `ReleaseDetail` from `../../use-release-detail`; `OutcomesTable`; `ConfigDiff`; `diffSnapshots` from `../../release-snapshot-diff`; `StackRelease`.
- Produces:
  ```ts
  export interface ReleasePostMortemProps {
    detail: ReleaseDetail;        // from useReleaseDetail
    release: StackRelease;        // the list row (state, message, sequence, id)
    prevReleaseId?: string;
    prevSeq?: number;
  }
  export function ReleasePostMortem(props: ReleasePostMortemProps): JSX.Element;
  ```
  On mount calls `detail.ensure(release.id)` + `detail.ensure(prevReleaseId)`. Renders: Why-it-failed (when `release.state === "Failed" && release.message`), OutcomesTable (from `peek(release.id).data.outcome.resources`), ConfigDiff (`diffSnapshots(prev.snapshot, cur.snapshot)`). Loading + error lines.

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { useReleaseDetail } from "../../use-release-detail";
import { ReleasePostMortem } from "../release-post-mortem";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);

function Wrap({ release, prevId }: { release: StackRelease; prevId?: string }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <ReleasePostMortem detail={detail} release={release} prevReleaseId={prevId} prevSeq={12} />;
}

describe("ReleasePostMortem", () => {
  it("shows outcomes + config diff once loaded", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockImplementation((_o, _t, _s, id) =>
      Promise.resolve(id === "r-cur"
        ? { id, sequence: 13, outcome: { resources: { web: { phase: "Ready", ready_replicas: 1, replicas: 1 } } }, snapshot: { resources: [{ name: "web", image_config: { image: "web:2" } }] } }
        : { id, sequence: 12, snapshot: { resources: [{ name: "web", image_config: { image: "web:1" } }] } }));
    render(<Wrap release={{ id: "r-cur", sequence: 13, state: "Released" } as StackRelease} prevId="r-prev" />);
    await waitFor(() => expect(screen.getByText("web")).toBeInTheDocument());
    expect(screen.getByText("Resource outcomes")).toBeInTheDocument();
    expect(screen.getByText("web:2")).toBeInTheDocument();
  });

  it("shows why-it-failed for a failed release", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({ id: "r-cur", sequence: 9, outcome: { resources: {} }, snapshot: { resources: [] } });
    render(<Wrap release={{ id: "r-cur", sequence: 9, state: "Failed", message: "apply error: quota" } as StackRelease} />);
    await waitFor(() => expect(screen.getByText(/apply error: quota/)).toBeInTheDocument());
    expect(screen.getByText("Why it failed")).toBeInTheDocument();
  });

  it("shows an error line if the fetch fails", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockRejectedValue(new Error("nope"));
    render(<Wrap release={{ id: "r-cur", sequence: 5, state: "Released" } as StackRelease} />);
    await waitFor(() => expect(screen.getByText(/nope/)).toBeInTheDocument());
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- release-post-mortem`
Expected: FAIL — cannot find `../release-post-mortem`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { useEffect } from "react";
import type { StackRelease } from "@/api/releases";
import type { ReleaseDetail } from "../../use-release-detail";
import { diffSnapshots } from "../../release-snapshot-diff";
import { OutcomesTable } from "./outcomes-table";
import { ConfigDiff } from "./config-diff";

const Marker = ({ children, tone }: { children: React.ReactNode; tone?: string }) => (
  <div className={`mb-2.5 font-mono text-[11px] uppercase tracking-wide ${tone ?? "text-fg-muted"}`}>{children}</div>
);

export interface ReleasePostMortemProps {
  detail: ReleaseDetail;
  release: StackRelease;
  prevReleaseId?: string;
  prevSeq?: number;
}

export function ReleasePostMortem({ detail, release, prevReleaseId, prevSeq }: ReleasePostMortemProps) {
  useEffect(() => {
    if (release.id) detail.ensure(release.id);
    if (prevReleaseId) detail.ensure(prevReleaseId);
  }, [detail, release.id, prevReleaseId]);

  const cur = detail.peek(release.id);
  const prev = detail.peek(prevReleaseId);

  if (cur.loading && !cur.data) return <div className="px-0.5 py-3 text-[12.5px] text-fg-muted">Loading release detail…</div>;
  if (cur.error) return <div className="px-0.5 py-3 text-[12.5px] text-danger">Could not load detail: {cur.error}</div>;

  const data = cur.data;
  const outcomes = (data as unknown as { outcome?: { resources?: Record<string, { phase?: string; ready_replicas?: number; replicas?: number; message?: string }> } })?.outcome?.resources ?? {};
  const snap = (data as unknown as { snapshot?: unknown })?.snapshot;
  const prevSnap = (prev.data as unknown as { snapshot?: unknown })?.snapshot;
  const diffs = diffSnapshots(prevSnap, snap);

  return (
    <div className="space-y-4 px-0.5 pb-1.5 pt-3.5">
      {release.state === "Failed" && release.message && (
        <div>
          <Marker tone="text-danger">Why it failed</Marker>
          <div className="flex items-start gap-2 font-mono text-[11.5px] leading-relaxed text-foreground">
            <span className="flex-none text-danger">⊘</span><span>{release.message}</span>
          </div>
        </div>
      )}
      {Object.keys(outcomes).length > 0 && (
        <div><Marker>Resource outcomes</Marker><OutcomesTable outcomes={outcomes} /></div>
      )}
      {prevReleaseId && (
        <div><Marker>Config changes · vs #{prevSeq ?? "previous"}</Marker><ConfigDiff diffs={diffs} prevSeq={prevSeq} /></div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- release-post-mortem`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/release-post-mortem.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/release-post-mortem.test.tsx
git commit -m "feat(frontend): inline ReleasePostMortem (why-failed + outcomes + diff)"
```

---

### Task 8: ResourceRow (live resource, expandable failure)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/resource-row.tsx`
- Test: `.../timeline/tests/resource-row.test.tsx`

**Interfaces:**
- Consumes: `FailingResource` from `../../derive`; `phaseTone`/`toneTextClass`/`toneDotClass` from `../../derive`; `FailureCard`, `LogSnapshot` from `@/components/branded`; `fetchLogSnapshot` from `@/api/observability`.
- Produces:
  ```ts
  export interface LogContext { orgId: string; teamName: string; stackId: string; }
  export interface ResourceRowVM { name: string; phase: string; replicas?: string; msg?: string; tag?: string; failure?: FailingResource; }
  export interface ResourceRowProps { vm: ResourceRowVM; logContext?: LogContext; onOpenLogs?: (name: string) => void; }
  export function ResourceRow(props: ResourceRowProps): JSX.Element;
  ```
  Only rows with `vm.failure` are clickable/expandable. Expanded → `FailureCard` (+ best-effort `LogSnapshot` via `fetchLogSnapshot` for `runtime_crash`, hidden when empty) + an "Open in Logs →" button (calls `onOpenLogs?.(name)`).

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
import { ResourceRow } from "../resource-row";

afterEach(cleanup);

describe("ResourceRow", () => {
  it("renders a healthy row without an expander", () => {
    render(<ResourceRow vm={{ name: "redis", phase: "Ready", replicas: "1/1" }} />);
    expect(screen.getByText("redis")).toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("expands a failing row to show the FailureCard and Open in Logs", async () => {
    const onOpenLogs = vi.fn();
    render(<ResourceRow
      vm={{ name: "web", phase: "CrashLoopBackOff", replicas: "0/1", failure: { name: "web", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", message: "exit 1", restartCount: 7 } }}
      logContext={{ orgId: "o", teamName: "t", stackId: "s" }}
      onOpenLogs={onOpenLogs}
    />);
    await userEvent.click(screen.getByText("web"));
    expect(screen.getByText("exit 1")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /open in logs/i }));
    expect(onOpenLogs).toHaveBeenCalledWith("web");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- resource-row`
Expected: FAIL — cannot find `../resource-row`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { useEffect, useState } from "react";
import { ChevronDown } from "lucide-react";
import { FailureCard, LogSnapshot } from "@/components/branded";
import { fetchLogSnapshot } from "@/api/observability";
import type { FailingResource } from "../../derive";
import { phaseTone, toneTextClass, toneDotClass } from "../../derive";

export interface LogContext { orgId: string; teamName: string; stackId: string; }
export interface ResourceRowVM { name: string; phase: string; replicas?: string; msg?: string; tag?: string; failure?: FailingResource; }
export interface ResourceRowProps { vm: ResourceRowVM; logContext?: LogContext; onOpenLogs?: (name: string) => void; }

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

function CrashLog({ ctx, name }: { ctx: LogContext; name: string }) {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    void fetchLogSnapshot(ctx.orgId, ctx.teamName, ctx.stackId, name, 50).then((l) => { if (alive) setLines(l); });
    return () => { alive = false; };
  }, [ctx.orgId, ctx.teamName, ctx.stackId, name]);
  if (!lines.length) return null;
  return <div className="mt-2.5"><LogSnapshot lines={lines} /></div>;
}

export function ResourceRow({ vm, logContext, onOpenLogs }: ResourceRowProps) {
  const [open, setOpen] = useState(false);
  const tone = phaseTone(vm.phase);
  const failing = !!vm.failure;
  return (
    <div>
      <div
        className={`flex items-center gap-2.5 py-2 ${failing ? "cursor-pointer" : ""}`}
        onClick={failing ? () => setOpen((o) => !o) : undefined}
      >
        <span className={`h-2 w-2 flex-none rounded-full ${toneDotClass(tone)}`} />
        <span className="w-[82px] flex-none font-mono text-[12.5px] font-semibold text-foreground">{vm.name}</span>
        <span className={`min-w-0 flex-1 text-[13px] ${toneTextClass(tone)}`}>{vm.phase}</span>
        {vm.tag && <span className="rounded border border-warn px-1.5 py-0.5 font-mono text-[9px] uppercase text-warn">{vm.tag}</span>}
        {vm.replicas && <span className="font-mono text-[11px] text-fg-muted">{vm.replicas}</span>}
        {vm.msg && <span className="max-w-[220px] truncate font-mono text-[11px] text-fg-muted">{vm.msg}</span>}
        {failing && <ChevronDown className={`h-3.5 w-3.5 text-fg-muted transition-transform ${open ? "rotate-180" : ""}`} />}
      </div>
      {open && vm.failure && (
        <div className="ml-[18px] mb-1 rounded-md border border-border bg-muted p-3">
          <FailureCard
            resourceName={vm.failure.name}
            stage={stageForCard(vm.failure.stage)}
            reason={vm.failure.reason}
            message={vm.failure.message}
            exitCode={vm.failure.exitCode}
            restartCount={vm.failure.restartCount}
          />
          {logContext && vm.failure.type === "runtime_crash" && <CrashLog ctx={logContext} name={vm.failure.name} />}
          {onOpenLogs && (
            <button
              onClick={() => onOpenLogs(vm.name)}
              className="mt-2.5 rounded bg-primary px-3 py-1.5 font-sans text-[12px] font-medium text-primary-foreground"
            >
              Open in Logs →
            </button>
          )}
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- resource-row`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/resource-row.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/resource-row.test.tsx
git commit -m "feat(frontend): expandable ResourceRow with failure detail + Open in Logs"
```

---

### Task 9: ReleaseMenu (⋮ state-gated)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/release-menu.tsx`
- Test: `.../timeline/tests/release-menu.test.tsx`

**Interfaces:**
- Consumes: `ui/dropdown-menu`; `StackRelease`.
- Produces:
  ```ts
  export interface ReleaseMenuProps {
    release: StackRelease;
    onView: (id: string) => void;
    onRollback: (id: string) => void;
    onCancel: (id: string) => void;
    onCopyId: (id: string) => void;
  }
  export function ReleaseMenu(props: ReleaseMenuProps): JSX.Element;
  ```
  Items: View details (always) · Rollback to this (state `Released`) · Cancel release (state `Pending`/`InProgress`) · Copy release ID (always).

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { ReleaseMenu } from "../release-menu";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = {
    hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined,
  };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const rel = (over: Partial<StackRelease>) => ({ id: "r1", sequence: 5, state: "Released", ...over } as StackRelease);

describe("ReleaseMenu", () => {
  it("shows Rollback for Released and Copy id; hides Cancel", async () => {
    const onRollback = vi.fn();
    render(<ReleaseMenu release={rel({ state: "Released" })} onView={vi.fn()} onRollback={onRollback} onCancel={vi.fn()} onCopyId={vi.fn()} />);
    await userEvent.click(screen.getByRole("button", { name: /release actions/i }));
    expect(screen.getByText("Rollback to this")).toBeInTheDocument();
    expect(screen.queryByText("Cancel release")).not.toBeInTheDocument();
    await userEvent.click(screen.getByText("Rollback to this"));
    expect(onRollback).toHaveBeenCalledWith("r1");
  });

  it("shows Cancel for Pending", async () => {
    render(<ReleaseMenu release={rel({ state: "Pending" })} onView={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} onCopyId={vi.fn()} />);
    await userEvent.click(screen.getByRole("button", { name: /release actions/i }));
    expect(screen.getByText("Cancel release")).toBeInTheDocument();
    expect(screen.queryByText("Rollback to this")).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- release-menu`
Expected: FAIL — cannot find `../release-menu`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { MoreVertical } from "lucide-react";
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { StackRelease } from "@/api/releases";

export interface ReleaseMenuProps {
  release: StackRelease;
  onView: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
}

export function ReleaseMenu({ release, onView, onRollback, onCancel, onCopyId }: ReleaseMenuProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label="Release actions" className="h-7 w-7" onClick={(e) => e.stopPropagation()}>
          <MoreVertical className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="min-w-[180px]">
        <DropdownMenuItem onClick={() => onView(id)}>View details</DropdownMenuItem>
        {state === "Released" && release.id && <DropdownMenuItem onClick={() => onRollback(release.id!)}>Rollback to this</DropdownMenuItem>}
        {(state === "Pending" || state === "InProgress") && release.id && (
          <DropdownMenuItem variant="destructive" onClick={() => onCancel(release.id!)}>Cancel release</DropdownMenuItem>
        )}
        <DropdownMenuItem onClick={() => onCopyId(id)}>Copy release ID</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- release-menu`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/release-menu.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/release-menu.test.tsx
git commit -m "feat(frontend): state-gated release ⋮ menu"
```

---

### Task 10: CurrentReleaseNode

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/current-release-node.tsx`
- Test: `.../timeline/tests/current-release-node.test.tsx`

**Interfaces:**
- Consumes: `StatusPill`/`variantFromState`, `StageTracker` (branded); `deriveStages`/`deriveFailingResources`/`deriveRecovered`/`releaseGitSha`/`formatDuration` (`../../derive`); `ResourceRow`/`LogContext`; `ReleasePostMortem`; `ReleaseDetail`; `StackRelease`, `Stack`.
- Produces:
  ```ts
  export interface CurrentReleaseNodeProps {
    release: StackRelease; stack: Stack; logContext?: LogContext;
    onOpenLogs?: (name: string) => void;
    detail: ReleaseDetail; prevReleaseId?: string; prevSeq?: number;
  }
  export function CurrentReleaseNode(props: CurrentReleaseNodeProps): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { useReleaseDetail } from "../../use-release-detail";
import { CurrentReleaseNode } from "../current-release-node";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);

const stack = (over: Record<string, unknown> = {}) => ({
  status: { resources: [{ name: "web", phase: "Ready", available_replicas: 1, replicas: 1 }], last_converged: { release_id: "r1" } },
  spec: { stack_resources: [{ name: "web", status: { state: "Ready" } }] },
  ...over,
}) as unknown as Stack;

function Wrap({ release, st }: { release: StackRelease; st: Stack }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <CurrentReleaseNode release={release} stack={st} detail={detail} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} />;
}

describe("CurrentReleaseNode", () => {
  it("renders status, sequence and a Ready resource row", () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released", snapshot_revision: "abcdef1234" } as StackRelease} st={stack()} />);
    expect(screen.getByText("Released")).toBeInTheDocument();
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText("web")).toBeInTheDocument();
  });

  it("renders the release-error block for a Failed release with no per-resource failure", () => {
    const st = stack({ status: { resources: [] }, spec: { stack_resources: [] } });
    render(<Wrap release={{ id: "r1", sequence: 17, state: "Failed", message: "apply error: unknown addon" } as StackRelease} st={st} />);
    expect(screen.getByText(/apply error: unknown addon/)).toBeInTheDocument();
  });

  it("toggles its own changelog", async () => {
    render(<Wrap release={{ id: "r1", sequence: 14, state: "Released" } as StackRelease} st={stack()} />);
    await userEvent.click(screen.getByRole("button", { name: /view changelog/i }));
    expect(await screen.findByText(/config changes|nothing to compare/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- current-release-node`
Expected: FAIL — cannot find `../current-release-node`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { useState } from "react";
import { StatusPill, StageTracker, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, releaseGitSha, formatDuration } from "../../derive";
import type { ReleaseDetail } from "../../use-release-detail";
import { ResourceRow, type LogContext } from "./resource-row";
import { ReleasePostMortem } from "./release-post-mortem";

export interface CurrentReleaseNodeProps {
  release: StackRelease; stack: Stack; logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  detail: ReleaseDetail; prevReleaseId?: string; prevSeq?: number;
}

function meta(release: StackRelease): string {
  const parts: string[] = [];
  if (release.snapshot_revision) parts.push(`config ${release.snapshot_revision.slice(0, 7)}`);
  const dur = formatDuration(release.rendered_at, release.completed_at);
  if (dur !== "—") parts.push(`took ${dur}`);
  return parts.join(" · ");
}

export function CurrentReleaseNode({ release, stack, logContext, onOpenLogs, detail, prevReleaseId, prevSeq }: CurrentReleaseNodeProps) {
  const [showChangelog, setShowChangelog] = useState(false);
  const failing = deriveFailingResources(stack);
  const recovered = deriveRecovered(stack);
  const recoveredNames = new Set(recovered.map((r) => r.name));
  const failingByName = new Map(failing.map((f) => [f.name, f]));
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];
  const releaseLevelError = release.state === "Failed" && failing.length === 0 && release.message;

  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex flex-wrap items-baseline justify-between gap-3">
        <div className="flex min-w-0 items-center gap-2.5">
          <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>
          <span className="font-sans text-[16px] font-semibold text-foreground">#{release.sequence}</span>
        </div>
        <span className="font-mono text-[11px] text-fg-muted">{meta(release)}</span>
      </div>

      <div className="mt-3.5"><StageTracker stages={stages} /></div>

      {releaseLevelError && (
        <div className="mt-3.5 rounded-md border border-danger-border bg-danger-bg p-3.5">
          <div className="mb-1.5 flex items-center gap-2 font-sans text-[13px] font-semibold text-danger">
            <span>⊘</span> Release failed
          </div>
          <div className="font-mono text-[11.5px] leading-relaxed text-foreground">{release.message}</div>
        </div>
      )}

      {summaries.length > 0 && (
        <div className="mt-4 divide-y divide-border">
          {summaries.map((s, i) => (
            <ResourceRow
              key={s.name ?? i}
              vm={{
                name: s.name ?? "",
                phase: s.phase ?? "",
                replicas: `${s.available_replicas ?? 0}/${s.replicas ?? 0}`,
                msg: s.message,
                tag: recoveredNames.has(s.name ?? "") ? "RECOVERED" : undefined,
                failure: failingByName.get(s.name ?? ""),
              }}
              logContext={logContext}
              onOpenLogs={onOpenLogs}
            />
          ))}
        </div>
      )}

      {recovered.length > 0 && (
        <div className="mt-4 rounded-md border border-warn-border bg-warn-bg px-3.5 py-2.5 text-[12.5px] text-fg-muted">
          {recovered.map((r) => (
            <div key={r.name}>
              <span className="font-medium text-foreground">{r.name}</span> recovered
              {r.restartCount != null ? ` after ${r.restartCount} ${r.restartCount === 1 ? "restart" : "restarts"}` : ""} — last failure{" "}
              <span className="text-warn">{r.reason}</span>
            </div>
          ))}
        </div>
      )}

      <button onClick={() => setShowChangelog((v) => !v)} className="mt-4 font-sans text-[12.5px] font-medium text-primary">
        {showChangelog ? "Hide changelog" : "View changelog"}
      </button>
      {showChangelog && (
        <ReleasePostMortem detail={detail} release={release} prevReleaseId={prevReleaseId} prevSeq={prevSeq} />
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- current-release-node`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/current-release-node.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/current-release-node.test.tsx
git commit -m "feat(frontend): CurrentReleaseNode (live status + changelog)"
```

---

### Task 11: HistoryRow (expandable post-mortem)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/history-row.tsx`
- Test: `.../timeline/tests/history-row.test.tsx`

**Interfaces:**
- Consumes: `StatusPill`/`variantFromState`; `causeLabel`/`releaseGitSha`/`formatDuration` (`../../derive`); `ReleaseMenu`; `ReleasePostMortem`; `ReleaseDetail`; `StackRelease`.
- Produces:
  ```ts
  export interface HistoryRowProps {
    release: StackRelease;
    prevReleaseId?: string; prevSeq?: number;
    detail: ReleaseDetail;
    isOpen: boolean;
    onToggle: (id: string) => void;
    onRollback: (id: string) => void;
    onCancel: (id: string) => void;
    onCopyId: (id: string) => void;
  }
  export function HistoryRow(props: HistoryRowProps): JSX.Element;
  ```
  Clicking the row (or the menu's "View details") toggles the post-mortem. Subline: `Failed` → `release.message` (danger); `Released` → `git <sha> · <duration>`.

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 13, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { useReleaseDetail } from "../../use-release-detail";
import { HistoryRow } from "../history-row";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

function Wrap({ release, isOpen, onToggle }: { release: StackRelease; isOpen: boolean; onToggle: (id: string) => void }) {
  const detail = useReleaseDetail("o", "t", "s");
  return <HistoryRow release={release} detail={detail} isOpen={isOpen} onToggle={onToggle} onRollback={vi.fn()} onCancel={vi.fn()} onCopyId={vi.fn()} />;
}

describe("HistoryRow", () => {
  it("shows cause + state and toggles open on click", async () => {
    const onToggle = vi.fn();
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released", cause: { kind: "rollback", detail: "9" } } as StackRelease} isOpen={false} onToggle={onToggle} />);
    expect(screen.getByText("#13")).toBeInTheDocument();
    expect(screen.getByText("Rollback to #9")).toBeInTheDocument();
    await userEvent.click(screen.getByText("#13"));
    expect(onToggle).toHaveBeenCalledWith("r1");
  });

  it("renders the post-mortem when open", async () => {
    render(<Wrap release={{ id: "r1", sequence: 13, state: "Released" } as StackRelease} isOpen onToggle={vi.fn()} />);
    expect(await screen.findByText(/config changes|nothing to compare/i)).toBeInTheDocument();
  });

  it("shows the failure message in danger for a Failed release", () => {
    render(<Wrap release={{ id: "r1", sequence: 9, state: "Failed", message: "apply quota error" } as StackRelease} isOpen={false} onToggle={vi.fn()} />);
    expect(screen.getByText(/apply quota error/)).toHaveClass("text-danger");
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- history-row`
Expected: FAIL — cannot find `../history-row`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { StatusPill, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import { causeLabel, releaseGitSha, formatDuration } from "../../derive";
import type { ReleaseDetail } from "../../use-release-detail";
import { ReleaseMenu } from "./release-menu";
import { ReleasePostMortem } from "./release-post-mortem";

export interface HistoryRowProps {
  release: StackRelease;
  prevReleaseId?: string; prevSeq?: number;
  detail: ReleaseDetail;
  isOpen: boolean;
  onToggle: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
}

export function HistoryRow({ release, prevReleaseId, prevSeq, detail, isOpen, onToggle, onRollback, onCancel, onCopyId }: HistoryRowProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  const sha = releaseGitSha(release);
  const subline = state === "Failed"
    ? release.message
    : state === "Released"
      ? [sha && `git ${sha}`, formatDuration(release.rendered_at, release.completed_at)].filter(Boolean).join(" · ")
      : undefined;

  return (
    <div>
      <div className="-mx-2 flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 hover:bg-muted" onClick={() => onToggle(id)}>
        <StatusPill variant={variantFromState(state)}>{state}</StatusPill>
        <span className="font-sans text-[13px] font-semibold text-foreground">#{release.sequence}</span>
        <span className="min-w-0 flex-1 truncate text-[13px] text-fg-muted">
          {causeLabel(release.cause)}
          {subline ? <span className={state === "Failed" ? "text-danger" : ""}> · {subline}</span> : null}
        </span>
        <ReleaseMenu release={release} onView={() => onToggle(id)} onRollback={onRollback} onCancel={onCancel} onCopyId={onCopyId} />
      </div>
      {isOpen && <ReleasePostMortem detail={detail} release={release} prevReleaseId={prevReleaseId} prevSeq={prevSeq} />}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- history-row`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/history-row.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/history-row.test.tsx
git commit -m "feat(frontend): HistoryRow with inline post-mortem + menu"
```

---

### Task 12: Banners (drift + release error)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/banners.tsx`
- Test: `.../timeline/tests/banners.test.tsx`

**Interfaces:**
- Consumes: `AlertBanner` (branded).
- Produces:
  ```ts
  export function DriftBanner(props: { onDeploy: () => void; busy: boolean }): JSX.Element;
  export function ReleaseErrorBanner(props: { lead: string; text: string; onView?: () => void }): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { DriftBanner, ReleaseErrorBanner } from "../banners";

afterEach(cleanup);

describe("banners", () => {
  it("drift banner deploys", async () => {
    const onDeploy = vi.fn();
    render(<DriftBanner onDeploy={onDeploy} busy={false} />);
    await userEvent.click(screen.getByRole("button", { name: /deploy/i }));
    expect(onDeploy).toHaveBeenCalled();
  });
  it("drift Deploy is disabled while busy", () => {
    render(<DriftBanner onDeploy={vi.fn()} busy />);
    expect(screen.getByRole("button", { name: /deploying/i })).toBeDisabled();
  });
  it("release error banner shows text + View details", async () => {
    const onView = vi.fn();
    render(<ReleaseErrorBanner lead="Deploy #16 failed" text="1 of 3 failing" onView={onView} />);
    expect(screen.getByText(/1 of 3 failing/)).toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /view details/i }));
    expect(onView).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- timeline/tests/banners`
Expected: FAIL — cannot find `../banners`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { Rocket } from "lucide-react";
import { AlertBanner } from "@/components/branded";

export function DriftBanner({ onDeploy, busy }: { onDeploy: () => void; busy: boolean }) {
  return (
    <AlertBanner action={{ label: busy ? "Deploying…" : "Deploy changes", onClick: onDeploy, disabled: busy }}>
      <span className="flex items-center gap-2">
        <Rocket className="h-3.5 w-3.5" />
        Unreleased changes — your saved configuration differs from the active deployment (approximate).
      </span>
    </AlertBanner>
  );
}

export function ReleaseErrorBanner({ lead, text, onView }: { lead: string; text: string; onView?: () => void }) {
  return (
    <AlertBanner action={onView ? { label: "View details", onClick: onView } : undefined}>
      <span className="flex items-center gap-2">
        <span className="font-semibold text-foreground">{lead}</span>
        <span className="truncate">{text}</span>
      </span>
    </AlertBanner>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- timeline/tests/banners`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/banners.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/banners.test.tsx
git commit -m "feat(frontend): drift + release-error timeline banners"
```

---

### Task 13: TimelineRail (composition + show-more window)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/timeline/timeline-rail.tsx`
- Test: `.../timeline/tests/timeline-rail.test.tsx`

**Interfaces:**
- Consumes: `RailNode`; `CurrentReleaseNode`; `HistoryRow`; `EmptyState` (branded); `useReleaseDetail`/`ReleaseDetail`; `StackRelease`, `Stack`; `LogContext`.
- Produces:
  ```ts
  export interface TimelineRailProps {
    releases: StackRelease[];       // newest-first
    activeRelease?: StackRelease;   // = releases[0] (undefined only when empty)
    stack: Stack;
    logContext?: LogContext;
    onOpenLogs?: (name: string) => void;
    banner?: React.ReactNode;       // DriftBanner / ReleaseErrorBanner, decided by container
    onRollback: (id: string) => void;
    onCancel: (id: string) => void;
    onCopyId: (id: string) => void;
    initialWindow?: number;         // default 15
  }
  export function TimelineRail(props: TimelineRailProps): JSX.Element;
  ```
  Layout: optional banner; if `activeRelease` → big `RailNode` → `CurrentReleaseNode`; then "Earlier releases" marker + `HistoryRow[]` (= `releases.slice(1)`, windowed) each in a small `RailNode`; "Show more" if windowed; `EmptyState` rail node when there are zero releases. `prevReleaseId`/`prevSeq` for `releases[i]` = `releases[i+1]`. Expansion (`openId`) is local state; the menu's View uses the same toggle.

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/api/releases", () => ({ getRelease: vi.fn().mockResolvedValue({ id: "x", sequence: 1, outcome: { resources: {} }, snapshot: { resources: [] } }) }));
import { TimelineRail } from "../timeline-rail";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;
const rels = (n: number): StackRelease[] => Array.from({ length: n }, (_, i) => ({ id: `r${n - i}`, sequence: n - i, state: "Released", cause: { kind: "manual" } } as StackRelease));

const base = { stack, onRollback: vi.fn(), onCancel: vi.fn(), onCopyId: vi.fn() };

describe("TimelineRail", () => {
  it("renders the empty state with no releases", () => {
    render(<TimelineRail releases={[]} {...base} />);
    expect(screen.getByText("No deployments yet")).toBeInTheDocument();
  });

  it("renders current node + earlier releases", () => {
    const r = rels(3);
    render(<TimelineRail releases={r} activeRelease={r[0]} {...base} />);
    expect(screen.getByText("Earlier releases")).toBeInTheDocument();
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("windows earlier releases behind Show more", async () => {
    const r = rels(20);
    render(<TimelineRail releases={r} activeRelease={r[0]} initialWindow={5} {...base} />);
    expect(screen.queryByText("#1")).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole("button", { name: /show more/i }));
    expect(screen.getByText("#1")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- timeline-rail`
Expected: FAIL — cannot find `../timeline-rail`.

- [ ] **Step 3: Write minimal implementation**

```tsx
import { useState } from "react";
import { EmptyState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { useReleaseDetail } from "../../use-release-detail";
import { RailNode } from "./rail-node";
import { CurrentReleaseNode } from "./current-release-node";
import { HistoryRow } from "./history-row";
import type { LogContext } from "./resource-row";

const TERMINAL = new Set(["Released", "Failed", "Superseded", "Cancelled"]);

export interface TimelineRailProps {
  releases: StackRelease[];
  activeRelease?: StackRelease;
  stack: Stack;
  logContext?: LogContext;
  onOpenLogs?: (name: string) => void;
  banner?: React.ReactNode;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
  onCopyId: (id: string) => void;
  initialWindow?: number;
}

export function TimelineRail(props: TimelineRailProps) {
  const { releases, activeRelease, stack, logContext, onOpenLogs, banner, onRollback, onCancel, onCopyId, initialWindow = 15 } = props;
  const detail = useReleaseDetail(logContext?.orgId ?? "", logContext?.teamName ?? "", logContext?.stackId ?? "");
  const [openId, setOpenId] = useState<string | null>(null);
  const [windowN, setWindowN] = useState(initialWindow);
  const toggle = (id: string) => setOpenId((cur) => (cur === id ? null : id));

  const earlier = releases.slice(1); // activeRelease = releases[0]
  const shown = earlier.slice(0, windowN);
  const prevIdFor = (idx: number) => releases[idx + 1]?.id;   // idx is the index in `releases`
  const prevSeqFor = (idx: number) => releases[idx + 1]?.sequence;

  return (
    <div className="space-y-0">
      {banner && <div className="mb-5">{banner}</div>}

      {activeRelease ? (
        <RailNode tone="ok" big pulse={!TERMINAL.has(activeRelease.state ?? "")} isLast={earlier.length === 0}>
          <CurrentReleaseNode
            release={activeRelease}
            stack={stack}
            logContext={logContext}
            onOpenLogs={onOpenLogs}
            detail={detail}
            prevReleaseId={prevIdFor(0)}
            prevSeq={prevSeqFor(0)}
          />
        </RailNode>
      ) : releases.length === 0 ? (
        <RailNode tone="muted" isLast>
          <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
        </RailNode>
      ) : null}

      {earlier.length > 0 && (
        <div className="ml-12 py-2 font-mono text-[11px] uppercase tracking-wide text-fg-muted">Earlier releases</div>
      )}

      {shown.map((r, i) => {
        const idx = i + 1; // position in `releases`
        return (
          <RailNode key={r.id ?? idx} tone="muted" isLast={idx === releases.length - 1}>
            <HistoryRow
              release={r}
              prevReleaseId={prevIdFor(idx)}
              prevSeq={prevSeqFor(idx)}
              detail={detail}
              isOpen={openId === r.id}
              onToggle={toggle}
              onRollback={onRollback}
              onCancel={onCancel}
              onCopyId={onCopyId}
            />
          </RailNode>
        );
      })}

      {earlier.length > windowN && (
        <div className="ml-12 pt-2">
          <button onClick={() => setWindowN((n) => n + initialWindow)} className="font-sans text-[12.5px] font-medium text-primary">
            Show more ({earlier.length - windowN})
          </button>
        </div>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `pnpm --prefix frontend test:run -- timeline-rail`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/timeline/timeline-rail.tsx frontend/src/pages/stacks/components/detail/deployments/timeline/tests/timeline-rail.test.tsx
git commit -m "feat(frontend): TimelineRail composition with show-more window"
```

---

### Task 14: Rewrite DeploymentsTab container + wire Logs switch

**Files:**
- Modify (rewrite): `frontend/src/pages/stacks/components/detail/deployments/deployments-tab.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (`Tabs` → controlled; pass `onOpenLogs`)
- Test (rewrite): `frontend/src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx`

**Interfaces:**
- Consumes: `useReleases`; `deriveFailingResources`; `TimelineRail`; `DriftBanner`/`ReleaseErrorBanner`; `createRelease`/`rollbackRelease`/`cancelRelease`; `EmptyState` (branded).
- Produces:
  ```ts
  export interface DeploymentsTabProps {
    orgId: string; teamName: string; stackId: string; stack: Stack; canDeploy: boolean;
    onOpenLogs?: (resourceName?: string) => void;
  }
  export function DeploymentsTab(props: DeploymentsTabProps): JSX.Element;
  ```

- [ ] **Step 1: Write the failing test** (replace the file)

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach, beforeAll } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
vi.mock("@/api/observability", () => ({ fetchLogSnapshot: vi.fn().mockResolvedValue([]) }));
vi.mock("@/components/ui/use-toast", () => ({ useToast: () => ({ toast: vi.fn() }) }));
vi.mock("@/api/releases", () => ({
  listReleases: vi.fn().mockResolvedValue({ items: [{ id: "r1", sequence: 14, state: "Released", cause: { kind: "manual" } }], total: 1 }),
  getRelease: vi.fn().mockResolvedValue({ id: "r1", sequence: 14, outcome: { resources: {} }, snapshot: { resources: [] } }),
  createRelease: vi.fn().mockResolvedValue({}), rollbackRelease: vi.fn(), cancelRelease: vi.fn(),
}));
import { listReleases, createRelease } from "@/api/releases";
import { DeploymentsTab } from "../deployments-tab";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
beforeAll(() => {
  const stubs: Record<string, () => unknown> = { hasPointerCapture: () => false, setPointerCapture: () => undefined, releasePointerCapture: () => undefined, scrollIntoView: () => undefined };
  for (const [k, v] of Object.entries(stubs)) (Element.prototype as unknown as Record<string, unknown>)[k] = v;
});

const stack = { status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;

describe("DeploymentsTab", () => {
  it("deploy button creates a release and refetches", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s" stack={stack} canDeploy />);
    await waitFor(() => expect(screen.getByText("#14")).toBeInTheDocument());
    const before = (listReleases as ReturnType<typeof vi.fn>).mock.calls.length;
    await userEvent.click(screen.getByRole("button", { name: /^deploy$/i }));
    expect(createRelease).toHaveBeenCalledWith("o", "t", "s");
    await waitFor(() => expect((listReleases as ReturnType<typeof vi.fn>).mock.calls.length).toBeGreaterThan(before));
  });

  it("hides the Deploy button when canDeploy is false", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s" stack={stack} canDeploy={false} />);
    await waitFor(() => expect(screen.getByText("#14")).toBeInTheDocument());
    expect(screen.queryByRole("button", { name: /^deploy$/i })).not.toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run test to verify it fails**

Run: `pnpm --prefix frontend test:run -- deployments-tab`
Expected: FAIL (new TimelineRail import / shape mismatch).

- [ ] **Step 3: Write minimal implementation** (replace `deployments-tab.tsx`)

```tsx
import { useState } from "react";
import { Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import type { Stack } from "@/api/stacks";
import { createRelease, rollbackRelease, cancelRelease } from "@/api/releases";
import { useReleases } from "./use-releases";
import { deriveFailingResources } from "./derive";
import { TimelineRail } from "./timeline/timeline-rail";
import { DriftBanner, ReleaseErrorBanner } from "./timeline/banners";

export interface DeploymentsTabProps {
  orgId: string; teamName: string; stackId: string; stack: Stack; canDeploy: boolean;
  onOpenLogs?: (resourceName?: string) => void;
}

export function DeploymentsTab({ orgId, teamName, stackId, stack, canDeploy, onOpenLogs }: DeploymentsTabProps) {
  const { releases, activeRelease, loading, error, refetch } = useReleases({ orgId, teamName, stackId, enabled: true });
  const [busy, setBusy] = useState(false);
  const { toast } = useToast();

  const run = async (fn: () => Promise<unknown>, ok: string) => {
    setBusy(true);
    try { await fn(); toast({ title: ok }); refetch(); }
    catch (e) { toast({ title: "Action failed", description: e instanceof Error ? e.message : "", variant: "destructive" }); }
    finally { setBusy(false); }
  };

  const onDeploy = () => run(() => createRelease(orgId, teamName, stackId), "Deploy started");
  const onRollback = (id: string) => run(() => rollbackRelease(orgId, teamName, stackId, id), "Rollback started");
  const onCancel = (id: string) => run(() => cancelRelease(orgId, teamName, stackId, id), "Release cancelled");
  const onCopyId = (id: string) => { void navigator.clipboard?.writeText(id); toast({ title: "Release ID copied" }); };

  const stackUpdated = (stack as unknown as { updated_at?: string }).updated_at;
  const hasDrift = !!activeRelease && !!stackUpdated && !!activeRelease.completed_at
    && new Date(stackUpdated) > new Date(activeRelease.completed_at);

  const failing = deriveFailingResources(stack);
  let banner: React.ReactNode = null;
  if (hasDrift) banner = <DriftBanner onDeploy={onDeploy} busy={busy} />;
  else if (activeRelease && failing.length > 0) {
    const total = stack.status?.resources?.length ?? failing.length;
    banner = <ReleaseErrorBanner lead={`Deploy #${activeRelease.sequence} ${activeRelease.state === "Failed" ? "failed" : "failing"}`} text={`${failing.length} of ${total} resources failing`} />;
  }

  if (error) return <EmptyState title="Could not load deployments" description={error} />;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-[15px] font-medium">Deployments</h2>
        {canDeploy && (
          <Button onClick={onDeploy} disabled={busy} className="gap-1.5">
            <Rocket className="h-3.5 w-3.5" /> Deploy
          </Button>
        )}
      </div>

      {loading && releases.length === 0 ? (
        <p className="text-[13px] text-fg-muted">Loading deployments…</p>
      ) : (
        <TimelineRail
          releases={releases}
          activeRelease={activeRelease}
          stack={stack}
          logContext={{ orgId, teamName, stackId }}
          onOpenLogs={onOpenLogs ? (name) => onOpenLogs(name) : undefined}
          banner={banner}
          onRollback={onRollback}
          onCancel={onCancel}
          onCopyId={onCopyId}
        />
      )}
    </div>
  );
}
```

- [ ] **Step 4: Wire the Logs switch in `index.tsx`**

Make `Tabs` controlled and pass `onOpenLogs`. Change line 569 `<Tabs defaultValue="configuration" className="w-full">` to controlled, adding state near the other `useState` calls (~line 85):

```tsx
const [activeTab, setActiveTab] = useState("configuration");
```
```tsx
<Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
```
And in the `<TabsContent value="deployments">` block (~line 724-736), add the prop to `<DeploymentsTab .../>`:
```tsx
onOpenLogs={() => setActiveTab("logs")}
```

- [ ] **Step 5: Run tests + typecheck**

Run: `pnpm --prefix frontend test:run -- deployments-tab`
Expected: PASS (2 tests).
Run: `pnpm --prefix frontend exec tsc -b`
Expected: only the pre-existing `postgres-backups.ts` error.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/deployments-tab.tsx frontend/src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(frontend): wire DeploymentsTab to timeline rail + Logs switch"
```

---

### Task 15: Remove superseded components + the demo harness

**Files:**
- Delete: `current-deployment-card.tsx`, `release-history.tsx`, `release-row.tsx`, `unreleased-changes-banner.tsx`, `failing-resources-accordion.tsx`, `release-detail-drawer.tsx`, `release-diff.ts` (in `deployments/`) + their `tests/*` files.
- Delete: `frontend/src/pages/__demo/` (DeploymentsDemo.tsx, deployments-fixtures.ts).
- Delete: `deploy-*.jpeg` (worktree root), `docs/superpowers/handoffs/2026-06-21-deployments-LOCAL-VERIFICATION.md`.
- Modify: `frontend/src/App.tsx` (remove the `DeploymentsDemo` import line and the `/__demo/deployments` route line).

- [ ] **Step 1: Remove the demo route from `App.tsx`**

Delete the line `import DeploymentsDemo from "@/pages/__demo/DeploymentsDemo" // DEMO-ONLY: remove with __demo folder` and the line `<Route path="/__demo/deployments" element={<DeploymentsDemo />} /> {/* DEMO-ONLY */}`.

- [ ] **Step 2: Delete superseded files**

```bash
cd frontend/src/pages/stacks/components/detail/deployments
git rm current-deployment-card.tsx release-history.tsx release-row.tsx unreleased-changes-banner.tsx failing-resources-accordion.tsx release-detail-drawer.tsx release-diff.ts
git rm tests/current-deployment-card.test.tsx tests/release-history.test.tsx tests/release-row.test.tsx tests/unreleased-changes-banner.test.tsx tests/failing-resources-accordion.test.tsx tests/release-detail-drawer.test.tsx tests/release-diff.test.ts
cd -
git rm -r frontend/src/pages/__demo
git rm deploy-01-released-light.jpeg deploy-02-recovered-light.jpeg deploy-03-build-failure-expanded-light.jpeg deploy-04-runtime-crash-multi-light.jpeg deploy-05-release-level-error-light.jpeg deploy-06-drift-light.jpeg deploy-07-history-mixed-light.jpeg deploy-08-history-mixed-dark.jpeg deploy-09-runtime-crash-multi-dark.jpeg
git rm docs/superpowers/handoffs/2026-06-21-deployments-LOCAL-VERIFICATION.md
```
(If any path was already untracked, use plain `rm` for that path.)

- [ ] **Step 3: Verify no dangling imports**

Run: `pnpm --prefix frontend exec tsc -b 2>&1 | grep -v postgres-backups`
Expected: no output (no new errors; no references to the deleted modules).
Run: `grep -rn "current-deployment-card\|release-history\|release-row\|unreleased-changes-banner\|failing-resources-accordion\|release-detail-drawer\|release-diff\|__demo" frontend/src`
Expected: no matches.

- [ ] **Step 4: Run the full suite + lint**

Run: `pnpm --prefix frontend test:run`
Expected: all PASS.
Run: `pnpm --prefix frontend lint`
Expected: zero new errors.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore(frontend): remove card/history/drawer components + demo harness superseded by timeline"
```

---

### Task 16: Final review + visual QA + backend follow-ups

**Files:** none (review + issues).

- [ ] **Step 1: Self-check the diff vs the 12 scenarios**

Run: `pnpm --prefix frontend test:run` (full suite green) and `pnpm --prefix frontend exec tsc -b` (only pre-existing error). Confirm every scenario from spec §7 maps to a passing test across Tasks 1–14.

- [ ] **Step 2: Dispatch a code-review subagent**

Use superpowers:requesting-code-review with `BASE_SHA=origin/main`, `HEAD_SHA=HEAD`. Focus: two-array correctness, lazy-fetch/cache + prev-release reuse, token/brand fidelity (no raw hex), removal completeness, the no-duplication current-node rule. Fix Critical/Important before finishing.

- [ ] **Step 3: Visual QA (manual, optional)**

With `mage run` + a stack that has releases, open the Deployments tab and eyeball the live states (released / in-progress / failing / drift / history-expand). The throwaway demo harness is intentionally gone; a backend-backed pass replaces it. Capture nothing into git.

- [ ] **Step 4: File backend follow-up issues** (`gh issue create`, repo `Stackdome/stackdome`)

  - `listReleases` pagination: add `limit` + cursor (`before_sequence`) to OpenAPI + handler + service + store; enables true infinite scroll (frontend swaps the client window for fetch-more). Reference spec §6.
  - Rollback + cancel handler/service tests (no coverage today). Reference spec §11.

- [ ] **Step 5: Commit any review fixes**

```bash
git add -A && git commit -m "fix(frontend): address timeline code-review findings"
```

---

## Self-Review

**Spec coverage:**
- §2 architecture → Tasks 4,10,11,13,14 (rail, current node, history row, composition, container). ✓
- §3 data flow (lazy getRelease + cache + prev reuse; current node = releases[0]; earlier = slice(1); changelog on current node) → Tasks 3,10,13. ✓
- §4 resource-grouped config diff (real field paths, no `replicas`) → Tasks 1,6. ✓
- §5 error/empty (list error, empty state, release-error block, post-mortem error, initial-release diff, log best-effort) → Tasks 7,8,10,13,14. ✓
- §6 pagination deferral (component-local window) → Task 13; backend issue → Task 16. ✓
- §7 twelve scenarios → covered across Tasks 1–14 tests; released/recovered/drift/pending/inprogress/build_fail/crash_single/crash_multi/release_err/pre_cluster/history/empty all expressible via `derive.ts` + the rail. ✓
- §8 token mapping (`--success`, `text-warn`, `text-danger`, etc., branded primitives) → used throughout; no raw hex. ✓
- §9 testing (incl. prev-release cache reuse) → Task 3 + per-component tests. ✓
- §10 cleanup (7 components + demo + screenshots + verification doc + App.tsx route) → Task 15. ✓
- §11 follow-ups → Task 16. ✓

**Placeholder scan:** no TBD/TODO; every code step shows complete code; `index.tsx` edit gives exact line targets + code. ✓

**Type consistency:** `ResourceDiff`/`DiffRow`/`DiffSection` (Task 1) consumed unchanged by Tasks 6,7. `ReleaseDetail.ensure/peek` (Task 3) consumed by Tasks 7,10,11,13. `LogContext` defined in Task 8, imported by Tasks 10,13. `ResourceRowVM` (Task 8) built in Task 10. `TimelineRailProps` (Task 13) consumed by Task 14. Branded prop names (`StageTracker stages`, `StatusPill variant`, `FailureCard` props, `AlertBanner action`) match existing primitives. ✓

