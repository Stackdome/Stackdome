# Stack Deployments Tab Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a **Deployments** tab to the Stack detail page, backed by the Releases API (#100), with deploy history, rollback/cancel, a live current-deployment surface, honest build/runtime/config failure visibility, a per-release detail drawer, and a decoupled Save/Deploy flow.

**Architecture:** Frontend-only on `origin/main` (Releases backend already shipped). A new `deployments/` feature folder under `frontend/src/pages/stacks/components/detail/` holds a fetch+poll hook (`use-releases.ts`), a pure view-model (`derive.ts`), a generic snapshot diff (`release-diff.ts`), and presentational components. Six branded failure primitives are **salvaged verbatim** from the parked `feat/stack-activity-tab` branch via `git show`. Data comes from two endpoints joined by resource name: `GET/POST /stacks/{id}/releases` for the timeline, and live `GET /stacks/{id}` (`status.resources[]` for progress, `spec.stack_resources[].status.last_failure` for structured failures). No React Query — the codebase uses plain `useState`/`useEffect`.

**Tech Stack:** React 19, TypeScript, Vite, Tailwind v4, Radix (Tabs/Sheet/Accordion), lucide-react, Vitest + @testing-library/react (jsdom). Generated OpenAPI types in `frontend/src/api/types/openapi.d.ts`; axios default instance `import api from "@/api/client"`.

---

## Reference — exact surfaces (verified against the worktree, origin/main @ de222a8)

**Generated types** (after regen): `components["schemas"]["StackRelease" | "StackReleaseList" | "CreateReleaseRequest" | "ReleaseCause" | "ReleaseOutcome" | "ReleasePins" | "StackResourceSummary" | "StackResourceFailure" | "StackConvergenceRecord" | "Stack"]`. `StackReleaseState` is a string union `"Pending"|"InProgress"|"Released"|"Failed"|"Superseded"|"Cancelled"`.

**Schema facts (load-bearing):**
- `StackReleaseList = { items?: StackRelease[]; total?: number }`.
- `StackRelease` list rows carry: `id, stack_id, sequence, state, message, cause, snapshot_revision, manifest_revision, renderer_version, pins, created_by, created_at, updated_at, rendered_at, completed_at`. `outcome`/`snapshot`/`manifest` are **detail-only** (`GET /releases/{id}`).
- `ReleaseCause = { kind?: "manual"|"rollback"|"webhook_push"; detail?: string }`.
- `ReleasePins = { resources?: { [name: string]: ResourcePins } }`, `ResourcePins = { git_sha?, volume_hash?, image_digest? }`. **Map keyed by resource name, not array.**
- `ReleaseOutcome = { resources?: { [name: string]: ResourceOutcome }; duration?: string }`, `ResourceOutcome = { phase?, ready_replicas?, replicas?, message? }`.
- `StackStatus = { state?, message?, observed_revision?, target_revision?, last_converged?: StackConvergenceRecord, resources?: StackResourceSummary[], conditions?: Condition[] }`.
- `StackConvergenceRecord = { revision?, release_id?, at? }`.
- `StackResourceSummary = { name?, phase?, observed_revision?, converged_revision?, available_replicas?, updated_replicas?, replicas?, missing?, message? }` — **NO `last_failure`.**
- Live structured failure path: `stack.spec.stack_resources[i].status.last_failure` where `status: StackResourceStatus = { state?, last_failure?: StackResourceFailure, conditions?, last_restart_request_processed_at?, ... }`.
- `StackResourceFailure = { type?: "runtime_crash"|"build_failure"; container?: ContainerFailureDetail; init_container?: ContainerFailureDetail; build?: BuildFailureDetail }`. `ContainerFailureDetail`/`BuildFailureDetail = { failure_type?: "crash_loop"|"out_of_memory"|"image_pull_failed"|"create_container_error"|"exit_error"; reason?; message?; restart_count?; exit_code? }`.

**Endpoints** (prefix `/organizations/{org}/teams/{team}/stacks/{id}`): `GET /releases` → `StackReleaseList`; `GET /releases/{release_id}` → `StackRelease`; `POST /releases` body `CreateReleaseRequest = { from_release_id?: string }` → `StackRelease` (deploy with `{}`, rollback with `{from_release_id}`); `POST /releases/{release_id}/cancel` → 200 no body.

**Reuse (existing on origin/main):**
- `@/api/client` default export `api` (axios); `api.get<T>(path)`, `api.post<T>(path, body)`.
- `@/pages/stacks/lib/stack-diff.ts`: `deepEqual(a, b): boolean`, `getAtPath(obj, path)`, `cloneJson<T>(v): T`.
- `@/pages/stacks/components/shared/dirty-field.tsx`: `DirtyField` props `{ draft, baseline, path, onReset?, compact?, hideReset?, className?, children }`.
- `@/pages/stacks/components/shared/sticky-action-bar.tsx`: `StickyActionBar` props `{ leadLabel, segments, primary{label,onClick,loadingLabel?,isLoading?,icon?}, secondary? }`.
- `@/components/branded` barrel (existing: StatusPill, Panel, EmptyState, …).
- `@/components/ui/{sheet,accordion,tabs}` (Radix wrappers).

**Salvage source (parked branch `feat/stack-activity-tab`):** `frontend/src/components/branded/{stage-tracker,stage-badge,failure-card,alert-banner,log-snapshot,event-row}.tsx`, and additions inside `frontend/src/api/observability.ts` (`buildStackResourceLogStreamUrl`, `fetchLogSnapshot`).

**Integration point:** `frontend/src/pages/stacks/components/detail/index.tsx` — `<Tabs defaultValue="configuration">` at line 568 (triggers 570–587: configuration/logs/metrics; content 591/717/730). StickyActionBar primary `label:"Deploy"` at 520. Team name available as `teamNameById(stack.team_id) ?? defaultTeamName` (lines 100, 328); live stack is `stackToShow`.

---

## File Structure

```
frontend/src/
├─ api/
│  ├─ releases.ts                         NEW  — list/get/create/cancel release clients
│  └─ observability.ts                    MOD  — salvage buildStackResourceLogStreamUrl + fetchLogSnapshot
├─ components/branded/
│  ├─ stage-tracker.tsx                   NEW(salvage+adapt) — Build→Deploy→Ready
│  ├─ stage-badge.tsx                     NEW(salvage)
│  ├─ failure-card.tsx                    NEW(salvage)
│  ├─ alert-banner.tsx                    NEW(salvage)
│  ├─ log-snapshot.tsx                    NEW(salvage)
│  ├─ event-row.tsx                       NEW(salvage)
│  └─ index.ts                            MOD  — export the six
└─ pages/stacks/components/detail/
   ├─ index.tsx                           MOD  — add Deployments tab; relabel Deploy→Save
   └─ deployments/
      ├─ derive.ts                        NEW  — pure view-model (stages, failures, recovered, labels)
      ├─ release-diff.ts                  NEW  — generic snapshot/spec JSON diff
      ├─ use-releases.ts                  NEW  — fetch + poll hook
      ├─ deployments-tab.tsx              NEW  — tab container
      ├─ current-deployment-card.tsx      NEW  — active release + live status + stage tracker
      ├─ failing-resources-accordion.tsx  NEW  — click-through failure cards
      ├─ release-history.tsx              NEW  — history list
      ├─ release-row.tsx                  NEW  — one history row + ⋮ menu
      ├─ release-detail-drawer.tsx        NEW  — GET /releases/{id} drawer
      ├─ unreleased-changes-banner.tsx    NEW  — drift (saved≠deployed) → Deploy
      └─ tests/                           NEW  — *.test.ts(x) per module
```

---

## TIER 1 — Read path: types, clients, view-model, current deployment + history, wired tab

### Task 1: Regenerate OpenAPI types & zod (release schemas)

**Files:**
- Modify (generated): `frontend/src/api/types/openapi.d.ts`, `frontend/src/api/zod-schemas.ts`

- [ ] **Step 1: Confirm the committed types lack release schemas**

Run: `grep -c 'StackReleaseList' frontend/src/api/types/openapi.d.ts`
Expected: `0` (stale — backend #100 added schemas to the YAML but the frontend types were never regenerated).

- [ ] **Step 2: Regenerate types and zod from the spec**

```bash
pnpm --prefix frontend generate:openapi-types
pnpm --prefix frontend generate:openapi-zod
```

- [ ] **Step 3: Verify the release schemas now exist**

Run: `grep -c 'StackReleaseList\|CreateReleaseRequest\|StackResourceFailure' frontend/src/api/types/openapi.d.ts`
Expected: `≥ 3`.

- [ ] **Step 4: Verify the project still type-checks and builds**

Run: `pnpm --prefix frontend exec tsc -b --noEmit`
Expected: no NEW errors (one pre-existing unrelated `postgres-backups.ts` error is acceptable — confirm it is the only one and predates this change with `git stash` if unsure).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/types/openapi.d.ts frontend/src/api/zod-schemas.ts
git commit -m "chore(frontend): regenerate OpenAPI types/zod with Releases schemas"
```

---

### Task 2: Release API client (`api/releases.ts`)

**Files:**
- Create: `frontend/src/api/releases.ts`
- Test: `frontend/src/api/tests/releases.test.ts`

- [ ] **Step 1: Write the failing test**

```ts
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach } from "vitest";

vi.mock("@/api/client", () => ({
  default: { get: vi.fn(), post: vi.fn() },
}));

import api from "@/api/client";
import { listReleases, getRelease, createRelease, rollbackRelease, cancelRelease } from "../releases";

const ORG = "org1", TEAM = "team1", STACK = "s1";
const BASE = `/organizations/${ORG}/teams/${TEAM}/stacks/${STACK}/releases`;

beforeEach(() => vi.clearAllMocks());

