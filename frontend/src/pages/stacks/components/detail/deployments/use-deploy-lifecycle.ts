import { useEffect } from "react";
import type { Stack } from "@/api/stacks";
import type { StackRelease, StackReleaseSnapshot } from "@/api/releases";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";
import type { ReleaseDetail } from "./use-release-detail";
import { diffSnapshots, type SnapshotDiff } from "./release-snapshot-diff";
import { specToSnapshot } from "./spec-to-snapshot";
import { isTerminal } from "./release-states";

export type DeployPhase = "editing" | "staged" | "deploying" | "clean";

export interface DeployLifecycle {
  phase: DeployPhase;
  /** Per-field diff backing the staged draft — present only when staged and the
   *  comparison snapshot was available to diff against. */
  stagedDiff?: SnapshotDiff;
  /** Sequence the staged diff is computed against: the in-flight release when the
   *  draft would supersede it, otherwise the live release. Shown as "vs #N". */
  vsSeq?: number;
  /** Sequence the next deploy would create. */
  nextSeq: number;
}

export interface DeriveDeployLifecycleArgs {
  stack: Stack | undefined;
  dirty: StackDiff;
  isActive: boolean;
  /** The latest release attempt (releases[0]). */
  activeRelease?: StackRelease;
  /** The release currently serving traffic (stack.status.last_converged). */
  liveRelease?: StackRelease;
  /** Snapshot of the latest attempt — what an in-flight deploy is shipping. */
  activeSnapshot?: StackReleaseSnapshot;
  /** Snapshot of the live release — what is actually running. */
  liveSnapshot?: StackReleaseSnapshot;
}

function dirtyCount(d: StackDiff): number {
  return d.dirtyResourceIdx.size + d.dirtyVolumeIdx.size + d.addonLinkCount;
}

function diffIsEmpty(d: SnapshotDiff): boolean {
  return d.resources.length === 0 && d.volumes.length === 0 && d.connections.length === 0;
}

/**
 * Pure lifecycle derivation. Phases are mutually exclusive, evaluated in priority:
 *
 *   editing   — unsaved edits in the session.
 *   deploying — a release is in flight AND it is shipping exactly the current saved
 *               spec (nothing new to deploy). Just a status.
 *   staged    — the saved spec is not what's live AND it isn't already being deployed
 *               as-is. This covers a fresh edit, a fresh edit made mid-deploy (which
 *               can supersede the in-flight release), and retrying a failed release.
 *   clean     — the saved spec matches what's live.
 *
 * Two comparisons drive this: saved-vs-in-flight (has anything new beyond the
 * release in flight?) and saved-vs-live (is the saved config actually running?).
 */
export function deriveDeployLifecycle(args: DeriveDeployLifecycleArgs): DeployLifecycle {
  const { stack, dirty, isActive, activeRelease, liveRelease, activeSnapshot, liveSnapshot } = args;
  const liveSeq = liveRelease?.sequence;
  const nextSeq = (activeRelease?.sequence ?? 0) + 1;
  if (!stack) return { phase: "clean", nextSeq };

  if (isActive && dirtyCount(dirty) > 0) {
    return { phase: "editing", nextSeq };
  }

  const spec = specToSnapshot(stack);
  const deploying = !!activeRelease && !isTerminal(activeRelease.state);

  // A deploy is in flight. The draft is measured against the IN-FLIGHT release,
  // not what's live: if the saved spec matches what's deploying there's nothing
  // new (just a status); if it differs, it's a draft that would supersede the
  // in-flight release — even when the saved spec happens to equal what's live.
  if (deploying) {
    if (!activeSnapshot) return { phase: "deploying", nextSeq }; // snapshot loading
    const d = diffSnapshots(activeSnapshot, spec);
    return diffIsEmpty(d)
      ? { phase: "deploying", nextSeq }
      : { phase: "staged", stagedDiff: d, vsSeq: activeRelease!.sequence, nextSeq };
  }

  // Nothing in flight: the draft is measured against what's live.
  if (liveSnapshot) {
    const d = diffSnapshots(liveSnapshot, spec);
    return diffIsEmpty(d)
      ? { phase: "clean", nextSeq }
      : { phase: "staged", stagedDiff: d, vsSeq: liveSeq, nextSeq };
  }

  const liveReleaseId = stack.status?.last_converged?.release_id;
  if (!liveReleaseId) {
    // Never converged a release — any saved resources are staged for the first deploy.
    const hasSpec = (stack.spec?.stack_resources?.length ?? 0) > 0;
    return { phase: hasSpec ? "staged" : "clean", vsSeq: liveSeq, nextSeq };
  }

  // Live snapshot not loaded yet. If the row is resolved fall back to a timestamp
  // drift heuristic; otherwise stay clean rather than flashing a false "staged".
  // Known false-positive: a metadata-only stack update (or server clock skew) can
  // read as drift — accepted because it's a last-resort guess that self-corrects
  // the moment the live snapshot loads and the real diff is computed.
  if (liveRelease) {
    const drift = !!stack.updated_at && !!liveRelease.completed_at
      && new Date(stack.updated_at) > new Date(liveRelease.completed_at);
    return { phase: drift ? "staged" : "clean", vsSeq: liveSeq, nextSeq };
  }
  return { phase: "clean", nextSeq };
}

export interface UseDeployLifecycleArgs {
  stack: Stack | undefined;
  dirty: StackDiff;
  isActive: boolean;
  releases: StackRelease[];
  activeRelease?: StackRelease;
  detail: ReleaseDetail;
}

/**
 * React wrapper: resolves the live + latest-attempt releases, lazily loads both
 * snapshots, and derives the lifecycle phase.
 */
export function useDeployLifecycle({ stack, dirty, isActive, releases, activeRelease, detail }: UseDeployLifecycleArgs): DeployLifecycle {
  const liveReleaseId = stack?.status?.last_converged?.release_id;
  const liveRelease = liveReleaseId ? releases.find((r) => r.id === liveReleaseId) : undefined;
  const activeId = activeRelease?.id;

  useEffect(() => {
    if (liveReleaseId) detail.ensure(liveReleaseId);
    if (activeId) detail.ensure(activeId);
  }, [detail, liveReleaseId, activeId]);

  const liveSnapshot = detail.peek(liveReleaseId).data?.snapshot;
  const activeSnapshot = detail.peek(activeId).data?.snapshot;
  return deriveDeployLifecycle({ stack, dirty, isActive, activeRelease, liveRelease, activeSnapshot, liveSnapshot });
}
