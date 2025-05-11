import { useParams, useNavigate } from "react-router-dom";
import { useEffect, useState } from "react";
import { useDeleteCluster } from "../../hooks/use-clusters";
import * as clusterApi from "@/api/clusters";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { ChevronLeft, Trash2, Info, Globe, Shield, Key, Check, AlertCircle } from "lucide-react";
import { getCurrentOrganizationId } from "@/helpers/common";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
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
        // After successful deletion, navigate back to clusters page
        // which will show the form to create a new cluster since no clusters exist
        navigate("/clusters");
      }
    } catch (error) {
      console.error("Failed to delete cluster:", error);
      // Error handling is done by the useDeleteCluster hook
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
    <div className="container max-w-4xl mx-auto py-8 px-4">
      <div className="flex items-center mb-6">
        <Button 
          variant="ghost" 
          className="mr-2 p-2" 
          onClick={() => navigate("/clusters")}
        >
          <ChevronLeft size={20} />
        </Button>
        <h1 className="text-2xl font-bold">Cluster Details</h1>
      </div>
      
      <Card className="mb-6">
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="text-xl">{cluster.name}</CardTitle>
              <CardDescription>Kubernetes Cluster</CardDescription>
            </div>
            <div className="flex items-center space-x-1 bg-green-100 text-green-600 px-2 py-1 rounded text-sm">
              <Check size={16} />
              <span>Connected</span>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-1">
              <div className="text-sm text-gray-500 flex items-center">
                <Globe size={16} className="mr-2" />
                API Server URL
              </div>
              <div className="font-mono text-sm bg-gray-50 p-2 rounded border">
                {cluster.cluster_url}
              </div>
            </div>
            
            <div className="space-y-1">
              <div className="text-sm text-gray-500 flex items-center">
                <Shield size={16} className="mr-2" />
                CA Certificate
              </div>
              <div className="font-mono text-sm bg-gray-50 p-2 rounded border overflow-hidden text-ellipsis whitespace-nowrap">
                {cluster.cluster_ca_data?.substring(0, 20)}...
              </div>
            </div>
            
            <div className="space-y-1 md:col-span-2">
              <div className="text-sm text-gray-500 flex items-center">
                <Key size={16} className="mr-2" />
                Service Account Token
              </div>
              <div className="font-mono text-sm bg-gray-50 p-2 rounded border overflow-hidden text-ellipsis whitespace-nowrap">
                {cluster.cluster_sa_token?.substring(0, 20)}...
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex justify-between border-t pt-4">
          <div className="text-sm text-gray-500 flex items-center">
            <Info size={16} className="mr-2" />
            ID: {cluster.id}
          </div>
          <Button 
            variant="destructive" 
            size="sm"
            onClick={() => setShowDeleteDialog(true)}
            className="flex items-center"
          >
            <Trash2 size={16} className="mr-2" />
            Delete Cluster
          </Button>
        </CardFooter>
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
