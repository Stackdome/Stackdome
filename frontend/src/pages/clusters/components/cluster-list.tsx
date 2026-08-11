import type { Cluster } from "../types";
import { Boxes, ChevronRight } from "lucide-react";

interface ClusterListProps {
  clusters: Cluster[];
  onOpen: (cluster: Cluster) => void;
}

export function ClusterList({ clusters, onOpen }: ClusterListProps) {
  if (!clusters.length) {
    return <div className="text-muted-foreground p-4">No clusters found.</div>;
  }

  return (
    <div className="divide-y divide-border">
      {clusters.map((cluster) => (
        <button
          key={cluster.id}
          type="button"
          onClick={() => onOpen(cluster)}
          className="flex w-full items-center gap-4 px-4 py-3 text-left transition-colors hover:bg-muted/50 focus-visible:outline-2 focus-visible:outline-[var(--ring)] focus-visible:outline-offset-2"
        >
          <div className="flex min-w-0 flex-1 items-center gap-3">
            <div aria-hidden className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg border border-border bg-card">
              <Boxes className="h-5 w-5 shrink-0 text-muted-foreground" />
            </div>
            <div className="min-w-0">
              <p className="truncate text-[15px] font-medium text-foreground">{cluster.name}</p>
              <p className="truncate font-mono text-[11.5px] text-fg-muted">{cluster.id}</p>
            </div>
          </div>
          <ChevronRight aria-hidden className="h-4 w-4 shrink-0 text-muted-foreground" />
        </button>
      ))}
    </div>
  );
}
