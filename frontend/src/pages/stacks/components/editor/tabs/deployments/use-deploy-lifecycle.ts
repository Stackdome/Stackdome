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
  /** The release currently serving traffic (stack.converged_release). */
  liveRelease?: StackRelease;
  /** Snapshot of the latest attempt — what an in-flight deploy is shipping. */
  activeSnapshot?: StackReleaseSnapshot;
  /** Snapshot of the live release — what is actually running. */
  liveSnapshot?: StackReleaseSnapshot;
  /** Live edit-session draft in snapshot shape. When present it is the "user's
   *  spec" side of every diff, replacing the server-fetched spec — staged
   *  surfaces then update on keystroke instead of after the autosave
   *  round-trip. Absent → server spec (no session, read-only viewers). */
  draftSnapshot?: StackReleaseSnapshot;
}

function diffIsEmpty(d: SnapshotDiff): boolean {
  return d.resources.length === 0 && d.volumes.length === 0;
}

/** Baseline for a stack that has never converged: every saved resource reads as "added". */
const EMPTY_SNAPSHOT: StackReleaseSnapshot = { resources: [], volumes: [], connections: [] };

/**
 * Baseline snapshot the saved/draft spec is diffed against, mirroring the phase rules:
 * the in-flight release while one is deploying, otherwise the live release, and an empty
 * baseline when the stack has never converged (first deploy). Undefined when a live release
 * exists but its snapshot hasn't loaded yet — no honest baseline is available, so no diff
 * is derivable (the drift heuristic covers phase in that transient window).
 */
function baselineSnapshot(args: DeriveDeployLifecycleArgs): StackReleaseSnapshot | undefined {
  const { stack, activeRelease, activeSnapshot, liveSnapshot } = args;
  const deploying = !!activeRelease && !isTerminal(activeRelease.state);
  if (deploying) return activeSnapshot;
  if (liveSnapshot) return liveSnapshot;
  if (!stack?.converged_release?.id) return EMPTY_SNAPSHOT;
  return undefined;
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
    // Keep the "editing" phase, but surface a diff so the changes panel lists the
    // saved-so-far changes instead of a placeholder. Baseline follows the same rules
    // as the staged path; undefined only while a live snapshot is still loading.
    const base = baselineSnapshot(args);
    const stagedDiff = base ? diffSnapshots(base, args.draftSnapshot ?? specToSnapshot(stack)) : undefined;
    // Mirror the staged path's "vs #N" anchor so the label doesn't vanish while an
    // edit is still autosaving: the in-flight release when deploying, else live.
    const deploying = !!activeRelease && !isTerminal(activeRelease.state);
    const vsSeq = deploying ? activeRelease!.sequence : liveSeq;
    return { phase: "editing", stagedDiff, vsSeq, nextSeq };
  }

  const spec = args.draftSnapshot ?? specToSnapshot(stack);
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

  const liveReleaseId = stack.converged_release?.id;
  if (!liveReleaseId) {
    // Never converged a release — the whole saved spec is staged for the first deploy,
    // so diff against an empty baseline (every resource/volume/connection reads as added).
    const d = diffSnapshots(EMPTY_SNAPSHOT, spec);
    return diffIsEmpty(d)
      ? { phase: "clean", vsSeq: liveSeq, nextSeq }
      : { phase: "staged", stagedDiff: d, vsSeq: liveSeq, nextSeq };
  }

  // Live snapshot not loaded yet: fall back to a timestamp drift heuristic (else stay clean,
  // not a false "staged"). No stagedDiff here on purpose — a live release exists but its
  // snapshot is missing, so there is no honest baseline to diff against (an empty baseline
  // would mislabel already-deployed resources as added). The panel body shows its "saving"
  // hint for this transient window; it self-corrects once the live snapshot loads and the
  // real diff runs. Known false-positive: a metadata-only update or clock skew reads as drift.
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
  /** See DeriveDeployLifecycleArgs.draftSnapshot. */
  draftSnapshot?: StackReleaseSnapshot;
}

/**
 * React wrapper: resolves the live + latest-attempt releases, lazily loads both
 * snapshots, and derives the lifecycle phase.
 */
export function useDeployLifecycle({ stack, unsaved, releases, activeRelease, detail, draftSnapshot }: UseDeployLifecycleArgs): DeployLifecycle {
  const liveReleaseId = stack?.converged_release?.id;
  const liveRelease = liveReleaseId ? releases.find((r) => r.id === liveReleaseId) : undefined;
  const activeId = activeRelease?.id;

  const ensure = detail.ensure;
  useEffect(() => {
    if (liveReleaseId) ensure(liveReleaseId);
    if (activeId) ensure(activeId);
  }, [ensure, liveReleaseId, activeId]);

  const liveSnapshot = detail.peek(liveReleaseId).data?.snapshot;
  const activeSnapshot = detail.peek(activeId).data?.snapshot;
  return deriveDeployLifecycle({ stack, unsaved, activeRelease, liveRelease, activeSnapshot, liveSnapshot, draftSnapshot });
}
