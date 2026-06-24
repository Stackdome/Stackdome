import { useEffect } from "react";
import type { Stack } from "@/api/stacks";
import type { StackRelease, StackReleaseSnapshot } from "@/api/releases";
import type { StackDiff } from "@/pages/stacks/lib/stack-diff";
import type { ReleaseDetail } from "./use-release-detail";
import { diffSnapshots, type SnapshotDiff } from "./release-snapshot-diff";
import { specToSnapshot } from "./spec-to-snapshot";

export type DeployPhase = "editing" | "staged" | "deploying" | "clean";

const TERMINAL = new Set<string>(["Released", "Failed", "Superseded", "Cancelled"]);

export interface DeployLifecycle {
  phase: DeployPhase;
  /** Per-field diff of the saved spec vs the live release — present only when staged
   *  and the live snapshot was available to diff against. */
  stagedDiff?: SnapshotDiff;
  /** Sequence of the release currently serving traffic (the live release). */
  liveSeq?: number;
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
  if (!stack) return { phase: "clean", liveSeq, nextSeq };

  if (isActive && dirtyCount(dirty) > 0) {
    return { phase: "editing", liveSeq, nextSeq };
  }

  const spec = specToSnapshot(stack);
  const deploying = !!activeRelease && !TERMINAL.has(activeRelease.state ?? "");

  // A deploy is in flight. If it is shipping exactly the saved spec there's nothing
  // new to deploy — show the deploying status. If the snapshot hasn't loaded yet,
  // assume it matches (show status) rather than flashing a Deploy action.
  if (deploying) {
    if (!activeSnapshot || diffIsEmpty(diffSnapshots(activeSnapshot, spec))) {
      return { phase: "deploying", liveSeq, nextSeq };
    }
    // else: there are changes beyond the in-flight release — fall through to staged.
  }

  // staged: the saved spec differs from what is live.
  if (liveSnapshot) {
    const stagedDiff = diffSnapshots(liveSnapshot, spec);
    return diffIsEmpty(stagedDiff)
      ? { phase: "clean", liveSeq, nextSeq }
      : { phase: "staged", stagedDiff, liveSeq, nextSeq };
  }

  const liveReleaseId = stack.status?.last_converged?.release_id;
  if (!liveReleaseId) {
    // Never converged a release — any saved resources are staged for the first deploy.
    const hasSpec = (stack.spec?.stack_resources?.length ?? 0) > 0;
    return { phase: hasSpec ? "staged" : "clean", liveSeq, nextSeq };
  }

  // Live snapshot not loaded yet. If the row is resolved fall back to a timestamp
  // drift heuristic; otherwise stay clean rather than flashing a false "staged".
  if (liveRelease) {
    const stackUpdated = (stack as { updated_at?: string }).updated_at;
    const drift = !!stackUpdated && !!liveRelease.completed_at
      && new Date(stackUpdated) > new Date(liveRelease.completed_at);
    return { phase: drift ? "staged" : "clean", liveSeq, nextSeq };
  }
  return { phase: "clean", liveSeq, nextSeq };
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
