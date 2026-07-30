import { createContext, createElement, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from "react";
import { getRelease, type StackReleaseDetail } from "@/api/releases";

export interface DetailState { data?: StackReleaseDetail; loading: boolean; error?: string; }
export interface ReleaseDetail {
  ensure: (id: string) => void;
  peek: (id?: string) => DetailState;
  refresh: (id: string) => void;
}

const EMPTY: DetailState = { loading: false };

// A cached error is not permanent — a transient failure loading a (previous)
// release's snapshot must be retriable, or a consumer gating on `!data` spins
// forever. Cool down between auto-retries so ensure() can't hammer a hard 4xx.
const RETRY_AFTER_ERROR_MS = 10000;

export function useReleaseDetail(orgId: string, projectName: string, stackId: string): ReleaseDetail {
  const [cache, setCache] = useState<Record<string, DetailState>>({});
  // Keep a ref to the latest cache so ensure() can read it without being in deps
  const cacheRef = useRef(cache);
  cacheRef.current = cache;
  const inFlight = useRef<Set<string>>(new Set());
  const erroredAt = useRef<Map<string, number>>(new Map());
  const refreshQueued = useRef<Set<string>>(new Set());

  // projectName resolves asynchronously (org project list fetch) after stackId is
  // already known — no-op until all three ids are in, to avoid /projects//... calls.
  const ready = !!orgId && !!projectName && !!stackId;

  // Shared fetch. On settle, drain a refresh queued while this was in flight — whichever
  // caller (ensure or refresh) held the slot, the trailing refresh must run once.
  const fetchInto = useCallback((id: string) => {
    inFlight.current.add(id);
    setCache((c) => ({ ...c, [id]: { ...(c[id] ?? {}), loading: true } }));
    getRelease(orgId, projectName, stackId, id)
      .then((data) => { erroredAt.current.delete(id); setCache((c) => ({ ...c, [id]: { loading: false, data } })); })
      .catch((e) => { erroredAt.current.set(id, Date.now()); setCache((c) => ({ ...c, [id]: { loading: false, error: e instanceof Error ? e.message : "Failed to load release" } })); })
      .finally(() => {
        inFlight.current.delete(id);
        if (refreshQueued.current.has(id) && ready) { refreshQueued.current.delete(id); fetchInto(id); }
      });
  }, [ready, orgId, projectName, stackId]);

  const ensure = useCallback((id: string) => {
    if (!ready || !id) return;
    const current = cacheRef.current[id];
    if (current?.data || current?.loading || inFlight.current.has(id)) return;
    // A prior error blocks re-fetch only during the cooldown, then one retry is allowed.
    if (current?.error && Date.now() - (erroredAt.current.get(id) ?? 0) < RETRY_AFTER_ERROR_MS) return;
    fetchInto(id);
  }, [ready, fetchInto]);

  const refresh = useCallback((id: string) => {
    if (!ready || !id) return;
    // A refresh arriving while one is in flight (e.g. the terminal event landing on top
    // of a prior event's or ensure's fetch) must run once the current fetch settles — the
    // in-flight response predates the event, so dropping it strands stale live_status.
    if (inFlight.current.has(id)) { refreshQueued.current.add(id); return; }
    fetchInto(id);
  }, [ready, fetchInto]);

  const peek = useCallback((id?: string): DetailState => (id ? cache[id] ?? EMPTY : EMPTY), [cache]);
  // Memoized: peek still changes identity on cache updates, but ensure/refresh stay
  // stable, so effects can safely depend on the callbacks they use.
  return useMemo(() => ({ ensure, peek, refresh }), [ensure, peek, refresh]);
}

const ReleaseDetailContext = createContext<ReleaseDetail | null>(null);

export function ReleaseDetailProvider({ value, children }: { value: ReleaseDetail; children: ReactNode }) {
  return createElement(ReleaseDetailContext.Provider, { value }, children);
}

export function useReleaseDetailContext(): ReleaseDetail {
  return useContext(ReleaseDetailContext)!;
}