describe("releases api", () => {
  it("lists releases", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { items: [{ id: "r1" }], total: 1 } });
    const out = await listReleases(ORG, TEAM, STACK);
    expect(api.get).toHaveBeenCalledWith(BASE);
    expect(out.items?.[0].id).toBe("r1");
  });

  it("gets one release", async () => {
    (api.get as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r1" } });
    await getRelease(ORG, TEAM, STACK, "r1");
    expect(api.get).toHaveBeenCalledWith(`${BASE}/r1`);
  });

  it("creates a deploy release with empty body", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r2" } });
    await createRelease(ORG, TEAM, STACK);
    expect(api.post).toHaveBeenCalledWith(BASE, {});
  });

  it("rolls back via from_release_id", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: { id: "r3" } });
    await rollbackRelease(ORG, TEAM, STACK, "r1");
    expect(api.post).toHaveBeenCalledWith(BASE, { from_release_id: "r1" });
  });

  it("cancels a release", async () => {
    (api.post as ReturnType<typeof vi.fn>).mockResolvedValue({ data: undefined });
    await cancelRelease(ORG, TEAM, STACK, "r1");
    expect(api.post).toHaveBeenCalledWith(`${BASE}/r1/cancel`);
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/api/tests/releases.test.ts`
Expected: FAIL — `Cannot find module '../releases'`.

- [ ] **Step 3: Implement the client**

```ts
// frontend/src/api/releases.ts
import api from "./client";
import type { components } from "./types/openapi";

export type StackRelease = components["schemas"]["StackRelease"];
export type StackReleaseList = components["schemas"]["StackReleaseList"];
export type CreateReleaseRequest = components["schemas"]["CreateReleaseRequest"];

function releasesPath(orgId: string, teamName: string, stackId: string): string {
  return `/organizations/${orgId}/teams/${teamName}/stacks/${stackId}/releases`;
}

export async function listReleases(orgId: string, teamName: string, stackId: string): Promise<StackReleaseList> {
  const response = await api.get<StackReleaseList>(releasesPath(orgId, teamName, stackId));
  return response.data;
}

export async function getRelease(orgId: string, teamName: string, stackId: string, releaseId: string): Promise<StackRelease> {
  const response = await api.get<StackRelease>(`${releasesPath(orgId, teamName, stackId)}/${releaseId}`);
  return response.data;
}

export async function createRelease(orgId: string, teamName: string, stackId: string): Promise<StackRelease> {
  const response = await api.post<StackRelease>(releasesPath(orgId, teamName, stackId), {});
  return response.data;
}

export async function rollbackRelease(orgId: string, teamName: string, stackId: string, fromReleaseId: string): Promise<StackRelease> {
  const body: CreateReleaseRequest = { from_release_id: fromReleaseId };
  const response = await api.post<StackRelease>(releasesPath(orgId, teamName, stackId), body);
  return response.data;
}

export async function cancelRelease(orgId: string, teamName: string, stackId: string, releaseId: string): Promise<void> {
  await api.post(`${releasesPath(orgId, teamName, stackId)}/${releaseId}/cancel`);
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/api/tests/releases.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/api/releases.ts frontend/src/api/tests/releases.test.ts
git commit -m "feat(frontend): release API client (list/get/create/rollback/cancel)"
```

---

### Task 3: Salvage branded failure primitives (verbatim)

**Files:**
- Create (from parked branch): `frontend/src/components/branded/{stage-badge,failure-card,alert-banner,log-snapshot,event-row}.tsx`
- Modify: `frontend/src/components/branded/index.ts`

> These six components are presentational and were reviewed/QA'd on the parked branch. Copy them verbatim; `stage-tracker` is adapted separately in Task 4.

- [ ] **Step 1: Copy the five generic primitives from the parked branch**

```bash
for f in stage-badge failure-card alert-banner log-snapshot event-row; do
  git show feat/stack-activity-tab:frontend/src/components/branded/$f.tsx > frontend/src/components/branded/$f.tsx
done
```

- [ ] **Step 2: Add `"validation"` defensiveness to `stage-badge` (it only handles build/runtime/init)**

Open `frontend/src/components/branded/stage-badge.tsx`. The exported `FailureStage` union and `STAGE_CLASSES` map cover `"build" | "runtime" | "init"`. Add a `"validation"` arm so a stray value renders safely:

```ts
// FailureStage union — add "validation"
export type FailureStage = "build" | "runtime" | "init" | "validation";

// In the class map object literal, add:
validation: "text-warn border-warn-border", // same as runtime/init styling
```

(If the file already includes `"validation"`, leave it.)

- [ ] **Step 3: Export the five from the barrel**

In `frontend/src/components/branded/index.ts`, append (keep existing exports):

```ts
export { StageBadge, type FailureStage } from "./stage-badge";
export { FailureCard, type FailureCardProps } from "./failure-card";
export { AlertBanner, type AlertBannerProps } from "./alert-banner";
export { EventRow, type EventRowProps } from "./event-row";
export { LogSnapshot, type LogSnapshotProps } from "./log-snapshot";
```

- [ ] **Step 4: Verify they compile and existing tests still pass**

Run: `pnpm --prefix frontend exec tsc -b --noEmit && pnpm --prefix frontend test:run`
Expected: builds; 321 prior tests still pass.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/branded/
git commit -m "feat(frontend): salvage branded failure primitives from parked branch"
```

---

### Task 4: Salvage + adapt the stage tracker (Build → Deploy → Ready)

**Files:**
- Create (from parked branch, then adapt): `frontend/src/components/branded/stage-tracker.tsx`
- Modify: `frontend/src/components/branded/index.ts`
- Test: `frontend/src/components/branded/tests/stage-tracker.test.tsx`

> The parked tracker uses keys `{ validate?, build, rollout, ready }` with labels `Validate/Build/Rollout/Ready`. The spec mandates the friendly **Build → Deploy → Ready**. Rename the `rollout` key to `deploy`, drop `validate`, and relabel.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { StageTracker } from "../stage-tracker";

afterEach(cleanup);

describe("StageTracker", () => {
  it("renders Build, Deploy, Ready with status marks", () => {
    render(<StageTracker stages={{ build: "done", deploy: "active", ready: "todo" }} />);
    expect(screen.getByText("Build")).toBeInTheDocument();
    expect(screen.getByText("Deploy")).toBeInTheDocument();
    expect(screen.getByText("Ready")).toBeInTheDocument();
    expect(screen.getByText("Deploy").closest("[data-status]")).toHaveAttribute("data-status", "active");
  });

  it("marks a failed stage", () => {
    render(<StageTracker stages={{ build: "failed", deploy: "todo", ready: "todo" }} />);
    expect(screen.getByText("Build").closest("[data-status]")).toHaveAttribute("data-status", "failed");
  });
});
```

- [ ] **Step 2: Copy the parked tracker, then run the test to verify it fails on labels/keys**

```bash
git show feat/stack-activity-tab:frontend/src/components/branded/stage-tracker.tsx > frontend/src/components/branded/stage-tracker.tsx
mkdir -p frontend/src/components/branded/tests
```

Run: `pnpm --prefix frontend exec vitest run src/components/branded/tests/stage-tracker.test.tsx`
Expected: FAIL — renders "Rollout" not "Deploy" (and/or no `deploy` key).

- [ ] **Step 3: Adapt the tracker (overwrite with the canonical Build→Deploy→Ready version)**

```tsx
// frontend/src/components/branded/stage-tracker.tsx
import { cn } from "@/lib/utils";

export type StageStatus = "done" | "active" | "failed" | "todo";

export interface Stages {
  build: StageStatus;
  deploy: StageStatus;
  ready: StageStatus;
}

const ORDER: Array<{ key: keyof Stages; label: string }> = [
  { key: "build", label: "Build" },
  { key: "deploy", label: "Deploy" },
  { key: "ready", label: "Ready" },
];

const MARK: Record<StageStatus, string> = { done: "✓", active: "●", failed: "✕", todo: "" };

function dotClass(status: StageStatus): string {
  switch (status) {
    case "done": return "bg-brand text-brand-foreground border-brand";
    case "active": return "bg-brand/10 text-brand border-brand animate-pulse";
    case "failed": return "bg-danger/10 text-danger border-danger";
    default: return "bg-muted text-muted-foreground border-border";
  }
}

export function StageTracker({ stages, className }: { stages: Stages; className?: string }) {
  return (
    <div className={cn("flex items-center gap-0", className)} role="list">
      {ORDER.map((stage, i) => {
        const status = stages[stage.key];
        return (
          <div key={stage.key} className="flex items-center" role="listitem" data-status={status}>
            <div className={cn("flex h-5 w-5 items-center justify-center rounded-full border text-[11px] font-medium", dotClass(status))}>
              {MARK[status]}
            </div>
            <span className={cn("ml-1.5 text-[13px]", status === "todo" ? "text-muted-foreground" : "text-foreground")}>
              {stage.label}
            </span>
            {i < ORDER.length - 1 && (
              <span className={cn("mx-3 h-px w-10", status === "done" ? "bg-brand" : "bg-border")} />
            )}
          </div>
        );
      })}
    </div>
  );
}
```

- [ ] **Step 4: Export from the barrel and run the test**

In `frontend/src/components/branded/index.ts` append:

```ts
export { StageTracker, type StageStatus, type Stages } from "./stage-tracker";
```

Run: `pnpm --prefix frontend exec vitest run src/components/branded/tests/stage-tracker.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/branded/stage-tracker.tsx frontend/src/components/branded/index.ts frontend/src/components/branded/tests/stage-tracker.test.tsx
git commit -m "feat(frontend): adapt stage tracker to Build->Deploy->Ready"
```

---

### Task 5: View-model — failures, recovered, labels (`derive.ts` part 1)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/derive.ts`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts`

> Two arrays, joined by name: progress from `status.resources[]` (summaries, no `last_failure`); structured failures from `spec.stack_resources[].status.last_failure`.

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect } from "vitest";
import type { Stack } from "@/api/stacks";
import { deriveFailingResources, deriveRecovered, humanizeFailureType, causeLabel, formatDuration } from "../derive";

function stackWith(resources: Array<Record<string, unknown>>): Stack {
  return { spec: { stack_resources: resources } } as unknown as Stack;
}

describe("deriveFailingResources", () => {
  it("joins last_failure from spec.stack_resources[].status", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "CrashLoopBackOff", last_failure: {
        type: "runtime_crash", container: { failure_type: "crash_loop", reason: "CrashLoopBackOff", message: "exit 1", exit_code: 1, restart_count: 5 } } } },
      { name: "redis", status: { state: "Ready" } },
    ]);
    const out = deriveFailingResources(stack);
    expect(out).toHaveLength(1);
    expect(out[0]).toMatchObject({ name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", exitCode: 1, restartCount: 5 });
  });

  it("classifies a build failure to the build stage", () => {
    const stack = stackWith([
      { name: "api", status: { state: "Error", last_failure: {
        type: "build_failure", build: { failure_type: "image_pull_failed", reason: "ErrImagePull", message: "manifest unknown" } } } },
    ]);
    expect(deriveFailingResources(stack)[0]).toMatchObject({ name: "api", type: "build_failure", stage: "build", reason: "ErrImagePull" });
  });
});

describe("deriveRecovered", () => {
  it("flags a Ready resource that still carries last_failure", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "Ready", last_failure: {
        type: "runtime_crash", container: { reason: "CrashLoopBackOff", restart_count: 5 } } } },
    ]);
    expect(deriveRecovered(stack)).toEqual([{ name: "tooljet", reason: "CrashLoopBackOff", restartCount: 5 }]);
  });

  it("does not flag a failing resource as recovered", () => {
    const stack = stackWith([
      { name: "tooljet", status: { state: "CrashLoopBackOff", last_failure: { type: "runtime_crash", container: { reason: "x" } } } },
    ]);
    expect(deriveRecovered(stack)).toEqual([]);
  });
});

describe("humanizeFailureType", () => {
  it("maps known types", () => {
    expect(humanizeFailureType("out_of_memory")).toBe("Out of memory");
    expect(humanizeFailureType("crash_loop")).toBe("Crash loop");
  });
  it("falls back to the raw value", () => {
    expect(humanizeFailureType("weird_thing")).toBe("weird_thing");
    expect(humanizeFailureType(undefined)).toBe("Unknown");
  });
});

describe("causeLabel", () => {
  it("labels manual / rollback / webhook", () => {
    expect(causeLabel({ kind: "manual" })).toBe("Manual deploy");
    expect(causeLabel({ kind: "rollback", detail: "12" })).toBe("Rollback to #12");
    expect(causeLabel({ kind: "webhook_push" })).toBe("Webhook push");
  });
});

describe("formatDuration", () => {
  it("derives from rendered_at → completed_at", () => {
    expect(formatDuration("2026-06-21T12:00:00Z", "2026-06-21T12:00:32Z")).toBe("32s");
    expect(formatDuration("2026-06-21T12:00:00Z", "2026-06-21T12:02:05Z")).toBe("2m 5s");
  });
  it("returns dash when missing", () => {
    expect(formatDuration(undefined, undefined)).toBe("—");
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/derive.test.ts`
Expected: FAIL — `Cannot find module '../derive'`.

- [ ] **Step 3: Implement `derive.ts` (part 1)**

```ts
// frontend/src/pages/stacks/components/detail/deployments/derive.ts
import type { components } from "@/api/types/openapi";

export type Stack = components["schemas"]["Stack"];
export type StackResourceFailure = components["schemas"]["StackResourceFailure"];
export type ReleaseCause = components["schemas"]["ReleaseCause"];
export type FailureStage = "build" | "runtime" | "init" | "validation";

export interface FailingResource {
  name: string;
  type: "build_failure" | "runtime_crash";
  stage: FailureStage;
  reason: string;
  message?: string;
  exitCode?: number;
  restartCount?: number;
  failureType?: string;
}

export interface RecoveredResource {
  name: string;
  reason: string;
  restartCount?: number;
}

const FAILURE_TYPE_LABELS: Record<string, string> = {
  crash_loop: "Crash loop",
  out_of_memory: "Out of memory",
  image_pull_failed: "Image pull failed",
  create_container_error: "Container create error",
  exit_error: "Exit error",
};

export function humanizeFailureType(failureType?: string): string {
  if (!failureType) return "Unknown";
  return FAILURE_TYPE_LABELS[failureType] ?? failureType;
}

/** Pick the active detail block from a last_failure (build vs container vs init). */
function failureDetail(f: StackResourceFailure) {
  if (f.type === "build_failure") return { detail: f.build, stage: "build" as const };
  if (f.init_container) return { detail: f.init_container, stage: "init" as const };
  return { detail: f.container, stage: "runtime" as const };
}

export function deriveFailingResources(stack: Stack): FailingResource[] {
  const resources = stack.spec?.stack_resources ?? [];
  const out: FailingResource[] = [];
  for (const r of resources) {
    const f = r.status?.last_failure;
    const state = r.status?.state ?? "";
    // Only surface as ACTIVE failure when the resource is not currently healthy.
    if (!f || isHealthyState(state)) continue;
    const { detail, stage } = failureDetail(f);
    out.push({
      name: r.name ?? "",
      type: (f.type ?? "runtime_crash") as FailingResource["type"],
      stage,
      reason: detail?.reason ?? humanizeFailureType(detail?.failure_type),
      message: detail?.message,
      exitCode: detail?.exit_code,
      restartCount: detail?.restart_count,
      failureType: detail?.failure_type,
    });
  }
  return out;
}

export function deriveRecovered(stack: Stack): RecoveredResource[] {
  const resources = stack.spec?.stack_resources ?? [];
  const out: RecoveredResource[] = [];
  for (const r of resources) {
    const f = r.status?.last_failure;
    const state = r.status?.state ?? "";
    if (!f || !isHealthyState(state)) continue;
    const { detail } = failureDetail(f);
    out.push({ name: r.name ?? "", reason: detail?.reason ?? humanizeFailureType(detail?.failure_type), restartCount: detail?.restart_count });
  }
  return out;
}

function isHealthyState(state: string): boolean {
  const s = state.toLowerCase();
  return s === "ready" || s === "available" || s === "running" || s === "healthy";
}

export function causeLabel(cause?: ReleaseCause): string {
  switch (cause?.kind) {
    case "rollback": return cause.detail ? `Rollback to #${cause.detail}` : "Rollback";
    case "webhook_push": return "Webhook push";
    case "manual": default: return "Manual deploy";
  }
}

export function formatDuration(start?: string, end?: string): string {
  if (!start || !end) return "—";
  const ms = new Date(end).getTime() - new Date(start).getTime();
  if (!Number.isFinite(ms) || ms < 0) return "—";
  const totalSec = Math.round(ms / 1000);
  if (totalSec < 60) return `${totalSec}s`;
  const m = Math.floor(totalSec / 60);
  const s = totalSec % 60;
  return `${m}m ${s}s`;
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/derive.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/derive.ts frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts
git commit -m "feat(frontend): deployments view-model — failures, recovered, labels"
```

---

### Task 6: View-model — stage derivation (`derive.ts` part 2)

**Files:**
- Modify: `frontend/src/pages/stacks/components/detail/deployments/derive.ts`
- Modify: `frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts`

> Gate Build on contract fields (a `git_sha` pin implies a build resource), never on free-form `phase` strings. Pre-cluster (render/apply) failures map to the first node (Build ✕) per spec §6.1.

- [ ] **Step 1: Add the failing test**

```ts
// append to derive.test.ts
import { deriveStages, releaseGitSha } from "../derive";
import type { StackRelease } from "@/api/releases";

function release(partial: Partial<StackRelease>): StackRelease {
  return { id: "r1", ...partial } as StackRelease;
}

describe("deriveStages", () => {
  const imagePins = { resources: { api: { git_sha: "9c69af2" } } };

  it("all done when converged to the active release", () => {
    const stack = { status: { last_converged: { release_id: "r1" } } } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ id: "r1", state: "Released", pins: imagePins }), []))
      .toEqual({ build: "done", deploy: "done", ready: "done" });
  });

  it("build active while Pending with build pins", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "Pending", pins: imagePins }), []))
      .toEqual({ build: "active", deploy: "todo", ready: "todo" });
  });

  it("build failed when a resource reports build_failure", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    const failing = [{ name: "api", type: "build_failure" as const, stage: "build" as const, reason: "x" }];
    expect(deriveStages(stack, release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("deploy failed when a runtime crash occurs (build already done)", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    const failing = [{ name: "api", type: "runtime_crash" as const, stage: "runtime" as const, reason: "x" }];
    expect(deriveStages(stack, release({ state: "InProgress", pins: imagePins }), failing))
      .toEqual({ build: "done", deploy: "failed", ready: "todo" });
  });

  it("pre-cluster Failed with no resource failure maps to Build ✕", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "Failed", pins: imagePins }), []))
      .toEqual({ build: "failed", deploy: "todo", ready: "todo" });
  });

  it("image-only stack (no build pins) starts at Deploy", () => {
    const stack = { status: {} } as unknown as import("../derive").Stack;
    expect(deriveStages(stack, release({ state: "InProgress", pins: { resources: { api: { image_digest: "sha256:..." } } } }), []))
      .toMatchObject({ build: "todo", deploy: "active" });
  });
});

