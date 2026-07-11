import { useCallback, useEffect, useRef, useState } from "react";
import { listReleases, type StackRelease } from "@/api/releases";

// Slow background refresh so a deploy started elsewhere (webhook push, another
// user) shows up while the page sits idle. Non-terminal releases no longer need
// a fast poll of their own — the release events stream (useReleaseEvents) pushes
// refetches on release-scoped events instead.
const IDLE_POLL_MS = 30000;

function isVisible(): boolean {
  return typeof document === "undefined" || document.visibilityState !== "hidden";
}

export interface UseReleasesArgs {
  orgId: string;
  teamName: string;
  stackId: string;
  enabled: boolean;
}

export interface UseReleasesResult {
  releases: StackRelease[];
  activeRelease: StackRelease | undefined;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

function bySequenceDesc(a: StackRelease, b: StackRelease): number {
  return (b.sequence ?? 0) - (a.sequence ?? 0);
}

export function useReleases({ orgId, teamName, stackId, enabled }: UseReleasesArgs): UseReleasesResult {
  const [releases, setReleases] = useState<StackRelease[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const inFlight = useRef(false);
  const mounted = useRef(true);

  const fetchOnce = useCallback(async () => {
    if (!enabled || inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    try {
      const data = await listReleases(orgId, teamName, stackId);
      if (!mounted.current) return;
      const sorted = [...(data.items ?? [])].sort(bySequenceDesc);
      setReleases(sorted);
      setError(null);
    } catch (e) {
      if (mounted.current) setError(e instanceof Error ? e.message : "Failed to load releases");
    } finally {
      if (mounted.current) setLoading(false);
      inFlight.current = false;
    }
  }, [orgId, teamName, stackId, enabled]);

  useEffect(() => {
    mounted.current = true;
    return () => { mounted.current = false; };
  }, []);

  useEffect(() => {
    if (!enabled) return;
    void fetchOnce();
  }, [enabled, fetchOnce]);

  // Idle safety net, paused while hidden, plus an immediate refetch on returning to the tab.
  useEffect(() => {
    if (!enabled) return;
    const slow = setInterval(() => { if (isVisible()) void fetchOnce(); }, IDLE_POLL_MS);
    const onVisible = () => { if (isVisible()) void fetchOnce(); };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      clearInterval(slow);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [enabled, fetchOnce]);

  const active = releases[0];

  return { releases, activeRelease: active, loading, error, refetch: fetchOnce };
}
