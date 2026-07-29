import { format, isToday, isYesterday } from "date-fns";
import type { components } from "@/api/types/openapi";
import { ReleaseEventType, type ReleaseEvent, type StackRelease, type ReleaseLiveStatus } from "@/api/releases";
import type { Stages } from "@/components/branded";
import { statusVariant, type StatusVariant } from "@/components/branded/status-variant";
import { ReleaseState, isTerminal } from "./release-states";
import { gitUnpinned } from "./release-snapshot-diff";

export type Stack = components["schemas"]["Stack"];
export type ReleaseHealth = components["schemas"]["ReleaseHealth"];
export type StackResource = components["schemas"]["StackResource"];
export type StackResourceFailure = components["schemas"]["StackResourceFailure"];
export type ReleaseCause = components["schemas"]["ReleaseCause"];
export type FailureStage = "build" | "runtime" | "init" | "validation";

/**
 * Health driving the stack header pill: an in-flight latest release always reads
 * "progressing"; otherwise the live release's rollup, falling back to "failed" for a
 * Failed latest with nothing live. Undefined → nothing ever deployed (no releases,
 * or only cancelled/superseded attempts) → callers render a neutral "Not deployed".
 */
export function deriveHeaderHealth(stack: Stack): ReleaseHealth | undefined {
  const latest = stack.latest_release;
  if (!latest) return undefined;
  if (!isTerminal(latest.state)) return "progressing";
  if (stack.converged_release?.health) return stack.converged_release.health;
  return latest.state === ReleaseState.Failed ? "failed" : undefined;
}

/**
 * The deployed snapshot stores the RESOLVED git revision (branch/commit written by
 * the pin resolver at deploy time). When the saved spec doesn't pin one, those are
 * deploy-time facts, not config drift — strip them so the diff baseline compares
 * intent with intent instead of reading every unpinned git resource as dirty.
 */
export function stripUnpinnedGitRevisions(
  snapshotResources: StackResource[],
  savedResources: StackResource[],
): StackResource[] {
  const savedByName = new Map(savedResources.map((r) => [r.name, r]));
  return snapshotResources.map((r) => {
    const git = r.source?.git;
    if (!git || !gitUnpinned(savedByName.get(r.name))) return r;
    const { branch: _b, commit: _c, tag: _t, ...unpinned } = git;
    return { ...r, source: { ...r.source, git: unpinned } };
  });
}

/**
 * True when the latest release failed AND a different release is currently live —
 * the header's secondary "deploy failed" hint. Mutually exclusive with the main pill
 * by construction: fires only when deriveHeaderHealth is showing something other than
 * "failed" (a live release masking the failed attempt), never doubling up the error.
 */
export function latestDeployFailed(stack: Stack): boolean {
  const latest = stack.latest_release;
  if (!latest || latest.state !== ReleaseState.Failed) return false;
  if (!stack.converged_release || latest.id === stack.converged_release.id) return false;
  return deriveHeaderHealth(stack) !== "failed";
}

/**
 * True while the stack's own release summaries lag a terminal release seen in the
 * polled releases list — the signal to (re)fetch the stack until they catch up.
 * The backend updates latest_release/converged_release asynchronously after a
 * release terminates, so a single refetch can race the pointers and capture the
 * old summaries; staleness must be judged from the CONTENT of the fetched stack,
 * never from "we already refetched once". For a Released release the
 * converged_release pointer must catch up too; failed/cancelled releases leave
 * converged on the previous live release by design.
 */
export function stackSummariesStale(
  active: StackRelease | undefined,
  stack: Pick<Stack, "latest_release" | "converged_release"> | undefined,
): boolean {
  if (!active || !isTerminal(active.state) || !stack) return false;
  const latest = stack.latest_release;
  if (!latest || latest.id !== active.id || latest.state !== active.state) return true;
  return active.state === ReleaseState.Released && stack.converged_release?.id !== active.id;
}

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

/** The generated `StackResourceFailure.type` is a types-only union; this is its runtime mirror. */
export type ResourceFailureTypeValue = NonNullable<StackResourceFailure["type"]>;
export const ResourceFailureType = {
  Build: "build_failure",
  Runtime: "runtime_crash",
} as const satisfies Record<string, ResourceFailureTypeValue>;

export interface FailingResource {
  name: string;
  type: ResourceFailureTypeValue;
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
  if (f.type === ResourceFailureType.Build) return { detail: f.build, stage: "build" as const };
  if (f.init_container) return { detail: f.init_container, stage: "init" as const };
  return { detail: f.container, stage: "runtime" as const };
}

