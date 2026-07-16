import { SYNC_STATUS } from "@/pages/stacks/lib/draft-sync/constants";

type SyncStatus = (typeof SYNC_STATUS)[keyof typeof SYNC_STATUS];

export interface StagedDiffCounts {
  resources: readonly unknown[];
  volumes: readonly unknown[];
  connections: readonly unknown[];
}

/** "View changes" badge + modal count must agree with the modal's BODY. The body
 * renders the staged diff (saved spec vs release), which only sees SAVED content.
 * When a sync is in flight or has errored (e.g. a fresh resource whose autosave
 * 400s), the unsaved edit is absent from the staged diff — so the diff can be
 * empty while real session dirt exists. In that unsettled state, never let the
 * (stale) staged count hide the dirt: take the larger of the two. Once the sync
 * settles, the staged diff is authoritative (session dirt can overcount, e.g. a
 * mount added then removed nets to zero staged while the session still reads
 * dirty), so trust it and fall back to dirt only until the diff resolves. */
export function resolveChangeCount(
  stagedDiff: StagedDiffCounts | null | undefined,
  dirtyTotal: number,
  syncStatus: SyncStatus,
): number {
  const stagedCount = stagedDiff
    ? stagedDiff.resources.length + stagedDiff.volumes.length + stagedDiff.connections.length
    : null;
  const syncUnsettled = syncStatus === SYNC_STATUS.saving || syncStatus === SYNC_STATUS.error;
  return syncUnsettled ? Math.max(stagedCount ?? 0, dirtyTotal) : stagedCount ?? dirtyTotal;
}
