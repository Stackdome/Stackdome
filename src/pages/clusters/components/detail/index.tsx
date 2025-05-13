import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useDeleteCluster } from "../../hooks/use-clusters";
import * as clusterApi from "@/api/clusters";
import { Button } from "@/components/ui/button";
import { Trash2, AlertCircle, Info } from "lucide-react";
import { getCurrentOrganizationId } from "@/helpers/common";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Switch } from "@/components/ui/switch";

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [cluster, setCluster] = useState<clusterApi.Cluster | null>(null);
  const [loading, setLoading] = useState(true);
  const { deleteCluster, loading: deleting, error: deleteError } = useDeleteCluster();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    clusterApi.getCluster(getCurrentOrganizationId(), id)
      .then(setCluster)
      .finally(() => setLoading(false));
  }, [id]);

  const handleDelete = async () => {
    if (!id) return;
    try {
      const success = await deleteCluster(id);
      if (success) {
        navigate("/clusters");
      }
    } catch (error) {
      console.error("Failed to delete cluster:", error);
    }
  };

  if (loading) return (
    <div className="flex items-center justify-center h-screen">
      <div className="text-gray-500">Loading cluster details...</div>
    </div>
  );
  
  if (!cluster) return (
    <div className="flex flex-col items-center justify-center h-screen">
      <div className="text-xl font-medium mb-2">Cluster not found</div>
      <p className="text-gray-500 mb-4">The cluster you're looking for doesn't exist or you don't have access.</p>
      <Button onClick={() => navigate("/clusters")}>
        Back to Clusters
      </Button>
    </div>
  );

  return (
    <div className="container max-w-2xl mx-auto py-8 px-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-xl">{cluster.name}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          <TooltipProvider>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">API Server URL</span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent>
                    The endpoint for your Kubernetes API server.
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="font-mono text-sm bg-muted p-2 rounded border">{cluster.cluster_url}</div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">CA Certificate</span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent>
                    The certificate authority data for your cluster.
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="font-mono text-sm bg-muted p-2 rounded border overflow-hidden text-ellipsis whitespace-nowrap">{cluster.cluster_ca_data?.substring(0, 20)}...</div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <span className="text-xs text-muted-foreground">Service Account Token</span>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                  </TooltipTrigger>
                  <TooltipContent>
                    The service account token used for authentication.
                  </TooltipContent>
                </Tooltip>
              </div>
              <div className="font-mono text-sm bg-muted p-2 rounded border overflow-hidden text-ellipsis whitespace-nowrap">{cluster.cluster_sa_token?.substring(0, 20)}...</div>
            </div>
            <div className="flex items-center gap-2 pt-2">
              <Switch checked={!!(cluster as { image_registry_enabled?: boolean }).image_registry_enabled} disabled />
              <span className="text-xs text-muted-foreground">Enable Image Registry</span>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Info className="size-3.5 text-muted-foreground cursor-pointer" />
                </TooltipTrigger>
                <TooltipContent>
                  Allow this cluster to use the built-in image registry.
                </TooltipContent>
              </Tooltip>
            </div>
          </TooltipProvider>
        </CardContent>
        <div className="flex justify-between items-center pt-4 px-6 pb-6">
          <div className="text-xs text-muted-foreground">ID: {cluster.id}</div>
          <Button 
            variant="destructive" 
            size="sm"
            onClick={() => setShowDeleteDialog(true)}
            className="flex items-center"
          >
            <Trash2 className="size-4 mr-2" />
            Delete Cluster
          </Button>
        </div>
      </Card>
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Delete Cluster</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this cluster? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          {deleteError && (
            <div className="bg-red-50 text-red-500 p-3 rounded flex items-center">
              <AlertCircle size={16} className="mr-2 flex-shrink-0" />
              <span>{deleteError}</span>
            </div>
          )}
          <DialogFooter className="flex justify-end space-x-2">
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
              disabled={deleting}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting}
              className="flex items-center"
            >
              {deleting ? "Deleting..." : "Delete Cluster"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
