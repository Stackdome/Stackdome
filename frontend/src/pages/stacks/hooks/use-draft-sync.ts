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
import { serverConnectionIndex, type ServerConnectionIndex } from "@/pages/stacks/lib/draft-sync/server-state";
import { canonicalFromStack } from "@/pages/stacks/lib/stack-model/from-api";
import { formResourcesFromSpec } from "@/pages/stacks/lib/spec-to-form";
import { deepEqual } from "@/pages/stacks/lib/stack-model/equal";
import { canonicalFromDraft } from "@/pages/stacks/lib/stack-model/from-form";
import type { CanonicalStack } from "@/pages/stacks/lib/stack-model/canonical";
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

  /** What the server holds, in the shape every comparison speaks, plus the
   *  connection ids the write path needs. */
  const mirrorRef = useRef<{ stack: CanonicalStack; connections: ServerConnectionIndex } | null>(null);
  const runningRef = useRef<Promise<boolean> | null>(null);
  const queuedRef = useRef(false);
  const failuresRef = useRef(0);
  const idleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const maxWaitTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  /** The draft the engine itself wrote while absorbing the server's answer.
   *  Identity, not a flag: a boolean would also swallow a keystroke that landed
   *  in the same React commit, and that edit would never be scheduled. */
  const adoptedDraftRef = useRef<unknown>(null);

  // Seed the mirror once from the fetched stack; afterwards the engine's own
  // refetches (and notifyExternalUpdate) keep it truthful.
  useEffect(() => {
    if (enabled && stack && !mirrorRef.current) {
      mirrorRef.current = { stack: canonicalFromStack(stack), connections: serverConnectionIndex(stack) };
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

  /**
   * Take the server's version of the resources the user was not editing.
   *
   * The draft is seeded once, when the session starts, and the server can
   * legitimately answer a write with more than it was sent — a default it
   * applied, a field it normalized. Left unreconciled that difference is
   * permanent: the draft says one thing, the server another, and every surface
   * downstream reports a change nobody made until the page is reloaded.
   *
   * Only entries whose object identity survived the whole cycle are adopted.
   * A resource typed into while the save was in flight keeps the user's version
   * and heals on the next cycle — a keystroke must never lose to a response.
   */
  const adoptServerState = useCallback((fresh: Stack, untouched: Set<object>) => {
    const session = sessionRef.current;
    if (!session.isActive) return;
    const serverForms = formResourcesFromSpec(fresh.spec?.stack_resources, fresh.spec?.connections);
    const byName = new Map(serverForms.map((r) => [r.name, r]));
    session.updateResources((current) => {
      let changed = false;
      const next = current.map((draftResource) => {
        if (!untouched.has(draftResource as unknown as object)) return draftResource;
        const server = draftResource.name ? byName.get(draftResource.name) : undefined;
        if (!server || deepEqual(draftResource, server)) return draftResource;
        changed = true;
        // The source stash is UI-only — the server has never heard of it, and
        // dropping it would empty the "Build from" toggle's memory mid-session.
        return {
          ...server,
          ...(draftResource.stashedGitSource ? { stashedGitSource: draftResource.stashedGitSource } : {}),
          ...(draftResource.stashedImageSource ? { stashedImageSource: draftResource.stashedImageSource } : {}),
        };
      });
      if (!changed) return current;
      adoptedDraftRef.current = next;
      return next;
    });
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
      // Identity of every resource as the cycle starts. Anything still holding
      // one of these references at the end was not typed into while the save
      // was in flight, and is safe to refresh from the server's answer.
      const untouched = new Set<object>(s.draft.resources as unknown as object[]);
      const ops = computeSyncOps(mirror.stack, canonicalFromDraft(snapshot), mirror.connections);
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
        mirrorRef.current = { stack: canonicalFromStack(fresh), connections: serverConnectionIndex(fresh) };
        if (!fetchFresh) onRefreshedRef.current(fresh);
        adoptServerState(fresh, untouched);
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
          mirrorRef.current = { stack: canonicalFromStack(fresh), connections: serverConnectionIndex(fresh) };
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
        // An edit made mid-cycle may also have re-armed the debounce; clear it
        // so the drain timer below is the only pending firing (a leaked handle
        // would fire a spurious extra cycle).
        if (idleTimerRef.current) clearTimeout(idleTimerRef.current);
        idleTimerRef.current = setTimeout(() => {
          void startCycle();
        }, 0);
      } else if (!idleTimerRef.current && !maxWaitTimerRef.current) {
        // Nothing queued AND no debounce armed: every known edit has been
        // pushed (or is on the retry path, which keeps failureCount > 0). An
        // armed timer means an edit landed DURING this cycle — it hasn't been
        // sent, so pending must survive until its own cycle settles.
        setPending(false);
      }
    });
    runningRef.current = run;
    return run;
  }, [adoptServerState]);

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
      // Disabled / session ended: no cycle will run to clear pending, so
      // reset it here rather than leaving the lifecycle stuck on "editing".
      setPending(false);
      return;
    }
    // The engine rewrote the draft to match the server it just read. That is
    // not an edit waiting to be saved, and there is nothing to send back — but
    // only for that exact draft; anything later is the user typing.
    if (draft.resources === adoptedDraftRef.current) {
      adoptedDraftRef.current = null;
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
    mirrorRef.current = { stack: canonicalFromStack(fresh), connections: serverConnectionIndex(fresh) };
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
