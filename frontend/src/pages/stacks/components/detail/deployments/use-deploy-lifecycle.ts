import { useEffect } from "react";
import type { Stack } from "@/api/stacks";
import type { StackRelease, StackReleaseSnapshot } from "@/api/releases";
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
  /** Edits not yet persisted by the autosave engine (sync in flight or failing).
   *  Session dirt vs the DEPLOYED baseline is intentionally not "editing" — with
   *  autosave, saved-but-undeployed content is the "staged" phase's job. */
  unsaved: boolean;
  /** The latest release attempt (releases[0]). */
  activeRelease?: StackRelease;
  /** The release currently serving traffic (stack.status.last_converged). */
  liveRelease?: StackRelease;
  /** Snapshot of the latest attempt — what an in-flight deploy is shipping. */
  activeSnapshot?: StackReleaseSnapshot;
  /** Snapshot of the live release — what is actually running. */
  liveSnapshot?: StackReleaseSnapshot;
}

function diffIsEmpty(d: SnapshotDiff): boolean {
  return d.resources.length === 0 && d.volumes.length === 0 && d.connections.length === 0;
}

/**
 * Pure lifecycle derivation. Mutually-exclusive phases, in priority:
 *   editing   — unsaved edits in the session.
 *   deploying — a release is in flight shipping exactly the saved spec (nothing new).
 *   staged    — saved spec differs from live and isn't already deploying as-is
 *               (fresh edit, mid-deploy supersede, or retry of a failed release).
 *   clean     — saved spec matches what's live.
 *
 * Driven by two comparisons: saved-vs-in-flight and saved-vs-live.
 */
export function deriveDeployLifecycle(args: DeriveDeployLifecycleArgs): DeployLifecycle {
  const { stack, unsaved, activeRelease, liveRelease, activeSnapshot, liveSnapshot } = args;
  const liveSeq = liveRelease?.sequence;
  const nextSeq = (activeRelease?.sequence ?? 0) + 1;
  if (!stack) return { phase: "clean", nextSeq };

  if (unsaved) {
    return { phase: "editing", nextSeq };
  }

  const spec = specToSnapshot(stack);
  const deploying = !!activeRelease && !isTerminal(activeRelease.state);

  // Deploy in flight: measure the draft against the IN-FLIGHT release, not live. Matches it →
  // nothing new (status only); differs → a draft that would supersede the in-flight release.
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

  // Live snapshot not loaded yet: fall back to a timestamp drift heuristic (else stay clean,
  // not a false "staged"). Known false-positive: a metadata-only update or clock skew reads as
  // drift — accepted, since it self-corrects once the live snapshot loads and the real diff runs.
  if (liveRelease) {
    const drift = !!stack.updated_at && !!liveRelease.completed_at
      && new Date(stack.updated_at) > new Date(liveRelease.completed_at);
    return { phase: drift ? "staged" : "clean", vsSeq: liveSeq, nextSeq };
  }
  return { phase: "clean", nextSeq };
}

export interface UseDeployLifecycleArgs {
  stack: Stack | undefined;
  unsaved: boolean;
  releases: StackRelease[];
  activeRelease?: StackRelease;
  detail: ReleaseDetail;
}

/**
 * React wrapper: resolves the live + latest-attempt releases, lazily loads both
 * snapshots, and derives the lifecycle phase.
 */
export function useDeployLifecycle({ stack, unsaved, releases, activeRelease, detail }: UseDeployLifecycleArgs): DeployLifecycle {
  const liveReleaseId = stack?.status?.last_converged?.release_id;
  const liveRelease = liveReleaseId ? releases.find((r) => r.id === liveReleaseId) : undefined;
  const activeId = activeRelease?.id;

  useEffect(() => {
    if (liveReleaseId) detail.ensure(liveReleaseId);
    if (activeId) detail.ensure(activeId);
  }, [detail, liveReleaseId, activeId]);

  const liveSnapshot = detail.peek(liveReleaseId).data?.snapshot;
  const activeSnapshot = detail.peek(activeId).data?.snapshot;
  return deriveDeployLifecycle({ stack, unsaved, activeRelease, liveRelease, activeSnapshot, liveSnapshot });
}
