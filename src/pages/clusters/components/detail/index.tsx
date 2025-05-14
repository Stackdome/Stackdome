import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useDeleteCluster } from "../../hooks/use-clusters";
import * as clusterApi from "@/api/clusters";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Trash2, AlertCircle, Info, EyeOff } from "lucide-react";
import { getCurrentOrganizationId } from "@/helpers/common";
import {
  Card,
  CardContent,
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
import { useBreadcrumb } from "@/hooks/use-breadcrumb";
import { useToast } from "@/components/ui/use-toast";

export default function ClusterDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [cluster, setCluster] = useState<clusterApi.Cluster | null>(null);
  const [loading, setLoading] = useState(true);
  const { deleteCluster, loading: deleting, error: deleteError } = useDeleteCluster();
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const { setCustomLabel, setPathLoading } = useBreadcrumb();
  const { toast } = useToast();

  useEffect(() => {
    if (!id) return;
    const orgId = getCurrentOrganizationId();
    if (!orgId) return;

    const currentPath = `/clusters/${id}`;
    setPathLoading(currentPath, true);
    setLoading(true);
    clusterApi.getCluster(orgId, id)
      .then((data) => {
        setCluster(data);
        if (data && data.name) {
          setCustomLabel(currentPath, data.name);
        }
      })
      .finally(() => {
        setPathLoading(currentPath, false);
        setLoading(false);
      });
  }, [id, setCustomLabel, setPathLoading]);

  const handleDelete = async () => {
    if (!id) return;
    try {
      await deleteCluster(id);
      toast({
        title: "Success",
        description: "Cluster deleted successfully",
        variant: "success",
        duration: 3000,
      });
      navigate("/clusters");
    } catch (err) {
      console.error("Failed to delete cluster:", err);
      toast({
        title: "Failed to delete cluster",
        description: err instanceof Error ? err.message : "Failed to delete cluster. Please try again.",
        variant: "destructive",
        duration: 5000,
      });
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
    <div className="p-4 pt-0 h-full">
      <div className="flex h-[calc(100vh-64px)]">
        <div className="flex-grow p-6 overflow-y-auto">
          <div className="max-w-3xl mx-auto">
            <h2 className="text-xl font-medium mb-6">Cluster Details</h2>
            <Card className="w-full">
              <CardHeader>
                <CardTitle>
                  <span className="flex items-center gap-2">
                    {cluster.name}
                  </span>
                </CardTitle>
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
                    <div className="relative">
                      <Input 
                        value={cluster.cluster_url} 
                        disabled 
                        className="font-mono bg-muted" 
                      />
                    </div>
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
                    <div className="relative">
                      <Input 
                        type="password" 
                        value="••••••••••••••••••••••" 
                        disabled 
                        className="font-mono bg-muted" 
                      />
                      <div className="absolute inset-y-0 right-0 flex items-center px-3">
                        <EyeOff className="h-4 w-4 text-gray-400" />
                      </div>
                    </div>
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
                    <div className="relative">
                      <Input 
                        type="password" 
                        value="••••••••••••••••••••••" 
                        disabled 
                        className="font-mono bg-muted" 
                      />
                      <div className="absolute inset-y-0 right-0 flex items-center px-3">
                        <EyeOff className="h-4 w-4 text-gray-400" />
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 pt-2">
                    <Switch checked={!!(cluster as { image_registry_enabled?: boolean }).image_registry_enabled} disabled />
                    <span className="text-xs text-muted-foreground">Stackdome Image Registry</span>
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
                    <p>Are you sure you want to delete this cluster?</p>
                    <p>This action cannot be undone.</p>
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
        </div>
      </div>
    </div>
  );
}
