import type { Cluster } from "../types";
import { Button } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Boxes, Edit, Trash2 } from "lucide-react";

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
    <TooltipProvider>
      <div>
        {clusters.map((cluster) => (
          <div key={cluster.id} className="flex items-center justify-between p-4 border-b last:border-b-0">
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2">
                <Boxes className="h-4 w-4 text-muted-foreground" />
                <div>
                  <div className="font-medium">{cluster.name}</div>
                  <div className="text-xs text-muted-foreground">ID: {cluster.id}</div>
                </div>
              </div>
            </div>
            <div className="flex gap-1">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEdit(cluster)}
                    className="h-8 w-8 p-0"
                  >
                    <Edit className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Edit cluster</p>
                </TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(cluster)}
                    className="h-8 w-8 p-0 text-danger hover:text-danger hover:bg-danger-bg"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Delete cluster</p>
                </TooltipContent>
              </Tooltip>
            </div>
          </div>
        ))}
      </div>
    </TooltipProvider>
  );
}
