// frontend/src/pages/stacks/hooks/use-stack-topology.ts
// Server topology feed for the canvas. Silent enhancement layer: failures fall
// back to null and the canvas renders the local graph alone.
import { useEffect, useState } from "react";
import { getStackTopology, type StackTopology } from "@/api/topology";

export interface UseStackTopologyArgs {
  /** Null disables the fetch (draft stacks have no server topology). */
  ids: { orgId: string; projectName: string; stackId: string } | null;
  /** Bump to refetch — wired to autosave refreshes. */
  refreshKey: number;
}

export function useStackTopology({ ids, refreshKey }: UseStackTopologyArgs): { topology: StackTopology | null } {
  const [topology, setTopology] = useState<StackTopology | null>(null);

  useEffect(() => {
    if (!ids) {
      setTopology(null);
      return;
    }
    let cancelled = false;
    getStackTopology(ids.orgId, ids.projectName, ids.stackId)
      .then((t) => {
        if (!cancelled) setTopology(t);
      })
      .catch((error) => {
        console.debug("stack topology fetch failed; canvas falls back to local graph", error);
        if (!cancelled) setTopology(null);
      });
    return () => {
      cancelled = true;
    };
  }, [ids?.orgId, ids?.projectName, ids?.stackId, refreshKey]); // eslint-disable-line react-hooks/exhaustive-deps -- keyed on primitive id parts

  return { topology };
}
