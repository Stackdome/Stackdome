import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { Stack } from "@/api/stacks";
import { getStackById } from "@/api/stacks";
import { createStackResource, updateStackResource, deleteStackResource } from "@/api/stack-resources";
import { createStackConnection, updateStackConnection, deleteStackConnection } from "@/api/connections";
import { createStackVolume } from "@/api/volumes";
import { cloneJson } from "@/pages/stacks/lib/stack-diff";
import {
  DEBOUNCE_IDLE_MS,
  DEBOUNCE_MAX_WAIT_MS,
  RETRY_BASE_MS,
  RETRY_MAX_MS,
  SYNC_STATUS,
  type SyncStatus,
} from "@/pages/stacks/lib/draft-sync/constants";
import { serverStateFromStack, type ServerStackState } from "@/pages/stacks/lib/draft-sync/server-state";
import { buildDesiredState } from "@/pages/stacks/lib/draft-sync/desired-state";
import { computeSyncOps, type SyncOp } from "@/pages/stacks/lib/draft-sync/ops";
import type { EditSessionDraft, UseStackEditSession } from "./use-stack-edit-session";

export interface UseDraftSyncArgs {
  enabled: boolean;
  stack: Stack | undefined;
  session: UseStackEditSession;
  ids: { orgId: string; teamName: string; stackId: string } | null;
  onStackRefreshed: (stack: Stack) => void;
}

export interface UseDraftSync {
  status: SyncStatus;
  failureCount: number;
  flush: () => Promise<boolean>;
  /** External writers (revert) hand the fresh stack here so the mirror stays truthful. */
  notifyExternalUpdate: (stack: Stack) => void;
}

type Ids = NonNullable<UseDraftSyncArgs["ids"]>;

async function executeOp(op: SyncOp, ids: Ids): Promise<void> {
  switch (op.kind) {
    case "createVolume":
      await createStackVolume(ids.orgId, ids.teamName, ids.stackId, op.volume);
      return;
    case "createResource":
      await createStackResource(ids.orgId, ids.teamName, ids.stackId, op.resource);
      return;
    case "updateResource":
      await updateStackResource(ids.orgId, ids.teamName, ids.stackId, op.name, op.resource);
      return;
    case "deleteResource":
      await deleteStackResource(ids.orgId, ids.teamName, ids.stackId, op.name);
      return;
    case "createConnection":
      await createStackConnection(ids.orgId, ids.teamName, ids.stackId, op.conn);
      return;
    case "updateConnection":
      await updateStackConnection(ids.orgId, ids.teamName, ids.stackId, op.id, op.conn);
      return;
    case "deleteConnection":
      await deleteStackConnection(ids.orgId, ids.teamName, ids.stackId, op.id);
      return;
  }
}

/**
 * Debounced, single-flight autosave of the edit-session draft through the thin
 * stack endpoints. Owns the server mirror (the only place server ids live).
 * On success the session baseline advances to the synced snapshot, so edits
 * made mid-flight stay dirty and trigger the next cycle.
 */
