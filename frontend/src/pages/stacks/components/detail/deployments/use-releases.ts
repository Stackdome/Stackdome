import { useCallback, useEffect, useRef, useState } from "react";
import { listReleases, type StackRelease } from "@/api/releases";
import { isTerminal } from "./release-states";

const POLL_MS = 5000;
// Slow background refresh so a deploy started elsewhere (webhook push, another
// user) shows up while the page sits idle with everything terminal.
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
  // Ref tracks whether ANY release is still non-terminal so the poll keeps the
  // list fresh until every release settles — this catches trailing supersessions
  // of earlier releases after the latest one has already finished. Gating only on
  // the latest would stop polling while an earlier Pending release is still
  // resolving, leaving it stuck on a stale state until a manual refresh.
  const hasPendingWork = useRef(false);

  const fetchOnce = useCallback(async () => {
    if (!enabled || inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    try {
      const data = await listReleases(orgId, teamName, stackId);
      if (!mounted.current) return;
      const sorted = [...(data.items ?? [])].sort(bySequenceDesc);
      hasPendingWork.current = sorted.some((r) => !isTerminal(r.state));
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

  // Two cadences, both paused while the tab is hidden:
  //  - fast (POLL_MS): only while a release is still settling.
  //  - slow (IDLE_POLL_MS): always, so external/idle changes still surface.
  // Returning to a hidden tab refetches immediately rather than waiting a tick.
  useEffect(() => {
    if (!enabled) return;
    const fast = setInterval(() => { if (hasPendingWork.current && isVisible()) void fetchOnce(); }, POLL_MS);
    const slow = setInterval(() => { if (isVisible()) void fetchOnce(); }, IDLE_POLL_MS);
    const onVisible = () => { if (isVisible()) void fetchOnce(); };
    document.addEventListener("visibilitychange", onVisible);
    return () => {
      clearInterval(fast);
      clearInterval(slow);
      document.removeEventListener("visibilitychange", onVisible);
    };
  }, [enabled, fetchOnce]);

  const active = releases[0];

  return { releases, activeRelease: active, loading, error, refetch: fetchOnce };
}
