import type { Cluster } from "../types";
import { Button } from "@/components/ui/button";

interface ClusterListProps {
  clusters: Cluster[];
  onEdit: (cluster: Cluster) => void;
  onDelete: (cluster: Cluster) => void;
}

export function ClusterList({ clusters, onEdit, onDelete }: ClusterListProps) {
  if (!clusters.length) {
    return <div className="text-muted-foreground p-4">No clusters found.</div>;
  }
  return (
    <div className="space-y-2">
      {clusters.map((cluster) => (
        <div key={cluster.id} className="flex items-center justify-between border rounded p-3 bg-card">
          <div>
            <div className="font-medium">{cluster.name}</div>
            <div className="text-xs text-muted-foreground">ID: {cluster.id}</div>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => onEdit(cluster)}>
              Edit
            </Button>
            <Button variant="destructive" size="sm" onClick={() => onDelete(cluster)}>
              Delete
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