/** Live per-resource statuses come from the release's live_status (present only while the
 *  release is live/converged or actively deploying); undefined until that's wired in.
 *  `_release` isn't read yet — kept for signature parity with deriveStages/deriveRecovered. */
export function deriveFailingResources(_release: StackRelease, liveStatus?: ReleaseLiveStatus): FailingResource[] {
  const resources = liveStatus?.resources ?? {};
  const out: FailingResource[] = [];
  for (const [name, r] of Object.entries(resources)) {
    const f = r.last_failure;
    const state = r.state ?? "";
    // Only surface as ACTIVE failure when the resource is not currently healthy.
    if (!f || isHealthyState(state)) continue;
    const { detail, stage } = failureDetail(f);
    out.push({
      name,
      type: f.type ?? ResourceFailureType.Runtime,
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

export function deriveRecovered(_release: StackRelease, liveStatus?: ReleaseLiveStatus): RecoveredResource[] {
  const resources = liveStatus?.resources ?? {};
  const out: RecoveredResource[] = [];
  for (const [name, r] of Object.entries(resources)) {
    const f = r.last_failure;
    const state = r.state ?? "";
    if (!f || !isHealthyState(state)) continue;
    const { detail } = failureDetail(f);
    out.push({ name, reason: detail?.reason ?? humanizeFailureType(detail?.failure_type), restartCount: detail?.restart_count });
  }
  return out;
}

// Delegate to the single word→variant brain so "healthy" can't drift from the rest
// of the app. Resource live_status.state is the cluster-agent rollout vocabulary.
function isHealthyState(state: string): boolean {
  return statusVariant("resource", state) === "ready";
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
 * deriveFailingResources(release, liveStatus) — recovered resources excluded, so any
 * failure here is CURRENT. Convergence is keyed on release state alone: live_status
 * is present for ACTIVE releases too (overlay presence rule), so its mere presence
 * says nothing about being converged.
 */
export function deriveStages(release: StackRelease, failing: FailingResource[], _liveStatus?: ReleaseLiveStatus): Stages {
  const buildFailed = failing.some((f) => f.type === ResourceFailureType.Build);
  const runtimeFailed = failing.some((f) => f.type === ResourceFailureType.Runtime);
  const hasBuild = hasBuildResources(release);
  const state = release.state;

  // Image-only stack has no build step → Build "skipped" (inert), not "todo".
  if (state === ReleaseState.Released) {
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
  const build = failing.find((f) => f.type === ResourceFailureType.Build);
  const crash = failing.find((f) => f.type === ResourceFailureType.Runtime);
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

/** Timeline render vocabulary. Derived from the canonical variant so the
 *  timeline can never drift from the rest of the app again. */
export function toneFromVariant(v: StatusVariant): Tone {
  switch (v) {
    case "ready": return "ok";
    case "pending": return "amber";
    case "error": return "err";
    default: return "muted"; // info | neutral
  }
}

export function phaseTone(phase: string): Tone {
  return toneFromVariant(statusVariant("rollout", phase));
}

/** Tone for a release's rail dot, keyed off its lifecycle state. */
export function stateTone(state: string): Tone {
  return toneFromVariant(statusVariant("release", state));
}

export function toneTextClass(t: Tone): string {
  return { ok: "text-success", amber: "text-warn", err: "text-danger", muted: "text-fg-muted" }[t];
}

export function toneDotClass(t: Tone): string {
  return { ok: "bg-success", amber: "bg-warn", err: "bg-danger", muted: "bg-fg-muted" }[t];
}

/**
 * Console messages drop the resource name (it sits in its own column) and lead with
 * the verb; the detail after the first ": " is kept. Unknown types pass through.
 */
export function compactEventMessage(e: ReleaseEvent): string {
  const msg = e.message ?? "";
  const detail = (prefix: string) => {
    const i = msg.indexOf(": ");
    return i >= 0 ? `${prefix} — ${msg.slice(i + 2)}` : prefix;
  };
  switch (e.type) {
    case ReleaseEventType.ResourceDeploying: return detail("Deploying");
    case ReleaseEventType.ResourceWaiting: return detail("Waiting");
    case ReleaseEventType.ResourceReady: return "Ready";
    case ReleaseEventType.ResourceFailed: return detail("Failed to start");
    default: return msg;
  }
}

/** Outlined status-pill treatment (split console rail). */
export function tonePillClass(t: Tone): string {
  return {
    ok: "border-success text-success",
    amber: "border-warn text-warn",
    err: "border-danger text-danger",
    muted: "border-fg-muted text-fg-muted",
  }[t];
}
