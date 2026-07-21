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
import { isBadRequestError } from "@/api/client";
import { parseApiError, type ParsedFieldError } from "@/api/errors";
import type { EditSessionDraft, UseStackEditSession } from "./use-stack-edit-session";

/** Surfaced to the caller when an autosave op fails, carrying the offending op
 *  so the UI can toast the reason and mark the responsible field inline. `op` is
 *  absent when the ops all landed but the post-save refetch failed — the save
 *  actually succeeded, so there is no op-level field error to attribute.
 *  `fieldErrors` are the structured per-field validation errors from the
 *  response (empty for non-validation failures). */
export interface SyncErrorInfo {
  status?: number;
  message: string;
  op?: SyncOp;
  fieldErrors: ParsedFieldError[];
}

export interface UseDraftSyncArgs {
  enabled: boolean;
  stack: Stack | undefined;
  session: UseStackEditSession;
  ids: { orgId: string; projectName: string; stackId: string } | null;
  onStackRefreshed: (stack: Stack) => void;
  /** Called when a sync op throws, before the mirror heals. */
  onSyncError?: (info: SyncErrorInfo) => void;
  /** When provided, replaces the hook's own getStackById for post-save and
   *  error-heal refetches. The provider owns applying the result to UI state
   *  (ticket-guarded), so onStackRefreshed is NOT called on those paths — the
   *  hook still heals its mirror from the returned stack. */
  refetchStack?: () => Promise<Stack>;
}

export interface UseDraftSync {
  status: SyncStatus;
  failureCount: number;
  /** True from the moment a draft change arms the debounce until the settling
   *  cycle finishes with nothing queued. Distinguishes "an edit exists that the
   *  server hasn't seen yet" from the status flags, which read settled during
   *  the pre-PUT debounce window. */
  pending: boolean;
  flush: () => Promise<boolean>;
  /** External writers (revert) hand the fresh stack here so the mirror stays truthful. */
  notifyExternalUpdate: (stack: Stack) => void;
}

type Ids = NonNullable<UseDraftSyncArgs["ids"]>;