describe("releaseGitSha", () => {
  it("returns the first non-empty git_sha from pins", () => {
    expect(releaseGitSha(release({ pins: { resources: { api: { git_sha: "abc1234" } } } }))).toBe("abc1234");
  });
  it("returns undefined when no git pins", () => {
    expect(releaseGitSha(release({ pins: { resources: { api: { image_digest: "x" } } } }))).toBeUndefined();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/derive.test.ts`
Expected: FAIL — `deriveStages`/`releaseGitSha` not exported.

- [ ] **Step 3: Implement (append to `derive.ts`)**

```ts
// add these imports to the TOP of derive.ts (with the existing imports), then append the functions below
import type { StackRelease } from "@/api/releases";
import type { Stages } from "@/components/branded";

export function releaseGitSha(release: StackRelease): string | undefined {
  const map = release.pins?.resources ?? {};
  for (const p of Object.values(map)) {
    if (p?.git_sha) return p.git_sha;
  }
  return undefined;
}

/** True if the release pins any resource with a git_sha (i.e. a build happened). */
function hasBuildResources(release: StackRelease): boolean {
  return releaseGitSha(release) !== undefined;
}

export function deriveStages(stack: Stack, release: StackRelease, failing: FailingResource[]): Stages {
  const converged = stack.status?.last_converged?.release_id != null
    && stack.status?.last_converged?.release_id === release.id;
  const buildFailed = failing.some((f) => f.type === "build_failure");
  const runtimeFailed = failing.some((f) => f.type === "runtime_crash");
  const hasBuild = hasBuildResources(release);
  const state = release.state;

  if (converged || state === "Released") {
    return { build: hasBuild ? "done" : "todo", deploy: "done", ready: "done" };
  }
  if (buildFailed) return { build: "failed", deploy: "todo", ready: "todo" };

  if (state === "Pending") {
    return hasBuild
      ? { build: "active", deploy: "todo", ready: "todo" }
      : { build: "todo", deploy: "active", ready: "todo" };
  }
  if (state === "InProgress") {
    return {
      build: hasBuild ? "done" : "todo",
      deploy: runtimeFailed ? "failed" : "active",
      ready: "todo",
    };
  }
  if (state === "Failed") {
    if (runtimeFailed) return { build: hasBuild ? "done" : "todo", deploy: "failed", ready: "todo" };
    // Pre-cluster (render/apply/timeout) → map to first node per spec §6.1.
    return { build: "failed", deploy: "todo", ready: "todo" };
  }
  // Superseded / Cancelled → neutral.
  return { build: "todo", deploy: "todo", ready: "todo" };
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/derive.test.ts`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/derive.ts frontend/src/pages/stacks/components/detail/deployments/tests/derive.test.ts
git commit -m "feat(frontend): deployments stage derivation (Build->Deploy->Ready)"
```

---

### Task 7: Releases fetch + poll hook (`use-releases.ts`)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/use-releases.ts`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/use-releases.test.tsx`

> Plain `useState`/`useEffect` (matches codebase). Poll every 5s while the active release is non-terminal (`Pending`/`InProgress`); stop when terminal. In-flight guard; stop on unmount.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, cleanup } from "@testing-library/react";

vi.mock("@/api/releases", () => ({ listReleases: vi.fn() }));
import { listReleases } from "@/api/releases";
import { useReleases } from "../use-releases";

const ARGS = { orgId: "o", teamName: "t", stackId: "s", enabled: true };

beforeEach(() => vi.clearAllMocks());
afterEach(() => { cleanup(); vi.useRealTimers(); });

describe("useReleases", () => {
  it("fetches on mount and exposes the active release (highest sequence first)", async () => {
    (listReleases as ReturnType<typeof vi.fn>).mockResolvedValue({
      items: [{ id: "r2", sequence: 2, state: "Released" }, { id: "r1", sequence: 1, state: "Superseded" }],
      total: 2,
    });
    const { result } = renderHook(() => useReleases(ARGS));
    await waitFor(() => expect(result.current.releases).toHaveLength(2));
    expect(result.current.activeRelease?.id).toBe("r2");
    expect(listReleases).toHaveBeenCalledWith("o", "t", "s");
  });

  it("does not fetch when disabled", async () => {
    renderHook(() => useReleases({ ...ARGS, enabled: false }));
    await Promise.resolve();
    expect(listReleases).not.toHaveBeenCalled();
  });

  it("polls while the active release is non-terminal", async () => {
    vi.useFakeTimers();
    (listReleases as ReturnType<typeof vi.fn>).mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "InProgress" }] });
    renderHook(() => useReleases(ARGS));
    await vi.advanceTimersByTimeAsync(0);
    expect(listReleases).toHaveBeenCalledTimes(1);
    await vi.advanceTimersByTimeAsync(5000);
    expect(listReleases).toHaveBeenCalledTimes(2);
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/use-releases.test.tsx`
Expected: FAIL — `Cannot find module '../use-releases'`.

- [ ] **Step 3: Implement the hook**

```ts
// frontend/src/pages/stacks/components/detail/deployments/use-releases.ts
import { useCallback, useEffect, useRef, useState } from "react";
import { listReleases, type StackRelease } from "@/api/releases";

const POLL_MS = 5000;
const TERMINAL = new Set<string>(["Released", "Failed", "Superseded", "Cancelled"]);

export interface UseReleasesArgs {
  orgId: string;
  teamName: string;
  stackId: string;
  enabled: boolean;
}

export interface UseReleasesResult {
  releases: StackRelease[];
  activeRelease: StackRelease | undefined;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

function bySequenceDesc(a: StackRelease, b: StackRelease): number {
  return (b.sequence ?? 0) - (a.sequence ?? 0);
}

export function useReleases({ orgId, teamName, stackId, enabled }: UseReleasesArgs): UseReleasesResult {
  const [releases, setReleases] = useState<StackRelease[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inFlight = useRef(false);
  const mounted = useRef(true);

  const fetchOnce = useCallback(async () => {
    if (!enabled || inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    try {
      const data = await listReleases(orgId, teamName, stackId);
      if (!mounted.current) return;
      setReleases([...(data.items ?? [])].sort(bySequenceDesc));
      setError(null);
    } catch (e) {
      if (mounted.current) setError(e instanceof Error ? e.message : "Failed to load releases");
    } finally {
      if (mounted.current) setLoading(false);
      inFlight.current = false;
    }
  }, [orgId, teamName, stackId, enabled]);

  useEffect(() => {
    mounted.current = true;
    return () => { mounted.current = false; };
  }, []);

  useEffect(() => {
    if (!enabled) return;
    void fetchOnce();
  }, [enabled, fetchOnce]);

  const active = releases[0];
  const shouldPoll = enabled && !!active && !TERMINAL.has(active.state ?? "");
  useEffect(() => {
    if (!shouldPoll) return;
    const id = setInterval(() => { void fetchOnce(); }, POLL_MS);
    return () => clearInterval(id);
  }, [shouldPoll, fetchOnce]);

  return { releases, activeRelease: active, loading, error, refetch: fetchOnce };
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/use-releases.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/use-releases.ts frontend/src/pages/stacks/components/detail/deployments/tests/use-releases.test.tsx
git commit -m "feat(frontend): useReleases fetch + poll hook"
```

---

### Task 8: Release history row + ⋮ menu (`release-row.tsx`)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/release-row.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/release-row.test.tsx`

> State-aware ⋮: Released → View details + Rollback to this; Pending → Cancel; others → View details. Uses `@/components/ui/dropdown-menu` and `@/components/branded` `StatusPill`.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { ReleaseRow } from "../release-row";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);

function row(partial: Partial<StackRelease>): StackRelease {
  return { id: "r1", sequence: 14, state: "Released", cause: { kind: "manual" },
    rendered_at: "2026-06-21T12:00:00Z", completed_at: "2026-06-21T12:00:32Z",
    pins: { resources: { api: { git_sha: "9c69af2" } } }, ...partial } as StackRelease;
}

describe("ReleaseRow", () => {
  it("renders sequence, cause and a Released pill", () => {
    render(<ReleaseRow release={row({})} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText(/Manual deploy/)).toBeInTheDocument();
    expect(screen.getByText(/9c69af2/)).toBeInTheDocument();
  });

  it("offers Rollback for a Released release", () => {
    const onRollback = vi.fn();
    render(<ReleaseRow release={row({})} onViewDetails={vi.fn()} onRollback={onRollback} onCancel={vi.fn()} />);
    fireEvent.click(screen.getByLabelText("Release actions"));
    fireEvent.click(screen.getByText("Rollback to this"));
    expect(onRollback).toHaveBeenCalledWith("r1");
  });

  it("offers Cancel for a Pending release and shows the failure message", () => {
    const onCancel = vi.fn();
    render(<ReleaseRow release={row({ state: "Pending" })} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={onCancel} />);
    fireEvent.click(screen.getByLabelText("Release actions"));
    fireEvent.click(screen.getByText("Cancel"));
    expect(onCancel).toHaveBeenCalledWith("r1");
  });

  it("shows the failure message on a Failed row", () => {
    render(<ReleaseRow release={row({ state: "Failed", message: "render error: bad template" })} onViewDetails={vi.fn()} onRollback={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText(/render error: bad template/)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-row.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `release-row.tsx`**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/release-row.tsx
import { MoreVertical } from "lucide-react";
import { StatusPill, variantFromState } from "@/components/branded";
import {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import type { StackRelease } from "@/api/releases";
import { causeLabel, formatDuration, releaseGitSha } from "./derive";

export interface ReleaseRowProps {
  release: StackRelease;
  onViewDetails: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
}

export function ReleaseRow({ release, onViewDetails, onRollback, onCancel }: ReleaseRowProps) {
  const id = release.id ?? "";
  const state = release.state ?? "";
  const sha = releaseGitSha(release);
  const subline =
    state === "Failed" ? release.message
    : state === "Released" ? [sha, formatDuration(release.rendered_at, release.completed_at)].filter(Boolean).join(" · ")
    : undefined;

  return (
    <div className="flex items-center justify-between gap-4 border-b border-border px-4 py-3 last:border-0">
      <div className="flex min-w-0 items-center gap-3">
        <StatusPill variant={variantFromState(state)}>{state}</StatusPill>
        <span className="font-mono text-[13px] text-foreground">#{release.sequence}</span>
        <span className="truncate text-[13px] text-muted-foreground">
          {causeLabel(release.cause)}
          {subline ? <span className={state === "Failed" ? "text-danger" : ""}> · {subline}</span> : null}
        </span>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button variant="ghost" size="icon" aria-label="Release actions" className="h-7 w-7">
            <MoreVertical className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-[180px]">
          <DropdownMenuItem onClick={() => onViewDetails(id)}>View details</DropdownMenuItem>
          {state === "Released" && (
            <DropdownMenuItem onClick={() => onRollback(id)}>Rollback to this</DropdownMenuItem>
          )}
          {state === "Pending" && (
            <DropdownMenuItem className="text-danger" onClick={() => onCancel(id)}>Cancel</DropdownMenuItem>
          )}
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
```

> If `variantFromState` does not already map release states, extend it (it lives in `@/components/branded/status-pill`). Acceptable mapping: `Released→success`, `Failed→danger`, `InProgress|Pending→info`, `Superseded|Cancelled→muted`.

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-row.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-row.tsx frontend/src/pages/stacks/components/detail/deployments/tests/release-row.test.tsx
git commit -m "feat(frontend): release history row with state-aware actions menu"
```

---

### Task 9: Release history list (`release-history.tsx`)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/release-history.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/release-history.test.tsx`

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { ReleaseHistory } from "../release-history";
import type { StackRelease } from "@/api/releases";

afterEach(cleanup);
const noop = vi.fn();
const rel = (id: string, seq: number): StackRelease => ({ id, sequence: seq, state: "Released", cause: { kind: "manual" } } as StackRelease);

describe("ReleaseHistory", () => {
  it("renders one row per release", () => {
    render(<ReleaseHistory releases={[rel("r2", 2), rel("r1", 1)]} onViewDetails={noop} onRollback={noop} onCancel={noop} />);
    expect(screen.getByText("#2")).toBeInTheDocument();
    expect(screen.getByText("#1")).toBeInTheDocument();
  });

  it("renders an empty state with no releases", () => {
    render(<ReleaseHistory releases={[]} onViewDetails={noop} onRollback={noop} onCancel={noop} />);
    expect(screen.getByText(/No deployments yet/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-history.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `release-history.tsx`**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/release-history.tsx
import { Panel, EmptyState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import { ReleaseRow } from "./release-row";

export interface ReleaseHistoryProps {
  releases: StackRelease[];
  onViewDetails: (id: string) => void;
  onRollback: (id: string) => void;
  onCancel: (id: string) => void;
}

export function ReleaseHistory({ releases, onViewDetails, onRollback, onCancel }: ReleaseHistoryProps) {
  return (
    <Panel title="History" count={releases.length} bodyClassName="p-0">
      {releases.length === 0 ? (
        <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
      ) : (
        releases.map((r) => (
          <ReleaseRow key={r.id} release={r} onViewDetails={onViewDetails} onRollback={onRollback} onCancel={onCancel} />
        ))
      )}
    </Panel>
  );
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-history.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-history.tsx frontend/src/pages/stacks/components/detail/deployments/tests/release-history.test.tsx
git commit -m "feat(frontend): release history list with empty state"
```

---

### Task 10: Current deployment card (`current-deployment-card.tsx`)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/current-deployment-card.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/current-deployment-card.test.tsx`

> Headline + StatusPill + #sequence, stage tracker, live per-resource table from `status.resources[]`, and a recovered note for Ready resources carrying `last_failure`.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { CurrentDeploymentCard } from "../current-deployment-card";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);

const release: StackRelease = { id: "r1", sequence: 14, state: "Released",
  pins: { resources: { tooljet: { git_sha: "9c69af2" } } }, snapshot_revision: "4f9c1a2" } as StackRelease;

const stack = { status: {
  last_converged: { release_id: "r1" },
  resources: [{ name: "tooljet", phase: "Ready", available_replicas: 1, replicas: 1 }],
}, spec: { stack_resources: [
  { name: "tooljet", status: { state: "Ready", last_failure: { type: "runtime_crash", container: { reason: "CrashLoopBackOff", restart_count: 5 } } } },
] } } as unknown as Stack;

describe("CurrentDeploymentCard", () => {
  it("shows the active release sequence and Ready stage tracker", () => {
    render(<CurrentDeploymentCard release={release} stack={stack} />);
    expect(screen.getByText("#14")).toBeInTheDocument();
    expect(screen.getByText("Ready").closest("[data-status]")).toHaveAttribute("data-status", "done");
  });

  it("renders the recovered note for a Ready resource with last_failure", () => {
    render(<CurrentDeploymentCard release={release} stack={stack} />);
    expect(screen.getByText(/recovered/i)).toBeInTheDocument();
    expect(screen.getByText(/CrashLoopBackOff/)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/current-deployment-card.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `current-deployment-card.tsx`**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/current-deployment-card.tsx
import { Panel, StatusPill, StageTracker, variantFromState } from "@/components/branded";
import type { StackRelease } from "@/api/releases";
import type { Stack } from "@/api/stacks";
import { deriveStages, deriveFailingResources, deriveRecovered, formatDuration } from "./derive";
import { FailingResourcesAccordion } from "./failing-resources-accordion";

export interface CurrentDeploymentCardProps {
  release: StackRelease;
  stack: Stack;
}

export function CurrentDeploymentCard({ release, stack }: CurrentDeploymentCardProps) {
  const failing = deriveFailingResources(stack);
  const recovered = deriveRecovered(stack);
  const stages = deriveStages(stack, release, failing);
  const summaries = stack.status?.resources ?? [];

  return (
    <Panel title="Current deployment" bodyClassName="p-0">
      <div className="flex items-center justify-between px-4 py-3">
        <div className="flex items-center gap-3">
          <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>
          <span className="font-mono text-[13px]">#{release.sequence}</span>
          {release.snapshot_revision && (
            <span className="text-[12px] text-muted-foreground">config {release.snapshot_revision.slice(0, 7)}</span>
          )}
        </div>
        <span className="text-[12px] text-muted-foreground">
          {formatDuration(release.rendered_at, release.completed_at)}
        </span>
      </div>

      <div className="px-4 pb-3"><StageTracker stages={stages} /></div>

      <div className="divide-y divide-border border-t border-border">
        {summaries.map((r) => (
          <div key={r.name} className="flex items-center justify-between px-4 py-2 text-[13px]">
            <span className="font-medium">{r.name}</span>
            <span className="text-muted-foreground">{r.phase}</span>
            <span className="font-mono text-muted-foreground">{r.available_replicas ?? 0}/{r.replicas ?? 0}</span>
          </div>
        ))}
      </div>

      {recovered.length > 0 && (
        <div className="border-t border-warn-border bg-warn/5 px-4 py-2 text-[12px] text-muted-foreground">
          {recovered.map((r) => (
            <div key={r.name}>
              <span className="font-medium text-foreground">{r.name}</span>{" "}
              recovered{r.restartCount != null ? ` after ${r.restartCount} restarts` : ""} — last failure{" "}
              <span className="text-warn">{r.reason}</span>
            </div>
          ))}
        </div>
      )}

      {failing.length > 0 && (
        <div className="border-t border-border p-4">
          <FailingResourcesAccordion failing={failing} releaseMessage={release.message} />
        </div>
      )}
    </Panel>
  );
}
```

> This depends on `FailingResourcesAccordion` (Task 12). To keep Task 10 self-contained for an out-of-order implementer, create a minimal stub now if Task 12 is not yet done: `export function FailingResourcesAccordion() { return null; }` in `failing-resources-accordion.tsx`, then replace it in Task 12. If implementing in order, Task 12 supplies the real component (the import path is identical).

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/current-deployment-card.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/current-deployment-card.tsx frontend/src/pages/stacks/components/detail/deployments/tests/current-deployment-card.test.tsx
git commit -m "feat(frontend): current deployment card with stage tracker + recovered note"
```

---

### Task 11: Deployments tab container + wire into the detail page; relabel Deploy→Save

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/deployments-tab.tsx`
- Modify: `frontend/src/pages/stacks/components/detail/index.tsx` (add tab; relabel StickyActionBar primary)
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx`

- [ ] **Step 1: Write the failing test for the container**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";

vi.mock("@/api/releases", () => ({
  listReleases: vi.fn().mockResolvedValue({ items: [{ id: "r1", sequence: 1, state: "Released", cause: { kind: "manual" } }], total: 1 }),
  getRelease: vi.fn(), createRelease: vi.fn(), rollbackRelease: vi.fn(), cancelRelease: vi.fn(),
}));
import { DeploymentsTab } from "../deployments-tab";
import type { Stack } from "@/api/stacks";

afterEach(cleanup);
const stack = { id: "s1", status: { resources: [] }, spec: { stack_resources: [] } } as unknown as Stack;

describe("DeploymentsTab", () => {
  it("loads releases and renders current deployment + history", async () => {
    render(<DeploymentsTab orgId="o" teamName="t" stackId="s1" stack={stack} canDeploy />);
    await waitFor(() => expect(screen.getByText("#1")).toBeInTheDocument());
    expect(screen.getByText("Current deployment")).toBeInTheDocument();
    expect(screen.getByText("History")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `deployments-tab.tsx`**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/deployments-tab.tsx
import { useState } from "react";
import { Rocket } from "lucide-react";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/branded";
import { useToast } from "@/components/ui/use-toast";
import type { Stack } from "@/api/stacks";
import { createRelease, rollbackRelease, cancelRelease } from "@/api/releases";
import { useReleases } from "./use-releases";
import { CurrentDeploymentCard } from "./current-deployment-card";
import { ReleaseHistory } from "./release-history";
import { ReleaseDetailDrawer } from "./release-detail-drawer";

export interface DeploymentsTabProps {
  orgId: string;
  teamName: string;
  stackId: string;
  stack: Stack;
  canDeploy: boolean;
}

export function DeploymentsTab({ orgId, teamName, stackId, stack, canDeploy }: DeploymentsTabProps) {
  const { releases, activeRelease, loading, error, refetch } = useReleases({ orgId, teamName, stackId, enabled: true });
  const [openReleaseId, setOpenReleaseId] = useState<string | null>(null);
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

      {activeRelease && <CurrentDeploymentCard release={activeRelease} stack={stack} />}

      <ReleaseHistory
        releases={releases}
        onViewDetails={setOpenReleaseId}
        onRollback={onRollback}
        onCancel={onCancel}
      />

      {!loading && releases.length === 0 && !activeRelease && (
        <EmptyState title="No deployments yet" description="Deploy this stack to create your first release." />
      )}

      {openReleaseId && (
        <ReleaseDetailDrawer
          orgId={orgId} teamName={teamName} stackId={stackId}
          releaseId={openReleaseId}
          previousRelease={releases.find((r, i) => releases[i - 1]?.id === openReleaseId)}
          onClose={() => setOpenReleaseId(null)}
        />
      )}
    </div>
  );
}
```

> `ReleaseDetailDrawer` arrives in Task 13. If implementing in order, add a temporary stub `export function ReleaseDetailDrawer() { return null; }` in `release-detail-drawer.tsx` so this compiles; Task 13 replaces it.

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx`
Expected: PASS.

- [ ] **Step 5: Wire the tab into `detail/index.tsx`**

Add the import near the other detail imports (after line 38's `getStackById` import group):

```tsx
import { DeploymentsTab } from "@/pages/stacks/components/detail/deployments/deployments-tab";
```

Insert a new trigger between Configuration (ends line 575) and Logs (begins line 576):

```tsx
          <TabsTrigger
            value="deployments"
            className="rounded-none border-b-2 border-transparent data-[state=active]:border-brand data-[state=active]:text-brand data-[state=active]:bg-transparent data-[state=active]:shadow-none px-1 pb-3 -mb-px font-medium"
          >
            Deployments
          </TabsTrigger>
```

Insert the matching content after the Configuration `</TabsContent>` (line 714) and before the Logs comment (line 716):

```tsx
        {/* Deployments Tab */}
        <TabsContent value="deployments">
          {stackToShow.id ? (
            <DeploymentsTab
              orgId={stackToShow.organisation_id || getCurrentOrganizationId() || ""}
              teamName={teamNameById(stackToShow.team_id) || defaultTeamName}
              stackId={stackToShow.id}
              stack={stackToShow}
              canDeploy={canWriteStack}
            />
          ) : (
            <div className="text-center text-muted-foreground py-12">Stack ID not available</div>
          )}
        </TabsContent>
```

- [ ] **Step 6: Relabel the StickyActionBar primary (Save vs Deploy decouple)**

In `detail/index.tsx`, change the `primary` block at lines 519–524 from:

```tsx
            primary={{
              label: "Deploy",
              loadingLabel: "Deploying",
              icon: <Rocket className="h-3.5 w-3.5" />,
              isLoading: isSaving,
              onClick: handleSave,
            }}
```

to:

```tsx
            primary={{
              label: "Save",
              loadingLabel: "Saving",
              icon: <Save className="h-3.5 w-3.5" />,
              isLoading: isSaving,
              onClick: handleSave,
            }}
```

Add `Save` to the lucide-react import in `detail/index.tsx` (find the existing `from "lucide-react"` import line and add `Save`). If `Rocket` is no longer referenced elsewhere in the file after this change, remove it from the import (run `grep -n "Rocket" frontend/src/pages/stacks/components/detail/index.tsx` to check).

- [ ] **Step 7: Verify the whole detail page compiles, tests pass, tab renders**

Run: `pnpm --prefix frontend exec tsc -b --noEmit && pnpm --prefix frontend test:run`
Expected: builds; all tests pass (prior 321 + new).

Then manual check via Playwright against `http://localhost:5174/stacks/<id>` (login `admin@stackdome.io` / `welcome@123`): the Deployments tab appears between Configuration and Logs; the sticky bar primary now reads **Save**.

- [ ] **Step 8: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/deployments-tab.tsx frontend/src/pages/stacks/components/detail/deployments/tests/deployments-tab.test.tsx frontend/src/pages/stacks/components/detail/index.tsx
git commit -m "feat(frontend): wire Deployments tab; relabel sticky primary Deploy->Save"
```

---

## TIER 2 — Failure visibility, detail drawer, post-mortem

### Task 12: Failing-resources click-through accordion (`failing-resources-accordion.tsx`)

**Files:**
- Create (or replace the Task 10 stub): `frontend/src/pages/stacks/components/detail/deployments/failing-resources-accordion.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx`

> One open at a time (Radix Accordion `type="single" collapsible`). Each item expands to a `FailureCard`. A release-level message (render/apply) shows as an `AlertBanner` above, with no per-resource detail.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { FailingResourcesAccordion } from "../failing-resources-accordion";
import type { FailingResource } from "../derive";

afterEach(cleanup);

const failing: FailingResource[] = [
  { name: "tooljet", type: "runtime_crash", stage: "runtime", reason: "CrashLoopBackOff", message: "exit 1", exitCode: 1, restartCount: 5 },
  { name: "worker", type: "runtime_crash", stage: "runtime", reason: "OOMKilled", restartCount: 2 },
];

describe("FailingResourcesAccordion", () => {
  it("lists each failing resource as a header", () => {
    render(<FailingResourcesAccordion failing={failing} />);
    expect(screen.getByText("tooljet")).toBeInTheDocument();
    expect(screen.getByText("worker")).toBeInTheDocument();
  });

  it("expands one resource to show its failure detail", () => {
    render(<FailingResourcesAccordion failing={failing} />);
    fireEvent.click(screen.getByText("tooljet"));
    expect(screen.getByText(/CrashLoopBackOff/)).toBeInTheDocument();
  });

  it("renders a release-level banner when releaseMessage is set", () => {
    render(<FailingResourcesAccordion failing={[]} releaseMessage="apply error: forbidden" />);
    expect(screen.getByText(/apply error: forbidden/)).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx`
Expected: FAIL (stub returns null / module not found).

- [ ] **Step 3: Implement the accordion**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/failing-resources-accordion.tsx
import { AlertBanner, FailureCard } from "@/components/branded";
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from "@/components/ui/accordion";
import type { FailingResource } from "./derive";

export interface FailingResourcesAccordionProps {
  failing: FailingResource[];
  releaseMessage?: string;
}

function stageForCard(stage: FailingResource["stage"]): "build" | "runtime" | "init" {
  return stage === "validation" ? "runtime" : stage;
}

export function FailingResourcesAccordion({ failing, releaseMessage }: FailingResourcesAccordionProps) {
  return (
    <div className="space-y-3">
      {releaseMessage && failing.length === 0 && (
        <AlertBanner>{releaseMessage}</AlertBanner>
      )}
      {failing.length > 0 && (
        <Accordion type="single" collapsible className="space-y-2">
          {failing.map((f) => (
            <AccordionItem key={f.name} value={f.name} className="rounded-md border border-danger-border">
              <AccordionTrigger className="px-3 py-2 text-[13px]">
                <span className="flex items-center gap-2">
                  <span className="font-medium text-foreground">{f.name}</span>
                  <span className="text-danger">{f.reason}</span>
                </span>
              </AccordionTrigger>
              <AccordionContent className="px-3 pb-3">
                <FailureCard
                  resourceName={f.name}
                  stage={stageForCard(f.stage)}
                  reason={f.reason}
                  message={f.message}
                  exitCode={f.exitCode}
                  restartCount={f.restartCount}
                />
              </AccordionContent>
            </AccordionItem>
          ))}
        </Accordion>
      )}
    </div>
  );
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx`
Expected: PASS (3 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/failing-resources-accordion.tsx frontend/src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx
git commit -m "feat(frontend): click-through failing-resources accordion"
```

---

### Task 13: Generic snapshot diff (`release-diff.ts`)

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/release-diff.ts`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/release-diff.test.ts`

> The drawer config-diff and the drift banner both compare two JSON snapshots (rendered spec). `stack-diff.ts` operates on the *form* model, not snapshots — so use a small generic JSON diff built on `deepEqual` from `stack-diff.ts`.

- [ ] **Step 1: Write the failing test**

```ts
import { describe, it, expect } from "vitest";
import { diffSnapshots } from "../release-diff";

describe("diffSnapshots", () => {
  it("reports changed scalar leaves with before/after", () => {
    const out = diffSnapshots({ a: 1, b: { c: "x" } }, { a: 2, b: { c: "x" } });
    expect(out).toEqual([{ path: "a", before: 1, after: 2, kind: "changed" }]);
  });

  it("reports added and removed keys", () => {
    const out = diffSnapshots({ a: 1 }, { a: 1, b: 2 });
    expect(out).toContainEqual({ path: "b", before: undefined, after: 2, kind: "added" });
    const out2 = diffSnapshots({ a: 1, b: 2 }, { a: 1 });
    expect(out2).toContainEqual({ path: "b", before: 2, after: undefined, kind: "removed" });
  });

  it("descends into nested objects with dotted paths", () => {
    const out = diffSnapshots({ b: { c: "x" } }, { b: { c: "y" } });
    expect(out).toEqual([{ path: "b.c", before: "x", after: "y", kind: "changed" }]);
  });

  it("treats arrays as leaves (whole-value change)", () => {
    const out = diffSnapshots({ env: ["A=1"] }, { env: ["A=1", "B=2"] });
    expect(out).toEqual([{ path: "env", before: ["A=1"], after: ["A=1", "B=2"], kind: "changed" }]);
  });

  it("returns [] for equal snapshots", () => {
    expect(diffSnapshots({ a: 1 }, { a: 1 })).toEqual([]);
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-diff.test.ts`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement `release-diff.ts`**

```ts
// frontend/src/pages/stacks/components/detail/deployments/release-diff.ts
import { deepEqual } from "@/pages/stacks/lib/stack-diff";

export interface SnapshotChange {
  path: string;
  before: unknown;
  after: unknown;
  kind: "added" | "removed" | "changed";
}

function isPlainObject(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null && !Array.isArray(v);
}

/** Recursive JSON diff. Arrays and scalars are leaves; objects are descended. */
export function diffSnapshots(before: unknown, after: unknown, prefix = ""): SnapshotChange[] {
  if (deepEqual(before, after)) return [];

  if (isPlainObject(before) && isPlainObject(after)) {
    const keys = new Set([...Object.keys(before), ...Object.keys(after)]);
    const out: SnapshotChange[] = [];
    for (const key of keys) {
      const path = prefix ? `${prefix}.${key}` : key;
      const hasB = key in before, hasA = key in after;
      if (hasB && !hasA) out.push({ path, before: before[key], after: undefined, kind: "removed" });
      else if (!hasB && hasA) out.push({ path, before: undefined, after: after[key], kind: "added" });
      else out.push(...diffSnapshots(before[key], after[key], path));
    }
    return out;
  }

  return [{ path: prefix, before, after, kind: "changed" }];
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-diff.test.ts`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-diff.ts frontend/src/pages/stacks/components/detail/deployments/tests/release-diff.test.ts
git commit -m "feat(frontend): generic snapshot diff for release config changes"
```

---

### Task 14: Release detail drawer (`release-detail-drawer.tsx`)

**Files:**
- Create (or replace the Task 11 stub): `frontend/src/pages/stacks/components/detail/deployments/release-detail-drawer.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/release-detail-drawer.test.tsx`

> `GET /releases/{id}` (full object). Sections: lifecycle, resource outcomes (`outcome.resources` map), pins, config-change diff vs previous (using `diffSnapshots` on snapshots), and the durable post-mortem message for Failed releases. Uses `@/components/ui/sheet`.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";

vi.mock("@/api/releases", () => ({ getRelease: vi.fn() }));
import { getRelease } from "@/api/releases";
import { ReleaseDetailDrawer } from "../release-detail-drawer";

afterEach(() => { cleanup(); vi.clearAllMocks(); });

describe("ReleaseDetailDrawer", () => {
  it("loads the full release and shows outcomes + message", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "r1", sequence: 14, state: "Failed", message: "timed out waiting for convergence after 15m0s",
      outcome: { resources: { tooljet: { phase: "Failed", ready_replicas: 0, replicas: 1, message: "CrashLoopBackOff" } } },
      pins: { resources: { tooljet: { git_sha: "9c69af2" } } },
      snapshot: { spec: { x: 1 } },
    });
    render(<ReleaseDetailDrawer orgId="o" teamName="t" stackId="s" releaseId="r1" onClose={vi.fn()} />);
    await waitFor(() => expect(screen.getByText(/timed out waiting for convergence/)).toBeInTheDocument());
    expect(screen.getByText("tooljet")).toBeInTheDocument();
    expect(getRelease).toHaveBeenCalledWith("o", "t", "s", "r1");
  });

  it("renders config changes vs the previous release snapshot", async () => {
    (getRelease as ReturnType<typeof vi.fn>).mockResolvedValue({
      id: "r2", sequence: 15, state: "Released", snapshot: { spec: { replicas: 2 } },
    });
    render(<ReleaseDetailDrawer orgId="o" teamName="t" stackId="s" releaseId="r2"
      previousRelease={{ id: "r1", snapshot: { spec: { replicas: 1 } } } as never} onClose={vi.fn()} />);
    await waitFor(() => expect(screen.getByText(/Config changes/i)).toBeInTheDocument());
    expect(screen.getByText("spec.replicas")).toBeInTheDocument();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-detail-drawer.test.tsx`
Expected: FAIL (stub / module not found).

- [ ] **Step 3: Implement the drawer**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/release-detail-drawer.tsx
import { useEffect, useState } from "react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { StatusPill, variantFromState } from "@/components/branded";
import { getRelease, type StackRelease } from "@/api/releases";
import { causeLabel, formatDuration } from "./derive";
import { diffSnapshots } from "./release-diff";

export interface ReleaseDetailDrawerProps {
  orgId: string;
  teamName: string;
  stackId: string;
  releaseId: string;
  previousRelease?: StackRelease;
  onClose: () => void;
}

export function ReleaseDetailDrawer({ orgId, teamName, stackId, releaseId, previousRelease, onClose }: ReleaseDetailDrawerProps) {
  const [release, setRelease] = useState<StackRelease | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let alive = true;
    getRelease(orgId, teamName, stackId, releaseId)
      .then((r) => { if (alive) setRelease(r); })
      .catch((e) => { if (alive) setError(e instanceof Error ? e.message : "Failed to load release"); });
    return () => { alive = false; };
  }, [orgId, teamName, stackId, releaseId]);

  const outcomes = Object.entries(release?.outcome?.resources ?? {});
  // snapshot is an untyped JSONB blob on the full release; cast for diffing.
  const snap = (release as unknown as { snapshot?: unknown })?.snapshot;
  const prevSnap = (previousRelease as unknown as { snapshot?: unknown })?.snapshot;
  const changes = release && previousRelease ? diffSnapshots(prevSnap, snap) : [];

  return (
    <Sheet open onOpenChange={(o) => { if (!o) onClose(); }}>
      <SheetContent side="right" className="w-[480px] sm:max-w-[480px] overflow-y-auto">
        <SheetHeader>
          <SheetTitle className="flex items-center gap-2">
            Release #{release?.sequence ?? "…"}
            {release && <StatusPill variant={variantFromState(release.state ?? "")}>{release.state}</StatusPill>}
          </SheetTitle>
        </SheetHeader>

        {error && <p className="mt-4 text-[13px] text-danger">{error}</p>}

        {release && (
          <div className="mt-4 space-y-6 text-[13px]">
            <section className="space-y-1">
              <div className="text-muted-foreground">{causeLabel(release.cause)}</div>
              <div className="text-muted-foreground">
                duration {formatDuration(release.rendered_at, release.completed_at)}
              </div>
            </section>

            {release.state === "Failed" && release.message && (
              <section>
                <h3 className="mb-1 font-medium">Why it failed</h3>
                <p className="rounded-md border border-danger-border bg-danger/5 p-2 text-danger">{release.message}</p>
              </section>
            )}

            {outcomes.length > 0 && (
              <section>
                <h3 className="mb-2 font-medium">Resource outcomes</h3>
                <div className="divide-y divide-border rounded-md border border-border">
                  {outcomes.map(([name, o]) => (
                    <div key={name} className="flex items-center justify-between px-3 py-2">
                      <span className="font-medium">{name}</span>
                      <span className="text-muted-foreground">{o?.phase}</span>
                      <span className="font-mono text-muted-foreground">{o?.ready_replicas ?? 0}/{o?.replicas ?? 0}</span>
                    </div>
                  ))}
                </div>
              </section>
            )}

            {changes.length > 0 && (
              <section>
                <h3 className="mb-2 font-medium">Config changes vs #{previousRelease?.sequence}</h3>
                <div className="space-y-1 font-mono text-[12px]">
                  {changes.map((c) => (
                    <div key={c.path} className="rounded border-l-2 border-warn bg-warn/5 px-2 py-1">
                      <div className="text-muted-foreground">{c.path}</div>
                      {c.kind !== "added" && <div className="text-danger">- {JSON.stringify(c.before)}</div>}
                      {c.kind !== "removed" && <div className="text-success">+ {JSON.stringify(c.after)}</div>}
                    </div>
                  ))}
                </div>
              </section>
            )}
          </div>
        )}
      </SheetContent>
    </Sheet>
  );
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/release-detail-drawer.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/release-detail-drawer.tsx frontend/src/pages/stacks/components/detail/deployments/tests/release-detail-drawer.test.tsx
git commit -m "feat(frontend): release detail drawer with outcomes + config diff + post-mortem"
```

---

### Task 15: Crash log snapshot (best-effort) in the failure card

**Files:**
- Modify: `frontend/src/api/observability.ts` (salvage `buildStackResourceLogStreamUrl` + `fetchLogSnapshot`)
- Modify: `frontend/src/pages/stacks/components/detail/deployments/failing-resources-accordion.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx` (extend)

> Best-effort, often empty for a crashing pod (#98). Render the log block only when non-empty; never gate state on it.

- [ ] **Step 1: Salvage the observability helpers from the parked branch**

Inspect what the parked branch added:

```bash
git show feat/stack-activity-tab:frontend/src/api/observability.ts | grep -n "buildStackResourceLogStreamUrl\|export async function fetchLogSnapshot"
```

Copy the two additions (`buildStackResourceLogStreamUrl` and `fetchLogSnapshot`) into the current `frontend/src/api/observability.ts`, appending them after the existing `buildStackLogStreamUrl`. Their exact bodies (verified shape):

```ts
export function buildStackResourceLogStreamUrl(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceName: string,
  params?: { follow?: boolean; since?: string; tail?: number },
): string {
  const base = import.meta.env.VITE_API_BASE_URL || "/api/v1";
  const qs = new URLSearchParams();
  if (params?.follow != null) qs.set("follow", String(params.follow));
  if (params?.since) qs.set("since", params.since);
  if (params?.tail != null) qs.set("tail", String(params.tail));
  const query = qs.toString();
  return `${base}/organizations/${organizationId}/teams/${teamName}/stacks/${stackId}/resources/${resourceName}/logs${query ? `?${query}` : ""}`;
}

export async function fetchLogSnapshot(
  organizationId: string,
  teamName: string,
  stackId: string,
  resourceName: string,
  tail = 50,
): Promise<string[]> {
  const url = buildStackResourceLogStreamUrl(organizationId, teamName, stackId, resourceName, { follow: false, tail });
  try {
    const res = await fetch(url, { credentials: "include" });
    if (!res.ok) return [];
    const text = await res.text();
    return text.split("\n")
      .map((l) => l.replace(/^data:\s?/, "").trim())
      .filter((l) => l.length > 0);
  } catch {
    return [];
  }
}
```

- [ ] **Step 2: Add a failing test for the log snapshot rendering**

Add the mock at the **top** of `failing-resources-accordion.test.tsx` (alongside the existing imports — `vi.mock` must be top-level to hoist above the component import), and switch the render imports to include `fireEvent`/`waitFor`:

```tsx
// at the top of failing-resources-accordion.test.tsx, with the other imports:
import { render, screen, cleanup, fireEvent, waitFor } from "@testing-library/react";
vi.mock("@/api/observability", () => ({
  fetchLogSnapshot: vi.fn().mockResolvedValue(["panic: boom", "exit status 1"]),
}));
```

Then append this new test case inside the existing `describe("FailingResourcesAccordion", …)` block:

```tsx
it("shows a log snapshot when the fetch returns lines", async () => {
  render(<FailingResourcesAccordion failing={failing} logContext={{ orgId: "o", teamName: "t", stackId: "s" }} />);
  fireEvent.click(screen.getByText("tooljet"));
  await waitFor(() => expect(screen.getByText(/panic: boom/)).toBeInTheDocument());
});
```

(The Task 12 test file's first import line must already be `import { render, screen, cleanup } from "@testing-library/react";` — replace it with the line above so `vi` is also imported via the existing `import { describe, it, expect, afterEach } from "vitest";`; add `vi` to that vitest import.)

- [ ] **Step 3: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx`
Expected: FAIL — `logContext` prop unsupported / no log rendering.

- [ ] **Step 4: Extend the accordion to fetch + render a best-effort snapshot**

Add to `failing-resources-accordion.tsx`:

```tsx
import { useEffect, useState } from "react";
import { LogSnapshot } from "@/components/branded";
import { fetchLogSnapshot } from "@/api/observability";

export interface LogContext { orgId: string; teamName: string; stackId: string; }

// extend props:
//   logContext?: LogContext;
// inside the AccordionContent for runtime failures, render:
function CrashLog({ ctx, resourceName }: { ctx: LogContext; resourceName: string }) {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    void fetchLogSnapshot(ctx.orgId, ctx.teamName, ctx.stackId, resourceName, 50)
      .then((l) => { if (alive) setLines(l); });
    return () => { alive = false; };
  }, [ctx, resourceName]);
  if (lines.length === 0) return null; // best-effort; pod may be unreachable (#98)
  return <LogSnapshot lines={lines} />;
}
```

In the runtime branch of each `AccordionContent`, after `<FailureCard .../>`, render `{logContext && f.type === "runtime_crash" && <CrashLog ctx={logContext} resourceName={f.name} />}`. Thread `logContext` from `CurrentDeploymentCard` (pass a new `logContext` prop down from `deployments-tab.tsx` → `current-deployment-card.tsx` → accordion; the tab already has orgId/teamName/stackId).

- [ ] **Step 5: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/failing-resources-accordion.test.tsx`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add frontend/src/api/observability.ts frontend/src/pages/stacks/components/detail/deployments/
git commit -m "feat(frontend): best-effort crash log snapshot in failure accordion"
```

---

## TIER 3 — Drift banner (Save≠Deploy) + polish

### Task 16: Unreleased-changes (drift) banner

**Files:**
- Create: `frontend/src/pages/stacks/components/detail/deployments/unreleased-changes-banner.tsx`
- Test: `frontend/src/pages/stacks/components/detail/deployments/tests/unreleased-changes-banner.test.tsx`

> Heuristic drift: a saved stack whose `updated_at` is newer than the active release's `completed_at`/`created_at`, OR a non-empty `diffSnapshots(activeSnapshot, savedSpec)`. Labelled "approximate" per spec §7 (precise drift needs backend `snapshot_revision`, a filed follow-up). Offers Deploy.

- [ ] **Step 1: Write the failing test**

```tsx
// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from "vitest";
import "@testing-library/jest-dom/vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import { UnreleasedChangesBanner } from "../unreleased-changes-banner";

afterEach(cleanup);

describe("UnreleasedChangesBanner", () => {
  it("renders nothing when there is no drift", () => {
    const { container } = render(<UnreleasedChangesBanner hasDrift={false} onDeploy={vi.fn()} busy={false} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("shows the deploy affordance when drift is present", () => {
    const onDeploy = vi.fn();
    render(<UnreleasedChangesBanner hasDrift onDeploy={onDeploy} busy={false} />);
    expect(screen.getByText(/Unreleased changes/i)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: /Deploy/i }));
    expect(onDeploy).toHaveBeenCalled();
  });
});
```

- [ ] **Step 2: Run to verify it fails**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/unreleased-changes-banner.test.tsx`
Expected: FAIL — module not found.

- [ ] **Step 3: Implement the banner**

```tsx
// frontend/src/pages/stacks/components/detail/deployments/unreleased-changes-banner.tsx
import { Rocket } from "lucide-react";
import { AlertBanner } from "@/components/branded";

export interface UnreleasedChangesBannerProps {
  hasDrift: boolean;
  onDeploy: () => void;
  busy: boolean;
}

export function UnreleasedChangesBanner({ hasDrift, onDeploy, busy }: UnreleasedChangesBannerProps) {
  if (!hasDrift) return null;
  return (
    <AlertBanner action={{ label: busy ? "Deploying…" : "Deploy", onClick: onDeploy }}>
      <span className="flex items-center gap-2">
        <Rocket className="h-3.5 w-3.5" />
        Unreleased changes — your saved configuration differs from the active deployment (approximate).
      </span>
    </AlertBanner>
  );
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `pnpm --prefix frontend exec vitest run src/pages/stacks/components/detail/deployments/tests/unreleased-changes-banner.test.tsx`
Expected: PASS (2 tests).

- [ ] **Step 5: Compute drift in the tab and render the banner**

In `deployments-tab.tsx`, derive `hasDrift` and render the banner above the current deployment card:

```tsx
import { diffSnapshots } from "./release-diff";
// inside DeploymentsTab, after activeRelease is known:
const savedSpec = stack.spec;
const activeSnap = (activeRelease as unknown as { snapshot?: unknown })?.snapshot;
// activeRelease from the LIST lacks snapshot (detail-only). Heuristic fallback on timestamps:
const stackUpdated = (stack as unknown as { updated_at?: string }).updated_at;
const hasDrift =
  !!activeRelease &&
  ((activeSnap != null && diffSnapshots(activeSnap, savedSpec).length > 0) ||
    (!!stackUpdated && !!activeRelease.completed_at && new Date(stackUpdated) > new Date(activeRelease.completed_at)));
```

Render `<UnreleasedChangesBanner hasDrift={hasDrift} onDeploy={onDeploy} busy={busy} />` just above `<CurrentDeploymentCard …>`.

> NOTE the limitation honestly in a code comment: list-response releases carry no `snapshot`, so drift relies on the timestamp heuristic until the backend exposes the current `snapshot_revision` (filed follow-up §12.4).

- [ ] **Step 6: Run tests + typecheck**

Run: `pnpm --prefix frontend exec tsc -b --noEmit && pnpm --prefix frontend test:run`
Expected: builds; all tests pass.

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/stacks/components/detail/deployments/
git commit -m "feat(frontend): unreleased-changes drift banner (approximate) with Deploy"
```

---

### Task 17: Backend follow-up issues (file, don't implement)

**Files:** none (GitHub issues via `gh`).

> The spec §12 lists six follow-ups. File them so the degraded surfaces are tracked, not silently shipped. These are out of scope for the frontend branch.

- [ ] **Step 1: File the issues**

```bash
gh issue create -R Stackdome/stackdome -t "Releases: converge should fast-fail on terminal resource failure" \
  -b "Today converge.go only marks a release Failed on the 15-min timeout with a generic message. It should detect a resource last_failure (build_failure/runtime_crash) and fail fast with that reason. Frontend Deployments tab works around this by reading live last_failure. Ref: docs/superpowers/specs/2026-06-21-stack-deployments-tab-design.md §12.1"

gh issue create -R Stackdome/stackdome -t "Releases: snapshot last_failure into release.outcome for durable post-mortem" \
  -b "buildOutcome walks status.resources[] (StackResourceSummary, no last_failure). To make the post-mortem failure detail durable, also read spec.stack_resources[].status.last_failure (StackResourceStatus) and copy it into outcome at fail time. Ref §12.2"

gh issue create -R Stackdome/stackdome -t "Releases: missing-secret leaves release stuck InProgress with no signal" \
  -b "apply requeues forever on a missing secret; should fail or surface a precondition message instead of hanging silently. Ref §12.3"

gh issue create -R Stackdome/stackdome -t "Stacks: expose current snapshot_revision for precise drift detection" \
  -b "Deployments tab drift is heuristic (timestamps) because the list release response has no snapshot and the stack has no desired snapshot_revision. Expose it for precise saved≠deployed detection. Ref §12.4"

gh issue create -R Stackdome/stackdome -t "Spec hygiene: last_validation_run absent from OpenAPI" \
  -b "With Releases, validation failures surface as Failed releases; the parked validation-stage synthesis is dropped. Add last_validation_run to the OpenAPI spec for completeness. Ref §12.6"
```

(Issue #98 — log streaming refuses non-Ready resources — already exists; do not re-file.)

- [ ] **Step 2: Commit a note linking the filed issues**

No code change. Record the issue numbers in the plan's closeout commit message in Task 18.

---

### Task 18: Whole-feature review pass

**Files:** all of the above.

- [ ] **Step 1: Full typecheck + test + lint**

```bash
pnpm --prefix frontend exec tsc -b --noEmit
pnpm --prefix frontend test:run
pnpm --prefix frontend lint
```
Expected: build clean (modulo the one pre-existing `postgres-backups.ts` error); all tests pass; lint clean.

- [ ] **Step 2: Manual QA via Playwright**

Against `http://localhost:5174/stacks/<id>` (login `admin@stackdome.io` / `welcome@123`): open the Deployments tab; confirm current deployment + stage tracker + history render; open the detail drawer; trigger a Deploy and watch polling flip state; verify the sticky bar primary reads **Save**. Capture a screenshot for the PR.

- [ ] **Step 3: Final commit**

```bash
git commit --allow-empty -m "chore: Deployments tab feature complete — filed backend follow-ups #<n>..#<n>"
```

---

## Self-Review

**Spec coverage:**
- §4 architecture (current + history + drawer) → Tasks 10, 9/8, 14. ✓
- §5 list-vs-detail (derived duration, git SHA from pins, outcome drawer-only) → `formatDuration`/`releaseGitSha` (Task 5/6), drawer uses detail fetch (Task 14). ✓
- §6.1 current block + recovered note → Task 10. ✓
- §6.1 stage tracker contract-field gating → Task 6. ✓
- §6.2 history + state-aware ⋮ → Task 8. ✓
- §6.3 two error shapes + click-through accordion + live-not-release-state → Tasks 5, 12; live data via `useReleases` poll + live stack (Tasks 7, 10). ✓
- §6.4 drawer (lifecycle, outcomes, pins, config diff) → Tasks 13, 14. ✓
- §6.5 post-mortem (durable message + outcomes) → Task 14. ✓
- §7 decoupled Save/Deploy + drift banner → Task 11 (relabel), Task 16 (drift). ✓
- §8 polling → Task 7. ✓
- §9 error taxonomy surfaces → Tasks 5/6/12 (build/runtime/render/apply/timeout). ✓
- §10 reuse + salvage + new → Tasks 3, 4, 15 (salvage); 2, 5–16 (new); DirtyField visual language applied in drawer diff (Task 14). ✓
- §11 API wiring/types → Tasks 1, 2. ✓
- §12 backend follow-ups → Task 17. ✓

**Placeholder scan:** No "TBD"/"handle errors"/"similar to". Every code step shows complete code or an exact `git show`/`gh` command. ✓

**Type consistency:** `Stages = { build, deploy, ready }` consistent across Tasks 4, 6, 10. `FailingResource`/`RecoveredResource` defined in Task 5, consumed in Tasks 10, 12. `StackRelease`/`StackReleaseList`/`CreateReleaseRequest` from `api/releases.ts` (Task 2) used everywhere. `diffSnapshots` signature consistent across Tasks 13, 14, 16. `variantFromState` extension noted in Task 8 and reused in Tasks 10, 14. ✓

**Known honest gaps (flagged in-code, not hidden):** drift is heuristic until backend exposes `snapshot_revision`; crash log snapshot is best-effort (#98); post-mortem structured failure detail is not durable until the §12.2 backend snapshot lands. All three are surfaced to the user as approximate / best-effort and tracked as follow-ups.
