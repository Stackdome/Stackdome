import { useCallback, useEffect, useRef, useState } from "react";
import { listReleases, type StackRelease } from "@/api/releases";

const POLL_MS = 5000;
const TERMINAL = new Set<string>(["Released", "Failed", "Superseded", "Cancelled"]);

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
  // Ref tracks whether the active release is non-terminal so the poll interval
  // can check without depending on React state (avoids extra render cycle).
  const activeIsNonTerminal = useRef(false);

  const fetchOnce = useCallback(async () => {
    if (!enabled || inFlight.current) return;
    inFlight.current = true;
    setLoading(true);
    try {
      const data = await listReleases(orgId, teamName, stackId);
      if (!mounted.current) return;
      const sorted = [...(data.items ?? [])].sort(bySequenceDesc);
      const top = sorted[0];
      activeIsNonTerminal.current = !!top && !TERMINAL.has(top.state ?? "");
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

  // Always register the interval while enabled; gate the actual fetch on the ref
  // so we don't need an extra render cycle to start/stop polling.
  useEffect(() => {
    if (!enabled) return;
    const id = setInterval(() => {
      if (activeIsNonTerminal.current) void fetchOnce();
    }, POLL_MS);
    return () => clearInterval(id);
  }, [enabled, fetchOnce]);

  const active = releases[0];

  return { releases, activeRelease: active, loading, error, refetch: fetchOnce };
}