async function executeOp(op: SyncOp, ids: Ids): Promise<void> {
  switch (op.kind) {
    case "createVolume":
      await createStackVolume(ids.orgId, ids.projectName, ids.stackId, op.volume);
      return;
    case "createResource":
      await createStackResource(ids.orgId, ids.projectName, ids.stackId, op.resource);
      return;
    case "updateResource":
      await updateStackResource(ids.orgId, ids.projectName, ids.stackId, op.name, op.resource);
      return;
    case "deleteResource":
      await deleteStackResource(ids.orgId, ids.projectName, ids.stackId, op.name);
      return;
    case "createConnection":
      await createStackConnection(ids.orgId, ids.projectName, ids.stackId, op.conn);
      return;
    case "updateConnection":
      await updateStackConnection(ids.orgId, ids.projectName, ids.stackId, op.id, op.conn);
      return;
    case "deleteConnection":
      await deleteStackConnection(ids.orgId, ids.projectName, ids.stackId, op.id);
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
  onSyncError,
  refetchStack,
}: UseDraftSyncArgs): UseDraftSync {
  const [status, setStatus] = useState<SyncStatus>(SYNC_STATUS.idle);
  const [failureCount, setFailureCount] = useState(0);
  const [pending, setPending] = useState(false);

  const sessionRef = useRef(session);
  sessionRef.current = session;
  const idsRef = useRef(ids);
  idsRef.current = ids;
  const onRefreshedRef = useRef(onStackRefreshed);
  onRefreshedRef.current = onStackRefreshed;
  const onSyncErrorRef = useRef(onSyncError);
  onSyncErrorRef.current = onSyncError;
  const refetchStackRef = useRef(refetchStack);
  refetchStackRef.current = refetchStack;

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
      // Track the op in flight so a throw can be pinned to it (ops.length > 0
      // here, so ops[0] is a safe non-null seed for the type checker).
      let failingOp: SyncOp = ops[0];
      // Flips true once every op has landed; a later throw then comes from the
      // post-save refetch, not from an op, so it must not misattribute a field
      // error to the last (successful) op.
      let opsSucceeded = false;
      try {
        for (const op of ops) {
          failingOp = op;
          await executeOp(op, currentIds);
        }
        opsSucceeded = true;
        const fetchFresh = refetchStackRef.current;
        const fresh = fetchFresh
          ? await fetchFresh()
          : await getStackById(currentIds.orgId, currentIds.projectName, currentIds.stackId);
        mirrorRef.current = serverStateFromStack(fresh);
        if (!fetchFresh) onRefreshedRef.current(fresh);
        // Deliberately NOT rebasing the session here: the diff baseline stays
        // pinned to the deployed release so autosaved edits remain visibly
        // dirty/revertable. Only deploy or discard moves the baseline.
        failuresRef.current = 0;
        setFailureCount(0);
        setStatus(SYNC_STATUS.saved);
        return true;
      } catch (err) {
        failuresRef.current += 1;
        setFailureCount(failuresRef.current);
        setStatus(SYNC_STATUS.error);
        const parsed = parseApiError(err);
        onSyncErrorRef.current?.({
          status: parsed.status,
          message: parsed.topLevel,
          fieldErrors: parsed.fieldErrors,
          // Ops all landed → the refetch failed, not an op. Leave op undefined so
          // the caller doesn't pin a spurious field error on a save that landed.
          op: opsSucceeded ? undefined : failingOp,
        });
        // Heal the mirror from server truth; the draft stays authoritative locally.
        try {
          const fetchFresh = refetchStackRef.current;
          const fresh = fetchFresh
            ? await fetchFresh()
            : await getStackById(currentIds.orgId, currentIds.projectName, currentIds.stackId);
          mirrorRef.current = serverStateFromStack(fresh);
          if (!fetchFresh) onRefreshedRef.current(fresh);
        } catch {
          /* keep the stale mirror; the next attempt refetches again */
        }
        // A 400 is a client-side validation error: retrying the identical draft
        // can only fail again. Leave the error status set (terminal) and wait for
        // the next draft edit to re-trigger a cycle via the debounce effect.
        // Non-400 (network/5xx) stays on the backoff-retry path.
        if (!isBadRequestError(err)) {
          const backoff = Math.min(RETRY_BASE_MS * 2 ** (failuresRef.current - 1), RETRY_MAX_MS);
          if (retryTimerRef.current) clearTimeout(retryTimerRef.current);
          retryTimerRef.current = setTimeout(() => {
            void startCycle();
          }, backoff);
        }
        return false;
      }
    })().finally(() => {
      runningRef.current = null;
      if (queuedRef.current) {
        queuedRef.current = false;
        idleTimerRef.current = setTimeout(() => {
          void startCycle();
        }, 0);
      } else {
        // Nothing queued behind this cycle: every known edit has been pushed
        // (or is on the retry path, which keeps failureCount > 0).
        setPending(false);
      }
    });
    runningRef.current = run;
    return run;
  }, []);

  // Debounce: every draft change (while active+enabled) restarts the idle
  // window; a max-wait timer guarantees persistence under continuous typing.
  const draft = session.isActive ? session.draft : null;
  // The debounce effect also fires once when a session starts (initial draft
  // identity, no edit yet) — that arm schedules a no-op cycle and must not
  // read as "pending edits". Only arms after the first one mark pending.
  const firstArmRef = useRef(true);
  useEffect(() => {
    if (!enabled || !draft || !idsRef.current) {
      firstArmRef.current = true;
      return;
    }
    if (firstArmRef.current) firstArmRef.current = false;
    else setPending(true);
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
    () => ({ status, failureCount, pending, flush, notifyExternalUpdate }),
    [status, failureCount, pending, flush, notifyExternalUpdate],
  );
}