export function useDraftSync({
  enabled,
  stack,
  session,
  ids,
  onStackRefreshed,
}: UseDraftSyncArgs): UseDraftSync {
  const [status, setStatus] = useState<SyncStatus>(SYNC_STATUS.idle);
  const [failureCount, setFailureCount] = useState(0);

  const sessionRef = useRef(session);
  sessionRef.current = session;
  const idsRef = useRef(ids);
  idsRef.current = ids;
  const onRefreshedRef = useRef(onStackRefreshed);
  onRefreshedRef.current = onStackRefreshed;

  const mirrorRef = useRef<ServerStackState | null>(null);
  const runningRef = useRef<Promise<boolean> | null>(null);
  const queuedRef = useRef(false);
  const failuresRef = useRef(0);
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const maxWaitTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Seed the mirror once from the fetched stack; afterwards the engine's own
  // refetches (and notifyExternalUpdate) keep it truthful.
  useEffect(() => {
    if (enabled && stack && !mirrorRef.current) {
      mirrorRef.current = serverStateFromStack(stack);
    }
  }, [enabled, stack]);

  const clearDebounceTimers = useCallback(() => {
    if (idleTimerRef.current) {
      clearTimeout(idleTimerRef.current);
      idleTimerRef.current = null;
    }
    if (maxWaitTimerRef.current) {
      clearTimeout(maxWaitTimerRef.current);
      maxWaitTimerRef.current = null;
    }
  }, []);

  const startCycle = useCallback((): Promise<boolean> => {
    if (runningRef.current) {
      queuedRef.current = true;
      return runningRef.current;
    }
    const run = (async (): Promise<boolean> => {
      const s = sessionRef.current;
      const currentIds = idsRef.current;
      const mirror = mirrorRef.current;
      if (!currentIds || !mirror || !s.isActive) return true;

      const snapshot: EditSessionDraft = {
        resources: cloneJson(s.draft.resources),
        volumes: cloneJson(s.draft.volumes),
      };
      const desired = buildDesiredState(snapshot);
      const ops = computeSyncOps(mirror, desired);
      if (ops.length === 0) {
        failuresRef.current = 0;
        setFailureCount(0);
        setStatus((prev) => (prev === SYNC_STATUS.idle ? SYNC_STATUS.idle : SYNC_STATUS.saved));
        return true;
      }

      setStatus(SYNC_STATUS.saving);
      try {
        for (const op of ops) await executeOp(op, currentIds);
        const fresh = await getStackById(currentIds.orgId, currentIds.teamName, currentIds.stackId);
        mirrorRef.current = serverStateFromStack(fresh);
        onRefreshedRef.current(fresh);
        // Deliberately NOT rebasing the session here: the diff baseline stays
        // pinned to the deployed release so autosaved edits remain visibly
        // dirty/revertable. Only deploy or discard moves the baseline.
        failuresRef.current = 0;
        setFailureCount(0);
        setStatus(SYNC_STATUS.saved);
        return true;
      } catch {
        failuresRef.current += 1;
        setFailureCount(failuresRef.current);
        setStatus(SYNC_STATUS.error);
        // Heal the mirror from server truth; the draft stays authoritative locally.
        try {
          const fresh = await getStackById(currentIds.orgId, currentIds.teamName, currentIds.stackId);
          mirrorRef.current = serverStateFromStack(fresh);
          onRefreshedRef.current(fresh);
        } catch {
          /* keep the stale mirror; the next attempt refetches again */
        }
        const backoff = Math.min(RETRY_BASE_MS * 2 ** (failuresRef.current - 1), RETRY_MAX_MS);
        if (retryTimerRef.current) clearTimeout(retryTimerRef.current);
        retryTimerRef.current = setTimeout(() => {
          void startCycle();
        }, backoff);
        return false;
      }
    })().finally(() => {
      runningRef.current = null;
      if (queuedRef.current) {
        queuedRef.current = false;
        idleTimerRef.current = setTimeout(() => {
          void startCycle();
        }, 0);
      }
    });
    runningRef.current = run;
    return run;
  }, []);

  // Debounce: every draft change (while active+enabled) restarts the idle
  // window; a max-wait timer guarantees persistence under continuous typing.
  const draft = session.isActive ? session.draft : null;
  useEffect(() => {
    if (!enabled || !draft || !idsRef.current) return;
    if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
    idleTimerRef.current = setTimeout(() => {
      clearDebounceTimers();
      void startCycle();
    }, DEBOUNCE_IDLE_MS);
    if (!maxWaitTimerRef.current) {
      maxWaitTimerRef.current = setTimeout(() => {
        clearDebounceTimers();
        void startCycle();
      }, DEBOUNCE_MAX_WAIT_MS);
    }
    return () => {
      /* timers cleared by the next run or unmount */
    };
  }, [enabled, draft, clearDebounceTimers, startCycle]);

  const flush = useCallback(async (): Promise<boolean> => {
    clearDebounceTimers();
    if (retryTimerRef.current) {
      clearTimeout(retryTimerRef.current);
      retryTimerRef.current = null;
    }
    if (runningRef.current) {
      const ok = await runningRef.current;
      if (!ok) return false;
    }
    return startCycle();
  }, [clearDebounceTimers, startCycle]);
  const flushRef = useRef(flush);
  flushRef.current = flush;

  const notifyExternalUpdate = useCallback((fresh: Stack) => {
    mirrorRef.current = serverStateFromStack(fresh);
  }, []);

  // Best-effort persistence when the tab hides or the page unmounts.
  useEffect(() => {
    if (!enabled) return;
    const onVisibility = () => {
      if (document.visibilityState === "hidden") void flushRef.current();
    };
    document.addEventListener("visibilitychange", onVisibility);
    return () => {
      document.removeEventListener("visibilitychange", onVisibility);
      void flushRef.current();
    };
  }, [enabled]);

  useEffect(
    () => () => {
      clearDebounceTimers();
      if (retryTimerRef.current) clearTimeout(retryTimerRef.current);
    },
    [clearDebounceTimers],
  );

  return useMemo(
    () => ({ status, failureCount, flush, notifyExternalUpdate }),
    [status, failureCount, flush, notifyExternalUpdate],
  );
}
