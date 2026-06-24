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
  stack: Stack;
  dirty: StackDiff;
  isActive: boolean;
  activeRelease?: StackRelease;
  liveRelease?: StackRelease;
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
 * editing > deploying > staged > clean.
 *
 * `staged` (saved-but-undeployed) is detected by diffing the saved spec against the
 * live release's snapshot. When the live snapshot isn't loaded yet we fall back to
 * a timestamp drift heuristic; when the stack has never deployed, any saved
 * resources count as staged for the first release.
 */
export function deriveDeployLifecycle(args: DeriveDeployLifecycleArgs): DeployLifecycle {
  const { stack, dirty, isActive, activeRelease, liveRelease, liveSnapshot } = args;
  const liveSeq = liveRelease?.sequence;
  const nextSeq = (activeRelease?.sequence ?? 0) + 1;

  if (isActive && dirtyCount(dirty) > 0) {
    return { phase: "editing", liveSeq, nextSeq };
  }
  if (activeRelease && !TERMINAL.has(activeRelease.state ?? "")) {
    return { phase: "deploying", liveSeq, nextSeq };
  }

  // staged: the saved spec differs from what is live.
  if (liveSnapshot) {
    const stagedDiff = diffSnapshots(liveSnapshot, specToSnapshot(stack));
    return diffIsEmpty(stagedDiff)
      ? { phase: "clean", liveSeq, nextSeq }
      : { phase: "staged", stagedDiff, liveSeq, nextSeq };
  }

  if (!liveRelease) {
    // Never deployed — any saved resources are staged for the first release.
    const hasSpec = (stack.spec?.stack_resources?.length ?? 0) > 0;
    return { phase: hasSpec ? "staged" : "clean", liveSeq, nextSeq };
  }

  // Live release exists but its snapshot hasn't loaded — heuristic drift fallback.
  const stackUpdated = (stack as { updated_at?: string }).updated_at;
  const drift = !!stackUpdated && !!liveRelease.completed_at
    && new Date(stackUpdated) > new Date(liveRelease.completed_at);
  return { phase: drift ? "staged" : "clean", liveSeq, nextSeq };
}

export interface UseDeployLifecycleArgs {
  stack: Stack;
  dirty: StackDiff;
  isActive: boolean;
  releases: StackRelease[];
  activeRelease?: StackRelease;
  detail: ReleaseDetail;
}

/**
 * React wrapper: resolves the live release (stack.status.last_converged) from the
 * releases list, lazily loads its snapshot, and derives the lifecycle phase.
 */
export function useDeployLifecycle({ stack, dirty, isActive, releases, activeRelease, detail }: UseDeployLifecycleArgs): DeployLifecycle {
  const liveReleaseId = stack.status?.last_converged?.release_id;
  const liveRelease = liveReleaseId ? releases.find((r) => r.id === liveReleaseId) : undefined;

  useEffect(() => {
    if (liveReleaseId) detail.ensure(liveReleaseId);
  }, [detail, liveReleaseId]);

  const liveSnapshot = detail.peek(liveReleaseId).data?.snapshot;
  return deriveDeployLifecycle({ stack, dirty, isActive, activeRelease, liveRelease, liveSnapshot });
}
