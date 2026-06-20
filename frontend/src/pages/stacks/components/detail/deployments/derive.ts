import type { components } from "@/api/types/openapi";
import type { StackRelease } from "@/api/releases";
import type { Stages } from "@/components/branded";

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

export function releaseGitSha(release: StackRelease): string | undefined {
  const map = release.pins?.resources ?? {};
  // First non-empty git_sha wins; multi-service stacks normally share one source repo.
  for (const p of Object.values(map)) {
    if (p?.git_sha) return p.git_sha;
  }
  return undefined;
}

/** True if the release pins any resource with a git_sha (i.e. a build happened). */
function hasBuildResources(release: StackRelease): boolean {
  return releaseGitSha(release) !== undefined;
}

/**
 * Derives the Build→Deploy→Ready tracker state.
 * `failing` MUST be the live, currently-unhealthy failure set from
 * deriveFailingResources(stack) — healthy/recovered resources are already
 * excluded, so a buildFailed/runtimeFailed here always reflects a CURRENT failure.
 */
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
    // Pre-cluster (render/apply/timeout) → map to first node per spec.
    return { build: "failed", deploy: "todo", ready: "todo" };
  }
  // Superseded / Cancelled → neutral.
  return { build: "todo", deploy: "todo", ready: "todo" };
}
