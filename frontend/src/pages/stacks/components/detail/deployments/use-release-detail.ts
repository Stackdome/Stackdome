import { createContext, createElement, useCallback, useContext, useMemo, useRef, useState, type ReactNode } from "react";
import { getRelease, type StackReleaseDetail } from "@/api/releases";

export interface DetailState { data?: StackReleaseDetail; loading: boolean; error?: string; }
export interface ReleaseDetail {
  ensure: (id: string) => void;
  peek: (id?: string) => DetailState;
  refresh: (id: string) => void;
}

const EMPTY: DetailState = { loading: false };

export function useReleaseDetail(orgId: string, teamName: string, stackId: string): ReleaseDetail {
  const [cache, setCache] = useState<Record<string, DetailState>>({});
  // Keep a ref to the latest cache so ensure() can read it without being in deps
  const cacheRef = useRef(cache);
  cacheRef.current = cache;
  const inFlight = useRef<Set<string>>(new Set());

  const ensure = useCallback((id: string) => {
    if (!id) return;
    const current = cacheRef.current[id];
    // Already cached (data or error), loading, or in-flight — no-op
    if (current?.data || current?.error || current?.loading || inFlight.current.has(id)) return;
    inFlight.current.add(id);
    setCache((c) => ({ ...c, [id]: { loading: true } }));
    getRelease(orgId, teamName, stackId, id)
      .then((data) => setCache((c) => ({ ...c, [id]: { loading: false, data } })))
      .catch((e) => setCache((c) => ({ ...c, [id]: { loading: false, error: e instanceof Error ? e.message : "Failed to load release" } })))
      .finally(() => inFlight.current.delete(id));
  }, [orgId, teamName, stackId]);

  const refresh = useCallback((id: string) => {
    if (!id || inFlight.current.has(id)) return;
    inFlight.current.add(id);
    setCache((c) => ({ ...c, [id]: { ...(c[id] ?? {}), loading: true } }));
    getRelease(orgId, teamName, stackId, id)
      .then((data) => setCache((c) => ({ ...c, [id]: { loading: false, data } })))
      .catch((e) => setCache((c) => ({ ...c, [id]: { loading: false, error: e instanceof Error ? e.message : "Failed to load release" } })))
      .finally(() => inFlight.current.delete(id));
  }, [orgId, teamName, stackId]);

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
