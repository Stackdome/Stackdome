import { format, isToday, isYesterday } from "date-fns";
import type { components } from "@/api/types/openapi";
import type { StackRelease } from "@/api/releases";
import type { Stages } from "@/components/branded";
import { ReleaseState } from "./release-states";

export type Stack = components["schemas"]["Stack"];
export type StackResource = components["schemas"]["StackResource"];
export type StackResourceFailure = components["schemas"]["StackResourceFailure"];
export type ReleaseCause = components["schemas"]["ReleaseCause"];
export type FailureStage = "build" | "runtime" | "init" | "validation";

export interface ResourceSource {
  kind: "image" | "git";
  label: string;
}

// hide trivial "n/n" — surface only when not fully available.
export function replicaLabel(ready?: number, desired?: number): string | undefined {
  const r = ready ?? 0;
  const d = desired ?? 0;
  return r === d ? undefined : `${r}/${d}`;
}

// image resources → image ref; git resources → repo url. mirrors config page sourceType.
export function resourceSource(r?: StackResource): ResourceSource | undefined {
  if (!r) return undefined;
  if (r.source?.git) {
    const url = r.source.git.repo_url;
    return url ? { kind: "git", label: url } : undefined;
  }
  const image = r.source?.image?.ref;
  return image ? { kind: "image", label: image } : undefined;
}

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
    case "rollback": {
      // Backend sends detail as a sentence ("rollback to release #1"); pull the
      // trailing sequence number so we render a clean "Rollback to #1".
      const seq = cause.detail?.match(/(\d+)\s*$/);
      return seq ? `Rollback to #${seq[1]}` : "Rollback";
    }
    case "webhook_push": return "Webhook push";
    case "manual": default: return "Manual deploy";
  }
}

/** Short, day-relative timestamp for the history rail: "today 12:14", "yesterday 17:44", "Jun 3 09:02". */
export function formatReleaseTime(iso?: string): string {
  if (!iso) return "";
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const time = format(d, "HH:mm");
  if (isToday(d)) return `today ${time}`;
  if (isYesterday(d)) return `yesterday ${time}`;
  return `${format(d, "MMM d")} ${time}`;
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
 * Build→Deploy→Ready tracker state. `failing` MUST be the live unhealthy set from
 * deriveFailingResources(stack) — recovered resources excluded, so any failure here is CURRENT.
 */
export function deriveStages(stack: Stack, release: StackRelease, failing: FailingResource[]): Stages {
  const converged = stack.status?.last_converged?.release_id != null
    && stack.status?.last_converged?.release_id === release.id;
  const buildFailed = failing.some((f) => f.type === "build_failure");
  const runtimeFailed = failing.some((f) => f.type === "runtime_crash");
  const hasBuild = hasBuildResources(release);
  const state = release.state;

  // Image-only stack has no build step → Build "skipped" (inert), not "todo".
  if (converged || state === ReleaseState.Released) {
    return { build: hasBuild ? "done" : "skipped", deploy: "done", ready: "done" };
  }
  if (buildFailed) return { build: "failed", deploy: "todo", ready: "todo" };

  if (state === ReleaseState.Pending) {
    return hasBuild
      ? { build: "active", deploy: "todo", ready: "todo" }
      : { build: "skipped", deploy: "active", ready: "todo" };
  }
  if (state === ReleaseState.InProgress) {
    return {
      build: hasBuild ? "done" : "skipped",
      deploy: runtimeFailed ? "failed" : "active",
      ready: "todo",
    };
  }
  if (state === ReleaseState.Failed) {
    // Worker records per-resource `outcome` only after reaching the convergence loop
    // (render+apply succeeded, workload deployed). Pre-cluster failures store none.
    const reachedCluster = Object.keys(release.outcome?.resources ?? {}).length > 0;
    // Reached cluster but never Ready (timeout/runtime crash) → failure on Ready.
    if (runtimeFailed || reachedCluster) {
      return { build: hasBuild ? "done" : "skipped", deploy: "done", ready: "failed" };
    }
    // Pre-cluster failure (render/apply/secret): nothing deployed. Lands on Build if
    // the stack builds, else on Deploy (image-only stack skips Build).
    return hasBuild
      ? { build: "failed", deploy: "todo", ready: "todo" }
      : { build: "skipped", deploy: "failed", ready: "todo" };
  }
  // Superseded / Cancelled → neutral.
  return { build: "todo", deploy: "todo", ready: "todo" };
}

/** Short title after the sequence on the live release card, e.g. "Runtime crash — tooljet", "Build queued". */
export function deriveReleaseTitle(release: StackRelease, failing: FailingResource[], stages: Stages): string {
  const state = release.state ?? "";
  const build = failing.find((f) => f.type === "build_failure");
  const crash = failing.find((f) => f.type === "runtime_crash");
  if (build) return `Build failed: ${build.name}`;
  // A terminal Failed crash reads as "Deploy failed"; an in-flight one names the resource.
  if (crash && state !== ReleaseState.Failed) return `Runtime crash: ${crash.name}`;
  switch (state) {
    case ReleaseState.Pending: return stages.build === "active" ? "Build queued" : "Deploying";
    case ReleaseState.InProgress: return "Deploying";
    case ReleaseState.Failed: return "Deploy failed";
    case ReleaseState.Released: return "Released";
    case ReleaseState.Superseded: return "Superseded";
    case ReleaseState.Cancelled: return "Cancelled";
    default: return state;
  }
}

export type Tone = "ok" | "amber" | "err" | "muted";

export function phaseTone(phase: string): Tone {
  if (/ready|available|running|healthy/i.test(phase)) return "ok";
  if (/progress|building|deploying|pending.*build/i.test(phase)) return "amber";
  if (/crash|oom|error|failed|imagepull|backoff/i.test(phase)) return "err";
  return "muted";
}

/** Tone for a release's rail dot, keyed off its lifecycle state. */
export function stateTone(state: string): Tone {
  switch (state) {
    case ReleaseState.Released: return "ok";
    case ReleaseState.Failed: return "err";
    case ReleaseState.Pending:
    case ReleaseState.InProgress: return "amber";
    default: return "muted"; // Superseded, Cancelled, unknown
  }
}

export function toneTextClass(t: Tone): string {
  return { ok: "text-success", amber: "text-warn", err: "text-danger", muted: "text-fg-muted" }[t];
}

export function toneDotClass(t: Tone): string {
  return { ok: "bg-success", amber: "bg-warn", err: "bg-danger", muted: "bg-fg-muted" }[t];
}
